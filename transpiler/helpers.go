package transpiler

import (
	"fmt"
	"math"
	"strings"
)

// mathConstantMap maps English stdlib math constants (registered as environment
// variables) to their Python equivalents. When an Identifier with one of these
// names is encountered, math is imported and the Python name is emitted.
var mathConstantMap = map[string]string{
	"pi":       "math.pi",
	"e":        "math.e",
	"infinity": "math.inf",
}

// ─── Python helper function definitions ──────────────────────────────────────
//
// These small Python functions are injected at the top of the generated file
// when the corresponding English stdlib call is used and there is no single
// Python expression that exactly reproduces the behaviour.

// helperDefs maps a helper name to its Python source (no trailing newline).
var helperDefs = map[string]string{
	"_table_remove": `def _table_remove(d, k):
    result = dict(d)
    result.pop(k, None)
    return result`,

	"_flatten": `def _flatten(lst):
    return [item for sublist in lst for item in sublist]`,

	"_read_file": `def _read_file(path):
    with open(path, "r") as f:
        return f.read()`,

	"_write_file": `def _write_file(path, content):
    with open(path, "w") as f:
        f.write(str(content))`,

	"_is_nan": `def _is_nan(x):
    try:
        return math.isnan(float(x))
    except (TypeError, ValueError):
        return True`,

	"_is_infinite": `def _is_infinite(x):
    try:
        return math.isinf(float(x))
    except (TypeError, ValueError):
        return False`,

	"_sign": `def _sign(x):
    if x > 0:
        return 1
    elif x < 0:
        return -1
    return 0`,

	"_unique": `def _unique(lst):
    seen = []
    for item in lst:
        if item not in seen:
            seen.append(item)
    return seen`,

	"_product": `def _product(lst):
    result = 1
    for item in lst:
        result *= item
    return result`,

	"_zip_with": `def _zip_with(a, b):
    return [[x, y] for x, y in zip(a, b)]`,
}

// helperOrder defines the deterministic emission order for helper functions.
var helperOrder = []string{
	"_table_remove",
	"_flatten",
	"_read_file",
	"_write_file",
	"_is_nan",
	"_is_infinite",
	"_sign",
	"_unique",
	"_product",
	"_zip_with",
}

// ─── Numeric literal formatting ───────────────────────────────────────────────

// formatNumber renders a float64 as a compact Python numeric literal.
// Whole numbers are emitted without a decimal point (e.g. 5, not 5.0).
func formatNumber(v float64) string {
	if math.IsInf(v, 1) {
		return "float('inf')"
	}
	if math.IsInf(v, -1) {
		return "float('-inf')"
	}
	if math.IsNaN(v) {
		return "float('nan')"
	}
	if v == math.Trunc(v) && math.Abs(v) < 1e15 {
		return fmt.Sprintf("%d", int64(v))
	}
	return fmt.Sprintf("%g", v)
}

// isIntegerLiteral returns true when an expression is guaranteed to be an
// integer-valued number literal at parse time. Used to avoid redundant int()
// wrapping of constants like 0, 1, 2 …
func isIntegerLiteral(s string) bool {
	for _, c := range s {
		if c == '.' || c == 'e' || c == 'E' || c == '\'' {
			return false
		}
	}
	return len(s) > 0 && (s[0] >= '0' && s[0] <= '9' || s[0] == '-')
}

// maybeInt wraps expr in int() only when it is not already an integer literal.
// This keeps the generated code readable: range(5) instead of range(int(5)).
func maybeInt(expr string) string {
	if isIntegerLiteral(expr) {
		return expr
	}
	return fmt.Sprintf("int(%s)", expr)
}

// ─── Operator / type name mapping ────────────────────────────────────────────

// mapOperator converts an English operator string to the Python equivalent.
func mapOperator(op string) string {
	switch op {
	case "+":
		return "+"
	case "-":
		return "-"
	case "*":
		return "*"
	case "/":
		return "/"
	case "%", "remainder":
		return "%"
	case "**":
		return "**"
	case "is equal to", "==":
		return "=="
	case "is not equal to", "!=":
		return "!="
	case "is less than", "<":
		return "<"
	case "is greater than", ">":
		return ">"
	case "is less than or equal to", "<=":
		return "<="
	case "is greater than or equal to", ">=":
		return ">="
	case "and":
		return "and"
	case "or":
		return "or"
	default:
		return op
	}
}

// mapTypeName converts an English type name to a Python type annotation string.
// Note: Python's int is signed and arbitrarily precise. "unsigned integer" maps
// to int because Python has no separate unsigned integer type.
func mapTypeName(name string) string {
	switch strings.ToLower(name) {
	case "number", "float":
		return "float"
	case "integer", "int", "unsigned integer":
		// Python int is arbitrarily large and signed; it is the closest equivalent.
		return "int"
	case "text", "string":
		return "str"
	case "boolean", "bool":
		return "bool"
	case "list", "array":
		return "list"
	default:
		return name
	}
}
