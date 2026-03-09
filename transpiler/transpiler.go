// Package transpiler converts an English language AST to human-readable Python code.
package transpiler

import (
	"english/ast"
	"fmt"
	"math"
	"strings"
)

// Transpiler converts an English AST to Python source code.
type Transpiler struct {
	indent int
	buf    strings.Builder
	// Python module imports required by the generated code
	needsMath   bool
	needsCopy   bool
	needsRandom bool
	// Helper functions that must be emitted at the top of the output
	helpers map[string]bool
	// methodFields holds field names of the struct currently being transpiled,
	// so that bare field-name identifiers inside methods become self.<field>.
	methodFields map[string]bool
}

// NewTranspiler creates a new Transpiler instance.
func NewTranspiler() *Transpiler {
	return &Transpiler{
		helpers: make(map[string]bool),
	}
}

// Transpile converts a parsed English program AST to a Python source string.
func (t *Transpiler) Transpile(program *ast.Program) string {
// First pass: scan the program to collect import/helper requirements.
t.scanProgram(program)

// Second pass: generate the body code.
var bodyBuf strings.Builder
savedBuf := t.buf
t.buf = bodyBuf
for _, stmt := range program.Statements {
t.transpileStatement(stmt)
}
body := t.buf.String()
t.buf = savedBuf

// Build the final file: banner + imports + helpers + body.
var out strings.Builder
out.WriteString("# Transpiled from English language source\n")

if t.needsMath {
out.WriteString("import math\n")
}
if t.needsCopy {
out.WriteString("import copy\n")
}
if t.needsRandom {
out.WriteString("import random\n")
}
if t.needsMath || t.needsCopy || t.needsRandom {
out.WriteString("\n")
}

// Emit helper functions in a deterministic order.
helperOrder := []string{
"_table_remove",
"_flatten",
"_read_file",
"_write_file",
"_is_nan",
"_is_infinite",
}
emittedHelper := false
for _, name := range helperOrder {
if t.helpers[name] {
out.WriteString(helperDefs[name])
out.WriteString("\n")
emittedHelper = true
}
}
if emittedHelper {
out.WriteString("\n")
}

out.WriteString(body)
return out.String()
}

// helperDefs holds small Python helper functions for stdlib calls that have no
// single-expression equivalent.
var helperDefs = map[string]string{
"_table_remove": `def _table_remove(d, k):
    result = dict(d)
    result.pop(k, None)
    return result`,
"_flatten": `def _flatten(lst):
    return [item for sublist in lst for item in sublist]`,
"_read_file": `def _read_file(path):
    with open(path, "r") as f:
        return f.read()`,
"_write_file": `def _write_file(path, content):
    with open(path, "w") as f:
        f.write(str(content))`,
"_is_nan": `def _is_nan(x):
    try:
        return math.isnan(float(x))
    except (TypeError, ValueError):
        return True`,
"_is_infinite": `def _is_infinite(x):
    try:
        return math.isinf(float(x))
    except (TypeError, ValueError):
        return False`,
}

// ─── Scanning pass ────────────────────────────────────────────────────────────

func (t *Transpiler) scanProgram(program *ast.Program) {
for _, stmt := range program.Statements {
t.scanStmt(stmt)
}
}

