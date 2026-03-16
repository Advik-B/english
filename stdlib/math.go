package stdlib

import (
	"github.com/Advik-B/english/astvm"
	"math"
	"math/rand"
)

func evalMath(name string, args []vm.Value) (vm.Value, error) {
	switch name {
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
	}
	return nil, vm.NewRuntimeError("unknown math function: " + name)
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
