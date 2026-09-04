package transpiler

import (
	"fmt"
	"math"
	"strings"

	"github.com/Advik-B/english/pygen"
)

// mathConstantMap maps the English math constants to their Python
// equivalents. Shared with the bytecode decompiler.
var mathConstantMap = pygen.MathConstants

// sanitizeIdent escapes a Python reserved word used as an identifier.
// The rule and the word list are shared with the bytecode decompiler.
func sanitizeIdent(name string) string { return pygen.SanitizeIdent(name) }

// ─── Python helper function definitions ──────────────────────────────────────
//
// These small Python functions are injected at the top of the generated file
// when the corresponding English stdlib call is used and there is no single
// Python expression that exactly reproduces the behaviour.

// The injected helper definitions and their emission order are shared with
// the bytecode decompiler, which had its own copy that had already drifted.
var (
	helperDefs  = pygen.HelperDefs
	helperOrder = pygen.HelperOrder
)

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

// maybeInt returns expr unchanged. Python list/string indices do not require
// explicit int() wrapping; using a non-integer index raises a clear TypeError.
func maybeInt(expr string) string {
	return expr
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

// typeZeroValue returns the Python zero/default value literal for a given
// English type name. Used when a struct field has no explicit default so that
// struct instances can be created with no arguments.
func typeZeroValue(typeName string) string {
	switch strings.ToLower(typeName) {
	case "number", "float", "integer", "int", "unsigned integer":
		return "0"
	case "text", "string":
		return `""`
	case "boolean", "bool":
		return "False"
	default:
		return "None"
	}
}