func (t *Transpiler) scanStmt(stmt ast.Statement) {
switch s := stmt.(type) {
case *ast.FunctionDecl:
for _, c := range s.Body {
t.scanStmt(c)
}
case *ast.IfStatement:
for _, c := range s.Then {
t.scanStmt(c)
}
for _, elif := range s.ElseIf {
for _, c := range elif.Body {
t.scanStmt(c)
}
}
for _, c := range s.Else {
t.scanStmt(c)
}
case *ast.WhileLoop:
for _, c := range s.Body {
t.scanStmt(c)
}
case *ast.ForLoop:
for _, c := range s.Body {
t.scanStmt(c)
}
case *ast.ForEachLoop:
for _, c := range s.Body {
t.scanStmt(c)
}
case *ast.TryStatement:
for _, c := range s.TryBody {
t.scanStmt(c)
}
for _, c := range s.ErrorBody {
t.scanStmt(c)
}
for _, c := range s.FinallyBody {
t.scanStmt(c)
}
case *ast.StructDecl:
for _, m := range s.Methods {
for _, c := range m.Body {
t.scanStmt(c)
}
}
case *ast.VariableDecl:
t.scanExpr(s.Value)
case *ast.TypedVariableDecl:
t.scanExpr(s.Value)
case *ast.Assignment:
t.scanExpr(s.Value)
case *ast.IndexAssignment:
t.scanExpr(s.Index)
t.scanExpr(s.Value)
case *ast.FieldAssignment:
t.scanExpr(s.Value)
case *ast.LookupKeyAssignment:
t.scanExpr(s.Key)
t.scanExpr(s.Value)
case *ast.OutputStatement:
for _, v := range s.Values {
t.scanExpr(v)
}
case *ast.ReturnStatement:
t.scanExpr(s.Value)
case *ast.CallStatement:
if s.FunctionCall != nil {
t.scanFuncCall(s.FunctionCall.Name)
for _, a := range s.FunctionCall.Arguments {
t.scanExpr(a)
}
}
if s.MethodCall != nil {
t.scanExpr(s.MethodCall.Object)
}
case *ast.RaiseStatement:
t.scanExpr(s.Message)
}
}

func (t *Transpiler) scanExpr(expr ast.Expression) {
if expr == nil {
return
}
switch e := expr.(type) {
case *ast.FunctionCall:
t.scanFuncCall(e.Name)
for _, a := range e.Arguments {
t.scanExpr(a)
}
case *ast.MethodCall:
t.scanExpr(e.Object)
for _, a := range e.Arguments {
t.scanExpr(a)
}
case *ast.BinaryExpression:
t.scanExpr(e.Left)
t.scanExpr(e.Right)
case *ast.UnaryExpression:
t.scanExpr(e.Right)
case *ast.CopyExpression:
t.needsCopy = true
t.scanExpr(e.Value)
case *ast.IndexExpression:
t.scanExpr(e.List)
t.scanExpr(e.Index)
case *ast.LengthExpression:
t.scanExpr(e.List)
case *ast.FieldAccess:
t.scanExpr(e.Object)
case *ast.LookupKeyAccess:
t.scanExpr(e.Table)
t.scanExpr(e.Key)
case *ast.HasExpression:
t.scanExpr(e.Table)
t.scanExpr(e.Key)
case *ast.NilCheckExpression:
t.scanExpr(e.Value)
case *ast.CastExpression:
t.scanExpr(e.Value)
case *ast.TypeExpression:
t.scanExpr(e.Value)
case *ast.AskExpression:
t.scanExpr(e.Prompt)
case *ast.StructInstantiation:
for _, v := range e.FieldValues {
t.scanExpr(v)
}
case *ast.ErrorTypeCheckExpression:
t.scanExpr(e.Value)
case *ast.ListLiteral:
for _, el := range e.Elements {
t.scanExpr(el)
}
case *ast.ArrayLiteral:
for _, el := range e.Elements {
t.scanExpr(el)
}
}
}

func (t *Transpiler) scanFuncCall(name string) {
switch name {
case "sqrt", "pow", "floor", "ceil", "sin", "cos", "tan", "log", "log10", "log2", "exp":
t.needsMath = true
case "_is_nan", "is_nan", "_is_infinite", "is_infinite":
t.needsMath = true
t.helpers["_is_nan"] = true
t.helpers["_is_infinite"] = true
case "random", "random_between":
t.needsRandom = true
case "table_remove":
t.helpers["_table_remove"] = true
case "flatten":
t.helpers["_flatten"] = true
case "read_file":
t.helpers["_read_file"] = true
case "write_file":
t.helpers["_write_file"] = true
}
}

