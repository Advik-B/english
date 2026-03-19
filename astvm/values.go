package vm

import (
	"github.com/Advik-B/english/ast"
	"github.com/Advik-B/english/astvm/types"
	"fmt"
)

// Value is the universal runtime value interface for the English language.
// Every evaluated expression produces a Value.
type Value = interface{}

// BuiltinFunc is the signature for the injected standard-library evaluator.
// It is called whenever a built-in function (Body == nil) is invoked.
type BuiltinFunc func(name string, args []Value) (Value, error)

// FunctionValue represents a user-defined function.
// It lives in vm/ (not vm/types/) because it holds a *Environment closure.
type FunctionValue struct {
	Name       string
	Parameters []string
	Body       []ast.Statement
	Closure    *Environment
}

func (f *FunctionValue) String() string {
	return fmt.Sprintf("<function %s>", f.Name)
}

// ─── Control-flow sentinels ───────────────────────────────────────────────────

// ReturnValue wraps a function's return payload.
type ReturnValue struct{ Value Value }

// BreakValue signals a loop break.
type BreakValue struct{}

// ContinueValue signals a loop continue.
type ContinueValue struct{}

// ─── Runtime error (non-catchable) ───────────────────────────────────────────

// RuntimeError is a non-catchable interpreter error with an optional call stack.
type RuntimeError struct {
	Message   string
	CallStack []string
	Line      int // source line where the error occurred (0 = unknown)
}

func (e *RuntimeError) Error() string {
	result := fmt.Sprintf("Runtime Error: %s\n", e.Message)
	if e.Line > 0 {
		result = fmt.Sprintf("Runtime Error at line %d: %s\n", e.Line, e.Message)
	}
	if len(e.CallStack) > 0 {
		result += "\nCall Stack (most recent first):\n"
		for i, frame := range e.CallStack {
			result += fmt.Sprintf("  %d. %s\n", i+1, frame)
		}
	}
	return result
}

// RuntimeMessage implements the stacktraces.RuntimeError interface and returns
// the human-readable error message without call-stack details.
func (e *RuntimeError) RuntimeMessage() string { return e.Message }

// RuntimeLine implements the stacktraces.RuntimeError interface and returns
// the source line where the error occurred (0 = unknown).
func (e *RuntimeError) RuntimeLine() int { return e.Line }

// RuntimeCallStack implements the stacktraces.RuntimeError interface and
// returns the full call-stack slice (most-recent frame first).
func (e *RuntimeError) RuntimeCallStack() []string {
	if e.CallStack == nil {
		return []string{}
	}
	return e.CallStack
}

// NewRuntimeError creates a RuntimeError with a default stdlib call-stack frame.
func NewRuntimeError(message string) error {
	return &RuntimeError{Message: message, CallStack: []string{"<stdlib>"}}
}

// ─── Reference value ─────────────────────────────────────────────────────────

// ReferenceValue holds a reference to a named variable in a specific environment.
// Lives in vm/ because it references *Environment.
type ReferenceValue struct {
	Name string
	Env  *Environment
}

// ─── Type aliases for vm/types composite types ───────────────────────────────

// ArrayValue is re-exported from vm/types for convenience within vm/.
type ArrayValue = types.ArrayValue

// LookupTableValue is re-exported from vm/types for convenience within vm/.
type LookupTableValue = types.LookupTableValue

// RangeValue is re-exported from vm/types for convenience within vm/.
type RangeValue = types.RangeValue
