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

// maybeInt wraps an expression in int() where Python needs a whole number.
//
// English has one number type and it is a float, so an index, a length, a
// repeat count and a range bound all arrive as floats. Python requires an int
// for every one of those: "xs[1.0]" is a TypeError and "range(1.0)" is a
// TypeError, so a program that ran in English failed as soon as it was
// transpiled. This used to return its argument untouched, with a doc comment
// asserting that no wrapping was needed, at thirteen call sites — while the
// bytecode decompiler, the other Python back-end, wrapped correctly.
//
// An expression already known to be a whole number is left alone, so the
// output stays readable: int(int(x)) and int(len(xs)) are noise.
func maybeInt(expr string) string {
	if alreadyInt(expr) {
		return expr
	}
	return "int(" + expr + ")"
}

// alreadyInt reports whether an expression is certainly a Python int.
func alreadyInt(expr string) bool {
	expr = strings.TrimSpace(expr)
	if isIntegerLiteral(expr) {
		return true
	}
	for _, wrapper := range []string{"int(", "len(", "_round("} {
		if wrapsWhole(expr, wrapper) {
			return true
		}
	}
	return false
}

// wrapsWhole reports whether expr is a single call to the given function —
// "len(xs)" but not "len(xs) + len(ys)", whose parentheses close early.
func wrapsWhole(expr, open string) bool {
	if !strings.HasPrefix(expr, open) || !strings.HasSuffix(expr, ")") {
		return false
	}
	depth := 0
	for i := len(open) - 1; i < len(expr); i++ {
		switch expr[i] {
		case '(':
			depth++
		case ')':
			depth--
			if depth == 0 {
				return i == len(expr)-1
			}
		}
	}
	return false
}

// ─── Operator / type name mapping ────────────────────────────────────────────

// isRemainder reports whether an operator is English's remainder, which needs
// a helper rather than Python's %.
func isRemainder(op string) bool {
	return op == "%" || op == "remainder"
}

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
	// The remainder is not here: Python's % is a different operation, so it
	// goes through a helper. See isRemainder.
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
//
// It matches what the interpreter starts such a field as, which for a
// collection is an empty one — this used to answer None for those, so a field
// declared as a list began as nothing in Python and as an empty list in
// English, and the first thing done to it failed.
func typeZeroValue(typeName string) string {
	switch strings.ToLower(typeName) {
	case "number", "float", "integer", "int", "unsigned integer":
		return "0"
	case "text", "string":
		return `""`
	case "boolean", "bool":
		return "False"
	case "list", "array":
		return "[]"
	case "lookup table", "table":
		return "{}"
	default:
		return "None"
	}
}
