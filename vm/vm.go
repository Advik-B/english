// Package vm provides the virtual machine (evaluator) for the English programming language.
package vm

import (
	"english/ast"
	"fmt"
	"strconv"
	"strings"
)

// Value represents a runtime value in the interpreter
type Value interface{}

// FunctionValue represents a user-defined function
type FunctionValue struct {
	Name       string
	Parameters []string
	Body       []ast.Statement
	Closure    *Environment
}

// ReturnValue is used to implement return statements
type ReturnValue struct {
	Value Value
}

// BreakValue is used to implement break statements
type BreakValue struct{}

// RuntimeError represents an error during execution
type RuntimeError struct {
	Message   string
	CallStack []string
}

func (e *RuntimeError) Error() string {
	result := fmt.Sprintf("Runtime Error: %s\n", e.Message)
	if len(e.CallStack) > 0 {
		result += "\nCall Stack (most recent first):\n"
		for i, frame := range e.CallStack {
			result += fmt.Sprintf("  %d. %s\n", i+1, frame)
		}
	}
	return result
}

// Environment represents a scope for variables and functions
type Environment struct {
	variables map[string]Value
	constants map[string]bool
	functions map[string]*FunctionValue
	parent    *Environment
}

// NewEnvironment creates a new environment
func NewEnvironment() *Environment {
	return &Environment{
		variables: make(map[string]Value),
		constants: make(map[string]bool),
		functions: make(map[string]*FunctionValue),
	}
}

// NewChild creates a child environment
func (e *Environment) NewChild() *Environment {
	return &Environment{
		variables: make(map[string]Value),
		constants: make(map[string]bool),
		functions: make(map[string]*FunctionValue),
		parent:    e,
	}
}

// Get retrieves a variable from the environment
func (e *Environment) Get(name string) (Value, bool) {
	if val, ok := e.variables[name]; ok {
		return val, ok
	}
	if e.parent != nil {
		return e.parent.Get(name)
	}
	return nil, false
}

// Set assigns a value to a variable
func (e *Environment) Set(name string, value Value) error {
	if e.constants[name] {
		return fmt.Errorf("cannot reassign constant '%s'\n  Hint: Constants are declared with 'to be always' or 'to always be'", name)
	}
	if _, exists := e.variables[name]; exists {
		e.variables[name] = value
		return nil
	}
	// Look up the scope chain for the variable
	if e.parent != nil {
		// Check if variable exists anywhere in parent chain
		parentEnv := e.parent
		for parentEnv != nil {
			if _, exists := parentEnv.variables[name]; exists {
				return e.parent.Set(name, value)
			}
			parentEnv = parentEnv.parent
		}
	}
	e.variables[name] = value
	return nil
}

// Define declares a new variable
func (e *Environment) Define(name string, value Value, isConstant bool) error {
	if _, exists := e.variables[name]; exists {
		return fmt.Errorf("variable %s already defined", name)
	}
	e.variables[name] = value
	e.constants[name] = isConstant
	return nil
}

// GetFunction retrieves a function from the environment
func (e *Environment) GetFunction(name string) (*FunctionValue, bool) {
	if fn, ok := e.functions[name]; ok {
		return fn, ok
	}
	if e.parent != nil {
		return e.parent.GetFunction(name)
	}
	return nil, false
}

// DefineFunction declares a new function
func (e *Environment) DefineFunction(name string, fn *FunctionValue) {
	e.functions[name] = fn
}

// GetAllVariables returns a copy of all variables in the current scope
func (e *Environment) GetAllVariables() map[string]Value {
	result := make(map[string]Value)
	for k, v := range e.variables {
		result[k] = v
	}
	return result
}

// GetAllFunctions returns a copy of all functions in the current scope
func (e *Environment) GetAllFunctions() map[string]*FunctionValue {
	result := make(map[string]*FunctionValue)
	for k, v := range e.functions {
		result[k] = v
	}
	return result
}

// IsConstant returns whether a variable is a constant
func (e *Environment) IsConstant(name string) bool {
	return e.constants[name]
}

