package vm

import (
	"fmt"
	"strconv"
	"strings"
)

// ToNumber converts a value to a number
func ToNumber(v Value) (float64, error) {
	switch val := v.(type) {
	case float64:
		return val, nil
	case string:
		n, err := strconv.ParseFloat(val, 64)
		return n, err
	default:
		return 0, fmt.Errorf("cannot convert %T to number", v)
	}
}

// ToString converts a value to a string
func ToString(v Value) string {
	switch val := v.(type) {
	case float64:
		if val == float64(int64(val)) {
			return strconv.FormatInt(int64(val), 10)
		}
		return strconv.FormatFloat(val, 'f', -1, 64)
	case string:
		return val
	case bool:
		if val {
			return "true"
		}
		return "false"
	case []interface{}:
		var parts []string
		for _, elem := range val {
			parts = append(parts, ToString(elem))
		}
		return "[" + strings.Join(parts, " ") + "]"
	case *FunctionValue:
		return fmt.Sprintf("<function %s>", val.Name)
	case nil:
		return "nil"
	default:
		return fmt.Sprintf("%v", val)
	}
}

// ToBool converts a value to a boolean
func ToBool(v Value) bool {
	switch val := v.(type) {
	case bool:
		return val
	case float64:
		return val != 0
	case string:
		return val != ""
	case []interface{}:
		return len(val) > 0
	case nil:
		return false
	default:
		return true
	}
}
