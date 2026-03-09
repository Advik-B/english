package vm

import (
"english/vm/types"
"fmt"
)

// ─── Arithmetic ───────────────────────────────────────────────────────────────

// Add adds two values.  Strict rules:
//   - number + number   → number (arithmetic)
//   - text   + text     → text   (concatenation)
//   - array  + array    → array  (concatenation, same element type required)
//   - any other combination is a TypeError
func Add(left, right Value) (Value, error) {
switch l := left.(type) {
case float64:
r, ok := right.(float64)
if !ok {
return nil, types.NewTypeError("+", "number", typeKindName(inferTypeKind(right)))
}
return l + r, nil
case string:
r, ok := right.(string)
if !ok {
return nil, types.NewTypeError("+", "text", typeKindName(inferTypeKind(right)))
}
return l + r, nil
case *ArrayValue:
r, ok := right.(*ArrayValue)
if !ok {
return nil, types.NewTypeError("+", "array", typeKindName(inferTypeKind(right)))
}
if l.ElementType != r.ElementType {
return nil, fmt.Errorf(
"TypeError: cannot concatenate array of %s with array of %s",
typeKindName(l.ElementType), typeKindName(r.ElementType),
)
}
combined := make([]interface{}, len(l.Elements)+len(r.Elements))
copy(combined, l.Elements)
copy(combined[len(l.Elements):], r.Elements)
return &ArrayValue{ElementType: l.ElementType, Elements: combined}, nil
default:
return nil, fmt.Errorf(
"TypeError: '+' is not defined for %s",
typeKindName(inferTypeKind(left)),
)
}
}

// Subtract subtracts two numbers.
func Subtract(left, right Value) (Value, error) {
l, err := requireNumber(left, "-")
if err != nil {
return nil, err
}
r, err := requireNumber(right, "-")
if err != nil {
return nil, err
}
return l - r, nil
}

// Multiply multiplies two numbers.
func Multiply(left, right Value) (Value, error) {
l, err := requireNumber(left, "*")
if err != nil {
return nil, err
}
r, err := requireNumber(right, "*")
if err != nil {
return nil, err
}
return l * r, nil
}

// Divide divides two numbers.  Division by zero is an error.
func Divide(left, right Value) (Value, error) {
l, err := requireNumber(left, "/")
if err != nil {
return nil, err
}
r, err := requireNumber(right, "/")
if err != nil {
return nil, err
}
if r == 0 {
return nil, fmt.Errorf("RuntimeError: division by zero")
}
return l / r, nil
}

// Modulo computes the integer remainder of two numbers.
func Modulo(left, right Value) (Value, error) {
l, err := requireNumber(left, "remainder")
if err != nil {
return nil, err
}
r, err := requireNumber(right, "remainder")
if err != nil {
return nil, err
}
if r == 0 {
return nil, fmt.Errorf("RuntimeError: division by zero in remainder")
}
return float64(int64(l) % int64(r)), nil
}

// ─── Comparison ───────────────────────────────────────────────────────────────

// Compare evaluates a comparison expression and returns a boolean.
// Strict rules:
//   - "is equal to" / "is not equal to": any two values of the SAME type
//   - ordering operators: numbers only
func Compare(op string, left, right Value) (bool, error) {
switch op {
case "is equal to":
return strictEquals(left, right)
case "is not equal to":
eq, err := strictEquals(left, right)
return !eq, err
case "is less than":
return strictOrderCompare(left, right, func(a, b float64) bool { return a < b })
case "is greater than":
return strictOrderCompare(left, right, func(a, b float64) bool { return a > b })
case "is less than or equal to":
return strictOrderCompare(left, right, func(a, b float64) bool { return a <= b })
case "is greater than or equal to":
return strictOrderCompare(left, right, func(a, b float64) bool { return a >= b })
default:
return false, fmt.Errorf("unknown comparison operator: %s", op)
}
}

// strictEquals checks equality without any implicit conversion.
// Two values are equal only if they are the same type AND the same value.
// Exception: nil == nil is always true (nothing == nothing).
func strictEquals(left, right Value) (bool, error) {
if left == nil && right == nil {
return true, nil
}
if left == nil || right == nil {
return false, nil // nothing ≠ any concrete value
}
lk := types.Canonical(inferTypeKind(left))
rk := types.Canonical(inferTypeKind(right))
if lk != rk {
return false, nil // different types are never equal (no implicit conversion)
}
return Equals(left, right), nil
}

// Equals performs a same-type equality check (no type coercion).
func Equals(left, right Value) bool {
if left == nil && right == nil {
return true
}
if left == nil || right == nil {
return false
}
switch l := left.(type) {
case float64:
if r, ok := right.(float64); ok {
return l == r
}
case string:
if r, ok := right.(string); ok {
return l == r
}
case bool:
if r, ok := right.(bool); ok {
return l == r
}
}
return false
}

// strictOrderCompare applies an ordering predicate to two numbers.
func strictOrderCompare(left, right Value, pred func(float64, float64) bool) (bool, error) {
l, err := requireNumber(left, "comparison")
if err != nil {
return false, err
}
r, err := requireNumber(right, "comparison")
if err != nil {
return false, err
}
return pred(l, r), nil
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

// requireNumber unwraps any numeric Value to float64 or returns a TypeError.
func requireNumber(v Value, op string) (float64, error) {
switch val := v.(type) {
case float64:
return val, nil
case int32:
return float64(val), nil
case int64:
return float64(val), nil
case uint32:
return float64(val), nil
case uint64:
return float64(val), nil
case float32:
return float64(val), nil
default:
return 0, fmt.Errorf(
"TypeError: '%s' requires number, got %s",
op, typeKindName(inferTypeKind(v)),
)
}
}

// typeMismatchError builds a type mismatch error (kept for compatibility).
func typeMismatchError(left, right Value, op string) error {
return fmt.Errorf(
"TypeError: '%s' requires matching types, got %s and %s",
op, typeKindName(inferTypeKind(left)), typeKindName(inferTypeKind(right)),
)
}