// Evaluator executes the AST
type Evaluator struct {
	env       *Environment
	callStack []string
}

// NewEvaluator creates a new evaluator with the given environment
func NewEvaluator(env *Environment) *Evaluator {
	return &Evaluator{
		env:       env,
		callStack: []string{"<main>"},
	}
}

func (ev *Evaluator) runtimeError(message string) error {
	return &RuntimeError{
		Message:   message,
		CallStack: append([]string{}, ev.callStack...),
	}
}

// Eval evaluates an AST node
func (ev *Evaluator) Eval(node interface{}) (Value, error) {
	switch node := node.(type) {
	case *ast.Program:
		return ev.evalProgram(node)
	case *ast.VariableDecl:
		return ev.evalVariableDecl(node)
	case *ast.Assignment:
		return ev.evalAssignment(node)
	case *ast.IndexAssignment:
		return ev.evalIndexAssignment(node)
	case *ast.FunctionDecl:
		return ev.evalFunctionDecl(node)
	case *ast.CallStatement:
		return ev.evalCallStatement(node)
	case *ast.OutputStatement:
		return ev.evalOutput(node)
	case *ast.ReturnStatement:
		return ev.evalReturn(node)
	case *ast.IfStatement:
		return ev.evalIfStatement(node)
	case *ast.WhileLoop:
		return ev.evalWhileLoop(node)
	case *ast.ForLoop:
		return ev.evalForLoop(node)
	case *ast.ForEachLoop:
		return ev.evalForEachLoop(node)
	case *ast.ToggleStatement:
		return ev.evalToggle(node)
	case *ast.BreakStatement:
		return &BreakValue{}, nil
	case *ast.NumberLiteral:
		return node.Value, nil
	case *ast.StringLiteral:
		return node.Value, nil
	case *ast.BooleanLiteral:
		return node.Value, nil
	case *ast.ListLiteral:
		return ev.evalListLiteral(node)
	case *ast.Identifier:
		return ev.evalIdentifier(node)
	case *ast.BinaryExpression:
		return ev.evalBinaryExpression(node)
	case *ast.UnaryExpression:
		return ev.evalUnaryExpression(node)
	case *ast.FunctionCall:
		return ev.evalFunctionCall(node)
	case *ast.IndexExpression:
		return ev.evalIndexExpression(node)
	case *ast.LengthExpression:
		return ev.evalLengthExpression(node)
	case *ast.LocationExpression:
		return ev.evalLocationExpression(node)
	default:
		return nil, fmt.Errorf("unknown node type: %T", node)
	}
}

func (ev *Evaluator) evalProgram(prog *ast.Program) (Value, error) {
	var result Value
	for _, stmt := range prog.Statements {
		val, err := ev.Eval(stmt)
		if err != nil {
			return nil, err
		}
		if _, ok := val.(*ReturnValue); ok {
			return val, nil
		}
		result = val
	}
	return result, nil
}

func (ev *Evaluator) evalVariableDecl(vd *ast.VariableDecl) (Value, error) {
	value, err := ev.Eval(vd.Value)
	if err != nil {
		return nil, err
	}

	err = ev.env.Define(vd.Name, value, vd.IsConstant)
	return nil, err
}

func (ev *Evaluator) evalAssignment(a *ast.Assignment) (Value, error) {
	value, err := ev.Eval(a.Value)
	if err != nil {
		return nil, err
	}

	err = ev.env.Set(a.Name, value)
	return nil, err
}

func (ev *Evaluator) evalIndexAssignment(ia *ast.IndexAssignment) (Value, error) {
	// Get the list
	list, ok := ev.env.Get(ia.ListName)
	if !ok {
		return nil, ev.runtimeError(fmt.Sprintf("undefined variable '%s'", ia.ListName))
	}

	items, ok := list.([]interface{})
	if !ok {
		return nil, ev.runtimeError(fmt.Sprintf("cannot index into non-list type %T", list))
	}

	// Get the index
	indexVal, err := ev.Eval(ia.Index)
	if err != nil {
		return nil, err
	}
	index, err := ToNumber(indexVal)
	if err != nil {
		return nil, ev.runtimeError("index must be a number")
	}
	idx := int(index)
	if idx < 0 || idx >= len(items) {
		return nil, ev.runtimeError(fmt.Sprintf("index %d out of range for list of length %d", idx, len(items)))
	}

	// Get the value
	value, err := ev.Eval(ia.Value)
	if err != nil {
		return nil, err
	}

	// Update the list
	items[idx] = value
	return nil, nil
}

