package stdlib

import (
	"fmt"
	"sort"
	"strings"

	vm "github.com/Advik-B/english/astvm"
	"github.com/Advik-B/english/runtime"
	"github.com/Advik-B/english/types"
)

func evalList(name string, args []vm.Value) (vm.Value, error) {
	switch name {
	case "append":
		switch col := args[0].(type) {
		case []any:
			result := make([]any, len(col)+1)
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
			newElems := make([]any, len(col.Elements)+1)
			copy(newElems, col.Elements)
			newElems[len(col.Elements)] = args[1]
			et := col.ElementType
			if et == types.TypeUnknown && args[1] != nil {
				et = types.Canonical(elemKind)
			}
			return &types.ArrayValue{ElementType: et, Elements: newElems}, nil
		default:
			return nil, fmt.Errorf("TypeError: append expects list or array, got %s", types.NameOf(args[0]))
		}
	case "remove":
		list, ok := args[0].([]any)
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
		result := make([]any, len(list)-1)
		copy(result, list[:index])
		copy(result[index:], list[index+1:])
		return result, nil
	case "insert":
		list, ok := args[0].([]any)
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
		result := make([]any, len(list)+1)
		copy(result, list[:index])
		result[index] = args[2]
		copy(result[index+1:], list[index:])
		return result, nil
	case "sort":
		list, ok := args[0].([]any)
		if !ok {
			return nil, vm.NewRuntimeError("sort expects a list")
		}
		result := make([]any, len(list))
		copy(result, list)
		sort.SliceStable(result, orderBefore(result))
		return result, nil
	case "reverse":
		list, ok := args[0].([]any)
		if !ok {
			return nil, vm.NewRuntimeError("reverse expects a list")
		}
		result := make([]any, len(list))
		for i, item := range list {
			result[len(list)-1-i] = item
		}
		return result, nil
	case "sum":
		switch col := args[0].(type) {
		case []any:
			total := 0.0
			for _, item := range col {
				n, err := vm.ToNumber(item)
				if err != nil {
					return nil, fmt.Errorf("TypeError: sum requires a list or array of numbers, got %s element", types.NameOf(item))
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
			return nil, fmt.Errorf("TypeError: sum expects list or array, got %s", types.NameOf(args[0]))
		}
	case "unique":
		list, ok := args[0].([]any)
		if !ok {
			return nil, vm.NewRuntimeError("unique expects a list")
		}
		// Distinctness follows the language's own equality, which knows that
		// the number 1 and the text "1" are different values. Keying on their
		// rendered form made them the same, so unique([1, "1"]) returned one
		// element.
		result := []any{}
		for _, item := range list {
			duplicate := false
			for _, kept := range result {
				if runtime.Equals(kept, item) {
					duplicate = true
					break
				}
			}
			if !duplicate {
				result = append(result, item)
			}
		}
		return result, nil
	case "first":
		switch col := args[0].(type) {
		case []any:
			if len(col) == 0 {
				return nil, vm.NewRuntimeError("first called on empty list")
			}
			return col[0], nil
		case *types.ArrayValue:
			if len(col.Elements) == 0 {
				return nil, vm.NewRuntimeError("first called on empty array")
			}
			return col.Elements[0], nil
		default:
			return nil, fmt.Errorf("TypeError: first expects list or array, got %s", types.NameOf(args[0]))
		}
	case "last":
		switch col := args[0].(type) {
		case []any:
			if len(col) == 0 {
				return nil, vm.NewRuntimeError("last called on empty list")
			}
			return col[len(col)-1], nil
		case *types.ArrayValue:
			if len(col.Elements) == 0 {
				return nil, vm.NewRuntimeError("last called on empty array")
			}
			return col.Elements[len(col.Elements)-1], nil
		default:
			return nil, fmt.Errorf("TypeError: last expects list or array, got %s", types.NameOf(args[0]))
		}
	case "flatten":
		list, ok := args[0].([]any)
		if !ok {
			return nil, vm.NewRuntimeError("flatten expects a list")
		}
		var result []any
		for _, item := range list {
			if sublist, ok := item.([]any); ok {
				result = append(result, sublist...)
			} else {
				result = append(result, item)
			}
		}
		if result == nil {
			result = []any{}
		}
		return result, nil
	case "count":
		switch col := args[0].(type) {
		case []any:
			return float64(len(col)), nil
		case *types.ArrayValue:
			return float64(len(col.Elements)), nil
		case *types.LookupTableValue:
			return float64(len(col.Entries)), nil
		case string:
			return float64(len(col)), nil
		default:
			return nil, fmt.Errorf("TypeError: count expects list, array, lookup table, or text; got %s", types.NameOf(args[0]))
		}
	case "slice":
		list, ok := args[0].([]any)
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
			return []any{}, nil
		}
		result := make([]any, e-s)
		copy(result, list[s:e])
		return result, nil
	case "average":
		lst, err := requireList("average", args[0])
		if err != nil {
			return nil, err
		}
		if len(lst) == 0 {
			return nil, vm.NewRuntimeError("average called on empty list")
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
			return nil, vm.NewRuntimeError("min_value called on empty list")
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
			return nil, vm.NewRuntimeError("max_value called on empty list")
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
		result := make([]any, len(lst))
		copy(result, lst)
		sort.SliceStable(result, func(i, j int) bool {
			return compareValues(result[i], result[j]) > 0
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
		result := make([]any, length)
		for i := 0; i < length; i++ {
			result[i] = []any{lst[i], other[i]}
		}
		return result, nil
	}
	return nil, vm.NewRuntimeError("unknown list function: " + name)
}

// orderBefore returns the ordering predicate used by sort and sorted_desc.
//
// A comparison must be consistent for every pair, or the result is undefined.
// The previous comparator decided per pair: numeric when both happened to be
// numbers, and textual otherwise, which is not transitive on a mixed list — so
// sort.Slice was free to produce anything at all.
//
// Values are ordered by type first, then within a type, so a mixed list has a
// definite order rather than an accidental one.
func orderBefore(items []any) func(i, j int) bool {
	return func(i, j int) bool {
		return compareValues(items[i], items[j]) < 0
	}
}

// compareValues orders two values: negative if a sorts before b, positive if
// after, zero if they sort together.
func compareValues(a, b any) int {
	ra, rb := sortRank(a), sortRank(b)
	if ra != rb {
		return ra - rb
	}
	switch ra {
	case rankNumber:
		x, _ := vm.ToNumber(a)
		y, _ := vm.ToNumber(b)
		switch {
		case x < y:
			return -1
		case x > y:
			return 1
		}
		return 0
	case rankBool:
		x, y := a.(bool), b.(bool)
		switch {
		case !x && y:
			return -1
		case x && !y:
			return 1
		}
		return 0
	case rankText:
		return strings.Compare(a.(string), b.(string))
	}
	// Everything else keeps its relative order, which SliceStable preserves.
	return 0
}

// Sort ranks: values of different types sort in this order.
const (
	rankNothing = iota
	rankBool
	rankNumber
	rankText
	rankOther
)

func sortRank(v any) int {
	switch v.(type) {
	case nil:
		return rankNothing
	case bool:
		return rankBool
	case string:
		return rankText
	}
	if _, err := vm.ToNumber(v); err == nil {
		return rankNumber
	}
	return rankOther
}
