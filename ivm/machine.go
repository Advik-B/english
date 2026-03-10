package ivm

import (
	"bufio"
	"english/vm/types"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
)

// BuiltinFunc is the stdlib function dispatcher.
type BuiltinFunc func(name string, args []interface{}) (interface{}, error)

// ─── Value types ──────────────────────────────────────────────────────────────

// StructInstance is the ivm runtime representation of a struct instance.
// Field values are mutable through the pointer.
type StructInstance struct {
	DefName string
	DefRef  *StructDef
	Fields  map[string]interface{}
}

// ReferenceValue holds a reference to a named variable in a specific scope.
type ReferenceValue struct {
	Name string
	Env  *ivmEnv
}

// ─── Environment ──────────────────────────────────────────────────────────────

type envEntry struct {
	value    interface{}
	typeName string // declared type name ("" = inferred)
	isConst  bool
}

type ivmEnv struct {
	vars       map[string]*envEntry
	funcs      map[string]*FuncChunk
	errorTypes map[string]string // name -> parent type name ("" = root)
	structDefs map[string]*StructDef
	parent     *ivmEnv
}

func newIvmEnv() *ivmEnv {
	return &ivmEnv{
		vars:       make(map[string]*envEntry),
		funcs:      make(map[string]*FuncChunk),
		errorTypes: make(map[string]string),
		structDefs: make(map[string]*StructDef),
	}
}

func (e *ivmEnv) newChild() *ivmEnv {
	return &ivmEnv{
		vars:       make(map[string]*envEntry),
		funcs:      make(map[string]*FuncChunk),
		errorTypes: make(map[string]string),
		structDefs: make(map[string]*StructDef),
		parent:     e,
	}
}

func (e *ivmEnv) getVar(name string) (interface{}, bool) {
	if en, ok := e.vars[name]; ok {
		return en.value, true
	}
	if e.parent != nil {
		return e.parent.getVar(name)
	}
	return nil, false
}

func (e *ivmEnv) setVar(name string, value interface{}) error {
	if en, ok := e.vars[name]; ok {
		if en.isConst {
			return fmt.Errorf("TypeError: cannot reassign constant '%s'", name)
		}
		if value != nil && en.typeName != "" {
			actual := inferKindName(value)
			declared := types.Parse(en.typeName)
			actualKind := types.Infer(value)
			if declared != types.TypeNull && declared != types.TypeUnknown &&
				types.Canonical(actualKind) != types.Canonical(declared) {
				return fmt.Errorf("TypeError: cannot assign %s to variable '%s' (declared as %s)", actual, name, en.typeName)
			}
		}
		en.value = value
		return nil
	}
	if e.parent != nil {
		return e.parent.setVar(name, value)
	}
	// Auto-create (needed for internal variables)
	e.vars[name] = &envEntry{value: value}
	return nil
}

func (e *ivmEnv) defineVar(name string, value interface{}, isConst bool) error {
	if _, ok := e.vars[name]; ok {
		return fmt.Errorf("variable '%s' is already defined in this scope", name)
	}
	e.vars[name] = &envEntry{value: value, isConst: isConst}
	return nil
}

func (e *ivmEnv) defineTypedVar(name string, typeName string, value interface{}, isConst bool) error {
	if _, ok := e.vars[name]; ok {
		return fmt.Errorf("variable '%s' is already defined in this scope", name)
	}
	target := types.Parse(typeName)
	if target == types.TypeUnknown {
		return fmt.Errorf("TypeError: unknown type '%s'", typeName)
	}
	if value != nil {
		actual := types.Infer(value)
		if types.Canonical(actual) != types.Canonical(target) {
			return fmt.Errorf("TypeError: cannot initialize %s variable '%s' with %s value", typeName, name, types.Name(actual))
		}
	}
	e.vars[name] = &envEntry{value: value, typeName: typeName, isConst: isConst}
	return nil
}

func (e *ivmEnv) getFunc(name string) (*FuncChunk, bool) {
	if fn, ok := e.funcs[name]; ok {
		return fn, true
	}
	if e.parent != nil {
		return e.parent.getFunc(name)
	}
	return nil, false
}

