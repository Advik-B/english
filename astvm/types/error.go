package types

import "fmt"

// ErrorValue is a catchable runtime error value.
type ErrorValue struct {
	Message   string
	ErrorType string   // e.g. "TypeError", "RuntimeError"
	CallStack []string // most-recent first
}

func (e *ErrorValue) Error() string {
	result := fmt.Sprintf("%s: %s\n", e.ErrorType, e.Message)
	if len(e.CallStack) > 0 {
		result += "\nCall Stack (most recent first):\n"
		for i, frame := range e.CallStack {
			result += fmt.Sprintf("  %d. %s\n", i+1, frame)
		}
	}
	return result
}

// NewTypeError creates a clear, consistent type mismatch error.
func NewTypeError(operation, expected, got string) error {
	return fmt.Errorf("TypeError: '%s' requires %s, but got %s", operation, expected, got)
}
