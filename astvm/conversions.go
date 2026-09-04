package vm

import (
	"github.com/Advik-B/english/runtime"
)

// Conversions are defined in the runtime package, which both engines share.
// These are the names this engine has always exposed.

// ToString renders a value as text, for Print and for "cast to text". It is
// never applied implicitly.
func ToString(v Value) string { return runtime.ToString(v) }

// ToNumber converts a value to a number. Text is not converted implicitly;
// "cast to number" is how that is asked for.
func ToNumber(v Value) (float64, error) { return runtime.ToNumber(v) }

// ToBool converts a value for use as a condition. Only booleans and nothing
// are accepted: numbers and text are not truthy.
func ToBool(v Value) (bool, error) { return runtime.ToBool(v) }
