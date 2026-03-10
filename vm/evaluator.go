package vm

import (
	"bufio"
	"english/ast"
	"english/bytecode"
	"english/parser"
	"english/vm/types"
	"fmt"
	"os"
	"strings"
)

// Evaluator executes the AST
type Evaluator struct {
	env       *Environment
	callStack []string
	builtinFn BuiltinFunc // injected stdlib evaluator
}

// NewEvaluator creates a new evaluator with the given environment and optional builtin function.
func NewEvaluator(env *Environment, builtinFn BuiltinFunc) *Evaluator {
	return &Evaluator{
		env:       env,
		callStack: []string{"<main>"},
		builtinFn: builtinFn,
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
	case *ast.TypedVariableDecl:
		return ev.evalTypedVariableDecl(node)
	case *ast.ErrorTypeDecl:
		return ev.evalErrorTypeDecl(node)
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
	case *ast.ContinueStatement:
		return &ContinueValue{}, nil
	case *ast.NothingLiteral:
		return nil, nil
	case *ast.AskExpression:
		return ev.evalAskExpression(node)
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
	// Composite type nodes
	case *ast.ArrayLiteral:
		return ev.evalArrayLiteral(node)
	case *ast.LookupTableLiteral:
		return types.NewLookupTable(), nil
	case *ast.LookupKeyAccess:
		return ev.evalLookupKeyAccess(node)
	case *ast.LookupKeyAssignment:
		return ev.evalLookupKeyAssignment(node)
	case *ast.HasExpression:
		return ev.evalHasExpression(node)
	case *ast.NilCheckExpression:
		return ev.evalNilCheckExpression(node)
	case *ast.ErrorTypeCheckExpression:
		return ev.evalErrorTypeCheckExpression(node)
	// Struct / type / cast nodes
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
	case *ast.CommentStatement:
		// Comments are no-ops at runtime.
		return nil, nil
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
	// Import evaluates another English file and executes it in the current environment.
	// Supports selective imports, import all, and safe imports.
	// Uses bytecode caching (__engcache__) for faster loading.

	// Try to load from cache or parse the source file
	parseFunc := func(path string) (*ast.Program, error) {
		content, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("failed to read file: %w", err)
		}

		lexer := parser.NewLexer(string(content))
		tokens := lexer.TokenizeAll()
		p := parser.NewParser(tokens)
		return p.Parse()
	}

	program, _, err := bytecode.LoadCachedOrParse(is.Path, parseFunc)
	if err != nil {
		return nil, ev.runtimeError(fmt.Sprintf("failed to import '%s': %v", is.Path, err))
	}

	// Handle different import modes
	if is.IsSafe {
		// Safe import: only execute declarations, skip top-level statements
		return ev.evalSafeImport(program, is)
	} else if len(is.Items) > 0 {
		// Selective import: import specific items
		return ev.evalSelectiveImport(program, is)
	} else {
		// Import all: execute everything in current scope
		_, err = ev.evalProgram(program)
		if err != nil {
			return nil, err
		}
	}

	// Import statements don't produce a value
	return nil, nil
}

// evalSafeImport executes only declarations (functions, variables) and skips top-level statements
func (ev *Evaluator) evalSafeImport(program *ast.Program, is *ast.ImportStatement) (Value, error) {
	for _, stmt := range program.Statements {
		// Only execute declarations, skip other statements
		switch s := stmt.(type) {
		case *ast.VariableDecl:
			_, err := ev.evalVariableDecl(s)
			if err != nil {
				return nil, err
			}
		case *ast.FunctionDecl:
			_, err := ev.evalFunctionDecl(s)
			if err != nil {
				return nil, err
			}
			// Skip all other statement types (Print, Call, etc.)
		}
	}
	return nil, nil
}