// ─── Indentation helpers ──────────────────────────────────────────────────────

func (t *Transpiler) write(s string) {
t.buf.WriteString(s)
}

func (t *Transpiler) writeLine(s string) {
t.buf.WriteString(strings.Repeat("    ", t.indent))
t.buf.WriteString(s)
t.buf.WriteString("\n")
}

// ─── Statements ───────────────────────────────────────────────────────────────

func (t *Transpiler) transpileStatement(stmt ast.Statement) {
switch s := stmt.(type) {
case *ast.ImportStatement:
t.transpileImport(s)
case *ast.VariableDecl:
t.transpileVariableDecl(s)
case *ast.TypedVariableDecl:
t.transpileTypedVariableDecl(s)
case *ast.Assignment:
t.transpileAssignment(s)
case *ast.IndexAssignment:
t.transpileIndexAssignment(s)
case *ast.FieldAssignment:
t.transpileFieldAssignment(s)
case *ast.LookupKeyAssignment:
t.transpileLookupKeyAssignment(s)
case *ast.FunctionDecl:
t.transpileFunctionDecl(s)
case *ast.CallStatement:
t.transpileCallStatement(s)
case *ast.OutputStatement:
t.transpileOutput(s)
case *ast.ReturnStatement:
t.transpileReturn(s)
case *ast.IfStatement:
t.transpileIf(s)
case *ast.WhileLoop:
t.transpileWhile(s)
case *ast.ForLoop:
t.transpileForLoop(s)
case *ast.ForEachLoop:
t.transpileForEach(s)
case *ast.ToggleStatement:
t.writeLine(fmt.Sprintf("%s = not %s", s.Name, s.Name))
case *ast.BreakStatement:
t.writeLine("break")
case *ast.ContinueStatement:
t.writeLine("continue")
case *ast.TryStatement:
t.transpileTry(s)
case *ast.RaiseStatement:
t.transpileRaise(s)
case *ast.ErrorTypeDecl:
t.transpileErrorTypeDecl(s)
case *ast.SwapStatement:
t.writeLine(fmt.Sprintf("%s, %s = %s, %s", s.Name1, s.Name2, s.Name2, s.Name1))
case *ast.StructDecl:
t.transpileStructDecl(s)
default:
t.writeLine(fmt.Sprintf("# unsupported statement: %T", stmt))
}
}

func (t *Transpiler) transpileImport(s *ast.ImportStatement) {
if len(s.Items) > 0 {
t.writeLine(fmt.Sprintf("# from %q import %s", s.Path, strings.Join(s.Items, ", ")))
} else {
t.writeLine(fmt.Sprintf("# import %q", s.Path))
}
}

func (t *Transpiler) transpileVariableDecl(s *ast.VariableDecl) {
val := t.transpileExpr(s.Value)
if s.IsConstant {
t.writeLine(fmt.Sprintf("%s = %s  # constant", s.Name, val))
} else {
t.writeLine(fmt.Sprintf("%s = %s", s.Name, val))
}
}

func (t *Transpiler) transpileTypedVariableDecl(s *ast.TypedVariableDecl) {
val := t.transpileExpr(s.Value)
typeName := mapTypeName(s.TypeName)
if s.IsConstant {
t.writeLine(fmt.Sprintf("%s: %s = %s  # constant", s.Name, typeName, val))
} else {
t.writeLine(fmt.Sprintf("%s: %s = %s", s.Name, typeName, val))
}
}

func (t *Transpiler) transpileAssignment(s *ast.Assignment) {
val := t.transpileExpr(s.Value)
t.writeLine(fmt.Sprintf("%s = %s", s.Name, val))
}

func (t *Transpiler) transpileIndexAssignment(s *ast.IndexAssignment) {
idx := t.transpileExpr(s.Index)
val := t.transpileExpr(s.Value)
t.writeLine(fmt.Sprintf("%s[int(%s)] = %s", s.ListName, idx, val))
}

