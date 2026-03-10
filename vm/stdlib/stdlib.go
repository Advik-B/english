package stdlib

import (
	"bufio"
	"english/vm"
	"english/vm/types"
	"fmt"
	"math"
	"math/rand"
	"os"
	"sort"
	"strings"
)

// Register registers all standard library functions into env.
func Register(env *vm.Environment) {
	registerMathConstants(env)
	registerMathFunctions(env)
	registerStringFunctions(env)
	registerListFunctions(env)
	registerIOFunctions(env)
	registerLookupTableFunctions(env)
	registerNumberFunctions(env)
}

// Eval evaluates a built-in function by name with the provided arguments.
func Eval(name string, args []vm.Value) (vm.Value, error) {
	switch name {
	// ── Math ──────────────────────────────────────────────────────────────────
	case "sqrt":
		x, err := vm.ToNumber(args[0])
		if err != nil {
			return nil, err
		}
		return math.Sqrt(x), nil
	case "pow":
		base, err := vm.ToNumber(args[0])
		if err != nil {
			return nil, err
		}
		exp, err := vm.ToNumber(args[1])
		if err != nil {
			return nil, err
		}
		return math.Pow(base, exp), nil
	case "abs":
		x, err := vm.ToNumber(args[0])
		if err != nil {
			return nil, err
		}
		return math.Abs(x), nil
	case "floor":
		x, err := vm.ToNumber(args[0])
		if err != nil {
			return nil, err
		}
		return math.Floor(x), nil
	case "ceil":
		x, err := vm.ToNumber(args[0])
		if err != nil {
			return nil, err
		}
		return math.Ceil(x), nil
	case "round":
		x, err := vm.ToNumber(args[0])
		if err != nil {
			return nil, err
		}
		return math.Round(x), nil
	case "min":
		a, err := vm.ToNumber(args[0])
		if err != nil {
			return nil, err
		}
		b, err := vm.ToNumber(args[1])
		if err != nil {
			return nil, err
		}
		return math.Min(a, b), nil
	case "max":
		a, err := vm.ToNumber(args[0])
		if err != nil {
			return nil, err
		}
		b, err := vm.ToNumber(args[1])
		if err != nil {
			return nil, err
		}
		return math.Max(a, b), nil
	case "sin":
		x, err := vm.ToNumber(args[0])
		if err != nil {
			return nil, err
		}
		return math.Sin(x), nil
	case "cos":
		x, err := vm.ToNumber(args[0])
		if err != nil {
			return nil, err
		}
		return math.Cos(x), nil
	case "tan":
		x, err := vm.ToNumber(args[0])
		if err != nil {
			return nil, err
		}
		return math.Tan(x), nil
	case "log":
		x, err := vm.ToNumber(args[0])
		if err != nil {
			return nil, err
		}
		return math.Log(x), nil
	case "log10":
		x, err := vm.ToNumber(args[0])
		if err != nil {
			return nil, err
		}
		return math.Log10(x), nil
	case "log2":
		x, err := vm.ToNumber(args[0])
		if err != nil {
			return nil, err
		}
		return math.Log2(x), nil
	case "exp":
		x, err := vm.ToNumber(args[0])
		if err != nil {
			return nil, err
		}
		return math.Exp(x), nil
	case "random":
		return rand.Float64(), nil
	case "random_between":
		a, err := vm.ToNumber(args[0])
		if err != nil {
			return nil, vm.NewRuntimeError("random_between expects a number as first argument")
		}
		b, err := vm.ToNumber(args[1])
		if err != nil {
			return nil, vm.NewRuntimeError("random_between expects a number as second argument")
		}
		if a > b {
			return nil, vm.NewRuntimeError("random_between: min must be less than or equal to max")
		}
		return a + rand.Float64()*(b-a), nil
	case "is_nan":
		x, err := vm.ToNumber(args[0])
		if err != nil {
			return true, nil // non-numeric is NaN-like
		}
		return math.IsNaN(x), nil
	case "is_infinite":
		x, err := vm.ToNumber(args[0])
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
		sep := vm.ToString(args[1])
		parts := strings.Split(text, sep)
		result := make([]interface{}, len(parts))
		for i, part := range parts {
			result[i] = part
		}
		return result, nil
	case "join":
		list, ok := args[0].([]interface{})
		if !ok {
			return nil, fmt.Errorf("TypeError: join expects list, got %s", kindName(args[0]))
		}
		sep := vm.ToString(args[1])
		strs := make([]string, len(list))
		for i, item := range list {
			strs[i] = vm.ToString(item)
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
		old := vm.ToString(args[1])
		newStr := vm.ToString(args[2])
		return strings.ReplaceAll(text, old, newStr), nil
	case "contains":
		text, err := requireText("contains", args[0])
		if err != nil {
			return nil, err
		}
		substr := vm.ToString(args[1])
		return strings.Contains(text, substr), nil
	case "starts_with":
		text, err := requireText("starts_with", args[0])
		if err != nil {
			return nil, err
		}
		prefix := vm.ToString(args[1])
		return strings.HasPrefix(text, prefix), nil
	case "ends_with":
		text, err := requireText("ends_with", args[0])
		if err != nil {
			return nil, err
		}
		suffix := vm.ToString(args[1])
		return strings.HasSuffix(text, suffix), nil
	case "index_of":
		text, err := requireText("index_of", args[0])
		if err != nil {
			return nil, err
		}
		search := vm.ToString(args[1])
		idx := strings.Index(text, search)
		return float64(idx), nil
	case "substring":
		text, err := requireText("substring", args[0])
		if err != nil {
			return nil, err
		}
		start, err := vm.ToNumber(args[1])
		if err != nil {
			return nil, vm.NewRuntimeError("substring expects a number as second argument")
		}
		length, err := vm.ToNumber(args[2])
		if err != nil {
			return nil, vm.NewRuntimeError("substring expects a number as third argument")
		}
		s := int(start)
		l := int(length)
		if s < 0 || s > len(text) {
			return nil, vm.NewRuntimeError(fmt.Sprintf("substring start index %d out of range", s))
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
		n, err := vm.ToNumber(args[1])
		if err != nil {
			return nil, vm.NewRuntimeError("str_repeat expects a number as second argument")
		}
		if int(n) < 0 {
			return nil, vm.NewRuntimeError("str_repeat count must be non-negative")
		}
		return strings.Repeat(text, int(n)), nil
	case "count_occurrences":
		text, err := requireText("count_occurrences", args[0])
		if err != nil {
			return nil, err
		}
		sub := vm.ToString(args[1])
		return float64(strings.Count(text, sub)), nil
	case "pad_left":
		text, err := requireText("pad_left", args[0])
		if err != nil {
			return nil, err
		}
		width, err := vm.ToNumber(args[1])
		if err != nil {
			return nil, vm.NewRuntimeError("pad_left expects a number as second argument")
		}
		padChar := " "
		if len(args) > 2 {
			padChar = vm.ToString(args[2])
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
		width, err := vm.ToNumber(args[1])
		if err != nil {
			return nil, vm.NewRuntimeError("pad_right expects a number as second argument")
		}
		padChar := " "
		if len(args) > 2 {
			padChar = vm.ToString(args[2])
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
			return nil, vm.NewRuntimeError(fmt.Sprintf("cannot convert '%s' to a number", text))
		}
		return f, nil
	case "to_string":
		return vm.ToString(args[0]), nil
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
		text, err := requireText("title", args[0])
		if err != nil {
			return nil, err
		}
		// str.title(): uppercase the first letter of each word.
		// Implemented without the deprecated strings.Title.
		words := strings.Fields(strings.ToLower(text))
		for i, w := range words {
			if len(w) > 0 {
				words[i] = strings.ToUpper(w[:1]) + w[1:]
			}
		}
		return strings.Join(words, " "), nil
	case "capitalize":
		text, err := requireText("capitalize", args[0])
		if err != nil {
			return nil, err
		}
		if text == "" {
			return text, nil
		}
		return strings.ToUpper(text[:1]) + strings.ToLower(text[1:]), nil
	case "swapcase":
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
		text, err := requireText("trim_left", args[0])
		if err != nil {
			return nil, err
		}
		return strings.TrimLeftFunc(text, func(r rune) bool {
			return r == ' ' || r == '\t' || r == '\n' || r == '\r'
		}), nil
	case "trim_right":
		text, err := requireText("trim_right", args[0])
		if err != nil {
			return nil, err
		}
		return strings.TrimRightFunc(text, func(r rune) bool {
			return r == ' ' || r == '\t' || r == '\n' || r == '\r'
		}), nil
	case "is_digit":
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
		text, err := requireText("is_space", args[0])
		if err != nil {
			return nil, err
		}
		if text == "" {
			return false, nil
		}
		return strings.TrimSpace(text) == "", nil
	case "is_upper":
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
			var ferr error
			fillChar, ferr = requireText("center", args[2])
			if ferr != nil {
				return nil, ferr
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
		x, err := requireNumber("is_integer", args[0])
		if err != nil {
			return nil, err
		}
		return x == math.Trunc(x), nil
	case "clamp":
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
		lst, err := requireList("average", args[0])
		if err != nil {
			return nil, err
		}
		if len(lst) == 0 {
			return nil, fmt.Errorf("RuntimeError: average called on empty list")
		}
		total := 0.0
		for _, item := range lst {
			n, err := vm.ToNumber(item)
			if err != nil {
				return nil, fmt.Errorf("TypeError: average requires a list of numbers")
			}
			total += n
		}
		return total / float64(len(lst)), nil
	case "min_value":
		lst, err := requireList("min_value", args[0])
		if err != nil {
			return nil, err
		}
		if len(lst) == 0 {
			return nil, fmt.Errorf("RuntimeError: min_value called on empty list")
		}
		best, err := vm.ToNumber(lst[0])
		if err != nil {
			return nil, fmt.Errorf("TypeError: min_value requires a list of numbers")
		}
		for _, item := range lst[1:] {
			n, err := vm.ToNumber(item)
			if err != nil {
				return nil, fmt.Errorf("TypeError: min_value requires a list of numbers")
			}
			if n < best {
				best = n
			}
		}
		return best, nil
	case "max_value":
		lst, err := requireList("max_value", args[0])
		if err != nil {
			return nil, err
		}
		if len(lst) == 0 {
			return nil, fmt.Errorf("RuntimeError: max_value called on empty list")
		}
		best, err := vm.ToNumber(lst[0])
		if err != nil {
			return nil, fmt.Errorf("TypeError: max_value requires a list of numbers")
		}
		for _, item := range lst[1:] {
			n, err := vm.ToNumber(item)
			if err != nil {
				return nil, fmt.Errorf("TypeError: max_value requires a list of numbers")
			}
			if n > best {
				best = n
			}
		}
		return best, nil
	case "any_true":
		lst, err := requireList("any_true", args[0])
		if err != nil {
			return nil, err
		}
		for _, item := range lst {
			b, err := vm.ToBool(item)
			if err != nil {
				return nil, fmt.Errorf("TypeError: any_true requires a list of boolean values")
			}
			if b {
				return true, nil
			}
		}
		return false, nil
	case "all_true":
		lst, err := requireList("all_true", args[0])
		if err != nil {
			return nil, err
		}
		for _, item := range lst {
			b, err := vm.ToBool(item)
			if err != nil {
				return nil, fmt.Errorf("TypeError: all_true requires a list of boolean values")
			}
			if !b {
				return false, nil
			}
		}
		return true, nil
	case "product":
		lst, err := requireList("product", args[0])
		if err != nil {
			return nil, err
		}
		result := 1.0
		for _, item := range lst {
			n, err := vm.ToNumber(item)
			if err != nil {
				return nil, fmt.Errorf("TypeError: product requires a list of numbers")
			}
			result *= n
		}
		return result, nil
	case "sorted_desc":
		lst, err := requireList("sorted_desc", args[0])
		if err != nil {
			return nil, err
		}
		result := make([]interface{}, len(lst))
		copy(result, lst)
		sort.Slice(result, func(i, j int) bool {
			a, errA := vm.ToNumber(result[i])
			b, errB := vm.ToNumber(result[j])
			if errA == nil && errB == nil {
				return a > b
			}
			return vm.ToString(result[i]) > vm.ToString(result[j])
		})
		return result, nil
	case "zip_with":
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

	// ── List (existing) ────────────────────────────────────────────────────────
	case "append":
		switch col := args[0].(type) {
		case []interface{}:
			result := make([]interface{}, len(col)+1)
			copy(result, col)
			result[len(col)] = args[1]
			return result, nil
		case *types.ArrayValue:
			elemKind := types.Infer(args[1])
			if col.ElementType != types.TypeUnknown && args[1] != nil &&
				types.Canonical(elemKind) != types.Canonical(col.ElementType) {
				return nil, fmt.Errorf(
					"TypeError: cannot append %s to array of %s",
					types.Name(elemKind), types.Name(col.ElementType),
				)
			}
			newElems := make([]interface{}, len(col.Elements)+1)
			copy(newElems, col.Elements)
			newElems[len(col.Elements)] = args[1]
			et := col.ElementType
			if et == types.TypeUnknown && args[1] != nil {
				et = types.Canonical(elemKind)
			}
			return &types.ArrayValue{ElementType: et, Elements: newElems}, nil
		default:
			return nil, fmt.Errorf("TypeError: append expects list or array, got %s", kindName(args[0]))
		}
	case "remove":
		list, ok := args[0].([]interface{})
		if !ok {
			return nil, vm.NewRuntimeError("remove expects a list as first argument")
		}
		idx, err := vm.ToNumber(args[1])
		if err != nil {
			return nil, vm.NewRuntimeError("remove expects a number as second argument")
		}
		index := int(idx)
		if index < 0 || index >= len(list) {
			return nil, vm.NewRuntimeError("list index out of bounds")
		}
		result := make([]interface{}, len(list)-1)
		copy(result, list[:index])
		copy(result[index:], list[index+1:])
		return result, nil
	case "insert":
		list, ok := args[0].([]interface{})
		if !ok {
			return nil, vm.NewRuntimeError("insert expects a list as first argument")
		}
		idx, err := vm.ToNumber(args[1])
		if err != nil {
			return nil, vm.NewRuntimeError("insert expects a number as second argument")
		}
		index := int(idx)
		if index < 0 || index > len(list) {
			return nil, vm.NewRuntimeError("list index out of bounds")
		}
		result := make([]interface{}, len(list)+1)
		copy(result, list[:index])
		result[index] = args[2]
		copy(result[index+1:], list[index:])
		return result, nil
	case "sort":
		list, ok := args[0].([]interface{})
		if !ok {
			return nil, vm.NewRuntimeError("sort expects a list")
		}
		result := make([]interface{}, len(list))
		copy(result, list)
		sort.Slice(result, func(i, j int) bool {
			a, errA := vm.ToNumber(result[i])
			b, errB := vm.ToNumber(result[j])
			if errA == nil && errB == nil {
				return a < b
			}
			return vm.ToString(result[i]) < vm.ToString(result[j])
		})
		return result, nil
	case "reverse":
		list, ok := args[0].([]interface{})
		if !ok {
			return nil, vm.NewRuntimeError("reverse expects a list")
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
				n, err := vm.ToNumber(item)
				if err != nil {
					return nil, fmt.Errorf("TypeError: sum requires a list or array of numbers, got %s element", kindName(item))
				}
				total += n
			}
			return total, nil
		case *types.ArrayValue:
			if col.ElementType != types.TypeUnknown && !types.IsNumeric(col.ElementType) {
				return nil, fmt.Errorf("TypeError: sum requires a number array, got array of %s", types.Name(col.ElementType))
			}
			total := 0.0
			for _, item := range col.Elements {
				n, err := vm.ToNumber(item)
				if err != nil {
					return nil, fmt.Errorf("TypeError: sum requires a number array")
				}
				total += n
			}
			return total, nil
		default:
			return nil, fmt.Errorf("TypeError: sum expects list or array, got %s", kindName(args[0]))
		}
	case "unique":
		list, ok := args[0].([]interface{})
		if !ok {
			return nil, vm.NewRuntimeError("unique expects a list")
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
		case *types.ArrayValue:
			if len(col.Elements) == 0 {
				return nil, fmt.Errorf("RuntimeError: first called on empty array")
			}
			return col.Elements[0], nil
		default:
			return nil, fmt.Errorf("TypeError: first expects list or array, got %s", kindName(args[0]))
		}
	case "last":
		switch col := args[0].(type) {
		case []interface{}:
			if len(col) == 0 {
				return nil, fmt.Errorf("RuntimeError: last called on empty list")
			}
			return col[len(col)-1], nil
		case *types.ArrayValue:
			if len(col.Elements) == 0 {
				return nil, fmt.Errorf("RuntimeError: last called on empty array")
			}
			return col.Elements[len(col.Elements)-1], nil
		default:
			return nil, fmt.Errorf("TypeError: last expects list or array, got %s", kindName(args[0]))
		}
	case "flatten":
		list, ok := args[0].([]interface{})
		if !ok {
			return nil, vm.NewRuntimeError("flatten expects a list")
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
		case *types.ArrayValue:
			return float64(len(col.Elements)), nil
		case *types.LookupTableValue:
			return float64(len(col.Entries)), nil
		case string:
			return float64(len(col)), nil
		default:
			return nil, fmt.Errorf("TypeError: count expects list, array, lookup table, or text; got %s", kindName(args[0]))
		}
	case "slice":
		list, ok := args[0].([]interface{})
		if !ok {
			return nil, vm.NewRuntimeError("slice expects a list as first argument")
		}
		start, err := vm.ToNumber(args[1])
		if err != nil {
			return nil, vm.NewRuntimeError("slice expects a number as second argument")
		}
		end, err := vm.ToNumber(args[2])
		if err != nil {
			return nil, vm.NewRuntimeError("slice expects a number as third argument")
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
			fmt.Print(vm.ToString(args[0]))
		}
		reader := bufio.NewReader(os.Stdin)
		line, err := reader.ReadString('\n')
		if err != nil && len(line) == 0 {
			return "", nil
		}
		return strings.TrimRight(line, "\r\n"), nil

	// ── Lookup table (existing) ────────────────────────────────────────────────
	case "keys":
		if len(args) != 1 {
			return nil, fmt.Errorf("keys() expects 1 argument")
		}
		lt, ok := args[0].(*types.LookupTableValue)
		if !ok {
			return nil, fmt.Errorf("TypeError: keys() expects a lookup table, got %s", kindName(args[0]))
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
		lt, ok := args[0].(*types.LookupTableValue)
		if !ok {
			return nil, fmt.Errorf("TypeError: values() expects a lookup table, got %s", kindName(args[0]))
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
		lt, ok := args[0].(*types.LookupTableValue)
		if !ok {
			return nil, fmt.Errorf("TypeError: table_remove() expects a lookup table, got %s", kindName(args[0]))
		}
		serialKey, err := types.SerializeKey(args[1])
		if err != nil {
			return nil, err
		}
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
		lt, ok := args[0].(*types.LookupTableValue)
		if !ok {
			return nil, fmt.Errorf("TypeError: table_has() expects a lookup table, got %s", kindName(args[0]))
		}
		serialKey, err := types.SerializeKey(args[1])
		if err != nil {
			return nil, err
		}
		_, exists := lt.Entries[serialKey]
		return exists, nil
	}

	return nil, vm.NewRuntimeError("unknown built-in function: " + name)
}

// ── Registration helpers ───────────────────────────────────────────────────────

// PredefinedNames returns the names of all constants registered by the stdlib.
// Pass these to vm.Check so the compile-time checker can catch redeclarations.
func PredefinedNames() []string {
	return []string{"pi", "e", "infinity"}
}

func registerMathConstants(env *vm.Environment) {
	env.Define("pi", math.Pi, true)
	env.Define("e", math.E, true)
	env.Define("infinity", math.Inf(1), true)
}

func registerMathFunctions(env *vm.Environment) {
	for _, name := range []string{"sqrt", "abs", "floor", "ceil", "round", "sin", "cos", "tan", "log", "log10", "log2", "exp", "is_nan", "is_infinite"} {
		n := name
		env.DefineFunction(n, &vm.FunctionValue{Name: n, Parameters: []string{"x"}, Body: nil, Closure: env})
	}
	env.DefineFunction("random", &vm.FunctionValue{Name: "random", Parameters: []string{}, Body: nil, Closure: env})
	env.DefineFunction("pow", &vm.FunctionValue{Name: "pow", Parameters: []string{"base", "exponent"}, Body: nil, Closure: env})
	env.DefineFunction("min", &vm.FunctionValue{Name: "min", Parameters: []string{"a", "b"}, Body: nil, Closure: env})
	env.DefineFunction("max", &vm.FunctionValue{Name: "max", Parameters: []string{"a", "b"}, Body: nil, Closure: env})
	env.DefineFunction("random_between", &vm.FunctionValue{Name: "random_between", Parameters: []string{"min", "max"}, Body: nil, Closure: env})
}

func registerStringFunctions(env *vm.Environment) {
	single := []string{
		"uppercase", "lowercase", "casefold", "trim", "to_number", "to_string", "is_empty",
		"title", "capitalize", "swapcase", "trim_left", "trim_right",
		"is_digit", "is_alpha", "is_alnum", "is_space", "is_upper", "is_lower",
	}
	for _, name := range single {
		n := name
		env.DefineFunction(n, &vm.FunctionValue{Name: n, Parameters: []string{"text"}, Body: nil, Closure: env})
	}
	env.DefineFunction("split", &vm.FunctionValue{Name: "split", Parameters: []string{"text", "separator"}, Body: nil, Closure: env})
	env.DefineFunction("join", &vm.FunctionValue{Name: "join", Parameters: []string{"list", "separator"}, Body: nil, Closure: env})
	env.DefineFunction("replace", &vm.FunctionValue{Name: "replace", Parameters: []string{"text", "old", "new"}, Body: nil, Closure: env})
	env.DefineFunction("contains", &vm.FunctionValue{Name: "contains", Parameters: []string{"text", "substring"}, Body: nil, Closure: env})
	env.DefineFunction("starts_with", &vm.FunctionValue{Name: "starts_with", Parameters: []string{"text", "prefix"}, Body: nil, Closure: env})
	env.DefineFunction("ends_with", &vm.FunctionValue{Name: "ends_with", Parameters: []string{"text", "suffix"}, Body: nil, Closure: env})
	env.DefineFunction("index_of", &vm.FunctionValue{Name: "index_of", Parameters: []string{"text", "search"}, Body: nil, Closure: env})
	env.DefineFunction("substring", &vm.FunctionValue{Name: "substring", Parameters: []string{"text", "start", "length"}, Body: nil, Closure: env})
	env.DefineFunction("str_repeat", &vm.FunctionValue{Name: "str_repeat", Parameters: []string{"text", "n"}, Body: nil, Closure: env})
	env.DefineFunction("count_occurrences", &vm.FunctionValue{Name: "count_occurrences", Parameters: []string{"text", "substring"}, Body: nil, Closure: env})
	env.DefineFunction("pad_left", &vm.FunctionValue{Name: "pad_left", Parameters: []string{"text", "width", "char"}, Body: nil, Closure: env})
	env.DefineFunction("pad_right", &vm.FunctionValue{Name: "pad_right", Parameters: []string{"text", "width", "char"}, Body: nil, Closure: env})
	env.DefineFunction("center", &vm.FunctionValue{Name: "center", Parameters: []string{"text", "width", "char"}, Body: nil, Closure: env})
	env.DefineFunction("zfill", &vm.FunctionValue{Name: "zfill", Parameters: []string{"text", "width"}, Body: nil, Closure: env})
}

func registerListFunctions(env *vm.Environment) {
	env.DefineFunction("append", &vm.FunctionValue{Name: "append", Parameters: []string{"list", "item"}, Body: nil, Closure: env})
	env.DefineFunction("remove", &vm.FunctionValue{Name: "remove", Parameters: []string{"list", "index"}, Body: nil, Closure: env})
	env.DefineFunction("insert", &vm.FunctionValue{Name: "insert", Parameters: []string{"list", "index", "item"}, Body: nil, Closure: env})
	env.DefineFunction("sort", &vm.FunctionValue{Name: "sort", Parameters: []string{"list"}, Body: nil, Closure: env})
	env.DefineFunction("reverse", &vm.FunctionValue{Name: "reverse", Parameters: []string{"list"}, Body: nil, Closure: env})
	env.DefineFunction("sum", &vm.FunctionValue{Name: "sum", Parameters: []string{"list"}, Body: nil, Closure: env})
	env.DefineFunction("unique", &vm.FunctionValue{Name: "unique", Parameters: []string{"list"}, Body: nil, Closure: env})
	env.DefineFunction("first", &vm.FunctionValue{Name: "first", Parameters: []string{"list"}, Body: nil, Closure: env})
	env.DefineFunction("last", &vm.FunctionValue{Name: "last", Parameters: []string{"list"}, Body: nil, Closure: env})
	env.DefineFunction("flatten", &vm.FunctionValue{Name: "flatten", Parameters: []string{"list"}, Body: nil, Closure: env})
	env.DefineFunction("count", &vm.FunctionValue{Name: "count", Parameters: []string{"list"}, Body: nil, Closure: env})
	env.DefineFunction("slice", &vm.FunctionValue{Name: "slice", Parameters: []string{"list", "start", "end"}, Body: nil, Closure: env})
	// Python-equivalent
	env.DefineFunction("average", &vm.FunctionValue{Name: "average", Parameters: []string{"list"}, Body: nil, Closure: env})
	env.DefineFunction("min_value", &vm.FunctionValue{Name: "min_value", Parameters: []string{"list"}, Body: nil, Closure: env})
	env.DefineFunction("max_value", &vm.FunctionValue{Name: "max_value", Parameters: []string{"list"}, Body: nil, Closure: env})
	env.DefineFunction("any_true", &vm.FunctionValue{Name: "any_true", Parameters: []string{"list"}, Body: nil, Closure: env})
	env.DefineFunction("all_true", &vm.FunctionValue{Name: "all_true", Parameters: []string{"list"}, Body: nil, Closure: env})
	env.DefineFunction("product", &vm.FunctionValue{Name: "product", Parameters: []string{"list"}, Body: nil, Closure: env})
	env.DefineFunction("sorted_desc", &vm.FunctionValue{Name: "sorted_desc", Parameters: []string{"list"}, Body: nil, Closure: env})
	env.DefineFunction("zip_with", &vm.FunctionValue{Name: "zip_with", Parameters: []string{"list", "other"}, Body: nil, Closure: env})
}

func registerIOFunctions(env *vm.Environment) {
	env.DefineFunction("ask", &vm.FunctionValue{Name: "ask", Parameters: []string{"prompt"}, Body: nil, Closure: env})
}

func registerLookupTableFunctions(env *vm.Environment) {
	env.DefineFunction("keys", &vm.FunctionValue{Name: "keys", Parameters: []string{"table"}, Body: nil, Closure: env})
	env.DefineFunction("values", &vm.FunctionValue{Name: "values", Parameters: []string{"table"}, Body: nil, Closure: env})
	env.DefineFunction("table_remove", &vm.FunctionValue{Name: "table_remove", Parameters: []string{"table", "key"}, Body: nil, Closure: env})
	env.DefineFunction("table_has", &vm.FunctionValue{Name: "table_has", Parameters: []string{"table", "key"}, Body: nil, Closure: env})
	// Python-equivalent
	env.DefineFunction("merge", &vm.FunctionValue{Name: "merge", Parameters: []string{"table", "other"}, Body: nil, Closure: env})
	env.DefineFunction("get_or_default", &vm.FunctionValue{Name: "get_or_default", Parameters: []string{"table", "key", "default"}, Body: nil, Closure: env})
}

func registerNumberFunctions(env *vm.Environment) {
	env.DefineFunction("is_integer", &vm.FunctionValue{Name: "is_integer", Parameters: []string{"x"}, Body: nil, Closure: env})
	env.DefineFunction("clamp", &vm.FunctionValue{Name: "clamp", Parameters: []string{"x", "min", "max"}, Body: nil, Closure: env})
	env.DefineFunction("sign", &vm.FunctionValue{Name: "sign", Parameters: []string{"x"}, Body: nil, Closure: env})
}

// ── Type-guard helpers ─────────────────────────────────────────────────────────

func requireText(fn string, arg vm.Value) (string, error) {
	s, ok := arg.(string)
	if !ok {
		return "", fmt.Errorf("TypeError: %s expects text, got %s", fn, kindName(arg))
	}
	return s, nil
}

func requireNumber(fn string, arg vm.Value) (float64, error) {
	n, err := vm.ToNumber(arg)
	if err != nil {
		return 0, fmt.Errorf("TypeError: %s expects number, got %s", fn, kindName(arg))
	}
	return n, nil
}

func requireList(fn string, arg vm.Value) ([]interface{}, error) {
	lst, ok := arg.([]interface{})
	if !ok {
		return nil, fmt.Errorf("TypeError: %s expects list, got %s", fn, kindName(arg))
	}
	return lst, nil
}

func requireLookupTable(fn string, arg vm.Value) (*types.LookupTableValue, error) {
	lt, ok := arg.(*types.LookupTableValue)
	if !ok {
		return nil, fmt.Errorf("TypeError: %s expects lookup table, got %s", fn, kindName(arg))
	}
	return lt, nil
}

func kindName(v vm.Value) string {
	switch v.(type) {
	case *vm.FunctionValue:
		return "function"
	case *vm.StructInstance:
		return "struct"
	case *vm.ReferenceValue:
		return "reference"
	case nil:
		return "nothing"
	default:
		return types.Name(types.Infer(v))
	}
}
