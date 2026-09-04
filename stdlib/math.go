package stdlib

import (
	"math"
	"math/rand"

	vm "github.com/Advik-B/english/astvm"
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
		// A value that is not a number is a mistake, not a not-a-number:
		// answering true for it meant is_nan("hello") reported true, and
		// answering false in is_infinite meant the mistake vanished entirely.
		x, err := requireNumber("is_nan", args[0])
		if err != nil {
			return nil, err
		}
		return math.IsNaN(x), nil
	case "is_infinite":
		x, err := requireNumber("is_infinite", args[0])
		if err != nil {
			return nil, err
		}
		return math.IsInf(x, 0), nil
	}
	return nil, vm.NewRuntimeError("unknown math function: " + name)
}
