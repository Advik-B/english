package ivm

import (
	"github.com/Advik-B/english/astvm/types"
	"fmt"
	"math"
	"strconv"
	"strings"
)

// ─── Helpers ──────────────────────────────────────────────────────────────────

func doBinaryOp(op BinOp, left, right interface{}) (interface{}, error) {
	switch op {
	case BinAdd:
		return ivmAdd(left, right)
	case BinSub:
		return requireNumberBinary(left, right, "-", func(a, b float64) interface{} { return a - b })
	case BinMul:
		return requireNumberBinary(left, right, "*", func(a, b float64) interface{} { return a * b })
	case BinDiv:
		l, err := ivmToFloat(left, "/")
		if err != nil {
			return nil, err
		}
		r, err := ivmToFloat(right, "/")
		if err != nil {
			return nil, err
		}
		if r == 0 {
			return nil, fmt.Errorf("division by zero")
		}
		return l / r, nil
	case BinMod:
		l, err := ivmToFloat(left, "remainder")
		if err != nil {
			return nil, err
		}
		r, err := ivmToFloat(right, "remainder")
		if err != nil {
			return nil, err
		}
		if r == 0 {
			return nil, fmt.Errorf("division by zero in remainder")
		}
		return float64(int64(l) % int64(r)), nil
	case BinEq:
		return ivmStrictEquals(left, right)
	case BinNeq:
		eq, err := ivmStrictEquals(left, right)
		if err != nil {
			return nil, err
		}
		return !eq, nil
	case BinLt:
		return ivmOrderCompare(left, right, func(a, b float64) bool { return a < b })
	case BinLte:
		return ivmOrderCompare(left, right, func(a, b float64) bool { return a <= b })
	case BinGt:
		return ivmOrderCompare(left, right, func(a, b float64) bool { return a > b })
	case BinGte:
		return ivmOrderCompare(left, right, func(a, b float64) bool { return a >= b })
	}
	return nil, fmt.Errorf("unknown binary op: %d", op)
}

func doUnaryOp(op UnaryOp, val interface{}) (interface{}, error) {
	switch op {
	case UnaryNeg:
		n, err := ivmToFloat(val, "-")
		if err != nil {
			return nil, err
		}
		return -n, nil
	case UnaryNot:
		b, err := ivmToBool(val)
		if err != nil {
			return nil, err
		}
		return !b, nil
	}
	return nil, fmt.Errorf("unknown unary op: %d", op)
}

func ivmAdd(left, right interface{}) (interface{}, error) {
	switch l := left.(type) {
	case float64:
		r, ok := right.(float64)
		if !ok {
			return nil, fmt.Errorf("TypeError: '+' requires matching types")
		}
		return l + r, nil
	case string:
		r, ok := right.(string)
		if !ok {
			return nil, fmt.Errorf("TypeError: '+' requires matching types")
		}
		return l + r, nil
	case *types.ArrayValue:
		r, ok := right.(*types.ArrayValue)
		if !ok {
			return nil, fmt.Errorf("TypeError: cannot concatenate array with non-array")
		}
		if l.ElementType != r.ElementType {
			return nil, fmt.Errorf("TypeError: cannot concatenate arrays of different element types")
		}
		combined := make([]interface{}, len(l.Elements)+len(r.Elements))
		copy(combined, l.Elements)
		copy(combined[len(l.Elements):], r.Elements)
		return &types.ArrayValue{ElementType: l.ElementType, Elements: combined}, nil
	default:
		return nil, fmt.Errorf("TypeError: '+' is not defined for %s", ivmGetTypeName(left))
	}
}

func requireNumberBinary(left, right interface{}, op string, fn func(float64, float64) interface{}) (interface{}, error) {
	l, err := ivmToFloat(left, op)
	if err != nil {
		return nil, err
	}
	r, err := ivmToFloat(right, op)
	if err != nil {
		return nil, err
	}
	return fn(l, r), nil
}

func ivmToFloat(v interface{}, op string) (float64, error) {
	switch val := v.(type) {
	case float64:
		return val, nil
	case int32:
		return float64(val), nil
	case int64:
		return float64(val), nil
	case uint32:
		return float64(val), nil
	case uint64:
		return float64(val), nil
	case float32:
		return float64(val), nil
	default:
		return 0, fmt.Errorf("TypeError: '%s' requires number, got %s", op, ivmGetTypeName(v))
	}
}

func ivmToBool(v interface{}) (bool, error) {
	switch val := v.(type) {
	case bool:
		return val, nil
	case nil:
		return false, nil
	default:
		return false, fmt.Errorf("TypeError: conditions must be boolean, got %s", ivmGetTypeName(val))
	}
}

