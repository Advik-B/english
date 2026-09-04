package types

import (
	"fmt"
	"strconv"
	"strings"
)

// Cast performs an explicit type conversion requested by a "cast to" expression.
// It is the ONLY place where type conversion is allowed in the language.
// Implicit conversions are always a TypeError.
func Cast(v interface{}, target TypeKind) (interface{}, error) {
	// Unwrap TypedValue if present
	if tv, ok := v.(*TypedValue); ok {
		v = tv.Value
	}

	switch target {
	case TypeI32:
		switch val := v.(type) {
		case float64:
			return int32(val), nil
		case int64:
			return int32(val), nil
		case uint32:
			return int32(val), nil
		case uint64:
			return int32(val), nil
		case string:
			var i int32
			if _, err := fmt.Sscanf(val, "%d", &i); err != nil {
				return nil, fmt.Errorf("TypeError: cannot cast text %q to number", val)
			}
			return i, nil
		default:
			return nil, fmt.Errorf("TypeError: cannot cast %s to number", Name(Infer(v)))
		}

	case TypeI64:
		switch val := v.(type) {
		case float64:
			return int64(val), nil
		case int32:
			return int64(val), nil
		case uint32:
			return int64(val), nil
		case uint64:
			return int64(val), nil
		case string:
			var i int64
			if _, err := fmt.Sscanf(val, "%d", &i); err != nil {
				return nil, fmt.Errorf("TypeError: cannot cast text %q to number", val)
			}
			return i, nil
		default:
			return nil, fmt.Errorf("TypeError: cannot cast %s to number", Name(Infer(v)))
		}

	case TypeU32:
		switch val := v.(type) {
		case float64:
			if val < 0 {
				return nil, fmt.Errorf("TypeError: cannot cast negative number to unsigned integer")
			}
			return uint32(val), nil
		case int32:
			if val < 0 {
				return nil, fmt.Errorf("TypeError: cannot cast negative number to unsigned integer")
			}
			return uint32(val), nil
		case int64:
			if val < 0 {
				return nil, fmt.Errorf("TypeError: cannot cast negative number to unsigned integer")
			}
			return uint32(val), nil
		case string:
			var i uint32
			if _, err := fmt.Sscanf(val, "%d", &i); err != nil {
				return nil, fmt.Errorf("TypeError: cannot cast text %q to number", val)
			}
			return i, nil
		default:
			return nil, fmt.Errorf("TypeError: cannot cast %s to number", Name(Infer(v)))
		}

	case TypeF64:
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
		case bool:
			if val {
				return float64(1), nil
			}
			return float64(0), nil
		case string:
			f, err := strconv.ParseFloat(val, 64)
			if err != nil {
				return nil, fmt.Errorf("TypeError: cannot cast text %q to number", val)
			}
			return f, nil
		default:
			return nil, fmt.Errorf("TypeError: cannot cast %s to number", Name(Infer(v)))
		}

	case TypeString:
		return basicString(v), nil

	case TypeBool:
		switch val := v.(type) {
		case bool:
			return val, nil
		case float64:
			return val != 0, nil
		case int32:
			return val != 0, nil
		case int64:
			return val != 0, nil
		case string:
			normalized := strings.ToLower(val)
			switch normalized {
			case "true", "1", "yes":
				return true, nil
			case "false", "0", "no":
				return false, nil
			}
			return nil, fmt.Errorf("TypeError: cannot cast text %q to boolean", val)
		case nil:
			return false, nil
		default:
			return nil, fmt.Errorf("TypeError: cannot cast %s to boolean", Name(Infer(v)))
		}

	default:
		return nil, fmt.Errorf("TypeError: unsupported cast target type '%s'", Name(target))
	}
}

// basicString converts a primitive value to its text representation.
// This is intentionally limited to types known by vm/types/ so that the cast
// package remains free of vm dependencies.  The vm package's full ToString
// handles complex types (arrays, lookup tables, struct instances, etc.).
func basicString(v interface{}) string {
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
	case nil:
		return "nothing"
	default:
		return fmt.Sprintf("%v", v)
	}
}
