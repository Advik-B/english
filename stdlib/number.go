package stdlib

import (
	"math"

	vm "github.com/Advik-B/english/astvm"
)

func evalNumber(name string, args []vm.Value) (vm.Value, error) {
	switch name {
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
	}
	return nil, vm.NewRuntimeError("unknown number function: " + name)
}
