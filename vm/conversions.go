package vm

import (
	"english/vm/types"
	"fmt"
	"strconv"
	"strings"
)

// ToString converts any Value to its textual representation.
// This is used for display (Print/Write) and explicit "cast to text".
// It is NOT called automatically during arithmetic or comparisons.
func ToString(v Value) string {
	switch val := v.(type) {
	case float64:
		if val == float64(int64(val)) {
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
			parts[i] = ToString(elem)
		}
		return "[" + strings.Join(parts, " ") + "]"
	case *ArrayValue:
		parts := make([]string, len(val.Elements))
		for i, elem := range val.Elements {
			parts[i] = ToString(elem)
		}
		return "[" + strings.Join(parts, ", ") + "]"
	case *LookupTableValue:
		if len(val.KeyOrder) == 0 {
			return "{}"
		}
		parts := make([]string, 0, len(val.KeyOrder))
		for _, k := range val.KeyOrder {
			origKey, _, ok := types.DeserializeKey(k)
			keyStr := k
			if ok {
				keyStr = ToString(origKey)
			}
			parts = append(parts, fmt.Sprintf("%s: %s", keyStr, ToString(val.Entries[k])))
		}
		return "{" + strings.Join(parts, ", ") + "}"
	case *FunctionValue:
		return fmt.Sprintf("<function %s>", val.Name)
	case *StructInstance:
		return fmt.Sprintf("<%s instance>", val.Definition.Name)
	case *types.ErrorValue:
		return fmt.Sprintf("<error: %s>", val.Message)
	case *ReferenceValue:
		return fmt.Sprintf("<ref: %s>", val.Name)
	case nil:
		return "nothing"
	default:
		return fmt.Sprintf("%v", v)
	}
}

// ToNumber attempts to convert a Value to float64.
// Only float64 and integer types are accepted — no implicit string→number coercion.
// Callers that need explicit string→number conversion must use "cast to number".
func ToNumber(v Value) (float64, error) {
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
		return 0, fmt.Errorf(
			"TypeError: expected number, got %s\n  Hint: use 'cast to number' to convert explicitly",
			typeKindName(inferTypeKind(v)),
		)
	}
}

// ToBool converts a Value to bool for use in conditions.
// With strict typing, ONLY actual boolean values are accepted.
// Truthy/falsy coercion of numbers and strings is not permitted.
func ToBool(v Value) (bool, error) {
	switch val := v.(type) {
	case bool:
		return val, nil
	case nil:
		return false, nil // nothing is always falsy, useful for nil checks
	default:
		return false, fmt.Errorf(
			"TypeError: conditions must be boolean, got %s\n  Hint: use a comparison (e.g. 'x is greater than 0') or a boolean variable",
			typeKindName(inferTypeKind(val)),
		)
	}
}
