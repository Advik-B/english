package stdlib

import (
	"github.com/Advik-B/english/astvm"
	"github.com/Advik-B/english/astvm/types"
	"fmt"
)

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
