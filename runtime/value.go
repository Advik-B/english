// Package runtime holds the value semantics of the English language: what a
// value is, how it converts, and what the operators do.
//
// Both execution engines call into here. They previously each had their own
// copy of all of it, and the copies had drifted: the two disagreed on the
// length of a non-ASCII string, on whether text could be indexed, on whether
// "copy of" copied an array, on what a type is called in an error message, and
// on the wording of nearly every diagnostic. The parity suite compared only
// stdout, so none of it was caught.
package runtime

import (
	"github.com/Advik-B/english/types"
)

// Value is any English value.
//
// It is an alias for any rather than a tagged union, which is what
// makes the language dynamically typed underneath: nothing about a value's Go
// representation constrains what may be stored where. Static analysis is what
// keeps programs honest; these functions are the last line of defence.
type Value = any

// Displayer is implemented by engine-specific values that know how to render
// themselves for Print and for "cast to text".
//
// A function value is the obvious case: one engine holds an AST body and the
// other a bytecode chunk, so only the engine can name it, but both must print
// it the same way.
type Displayer interface {
	EnglishString() string
}

// Fielded is implemented by struct values, so that they can be compared,
// copied and printed without this package knowing either engine's definition
// record.
type Fielded interface {
	// EnglishTypeName is the declared struct name.
	EnglishTypeName() string
	// EnglishFields returns the live field map, which callers may read but
	// must not resize.
	EnglishFields() map[string]Value
}

// Copier is implemented by values that can only be duplicated by the engine
// that created them, which is every struct value: this package can read a
// struct's fields through Fielded but cannot construct an engine's definition
// record to build a new instance around.
type Copier interface {
	EnglishCopy() Value
}

// ─── Errors ──────────────────────────────────────────────────────────────────

// TypeErrorf builds the standard type-error message shape, so that both
// engines word the same failure identically.
func TypeErrorf(format string, args ...any) error {
	return &OperationError{Message: sprintf(format, args...)}
}

// OperationError is a failure in a value operation: a type mismatch, a
// division by zero, an index out of range.
//
// It carries no source position. The engines add that, because only they know
// which statement was running.
type OperationError struct {
	Message string
}

func (e *OperationError) Error() string { return e.Message }

// NameOf is the user-facing name of a value's type, for diagnostics.
func NameOf(v Value) string { return types.NameOf(v) }