func (ev *Evaluator) evalIndexExpression(ie *ast.IndexExpression) (Value, error) {
	// Get the list
	list, err := ev.Eval(ie.List)
	if err != nil {
		return nil, err
	}

	items, ok := list.([]interface{})
	if !ok {
		return nil, ev.runtimeError(fmt.Sprintf("cannot index into non-list type %T", list))
	}

	// Get the index
	indexVal, err := ev.Eval(ie.Index)
	if err != nil {
		return nil, err
	}
	index, err := ToNumber(indexVal)
	if err != nil {
		return nil, ev.runtimeError("index must be a number")
	}
	idx := int(index)
	if idx < 0 || idx >= len(items) {
		return nil, ev.runtimeError(fmt.Sprintf("index %d out of range for list of length %d", idx, len(items)))
	}

	return items[idx], nil
}

func (ev *Evaluator) evalLengthExpression(le *ast.LengthExpression) (Value, error) {
	list, err := ev.Eval(le.List)
	if err != nil {
		return nil, err
	}

	switch v := list.(type) {
	case []interface{}:
		return float64(len(v)), nil
	case string:
		return float64(len(v)), nil
	default:
		return nil, ev.runtimeError(fmt.Sprintf("cannot get length of %T", list))
	}
}

func (ev *Evaluator) evalLocationExpression(loc *ast.LocationExpression) (Value, error) {
	// Check if variable exists
	_, ok := ev.env.Get(loc.Name)
	if !ok {
		return nil, ev.runtimeError(fmt.Sprintf("undefined variable '%s'", loc.Name))
	}
	// Return a unique identifier based on the variable name and environment
	return fmt.Sprintf("0x%p:%s", ev.env, loc.Name), nil
}

func (ev *Evaluator) evalToggle(ts *ast.ToggleStatement) (Value, error) {
	// Get the current value
	val, ok := ev.env.Get(ts.Name)
	if !ok {
		return nil, ev.runtimeError(fmt.Sprintf("undefined variable '%s'", ts.Name))
	}

	// Check if it's a boolean
	boolVal, isBool := val.(bool)
	if !isBool {
		return nil, ev.runtimeError(fmt.Sprintf("cannot toggle non-boolean variable '%s' (type: %T)", ts.Name, val))
	}

	// Toggle the value
	err := ev.env.Set(ts.Name, !boolVal)
	return nil, err
}

func (ev *Evaluator) evalFunctionDecl(fd *ast.FunctionDecl) (Value, error) {
	fn := &FunctionValue{
		Name:       fd.Name,
		Parameters: fd.Parameters,
		Body:       fd.Body,
		Closure:    ev.env,
	}
	ev.env.DefineFunction(fd.Name, fn)
	return nil, nil
}

func (ev *Evaluator) evalCallStatement(cs *ast.CallStatement) (Value, error) {
	_, err := ev.evalFunctionCall(cs.FunctionCall)
	return nil, err
}

func (ev *Evaluator) evalOutput(os *ast.OutputStatement) (Value, error) {
	value, err := ev.Eval(os.Value)
	if err != nil {
		return nil, err
	}
	fmt.Println(ToString(value))
	return nil, nil
}

func (ev *Evaluator) evalReturn(rs *ast.ReturnStatement) (Value, error) {
	value, err := ev.Eval(rs.Value)
	if err != nil {
		return nil, err
	}
	return &ReturnValue{Value: value}, nil
}

