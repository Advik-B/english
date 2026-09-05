package runtime

import (
	"math"

	"github.com/Advik-B/english/types"
)

// Add adds two values.
//
//	number + number → number      (arithmetic)
//	text   + text   → text        (concatenation)
//	array  + array  → array       (concatenation, matching element types)
//
// Anything else is a type error. Note that this accepts the same numeric
// representations as every other arithmetic operation: it used to switch on
// float64 alone while subtraction accepted all of them, so after a cast to a
// sized integer `a - b` worked and `a + b` did not.
func Add(left, right Value) (Value, error) {
	if l, ok := asNumber(left); ok {
		r, ok := asNumber(right)
		if !ok {
			return nil, TypeErrorf("TypeError: '+' requires number, but got %s", NameOf(right))
		}
		return l + r, nil
	}
	switch l := left.(type) {
	case string:
		r, ok := right.(string)
		if !ok {
			return nil, TypeErrorf("TypeError: '+' requires text, but got %s", NameOf(right))
		}
		return l + r, nil
	case *types.ArrayValue:
		r, ok := right.(*types.ArrayValue)
		if !ok {
			return nil, TypeErrorf("TypeError: '+' requires array, but got %s", NameOf(right))
		}
		if l.ElementType != r.ElementType {
			return nil, TypeErrorf("TypeError: cannot concatenate array of %s with array of %s",
				types.Name(l.ElementType), types.Name(r.ElementType))
		}
		combined := make([]any, 0, len(l.Elements)+len(r.Elements))
		combined = append(combined, l.Elements...)
		combined = append(combined, r.Elements...)
		return &types.ArrayValue{ElementType: l.ElementType, Elements: combined}, nil
	}
	return nil, TypeErrorf("TypeError: '+' is not defined for %s", NameOf(left))
}

// Subtract, Multiply, Divide and Modulo operate on numbers.

func Subtract(left, right Value) (Value, error) {
	l, r, err := twoNumbers(left, right, "-")
	if err != nil {
		return nil, err
	}
	return l - r, nil
}

func Multiply(left, right Value) (Value, error) {
	l, r, err := twoNumbers(left, right, "*")
	if err != nil {
		return nil, err
	}
	return l * r, nil
}

func Divide(left, right Value) (Value, error) {
	l, r, err := twoNumbers(left, right, "/")
	if err != nil {
		return nil, err
	}
	if r == 0 {
		return nil, TypeErrorf("division by zero")
	}
	return l / r, nil
}

// Modulo is the remainder after truncating division, so it takes the sign of
// the left operand: the remainder of -7 divided by 3 is -1.
func Modulo(left, right Value) (Value, error) {
	l, r, err := twoNumbers(left, right, "remainder")
	if err != nil {
		return nil, err
	}
	if r == 0 {
		return nil, TypeErrorf("division by zero in remainder")
	}
	return math.Mod(math.Trunc(l), math.Trunc(r)), nil
}

// twoNumbers unwraps both operands or reports which one was wrong.
func twoNumbers(left, right Value, op string) (float64, float64, error) {
	l, ok := asNumber(left)
	if !ok {
		return 0, 0, TypeErrorf("TypeError: '%s' requires number, got %s", op, NameOf(left))
	}
	r, ok := asNumber(right)
	if !ok {
		return 0, 0, TypeErrorf("TypeError: '%s' requires number, got %s", op, NameOf(right))
	}
	return l, r, nil
}

// Negate applies unary minus.
func Negate(v Value) (Value, error) {
	n, ok := asNumber(v)
	if !ok {
		return nil, TypeErrorf("TypeError: '-' requires number, got %s", NameOf(v))
	}
	return -n, nil
}

// Not applies logical negation.
func Not(v Value) (Value, error) {
	b, err := ToBool(v)
	if err != nil {
		return nil, err
	}
	return !b, nil
}

// ─── Comparison ──────────────────────────────────────────────────────────────

// Compare evaluates a comparison operator.
//
// Equality accepts any two values of the same type; the ordering operators
// accept numbers only. Nothing is converted: comparing a number with text is
// false rather than a coincidence of representation.
func Compare(op string, left, right Value) (bool, error) {
	switch op {
	case "is equal to":
		return Equals(left, right), nil
	case "is not equal to":
		return !Equals(left, right), nil
	case "is less than":
		return order(left, right, func(a, b float64) bool { return a < b })
	case "is greater than":
		return order(left, right, func(a, b float64) bool { return a > b })
	case "is less than or equal to":
		return order(left, right, func(a, b float64) bool { return a <= b })
	case "is greater than or equal to":
		return order(left, right, func(a, b float64) bool { return a >= b })
	}
	return false, TypeErrorf("unknown comparison operator: %s", op)
}

func order(left, right Value, pred func(float64, float64) bool) (bool, error) {
	l, r, err := twoNumbers(left, right, "comparison")
	if err != nil {
		return false, err
	}
	return pred(l, r), nil
}

// Equals reports whether two values are the same type and the same value.
//
// It is total: every kind of value can be compared. It used to switch on
// number, text and boolean alone and return false for everything else, so two
// identical lists were unequal, two identical structs were unequal, and — via
// the canonical-kind check that let them through — two equal sized integers
// were unequal too.
func Equals(left, right Value) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	if tv, ok := left.(*types.TypedValue); ok {
		return Equals(tv.Value, right)
	}
	if tv, ok := right.(*types.TypedValue); ok {
		return Equals(left, tv.Value)
	}

	// Different types are never equal, with no conversion attempted.
	if types.Canonical(types.Infer(left)) != types.Canonical(types.Infer(right)) {
		return false
	}

	// Numbers compare by value, across representations.
	if l, ok := asNumber(left); ok {
		r, _ := asNumber(right)
		return l == r
	}

	switch l := left.(type) {
	case string:
		r, ok := right.(string)
		return ok && l == r
	case bool:
		r, ok := right.(bool)
		return ok && l == r
	case []any:
		r, ok := right.([]any)
		return ok && sameElements(l, r)
	case *types.ArrayValue:
		r, ok := right.(*types.ArrayValue)
		return ok && l.ElementType == r.ElementType && sameElements(l.Elements, r.Elements)
	case *types.RangeValue:
		r, ok := right.(*types.RangeValue)
		return ok && sameElements(l.ToSlice(), r.ToSlice())
	case *types.LookupTableValue:
		r, ok := right.(*types.LookupTableValue)
		return ok && sameEntries(l, r)
	case *types.ErrorValue:
		r, ok := right.(*types.ErrorValue)
		return ok && l.ErrorType == r.ErrorType && l.Message == r.Message
	}

	// Structs compare by type and field values.
	if l, ok := left.(Fielded); ok {
		r, ok := right.(Fielded)
		if !ok || l.EnglishTypeName() != r.EnglishTypeName() {
			return false
		}
		lf, rf := l.EnglishFields(), r.EnglishFields()
		if len(lf) != len(rf) {
			return false
		}
		for name, lv := range lf {
			rv, present := rf[name]
			if !present || !Equals(lv, rv) {
				return false
			}
		}
		return true
	}

	// Function values and references compare by identity.
	return left == right
}

func sameElements(a, b []any) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if !Equals(a[i], b[i]) {
			return false
		}
	}
	return true
}

func sameEntries(a, b *types.LookupTableValue) bool {
	if len(a.Entries) != len(b.Entries) {
		return false
	}
	for key, av := range a.Entries {
		bv, present := b.Entries[key]
		if !present || !Equals(av, bv) {
			return false
		}
	}
	return true
}
