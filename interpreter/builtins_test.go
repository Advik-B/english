package interpreter

import (
	"testing"
)

func TestToNumber(t *testing.T) {
	tests := []struct {
		input    Value
		expected float64
		hasError bool
	}{
		{float64(5), 5, false},
		{"10", 10, false},
		{"3.14", 3.14, false},
		{"invalid", 0, true},
		{[]interface{}{}, 0, true},
	}

	for _, test := range tests {
		result, err := toNumber(test.input)
		if test.hasError {
			if err == nil {
				t.Errorf("Expected error for %v, got nil", test.input)
			}
		} else {
			if err != nil {
				t.Errorf("Unexpected error for %v: %v", test.input, err)
			}
			if result != test.expected {
				t.Errorf("toNumber(%v) = %v, want %v", test.input, result, test.expected)
			}
		}
	}
}

func TestToString(t *testing.T) {
	tests := []struct {
		input    Value
		expected string
	}{
		{float64(5), "5"},
		{float64(3.14), "3.14"},
		{"hello", "hello"},
		{nil, "nil"},
		{[]interface{}{float64(1), float64(2), float64(3)}, "[1 2 3]"},
	}

	for _, test := range tests {
		result := toString(test.input)
		if result != test.expected {
			t.Errorf("toString(%v) = %q, want %q", test.input, result, test.expected)
		}
	}
}

func TestToBool(t *testing.T) {
	tests := []struct {
		input    Value
		expected bool
	}{
		{true, true},
		{false, false},
		{float64(1), true},
		{float64(0), false},
		{"hello", true},
		{"", false},
		{[]interface{}{1}, true},
		{[]interface{}{}, false},
		{nil, false},
	}

	for _, test := range tests {
		result := toBool(test.input)
		if result != test.expected {
			t.Errorf("toBool(%v) = %v, want %v", test.input, result, test.expected)
		}
	}
}

func TestCompare(t *testing.T) {
	tests := []struct {
		op       string
		left     Value
		right    Value
		expected bool
	}{
		{"is equal to", float64(5), float64(5), true},
		{"is equal to", float64(5), float64(10), false},
		{"is equal to", "hello", "hello", true},
		{"is equal to", "hello", "world", false},
		{"is less than", float64(5), float64(10), true},
		{"is less than", float64(10), float64(5), false},
		{"is greater than", float64(10), float64(5), true},
		{"is greater than", float64(5), float64(10), false},
		{"is less than or equal to", float64(5), float64(5), true},
		{"is less than or equal to", float64(5), float64(10), true},
		{"is greater than or equal to", float64(5), float64(5), true},
		{"is greater than or equal to", float64(10), float64(5), true},
		{"is not equal to", float64(5), float64(10), true},
		{"is not equal to", float64(5), float64(5), false},
	}

	for _, test := range tests {
		result, err := compare(test.op, test.left, test.right)
		if err != nil {
			t.Errorf("compare(%q, %v, %v) error: %v", test.op, test.left, test.right, err)
			continue
		}
		if result != test.expected {
			t.Errorf("compare(%q, %v, %v) = %v, want %v", test.op, test.left, test.right, result, test.expected)
		}
	}
}

func TestEquals(t *testing.T) {
	tests := []struct {
		left     Value
		right    Value
		expected bool
	}{
		{float64(5), float64(5), true},
		{float64(5), float64(10), false},
		{"hello", "hello", true},
		{"hello", "world", false},
		{nil, nil, true},
		{float64(5), "5", false}, // Different types
	}

	for _, test := range tests {
		result := equals(test.left, test.right)
		if result != test.expected {
			t.Errorf("equals(%v, %v) = %v, want %v", test.left, test.right, result, test.expected)
		}
	}
}

func TestAdd(t *testing.T) {
	tests := []struct {
		left     Value
		right    Value
		expected Value
	}{
		{float64(5), float64(3), float64(8)},
		{"Hello", " World", "Hello World"},
		{[]interface{}{1}, []interface{}{2}, []interface{}{1, 2}},
	}

	for _, test := range tests {
		result, err := add(test.left, test.right)
		if err != nil {
			t.Errorf("add(%v, %v) error: %v", test.left, test.right, err)
			continue
		}
		switch expected := test.expected.(type) {
		case float64:
			if result != expected {
				t.Errorf("add(%v, %v) = %v, want %v", test.left, test.right, result, expected)
			}
		case string:
			if result != expected {
				t.Errorf("add(%v, %v) = %v, want %v", test.left, test.right, result, expected)
			}
		}
	}
}

func TestSubtract(t *testing.T) {
	result, err := subtract(float64(10), float64(3))
	if err != nil {
		t.Fatalf("subtract error: %v", err)
	}
	if result != float64(7) {
		t.Errorf("subtract(10, 3) = %v, want 7", result)
	}

	// Test error case
	_, err = subtract("hello", float64(3))
	if err == nil {
		t.Error("Expected error for non-numeric subtraction")
	}
}

func TestMultiply(t *testing.T) {
	tests := []struct {
		left     Value
		right    Value
		expected Value
	}{
		{float64(5), float64(3), float64(15)},
		{"ab", float64(3), "ababab"},
	}

	for _, test := range tests {
		result, err := multiply(test.left, test.right)
		if err != nil {
			t.Errorf("multiply(%v, %v) error: %v", test.left, test.right, err)
			continue
		}
		if result != test.expected {
			t.Errorf("multiply(%v, %v) = %v, want %v", test.left, test.right, result, test.expected)
		}
	}
}

func TestDivide(t *testing.T) {
	result, err := divide(float64(10), float64(2))
	if err != nil {
		t.Fatalf("divide error: %v", err)
	}
	if result != float64(5) {
		t.Errorf("divide(10, 2) = %v, want 5", result)
	}

	// Test division by zero
	_, err = divide(float64(10), float64(0))
	if err == nil {
		t.Error("Expected error for division by zero")
	}

	// Test non-numeric division
	_, err = divide("hello", float64(2))
	if err == nil {
		t.Error("Expected error for non-numeric division")
	}
}

func TestFunctionValue(t *testing.T) {
	fn := &FunctionValue{
		Name:       "test",
		Parameters: []string{"a", "b"},
		Body:       []Statement{},
		Closure:    NewEnvironment(),
	}

	result := toString(fn)
	if result != "<function test>" {
		t.Errorf("toString(function) = %q, want '<function test>'", result)
	}
}