func (ev *Evaluator) evalIfStatement(is *ast.IfStatement) (Value, error) {
	cond, err := ev.Eval(is.Condition)
	if err != nil {
		return nil, err
	}

	oldEnv := ev.env

	if ToBool(cond) {
		// Create scoped environment for then block
		ev.env = oldEnv.NewChild()
		result, err := ev.evalStatements(is.Then)
		ev.env = oldEnv
		return result, err
	}

	for _, eif := range is.ElseIf {
		cond, err := ev.Eval(eif.Condition)
		if err != nil {
			return nil, err
		}
		if ToBool(cond) {
			// Create scoped environment for else-if block
			ev.env = oldEnv.NewChild()
			result, err := ev.evalStatements(eif.Body)
			ev.env = oldEnv
			return result, err
		}
	}

	if is.Else != nil {
		// Create scoped environment for else block
		ev.env = oldEnv.NewChild()
		result, err := ev.evalStatements(is.Else)
		ev.env = oldEnv
		return result, err
	}

	return nil, nil
}

func (ev *Evaluator) evalWhileLoop(wl *ast.WhileLoop) (Value, error) {
	var result Value
	for {
		cond, err := ev.Eval(wl.Condition)
		if err != nil {
			return nil, err
		}
		if !ToBool(cond) {
			break
		}

		// Create a new child environment for each iteration to support scoped variables
		childEnv := ev.env.NewChild()
		oldEnv := ev.env
		ev.env = childEnv

		val, err := ev.evalStatements(wl.Body)
		ev.env = oldEnv // Restore environment

		if err != nil {
			return nil, err
		}
		if _, ok := val.(*ReturnValue); ok {
			return val, nil
		}
		if _, ok := val.(*BreakValue); ok {
			break
		}
		result = val
	}
	return result, nil
}

func (ev *Evaluator) evalForLoop(fl *ast.ForLoop) (Value, error) {
	count, err := ev.Eval(fl.Count)
	if err != nil {
		return nil, err
	}

	num, err := ToNumber(count)
	if err != nil {
		return nil, err
	}

	var result Value
	for i := 0; i < int(num); i++ {
		// Create a new child environment for each iteration to support scoped variables
		childEnv := ev.env.NewChild()
		oldEnv := ev.env
		ev.env = childEnv

		val, err := ev.evalStatements(fl.Body)
		ev.env = oldEnv // Restore environment

		if err != nil {
			return nil, err
		}
		if _, ok := val.(*ReturnValue); ok {
			return val, nil
		}
		if _, ok := val.(*BreakValue); ok {
			break
		}
		result = val
	}
	return result, nil
}

func (ev *Evaluator) evalForEachLoop(fel *ast.ForEachLoop) (Value, error) {
	list, err := ev.Eval(fel.List)
	if err != nil {
		return nil, err
	}

	items, ok := list.([]interface{})
	if !ok {
		return nil, fmt.Errorf("cannot iterate over %T", list)
	}

	oldEnv := ev.env

	var result Value
	for _, item := range items {
		// Create a new child environment for each iteration to support scoped variables
		childEnv := oldEnv.NewChild()
		ev.env = childEnv
		ev.env.Define(fel.Item, item, false)

		val, err := ev.evalStatements(fel.Body)
		ev.env = oldEnv // Restore environment

		if err != nil {
			return nil, err
		}
		if _, ok := val.(*ReturnValue); ok {
			return val, nil
		}
		if _, ok := val.(*BreakValue); ok {
			break
		}
		result = val
	}
	return result, nil
}

func (ev *Evaluator) evalStatements(stmts []ast.Statement) (Value, error) {
	var result Value
	for _, stmt := range stmts {
		val, err := ev.Eval(stmt)
		if err != nil {
			return nil, err
		}
		if _, ok := val.(*ReturnValue); ok {
			return val, nil
		}
		result = val
	}
	return result, nil
}

func (ev *Evaluator) evalListLiteral(ll *ast.ListLiteral) (Value, error) {
	var result []interface{}
	for _, elem := range ll.Elements {
		val, err := ev.Eval(elem)
		if err != nil {
			return nil, err
		}
		result = append(result, val)
	}
	return result, nil
}

