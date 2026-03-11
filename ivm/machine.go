package ivm

import (
"bufio"
"english/astvm/types"
"fmt"
"os"
"strings"
)

// ─── Machine ──────────────────────────────────────────────────────────────────

type tryFrame struct {
catchOffset uint32
stackHeight int
envDepth    int // number of envs on envStack when TRY_BEGIN was emitted
}

type callFrame struct {
chunk    *Chunk
ip       int
stack    []interface{}
env      *ivmEnv
envStack []*ivmEnv // scopes pushed during this frame
tryStack []tryFrame
line     int
name     string
}

// Machine executes a compiled Chunk.
type Machine struct {
frames  []*callFrame
cur     *callFrame
builtin BuiltinFunc
// importHandler is called for OP_IMPORT; if nil, imports are silently skipped.
importHandler func(path string, items []interface{}, importAll, isSafe bool, env *ivmEnv) error
}

func newMachine(builtin BuiltinFunc) *Machine {
return &Machine{builtin: builtin}
}

func (m *Machine) push(v interface{}) {
m.cur.stack = append(m.cur.stack, v)
}

func (m *Machine) pop() interface{} {
n := len(m.cur.stack) - 1
v := m.cur.stack[n]
m.cur.stack = m.cur.stack[:n]
return v
}

func (m *Machine) peek() interface{} {
return m.cur.stack[len(m.cur.stack)-1]
}

func (m *Machine) env() *ivmEnv {
return m.cur.env
}

func (m *Machine) pushEnv() {
child := m.cur.env.newChild()
m.cur.envStack = append(m.cur.envStack, m.cur.env)
m.cur.env = child
}

func (m *Machine) popEnv() {
n := len(m.cur.envStack) - 1
m.cur.env = m.cur.envStack[n]
m.cur.envStack = m.cur.envStack[:n]
}

func (m *Machine) runtimeErr(msg string) error {
return &machineError{message: msg, line: m.cur.line, frame: m.cur.name}
}

type machineError struct {
message string
line    int
frame   string
}

func (e *machineError) Error() string {
if e.line > 0 {
return fmt.Sprintf("Runtime Error at line %d: %s", e.line, e.message)
}
return fmt.Sprintf("Runtime Error: %s", e.message)
}

// execute runs the machine until the outermost frame returns.
func (m *Machine) execute(env *ivmEnv) (interface{}, error) {
for {
frame := m.cur
if frame.ip >= len(frame.chunk.Code) {
// Implicit return nil at end of top-level code
if len(m.frames) == 0 {
return nil, nil
}
// If somehow frames remain, pop and continue
m.cur = m.frames[len(m.frames)-1]
m.frames = m.frames[:len(m.frames)-1]
m.push(interface{}(nil))
continue
}

instr := frame.chunk.Code[frame.ip]
frame.ip++

result, stop, err := m.step(instr, frame.chunk)
if err != nil {
// errCaughtByParent means callFuncChunk detected that handleError()
// already set m.cur to this frame's catch handler.  Just continue.
if _, ok := err.(errCaughtByParent); ok {
continue
}
// Check if there's a try frame to catch this
caught, jumpErr := m.handleError(err)
if jumpErr != nil {
return nil, jumpErr
}
if caught {
continue
}
return nil, err
}
if stop {
return result, nil
}
}
}

func (m *Machine) handleError(err error) (bool, error) {
// Walk up the call stack looking for a try frame
for {
frame := m.cur
if len(frame.tryStack) > 0 {
tf := frame.tryStack[len(frame.tryStack)-1]
frame.tryStack = frame.tryStack[:len(frame.tryStack)-1]

// Restore stack to height at TRY_BEGIN
frame.stack = frame.stack[:tf.stackHeight]

// Restore env scopes to depth at TRY_BEGIN
for len(frame.envStack) > tf.envDepth {
m.popEnv()
}

// Convert error to ErrorValue
var ev *types.ErrorValue
switch e := err.(type) {
case *types.ErrorValue:
ev = e
case *machineError:
ev = &types.ErrorValue{Message: e.message, ErrorType: "RuntimeError"}
default:
ev = &types.ErrorValue{Message: err.Error(), ErrorType: "RuntimeError"}
}
m.push(ev)
frame.ip = int(tf.catchOffset)
return true, nil
}

// No try frame in current frame; pop frame and propagate
if len(m.frames) == 0 {
return false, nil
}
// Pop the current frame and continue looking
m.cur = m.frames[len(m.frames)-1]
m.frames = m.frames[:len(m.frames)-1]
}
}

