package transpiler

import (
	"english/ast"
	"fmt"
	"strings"
)

// ─── stdlib function call translation ────────────────────────────────────────
//
// transpileFuncCallExpr maps every English standard-library function name to
// an idiomatic Python expression. User-defined functions are emitted as plain
// calls at the end of the switch.

func (t *Transpiler) transpileFuncCallExpr(e *ast.FunctionCall) string {
	args := make([]string, len(e.Arguments))
	for i, a := range e.Arguments {
		args[i] = t.transpileExpr(a)
	}
	// Helper so callers can grab an arg by index without bounds-checking.
	a := func(i int) string {
		if i < len(args) {
			return args[i]
		}
		return "None"
	}

	// User-defined functions take priority over any stdlib mapping with the
	// same name (e.g. a user can define their own "average" that takes numbers
	// rather than a list).
	if t.userFunctions[e.Name] {
		return fmt.Sprintf("%s(%s)", sanitizeIdent(e.Name), strings.Join(args, ", "))
	}

	switch e.Name {
	// ── Math ──────────────────────────────────────────────────────────────────
	case "sqrt":
		return fmt.Sprintf("math.sqrt(%s)", a(0))
	case "pow":
		return fmt.Sprintf("math.pow(%s, %s)", a(0), a(1))
	case "abs":
		return fmt.Sprintf("abs(%s)", a(0))
	case "floor":
		return fmt.Sprintf("math.floor(%s)", a(0))
	case "ceil":
		return fmt.Sprintf("math.ceil(%s)", a(0))
	case "round":
		return fmt.Sprintf("round(%s)", a(0))
	case "min":
		return fmt.Sprintf("min(%s, %s)", a(0), a(1))
	case "max":
		return fmt.Sprintf("max(%s, %s)", a(0), a(1))
	case "sin":
		return fmt.Sprintf("math.sin(%s)", a(0))
	case "cos":
		return fmt.Sprintf("math.cos(%s)", a(0))
	case "tan":
		return fmt.Sprintf("math.tan(%s)", a(0))
	case "log":
		return fmt.Sprintf("math.log(%s)", a(0))
	case "log10":
		return fmt.Sprintf("math.log10(%s)", a(0))
	case "log2":
		return fmt.Sprintf("math.log2(%s)", a(0))
	case "exp":
		return fmt.Sprintf("math.exp(%s)", a(0))
	case "random":
		return "random.random()"
	case "random_between":
		return fmt.Sprintf("random.uniform(%s, %s)", a(0), a(1))
	case "is_nan":
		return fmt.Sprintf("_is_nan(%s)", a(0))
	case "is_infinite":
		return fmt.Sprintf("_is_infinite(%s)", a(0))
	case "is_integer":
		// float.is_integer() is a Python built-in method
		return fmt.Sprintf("float(%s).is_integer()", a(0))
	case "clamp":
		// clamp(x, min, max) → max(min_val, min(max_val, x))
		return fmt.Sprintf("max(%s, min(%s, %s))", a(1), a(2), a(0))
	case "sign":
		return fmt.Sprintf("_sign(%s)", a(0))

	// ── String ────────────────────────────────────────────────────────────────
	case "uppercase":
		return fmt.Sprintf("%s.upper()", a(0))
	case "lowercase":
		return fmt.Sprintf("%s.lower()", a(0))
	case "casefold":
		return fmt.Sprintf("%s.casefold()", a(0))
	case "title":
		return fmt.Sprintf("%s.title()", a(0))
	case "capitalize":
		return fmt.Sprintf("%s.capitalize()", a(0))
	case "swapcase":
		return fmt.Sprintf("%s.swapcase()", a(0))
	case "split":
		return fmt.Sprintf("%s.split(%s)", a(0), a(1))
	case "join":
		// join(list, sep) → sep.join(list)
		return fmt.Sprintf("%s.join(%s)", a(1), a(0))
	case "trim":
		return fmt.Sprintf("%s.strip()", a(0))
	case "trim_left":
		return fmt.Sprintf("%s.lstrip()", a(0))
	case "trim_right":
		return fmt.Sprintf("%s.rstrip()", a(0))
	case "replace":
		return fmt.Sprintf("%s.replace(%s, %s)", a(0), a(1), a(2))
	case "contains":
		return fmt.Sprintf("(%s in %s)", a(1), a(0))
	case "starts_with":
		return fmt.Sprintf("%s.startswith(%s)", a(0), a(1))
	case "ends_with":
		return fmt.Sprintf("%s.endswith(%s)", a(0), a(1))
	case "index_of":
		return fmt.Sprintf("%s.find(%s)", a(0), a(1))
	case "substring":
		// substring(s, start, length) → s[start : start+length]
		start := a(1)
		length := a(2)
		return fmt.Sprintf("%s[%s:%s+%s]", a(0), maybeInt(start), maybeInt(start), maybeInt(length))
	case "str_repeat":
		return fmt.Sprintf("%s * %s", a(0), maybeInt(a(1)))
	case "count_occurrences":
		return fmt.Sprintf("%s.count(%s)", a(0), a(1))
	case "pad_left":
		if len(args) > 2 {
			return fmt.Sprintf("%s.rjust(%s, %s)", a(0), maybeInt(a(1)), a(2))
		}
		return fmt.Sprintf("%s.rjust(%s)", a(0), maybeInt(a(1)))
	case "pad_right":
		if len(args) > 2 {
			return fmt.Sprintf("%s.ljust(%s, %s)", a(0), maybeInt(a(1)), a(2))
		}
		return fmt.Sprintf("%s.ljust(%s)", a(0), maybeInt(a(1)))
	case "center":
		if len(args) > 2 {
			return fmt.Sprintf("%s.center(%s, %s)", a(0), maybeInt(a(1)), a(2))
		}
		return fmt.Sprintf("%s.center(%s)", a(0), maybeInt(a(1)))
	case "zfill":
		return fmt.Sprintf("%s.zfill(%s)", a(0), maybeInt(a(1)))
	case "to_number":
		return fmt.Sprintf("float(%s)", a(0))
	case "to_string":
		return fmt.Sprintf("str(%s)", a(0))
	case "is_empty":
		return fmt.Sprintf("(len(%s) == 0)", a(0))
	case "is_digit":
		return fmt.Sprintf("%s.isdigit()", a(0))
	case "is_alpha":
		return fmt.Sprintf("%s.isalpha()", a(0))
	case "is_alnum":
		return fmt.Sprintf("%s.isalnum()", a(0))
	case "is_space":
		return fmt.Sprintf("%s.isspace()", a(0))
	case "is_upper":
		return fmt.Sprintf("%s.isupper()", a(0))
	case "is_lower":
		return fmt.Sprintf("%s.islower()", a(0))

	// ── List ──────────────────────────────────────────────────────────────────
	case "count":
		return fmt.Sprintf("len(%s)", a(0))
	case "first":
		return fmt.Sprintf("%s[0]", a(0))
	case "last":
		return fmt.Sprintf("%s[-1]", a(0))
	case "sort":
		return fmt.Sprintf("sorted(%s)", a(0))
	case "sorted_desc":
		return fmt.Sprintf("sorted(%s, reverse=True)", a(0))
	case "reverse":
		return fmt.Sprintf("list(reversed(%s))", a(0))
	case "append":
		// append(list, item) → list + [item]
		return fmt.Sprintf("%s + [%s]", a(0), a(1))
	case "pop":
		return fmt.Sprintf("%s[:-1]", a(0))
	case "remove":
		// remove(list, index) → list without element at index
		return fmt.Sprintf("[v for i, v in enumerate(%s) if i != %s]", a(0), maybeInt(a(1)))
	case "insert":
		// insert(list, index, item) → new list with item inserted at index
		return fmt.Sprintf("(%s[:%s] + [%s] + %s[%s:])", a(0), maybeInt(a(1)), a(2), a(0), maybeInt(a(1)))
	case "sum":
		return fmt.Sprintf("sum(%s)", a(0))
	case "product":
		return fmt.Sprintf("_product(%s)", a(0))
	case "average":
		return fmt.Sprintf("(sum(%s) / len(%s))", a(0), a(0))
	case "min_value":
		return fmt.Sprintf("min(%s)", a(0))
	case "max_value":
		return fmt.Sprintf("max(%s)", a(0))
	case "any_true":
		return fmt.Sprintf("any(%s)", a(0))
	case "all_true":
		return fmt.Sprintf("all(%s)", a(0))
	case "unique":
		return fmt.Sprintf("_unique(%s)", a(0))
	case "flatten":
		return fmt.Sprintf("_flatten(%s)", a(0))
	case "slice":
		// slice(list, start, end) → list[start:end]
		return fmt.Sprintf("%s[%s:%s]", a(0), maybeInt(a(1)), maybeInt(a(2)))
	case "zip_with":
		return fmt.Sprintf("_zip_with(%s, %s)", a(0), a(1))

	// ── Lookup table ──────────────────────────────────────────────────────────
	case "keys":
		return fmt.Sprintf("list(%s.keys())", a(0))
	case "values":
		return fmt.Sprintf("list(%s.values())", a(0))
	case "table_remove":
		return fmt.Sprintf("_table_remove(%s, %s)", a(0), a(1))
	case "table_has":
		return fmt.Sprintf("(%s in %s)", a(1), a(0))
	case "merge":
		// merge(table, other) → {**table, **other}
		return fmt.Sprintf("{**%s, **%s}", a(0), a(1))
	case "get_or_default":
		return fmt.Sprintf("%s.get(%s, %s)", a(0), a(1), a(2))

	// ── I/O ───────────────────────────────────────────────────────────────────
	case "ask":
		return fmt.Sprintf("input(%s)", a(0))
	case "read_file":
		return fmt.Sprintf("_read_file(%s)", a(0))
	case "write_file":
		return fmt.Sprintf("_write_file(%s, %s)", a(0), a(1))
	}

	// Unknown / user-defined function — emit a direct call.
	return fmt.Sprintf("%s(%s)", sanitizeIdent(e.Name), strings.Join(args, ", "))
}