func (t *Transpiler) transpileFieldAssignment(s *ast.FieldAssignment) {
val := t.transpileExpr(s.Value)
t.writeLine(fmt.Sprintf("%s.%s = %s", s.ObjectName, s.Field, val))
}

func (t *Transpiler) transpileLookupKeyAssignment(s *ast.LookupKeyAssignment) {
key := t.transpileExpr(s.Key)
val := t.transpileExpr(s.Value)
t.writeLine(fmt.Sprintf("%s[%s] = %s", s.TableName, key, val))
}

func (t *Transpiler) transpileFunctionDecl(s *ast.FunctionDecl) {
params := strings.Join(s.Parameters, ", ")
t.writeLine(fmt.Sprintf("def %s(%s):", s.Name, params))
t.indent++
if len(s.Body) == 0 {
t.writeLine("pass")
} else {
for _, stmt := range s.Body {
t.transpileStatement(stmt)
}
}
t.indent--
t.write("\n")
}

func (t *Transpiler) transpileCallStatement(s *ast.CallStatement) {
if s.FunctionCall != nil {
t.writeLine(t.transpileFuncCallExpr(s.FunctionCall))
} else if s.MethodCall != nil {
t.writeLine(t.transpileMethodCallExpr(s.MethodCall))
}
}

func (t *Transpiler) transpileOutput(s *ast.OutputStatement) {
parts := make([]string, len(s.Values))
for i, v := range s.Values {
parts[i] = t.transpileExpr(v)
}
args := strings.Join(parts, ", ")
if s.Newline {
t.writeLine(fmt.Sprintf("print(%s)", args))
} else {
t.writeLine(fmt.Sprintf("print(%s, end=\"\")", args))
}
}

func (t *Transpiler) transpileReturn(s *ast.ReturnStatement) {
if s.Value == nil {
t.writeLine("return")
} else {
t.writeLine(fmt.Sprintf("return %s", t.transpileExpr(s.Value)))
}
}

func (t *Transpiler) transpileIf(s *ast.IfStatement) {
t.writeLine(fmt.Sprintf("if %s:", t.transpileExpr(s.Condition)))
t.indent++
t.transpileBody(s.Then)
t.indent--
for _, elif := range s.ElseIf {
t.writeLine(fmt.Sprintf("elif %s:", t.transpileExpr(elif.Condition)))
t.indent++
t.transpileBody(elif.Body)
t.indent--
}
if len(s.Else) > 0 {
t.writeLine("else:")
t.indent++
t.transpileBody(s.Else)
t.indent--
}
}

func (t *Transpiler) transpileWhile(s *ast.WhileLoop) {
t.writeLine(fmt.Sprintf("while %s:", t.transpileExpr(s.Condition)))
t.indent++
t.transpileBody(s.Body)
t.indent--
}

func (t *Transpiler) transpileForLoop(s *ast.ForLoop) {
t.writeLine(fmt.Sprintf("for _ in range(int(%s)):", t.transpileExpr(s.Count)))
t.indent++
t.transpileBody(s.Body)
t.indent--
}

func (t *Transpiler) transpileForEach(s *ast.ForEachLoop) {
t.writeLine(fmt.Sprintf("for %s in %s:", s.Item, t.transpileExpr(s.List)))
t.indent++
t.transpileBody(s.Body)
t.indent--
}

func (t *Transpiler) transpileTry(s *ast.TryStatement) {
t.writeLine("try:")
t.indent++
t.transpileBody(s.TryBody)
t.indent--

if len(s.ErrorBody) > 0 {
var excLine string
if s.ErrorType != "" {
if s.ErrorVar != "" {
excLine = fmt.Sprintf("except %s as %s:", s.ErrorType, s.ErrorVar)
} else {
excLine = fmt.Sprintf("except %s:", s.ErrorType)
}
} else {
if s.ErrorVar != "" {
excLine = fmt.Sprintf("except Exception as %s:", s.ErrorVar)
} else {
excLine = "except Exception:"
}
}
t.writeLine(excLine)
t.indent++
t.transpileBody(s.ErrorBody)
t.indent--
}

if len(s.FinallyBody) > 0 {
t.writeLine("finally:")
t.indent++
t.transpileBody(s.FinallyBody)
t.indent--
}
}