func (m *Machine) step(instr Instruction, chunk *Chunk) (result interface{}, stop bool, err error) {
op := instr.Op
operand := instr.Operand

switch op {
case OP_LOAD_CONST:
m.push(chunk.Constants[operand])

case OP_LOAD_NOTHING:
m.push(nil)

case OP_LOAD_VAR:
name := chunk.Names[operand]
val, ok := m.env().getVar(name)
if !ok {
return nil, false, m.runtimeErr(fmt.Sprintf("undefined variable '%s'", name))
}
m.push(val)

case OP_STORE_VAR:
name := chunk.Names[operand]
val := m.pop()
if err := m.env().setVar(name, val); err != nil {
return nil, false, m.runtimeErr(err.Error())
}

case OP_DEFINE_VAR:
name := chunk.Names[operand]
val := m.pop()
if err := m.env().defineVar(name, val, false); err != nil {
return nil, false, m.runtimeErr(err.Error())
}

case OP_DEFINE_CONST:
name := chunk.Names[operand]
val := m.pop()
if err := m.env().defineVar(name, val, true); err != nil {
return nil, false, m.runtimeErr(err.Error())
}

case OP_DEFINE_TYPED:
name := chunk.Names[operand]
val := m.pop()
typeName, ok := m.pop().(string)
if !ok {
return nil, false, m.runtimeErr("DEFINE_TYPED: expected type name string on stack")
}
if err := m.env().defineTypedVar(name, typeName, val, false); err != nil {
return nil, false, m.runtimeErr(err.Error())
}

case OP_DEFINE_TYPED_CONST:
name := chunk.Names[operand]
val := m.pop()
typeName, ok := m.pop().(string)
if !ok {
return nil, false, m.runtimeErr("DEFINE_TYPED_CONST: expected type name string on stack")
}
if err := m.env().defineTypedVar(name, typeName, val, true); err != nil {
return nil, false, m.runtimeErr(err.Error())
}

case OP_TOGGLE_VAR:
name := chunk.Names[operand]
val, ok := m.env().getVar(name)
if !ok {
return nil, false, m.runtimeErr(fmt.Sprintf("undefined variable '%s'", name))
}
b, ok := val.(bool)
if !ok {
return nil, false, m.runtimeErr(fmt.Sprintf("toggle: '%s' is not a boolean", name))
}
if err := m.env().setVar(name, !b); err != nil {
return nil, false, m.runtimeErr(err.Error())
}

case OP_BINARY_OP:
right := m.pop()
left := m.pop()
res, err := doBinaryOp(BinOp(operand), left, right)
if err != nil {
return nil, false, m.runtimeErr(err.Error())
}
m.push(res)

case OP_UNARY_OP:
val := m.pop()
res, err := doUnaryOp(UnaryOp(operand), val)
if err != nil {
return nil, false, m.runtimeErr(err.Error())
}
m.push(res)

case OP_JUMP:
m.cur.ip = int(operand)

case OP_JUMP_IF_FALSE:
val := m.pop()
b, err := ivmToBool(val)
if err != nil {
return nil, false, m.runtimeErr(err.Error())
}
if !b {
m.cur.ip = int(operand)
}

case OP_JUMP_IF_TRUE:
val := m.pop()
b, err := ivmToBool(val)
if err != nil {
return nil, false, m.runtimeErr(err.Error())
}
if b {
m.cur.ip = int(operand)
}

case OP_PUSH_SCOPE:
m.pushEnv()

case OP_POP_SCOPE:
m.popEnv()

case OP_DEFINE_FUNC:
fc := chunk.Funcs[operand]
m.env().defineFunc(fc.Name, fc)

case OP_CALL:
argc := int(operand >> 16)
nameIdx := operand & 0xFFFF
name := chunk.Names[nameIdx]
args := make([]interface{}, argc)
for i := argc - 1; i >= 0; i-- {
args[i] = m.pop()
}
res, err := m.callFunction(name, args, chunk)
if err != nil {
return nil, false, err
}
m.push(res)

case OP_CALL_METHOD:
argc := int(operand >> 16)
methodNameIdx := operand & 0xFFFF
methodName := chunk.Names[methodNameIdx]
args := make([]interface{}, argc)
for i := argc - 1; i >= 0; i-- {
args[i] = m.pop()
}
obj := m.pop()
res, err := m.callMethod(obj, methodName, args, chunk)
if err != nil {
return nil, false, err
}
m.push(res)

case OP_RETURN:
var retVal interface{}
if len(m.cur.stack) > 0 {
retVal = m.pop()
}
// Always signal stop; each loop (execute/callFuncChunk) handles frame restoration
return retVal, true, nil

case OP_PRINT:
newline := operand & 1
count := int(operand >> 1)
parts := make([]string, count)
for i := count - 1; i >= 0; i-- {
parts[i] = ivmToString(m.pop())
}
text := strings.Join(parts, " ")
if newline == 1 {
fmt.Println(text)
} else {
fmt.Print(text)
}

case OP_BUILD_LIST:
count := int(operand)
elems := make([]interface{}, count)
for i := count - 1; i >= 0; i-- {
elems[i] = m.pop()
}
m.push(elems)

case OP_BUILD_ARRAY:
count := int(operand)
typeName, ok := m.pop().(string)
if !ok {
return nil, false, m.runtimeErr("BUILD_ARRAY: expected type name string")
}
elems := make([]interface{}, count)
for i := count - 1; i >= 0; i-- {
elems[i] = m.pop()
}
elemKind := types.Parse(typeName)
m.push(&types.ArrayValue{ElementType: elemKind, Elements: elems})

case OP_BUILD_LOOKUP:
m.push(&types.LookupTableValue{
Entries:  make(map[string]interface{}),
KeyOrder: []string{},
})

case OP_INDEX_GET:
index := m.pop()
container := m.pop()
res, err := doIndexGet(container, index)
if err != nil {
return nil, false, m.runtimeErr(err.Error())
}
m.push(res)

case OP_INDEX_SET:
name := chunk.Names[operand]
val := m.pop()
index := m.pop()
container, ok := m.env().getVar(name)
if !ok {
return nil, false, m.runtimeErr(fmt.Sprintf("undefined variable '%s'", name))
}
if err := doIndexSet(container, index, val); err != nil {
return nil, false, m.runtimeErr(err.Error())
}

case OP_LENGTH:
val := m.pop()
n, err := doLength(val)
if err != nil {
return nil, false, m.runtimeErr(err.Error())
}
m.push(n)

case OP_LOOKUP_GET:
key := m.pop()
table := m.pop()
res, err := doLookupGet(table, key)
if err != nil {
return nil, false, m.runtimeErr(err.Error())
}
m.push(res)

case OP_LOOKUP_SET:
name := chunk.Names[operand]
val := m.pop()
key := m.pop()
tableVal, ok := m.env().getVar(name)
if !ok {
return nil, false, m.runtimeErr(fmt.Sprintf("undefined variable '%s'", name))
}
lt, ok := tableVal.(*types.LookupTableValue)
if !ok {
return nil, false, m.runtimeErr(fmt.Sprintf("'%s' is not a lookup table", name))
}
k, err := types.SerializeKey(key)
if err != nil {
return nil, false, m.runtimeErr(err.Error())
}
if _, exists := lt.Entries[k]; !exists {
lt.KeyOrder = append(lt.KeyOrder, k)
}
lt.Entries[k] = val

case OP_LOOKUP_HAS:
key := m.pop()
table := m.pop()
lt, ok := table.(*types.LookupTableValue)
if !ok {
return nil, false, m.runtimeErr("LOOKUP_HAS: not a lookup table")
}
k, err := types.SerializeKey(key)
if err != nil {
return nil, false, m.runtimeErr(err.Error())
}
_, has := lt.Entries[k]
m.push(has)

case OP_TYPEOF:
val := m.pop()
m.push(ivmGetTypeName(val))

case OP_CAST:
typeName := chunk.Names[operand]
val := m.pop()
target := types.Parse(typeName)
var res interface{}
if target == types.TypeString {
res = ivmToString(val)
} else {
var castErr error
res, castErr = types.Cast(val, target)
if castErr != nil {
return nil, false, &types.ErrorValue{Message: castErr.Error(), ErrorType: "TypeError"}
}
}
m.push(res)

case OP_NIL_CHECK:
val := m.pop()
if operand == 1 { // is_something
m.push(val != nil)
} else { // is_nothing
m.push(val == nil)
}

case OP_ERROR_TYPE_CHECK:
typeName := chunk.Names[operand]
val := m.pop()
ev, ok := val.(*types.ErrorValue)
if !ok {
m.push(false)
} else {
m.push(m.env().isSubtypeOf(ev.ErrorType, typeName))
}

case OP_ASK:
if operand == 1 {
p := m.pop()
fmt.Print(ivmToString(p))
}
scanner := bufio.NewScanner(os.Stdin)
if scanner.Scan() {
m.push(scanner.Text())
} else {
m.push("")
}

case OP_LOCATION:
name := chunk.Names[operand]
val, ok := m.env().getVar(name)
if !ok {
return nil, false, m.runtimeErr(fmt.Sprintf("undefined variable '%s'", name))
}
m.push(fmt.Sprintf("%p", &val))

case OP_DEFINE_STRUCT:
sd := chunk.StructDefs[operand]
m.env().defineStructDef(sd.Name, sd)

case OP_NEW_STRUCT:
fieldCount := int(operand >> 16)
snIdx := operand & 0xFFFF
structName := chunk.Names[snIdx]
sd, ok := m.env().getStructDef(structName)
if !ok {
return nil, false, m.runtimeErr(fmt.Sprintf("undefined struct '%s'", structName))
}

// Pop field values in reverse order (last field pushed = top of stack)
fieldVals := make([]interface{}, fieldCount)
for i := fieldCount - 1; i >= 0; i-- {
fieldVals[i] = m.pop()
}

inst := &StructInstance{
DefName: structName,
DefRef:  sd,
Fields:  make(map[string]interface{}),
}

// Assign fields in FieldOrder from StructInstantiation (same order as compiled)
// The fieldCount fields were compiled from FieldOrder, so we use sd.Fields for defaults
// and the compiled values for specified ones.
// Since we compile in FieldOrder (from StructInstantiation), we need to map them back.
// However, the compiler pushes them in the FieldOrder from the instantiation,
// not necessarily the struct definition order.
// We store them positionally, so we need the same order.
// For simplicity: if fieldCount > 0, use fieldVals as-is in the order they were compiled.
// The struct definition's field order is in sd.Fields (slice).
// The instantiation's field order is what was compiled. We don't have that info here.
// Solution: compile ALL struct fields in struct definition order (see compileStructInstantiation).

// We reconstruct by matching against sd.Fields order
for i, fd := range sd.Fields {
var fval interface{}
if i < fieldCount {
fval = fieldVals[i]
}
if fval == nil && fd.DefaultExprChunk != nil {
// Execute default expression chunk
var defErr error
fval, defErr = m.executeDefaultChunk(fd.DefaultExprChunk)
if defErr != nil {
return nil, false, m.runtimeErr(defErr.Error())
}
}
if fval == nil {
// Use type default
fval = typeDefault(fd.TypeName)
}
inst.Fields[fd.Name] = fval
}
m.push(inst)

case OP_GET_FIELD:
fieldName := chunk.Names[operand]
obj := m.pop()
si, ok := obj.(*StructInstance)
if !ok {
return nil, false, m.runtimeErr(fmt.Sprintf("GET_FIELD: not a struct instance (got %T)", obj))
}
val, exists := si.Fields[fieldName]
if !exists {
return nil, false, m.runtimeErr(fmt.Sprintf("struct '%s' has no field '%s'", si.DefName, fieldName))
}
m.push(val)

case OP_SET_FIELD:
fieldName := chunk.Names[operand]
newVal := m.pop()
obj := m.pop()
si, ok := obj.(*StructInstance)
if !ok {
return nil, false, m.runtimeErr(fmt.Sprintf("SET_FIELD: not a struct instance (got %T)", obj))
}
si.Fields[fieldName] = newVal

case OP_RAISE:
msg := ivmToString(m.pop())
var errType string
if operand > 0 && int(operand-1) < len(chunk.Names) {
errType = chunk.Names[operand-1]
} else {
errType = "RuntimeError"
}
return nil, false, &types.ErrorValue{Message: msg, ErrorType: errType}

case OP_TRY_BEGIN:
tf := tryFrame{
catchOffset: operand,
stackHeight: len(m.cur.stack),
envDepth:    len(m.cur.envStack),
}
m.cur.tryStack = append(m.cur.tryStack, tf)

case OP_TRY_END:
// Pop the try frame (no error occurred)
if len(m.cur.tryStack) > 0 {
m.cur.tryStack = m.cur.tryStack[:len(m.cur.tryStack)-1]
}
// Jump past the catch body to the finally/end section
m.cur.ip = int(operand)

case OP_CATCH:
// The error value is on top of stack (pushed by handleError)
errVarIdx := operand >> 16
errTypeIdx := operand & 0xFFFF

errVal, ok := m.peek().(*types.ErrorValue)
if !ok {
// Not a typed error, just leave it (unlikely)
break
}

// Check error type filter
if errTypeIdx > 0 {
errTypeName := chunk.Names[errTypeIdx-1]
if !m.env().isSubtypeOf(errVal.ErrorType, errTypeName) {
// Type mismatch: re-raise the error
_ = m.pop()
return nil, false, errVal
}
}

// Bind error to variable
if errVarIdx < uint32(len(chunk.Names)) {
errVarName := chunk.Names[errVarIdx]
if errVarName != "" {
_ = m.pop()
m.env().defineVar(errVarName, errVal, false)
} else {
m.pop()
}
} else {
m.pop()
}

case OP_DEFINE_ERROR_TYPE:
nameIdx := operand >> 16
parentIdx := operand & 0xFFFF
name := chunk.Names[nameIdx]
var parent string
if parentIdx > 0 {
parent = chunk.Names[parentIdx-1]
}
m.env().defineErrorType(name, parent)

case OP_MAKE_REFERENCE:
name := chunk.Names[operand]
_, ok := m.env().getVar(name)
if !ok {
return nil, false, m.runtimeErr(fmt.Sprintf("undefined variable '%s'", name))
}
m.push(&ReferenceValue{Name: name, Env: m.env()})

case OP_MAKE_COPY:
val := m.pop()
m.push(deepCopyValue(val))

case OP_SWAP_VARS:
n1Idx := operand >> 16
n2Idx := operand & 0xFFFF
name1 := chunk.Names[n1Idx]
name2 := chunk.Names[n2Idx]
v1, ok1 := m.env().getVar(name1)
if !ok1 {
return nil, false, m.runtimeErr(fmt.Sprintf("undefined variable '%s'", name1))
}
v2, ok2 := m.env().getVar(name2)
if !ok2 {
return nil, false, m.runtimeErr(fmt.Sprintf("undefined variable '%s'", name2))
}
if err := m.env().setVar(name1, v2); err != nil {
return nil, false, m.runtimeErr(err.Error())
}
if err := m.env().setVar(name2, v1); err != nil {
return nil, false, m.runtimeErr(err.Error())
}

case OP_IMPORT:
flags := operand
hasItems := flags&1 != 0
isSafe := flags&2 != 0
importAll := flags&4 != 0

var items []interface{}
if hasItems {
itemsVal := m.pop()
items, _ = itemsVal.([]interface{})
}
path, ok := m.pop().(string)
if !ok {
return nil, false, m.runtimeErr("IMPORT: expected path string on stack")
}

if m.importHandler != nil {
if err := m.importHandler(path, items, importAll, isSafe, m.env()); err != nil {
return nil, false, m.runtimeErr(err.Error())
}
}
// If no handler, silently skip import

case OP_SET_LINE:
m.cur.line = int(operand)

case OP_POP:
if len(m.cur.stack) > 0 {
m.pop()
}

default:
return nil, false, m.runtimeErr(fmt.Sprintf("unknown opcode: %d (%s)", op, OpName(op)))
}

return nil, false, nil
}

