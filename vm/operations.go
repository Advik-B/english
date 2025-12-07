package vm

import "fmt"

// Compare compares two values based on an operator
func Compare(op string, left, right Value) (bool, error) {
	switch op {
	case "is equal to":
		return Equals(left, right), nil
	case "is less than":
		l, err := ToNumber(left)
		if err != nil {
			return false, err
		}
		r, err := ToNumber(right)
		if err != nil {
			return false, err
		}
		return l < r, nil
	case "is greater than":
		l, err := ToNumber(left)
		if err != nil {
			return false, err
		}
		r, err := ToNumber(right)
		if err != nil {
			return false, err
		}
		return l > r, nil
	case "is less than or equal to":
		l, err := ToNumber(left)
		if err != nil {
			return false, err
		}
		r, err := ToNumber(right)
		if err != nil {
			return false, err
		}
		return l <= r, nil
	case "is greater than or equal to":
		l, err := ToNumber(left)
		if err != nil {
			return false, err
		}
		r, err := ToNumber(right)
		if err != nil {
			return false, err
		}
		return l >= r, nil
	case "is not equal to":
		return !Equals(left, right), nil
	default:
		return false, fmt.Errorf("unknown comparison operator: %s", op)
	}
}

// Equals checks if two values are equal
func Equals(left, right Value) bool {
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
	case bool:
		switch r := right.(type) {
		case bool:
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

// Add adds two values
func Add(left, right Value) (Value, error) {
	switch l := left.(type) {
	case float64:
		// When left is a number, right must also be a number
		switch right.(type) {
		case float64:
			r, _ := ToNumber(right)
			return l + r, nil
		case string:
			return nil, fmt.Errorf("mismatched types %s and %s for operation \"add\"", getTypeName(left), getTypeName(right))
		default:
			r, err := ToNumber(right)
			if err != nil {
				return nil, fmt.Errorf("mismatched types %s and %s for operation \"add\"", getTypeName(left), getTypeName(right))
			}
			return l + r, nil
		}
	case string:
		// When left is a string, concatenate with string representation of right
		return l + ToString(right), nil
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

// Subtract subtracts two values
func Subtract(left, right Value) (Value, error) {
	l, err := ToNumber(left)
	if err != nil {
		return nil, fmt.Errorf("cannot subtract: left operand is not a number (got %T)\n  Hint: Subtraction only works with numbers", left)
	}
	r, err := ToNumber(right)
	if err != nil {
		return nil, fmt.Errorf("cannot subtract: right operand is not a number (got %T)\n  Hint: Subtraction only works with numbers", right)
	}
	return l - r, nil
}

// Multiply multiplies two values
func Multiply(left, right Value) (Value, error) {
	switch l := left.(type) {
	case float64:
		r, err := ToNumber(right)
		if err != nil {
			return nil, err
		}
		return l * r, nil
	case string:
		r, err := ToNumber(right)
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

// Divide divides two values
func Divide(left, right Value) (Value, error) {
	l, err := ToNumber(left)
	if err != nil {
		return nil, fmt.Errorf("cannot divide: left operand is not a number (got %T)\n  Hint: Division only works with numbers", left)
	}
	r, err := ToNumber(right)
	if err != nil {
		return nil, fmt.Errorf("cannot divide: right operand is not a number (got %T)\n  Hint: Division only works with numbers", right)
	}
	if r == 0 {
		return nil, fmt.Errorf("division by zero\n  Hint: You cannot divide by zero - check your expression")
	}
	return l / r, nil
}

// Modulo calculates the remainder of two values
func Modulo(left, right Value) (Value, error) {
	l, err := ToNumber(left)
	if err != nil {
		return nil, fmt.Errorf("cannot get remainder: left operand is not a number (got %T)\n  Hint: Remainder only works with numbers", left)
	}
	r, err := ToNumber(right)
	if err != nil {
		return nil, fmt.Errorf("cannot get remainder: right operand is not a number (got %T)\n  Hint: Remainder only works with numbers", right)
	}
	if r == 0 {
		return nil, fmt.Errorf("division by zero\n  Hint: You cannot get remainder when dividing by zero")
	}
	return float64(int64(l) % int64(r)), nil
}