func (t *Transpiler) transpileRaise(s *ast.RaiseStatement) {
msg := t.transpileExpr(s.Message)
if s.ErrorType != "" {
t.writeLine(fmt.Sprintf("raise %s(%s)", s.ErrorType, msg))
} else {
t.writeLine(fmt.Sprintf("raise Exception(%s)", msg))
}
}

func (t *Transpiler) transpileErrorTypeDecl(s *ast.ErrorTypeDecl) {
parent := "Exception"
if s.ParentType != "" {
parent = s.ParentType
}
t.writeLine(fmt.Sprintf("class %s(%s): pass", s.Name, parent))
t.write("\n")
}

func (t *Transpiler) transpileStructDecl(s *ast.StructDecl) {
	t.writeLine(fmt.Sprintf("class %s:", s.Name))
	t.indent++
	if len(s.Fields) > 0 {
		params := make([]string, 0, len(s.Fields)+1)
		params = append(params, "self")
		for _, field := range s.Fields {
			if field.DefaultValue != nil {
				defVal := t.transpileExpr(field.DefaultValue)
				params = append(params, fmt.Sprintf("%s=%s", field.Name, defVal))
			} else {
				params = append(params, field.Name)
			}
		}
		t.writeLine(fmt.Sprintf("def __init__(%s):", strings.Join(params, ", ")))
		t.indent++
		for _, field := range s.Fields {
			t.writeLine(fmt.Sprintf("self.%s = %s", field.Name, field.Name))
		}
		t.indent--
	} else {
		t.writeLine("pass")
	}

	// Build the set of field names so method bodies can prefix them with "self.".
	fields := make(map[string]bool, len(s.Fields))
	for _, field := range s.Fields {
		fields[field.Name] = true
	}
	savedFields := t.methodFields
	t.methodFields = fields
	defer func() { t.methodFields = savedFields }()

	for _, method := range s.Methods {
		t.write("\n")
		params := make([]string, 0, len(method.Parameters)+1)
		params = append(params, "self")
		params = append(params, method.Parameters...)
		t.writeLine(fmt.Sprintf("def %s(%s):", method.Name, strings.Join(params, ", ")))
		t.indent++
		if len(method.Body) == 0 {
			t.writeLine("pass")
		} else {
			for _, stmt := range method.Body {
				t.transpileStatement(stmt)
			}
		}
		t.indent--
	}
	t.indent--
	t.write("\n")
}

// transpileBody writes a list of statements; emits "pass" for empty bodies.
func (t *Transpiler) transpileBody(stmts []ast.Statement) {
if len(stmts) == 0 {
t.writeLine("pass")
return
}
for _, stmt := range stmts {
t.transpileStatement(stmt)
}
}

// ─── Expressions ─────────────────────────────────────────────────────────────