func (m *Machine) callFunction(name string, args []interface{}, callerChunk *Chunk) (interface{}, error) {
// Look up user-defined function
fn, ok := m.env().getFunc(name)
if ok {
return m.callFuncChunk(fn, args, nil)
}
// Fall back to builtin
if m.builtin != nil {
res, err := m.builtin(name, args)
if err != nil {
// stdlib.Eval returns "unknown built-in function: X" for names it
// doesn't recognise. That message is misleading when the function
// was supposed to be user-defined (e.g. imported but import was
// skipped). Normalise to "undefined function" so the user gets a
// clear, actionable error.
if strings.Contains(err.Error(), "unknown built-in function:") {
return nil, m.runtimeErr(fmt.Sprintf("undefined function '%s'", name))
}
return nil, err
}
return res, nil
}
return nil, m.runtimeErr(fmt.Sprintf("undefined function '%s'", name))
}

// errCaughtByParent is a sentinel returned by callFuncChunk when an error
// escapes the function and is caught by a try block in a *parent* frame.
// handleError has already rewound the frame stack and set m.cur to the parent
// frame's catch handler.  The caller must not call handleError again on this
// sentinel; instead it should propagate it upward until it reaches execute(),
// which simply continues the main loop (now running the catch handler).
type errCaughtByParent struct{}

