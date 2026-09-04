package types

import (
	"fmt"
	"strconv"
	"strings"
)

// Cast performs an explicit conversion requested by a "cast to" expression.
//
// It is the only place a conversion happens: nothing is converted implicitly,
// which is why the expression exists.
//
// There are three targets, because there are three convertible types. It used
// to accept nine, one for each sized numeric kind, and was incomplete in both
// directions: it had no case for u64 or f32 at all, and every numeric case
// omitted its own type, so casting an i32 to i32 failed. The error for an
// unsupported target read "cannot cast number to number", because all nine
// kinds are called "number".
func Cast(v interface{}, target TypeKind) (interface{}, error) {
	if tv, ok := v.(*TypedValue); ok {
		v = tv.Value
	}

	switch target {
	case TypeF64:
		return castToNumber(v)
	case TypeString:
		return castToText(v)
	case TypeBool:
		return castToBoolean(v)
	}
	return nil, fmt.Errorf("cannot cast to %s\n  Hint: a cast can produce a number, text or a boolean",
		Name(target))
}

func castToNumber(v interface{}) (interface{}, error) {
	if f, ok := numeric(v); ok {
		return f, nil
	}
	switch val := v.(type) {
	case bool:
		if val {
			return float64(1), nil
		}
		return float64(0), nil
	case string:
		f, err := strconv.ParseFloat(strings.TrimSpace(val), 64)
		if err != nil {
			return nil, fmt.Errorf("cannot cast the text %q to a number", val)
		}
		return f, nil
	}
	return nil, fmt.Errorf("cannot cast %s to a number", NameOf(v))
}

func castToText(v interface{}) (interface{}, error) {
	// Rendering a composite value needs the full renderer, which lives in the
	// runtime package and cannot be imported here without a cycle. The engines
	// route "cast to text" through that renderer directly; this handles the
	// primitives for any caller that reaches Cast with one.
	switch val := v.(type) {
	case nil:
		return "nothing", nil
	case string:
		return val, nil
	case bool:
		if val {
			return "true", nil
		}
		return "false", nil
	}
	if f, ok := numeric(v); ok {
		if f == float64(int64(f)) {
			return strconv.FormatInt(int64(f), 10), nil
		}
		return strconv.FormatFloat(f, 'f', -1, 64), nil
	}
	return fmt.Sprintf("%v", v), nil
}

func castToBoolean(v interface{}) (interface{}, error) {
	switch val := v.(type) {
	case bool:
		return val, nil
	case nil:
		return false, nil
	case string:
		switch strings.ToLower(strings.TrimSpace(val)) {
		case "true", "yes", "1":
			return true, nil
		case "false", "no", "0":
			return false, nil
		}
		return nil, fmt.Errorf("cannot cast the text %q to a boolean\n  Hint: write true, false, yes or no", val)
	}
	if f, ok := numeric(v); ok {
		return f != 0, nil
	}
	return nil, fmt.Errorf("cannot cast %s to a boolean", NameOf(v))
}

// numeric unwraps any numeric representation to a float64.
//
// Only float64 can be produced now, but a value decoded from bytecode or held
// in a TypedValue may still be one of the others, so this stays total.
func numeric(v interface{}) (float64, bool) {
	switch val := v.(type) {
	case float64:
		return val, true
	case float32:
		return float64(val), true
	case int:
		return float64(val), true
	case int32:
		return float64(val), true
	case int64:
		return float64(val), true
	case uint32:
		return float64(val), true
	case uint64:
		return float64(val), true
	case *TypedValue:
		return numeric(val.Value)
	}
	return 0, false
}
