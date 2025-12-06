package vm

import (
	"english/ast"
	"fmt"
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
