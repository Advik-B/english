package vm

import (
	"bufio"
	"english/vm/types"
	"fmt"
	"math"
	"math/rand"
	"os"
	"sort"
	"strings"
)

// RegisterStdlib registers all standard library functions
func RegisterStdlib(env *Environment) {
	registerMathFunctions(env)
	registerStringFunctions(env)
	registerListFunctions(env)
	registerIOFunctions(env)
	registerLookupTableFunctions(env)
	registerNumberFunctions(env)
	registerMathConstants(env)
}

// registerMathConstants registers mathematical constants as read-only variables
func registerMathConstants(env *Environment) {
	env.Define("pi", math.Pi, true)
	env.Define("e", math.E, true)
	env.Define("infinity", math.Inf(1), true)
}

// registerMathFunctions registers all math functions
func registerMathFunctions(env *Environment) {
	env.DefineFunction("sqrt", &FunctionValue{Name: "sqrt", Parameters: []string{"x"}, Body: nil, Closure: env})
	env.DefineFunction("pow", &FunctionValue{Name: "pow", Parameters: []string{"base", "exponent"}, Body: nil, Closure: env})
	env.DefineFunction("abs", &FunctionValue{Name: "abs", Parameters: []string{"x"}, Body: nil, Closure: env})
	env.DefineFunction("floor", &FunctionValue{Name: "floor", Parameters: []string{"x"}, Body: nil, Closure: env})
	env.DefineFunction("ceil", &FunctionValue{Name: "ceil", Parameters: []string{"x"}, Body: nil, Closure: env})
	env.DefineFunction("round", &FunctionValue{Name: "round", Parameters: []string{"x"}, Body: nil, Closure: env})
	env.DefineFunction("min", &FunctionValue{Name: "min", Parameters: []string{"a", "b"}, Body: nil, Closure: env})
	env.DefineFunction("max", &FunctionValue{Name: "max", Parameters: []string{"a", "b"}, Body: nil, Closure: env})
	env.DefineFunction("sin", &FunctionValue{Name: "sin", Parameters: []string{"x"}, Body: nil, Closure: env})
	env.DefineFunction("cos", &FunctionValue{Name: "cos", Parameters: []string{"x"}, Body: nil, Closure: env})
	env.DefineFunction("tan", &FunctionValue{Name: "tan", Parameters: []string{"x"}, Body: nil, Closure: env})
	env.DefineFunction("log", &FunctionValue{Name: "log", Parameters: []string{"x"}, Body: nil, Closure: env})
	env.DefineFunction("log10", &FunctionValue{Name: "log10", Parameters: []string{"x"}, Body: nil, Closure: env})
	env.DefineFunction("log2", &FunctionValue{Name: "log2", Parameters: []string{"x"}, Body: nil, Closure: env})
	env.DefineFunction("exp", &FunctionValue{Name: "exp", Parameters: []string{"x"}, Body: nil, Closure: env})
	env.DefineFunction("random", &FunctionValue{Name: "random", Parameters: []string{}, Body: nil, Closure: env})
	env.DefineFunction("random_between", &FunctionValue{Name: "random_between", Parameters: []string{"min", "max"}, Body: nil, Closure: env})
	env.DefineFunction("is_nan", &FunctionValue{Name: "is_nan", Parameters: []string{"x"}, Body: nil, Closure: env})
	env.DefineFunction("is_infinite", &FunctionValue{Name: "is_infinite", Parameters: []string{"x"}, Body: nil, Closure: env})
}