func (ev *Evaluator) evalIdentifier(id *ast.Identifier) (Value, error) {
	val, ok := ev.env.Get(id.Name)
	if !ok {
		suggestion := ev.findSimilarVariable(id.Name)
		if suggestion != "" {
			return nil, ev.runtimeError(fmt.Sprintf("undefined variable '%s'\n  Perhaps you meant: '%s'", id.Name, suggestion))
		}
		return nil, ev.runtimeError(fmt.Sprintf("undefined variable '%s'", id.Name))
	}
	return val, nil
}

func (ev *Evaluator) findSimilarVariable(name string) string {
	name = strings.ToLower(name)
	var candidates []string

	// Collect all variables from current and parent scopes
	env := ev.env
	for env != nil {
		for varName := range env.variables {
			candidates = append(candidates, varName)
		}
		env = env.parent
	}

	// Simple similarity check (case-insensitive match or one-char difference)
	for _, candidate := range candidates {
		if strings.ToLower(candidate) == name {
			return candidate
		}
		if levenshteinDistance(strings.ToLower(candidate), name) <= 2 {
			return candidate
		}
	}

	return ""
}

func levenshteinDistance(s1, s2 string) int {
	if len(s1) == 0 {
		return len(s2)
	}
	if len(s2) == 0 {
		return len(s1)
	}

	if len(s1) > len(s2) {
		s1, s2 = s2, s1
	}

	lenS1 := len(s1)
	lenS2 := len(s2)

	distances := make([]int, lenS1+1)
	for i := range distances {
		distances[i] = i
	}

	for i := 1; i <= lenS2; i++ {
		prev := i
		for j := 1; j <= lenS1; j++ {
			current := distances[j-1]
			if s2[i-1] != s1[j-1] {
				current = min(min(distances[j-1]+1, distances[j]+1), prev+1)
			}
			distances[j-1] = prev
			prev = current
		}
		distances[lenS1] = prev
	}

	return distances[lenS1]
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (ev *Evaluator) evalBinaryExpression(be *ast.BinaryExpression) (Value, error) {
	left, err := ev.Eval(be.Left)
	if err != nil {
		return nil, err
	}

	right, err := ev.Eval(be.Right)
	if err != nil {
		return nil, err
	}

	switch be.Operator {
	case "+":
		return Add(left, right)
	case "-":
		return Subtract(left, right)
	case "*":
		return Multiply(left, right)
	case "/":
		return Divide(left, right)
	case "%":
		return Modulo(left, right)
	case "is equal to", "is less than", "is greater than", "is less than or equal to", "is greater than or equal to", "is not equal to":
		result, err := Compare(be.Operator, left, right)
		return result, err
	default:
		return nil, fmt.Errorf("unknown operator: %s", be.Operator)
	}
}

func (ev *Evaluator) evalUnaryExpression(ue *ast.UnaryExpression) (Value, error) {
	right, err := ev.Eval(ue.Right)
	if err != nil {
		return nil, err
	}

	switch ue.Operator {
	case "-":
		num, err := ToNumber(right)
		if err != nil {
			return nil, err
		}
		return -num, nil
	default:
		return nil, fmt.Errorf("unknown unary operator: %s", ue.Operator)
	}
}

func (ev *Evaluator) evalFunctionCall(fc *ast.FunctionCall) (Value, error) {
	fn, ok := ev.env.GetFunction(fc.Name)
	if !ok {
		suggestion := ev.findSimilarFunction(fc.Name)
		if suggestion != "" {
			return nil, ev.runtimeError(fmt.Sprintf("undefined function '%s'\n  Perhaps you meant: '%s'", fc.Name, suggestion))
		}
		return nil, ev.runtimeError(fmt.Sprintf("undefined function '%s'", fc.Name))
	}

	// Evaluate arguments
	args := make([]Value, len(fc.Arguments))
	for i, arg := range fc.Arguments {
		val, err := ev.Eval(arg)
		if err != nil {
			return nil, err
		}
		args[i] = val
	}

	// Check parameter count
	if len(args) != len(fn.Parameters) {
		expected := "no arguments"
		if len(fn.Parameters) == 1 {
			expected = "1 argument"
		} else if len(fn.Parameters) > 1 {
			expected = fmt.Sprintf("%d arguments", len(fn.Parameters))
		}

		got := "none"
		if len(args) == 1 {
			got = "1"
		} else if len(args) > 1 {
			got = fmt.Sprintf("%d", len(args))
		}

		paramList := ""
		if len(fn.Parameters) > 0 {
			paramList = fmt.Sprintf("\n  Expected parameters: %s", strings.Join(fn.Parameters, ", "))
		}

		return nil, ev.runtimeError(fmt.Sprintf("function '%s' expects %s, got %s%s", fc.Name, expected, got, paramList))
	}

	// Create new environment for function execution
	funcEnv := fn.Closure.NewChild()

	// Bind parameters
	for i, param := range fn.Parameters {
		funcEnv.Define(param, args[i], false)
	}

	// Execute function body with new environment
	oldEnv := ev.env
	ev.env = funcEnv

	// Add to call stack
	ev.callStack = append(ev.callStack, fmt.Sprintf("%s(%s)", fc.Name, strings.Join(fn.Parameters, ", ")))

	defer func() {
		ev.env = oldEnv
		ev.callStack = ev.callStack[:len(ev.callStack)-1]
	}()

	for _, stmt := range fn.Body {
		val, err := ev.Eval(stmt)
		if err != nil {
			return nil, err
		}
		if retVal, ok := val.(*ReturnValue); ok {
			return retVal.Value, nil
		}
	}

	return nil, nil
}

func (ev *Evaluator) findSimilarFunction(name string) string {
	name = strings.ToLower(name)
	var candidates []string

	// Collect all functions
	env := ev.env
	for env != nil {
		for fnName := range env.functions {
			candidates = append(candidates, fnName)
		}
		env = env.parent
	}

	// Simple similarity check
	for _, candidate := range candidates {
		if strings.ToLower(candidate) == name {
			return candidate
		}
		if levenshteinDistance(strings.ToLower(candidate), name) <= 2 {
			return candidate
		}
	}

	return ""
}

// ToNumber converts a value to a number
func ToNumber(v Value) (float64, error) {
	switch val := v.(type) {
	case float64:
		return val, nil
	case string:
		n, err := strconv.ParseFloat(val, 64)
		return n, err
	default:
		return 0, fmt.Errorf("cannot convert %T to number", v)
	}
}

// ToString converts a value to a string
func ToString(v Value) string {
	switch val := v.(type) {
	case float64:
		if val == float64(int64(val)) {
			return strconv.FormatInt(int64(val), 10)
		}
		return strconv.FormatFloat(val, 'f', -1, 64)
	case string:
		return val
	case bool:
		if val {
			return "true"
		}
		return "false"
	case []interface{}:
		var parts []string
		for _, elem := range val {
			parts = append(parts, ToString(elem))
		}
		return "[" + strings.Join(parts, " ") + "]"
	case *FunctionValue:
		return fmt.Sprintf("<function %s>", val.Name)
	case nil:
		return "nil"
	default:
		return fmt.Sprintf("%v", val)
	}
}

// ToBool converts a value to a boolean
func ToBool(v Value) bool {
	switch val := v.(type) {
	case bool:
		return val
	case float64:
		return val != 0
	case string:
		return val != ""
	case []interface{}:
		return len(val) > 0
	case nil:
		return false
	default:
		return true
	}
}

// Compare compares two values based on an operator
func Compare(op string, left, right Value) (bool, error) {
	switch op {
	case "is equal to":
		return Equals(left, right), nil
	case "is less than":
		l, err := ToNumber(left)
		if err != nil {
			return false, err
		}
		r, err := ToNumber(right)
		if err != nil {
			return false, err
		}
		return l < r, nil
	case "is greater than":
		l, err := ToNumber(left)
		if err != nil {
			return false, err
		}
		r, err := ToNumber(right)
		if err != nil {
			return false, err
		}
		return l > r, nil
	case "is less than or equal to":
		l, err := ToNumber(left)
		if err != nil {
			return false, err
		}
		r, err := ToNumber(right)
		if err != nil {
			return false, err
		}
		return l <= r, nil
	case "is greater than or equal to":
		l, err := ToNumber(left)
		if err != nil {
			return false, err
		}
		r, err := ToNumber(right)
		if err != nil {
			return false, err
		}
		return l >= r, nil
	case "is not equal to":
		return !Equals(left, right), nil
	default:
		return false, fmt.Errorf("unknown comparison operator: %s", op)
	}
}

// Equals checks if two values are equal
func Equals(left, right Value) bool {
	switch l := left.(type) {
	case float64:
		switch r := right.(type) {
		case float64:
			return l == r
		default:
			return false
		}
	case string:
		switch r := right.(type) {
		case string:
			return l == r
		default:
			return false
		}
	case bool:
		switch r := right.(type) {
		case bool:
			return l == r
		default:
			return false
		}
	case nil:
		return right == nil
	default:
		return false
	}
}

// Add adds two values
func Add(left, right Value) (Value, error) {
	switch l := left.(type) {
	case float64:
		r, err := ToNumber(right)
		if err != nil {
			return nil, err
		}
		return l + r, nil
	case string:
		return l + ToString(right), nil
	case []interface{}:
		switch r := right.(type) {
		case []interface{}:
			return append(l, r...), nil
		default:
			return append(l, r), nil
		}
	default:
		return nil, fmt.Errorf("cannot add %T and %T", left, right)
	}
}

// Subtract subtracts two values
func Subtract(left, right Value) (Value, error) {
	l, err := ToNumber(left)
	if err != nil {
		return nil, fmt.Errorf("cannot subtract: left operand is not a number (got %T)\n  Hint: Subtraction only works with numbers", left)
	}
	r, err := ToNumber(right)
	if err != nil {
		return nil, fmt.Errorf("cannot subtract: right operand is not a number (got %T)\n  Hint: Subtraction only works with numbers", right)
	}
	return l - r, nil
}

// Multiply multiplies two values
func Multiply(left, right Value) (Value, error) {
	switch l := left.(type) {
	case float64:
		r, err := ToNumber(right)
		if err != nil {
			return nil, err
		}
		return l * r, nil
	case string:
		r, err := ToNumber(right)
		if err != nil {
			return nil, err
		}
		result := ""
		for i := 0; i < int(r); i++ {
			result += l
		}
		return result, nil
	default:
		return nil, fmt.Errorf("cannot multiply %T and %T", left, right)
	}
}

// Divide divides two values
func Divide(left, right Value) (Value, error) {
	l, err := ToNumber(left)
	if err != nil {
		return nil, fmt.Errorf("cannot divide: left operand is not a number (got %T)\n  Hint: Division only works with numbers", left)
	}
	r, err := ToNumber(right)
	if err != nil {
		return nil, fmt.Errorf("cannot divide: right operand is not a number (got %T)\n  Hint: Division only works with numbers", right)
	}
	if r == 0 {
		return nil, fmt.Errorf("division by zero\n  Hint: You cannot divide by zero - check your expression")
	}
	return l / r, nil
}

// Modulo calculates the remainder of two values
func Modulo(left, right Value) (Value, error) {
	l, err := ToNumber(left)
	if err != nil {
		return nil, fmt.Errorf("cannot get remainder: left operand is not a number (got %T)\n  Hint: Remainder only works with numbers", left)
	}
	r, err := ToNumber(right)
	if err != nil {
		return nil, fmt.Errorf("cannot get remainder: right operand is not a number (got %T)\n  Hint: Remainder only works with numbers", right)
	}
	if r == 0 {
		return nil, fmt.Errorf("division by zero\n  Hint: You cannot get remainder when dividing by zero")
	}
	return float64(int64(l) % int64(r)), nil
}