func (t *Transpiler) transpileExpr(expr ast.Expression) string {
if expr == nil {
return "None"
}
switch e := expr.(type) {
case *ast.NumberLiteral:
return formatNumber(e.Value)
case *ast.StringLiteral:
return fmt.Sprintf("%q", e.Value)
case *ast.BooleanLiteral:
if e.Value {
return "True"
}
return "False"
case *ast.NothingLiteral:
return "None"
	case *ast.Identifier:
		if t.methodFields[e.Name] {
			return "self." + e.Name
		}
		return e.Name
case *ast.ListLiteral:
return t.transpileListLit(e.Elements)
case *ast.ArrayLiteral:
return t.transpileListLit(e.Elements)
case *ast.LookupTableLiteral:
return "{}"
case *ast.BinaryExpression:
return t.transpileBinaryExpr(e)
case *ast.UnaryExpression:
return t.transpileUnaryExpr(e)
case *ast.FunctionCall:
return t.transpileFuncCallExpr(e)
case *ast.MethodCall:
return t.transpileMethodCallExpr(e)
case *ast.IndexExpression:
return fmt.Sprintf("%s[int(%s)]", t.transpileExpr(e.List), t.transpileExpr(e.Index))
case *ast.LengthExpression:
return fmt.Sprintf("len(%s)", t.transpileExpr(e.List))
case *ast.FieldAccess:
return fmt.Sprintf("%s.%s", t.transpileExpr(e.Object), e.Field)
case *ast.StructInstantiation:
return t.transpileStructInst(e)
case *ast.TypeExpression:
return fmt.Sprintf("type(%s).__name__", t.transpileExpr(e.Value))
case *ast.CastExpression:
return t.transpileCast(e)
case *ast.ReferenceExpression:
return e.Name
case *ast.CopyExpression:
return fmt.Sprintf("copy.copy(%s)", t.transpileExpr(e.Value))
case *ast.LocationExpression:
return fmt.Sprintf("hex(id(%s))", e.Name)
case *ast.AskExpression:
if e.Prompt == nil {
return "input()"
}
return fmt.Sprintf("input(%s)", t.transpileExpr(e.Prompt))
case *ast.LookupKeyAccess:
return fmt.Sprintf("%s[%s]", t.transpileExpr(e.Table), t.transpileExpr(e.Key))
case *ast.HasExpression:
return fmt.Sprintf("%s in %s", t.transpileExpr(e.Key), t.transpileExpr(e.Table))
case *ast.NilCheckExpression:
inner := t.transpileExpr(e.Value)
if e.IsSomethingCheck {
return fmt.Sprintf("%s is not None", inner)
}
return fmt.Sprintf("%s is None", inner)
case *ast.ErrorTypeCheckExpression:
return fmt.Sprintf("isinstance(%s, %s)", t.transpileExpr(e.Value), e.TypeName)
default:
return fmt.Sprintf("None  # unsupported expression: %T", expr)
}
}

func (t *Transpiler) transpileListLit(elements []ast.Expression) string {
parts := make([]string, len(elements))
for i, el := range elements {
parts[i] = t.transpileExpr(el)
}
return "[" + strings.Join(parts, ", ") + "]"
}

func (t *Transpiler) transpileBinaryExpr(e *ast.BinaryExpression) string {
left := t.transpileExpr(e.Left)
right := t.transpileExpr(e.Right)
op := mapOperator(e.Operator)

// Parenthesise nested binary expressions to make precedence explicit.
if _, ok := e.Left.(*ast.BinaryExpression); ok {
left = "(" + left + ")"
}
if _, ok := e.Right.(*ast.BinaryExpression); ok {
right = "(" + right + ")"
}
return fmt.Sprintf("%s %s %s", left, op, right)
}

func (t *Transpiler) transpileUnaryExpr(e *ast.UnaryExpression) string {
right := t.transpileExpr(e.Right)
switch e.Operator {
case "-":
return fmt.Sprintf("-%s", right)
case "not":
return fmt.Sprintf("not %s", right)
default:
return fmt.Sprintf("%s%s", e.Operator, right)
}
}

