package vm

import (
	"math"
	"sort"
	"strings"
)

// RegisterStdlib registers all standard library functions
func RegisterStdlib(env *Environment) {
	registerMathFunctions(env)
	registerStringFunctions(env)
	registerListFunctions(env)
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
}

// registerStringFunctions registers all string functions
func registerStringFunctions(env *Environment) {
	env.DefineFunction("uppercase", &FunctionValue{Name: "uppercase", Parameters: []string{"text"}, Body: nil, Closure: env})
	env.DefineFunction("lowercase", &FunctionValue{Name: "lowercase", Parameters: []string{"text"}, Body: nil, Closure: env})
	env.DefineFunction("split", &FunctionValue{Name: "split", Parameters: []string{"text", "separator"}, Body: nil, Closure: env})
	env.DefineFunction("join", &FunctionValue{Name: "join", Parameters: []string{"list", "separator"}, Body: nil, Closure: env})
	env.DefineFunction("trim", &FunctionValue{Name: "trim", Parameters: []string{"text"}, Body: nil, Closure: env})
	env.DefineFunction("replace", &FunctionValue{Name: "replace", Parameters: []string{"text", "old", "new"}, Body: nil, Closure: env})
	env.DefineFunction("contains", &FunctionValue{Name: "contains", Parameters: []string{"text", "substring"}, Body: nil, Closure: env})
}

// registerListFunctions registers all list functions
func registerListFunctions(env *Environment) {
	env.DefineFunction("append", &FunctionValue{Name: "append", Parameters: []string{"list", "item"}, Body: nil, Closure: env})
	env.DefineFunction("remove", &FunctionValue{Name: "remove", Parameters: []string{"list", "index"}, Body: nil, Closure: env})
	env.DefineFunction("insert", &FunctionValue{Name: "insert", Parameters: []string{"list", "index", "item"}, Body: nil, Closure: env})
	env.DefineFunction("sort", &FunctionValue{Name: "sort", Parameters: []string{"list"}, Body: nil, Closure: env})
	env.DefineFunction("reverse", &FunctionValue{Name: "reverse", Parameters: []string{"list"}, Body: nil, Closure: env})
}

// evalBuiltinFunction evaluates a built-in stdlib function
func evalBuiltinFunction(name string, args []Value) (Value, error) {
	// Math functions
	switch name {
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

	// String functions
	case "uppercase":
		return strings.ToUpper(ToString(args[0])), nil
	case "lowercase":
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
		new := ToString(args[2])
		return strings.ReplaceAll(text, old, new), nil
	case "contains":
		text := ToString(args[0])
		substr := ToString(args[1])
		return strings.Contains(text, substr), nil

	// List functions
	case "append":
		list, ok := args[0].([]interface{})
		if !ok {
			return nil, NewRuntimeError("append expects a list as first argument")
		}
		result := make([]interface{}, len(list)+1)
		copy(result, list)
		result[len(list)] = args[1]
		return result, nil
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
	}

	return nil, NewRuntimeError("unknown built-in function: " + name)
}