func ivmStrictEquals(left, right interface{}) (bool, error) {
	if left == nil && right == nil {
		return true, nil
	}
	if left == nil || right == nil {
		return false, nil
	}
	lk := types.Canonical(types.Infer(left))
	rk := types.Canonical(types.Infer(right))
	if lk != rk {
		return false, nil
	}
	switch l := left.(type) {
	case float64:
		if r, ok := right.(float64); ok {
			return l == r, nil
		}
	case string:
		if r, ok := right.(string); ok {
			return l == r, nil
		}
	case bool:
		if r, ok := right.(bool); ok {
			return l == r, nil
		}
	}
	return false, nil
}

func ivmOrderCompare(left, right interface{}, pred func(float64, float64) bool) (bool, error) {
	l, err := ivmToFloat(left, "comparison")
	if err != nil {
		return false, err
	}
	r, err := ivmToFloat(right, "comparison")
	if err != nil {
		return false, err
	}
	return pred(l, r), nil
}

func doIndexGet(container, index interface{}) (interface{}, error) {
	switch c := container.(type) {
	case []interface{}:
		idx, err := ivmToFloat(index, "index")
		if err != nil {
			return nil, err
		}
		i := int(idx)
		if i < 0 || i >= len(c) {
			return nil, fmt.Errorf("index %d out of range for list of length %d", i, len(c))
		}
		return c[i], nil
	case *types.ArrayValue:
		idx, err := ivmToFloat(index, "index")
		if err != nil {
			return nil, err
		}
		i := int(idx)
		if i < 0 || i >= len(c.Elements) {
			return nil, fmt.Errorf("index %d out of range for array of length %d", i, len(c.Elements))
		}
		return c.Elements[i], nil
	case *types.RangeValue:
		idx, err := ivmToFloat(index, "index")
		if err != nil {
			return nil, err
		}
		i := int(idx)
		val, ok := c.Get(i)
		if !ok {
			return nil, fmt.Errorf("index %d out of range for range of length %d", i, c.Length())
		}
		return val, nil
	case string:
		idx, err := ivmToFloat(index, "index")
		if err != nil {
			return nil, err
		}
		i := int(idx)
		runes := []rune(c)
		if i < 0 || i >= len(runes) {
			return nil, fmt.Errorf("index %d out of range for string of length %d", i, len(runes))
		}
		return string(runes[i]), nil
	case *types.LookupTableValue:
		// Integer indexing into a lookup table yields the key at that position
		// (used by the for-each loop when iterating over a lookup table).
		idx, err := ivmToFloat(index, "index")
		if err != nil {
			return nil, err
		}
		return lookupTableGetByIndex(c, int(idx))
	default:
		return nil, fmt.Errorf("cannot index into %s", ivmGetTypeName(container))
	}
}

// lookupTableGetByIndex returns the key at position i (0-based) in a lookup table.
// Used by the for-each loop when iterating over a lookup table (yields keys in insertion order).
func lookupTableGetByIndex(lt *types.LookupTableValue, i int) (interface{}, error) {
	if i < 0 || i >= len(lt.KeyOrder) {
		return nil, fmt.Errorf("index %d out of range for lookup table of length %d", i, len(lt.KeyOrder))
	}
	serialKey := lt.KeyOrder[i]
	origKey, _, ok := types.DeserializeKey(serialKey)
	if !ok {
		origKey = serialKey
	}
	return origKey, nil
}

func doIndexSet(container, index, value interface{}) error {
	switch c := container.(type) {
	case []interface{}:
		idx, err := ivmToFloat(index, "index")
		if err != nil {
			return err
		}
		i := int(idx)
		if i < 0 || i >= len(c) {
			return fmt.Errorf("index %d out of range for list of length %d", i, len(c))
		}
		c[i] = value
		return nil
	case *types.ArrayValue:
		idx, err := ivmToFloat(index, "index")
		if err != nil {
			return err
		}
		i := int(idx)
		if i < 0 || i >= len(c.Elements) {
			return fmt.Errorf("index %d out of range for array of length %d", i, len(c.Elements))
		}
		c.Elements[i] = value
		return nil
	case *types.RangeValue:
		return fmt.Errorf("cannot modify a range")
	default:
		return fmt.Errorf("cannot assign to index of %s", ivmGetTypeName(container))
	}
}

func doLength(val interface{}) (float64, error) {
	switch v := val.(type) {
	case []interface{}:
		return float64(len(v)), nil
	case *types.ArrayValue:
		return float64(len(v.Elements)), nil
	case *types.RangeValue:
		return float64(v.Length()), nil
	case string:
		return float64(len([]rune(v))), nil
	case *types.LookupTableValue:
		return float64(len(v.KeyOrder)), nil
	default:
		return 0, fmt.Errorf("cannot get length of %s", ivmGetTypeName(val))
	}
}