// transpileFuncCallExpr translates an English stdlib/user function call to Python.
// Where possible the call is inlined as a Python expression (e.g. method calls,
// built-in function calls). For the few helpers that have no direct equivalent,
// a small helper function is emitted at the top of the file.
func (t *Transpiler) transpileFuncCallExpr(e *ast.FunctionCall) string {
args := make([]string, len(e.Arguments))
for i, a := range e.Arguments {
args[i] = t.transpileExpr(a)
}
a := func(i int) string {
if i < len(args) {
return args[i]
}
return "None"
}

switch e.Name {
// ── Math ──────────────────────────────────────────────────────────────────
case "sqrt":
return fmt.Sprintf("math.sqrt(%s)", a(0))
case "pow":
return fmt.Sprintf("math.pow(%s, %s)", a(0), a(1))
case "abs":
return fmt.Sprintf("abs(%s)", a(0))
case "floor":
return fmt.Sprintf("math.floor(%s)", a(0))
case "ceil":
return fmt.Sprintf("math.ceil(%s)", a(0))
case "round":
return fmt.Sprintf("round(%s)", a(0))
case "min":
return fmt.Sprintf("min(%s, %s)", a(0), a(1))
case "max":
return fmt.Sprintf("max(%s, %s)", a(0), a(1))
case "sin":
return fmt.Sprintf("math.sin(%s)", a(0))
case "cos":
return fmt.Sprintf("math.cos(%s)", a(0))
case "tan":
return fmt.Sprintf("math.tan(%s)", a(0))
case "log":
return fmt.Sprintf("math.log(%s)", a(0))
case "log10":
return fmt.Sprintf("math.log10(%s)", a(0))
case "log2":
return fmt.Sprintf("math.log2(%s)", a(0))
case "exp":
return fmt.Sprintf("math.exp(%s)", a(0))
case "random":
return "random.random()"
case "random_between":
return fmt.Sprintf("random.uniform(%s, %s)", a(0), a(1))
case "is_nan":
return fmt.Sprintf("_is_nan(%s)", a(0))
case "is_infinite":
return fmt.Sprintf("_is_infinite(%s)", a(0))

// ── String ────────────────────────────────────────────────────────────────
case "uppercase":
return fmt.Sprintf("%s.upper()", a(0))
case "lowercase":
return fmt.Sprintf("%s.lower()", a(0))
case "casefold":
return fmt.Sprintf("%s.casefold()", a(0))
case "split":
return fmt.Sprintf("%s.split(%s)", a(0), a(1))
case "join":
// join(list, sep) → sep.join(list)
return fmt.Sprintf("%s.join(%s)", a(1), a(0))
case "trim":
return fmt.Sprintf("%s.strip()", a(0))
case "replace":
return fmt.Sprintf("%s.replace(%s, %s)", a(0), a(1), a(2))
case "contains":
return fmt.Sprintf("(%s in %s)", a(1), a(0))
case "starts_with":
return fmt.Sprintf("%s.startswith(%s)", a(0), a(1))
case "ends_with":
return fmt.Sprintf("%s.endswith(%s)", a(0), a(1))
case "index_of":
return fmt.Sprintf("%s.index(%s)", a(0), a(1))
case "substring":
// substring(s, start, length) → s[start : start+length]
return fmt.Sprintf("%s[int(%s):int(%s)+int(%s)]", a(0), a(1), a(1), a(2))
case "str_repeat":
return fmt.Sprintf("%s * int(%s)", a(0), a(1))
case "count_occurrences":
return fmt.Sprintf("%s.count(%s)", a(0), a(1))
case "pad_left":
if len(args) > 2 {
return fmt.Sprintf("%s.rjust(int(%s), %s)", a(0), a(1), a(2))
}
return fmt.Sprintf("%s.rjust(int(%s))", a(0), a(1))
case "pad_right":
if len(args) > 2 {
return fmt.Sprintf("%s.ljust(int(%s), %s)", a(0), a(1), a(2))
}
return fmt.Sprintf("%s.ljust(int(%s))", a(0), a(1))
case "to_number":
return fmt.Sprintf("float(%s)", a(0))
case "to_string":
return fmt.Sprintf("str(%s)", a(0))
case "is_empty":
	return fmt.Sprintf("(len(%s) == 0)", a(0))

	// ── List ──────────────────────────────────────────────────────────────────
	case "count":
		return fmt.Sprintf("len(%s)", a(0))
	case "first":
		return fmt.Sprintf("%s[0]", a(0))
	case "last":
		return fmt.Sprintf("%s[-1]", a(0))
	case "sort":
		return fmt.Sprintf("sorted(%s)", a(0))
	case "reverse":
		return fmt.Sprintf("list(reversed(%s))", a(0))
	case "append":
		// append(list, item) → list + [item]
		return fmt.Sprintf("%s + [%s]", a(0), a(1))
	case "pop":
		return fmt.Sprintf("%s[:-1]", a(0))
	case "flatten":
		return fmt.Sprintf("_flatten(%s)", a(0))

	// ── Lookup table ──────────────────────────────────────────────────────────
	case "keys":
		return fmt.Sprintf("list(%s.keys())", a(0))
	case "values":
		return fmt.Sprintf("list(%s.values())", a(0))
	case "table_remove":
		return fmt.Sprintf("_table_remove(%s, %s)", a(0), a(1))

	// ── I/O ───────────────────────────────────────────────────────────────────
	case "ask":
		return fmt.Sprintf("input(%s)", a(0))
	case "read_file":
		return fmt.Sprintf("_read_file(%s)", a(0))
	case "write_file":
		return fmt.Sprintf("_write_file(%s, %s)", a(0), a(1))
	}

	// Unknown / user-defined function — emit a direct call.
	return fmt.Sprintf("%s(%s)", e.Name, strings.Join(args, ", "))
}

