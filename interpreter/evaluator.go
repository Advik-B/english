package interpreter

import (
	"fmt"
	"strings"
)

// Environment represents a scope for variables and functions
type Environment struct {
	variables map[string]Value
	constants map[string]bool
	functions map[string]*FunctionValue
	parent    *Environment
}

func NewEnvironment() *Environment {
	return &Environment{
		variables: make(map[string]Value),
		constants: make(map[string]bool),
		functions: make(map[string]*FunctionValue),
	}
}

func (e *Environment) NewChild() *Environment {
	return &Environment{
		variables: make(map[string]Value),
		constants: make(map[string]bool),
		functions: make(map[string]*FunctionValue),
		parent:    e,
	}
}

func (e *Environment) Get(name string) (Value, bool) {
	if val, ok := e.variables[name]; ok {
		return val, ok
	}
	if e.parent != nil {
		return e.parent.Get(name)
	}
	return nil, false
}

func (e *Environment) Set(name string, value Value) error {
	if e.constants[name] {
		return fmt.Errorf("cannot reassign constant '%s'\n  Hint: Constants are declared with 'to be always' or 'to always be'", name)
	}
	if _, exists := e.variables[name]; exists {
		e.variables[name] = value
		return nil
	}
	if e.parent != nil {
		if _, exists := e.parent.variables[name]; exists {
			return e.parent.Set(name, value)
		}
	}
	e.variables[name] = value
	return nil
}

func (e *Environment) Define(name string, value Value, isConstant bool) error {
	if _, exists := e.variables[name]; exists {
		return fmt.Errorf("variable %s already defined", name)
	}
	e.variables[name] = value
	e.constants[name] = isConstant
	return nil
}

func (e *Environment) GetFunction(name string) (*FunctionValue, bool) {
	if fn, ok := e.functions[name]; ok {
		return fn, ok
	}
	if e.parent != nil {
		return e.parent.GetFunction(name)
	}
	return nil, false
}

func (e *Environment) DefineFunction(name string, fn *FunctionValue) {
	e.functions[name] = fn
}

// ReturnValue is used to implement return statements
type ReturnValue struct {
	Value Value
}

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

// Evaluator executes the AST
type Evaluator struct {
	env       *Environment
	callStack []string
}

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

func (ev *Evaluator) Eval(node interface{}) (Value, error) {
	switch node := node.(type) {
	case *Program:
		return ev.evalProgram(node)
	case *VariableDecl:
		return ev.evalVariableDecl(node)
	case *Assignment:
		return ev.evalAssignment(node)
	case *FunctionDecl:
		return ev.evalFunctionDecl(node)
	case *CallStatement:
		return ev.evalCallStatement(node)
	case *OutputStatement:
		return ev.evalOutput(node)
	case *ReturnStatement:
		return ev.evalReturn(node)
	case *IfStatement:
		return ev.evalIfStatement(node)
	case *WhileLoop:
		return ev.evalWhileLoop(node)
	case *ForLoop:
		return ev.evalForLoop(node)
	case *ForEachLoop:
		return ev.evalForEachLoop(node)
	case *NumberLiteral:
		return node.Value, nil
	case *StringLiteral:
		return node.Value, nil
	case *ListLiteral:
		return ev.evalListLiteral(node)
	case *Identifier:
		return ev.evalIdentifier(node)
	case *BinaryExpression:
		return ev.evalBinaryExpression(node)
	case *UnaryExpression:
		return ev.evalUnaryExpression(node)
	case *FunctionCall:
		return ev.evalFunctionCall(node)
	default:
		return nil, fmt.Errorf("unknown node type: %T", node)
	}
}