// registerStringFunctions registers all string functions
func registerStringFunctions(env *Environment) {
	env.DefineFunction("uppercase", &FunctionValue{Name: "uppercase", Parameters: []string{"text"}, Body: nil, Closure: env})
	env.DefineFunction("lowercase", &FunctionValue{Name: "lowercase", Parameters: []string{"text"}, Body: nil, Closure: env})
	env.DefineFunction("casefold", &FunctionValue{Name: "casefold", Parameters: []string{"text"}, Body: nil, Closure: env})
	env.DefineFunction("split", &FunctionValue{Name: "split", Parameters: []string{"text", "separator"}, Body: nil, Closure: env})
	env.DefineFunction("join", &FunctionValue{Name: "join", Parameters: []string{"list", "separator"}, Body: nil, Closure: env})
	env.DefineFunction("trim", &FunctionValue{Name: "trim", Parameters: []string{"text"}, Body: nil, Closure: env})
	env.DefineFunction("replace", &FunctionValue{Name: "replace", Parameters: []string{"text", "old", "new"}, Body: nil, Closure: env})
	env.DefineFunction("contains", &FunctionValue{Name: "contains", Parameters: []string{"text", "substring"}, Body: nil, Closure: env})
	env.DefineFunction("starts_with", &FunctionValue{Name: "starts_with", Parameters: []string{"text", "prefix"}, Body: nil, Closure: env})
	env.DefineFunction("ends_with", &FunctionValue{Name: "ends_with", Parameters: []string{"text", "suffix"}, Body: nil, Closure: env})
	env.DefineFunction("index_of", &FunctionValue{Name: "index_of", Parameters: []string{"text", "search"}, Body: nil, Closure: env})
	env.DefineFunction("substring", &FunctionValue{Name: "substring", Parameters: []string{"text", "start", "length"}, Body: nil, Closure: env})
	env.DefineFunction("str_repeat", &FunctionValue{Name: "str_repeat", Parameters: []string{"text", "n"}, Body: nil, Closure: env})
	env.DefineFunction("count_occurrences", &FunctionValue{Name: "count_occurrences", Parameters: []string{"text", "substring"}, Body: nil, Closure: env})
	env.DefineFunction("pad_left", &FunctionValue{Name: "pad_left", Parameters: []string{"text", "width", "char"}, Body: nil, Closure: env})
	env.DefineFunction("pad_right", &FunctionValue{Name: "pad_right", Parameters: []string{"text", "width", "char"}, Body: nil, Closure: env})
	env.DefineFunction("to_number", &FunctionValue{Name: "to_number", Parameters: []string{"text"}, Body: nil, Closure: env})
	env.DefineFunction("to_string", &FunctionValue{Name: "to_string", Parameters: []string{"value"}, Body: nil, Closure: env})
	env.DefineFunction("is_empty", &FunctionValue{Name: "is_empty", Parameters: []string{"value"}, Body: nil, Closure: env})
	// Python-equivalent string methods
	env.DefineFunction("title", &FunctionValue{Name: "title", Parameters: []string{"text"}, Body: nil, Closure: env})
	env.DefineFunction("capitalize", &FunctionValue{Name: "capitalize", Parameters: []string{"text"}, Body: nil, Closure: env})
	env.DefineFunction("swapcase", &FunctionValue{Name: "swapcase", Parameters: []string{"text"}, Body: nil, Closure: env})
	env.DefineFunction("trim_left", &FunctionValue{Name: "trim_left", Parameters: []string{"text"}, Body: nil, Closure: env})
	env.DefineFunction("trim_right", &FunctionValue{Name: "trim_right", Parameters: []string{"text"}, Body: nil, Closure: env})
	env.DefineFunction("is_digit", &FunctionValue{Name: "is_digit", Parameters: []string{"text"}, Body: nil, Closure: env})
	env.DefineFunction("is_alpha", &FunctionValue{Name: "is_alpha", Parameters: []string{"text"}, Body: nil, Closure: env})
	env.DefineFunction("is_alnum", &FunctionValue{Name: "is_alnum", Parameters: []string{"text"}, Body: nil, Closure: env})
	env.DefineFunction("is_space", &FunctionValue{Name: "is_space", Parameters: []string{"text"}, Body: nil, Closure: env})
	env.DefineFunction("is_upper", &FunctionValue{Name: "is_upper", Parameters: []string{"text"}, Body: nil, Closure: env})
	env.DefineFunction("is_lower", &FunctionValue{Name: "is_lower", Parameters: []string{"text"}, Body: nil, Closure: env})
	env.DefineFunction("center", &FunctionValue{Name: "center", Parameters: []string{"text", "width", "char"}, Body: nil, Closure: env})
	env.DefineFunction("zfill", &FunctionValue{Name: "zfill", Parameters: []string{"text", "width"}, Body: nil, Closure: env})
}

// registerListFunctions registers all list functions
func registerListFunctions(env *Environment) {
	env.DefineFunction("append", &FunctionValue{Name: "append", Parameters: []string{"list", "item"}, Body: nil, Closure: env})
	env.DefineFunction("remove", &FunctionValue{Name: "remove", Parameters: []string{"list", "index"}, Body: nil, Closure: env})
	env.DefineFunction("insert", &FunctionValue{Name: "insert", Parameters: []string{"list", "index", "item"}, Body: nil, Closure: env})
	env.DefineFunction("sort", &FunctionValue{Name: "sort", Parameters: []string{"list"}, Body: nil, Closure: env})
	env.DefineFunction("reverse", &FunctionValue{Name: "reverse", Parameters: []string{"list"}, Body: nil, Closure: env})
	env.DefineFunction("sum", &FunctionValue{Name: "sum", Parameters: []string{"list"}, Body: nil, Closure: env})
	env.DefineFunction("unique", &FunctionValue{Name: "unique", Parameters: []string{"list"}, Body: nil, Closure: env})
	env.DefineFunction("first", &FunctionValue{Name: "first", Parameters: []string{"list"}, Body: nil, Closure: env})
	env.DefineFunction("last", &FunctionValue{Name: "last", Parameters: []string{"list"}, Body: nil, Closure: env})
	env.DefineFunction("flatten", &FunctionValue{Name: "flatten", Parameters: []string{"list"}, Body: nil, Closure: env})
	env.DefineFunction("count", &FunctionValue{Name: "count", Parameters: []string{"list"}, Body: nil, Closure: env})
	env.DefineFunction("slice", &FunctionValue{Name: "slice", Parameters: []string{"list", "start", "end"}, Body: nil, Closure: env})
	// Python-equivalent list methods
	env.DefineFunction("average", &FunctionValue{Name: "average", Parameters: []string{"list"}, Body: nil, Closure: env})
	env.DefineFunction("min_value", &FunctionValue{Name: "min_value", Parameters: []string{"list"}, Body: nil, Closure: env})
	env.DefineFunction("max_value", &FunctionValue{Name: "max_value", Parameters: []string{"list"}, Body: nil, Closure: env})
	env.DefineFunction("any_true", &FunctionValue{Name: "any_true", Parameters: []string{"list"}, Body: nil, Closure: env})
	env.DefineFunction("all_true", &FunctionValue{Name: "all_true", Parameters: []string{"list"}, Body: nil, Closure: env})
	env.DefineFunction("product", &FunctionValue{Name: "product", Parameters: []string{"list"}, Body: nil, Closure: env})
	env.DefineFunction("sorted_desc", &FunctionValue{Name: "sorted_desc", Parameters: []string{"list"}, Body: nil, Closure: env})
	env.DefineFunction("zip_with", &FunctionValue{Name: "zip_with", Parameters: []string{"list", "other"}, Body: nil, Closure: env})
}