func (errCaughtByParent) Error() string { return "caught by parent frame" }

func (m *Machine) callFuncChunk(fn *FuncChunk, args []interface{}, selfEnv *ivmEnv) (interface{}, error) {
if len(args) != len(fn.Params) {
return nil, m.runtimeErr(fmt.Sprintf("function '%s' expects %d argument(s), got %d", fn.Name, len(fn.Params), len(args)))
}

// Create a new environment for the function call
var parentEnv *ivmEnv
if selfEnv != nil {
parentEnv = selfEnv
} else {
// Use the current env's root as the closure env (simple lexical scoping)
parentEnv = m.env()
}
funcEnv := parentEnv.newChild()
for i, param := range fn.Params {
funcEnv.defineVar(param, args[i], false)
}

// Push current frame, start new frame.
// Remember which frame belongs to this function so we can detect if
// handleError() has swapped m.cur to a parent frame.
m.frames = append(m.frames, m.cur)
funcFrame := &callFrame{
chunk: fn.Body,
ip:    0,
stack: []interface{}{},
env:   funcEnv,
name:  fn.Name,
}
m.cur = funcFrame

// Run until RETURN or end of code
for {
frame := m.cur
if frame != funcFrame {
// m.cur was changed to a parent frame by handleError (catch handler
// found in a parent frame).  Return the sentinel; execute() will
// continue at the catch handler.
return nil, errCaughtByParent{}
}
if frame.ip >= len(frame.chunk.Code) {
// Implicit nil return at end of function; restore caller frame.
m.cur = m.frames[len(m.frames)-1]
m.frames = m.frames[:len(m.frames)-1]
return nil, nil
}

instr := frame.chunk.Code[frame.ip]
frame.ip++

result, stop, err := m.step(instr, frame.chunk)
if err != nil {
// Propagate the sentinel without calling handleError again.
if _, ok := err.(errCaughtByParent); ok {
return nil, err
}
caught, jumpErr := m.handleError(err)
if jumpErr != nil {
return nil, jumpErr
}
if caught {
// If the handler is in a parent frame, m.cur has been updated.
// The sentinel check at the top of the loop will catch this.
continue
}
// Error was not caught anywhere.  handleError has already unwound
// m.cur and m.frames; do NOT touch them again.
return nil, err
}
if stop {
// OP_RETURN: restore caller frame, return the value.
m.cur = m.frames[len(m.frames)-1]
m.frames = m.frames[:len(m.frames)-1]
return result, nil
}
}
}