func (ev *Evaluator) evalProgram(prog *Program) (Value, error) {
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

func (ev *Evaluator) evalVariableDecl(vd *VariableDecl) (Value, error) {
	value, err := ev.Eval(vd.Value)
	if err != nil {
		return nil, err
	}

	err = ev.env.Define(vd.Name, value, vd.IsConstant)
	return nil, err
}

func (ev *Evaluator) evalAssignment(a *Assignment) (Value, error) {
	value, err := ev.Eval(a.Value)
	if err != nil {
		return nil, err
	}

	err = ev.env.Set(a.Name, value)
	return nil, err
}

func (ev *Evaluator) evalFunctionDecl(fd *FunctionDecl) (Value, error) {
	fn := &FunctionValue{
		Name:       fd.Name,
		Parameters: fd.Parameters,
		Body:       fd.Body,
		Closure:    ev.env,
	}
	ev.env.DefineFunction(fd.Name, fn)
	return nil, nil
}

func (ev *Evaluator) evalCallStatement(cs *CallStatement) (Value, error) {
	_, err := ev.evalFunctionCall(cs.FunctionCall)
	return nil, err
}

func (ev *Evaluator) evalOutput(os *OutputStatement) (Value, error) {
	value, err := ev.Eval(os.Value)
	if err != nil {
		return nil, err
	}
	fmt.Println(toString(value))
	return nil, nil
}

func (ev *Evaluator) evalReturn(rs *ReturnStatement) (Value, error) {
	value, err := ev.Eval(rs.Value)
	if err != nil {
		return nil, err
	}
	return &ReturnValue{Value: value}, nil
}

func (ev *Evaluator) evalIfStatement(is *IfStatement) (Value, error) {
	cond, err := ev.Eval(is.Condition)
	if err != nil {
		return nil, err
	}

	if toBool(cond) {
		return ev.evalStatements(is.Then)
	}

	for _, eif := range is.ElseIf {
		cond, err := ev.Eval(eif.Condition)
		if err != nil {
			return nil, err
		}
		if toBool(cond) {
			return ev.evalStatements(eif.Body)
		}
	}

	if is.Else != nil {
		return ev.evalStatements(is.Else)
	}

	return nil, nil
}

func (ev *Evaluator) evalWhileLoop(wl *WhileLoop) (Value, error) {
	var result Value
	for {
		cond, err := ev.Eval(wl.Condition)
		if err != nil {
			return nil, err
		}
		if !toBool(cond) {
			break
		}

		val, err := ev.evalStatements(wl.Body)
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

func (ev *Evaluator) evalForLoop(fl *ForLoop) (Value, error) {
	count, err := ev.Eval(fl.Count)
	if err != nil {
		return nil, err
	}

	num, err := toNumber(count)
	if err != nil {
		return nil, err
	}

	var result Value
	for i := 0; i < int(num); i++ {
		val, err := ev.evalStatements(fl.Body)
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

func (ev *Evaluator) evalForEachLoop(fel *ForEachLoop) (Value, error) {
	list, err := ev.Eval(fel.List)
	if err != nil {
		return nil, err
	}

	items, ok := list.([]interface{})
	if !ok {
		return nil, fmt.Errorf("cannot iterate over %T", list)
	}

	childEnv := ev.env.NewChild()
	oldEnv := ev.env
	ev.env = childEnv
	defer func() { ev.env = oldEnv }()

	var result Value
	for _, item := range items {
		// Use direct assignment to update the loop variable each iteration
		ev.env.variables[fel.Item] = item
		val, err := ev.evalStatements(fel.Body)
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

func (ev *Evaluator) evalStatements(stmts []Statement) (Value, error) {
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

func (ev *Evaluator) evalListLiteral(ll *ListLiteral) (Value, error) {
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

func (ev *Evaluator) evalIdentifier(id *Identifier) (Value, error) {
	val, ok := ev.env.Get(id.Name)
	if !ok {
		// Try to find similar variable names
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

	// Create distance matrix
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

func (ev *Evaluator) evalBinaryExpression(be *BinaryExpression) (Value, error) {
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
		return add(left, right)
	case "-":
		return subtract(left, right)
	case "*":
		return multiply(left, right)
	case "/":
		return divide(left, right)
	case "is equal to", "is less than", "is greater than", "is less than or equal to", "is greater than or equal to", "is not equal to":
		result, err := compare(be.Operator, left, right)
		return result, err
	default:
		return nil, fmt.Errorf("unknown operator: %s", be.Operator)
	}
}

func (ev *Evaluator) evalUnaryExpression(ue *UnaryExpression) (Value, error) {
	right, err := ev.Eval(ue.Right)
	if err != nil {
		return nil, err
	}

	switch ue.Operator {
	case "-":
		num, err := toNumber(right)
		if err != nil {
			return nil, err
		}
		return -num, nil
	default:
		return nil, fmt.Errorf("unknown unary operator: %s", ue.Operator)
	}
}

func (ev *Evaluator) evalFunctionCall(fc *FunctionCall) (Value, error) {
	fn, ok := ev.env.GetFunction(fc.Name)
	if !ok {
		// Try to find similar function names
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
