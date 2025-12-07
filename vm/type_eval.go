package vm

import (
	"english/ast"
	"fmt"
)

// evalTryStatement evaluates a try/error/finally block
func (ev *Evaluator) evalTryStatement(node *ast.TryStatement) (Value, error) {
	var tryResult Value
	var tryError error

	// Execute try block
	for _, stmt := range node.TryBody {
		val, err := ev.Eval(stmt)
		if err != nil {
			tryError = err
			break
		}
		if _, ok := val.(*ReturnValue); ok {
			tryResult = val
			break
		}
		if _, ok := val.(*BreakValue); ok {
			tryResult = val
			break
		}
		tryResult = val
	}

	// Execute error handler if there was an error
	if tryError != nil && len(node.ErrorBody) > 0 {
		// Convert error to ErrorValue
		var errorVal *ErrorValue
		if errVal, ok := tryError.(*ErrorValue); ok {
			errorVal = errVal
		} else if re, ok := tryError.(*RuntimeError); ok {
			errorVal = &ErrorValue{
				Message:   re.Message,
				ErrorType: "RuntimeError",
				CallStack: re.CallStack,
			}
		} else {
			errorVal = &ErrorValue{
				Message:   tryError.Error(),
				ErrorType: "RuntimeError",
				CallStack: append([]string{}, ev.callStack...),
			}
		}

		// Bind error to variable in error handler scope
		errorEnv := ev.env.NewChild()
		errorEnv.Define(node.ErrorVar, errorVal, false)

		// Save current environment and switch to error environment
		oldEnv := ev.env
		ev.env = errorEnv

		// Execute error handler
		for _, stmt := range node.ErrorBody {
			val, err := ev.Eval(stmt)
			if err != nil {
				// Error in error handler - restore environment and execute finally
				ev.env = oldEnv
				if len(node.FinallyBody) > 0 {
					ev.executeFinallyBlock(node.FinallyBody)
				}
				return nil, err
			}
			if _, ok := val.(*ReturnValue); ok {
				tryResult = val
				break
			}
			if _, ok := val.(*BreakValue); ok {
				tryResult = val
				break
			}
			tryResult = val
		}

		// Restore environment
		ev.env = oldEnv

		// Clear the error since it was handled
		tryError = nil
	}

	// Execute finally block
	if len(node.FinallyBody) > 0 {
		ev.executeFinallyBlock(node.FinallyBody)
	}

	// If there was an unhandled error, return it
	if tryError != nil {
		return nil, tryError
	}

	return tryResult, nil
}

// executeFinallyBlock executes the finally block (ignoring errors)
func (ev *Evaluator) executeFinallyBlock(finallyBody []ast.Statement) {
	for _, stmt := range finallyBody {
		ev.Eval(stmt) // Ignore errors in finally block
	}
}

// evalRaiseStatement evaluates a raise statement
func (ev *Evaluator) evalRaiseStatement(node *ast.RaiseStatement) (Value, error) {
	// Evaluate the message
	msgVal, err := ev.Eval(node.Message)
	if err != nil {
		return nil, err
	}

	message := ToString(msgVal)

	// Create and return error value
	return nil, &ErrorValue{
		Message:   message,
		ErrorType: node.ErrorType,
		CallStack: append([]string{}, ev.callStack...),
	}
}

// evalSwapStatement evaluates a swap statement
func (ev *Evaluator) evalSwapStatement(node *ast.SwapStatement) (Value, error) {
	// Get first variable
	val1, ok := ev.env.Get(node.Name1)
	if !ok {
		return nil, ev.runtimeError(fmt.Sprintf("undefined variable '%s'", node.Name1))
	}

	// Get second variable
	val2, ok := ev.env.Get(node.Name2)
	if !ok {
		return nil, ev.runtimeError(fmt.Sprintf("undefined variable '%s'", node.Name2))
	}

	// Swap values
	if err := ev.env.Set(node.Name1, val2); err != nil {
		return nil, ev.runtimeError(err.Error())
	}

	if err := ev.env.Set(node.Name2, val1); err != nil {
		return nil, ev.runtimeError(err.Error())
	}

	return nil, nil
}

// evalTypeExpression evaluates getting the type of a value
func (ev *Evaluator) evalTypeExpression(node *ast.TypeExpression) (Value, error) {
	// Evaluate the value
	val, err := ev.Eval(node.Value)
	if err != nil {
		return nil, err
	}

	// Get type info
	typeInfo := GetType(val)

	// Return type name as string
	return typeInfo.String(), nil
}

// evalCastExpression evaluates casting a value to a type
func (ev *Evaluator) evalCastExpression(node *ast.CastExpression) (Value, error) {
	// Evaluate the value to cast
	val, err := ev.Eval(node.Value)
	if err != nil {
		return nil, err
	}

	// Parse target type
	targetType := ParseTypeString(node.TypeName)

	// Attempt to cast
	result, err := CastValue(val, targetType)
	if err != nil {
		return nil, &ErrorValue{
			Message:   err.Error(),
			ErrorType: "TypeError",
			CallStack: append([]string{}, ev.callStack...),
		}
	}

	return result, nil
}

// evalReferenceExpression evaluates creating a reference to a variable
func (ev *Evaluator) evalReferenceExpression(node *ast.ReferenceExpression) (Value, error) {
	// Check if variable exists
	_, ok := ev.env.Get(node.Name)
	if !ok {
		return nil, ev.runtimeError(fmt.Sprintf("undefined variable '%s'", node.Name))
	}

	// Create a reference value
	return &ReferenceValue{
		Name: node.Name,
		Env:  ev.env,
	}, nil
}

// evalCopyExpression evaluates creating a deep copy of a value
func (ev *Evaluator) evalCopyExpression(node *ast.CopyExpression) (Value, error) {
	// Evaluate the value to copy
	val, err := ev.Eval(node.Value)
	if err != nil {
		return nil, err
	}

	// Perform deep copy based on type
	return deepCopy(val), nil
}

// deepCopy performs a deep copy of a value
func deepCopy(val Value) Value {
	switch v := val.(type) {
	case []interface{}:
		// Deep copy list
		copied := make([]interface{}, len(v))
		for i, elem := range v {
			copied[i] = deepCopy(elem)
		}
		return copied
	case *StructInstance:
		// Deep copy struct instance
		copiedFields := make(map[string]Value)
		for fieldName, fieldVal := range v.Fields {
			copiedFields[fieldName] = deepCopy(fieldVal)
		}
		return &StructInstance{
			Definition: v.Definition,
			Fields:     copiedFields,
		}
	case *TypedValue:
		// Deep copy typed value
		return &TypedValue{
			Value:    deepCopy(v.Value),
			TypeInfo: v.TypeInfo,
		}
	default:
		// For primitive types, just return the value (they're immutable or copied by value)
		return val
	}
}