func doLookupGet(table, key interface{}) (interface{}, error) {
	lt, ok := table.(*types.LookupTableValue)
	if !ok {
		return nil, fmt.Errorf("LOOKUP_GET: not a lookup table")
	}
	k, err := types.SerializeKey(key)
	if err != nil {
		return nil, err
	}
	val, exists := lt.Entries[k]
	if !exists {
		return nil, nil
	}
	return val, nil
}

func ivmGetTypeName(v interface{}) string {
	switch val := v.(type) {
	case float64:
		return "f64"
	case int32:
		return "i32"
	case int64:
		return "i64"
	case uint32:
		return "u32"
	case uint64:
		return "u64"
	case float32:
		return "f32"
	case string:
		return "text"
	case bool:
		return "boolean"
	case []interface{}:
		return "list"
	case *types.ArrayValue:
		elemTypeInfo := &types.TypeInfo{Kind: val.ElementType}
		return fmt.Sprintf("array of %s", elemTypeInfo.String())
	case *types.RangeValue:
		return "range"
	case *types.LookupTableValue:
		return "lookup table"
	case *types.ErrorValue:
		return "error"
	case *StructInstance:
		return val.DefName
	case *ReferenceValue:
		return "reference"
	case *FuncChunk:
		return "function"
	case nil:
		return "nothing"
	default:
		return fmt.Sprintf("%T", v)
	}
}

func inferKindName(v interface{}) string {
	return types.Name(types.Infer(v))
}

func ivmToString(v interface{}) string {
	switch val := v.(type) {
	case float64:
		if val == float64(int64(val)) && !math.IsInf(val, 0) {
			return strconv.FormatInt(int64(val), 10)
		}
		return strconv.FormatFloat(val, 'f', -1, 64)
	case int32:
		return strconv.FormatInt(int64(val), 10)
	case int64:
		return strconv.FormatInt(val, 10)
	case uint32:
		return strconv.FormatUint(uint64(val), 10)
	case uint64:
		return strconv.FormatUint(val, 10)
	case float32:
		return strconv.FormatFloat(float64(val), 'f', -1, 32)
	case string:
		return val
	case bool:
		if val {
			return "true"
		}
		return "false"
	case []interface{}:
		parts := make([]string, len(val))
		for i, elem := range val {
			parts[i] = ivmToString(elem)
		}
		return "[" + strings.Join(parts, " ") + "]"
	case *types.ArrayValue:
		parts := make([]string, len(val.Elements))
		for i, elem := range val.Elements {
			parts[i] = ivmToString(elem)
		}
		return "[" + strings.Join(parts, ", ") + "]"
	case *types.LookupTableValue:
		if len(val.KeyOrder) == 0 {
			return "{}"
		}
		parts := make([]string, 0, len(val.KeyOrder))
		for _, k := range val.KeyOrder {
			origKey, _, ok := types.DeserializeKey(k)
			keyStr := k
			if ok {
				keyStr = ivmToString(origKey)
			}
			parts = append(parts, fmt.Sprintf("%s: %s", keyStr, ivmToString(val.Entries[k])))
		}
		return "{" + strings.Join(parts, ", ") + "}"
	case *StructInstance:
		return fmt.Sprintf("<%s instance>", val.DefName)
	case *types.ErrorValue:
		return fmt.Sprintf("<error: %s>", val.Message)
	case *ReferenceValue:
		return fmt.Sprintf("<ref: %s>", val.Name)
	case *FuncChunk:
		return fmt.Sprintf("<function %s>", val.Name)
	case nil:
		return "nothing"
	default:
		return fmt.Sprintf("%v", v)
	}
}

func deepCopyValue(val interface{}) interface{} {
	switch v := val.(type) {
	case []interface{}:
		copied := make([]interface{}, len(v))
		for i, elem := range v {
			copied[i] = deepCopyValue(elem)
		}
		return copied
	case *types.ArrayValue:
		elems := make([]interface{}, len(v.Elements))
		for i, elem := range v.Elements {
			elems[i] = deepCopyValue(elem)
		}
		return &types.ArrayValue{ElementType: v.ElementType, Elements: elems}
	case *StructInstance:
		copied := &StructInstance{
			DefName: v.DefName,
			DefRef:  v.DefRef,
			Fields:  make(map[string]interface{}),
		}
		for k, fv := range v.Fields {
			copied.Fields[k] = deepCopyValue(fv)
		}
		return copied
	default:
		return val
	}
}

func typeDefault(typeName string) interface{} {
	switch typeName {
	case "number", "f64", "f32", "i32", "i64", "u32", "u64":
		return float64(0)
	case "text":
		return ""
	case "boolean":
		return false
	case "list":
		return []interface{}{}
	default:
		return nil
	}
}
