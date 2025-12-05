package interpreter

import (
	"fmt"
	"strconv"
)

// Value represents a runtime value in the interpreter
type Value interface{}

// Type assertions for different value types
func toNumber(v Value) (float64, error) {
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

func toString(v Value) string {
	switch val := v.(type) {
	case float64:
		if val == float64(int64(val)) {
			return strconv.FormatInt(int64(val), 10)
		}
		return strconv.FormatFloat(val, 'f', -1, 64)
	case string:
		return val
	case []interface{}:
		var parts []string
		for _, elem := range val {
			parts = append(parts, toString(elem))
		}
		return "[" + fmt.Sprintf("%v", parts)[1:len(fmt.Sprintf("%v", parts))-1] + "]"
	case *FunctionValue:
		return fmt.Sprintf("<function %s>", val.Name)
	case nil:
		return "nil"
	default:
		return fmt.Sprintf("%v", val)
	}
}

func toBool(v Value) bool {
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

// FunctionValue represents a user-defined function
type FunctionValue struct {
	Name       string
	Parameters []string
	Body       []Statement
	Closure    *Environment
}

// BuiltinFunction represents a built-in function
type BuiltinFunction func(...Value) (Value, error)

// Compare values based on an operator
func compare(op string, left, right Value) (bool, error) {
	switch op {
	case "is equal to":
		return equals(left, right), nil
	case "is less than":
		l, err := toNumber(left)
		if err != nil {
			return false, err
		}
		r, err := toNumber(right)
		if err != nil {
			return false, err
		}
		return l < r, nil
	case "is greater than":
		l, err := toNumber(left)
		if err != nil {
			return false, err
		}
		r, err := toNumber(right)
		if err != nil {
			return false, err
		}
		return l > r, nil
	case "is less than or equal to":
		l, err := toNumber(left)
		if err != nil {
			return false, err
		}
		r, err := toNumber(right)
		if err != nil {
			return false, err
		}
		return l <= r, nil
	case "is greater than or equal to":
		l, err := toNumber(left)
		if err != nil {
			return false, err
		}
		r, err := toNumber(right)
		if err != nil {
			return false, err
		}
		return l >= r, nil
	case "is not equal to":
		return !equals(left, right), nil
	default:
		return false, fmt.Errorf("unknown comparison operator: %s", op)
	}
}

func equals(left, right Value) bool {
	switch l := left.(type) {
	case float64:
		switch r := right.(type) {
		case float64:
			return l == r
		default:
			return false
		}
	case string:
		switch r := right.(type) {
		case string:
			return l == r
		default:
			return false
		}
	case nil:
		return right == nil
	default:
		return false
	}
}

// Arithmetic operations
func add(left, right Value) (Value, error) {
	switch l := left.(type) {
	case float64:
		r, err := toNumber(right)
		if err != nil {
			return nil, err
		}
		return l + r, nil
	case string:
		return l + toString(right), nil
	case []interface{}:
		switch r := right.(type) {
		case []interface{}:
			return append(l, r...), nil
		default:
			return append(l, r), nil
		}
	default:
		return nil, fmt.Errorf("cannot add %T and %T", left, right)
	}
}

func subtract(left, right Value) (Value, error) {
	l, err := toNumber(left)
	if err != nil {
		return nil, fmt.Errorf("cannot subtract: left operand is not a number (got %T)\n  Hint: Subtraction only works with numbers", left)
	}
	r, err := toNumber(right)
	if err != nil {
		return nil, fmt.Errorf("cannot subtract: right operand is not a number (got %T)\n  Hint: Subtraction only works with numbers", right)
	}
	return l - r, nil
}

func multiply(left, right Value) (Value, error) {
	switch l := left.(type) {
	case float64:
		r, err := toNumber(right)
		if err != nil {
			return nil, err
		}
		return l * r, nil
	case string:
		r, err := toNumber(right)
		if err != nil {
			return nil, err
		}
		result := ""
		for i := 0; i < int(r); i++ {
			result += l
		}
		return result, nil
	default:
		return nil, fmt.Errorf("cannot multiply %T and %T", left, right)
	}
}

func divide(left, right Value) (Value, error) {
	l, err := toNumber(left)
	if err != nil {
		return nil, fmt.Errorf("cannot divide: left operand is not a number (got %T)\n  Hint: Division only works with numbers", left)
	}
	r, err := toNumber(right)
	if err != nil {
		return nil, fmt.Errorf("cannot divide: right operand is not a number (got %T)\n  Hint: Division only works with numbers", right)
	}
	if r == 0 {
		return nil, fmt.Errorf("division by zero\n  Hint: You cannot divide by zero - check your expression")
	}
	return l / r, nil
}
