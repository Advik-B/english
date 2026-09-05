package runtime

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/Advik-B/english/types"
)

// sprintf is fmt.Sprintf, kept local so value.go need not import fmt.
func sprintf(format string, args ...any) string {
	return fmt.Sprintf(format, args...)
}

// ToString renders a value as text.
//
// It is used for Print and for "cast to text", and is never applied
// implicitly: arithmetic and comparison do not convert their operands.
func ToString(v Value) string {
	switch val := v.(type) {
	case nil:
		return "nothing"
	case string:
		return val
	case bool:
		if val {
			return "true"
		}
		return "false"
	case float64:
		return formatNumber(val)
	case int32:
		return strconv.FormatInt(int64(val), 10)
	case int64:
		return strconv.FormatInt(val, 10)
	case uint32:
		return strconv.FormatUint(uint64(val), 10)
	case uint64:
		return strconv.FormatUint(val, 10)
	case float32:
		return formatNumber(float64(val))
	case []any:
		// A sequence renders the same way whichever container holds it. A list
		// used to render as "[1 2 3]" and an array as "[1, 2, 3]", so the same
		// three numbers printed two different ways depending on how they were
		// declared; the disassembler and both Python back-ends had already
		// settled on the separated form.
		parts := make([]string, len(val))
		for i, el := range val {
			parts[i] = ToString(el)
		}
		return "[" + strings.Join(parts, ", ") + "]"
	case *types.ArrayValue:
		parts := make([]string, len(val.Elements))
		for i, el := range val.Elements {
			parts[i] = ToString(el)
		}
		return "[" + strings.Join(parts, ", ") + "]"
	case *types.RangeValue:
		parts := make([]string, 0, val.Length())
		for _, el := range val.ToSlice() {
			parts = append(parts, ToString(el))
		}
		return "[" + strings.Join(parts, ", ") + "]"
	case *types.LookupTableValue:
		return lookupTableString(val)
	case *types.ErrorValue:
		return "<error: " + val.Message + ">"
	case *types.TypedValue:
		return ToString(val.Value)
	}

	// Struct values render by name, and engine-specific values render
	// themselves.
	if f, ok := v.(Fielded); ok {
		return "<" + f.EnglishTypeName() + " instance>"
	}
	if d, ok := v.(Displayer); ok {
		return d.EnglishString()
	}
	return fmt.Sprintf("%v", v)
}

// formatNumber renders a number without a trailing ".0", so that whole values
// print as "5" rather than "5.0".
//
// Infinity and not-a-number are excluded from the wholeness test because
// converting them to int64 is undefined. One engine had that guard and the
// other did not, which left the other relying on undefined behaviour to
// produce the right answer.
//
// They still render the way Go writes them, "+Inf" and "NaN", which sits
// oddly beside a language that prints "true" and "nothing" in its own words
// and offers an "infinity" constant. Changing that is a decision about the
// language rather than a divergence to settle, so it is left alone here.
func formatNumber(f float64) string {
	if math.IsInf(f, 0) || math.IsNaN(f) {
		return strconv.FormatFloat(f, 'f', -1, 64)
	}
	if f == math.Trunc(f) && math.Abs(f) < 1e15 {
		return strconv.FormatInt(int64(f), 10)
	}
	return strconv.FormatFloat(f, 'f', -1, 64)
}

func lookupTableString(lt *types.LookupTableValue) string {
	if len(lt.KeyOrder) == 0 {
		return "{}"
	}
	parts := make([]string, 0, len(lt.KeyOrder))
	for _, serial := range lt.KeyOrder {
		key := serial
		if orig, _, ok := types.DeserializeKey(serial); ok {
			key = ToString(orig)
		}
		parts = append(parts, key+": "+ToString(lt.Entries[serial]))
	}
	return "{" + strings.Join(parts, ", ") + "}"
}

// ToNumber converts a value to a number.
//
// Only numeric values are accepted. Text is never converted implicitly, which
// is why "cast to number" exists.
func ToNumber(v Value) (float64, error) {
	if f, ok := asNumber(v); ok {
		return f, nil
	}
	return 0, TypeErrorf(
		"TypeError: expected number, got %s\n  Hint: use 'cast to number' to convert explicitly",
		NameOf(v))
}

// asNumber unwraps any numeric representation to a float64.
func asNumber(v Value) (float64, bool) {
	switch val := v.(type) {
	case float64:
		return val, true
	case int32:
		return float64(val), true
	case int64:
		return float64(val), true
	case uint32:
		return float64(val), true
	case uint64:
		return float64(val), true
	case float32:
		return float64(val), true
	case *types.TypedValue:
		return asNumber(val.Value)
	}
	return 0, false
}

// ToBool converts a value for use as a condition.
//
// Only booleans are accepted, plus nothing, which is falsy so that a nil check
// reads naturally. Numbers and text are not truthy: a condition must say what
// it means.
func ToBool(v Value) (bool, error) {
	switch val := v.(type) {
	case bool:
		return val, nil
	case nil:
		return false, nil
	case *types.TypedValue:
		return ToBool(val.Value)
	}
	return false, TypeErrorf(
		"TypeError: conditions must be boolean, got %s\n  Hint: use a comparison (e.g. 'x is greater than 0') or a boolean variable",
		NameOf(v))
}

// DeepCopy returns an independent copy of a value.
//
// Lists, arrays, lookup tables and structs are copied through; everything else
// is immutable or copied by value. One engine used to copy arrays and the
// other returned the same pointer, so "a copy of" aliased its source in one of
// them.
func DeepCopy(v Value) Value {
	switch val := v.(type) {
	case []any:
		out := make([]any, len(val))
		for i, el := range val {
			out[i] = DeepCopy(el)
		}
		return out
	case *types.ArrayValue:
		out := make([]any, len(val.Elements))
		for i, el := range val.Elements {
			out[i] = DeepCopy(el)
		}
		return &types.ArrayValue{ElementType: val.ElementType, Elements: out}
	case *types.LookupTableValue:
		out := types.NewLookupTable()
		for _, serial := range val.KeyOrder {
			out.Set(serial, DeepCopy(val.Entries[serial]))
		}
		return out
	case *types.TypedValue:
		return &types.TypedValue{Value: DeepCopy(val.Value), TypeInfo: val.TypeInfo}
	}
	// A struct value can only be rebuilt by the engine that defined it.
	if c, ok := v.(Copier); ok {
		return c.EnglishCopy()
	}
	return v
}