// registerIOFunctions registers input/output functions
func registerIOFunctions(env *Environment) {
	env.DefineFunction("ask", &FunctionValue{Name: "ask", Parameters: []string{"prompt"}, Body: nil, Closure: env})
}

// registerLookupTableFunctions registers lookup table stdlib functions
func registerLookupTableFunctions(env *Environment) {
	env.DefineFunction("keys", &FunctionValue{Name: "keys", Parameters: []string{"table"}, Body: nil, Closure: env})
	env.DefineFunction("values", &FunctionValue{Name: "values", Parameters: []string{"table"}, Body: nil, Closure: env})
	env.DefineFunction("table_remove", &FunctionValue{Name: "table_remove", Parameters: []string{"table", "key"}, Body: nil, Closure: env})
	env.DefineFunction("table_has", &FunctionValue{Name: "table_has", Parameters: []string{"table", "key"}, Body: nil, Closure: env})
	// Python-equivalent lookup table methods
	env.DefineFunction("merge", &FunctionValue{Name: "merge", Parameters: []string{"table", "other"}, Body: nil, Closure: env})
	env.DefineFunction("get_or_default", &FunctionValue{Name: "get_or_default", Parameters: []string{"table", "key", "default"}, Body: nil, Closure: env})
}

// registerNumberFunctions registers number-specific functions
func registerNumberFunctions(env *Environment) {
	env.DefineFunction("is_integer", &FunctionValue{Name: "is_integer", Parameters: []string{"x"}, Body: nil, Closure: env})
	env.DefineFunction("clamp", &FunctionValue{Name: "clamp", Parameters: []string{"x", "min", "max"}, Body: nil, Closure: env})
	env.DefineFunction("sign", &FunctionValue{Name: "sign", Parameters: []string{"x"}, Body: nil, Closure: env})
}

// requireText checks that arg is a string value and returns a descriptive
// TypeError if it is not.  Used by text-specific stdlib functions so that
// calling e.g. `42's title` produces a clear error instead of silently
// converting the number.
func requireText(fn string, arg Value) (string, error) {
	s, ok := arg.(string)
	if !ok {
		return "", fmt.Errorf("TypeError: %s expects text, got %s", fn, typeKindName(inferTypeKind(arg)))
	}
	return s, nil
}

// requireNumber checks that arg is a numeric value and returns a descriptive
// TypeError if it is not.
func requireNumber(fn string, arg Value) (float64, error) {
	n, err := ToNumber(arg)
	if err != nil {
		return 0, fmt.Errorf("TypeError: %s expects number, got %s", fn, typeKindName(inferTypeKind(arg)))
	}
	return n, nil
}

// requireList checks that arg is a list ([]interface{}) and returns a
// descriptive TypeError if it is not.
func requireList(fn string, arg Value) ([]interface{}, error) {
	lst, ok := arg.([]interface{})
	if !ok {
		return nil, fmt.Errorf("TypeError: %s expects list, got %s", fn, typeKindName(inferTypeKind(arg)))
	}
	return lst, nil
}

// requireLookupTable checks that arg is a *LookupTableValue and returns a
// descriptive TypeError if it is not.
func requireLookupTable(fn string, arg Value) (*LookupTableValue, error) {
	lt, ok := arg.(*LookupTableValue)
	if !ok {
		return nil, fmt.Errorf("TypeError: %s expects lookup table, got %s", fn, typeKindName(inferTypeKind(arg)))
	}
	return lt, nil
}

