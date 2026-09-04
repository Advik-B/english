package stdlib

import (
	"fmt"

	vm "github.com/Advik-B/english/astvm"
	"github.com/Advik-B/english/types"
)

func requireText(fn string, arg vm.Value) (string, error) {
	s, ok := arg.(string)
	if !ok {
		return "", fmt.Errorf("TypeError: %s expects text, got %s", fn, types.NameOf(arg))
	}
	return s, nil
}

func requireNumber(fn string, arg vm.Value) (float64, error) {
	n, err := vm.ToNumber(arg)
	if err != nil {
		return 0, fmt.Errorf("TypeError: %s expects number, got %s", fn, types.NameOf(arg))
	}
	return n, nil
}

func requireList(fn string, arg vm.Value) ([]interface{}, error) {
	lst, ok := arg.([]interface{})
	if !ok {
		return nil, fmt.Errorf("TypeError: %s expects list, got %s", fn, types.NameOf(arg))
	}
	return lst, nil
}

func requireLookupTable(fn string, arg vm.Value) (*types.LookupTableValue, error) {
	lt, ok := arg.(*types.LookupTableValue)
	if !ok {
		return nil, fmt.Errorf("TypeError: %s expects lookup table, got %s", fn, types.NameOf(arg))
	}
	return lt, nil
}
