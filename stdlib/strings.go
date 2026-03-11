package stdlib

import (
	"english/astvm"
	"fmt"
	"strings"
)

func evalString(name string, args []vm.Value) (vm.Value, error) {
	switch name {
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
	case "title":
		text, err := requireText("title", args[0])
		if err != nil {
			return nil, err
		}
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
	}
	return nil, vm.NewRuntimeError("unknown string function: " + name)
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