func (t *Transpiler) transpileMethodCallExpr(e *ast.MethodCall) string {
obj := t.transpileExpr(e.Object)
args := make([]string, len(e.Arguments))
for i, a := range e.Arguments {
args[i] = t.transpileExpr(a)
}
return fmt.Sprintf("%s.%s(%s)", obj, e.MethodName, strings.Join(args, ", "))
}

func (t *Transpiler) transpileCast(e *ast.CastExpression) string {
inner := t.transpileExpr(e.Value)
switch strings.ToLower(e.TypeName) {
case "number", "float":
return fmt.Sprintf("float(%s)", inner)
case "integer", "int":
return fmt.Sprintf("int(%s)", inner)
case "text", "string", "str":
return fmt.Sprintf("str(%s)", inner)
case "boolean", "bool":
return fmt.Sprintf("bool(%s)", inner)
default:
return fmt.Sprintf("%s(%s)", e.TypeName, inner)
}
}

func (t *Transpiler) transpileStructInst(e *ast.StructInstantiation) string {
args := make([]string, 0, len(e.FieldOrder))
for _, name := range e.FieldOrder {
args = append(args, fmt.Sprintf("%s=%s", name, t.transpileExpr(e.FieldValues[name])))
}
return fmt.Sprintf("%s(%s)", e.StructName, strings.Join(args, ", "))
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

// formatNumber renders a float64 as a compact Python numeric literal.
func formatNumber(v float64) string {
if math.IsInf(v, 1) {
return "float('inf')"
}
if math.IsInf(v, -1) {
return "float('-inf')"
}
if math.IsNaN(v) {
return "float('nan')"
}
if v == math.Trunc(v) && math.Abs(v) < 1e15 {
return fmt.Sprintf("%d", int64(v))
}
return fmt.Sprintf("%g", v)
}

// mapOperator converts an English operator string to the Python equivalent.
func mapOperator(op string) string {
switch op {
case "+":
return "+"
case "-":
return "-"
case "*":
return "*"
case "/":
return "/"
case "%", "remainder":
return "%"
case "is equal to", "==":
return "=="
case "is not equal to", "!=":
return "!="
case "is less than", "<":
return "<"
case "is greater than", ">":
return ">"
case "is less than or equal to", "<=":
return "<="
case "is greater than or equal to", ">=":
return ">="
case "and":
return "and"
case "or":
return "or"
default:
return op
}
}

// mapTypeName converts an English type name to a Python type annotation string.
func mapTypeName(name string) string {
switch strings.ToLower(name) {
case "number", "float":
return "float"
case "integer", "int":
return "int"
case "text", "string":
return "str"
case "boolean", "bool":
return "bool"
case "list", "array":
return "list"
default:
return name
}
}
