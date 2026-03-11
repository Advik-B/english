package types

import "fmt"

// SerializeKey converts a hashable value into a string map key for use inside
// a LookupTableValue.  The type prefix prevents collisions between e.g. the
// number 5 and the text "5".
//
// Valid key types are: float64 (number), string (text), bool (boolean).
// Any other type returns a non-nil error.
func SerializeKey(v interface{}) (string, error) {
	switch val := v.(type) {
	case float64:
		return fmt.Sprintf("n:%v", val), nil
	case string:
		return "s:" + val, nil
	case bool:
		if val {
			return "b:true", nil
		}
		return "b:false", nil
	default:
		return "", fmt.Errorf(
			"TypeError: lookup table keys must be number, text, or boolean; got %T", v,
		)
	}
}

// DeserializeKey recovers the original value from a serialised key string.
// It mirrors SerializeKey and returns (value, type, ok).
func DeserializeKey(s string) (interface{}, TypeKind, bool) {
	if len(s) < 2 {
		return nil, TypeUnknown, false
	}
	prefix, payload := s[:2], s[2:]
	switch prefix {
	case "n:":
		var f float64
		_, err := fmt.Sscanf(payload, "%g", &f)
		if err != nil {
			return nil, TypeUnknown, false
		}
		return f, TypeF64, true
	case "s:":
		return payload, TypeString, true
	case "b:":
		return payload == "true", TypeBool, true
	default:
		return nil, TypeUnknown, false
	}
}
