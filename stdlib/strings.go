package stdlib

import (
	"fmt"
	"strconv"
	"strings"

	vm "github.com/Advik-B/english/astvm"
	"github.com/Advik-B/english/types"
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
		result := make([]any, len(parts))
		for i, part := range parts {
			result[i] = part
		}
		return result, nil
	case "join":
		list, ok := args[0].([]any)
		if !ok {
			return nil, fmt.Errorf("TypeError: join expects list, got %s", types.NameOf(args[0]))
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
		// Positions count characters, not bytes: slicing by byte offset can
		// cut a multi-byte character in half and produce invalid text.
		runes := []rune(text)
		s := int(start)
		l := int(length)
		if s < 0 || s > len(runes) {
			return nil, vm.NewRuntimeError(fmt.Sprintf("substring start index %d out of range", s))
		}
		end := s + l
		if end > len(runes) {
			end = len(runes)
		}
		if end < s {
			end = s
		}
		return string(runes[s:end]), nil
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
		return padTo(text, int(width), padChar, padLeft), nil
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
		return padTo(text, int(width), padChar, padRight), nil
	case "to_number":
		text, err := requireText("to_number", args[0])
		if err != nil {
			return nil, err
		}
		// The whole text must be a number. Sscanf stops at the first character
		// it cannot use, so to_number("12abc") quietly returned 12 while
		// "12abc" cast to number correctly refused.
		f, err := strconv.ParseFloat(strings.TrimSpace(text), 64)
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
		case []any:
			return len(v) == 0, nil
		case *types.ArrayValue:
			return len(v.Elements) == 0, nil
		case *types.LookupTableValue:
			return len(v.Entries) == 0, nil
		case nil:
			return true, nil
		}
		// Emptiness is not a property of a number or a boolean. Answering
		// false for them hid the mistake in a language that otherwise refuses
		// every implicit conversion.
		return nil, vm.NewRuntimeError(fmt.Sprintf(
			"is_empty expects text or a collection, got %s", types.NameOf(args[0])))
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
		return padTo(text, int(width), fillChar, padCentre), nil
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
		if len([]rune(text)) >= w {
			return text, nil
		}
		prefix := ""
		body := text
		if len(body) > 0 && (body[0] == '+' || body[0] == '-') {
			prefix = string(body[0])
			body = body[1:]
		}
		zeros := w - len([]rune(prefix)) - len([]rune(body))
		if zeros < 0 {
			zeros = 0
		}
		return prefix + strings.Repeat("0", zeros) + body, nil
	}
	return nil, vm.NewRuntimeError("unknown string function: " + name)
}

// Padding sides.
const (
	padLeft = iota
	padRight
	padCentre
)

// padTo pads text to a width, measured in characters.
//
// Both the width and the pad character used to be handled by byte: the width
// was compared against len(text), so any non-ASCII text was padded to the
// wrong visible length, and the pad character was sliced with [:1], which
// takes one *byte* of it — so padding with a multi-byte character produced
// invalid text.
func padTo(text string, width int, pad string, side int) string {
	fill := []rune(pad)
	if len(fill) == 0 {
		fill = []rune{' '}
	}
	filler := string(fill[0])

	missing := width - len([]rune(text))
	if missing <= 0 {
		return text
	}
	switch side {
	case padLeft:
		return strings.Repeat(filler, missing) + text
	case padRight:
		return text + strings.Repeat(filler, missing)
	}
	left := missing / 2
	return strings.Repeat(filler, left) + text + strings.Repeat(filler, missing-left)
}
