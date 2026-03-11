package ivm

import (
	"fmt"
	"strings"
)

// fmtFuncCall maps English stdlib function names to Python equivalents,
// mirroring the AST transpiler's stdlib.go mapping.
func (d *decompiler) fmtFuncCall(name string, args []string) string {
	a := func(i int) string {
		if i < len(args) {
			return args[i]
		}
		return "None"
	}
	joined := strings.Join(args, ", ")

	if d.userFuncs[name] {
		return fmt.Sprintf("%s(%s)", sanitizeDecompIdent(name), joined)
	}

	switch name {
	// Math
	case "sqrt":
		d.needsMath = true
		return fmt.Sprintf("math.sqrt(%s)", a(0))
	case "pow":
		d.needsMath = true
		return fmt.Sprintf("math.pow(%s, %s)", a(0), a(1))
	case "abs":
		return fmt.Sprintf("abs(%s)", a(0))
	case "floor":
		d.needsMath = true
		return fmt.Sprintf("math.floor(%s)", a(0))
	case "ceil":
		d.needsMath = true
		return fmt.Sprintf("math.ceil(%s)", a(0))
	case "round":
		return fmt.Sprintf("round(%s)", a(0))
	case "min":
		return fmt.Sprintf("min(%s, %s)", a(0), a(1))
	case "max":
		return fmt.Sprintf("max(%s, %s)", a(0), a(1))
	case "sin":
		d.needsMath = true
		return fmt.Sprintf("math.sin(%s)", a(0))
	case "cos":
		d.needsMath = true
		return fmt.Sprintf("math.cos(%s)", a(0))
	case "tan":
		d.needsMath = true
		return fmt.Sprintf("math.tan(%s)", a(0))
	case "log":
		d.needsMath = true
		return fmt.Sprintf("math.log(%s)", a(0))
	case "log10":
		d.needsMath = true
		return fmt.Sprintf("math.log10(%s)", a(0))
	case "log2":
		d.needsMath = true
		return fmt.Sprintf("math.log2(%s)", a(0))
	case "exp":
		d.needsMath = true
		return fmt.Sprintf("math.exp(%s)", a(0))
	case "random":
		d.needsRandom = true
		return "random.random()"
	case "random_between":
		d.needsRandom = true
		return fmt.Sprintf("random.uniform(%s, %s)", a(0), a(1))
	case "is_nan":
		d.helpers["_is_nan"] = true
		d.needsMath = true
		return fmt.Sprintf("_is_nan(%s)", a(0))
	case "is_infinite":
		d.helpers["_is_infinite"] = true
		d.needsMath = true
		return fmt.Sprintf("_is_infinite(%s)", a(0))
	case "is_integer":
		return fmt.Sprintf("float(%s).is_integer()", a(0))
	case "clamp":
		return fmt.Sprintf("max(%s, min(%s, %s))", a(1), a(2), a(0))
	case "sign":
		d.helpers["_sign"] = true
		return fmt.Sprintf("_sign(%s)", a(0))
	// String
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
		start := a(1)
		length := a(2)
		return fmt.Sprintf("%s[int(%s):int(%s)+int(%s)]", a(0), start, start, length)
	case "str_repeat":
		return fmt.Sprintf("%s * int(%s)", a(0), a(1))
	case "count_occurrences":
		return fmt.Sprintf("%s.count(%s)", a(0), a(1))
	case "pad_left":
		if len(args) > 2 {
			return fmt.Sprintf("%s.rjust(int(%s), %s)", a(0), a(1), a(2))
		}
		return fmt.Sprintf("%s.rjust(int(%s))", a(0), a(1))
	case "pad_right":
		if len(args) > 2 {
			return fmt.Sprintf("%s.ljust(int(%s), %s)", a(0), a(1), a(2))
		}
		return fmt.Sprintf("%s.ljust(int(%s))", a(0), a(1))
	case "center":
		if len(args) > 2 {
			return fmt.Sprintf("%s.center(int(%s), %s)", a(0), a(1), a(2))
		}
		return fmt.Sprintf("%s.center(int(%s))", a(0), a(1))
	case "zfill":
		return fmt.Sprintf("%s.zfill(int(%s))", a(0), a(1))
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
	// List
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
		return fmt.Sprintf("%s + [%s]", a(0), a(1))
	case "pop":
		return fmt.Sprintf("%s[:-1]", a(0))
	case "remove":
		return fmt.Sprintf("[v for i, v in enumerate(%s) if i != int(%s)]", a(0), a(1))
	case "insert":
		return fmt.Sprintf("(%s[:int(%s)] + [%s] + %s[int(%s):])", a(0), a(1), a(2), a(0), a(1))
	case "sum":
		return fmt.Sprintf("sum(%s)", a(0))
	case "product":
		d.helpers["_product"] = true
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
		d.helpers["_unique"] = true
		return fmt.Sprintf("_unique(%s)", a(0))
	case "flatten":
		d.helpers["_flatten"] = true
		return fmt.Sprintf("_flatten(%s)", a(0))
	case "slice":
		return fmt.Sprintf("%s[int(%s):int(%s)]", a(0), a(1), a(2))
	case "zip_with":
		d.helpers["_zip_with"] = true
		return fmt.Sprintf("_zip_with(%s, %s)", a(0), a(1))
	// Lookup table
	case "keys":
		return fmt.Sprintf("list(%s.keys())", a(0))
	case "values":
		return fmt.Sprintf("list(%s.values())", a(0))
	case "table_remove":
		d.helpers["_table_remove"] = true
		return fmt.Sprintf("_table_remove(%s, %s)", a(0), a(1))
	case "table_has":
		return fmt.Sprintf("(%s in %s)", a(1), a(0))
	case "merge":
		return fmt.Sprintf("{**%s, **%s}", a(0), a(1))
	case "get_or_default":
		return fmt.Sprintf("%s.get(%s, %s)", a(0), a(1), a(2))
	// I/O
	case "ask":
		return fmt.Sprintf("input(%s)", a(0))
	case "read_file":
		d.helpers["_read_file"] = true
		return fmt.Sprintf("_read_file(%s)", a(0))
	case "write_file":
		d.helpers["_write_file"] = true
		return fmt.Sprintf("_write_file(%s, %s)", a(0), a(1))
	}
	return fmt.Sprintf("%s(%s)", sanitizeDecompIdent(name), joined)
}