func (m *Machine) callMethod(obj interface{}, methodName string, args []interface{}, callerChunk *Chunk) (interface{}, error) {
// Check if it's a struct instance
si, ok := obj.(*StructInstance)
if ok {
// Look up method in struct definition
if si.DefRef != nil {
for _, method := range si.DefRef.Methods {
if method.Name == methodName {
// Create env with struct fields accessible
structFieldEnv := m.env().newChild()
for k, v := range si.Fields {
structFieldEnv.defineVar(k, v, false)
}
res, err := m.callFuncChunk(method, args, structFieldEnv)
if err != nil {
return nil, err
}
// Update struct fields from method execution
for k := range si.Fields {
if val, exists := structFieldEnv.vars[k]; exists {
si.Fields[k] = val.value
}
}
return res, nil
}
}
}
return nil, m.runtimeErr(fmt.Sprintf("struct '%s' has no method '%s'", si.DefName, methodName))
}

// Non-struct: fall back to calling function with obj as first argument
allArgs := append([]interface{}{obj}, args...)
return m.callFunction(methodName, allArgs, callerChunk)
}

func (m *Machine) executeDefaultChunk(chunk *Chunk) (interface{}, error) {
subMachine := &Machine{builtin: m.builtin}
env := m.env().newChild()
subMachine.cur = &callFrame{
chunk: chunk,
ip:    0,
stack: []interface{}{},
env:   env,
}
return subMachine.execute(env)
}