// evalBuiltinFunction evaluates a built-in stdlib function
func evalBuiltinFunction(name string, args []Value) (Value, error) {
	switch name {
	// ── Math ──────────────────────────────────────────────────────────────────
	case "sqrt":
		x, err := ToNumber(args[0])
		if err != nil {
			return nil, err
		}
		return math.Sqrt(x), nil
	case "pow":
		base, err := ToNumber(args[0])
		if err != nil {
			return nil, err
		}
		exp, err := ToNumber(args[1])
		if err != nil {
			return nil, err
		}
		return math.Pow(base, exp), nil
	case "abs":
		x, err := ToNumber(args[0])
		if err != nil {
			return nil, err
		}
		return math.Abs(x), nil
	case "floor":
		x, err := ToNumber(args[0])
		if err != nil {
			return nil, err
		}
		return math.Floor(x), nil
	case "ceil":
		x, err := ToNumber(args[0])
		if err != nil {
			return nil, err
		}
		return math.Ceil(x), nil
	case "round":
		x, err := ToNumber(args[0])
		if err != nil {
			return nil, err
		}
		return math.Round(x), nil
	case "min":
		a, err := ToNumber(args[0])
		if err != nil {
			return nil, err
		}
		b, err := ToNumber(args[1])
		if err != nil {
			return nil, err
		}
		return math.Min(a, b), nil
	case "max":
		a, err := ToNumber(args[0])
		if err != nil {
			return nil, err
		}
		b, err := ToNumber(args[1])
		if err != nil {
			return nil, err
		}
		return math.Max(a, b), nil
	case "sin":
		x, err := ToNumber(args[0])
		if err != nil {
			return nil, err
		}
		return math.Sin(x), nil
	case "cos":
		x, err := ToNumber(args[0])
		if err != nil {
			return nil, err
		}
		return math.Cos(x), nil
	case "tan":
		x, err := ToNumber(args[0])
		if err != nil {
			return nil, err
		}
		return math.Tan(x), nil
	case "log":
		x, err := ToNumber(args[0])
		if err != nil {
			return nil, err
		}
		return math.Log(x), nil
	case "log10":
		x, err := ToNumber(args[0])
		if err != nil {
			return nil, err
		}
		return math.Log10(x), nil
	case "log2":
		x, err := ToNumber(args[0])
		if err != nil {
			return nil, err
		}
		return math.Log2(x), nil
	case "exp":
		x, err := ToNumber(args[0])
		if err != nil {
			return nil, err
		}
		return math.Exp(x), nil
	case "random":
		return rand.Float64(), nil
	case "random_between":
		a, err := ToNumber(args[0])
		if err != nil {
			return nil, NewRuntimeError("random_between expects a number as first argument")
		}
		b, err := ToNumber(args[1])
		if err != nil {
			return nil, NewRuntimeError("random_between expects a number as second argument")
		}
		if a > b {
			return nil, NewRuntimeError("random_between: min must be less than or equal to max")
		}
		return a + rand.Float64()*(b-a), nil
	case "is_nan":
		x, err := ToNumber(args[0])
		if err != nil {
			return true, nil // non-numeric is NaN-like
		}
		return math.IsNaN(x), nil
	case "is_infinite":
		x, err := ToNumber(args[0])
		if err != nil {
			return false, nil
		}
		return math.IsInf(x, 0), nil

	// ── String ────────────────────────────────────────────────────────────────
	case "uppercase":
		text, err := requireText("uppercase", args[0])
		if err != nil {
			return nil, err
		}
		return strings.ToUpper(text), nil
	case "lowercase":
		text, err := requireText("lowercase", args[0])
		if err != nil {
			return nil, err
		}
		return strings.ToLower(text), nil
	case "casefold":
		// casefold converts text to a case-folded (lowercase) form suitable for
		// case-insensitive comparisons.  Equivalent to Python's str.casefold().
		text, err := requireText("casefold", args[0])
		if err != nil {
			return nil, err
		}
		return strings.ToLower(text), nil
	case "split":
		text, err := requireText("split", args[0])
		if err != nil {
			return nil, err
		}
		sep := ToString(args[1])
		parts := strings.Split(text, sep)
		result := make([]interface{}, len(parts))
		for i, part := range parts {
			result[i] = part
		}
		return result, nil
	case "join":
		list, ok := args[0].([]interface{})
		if !ok {
			return nil, fmt.Errorf("TypeError: join expects list, got %s", typeKindName(inferTypeKind(args[0])))
		}
		sep := ToString(args[1])
		strs := make([]string, len(list))
		for i, item := range list {
			strs[i] = ToString(item)
		}
		return strings.Join(strs, sep), nil
	case "trim":
		text, err := requireText("trim", args[0])
		if err != nil {
			return nil, err
		}
		return strings.TrimSpace(text), nil
	case "replace":
		text, err := requireText("replace", args[0])
		if err != nil {
			return nil, err
		}
		old := ToString(args[1])
		newStr := ToString(args[2])
		return strings.ReplaceAll(text, old, newStr), nil
	case "contains":
		text, err := requireText("contains", args[0])
		if err != nil {
			return nil, err
		}
		substr := ToString(args[1])
		return strings.Contains(text, substr), nil
	case "starts_with":
		text, err := requireText("starts_with", args[0])
		if err != nil {
			return nil, err
		}
		prefix := ToString(args[1])
		return strings.HasPrefix(text, prefix), nil
	case "ends_with":
		text, err := requireText("ends_with", args[0])
		if err != nil {
			return nil, err
		}
		suffix := ToString(args[1])
		return strings.HasSuffix(text, suffix), nil
	case "index_of":
		text, err := requireText("index_of", args[0])
		if err != nil {
			return nil, err
		}
		search := ToString(args[1])
		idx := strings.Index(text, search)
		return float64(idx), nil
	case "substring":
		text, err := requireText("substring", args[0])
		if err != nil {
			return nil, err
		}
		start, err := ToNumber(args[1])
		if err != nil {
			return nil, NewRuntimeError("substring expects a number as second argument")
		}
		length, err := ToNumber(args[2])
		if err != nil {
			return nil, NewRuntimeError("substring expects a number as third argument")
		}
		s := int(start)
		l := int(length)
		if s < 0 || s > len(text) {
			return nil, NewRuntimeError(fmt.Sprintf("substring start index %d out of range", s))
		}
		end := s + l
		if end > len(text) {
			end = len(text)
		}
		return text[s:end], nil
	case "str_repeat":
		text, err := requireText("str_repeat", args[0])
		if err != nil {
			return nil, err
		}
		n, err := ToNumber(args[1])
		if err != nil {
			return nil, NewRuntimeError("str_repeat expects a number as second argument")
		}
		if int(n) < 0 {
			return nil, NewRuntimeError("str_repeat count must be non-negative")
		}
		return strings.Repeat(text, int(n)), nil
	case "count_occurrences":
		text, err := requireText("count_occurrences", args[0])
		if err != nil {
			return nil, err
		}
		sub := ToString(args[1])
		return float64(strings.Count(text, sub)), nil
	case "pad_left":
		text, err := requireText("pad_left", args[0])
		if err != nil {
			return nil, err
		}
		width, err := ToNumber(args[1])
		if err != nil {
			return nil, NewRuntimeError("pad_left expects a number as second argument")
		}
		padChar := " "
		if len(args) > 2 {
			padChar = ToString(args[2])
			if len(padChar) == 0 {
				padChar = " "
			}
		}
		w := int(width)
		for len(text) < w {
			text = padChar[:1] + text
		}
		return text, nil
	case "pad_right":
		text, err := requireText("pad_right", args[0])
		if err != nil {
			return nil, err
		}
		width, err := ToNumber(args[1])
		if err != nil {
			return nil, NewRuntimeError("pad_right expects a number as second argument")
		}
		padChar := " "
		if len(args) > 2 {
			padChar = ToString(args[2])
			if len(padChar) == 0 {
				padChar = " "
			}
		}
		w := int(width)
		for len(text) < w {
			text = text + padChar[:1]
		}
		return text, nil
	case "to_number":
		text, err := requireText("to_number", args[0])
		if err != nil {
			return nil, err
		}
		var f float64
		_, err = fmt.Sscanf(text, "%g", &f)
		if err != nil {
			return nil, NewRuntimeError(fmt.Sprintf("cannot convert '%s' to a number", text))
		}
		return f, nil
	case "to_string":
		return ToString(args[0]), nil
	case "is_empty":
		switch v := args[0].(type) {
		case string:
			return len(v) == 0, nil
		case []interface{}:
			return len(v) == 0, nil
		case nil:
			return true, nil
		default:
			return false, nil
		}
	// ── String (Python-equivalent methods) ────────────────────────────────────
	case "title":
		// str.title() — uppercase the first letter of each word.
		text, err := requireText("title", args[0])
		if err != nil {
			return nil, err
		}
		return strings.Title(strings.ToLower(text)), nil //nolint:staticcheck
	case "capitalize":
		// str.capitalize() — uppercase first character, lowercase rest.
		text, err := requireText("capitalize", args[0])
		if err != nil {
			return nil, err
		}
		if text == "" {
			return text, nil
		}
		return strings.ToUpper(text[:1]) + strings.ToLower(text[1:]), nil
	case "swapcase":
		// str.swapcase() — swap the case of every character.
		text, err := requireText("swapcase", args[0])
		if err != nil {
			return nil, err
		}
		var sb strings.Builder
		for _, r := range text {
			if r >= 'A' && r <= 'Z' {
				sb.WriteRune(r + 32)
			} else if r >= 'a' && r <= 'z' {
				sb.WriteRune(r - 32)
			} else {
				sb.WriteRune(r)
			}
		}
		return sb.String(), nil
	case "trim_left":
		// str.lstrip() — strip leading whitespace.
		text, err := requireText("trim_left", args[0])
		if err != nil {
			return nil, err
		}
		return strings.TrimLeftFunc(text, func(r rune) bool { return r == ' ' || r == '\t' || r == '\n' || r == '\r' }), nil
	case "trim_right":
		// str.rstrip() — strip trailing whitespace.
		text, err := requireText("trim_right", args[0])
		if err != nil {
			return nil, err
		}
		return strings.TrimRightFunc(text, func(r rune) bool { return r == ' ' || r == '\t' || r == '\n' || r == '\r' }), nil
	case "is_digit":
		// str.isdigit() — true if all characters are digits and text is non-empty.
		text, err := requireText("is_digit", args[0])
		if err != nil {
			return nil, err
		}
		if text == "" {
			return false, nil
		}
		for _, r := range text {
			if r < '0' || r > '9' {
				return false, nil
			}
		}
		return true, nil
	case "is_alpha":
		// str.isalpha() — true if all characters are letters and text is non-empty.
		text, err := requireText("is_alpha", args[0])
		if err != nil {
			return nil, err
		}
		if text == "" {
			return false, nil
		}
		for _, r := range text {
			if !((r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z')) {
				return false, nil
			}
		}
		return true, nil
	case "is_alnum":
		// str.isalnum() — true if all characters are letters or digits and text is non-empty.
		text, err := requireText("is_alnum", args[0])
		if err != nil {
			return nil, err
		}
		if text == "" {
			return false, nil
		}
		for _, r := range text {
			if !((r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9')) {
				return false, nil
			}
		}
		return true, nil
	case "is_space":
		// str.isspace() — true if all characters are whitespace and text is non-empty.
		text, err := requireText("is_space", args[0])
		if err != nil {
			return nil, err
		}
		if text == "" {
			return false, nil
		}
		return strings.TrimSpace(text) == "", nil
	case "is_upper":
		// str.isupper() — true if all cased characters are uppercase.
		text, err := requireText("is_upper", args[0])
		if err != nil {
			return nil, err
		}
		if text == "" {
			return false, nil
		}
		hasCased := false
		for _, r := range text {
			if r >= 'a' && r <= 'z' {
				return false, nil
			}
			if r >= 'A' && r <= 'Z' {
				hasCased = true
			}
		}
		return hasCased, nil
	case "is_lower":
		// str.islower() — true if all cased characters are lowercase.
		text, err := requireText("is_lower", args[0])
		if err != nil {
			return nil, err
		}
		if text == "" {
			return false, nil
		}
		hasCased := false
		for _, r := range text {
			if r >= 'A' && r <= 'Z' {
				return false, nil
			}
			if r >= 'a' && r <= 'z' {
				hasCased = true
			}
		}
		return hasCased, nil
	case "center":
		// str.center(width[, fillchar]) — center text within a field of given width.
		text, err := requireText("center", args[0])
		if err != nil {
			return nil, err
		}
		width, err := requireNumber("center", args[1])
		if err != nil {
			return nil, err
		}
		fillChar := " "
		if len(args) > 2 {
			fillChar, err = requireText("center", args[2])
			if err != nil {
				return nil, err
			}
			if len(fillChar) == 0 {
				fillChar = " "
			}
		}
		w := int(width)
		pad := w - len(text)
		if pad <= 0 {
			return text, nil
		}
		left := pad / 2
		right := pad - left
		return strings.Repeat(fillChar[:1], left) + text + strings.Repeat(fillChar[:1], right), nil
	case "zfill":
		// str.zfill(width) — pad text with leading zeros to reach given width.
		text, err := requireText("zfill", args[0])
		if err != nil {
			return nil, err
		}
		width, err := requireNumber("zfill", args[1])
		if err != nil {
			return nil, err
		}
		w := int(width)
		if len(text) >= w {
			return text, nil
		}
		prefix := ""
		body := text
		if len(body) > 0 && (body[0] == '+' || body[0] == '-') {
			prefix = string(body[0])
			body = body[1:]
		}
		return prefix + strings.Repeat("0", w-len(prefix)-len(body)) + body, nil

	// ── Number (Python-equivalent methods) ────────────────────────────────────
	case "is_integer":
		// float.is_integer() — true if the number has no fractional part.
		x, err := requireNumber("is_integer", args[0])
		if err != nil {
			return nil, err
		}
		return x == math.Trunc(x), nil
	case "clamp":
		// clamp(x, min, max) — limit x to the range [min, max].
		x, err := requireNumber("clamp", args[0])
		if err != nil {
			return nil, err
		}
		lo, err := requireNumber("clamp", args[1])
		if err != nil {
			return nil, err
		}
		hi, err := requireNumber("clamp", args[2])
		if err != nil {
			return nil, err
		}
		if x < lo {
			return lo, nil
		}
		if x > hi {
			return hi, nil
		}
		return x, nil
	case "sign":
		// sign(x) — returns -1, 0, or 1 depending on the sign of x.
		x, err := requireNumber("sign", args[0])
		if err != nil {
			return nil, err
		}
		if x < 0 {
			return float64(-1), nil
		}
		if x > 0 {
			return float64(1), nil
		}
		return float64(0), nil

	// ── List (Python-equivalent methods) ──────────────────────────────────────
	case "average":
		// statistics.mean() / sum(l)/len(l) — average of a numeric list.
		lst, err := requireList("average", args[0])
		if err != nil {
			return nil, err
		}
		if len(lst) == 0 {
			return nil, fmt.Errorf("RuntimeError: average called on empty list")
		}
		total := 0.0
		for _, item := range lst {
			n, err := ToNumber(item)
			if err != nil {
				return nil, fmt.Errorf("TypeError: average requires a list of numbers")
			}
			total += n
		}
		return total / float64(len(lst)), nil
	case "min_value":
		// min(list) — minimum value in a numeric list.
		lst, err := requireList("min_value", args[0])
		if err != nil {
			return nil, err
		}
		if len(lst) == 0 {
			return nil, fmt.Errorf("RuntimeError: min_value called on empty list")
		}
		best, err := ToNumber(lst[0])
		if err != nil {
			return nil, fmt.Errorf("TypeError: min_value requires a list of numbers")
		}
		for _, item := range lst[1:] {
			n, err := ToNumber(item)
			if err != nil {
				return nil, fmt.Errorf("TypeError: min_value requires a list of numbers")
			}
			if n < best {
				best = n
			}
		}
		return best, nil
	case "max_value":
		// max(list) — maximum value in a numeric list.
		lst, err := requireList("max_value", args[0])
		if err != nil {
			return nil, err
		}
		if len(lst) == 0 {
			return nil, fmt.Errorf("RuntimeError: max_value called on empty list")
		}
		best, err := ToNumber(lst[0])
		if err != nil {
			return nil, fmt.Errorf("TypeError: max_value requires a list of numbers")
		}
		for _, item := range lst[1:] {
			n, err := ToNumber(item)
			if err != nil {
				return nil, fmt.Errorf("TypeError: max_value requires a list of numbers")
			}
			if n > best {
				best = n
			}
		}
		return best, nil
	case "any_true":
		// any() — true if at least one element is truthy.
		lst, err := requireList("any_true", args[0])
		if err != nil {
			return nil, err
		}
		for _, item := range lst {
			b, err := ToBool(item)
			if err != nil {
				return nil, fmt.Errorf("TypeError: any_true requires a list of boolean values")
			}
			if b {
				return true, nil
			}
		}
		return false, nil
	case "all_true":
		// all() — true if every element is truthy.
		lst, err := requireList("all_true", args[0])
		if err != nil {
			return nil, err
		}
		for _, item := range lst {
			b, err := ToBool(item)
			if err != nil {
				return nil, fmt.Errorf("TypeError: all_true requires a list of boolean values")
			}
			if !b {
				return false, nil
			}
		}
		return true, nil
	case "product":
		// math.prod() — product of all numbers in a list.
		lst, err := requireList("product", args[0])
		if err != nil {
			return nil, err
		}
		result := 1.0
		for _, item := range lst {
			n, err := ToNumber(item)
			if err != nil {
				return nil, fmt.Errorf("TypeError: product requires a list of numbers")
			}
			result *= n
		}
		return result, nil
	case "sorted_desc":
		// sorted(list, reverse=True) — return a new list sorted in descending order.
		lst, err := requireList("sorted_desc", args[0])
		if err != nil {
			return nil, err
		}
		result := make([]interface{}, len(lst))
		copy(result, lst)
		sort.Slice(result, func(i, j int) bool {
			a, errA := ToNumber(result[i])
			b, errB := ToNumber(result[j])
			if errA == nil && errB == nil {
				return a > b
			}
			return ToString(result[i]) > ToString(result[j])
		})
		return result, nil
	case "zip_with":
		// zip(a, b) — pair elements of two lists into a list of two-element lists.
		lst, err := requireList("zip_with", args[0])
		if err != nil {
			return nil, err
		}
		other, err := requireList("zip_with", args[1])
		if err != nil {
			return nil, err
		}
		length := len(lst)
		if len(other) < length {
			length = len(other)
		}
		result := make([]interface{}, length)
		for i := 0; i < length; i++ {
			result[i] = []interface{}{lst[i], other[i]}
		}
		return result, nil

	// ── Lookup table (Python-equivalent methods) ───────────────────────────────
	case "merge":
		// {**a, **b} — merge two lookup tables; second overwrites first on conflict.
		lt, err := requireLookupTable("merge", args[0])
		if err != nil {
			return nil, err
		}
		other, err := requireLookupTable("merge", args[1])
		if err != nil {
			return nil, err
		}
		newTable := types.NewLookupTable()
		for _, k := range lt.KeyOrder {
			newTable.Set(k, lt.Entries[k])
		}
		for _, k := range other.KeyOrder {
			newTable.Set(k, other.Entries[k])
		}
		return newTable, nil
	case "get_or_default":
		// dict.get(key, default) — return value for key, or default if absent.
		lt, err := requireLookupTable("get_or_default", args[0])
		if err != nil {
			return nil, err
		}
		serialKey, err := types.SerializeKey(args[1])
		if err != nil {
			return nil, err
		}
		if v, ok := lt.Entries[serialKey]; ok {
			return v, nil
		}
		return args[2], nil


	case "append":
		switch col := args[0].(type) {
		case []interface{}:
			result := make([]interface{}, len(col)+1)
			copy(result, col)
			result[len(col)] = args[1]
			return result, nil
		case *ArrayValue:
			// Enforce homogeneity
			elemKind := inferTypeKind(args[1])
			if col.ElementType != types.TypeUnknown && args[1] != nil &&
				types.Canonical(elemKind) != types.Canonical(col.ElementType) {
				return nil, fmt.Errorf(
					"TypeError: cannot append %s to array of %s",
					typeKindName(elemKind), typeKindName(col.ElementType),
				)
			}
			newElems := make([]interface{}, len(col.Elements)+1)
			copy(newElems, col.Elements)
			newElems[len(col.Elements)] = args[1]
			et := col.ElementType
			if et == types.TypeUnknown && args[1] != nil {
				et = types.Canonical(elemKind)
			}
			return &ArrayValue{ElementType: et, Elements: newElems}, nil
		default:
			return nil, fmt.Errorf("TypeError: append expects list or array, got %s", typeKindName(inferTypeKind(args[0])))
		}
	case "remove":
		list, ok := args[0].([]interface{})
		if !ok {
			return nil, NewRuntimeError("remove expects a list as first argument")
		}
		idx, err := ToNumber(args[1])
		if err != nil {
			return nil, NewRuntimeError("remove expects a number as second argument")
		}
		index := int(idx)
		if index < 0 || index >= len(list) {
			return nil, NewRuntimeError("list index out of bounds")
		}
		result := make([]interface{}, len(list)-1)
		copy(result, list[:index])
		copy(result[index:], list[index+1:])
		return result, nil
	case "insert":
		list, ok := args[0].([]interface{})
		if !ok {
			return nil, NewRuntimeError("insert expects a list as first argument")
		}
		idx, err := ToNumber(args[1])
		if err != nil {
			return nil, NewRuntimeError("insert expects a number as second argument")
		}
		index := int(idx)
		if index < 0 || index > len(list) {
			return nil, NewRuntimeError("list index out of bounds")
		}
		result := make([]interface{}, len(list)+1)
		copy(result, list[:index])
		result[index] = args[2]
		copy(result[index+1:], list[index:])
		return result, nil
	case "sort":
		list, ok := args[0].([]interface{})
		if !ok {
			return nil, NewRuntimeError("sort expects a list")
		}
		result := make([]interface{}, len(list))
		copy(result, list)
		sort.Slice(result, func(i, j int) bool {
			a, errA := ToNumber(result[i])
			b, errB := ToNumber(result[j])
			if errA == nil && errB == nil {
				return a < b
			}
			return ToString(result[i]) < ToString(result[j])
		})
		return result, nil
	case "reverse":
		list, ok := args[0].([]interface{})
		if !ok {
			return nil, NewRuntimeError("reverse expects a list")
		}
		result := make([]interface{}, len(list))
		for i, item := range list {
			result[len(list)-1-i] = item
		}
		return result, nil
	case "sum":
		switch col := args[0].(type) {
		case []interface{}:
			total := 0.0
			for _, item := range col {
				n, err := ToNumber(item)
				if err != nil {
					return nil, fmt.Errorf("TypeError: sum requires a list or array of numbers, got %s element", typeKindName(inferTypeKind(item)))
				}
				total += n
			}
			return total, nil
		case *ArrayValue:
			if col.ElementType != types.TypeUnknown && !types.IsNumeric(col.ElementType) {
				return nil, fmt.Errorf("TypeError: sum requires a number array, got array of %s", typeKindName(col.ElementType))
			}
			total := 0.0
			for _, item := range col.Elements {
				n, err := ToNumber(item)
				if err != nil {
					return nil, fmt.Errorf("TypeError: sum requires a number array")
				}
				total += n
			}
			return total, nil
		default:
			return nil, fmt.Errorf("TypeError: sum expects list or array, got %s", typeKindName(inferTypeKind(args[0])))
		}
	case "unique":
		list, ok := args[0].([]interface{})
		if !ok {
			return nil, NewRuntimeError("unique expects a list")
		}
		seen := make(map[string]bool)
		var result []interface{}
		for _, item := range list {
			key := fmt.Sprintf("%v", item)
			if !seen[key] {
				seen[key] = true
				result = append(result, item)
			}
		}
		if result == nil {
			result = []interface{}{}
		}
		return result, nil
	case "first":
		switch col := args[0].(type) {
		case []interface{}:
			if len(col) == 0 {
				return nil, fmt.Errorf("RuntimeError: first called on empty list")
			}
			return col[0], nil
		case *ArrayValue:
			if len(col.Elements) == 0 {
				return nil, fmt.Errorf("RuntimeError: first called on empty array")
			}
			return col.Elements[0], nil
		default:
			return nil, fmt.Errorf("TypeError: first expects list or array, got %s", typeKindName(inferTypeKind(args[0])))
		}
	case "last":
		switch col := args[0].(type) {
		case []interface{}:
			if len(col) == 0 {
				return nil, fmt.Errorf("RuntimeError: last called on empty list")
			}
			return col[len(col)-1], nil
		case *ArrayValue:
			if len(col.Elements) == 0 {
				return nil, fmt.Errorf("RuntimeError: last called on empty array")
			}
			return col.Elements[len(col.Elements)-1], nil
		default:
			return nil, fmt.Errorf("TypeError: last expects list or array, got %s", typeKindName(inferTypeKind(args[0])))
		}
	case "flatten":
		list, ok := args[0].([]interface{})
		if !ok {
			return nil, NewRuntimeError("flatten expects a list")
		}
		var result []interface{}
		for _, item := range list {
			if sublist, ok := item.([]interface{}); ok {
				result = append(result, sublist...)
			} else {
				result = append(result, item)
			}
		}
		if result == nil {
			result = []interface{}{}
		}
		return result, nil
	case "count":
		switch col := args[0].(type) {
		case []interface{}:
			return float64(len(col)), nil
		case *ArrayValue:
			return float64(len(col.Elements)), nil
		case *LookupTableValue:
			return float64(len(col.Entries)), nil
		case string:
			return float64(len(col)), nil
		default:
			return nil, fmt.Errorf("TypeError: count expects list, array, lookup table, or text; got %s", typeKindName(inferTypeKind(args[0])))
		}
	case "slice":
		list, ok := args[0].([]interface{})
		if !ok {
			return nil, NewRuntimeError("slice expects a list as first argument")
		}
		start, err := ToNumber(args[1])
		if err != nil {
			return nil, NewRuntimeError("slice expects a number as second argument")
		}
		end, err := ToNumber(args[2])
		if err != nil {
			return nil, NewRuntimeError("slice expects a number as third argument")
		}
		s := int(start)
		e := int(end)
		if s < 0 {
			s = 0
		}
		if e > len(list) {
			e = len(list)
		}
		if s >= e {
			return []interface{}{}, nil
		}
		result := make([]interface{}, e-s)
		copy(result, list[s:e])
		return result, nil

	// ── I/O ───────────────────────────────────────────────────────────────────
	case "ask":
		if len(args) > 0 {
			fmt.Print(ToString(args[0]))
		}
		reader := bufio.NewReader(os.Stdin)
		line, err := reader.ReadString('\n')
		if err != nil && len(line) == 0 {
			return "", nil
		}
		return strings.TrimRight(line, "\r\n"), nil

	// ─── Lookup table functions ───────────────────────────────────────────────
	case "keys":
		if len(args) != 1 {
			return nil, fmt.Errorf("keys() expects 1 argument")
		}
		lt, ok := args[0].(*LookupTableValue)
		if !ok {
			return nil, fmt.Errorf("TypeError: keys() expects a lookup table, got %s", typeKindName(inferTypeKind(args[0])))
		}
		result := make([]interface{}, 0, len(lt.KeyOrder))
		for _, k := range lt.KeyOrder {
			orig, _, ok := types.DeserializeKey(k)
			if ok {
				result = append(result, orig)
			} else {
				result = append(result, k)
			}
		}
		return result, nil

	case "values":
		if len(args) != 1 {
			return nil, fmt.Errorf("values() expects 1 argument")
		}
		lt, ok := args[0].(*LookupTableValue)
		if !ok {
			return nil, fmt.Errorf("TypeError: values() expects a lookup table, got %s", typeKindName(inferTypeKind(args[0])))
		}
		result := make([]interface{}, 0, len(lt.KeyOrder))
		for _, k := range lt.KeyOrder {
			result = append(result, lt.Entries[k])
		}
		return result, nil

	case "table_remove":
		if len(args) != 2 {
			return nil, fmt.Errorf("table_remove() expects 2 arguments")
		}
		lt, ok := args[0].(*LookupTableValue)
		if !ok {
			return nil, fmt.Errorf("TypeError: table_remove() expects a lookup table, got %s", typeKindName(inferTypeKind(args[0])))
		}
		serialKey, err := types.SerializeKey(args[1])
		if err != nil {
			return nil, err
		}
		// Return a new table without the key (functional style)
		newTable := types.NewLookupTable()
		for _, k := range lt.KeyOrder {
			if k != serialKey {
				newTable.Set(k, lt.Entries[k])
			}
		}
		return newTable, nil

	case "table_has":
		if len(args) != 2 {
			return nil, fmt.Errorf("table_has() expects 2 arguments")
		}
		lt, ok := args[0].(*LookupTableValue)
		if !ok {
			return nil, fmt.Errorf("TypeError: table_has() expects a lookup table, got %s", typeKindName(inferTypeKind(args[0])))
		}
		serialKey, err := types.SerializeKey(args[1])
		if err != nil {
			return nil, err
		}
		_, exists := lt.Entries[serialKey]
		return exists, nil

	}

	return nil, NewRuntimeError("unknown built-in function: " + name)
}
