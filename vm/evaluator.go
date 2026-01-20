package vm

import (
	"english/ast"
	"english/parser"
	"fmt"
	"os"
	"strings"
)

// Evaluator executes the AST
type Evaluator struct {
	env       *Environment
	callStack []string
}

// NewEvaluator creates a new evaluator with the given environment
func NewEvaluator(env *Environment) *Evaluator {
	// Register standard library functions
	RegisterStdlib(env)
	
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
	case *ast.ImportStatement:
		return ev.evalImport(node)
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
	case *ast.MethodCall:
		return ev.evalMethodCall(node)
	case *ast.IndexExpression:
		return ev.evalIndexExpression(node)
	case *ast.LengthExpression:
		return ev.evalLengthExpression(node)
	case *ast.LocationExpression:
		return ev.evalLocationExpression(node)
	// New AST node types
	case *ast.StructDecl:
		return ev.evalStructDecl(node)
	case *ast.StructInstantiation:
		return ev.evalStructInstantiation(node)
	case *ast.FieldAccess:
		return ev.evalFieldAccess(node)
	case *ast.FieldAssignment:
		return ev.evalFieldAssignment(node)
	case *ast.TryStatement:
		return ev.evalTryStatement(node)
	case *ast.RaiseStatement:
		return ev.evalRaiseStatement(node)
	case *ast.SwapStatement:
		return ev.evalSwapStatement(node)
	case *ast.TypeExpression:
		return ev.evalTypeExpression(node)
	case *ast.CastExpression:
		return ev.evalCastExpression(node)
	case *ast.ReferenceExpression:
		return ev.evalReferenceExpression(node)
	case *ast.CopyExpression:
		return ev.evalCopyExpression(node)
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

func (ev *Evaluator) evalImport(is *ast.ImportStatement) (Value, error) {
	// Import evaluates another English file and executes it in the current environment
	// This allows sharing variables and functions across files
	
	// Read the file content
	content, err := os.ReadFile(is.Path)
	if err != nil {
		return nil, ev.runtimeError(fmt.Sprintf("failed to import '%s': %v", is.Path, err))
	}

	// Parse the imported file
	lexer := parser.NewLexer(string(content))
	tokens := lexer.TokenizeAll()
	p := parser.NewParser(tokens)
	program, err := p.Parse()
	if err != nil {
		return nil, ev.runtimeError(fmt.Sprintf("failed to parse imported file '%s': %v", is.Path, err))
	}

	// Execute the imported program in the current environment
	// This makes all declarations from the imported file available in the current scope
	_, err = ev.evalProgram(program)
	if err != nil {
		return nil, err
	}

	return nil, nil
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
	if cs.MethodCall != nil {
		_, err := ev.evalMethodCall(cs.MethodCall)
		return nil, err
	}
	_, err := ev.evalFunctionCall(cs.FunctionCall)
	return nil, err
}

func (ev *Evaluator) evalOutput(os *ast.OutputStatement) (Value, error) {
	var parts []string
	for _, expr := range os.Values {
		value, err := ev.Eval(expr)
		if err != nil {
			return nil, err
		}
		parts = append(parts, ToString(value))
	}
	output := strings.Join(parts, " ")
	if os.Newline {
		fmt.Println(output)
	} else {
		fmt.Print(output)
	}
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
	var candidates []string

	// Collect all variables from current and parent scopes
	env := ev.env
	for env != nil {
		for varName := range env.variables {
			candidates = append(candidates, varName)
		}
		env = env.parent
	}

	return findSimilarName(name, candidates)
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

	// Check if it's a built-in function (stdlib)
	if fn.Body == nil {
		// This is a built-in function, delegate to stdlib
		return ev.evalBuiltinFunction(fc.Name, args)
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
	var candidates []string

	// Collect all functions
	env := ev.env
	for env != nil {
		for fnName := range env.functions {
			candidates = append(candidates, fnName)
		}
		env = env.parent
	}

	return findSimilarName(name, candidates)
}
