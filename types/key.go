package types

import "fmt"

// SerializeKey converts a hashable value into a string map key for use inside
// a LookupTableValue.  The type prefix prevents collisions between e.g. the
// number 5 and the text "5".
//
// Valid key types are: float64 (number), string (text), bool (boolean).
// Any other type returns a non-nil error.
func SerializeKey(v any) (string, error) {
	switch val := v.(type) {
	case string:
		return "s:" + val, nil
	case bool:
		if val {
			return "b:true", nil
		}
		return "b:false", nil
	}
	// Any numeric representation is a valid key. Only float64 was accepted
	// before, so a value that had been through a cast to a sized numeric type
	// could not be used as a key at all.
	if f, ok := numeric(v); ok {
		return fmt.Sprintf("n:%v", f), nil
	}
	return "", fmt.Errorf(
		"a lookup table key must be a number, text or a boolean; got %s", NameOf(v))
}

// DeserializeKey recovers the original value from a serialised key string.
// It mirrors SerializeKey and returns (value, type, ok).
func DeserializeKey(s string) (any, TypeKind, bool) {
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
