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
		return strings.ToUpper(ToString(args[0])), nil
	case "lowercase":
		return strings.ToLower(ToString(args[0])), nil
	case "casefold":
		// casefold converts text to a case-folded (lowercase) form suitable for
		// case-insensitive comparisons.  Equivalent to Python's str.casefold().
		return strings.ToLower(ToString(args[0])), nil
	case "split":
		text := ToString(args[0])
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
			return nil, NewRuntimeError("join expects a list as first argument")
		}
		sep := ToString(args[1])
		strs := make([]string, len(list))
		for i, item := range list {
			strs[i] = ToString(item)
		}
		return strings.Join(strs, sep), nil
	case "trim":
		return strings.TrimSpace(ToString(args[0])), nil
	case "replace":
		text := ToString(args[0])
		old := ToString(args[1])
		newStr := ToString(args[2])
		return strings.ReplaceAll(text, old, newStr), nil
	case "contains":
		text := ToString(args[0])
		substr := ToString(args[1])
		return strings.Contains(text, substr), nil
	case "starts_with":
		text := ToString(args[0])
		prefix := ToString(args[1])
		return strings.HasPrefix(text, prefix), nil
	case "ends_with":
		text := ToString(args[0])
		suffix := ToString(args[1])
		return strings.HasSuffix(text, suffix), nil
	case "index_of":
		text := ToString(args[0])
		search := ToString(args[1])
		idx := strings.Index(text, search)
		return float64(idx), nil
	case "substring":
		text := ToString(args[0])
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
		text := ToString(args[0])
		n, err := ToNumber(args[1])
		if err != nil {
			return nil, NewRuntimeError("str_repeat expects a number as second argument")
		}
		if int(n) < 0 {
			return nil, NewRuntimeError("str_repeat count must be non-negative")
		}
		return strings.Repeat(text, int(n)), nil
	case "count_occurrences":
		text := ToString(args[0])
		sub := ToString(args[1])
		return float64(strings.Count(text, sub)), nil
	case "pad_left":
		text := ToString(args[0])
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
		text := ToString(args[0])
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
		text := ToString(args[0])
		var f float64
		_, err := fmt.Sscanf(text, "%g", &f)
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

	// ── List ──────────────────────────────────────────────────────────────────
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