// evalSelectiveImport imports only specific items from the file
func (ev *Evaluator) evalSelectiveImport(program *ast.Program, is *ast.ImportStatement) (Value, error) {
	// Create a temporary environment for the imported file
	tempEnv := NewEnvironment()
	tempEval := NewEvaluator(tempEnv, ev.builtinFn)

	// Execute in temporary environment
	_, err := tempEval.evalProgram(program)
	if err != nil {
		return nil, err
	}

	// Import only the requested items
	for _, itemName := range is.Items {
		// Try to get as variable
		if val, ok := tempEnv.Get(itemName); ok {
			// Check if it's constant
			isConst := tempEnv.IsConstant(itemName)
			err := ev.env.Define(itemName, val, isConst)
			if err != nil {
				return nil, ev.runtimeError(fmt.Sprintf("failed to import '%s' from '%s': %v", itemName, is.Path, err))
			}
		} else if fn, ok := tempEnv.GetFunction(itemName); ok {
			// Import as function
			ev.env.DefineFunction(itemName, fn)
		} else {
			return nil, ev.runtimeError(fmt.Sprintf("'%s' not found in '%s'", itemName, is.Path))
		}
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
	list, ok := ev.env.Get(ia.ListName)
	if !ok {
		return nil, ev.runtimeError(fmt.Sprintf("undefined variable '%s'", ia.ListName))
	}

	indexVal, err := ev.Eval(ia.Index)
	if err != nil {
		return nil, err
	}
	index, err := ToNumber(indexVal)
	if err != nil {
		return nil, ev.runtimeError("index must be a number")
	}
	idx := int(index)

	value, err := ev.Eval(ia.Value)
	if err != nil {
		return nil, err
	}

	switch items := list.(type) {
	case []interface{}:
		if idx < 0 || idx >= len(items) {
			return nil, ev.runtimeError(fmt.Sprintf("index %d out of range for list of length %d", idx, len(items)))
		}
		items[idx] = value
	case *ArrayValue:
		if idx < 0 || idx >= len(items.Elements) {
			return nil, ev.runtimeError(fmt.Sprintf("index %d out of range for array of length %d", idx, len(items.Elements)))
		}
		// Type-check the new value against the array's element type
		if value != nil && items.ElementType != types.TypeUnknown {
			vk := types.Canonical(inferTypeKind(value))
			if vk != types.Canonical(items.ElementType) {
				return nil, ev.runtimeError(fmt.Sprintf(
					"TypeError: cannot assign %s to array of %s",
					typeKindName(inferTypeKind(value)), typeKindName(items.ElementType),
				))
			}
		}
		items.Elements[idx] = value
	default:
		return nil, ev.runtimeError(fmt.Sprintf("cannot index into %s", typeKindName(inferTypeKind(list))))
	}
	return nil, nil
}

func (ev *Evaluator) evalIndexExpression(ie *ast.IndexExpression) (Value, error) {
	list, err := ev.Eval(ie.List)
	if err != nil {
		return nil, err
	}

	indexVal, err := ev.Eval(ie.Index)
	if err != nil {
		return nil, err
	}
	index, err := ToNumber(indexVal)
	if err != nil {
		return nil, ev.runtimeError("index must be a number")
	}
	idx := int(index)

	switch items := list.(type) {
	case []interface{}:
		if idx < 0 || idx >= len(items) {
			return nil, ev.runtimeError(fmt.Sprintf("index %d out of range for list of length %d", idx, len(items)))
		}
		return items[idx], nil
	case *ArrayValue:
		if idx < 0 || idx >= len(items.Elements) {
			return nil, ev.runtimeError(fmt.Sprintf("index %d out of range for array of length %d", idx, len(items.Elements)))
		}
		return items.Elements[idx], nil
	default:
		return nil, ev.runtimeError(fmt.Sprintf("TypeError: cannot index into %s", typeKindName(inferTypeKind(list))))
	}
}

func (ev *Evaluator) evalLengthExpression(le *ast.LengthExpression) (Value, error) {
	list, err := ev.Eval(le.List)
	if err != nil {
		return nil, err
	}

	switch v := list.(type) {
	case []interface{}:
		return float64(len(v)), nil
	case *ArrayValue:
		return float64(len(v.Elements)), nil
	case *LookupTableValue:
		return float64(len(v.Entries)), nil
	case string:
		return float64(len(v)), nil
	default:
		return nil, ev.runtimeError(fmt.Sprintf("cannot get length of %s", typeKindName(inferTypeKind(list))))
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

func (ev *Evaluator) evalAskExpression(ae *ast.AskExpression) (Value, error) {
	// Display the prompt if provided
	if ae.Prompt != nil {
		prompt, err := ev.Eval(ae.Prompt)
		if err != nil {
			return nil, err
		}
		fmt.Print(ToString(prompt))
	}

	// Read a line from stdin
	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil {
		// EOF is acceptable (e.g. input from pipe)
		if len(line) == 0 {
			return "", nil
		}
	}
	// Trim trailing newline characters
	line = strings.TrimRight(line, "\r\n")
	return line, nil
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

	if condBool, condErr := ToBool(cond); condErr != nil {
		return nil, condErr
	} else if condBool {
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
		if condBool, condErr := ToBool(cond); condErr != nil {
			return nil, condErr
		} else if condBool {
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
		if condBool, condErr := ToBool(cond); condErr != nil {
			return nil, condErr
		} else if !condBool {
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
		if _, ok := val.(*ContinueValue); ok {
			continue
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
		if _, ok := val.(*ContinueValue); ok {
			continue
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

	oldEnv := ev.env
	var result Value

	switch col := list.(type) {
	case []interface{}:
		for _, item := range col {
			childEnv := oldEnv.NewChild()
			ev.env = childEnv
			ev.env.Define(fel.Item, item, false)
			val, err := ev.evalStatements(fel.Body)
			ev.env = oldEnv
			if err != nil {
				return nil, err
			}
			if _, ok := val.(*ReturnValue); ok {
				return val, nil
			}
			if _, ok := val.(*BreakValue); ok {
				break
			}
			if _, ok := val.(*ContinueValue); ok {
				continue
			}
			result = val
		}
	case *ArrayValue:
		for _, item := range col.Elements {
			childEnv := oldEnv.NewChild()
			ev.env = childEnv
			ev.env.Define(fel.Item, item, false)
			val, err := ev.evalStatements(fel.Body)
			ev.env = oldEnv
			if err != nil {
				return nil, err
			}
			if _, ok := val.(*ReturnValue); ok {
				return val, nil
			}
			if _, ok := val.(*BreakValue); ok {
				break
			}
			if _, ok := val.(*ContinueValue); ok {
				continue
			}
			result = val
		}
	case *LookupTableValue:
		// Iterating over a lookup table yields the keys in insertion order
		for _, serialKey := range col.KeyOrder {
			origKey, _, ok := types.DeserializeKey(serialKey)
			if !ok {
				origKey = serialKey
			}
			childEnv := oldEnv.NewChild()
			ev.env = childEnv
			ev.env.Define(fel.Item, origKey, false)
			val, err := ev.evalStatements(fel.Body)
			ev.env = oldEnv
			if err != nil {
				return nil, err
			}
			if _, ok := val.(*ReturnValue); ok {
				return val, nil
			}
			if _, ok := val.(*BreakValue); ok {
				break
			}
			if _, ok := val.(*ContinueValue); ok {
				continue
			}
			result = val
		}
	default:
		return nil, fmt.Errorf("TypeError: 'for each' requires list, array, or lookup table; got %s",
			typeKindName(inferTypeKind(list)))
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
		// Propagate control-flow signals so they escape nested blocks (e.g. break/continue
		// inside an if-block inside a loop). The enclosing loop evaluator consumes them.
		if _, ok := val.(*ReturnValue); ok {
			return val, nil
		}
		if _, ok := val.(*BreakValue); ok {
			return val, nil
		}
		if _, ok := val.(*ContinueValue); ok {
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
	// Short-circuit evaluation for logical operators
	if be.Operator == "and" {
		left, err := ev.Eval(be.Left)
		if err != nil {
			return nil, err
		}
		leftBool, leftErr := ToBool(left)
		if leftErr != nil {
			return nil, leftErr
		}
		if !leftBool {
			return false, nil
		}
		right, err := ev.Eval(be.Right)
		if err != nil {
			return nil, err
		}
		rightBool, rightErr := ToBool(right)
		if rightErr != nil {
			return nil, rightErr
		}
		return rightBool, nil
	}
	if be.Operator == "or" {
		left, err := ev.Eval(be.Left)
		if err != nil {
			return nil, err
		}
		leftBool, leftErr := ToBool(left)
		if leftErr != nil {
			return nil, leftErr
		}
		if leftBool {
			return true, nil
		}
		right, err := ev.Eval(be.Right)
		if err != nil {
			return nil, err
		}
		rightBool, rightErr := ToBool(right)
		if rightErr != nil {
			return nil, rightErr
		}
		return rightBool, nil
	}

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
		result, err := Divide(left, right)
		if err != nil {
			return nil, ev.runtimeError(err.Error())
		}
		return result, nil
	case "%":
		result, err := Modulo(left, right)
		if err != nil {
			return nil, ev.runtimeError(err.Error())
		}
		return result, nil
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
	case "not":
		rightBool, rightErr := ToBool(right)
		if rightErr != nil {
			return nil, rightErr
		}
		return !rightBool, nil
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

// callFunction invokes a named function with pre-evaluated argument values.
// This is used by evalMethodCall for the stdlib-fallback path.
func (ev *Evaluator) callFunction(name string, args []Value) (Value, error) {
	fn, ok := ev.env.GetFunction(name)
	if !ok {
		suggestion := ev.findSimilarFunction(name)
		if suggestion != "" {
			return nil, ev.runtimeError(fmt.Sprintf("undefined function '%s'\n  Perhaps you meant: '%s'", name, suggestion))
		}
		return nil, ev.runtimeError(fmt.Sprintf("undefined function '%s'", name))
	}

	// Built-in (stdlib) path
	if fn.Body == nil {
		return ev.evalBuiltinFunction(name, args)
	}

	// User-defined function path
	if len(args) != len(fn.Parameters) {
		return nil, ev.runtimeError(fmt.Sprintf("function '%s' expects %d argument(s), got %d", name, len(fn.Parameters), len(args)))
	}

	funcEnv := fn.Closure.NewChild()
	for i, param := range fn.Parameters {
		funcEnv.Define(param, args[i], false)
	}

	oldEnv := ev.env
	ev.env = funcEnv
	ev.callStack = append(ev.callStack, fmt.Sprintf("%s(...)", name))
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

// ─── Array ────────────────────────────────────────────────────────────────────

func (ev *Evaluator) evalArrayLiteral(al *ast.ArrayLiteral) (Value, error) {
	elements := make([]interface{}, 0, len(al.Elements))

	// Determine element type: from explicit hint or infer from first element
	elemType := types.TypeUnknown
	if al.ElementType != "" {
		elemType = types.Parse(al.ElementType)
	}

	for _, expr := range al.Elements {
		val, err := ev.Eval(expr)
		if err != nil {
			return nil, err
		}
		valType := types.Canonical(inferTypeKind(val))

		// Infer element type from first element if not explicitly given
		if elemType == types.TypeUnknown && val != nil {
			elemType = valType
		}

		// Enforce homogeneity
		if elemType != types.TypeUnknown && val != nil && types.Canonical(valType) != types.Canonical(elemType) {
			return nil, fmt.Errorf(
				"TypeError: array element has wrong type: expected %s, got %s",
				typeKindName(elemType), typeKindName(valType),
			)
		}
		elements = append(elements, val)
	}

	return &ArrayValue{ElementType: elemType, Elements: elements}, nil
}

// ─── Lookup table ─────────────────────────────────────────────────────────────

func (ev *Evaluator) evalLookupKeyAccess(la *ast.LookupKeyAccess) (Value, error) {
	tableVal, err := ev.Eval(la.Table)
	if err != nil {
		return nil, err
	}
	lt, ok := tableVal.(*LookupTableValue)
	if !ok {
		return nil, fmt.Errorf("TypeError: cannot index %s with a key; expected lookup table",
			typeKindName(inferTypeKind(tableVal)))
	}

	keyVal, err := ev.Eval(la.Key)
	if err != nil {
		return nil, err
	}
	serialKey, err := types.SerializeKey(keyVal)
	if err != nil {
		return nil, err
	}

	val, exists := lt.Entries[serialKey]
	if !exists {
		return nil, fmt.Errorf("KeyError: key %s not found in lookup table", ToString(keyVal))
	}
	return val, nil
}

func (ev *Evaluator) evalLookupKeyAssignment(la *ast.LookupKeyAssignment) (Value, error) {
	tableVal, ok := ev.env.Get(la.TableName)
	if !ok {
		return nil, fmt.Errorf("undefined variable '%s'", la.TableName)
	}
	lt, ok := tableVal.(*LookupTableValue)
	if !ok {
		return nil, fmt.Errorf("TypeError: '%s' is not a lookup table (got %s)",
			la.TableName, typeKindName(inferTypeKind(tableVal)))
	}

	keyVal, err := ev.Eval(la.Key)
	if err != nil {
		return nil, err
	}
	serialKey, err := types.SerializeKey(keyVal)
	if err != nil {
		return nil, err
	}

	value, err := ev.Eval(la.Value)
	if err != nil {
		return nil, err
	}

	lt.Set(serialKey, value)
	return nil, nil
}

func (ev *Evaluator) evalHasExpression(he *ast.HasExpression) (Value, error) {
	tableVal, err := ev.Eval(he.Table)
	if err != nil {
		return nil, err
	}
	lt, ok := tableVal.(*LookupTableValue)
	if !ok {
		return nil, fmt.Errorf("TypeError: 'has' requires a lookup table, got %s",
			typeKindName(inferTypeKind(tableVal)))
	}

	keyVal, err := ev.Eval(he.Key)
	if err != nil {
		return nil, err
	}
	serialKey, err := types.SerializeKey(keyVal)
	if err != nil {
		return nil, err
	}

	_, exists := lt.Entries[serialKey]
	return exists, nil
}

// evalNilCheckExpression evaluates "x is something" / "x has a value" (IsSomethingCheck=true)
// and "x is nothing" / "x has no value" (IsSomethingCheck=false).
// Both always return a boolean — compatible with strict boolean conditions.
func (ev *Evaluator) evalNilCheckExpression(nc *ast.NilCheckExpression) (Value, error) {
	val, err := ev.Eval(nc.Value)
	if err != nil {
		return nil, err
	}
	if nc.IsSomethingCheck {
		return val != nil, nil // true when the variable holds a non-nothing value
	}
	return val == nil, nil // true when the variable is nothing
}
