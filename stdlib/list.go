package stdlib

import (
	"english/astvm"
	"english/astvm/types"
	"fmt"
	"sort"
)

func evalList(name string, args []vm.Value) (vm.Value, error) {
	switch name {
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
				return nil, vm.NewRuntimeError("first called on empty list")
			}
			return col[0], nil
		case *types.ArrayValue:
			if len(col.Elements) == 0 {
				return nil, vm.NewRuntimeError("first called on empty array")
			}
			return col.Elements[0], nil
		default:
			return nil, fmt.Errorf("TypeError: first expects list or array, got %s", kindName(args[0]))
		}
	case "last":
		switch col := args[0].(type) {
		case []interface{}:
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
	}
	return nil, vm.NewRuntimeError("unknown list function: " + name)
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
	env.DefineFunction("average", &vm.FunctionValue{Name: "average", Parameters: []string{"list"}, Body: nil, Closure: env})
	env.DefineFunction("min_value", &vm.FunctionValue{Name: "min_value", Parameters: []string{"list"}, Body: nil, Closure: env})
	env.DefineFunction("max_value", &vm.FunctionValue{Name: "max_value", Parameters: []string{"list"}, Body: nil, Closure: env})
	env.DefineFunction("any_true", &vm.FunctionValue{Name: "any_true", Parameters: []string{"list"}, Body: nil, Closure: env})
	env.DefineFunction("all_true", &vm.FunctionValue{Name: "all_true", Parameters: []string{"list"}, Body: nil, Closure: env})
	env.DefineFunction("product", &vm.FunctionValue{Name: "product", Parameters: []string{"list"}, Body: nil, Closure: env})
	env.DefineFunction("sorted_desc", &vm.FunctionValue{Name: "sorted_desc", Parameters: []string{"list"}, Body: nil, Closure: env})
	env.DefineFunction("zip_with", &vm.FunctionValue{Name: "zip_with", Parameters: []string{"list", "other"}, Body: nil, Closure: env})
}