func (e *ivmEnv) defineFunc(name string, fn *FuncChunk) {
	e.funcs[name] = fn
}

func (e *ivmEnv) getStructDef(name string) (*StructDef, bool) {
	if s, ok := e.structDefs[name]; ok {
		return s, true
	}
	if e.parent != nil {
		return e.parent.getStructDef(name)
	}
	return nil, false
}

func (e *ivmEnv) defineStructDef(name string, sd *StructDef) {
	e.structDefs[name] = sd
}

// Root walks up to the root environment.
func (e *ivmEnv) root() *ivmEnv {
	r := e
	for r.parent != nil {
		r = r.parent
	}
	return r
}

func (e *ivmEnv) defineErrorType(name, parent string) {
	e.root().errorTypes[name] = parent
}

func (e *ivmEnv) isKnownErrorType(name string) bool {
	_, ok := e.root().errorTypes[name]
	return ok
}

func (e *ivmEnv) isSubtypeOf(child, parent string) bool {
	r := e.root()
	current := child
	for current != "" {
		if current == parent {
			return true
		}
		p, ok := r.errorTypes[current]
		if !ok {
			break
		}
		current = p
	}
	return false
}

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
			return nil, err
		}
		return res, nil
	}
	return nil, m.runtimeErr(fmt.Sprintf("undefined function '%s'", name))
}

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

	// Push current frame, start new frame
	m.frames = append(m.frames, m.cur)
	m.cur = &callFrame{
		chunk: fn.Body,
		ip:    0,
		stack: []interface{}{},
		env:   funcEnv,
		name:  fn.Name,
	}

	// Run until RETURN or end of code
	for {
		frame := m.cur
		if frame.ip >= len(frame.chunk.Code) {
			// Implicit nil return at end of function
			m.cur = m.frames[len(m.frames)-1]
			m.frames = m.frames[:len(m.frames)-1]
			return nil, nil
		}

		instr := frame.chunk.Code[frame.ip]
		frame.ip++

		result, stop, err := m.step(instr, frame.chunk)
		if err != nil {
			caught, jumpErr := m.handleError(err)
			if jumpErr != nil {
				// Restore caller frame before propagating
				m.cur = m.frames[len(m.frames)-1]
				m.frames = m.frames[:len(m.frames)-1]
				return nil, jumpErr
			}
			if caught {
				continue
			}
			// Error propagates out of this function
			m.cur = m.frames[len(m.frames)-1]
			m.frames = m.frames[:len(m.frames)-1]
			return nil, err
		}
		if stop {
			// OP_RETURN: restore caller frame, return the value
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

// ─── Helpers ──────────────────────────────────────────────────────────────────

func doBinaryOp(op BinOp, left, right interface{}) (interface{}, error) {
	switch op {
	case BinAdd:
		return ivmAdd(left, right)
	case BinSub:
		return requireNumberBinary(left, right, "-", func(a, b float64) interface{} { return a - b })
	case BinMul:
		return requireNumberBinary(left, right, "*", func(a, b float64) interface{} { return a * b })
	case BinDiv:
		l, err := ivmToFloat(left, "/")
		if err != nil {
			return nil, err
		}
		r, err := ivmToFloat(right, "/")
		if err != nil {
			return nil, err
		}
		if r == 0 {
			return nil, fmt.Errorf("division by zero")
		}
		return l / r, nil
	case BinMod:
		l, err := ivmToFloat(left, "remainder")
		if err != nil {
			return nil, err
		}
		r, err := ivmToFloat(right, "remainder")
		if err != nil {
			return nil, err
		}
		if r == 0 {
			return nil, fmt.Errorf("division by zero in remainder")
		}
		return float64(int64(l) % int64(r)), nil
	case BinEq:
		return ivmStrictEquals(left, right)
	case BinNeq:
		eq, err := ivmStrictEquals(left, right)
		if err != nil {
			return nil, err
		}
		return !eq, nil
	case BinLt:
		return ivmOrderCompare(left, right, func(a, b float64) bool { return a < b })
	case BinLte:
		return ivmOrderCompare(left, right, func(a, b float64) bool { return a <= b })
	case BinGt:
		return ivmOrderCompare(left, right, func(a, b float64) bool { return a > b })
	case BinGte:
		return ivmOrderCompare(left, right, func(a, b float64) bool { return a >= b })
	}
	return nil, fmt.Errorf("unknown binary op: %d", op)
}

func doUnaryOp(op UnaryOp, val interface{}) (interface{}, error) {
	switch op {
	case UnaryNeg:
		n, err := ivmToFloat(val, "-")
		if err != nil {
			return nil, err
		}
		return -n, nil
	case UnaryNot:
		b, err := ivmToBool(val)
		if err != nil {
			return nil, err
		}
		return !b, nil
	}
	return nil, fmt.Errorf("unknown unary op: %d", op)
}

func ivmAdd(left, right interface{}) (interface{}, error) {
	switch l := left.(type) {
	case float64:
		r, ok := right.(float64)
		if !ok {
			return nil, fmt.Errorf("TypeError: '+' requires matching types")
		}
		return l + r, nil
	case string:
		r, ok := right.(string)
		if !ok {
			return nil, fmt.Errorf("TypeError: '+' requires matching types")
		}
		return l + r, nil
	case *types.ArrayValue:
		r, ok := right.(*types.ArrayValue)
		if !ok {
			return nil, fmt.Errorf("TypeError: cannot concatenate array with non-array")
		}
		if l.ElementType != r.ElementType {
			return nil, fmt.Errorf("TypeError: cannot concatenate arrays of different element types")
		}
		combined := make([]interface{}, len(l.Elements)+len(r.Elements))
		copy(combined, l.Elements)
		copy(combined[len(l.Elements):], r.Elements)
		return &types.ArrayValue{ElementType: l.ElementType, Elements: combined}, nil
	default:
		return nil, fmt.Errorf("TypeError: '+' is not defined for %s", ivmGetTypeName(left))
	}
}

func requireNumberBinary(left, right interface{}, op string, fn func(float64, float64) interface{}) (interface{}, error) {
	l, err := ivmToFloat(left, op)
	if err != nil {
		return nil, err
	}
	r, err := ivmToFloat(right, op)
	if err != nil {
		return nil, err
	}
	return fn(l, r), nil
}

func ivmToFloat(v interface{}, op string) (float64, error) {
	switch val := v.(type) {
	case float64:
		return val, nil
	case int32:
		return float64(val), nil
	case int64:
		return float64(val), nil
	case uint32:
		return float64(val), nil
	case uint64:
		return float64(val), nil
	case float32:
		return float64(val), nil
	default:
		return 0, fmt.Errorf("TypeError: '%s' requires number, got %s", op, ivmGetTypeName(v))
	}
}

func ivmToBool(v interface{}) (bool, error) {
	switch val := v.(type) {
	case bool:
		return val, nil
	case nil:
		return false, nil
	default:
		return false, fmt.Errorf("TypeError: conditions must be boolean, got %s", ivmGetTypeName(val))
	}
}

func ivmStrictEquals(left, right interface{}) (bool, error) {
	if left == nil && right == nil {
		return true, nil
	}
	if left == nil || right == nil {
		return false, nil
	}
	lk := types.Canonical(types.Infer(left))
	rk := types.Canonical(types.Infer(right))
	if lk != rk {
		return false, nil
	}
	switch l := left.(type) {
	case float64:
		if r, ok := right.(float64); ok {
			return l == r, nil
		}
	case string:
		if r, ok := right.(string); ok {
			return l == r, nil
		}
	case bool:
		if r, ok := right.(bool); ok {
			return l == r, nil
		}
	}
	return false, nil
}

func ivmOrderCompare(left, right interface{}, pred func(float64, float64) bool) (bool, error) {
	l, err := ivmToFloat(left, "comparison")
	if err != nil {
		return false, err
	}
	r, err := ivmToFloat(right, "comparison")
	if err != nil {
		return false, err
	}
	return pred(l, r), nil
}

func doIndexGet(container, index interface{}) (interface{}, error) {
	switch c := container.(type) {
	case []interface{}:
		idx, err := ivmToFloat(index, "index")
		if err != nil {
			return nil, err
		}
		i := int(idx)
		if i < 0 || i >= len(c) {
			return nil, fmt.Errorf("index %d out of range for list of length %d", i, len(c))
		}
		return c[i], nil
	case *types.ArrayValue:
		idx, err := ivmToFloat(index, "index")
		if err != nil {
			return nil, err
		}
		i := int(idx)
		if i < 0 || i >= len(c.Elements) {
			return nil, fmt.Errorf("index %d out of range for array of length %d", i, len(c.Elements))
		}
		return c.Elements[i], nil
	case string:
		idx, err := ivmToFloat(index, "index")
		if err != nil {
			return nil, err
		}
		i := int(idx)
		runes := []rune(c)
		if i < 0 || i >= len(runes) {
			return nil, fmt.Errorf("index %d out of range for string of length %d", i, len(runes))
		}
		return string(runes[i]), nil
	case *types.LookupTableValue:
		// Integer indexing into a lookup table yields the key at that position
		// (used by the for-each loop when iterating over a lookup table).
		idx, err := ivmToFloat(index, "index")
		if err != nil {
			return nil, err
		}
		return lookupTableGetByIndex(c, int(idx))
	default:
		return nil, fmt.Errorf("cannot index into %s", ivmGetTypeName(container))
	}
}

// lookupTableGetByIndex returns the key at position i (0-based) in a lookup table.
// Used by the for-each loop when iterating over a lookup table (yields keys in insertion order).
func lookupTableGetByIndex(lt *types.LookupTableValue, i int) (interface{}, error) {
	if i < 0 || i >= len(lt.KeyOrder) {
		return nil, fmt.Errorf("index %d out of range for lookup table of length %d", i, len(lt.KeyOrder))
	}
	serialKey := lt.KeyOrder[i]
	origKey, _, ok := types.DeserializeKey(serialKey)
	if !ok {
		origKey = serialKey
	}
	return origKey, nil
}

func doIndexSet(container, index, value interface{}) error {
	switch c := container.(type) {
	case []interface{}:
		idx, err := ivmToFloat(index, "index")
		if err != nil {
			return err
		}
		i := int(idx)
		if i < 0 || i >= len(c) {
			return fmt.Errorf("index %d out of range for list of length %d", i, len(c))
		}
		c[i] = value
		return nil
	case *types.ArrayValue:
		idx, err := ivmToFloat(index, "index")
		if err != nil {
			return err
		}
		i := int(idx)
		if i < 0 || i >= len(c.Elements) {
			return fmt.Errorf("index %d out of range for array of length %d", i, len(c.Elements))
		}
		c.Elements[i] = value
		return nil
	default:
		return fmt.Errorf("cannot assign to index of %s", ivmGetTypeName(container))
	}
}

func doLength(val interface{}) (float64, error) {
	switch v := val.(type) {
	case []interface{}:
		return float64(len(v)), nil
	case *types.ArrayValue:
		return float64(len(v.Elements)), nil
	case string:
		return float64(len([]rune(v))), nil
	case *types.LookupTableValue:
		return float64(len(v.KeyOrder)), nil
	default:
		return 0, fmt.Errorf("cannot get length of %s", ivmGetTypeName(val))
	}
}

func doLookupGet(table, key interface{}) (interface{}, error) {
	lt, ok := table.(*types.LookupTableValue)
	if !ok {
		return nil, fmt.Errorf("LOOKUP_GET: not a lookup table")
	}
	k, err := types.SerializeKey(key)
	if err != nil {
		return nil, err
	}
	val, exists := lt.Entries[k]
	if !exists {
		return nil, nil
	}
	return val, nil
}

func ivmGetTypeName(v interface{}) string {
	switch val := v.(type) {
	case float64:
		return "f64"
	case int32:
		return "i32"
	case int64:
		return "i64"
	case uint32:
		return "u32"
	case uint64:
		return "u64"
	case float32:
		return "f32"
	case string:
		return "text"
	case bool:
		return "boolean"
	case []interface{}:
		return "list"
	case *types.ArrayValue:
		elemTypeInfo := &types.TypeInfo{Kind: val.ElementType}
		return fmt.Sprintf("array of %s", elemTypeInfo.String())
	case *types.LookupTableValue:
		return "lookup table"
	case *types.ErrorValue:
		return "error"
	case *StructInstance:
		return val.DefName
	case *ReferenceValue:
		return "reference"
	case *FuncChunk:
		return "function"
	case nil:
		return "nothing"
	default:
		return fmt.Sprintf("%T", v)
	}
}

func inferKindName(v interface{}) string {
	return types.Name(types.Infer(v))
}

func ivmToString(v interface{}) string {
	switch val := v.(type) {
	case float64:
		if val == float64(int64(val)) && !math.IsInf(val, 0) {
			return strconv.FormatInt(int64(val), 10)
		}
		return strconv.FormatFloat(val, 'f', -1, 64)
	case int32:
		return strconv.FormatInt(int64(val), 10)
	case int64:
		return strconv.FormatInt(val, 10)
	case uint32:
		return strconv.FormatUint(uint64(val), 10)
	case uint64:
		return strconv.FormatUint(val, 10)
	case float32:
		return strconv.FormatFloat(float64(val), 'f', -1, 32)
	case string:
		return val
	case bool:
		if val {
			return "true"
		}
		return "false"
	case []interface{}:
		parts := make([]string, len(val))
		for i, elem := range val {
			parts[i] = ivmToString(elem)
		}
		return "[" + strings.Join(parts, " ") + "]"
	case *types.ArrayValue:
		parts := make([]string, len(val.Elements))
		for i, elem := range val.Elements {
			parts[i] = ivmToString(elem)
		}
		return "[" + strings.Join(parts, ", ") + "]"
	case *types.LookupTableValue:
		if len(val.KeyOrder) == 0 {
			return "{}"
		}
		parts := make([]string, 0, len(val.KeyOrder))
		for _, k := range val.KeyOrder {
			origKey, _, ok := types.DeserializeKey(k)
			keyStr := k
			if ok {
				keyStr = ivmToString(origKey)
			}
			parts = append(parts, fmt.Sprintf("%s: %s", keyStr, ivmToString(val.Entries[k])))
		}
		return "{" + strings.Join(parts, ", ") + "}"
	case *StructInstance:
		return fmt.Sprintf("<%s instance>", val.DefName)
	case *types.ErrorValue:
		return fmt.Sprintf("<error: %s>", val.Message)
	case *ReferenceValue:
		return fmt.Sprintf("<ref: %s>", val.Name)
	case *FuncChunk:
		return fmt.Sprintf("<function %s>", val.Name)
	case nil:
		return "nothing"
	default:
		return fmt.Sprintf("%v", v)
	}
}

func deepCopyValue(val interface{}) interface{} {
	switch v := val.(type) {
	case []interface{}:
		copied := make([]interface{}, len(v))
		for i, elem := range v {
			copied[i] = deepCopyValue(elem)
		}
		return copied
	case *types.ArrayValue:
		elems := make([]interface{}, len(v.Elements))
		for i, elem := range v.Elements {
			elems[i] = deepCopyValue(elem)
		}
		return &types.ArrayValue{ElementType: v.ElementType, Elements: elems}
	case *StructInstance:
		copied := &StructInstance{
			DefName: v.DefName,
			DefRef:  v.DefRef,
			Fields:  make(map[string]interface{}),
		}
		for k, fv := range v.Fields {
			copied.Fields[k] = deepCopyValue(fv)
		}
		return copied
	default:
		return val
	}
}

func typeDefault(typeName string) interface{} {
	switch typeName {
	case "number", "f64", "f32", "i32", "i64", "u32", "u64":
		return float64(0)
	case "text":
		return ""
	case "boolean":
		return false
	case "list":
		return []interface{}{}
	default:
		return nil
	}
}
