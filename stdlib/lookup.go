package stdlib

import (
	"english/astvm"
	"english/astvm/types"
	"fmt"
)

func evalLookup(name string, args []vm.Value) (vm.Value, error) {
	switch name {
	case "keys":
		if len(args) != 1 {
			return nil, fmt.Errorf("keys() expects 1 argument")
		}
		lt, ok := args[0].(*types.LookupTableValue)
		if !ok {
			return nil, fmt.Errorf("TypeError: keys() expects a lookup table, got %s", kindName(args[0]))
		}
		result := make([]interface{}, 0, len(lt.KeyOrder))
		for _, k := range lt.KeyOrder {
			orig, _, ok := types.DeserializeKey(k)
			if ok {
				result = append(result, orig)
			} else {
				result = append(result, k)
			}
		}
		return result, nil
	case "values":
		if len(args) != 1 {
			return nil, fmt.Errorf("values() expects 1 argument")
		}
		lt, ok := args[0].(*types.LookupTableValue)
		if !ok {
			return nil, fmt.Errorf("TypeError: values() expects a lookup table, got %s", kindName(args[0]))
		}
		result := make([]interface{}, 0, len(lt.KeyOrder))
		for _, k := range lt.KeyOrder {
			result = append(result, lt.Entries[k])
		}
		return result, nil
	case "table_remove":
		if len(args) != 2 {
			return nil, fmt.Errorf("table_remove() expects 2 arguments")
		}
		lt, ok := args[0].(*types.LookupTableValue)
		if !ok {
			return nil, fmt.Errorf("TypeError: table_remove() expects a lookup table, got %s", kindName(args[0]))
		}
		serialKey, err := types.SerializeKey(args[1])
		if err != nil {
			return nil, err
		}
		newTable := types.NewLookupTable()
		for _, k := range lt.KeyOrder {
			if k != serialKey {
				newTable.Set(k, lt.Entries[k])
			}
		}
		return newTable, nil
	case "table_has":
		if len(args) != 2 {
			return nil, fmt.Errorf("table_has() expects 2 arguments")
		}
		lt, ok := args[0].(*types.LookupTableValue)
		if !ok {
			return nil, fmt.Errorf("TypeError: table_has() expects a lookup table, got %s", kindName(args[0]))
		}
		serialKey, err := types.SerializeKey(args[1])
		if err != nil {
			return nil, err
		}
		_, exists := lt.Entries[serialKey]
		return exists, nil
	case "merge":
		lt, err := requireLookupTable("merge", args[0])
		if err != nil {
			return nil, err
		}
		other, err := requireLookupTable("merge", args[1])
		if err != nil {
			return nil, err
		}
		newTable := types.NewLookupTable()
		for _, k := range lt.KeyOrder {
			newTable.Set(k, lt.Entries[k])
		}
		for _, k := range other.KeyOrder {
			newTable.Set(k, other.Entries[k])
		}
		return newTable, nil
	case "get_or_default":
		lt, err := requireLookupTable("get_or_default", args[0])
		if err != nil {
			return nil, err
		}
		serialKey, err := types.SerializeKey(args[1])
		if err != nil {
			return nil, err
		}
		if v, ok := lt.Entries[serialKey]; ok {
			return v, nil
		}
		return args[2], nil
	}
	return nil, vm.NewRuntimeError("unknown lookup table function: " + name)
}

func registerLookupTableFunctions(env *vm.Environment) {
	env.DefineFunction("keys", &vm.FunctionValue{Name: "keys", Parameters: []string{"table"}, Body: nil, Closure: env})
	env.DefineFunction("values", &vm.FunctionValue{Name: "values", Parameters: []string{"table"}, Body: nil, Closure: env})
	env.DefineFunction("table_remove", &vm.FunctionValue{Name: "table_remove", Parameters: []string{"table", "key"}, Body: nil, Closure: env})
	env.DefineFunction("table_has", &vm.FunctionValue{Name: "table_has", Parameters: []string{"table", "key"}, Body: nil, Closure: env})
	env.DefineFunction("merge", &vm.FunctionValue{Name: "merge", Parameters: []string{"table", "other"}, Body: nil, Closure: env})
	env.DefineFunction("get_or_default", &vm.FunctionValue{Name: "get_or_default", Parameters: []string{"table", "key", "default"}, Body: nil, Closure: env})
}
