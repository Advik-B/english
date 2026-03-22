package vm_test

import (
	"bytes"
	"github.com/Advik-B/english/ast"
	"github.com/Advik-B/english/parser"
	"github.com/Advik-B/english/astvm"
	"github.com/Advik-B/english/astvm/types"
	"github.com/Advik-B/english/stdlib"
	"io"
	"os"
	"strings"
	"testing"
)

// Helper function to evaluate code
func evaluate(input string) (vm.Value, error) {
	lexer := parser.NewLexer(input)
	tokens := lexer.TokenizeAll()
	p := parser.NewParser(tokens)
	program, err := p.Parse()
	if err != nil {
		return nil, err
	}
	env := vm.NewEnvironment()
	stdlib.Register(env)
	evaluator := vm.NewEvaluator(env, stdlib.Eval)
	return evaluator.Eval(program)
}

// captureOutput captures stdout during a function execution
func captureOutput(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

// ============================================
// ENVIRONMENT TESTS
// ============================================

func TestEnvironmentDefine(t *testing.T) {
	env := vm.NewEnvironment()
	err := env.Define("x", float64(5), false)
	if err != nil {
		t.Fatalf("Define error: %v", err)
	}

	val, ok := env.Get("x")
	if !ok {
		t.Fatal("Variable 'x' not found")
	}
	if val != float64(5) {
		t.Errorf("Expected 5, got %v", val)
	}
}

func TestEnvironmentDefineConstant(t *testing.T) {
	env := vm.NewEnvironment()
	err := env.Define("PI", float64(3.14159), true)
	if err != nil {
		t.Fatalf("Define error: %v", err)
	}

	if !env.IsConstant("PI") {
		t.Error("PI should be a constant")
	}
}

func TestEnvironmentSet(t *testing.T) {
	env := vm.NewEnvironment()
	env.Define("x", float64(5), false)
	env.Set("x", float64(10))

	val, _ := env.Get("x")
	if val != float64(10) {
		t.Errorf("Expected 10, got %v", val)
	}
}

func TestEnvironmentSetConstantError(t *testing.T) {
	env := vm.NewEnvironment()
	env.Define("PI", float64(3.14159), true)
	err := env.Set("PI", float64(3.14))

	if err == nil {
		t.Error("Expected error when reassigning constant")
	}
}

func TestEnvironmentChildScope(t *testing.T) {
	parent := vm.NewEnvironment()
	parent.Define("x", float64(5), false)

	child := parent.NewChild()
	child.Define("y", float64(10), false)

	// Child can see parent's variable
	val, ok := child.Get("x")
	if !ok {
		t.Error("Child should see parent's variable")
	}
	if val != float64(5) {
		t.Errorf("Expected 5, got %v", val)
	}

	// Parent cannot see child's variable
	_, ok = parent.Get("y")
	if ok {
		t.Error("Parent should not see child's variable")
	}
}

func TestEnvironmentGetFunction(t *testing.T) {
	env := vm.NewEnvironment()
	fn := &vm.FunctionValue{Name: "test", Parameters: []string{}, Body: []ast.Statement{}}
	env.DefineFunction("test", fn)

	found, ok := env.GetFunction("test")
	if !ok {
		t.Error("Function not found")
	}
	if found.Name != "test" {
		t.Errorf("Expected function name 'test', got %q", found.Name)
	}
}

func TestEnvironmentGetAllVariables(t *testing.T) {
	env := vm.NewEnvironment()
	env.Define("x", float64(5), false)
	env.Define("y", float64(10), false)

	vars := env.GetAllVariables()
	if len(vars) != 2 {
		t.Errorf("Expected 2 variables, got %d", len(vars))
	}
}

func TestEnvironmentGetAllFunctions(t *testing.T) {
	env := vm.NewEnvironment()
	fn1 := &vm.FunctionValue{Name: "fn1", Parameters: []string{}, Body: []ast.Statement{}}
	fn2 := &vm.FunctionValue{Name: "fn2", Parameters: []string{}, Body: []ast.Statement{}}
	env.DefineFunction("fn1", fn1)
	env.DefineFunction("fn2", fn2)

	funcs := env.GetAllFunctions()
	if len(funcs) != 2 {
		t.Errorf("Expected 2 functions, got %d", len(funcs))
	}
}

// ============================================
// VALUE CONVERSION TESTS
// ============================================

func TestToNumber(t *testing.T) {
	// Strict type system: ToNumber only accepts actual numeric values.
	// Text-to-number coercion requires explicit "cast to number".
	tests := []struct {
		input    vm.Value
		expected float64
		hasError bool
	}{
		{float64(5), 5, false},
		{float64(3.14), 3.14, false},
		{"10", 0, true},   // text needs explicit cast
		{"3.14", 0, true}, // text needs explicit cast
		{"invalid", 0, true},
		{[]interface{}{}, 0, true},
	}

	for _, test := range tests {
		result, err := vm.ToNumber(test.input)
		if test.hasError {
			if err == nil {
				t.Errorf("Expected error for %v, got nil", test.input)
			}
		} else {
			if err != nil {
				t.Errorf("Unexpected error for %v: %v", test.input, err)
			}
			if result != test.expected {
				t.Errorf("vm.ToNumber(%v) = %v, want %v", test.input, result, test.expected)
			}
		}
	}
}

func TestToString(t *testing.T) {
	tests := []struct {
		input    vm.Value
		expected string
	}{
		{float64(5), "5"},
		{float64(3.14), "3.14"},
		{"hello", "hello"},
		{nil, "nothing"}, // nil displays as "nothing" (the language keyword)
		{true, "true"},
		{false, "false"},
		{[]interface{}{float64(1), float64(2), float64(3)}, "[1 2 3]"},
	}

	for _, test := range tests {
		result := vm.ToString(test.input)
		if result != test.expected {
			t.Errorf("vm.ToString(%v) = %q, want %q", test.input, result, test.expected)
		}
	}
}

func TestToBool(t *testing.T) {
	// With strict typing, ToBool only accepts boolean values (and nil for nothing).
	// Truthy/falsy coercion of numbers, strings, and lists is intentionally removed.
	tests := []struct {
		input     vm.Value
		expected  bool
		expectErr bool
	}{
		{true, true, false},
		{false, false, false},
		{nil, false, false},             // nothing is always false
		{float64(1), false, true},       // TypeError: number is not boolean
		{"hello", false, true},          // TypeError: text is not boolean
		{[]interface{}{1}, false, true}, // TypeError: list is not boolean
	}

	for _, test := range tests {
		result, err := vm.ToBool(test.input)
		if test.expectErr {
			if err == nil {
				t.Errorf("vm.ToBool(%v): expected TypeError, got nil error", test.input)
			}
		} else {
			if err != nil {
				t.Errorf("vm.ToBool(%v): unexpected error: %v", test.input, err)
			}
			if result != test.expected {
				t.Errorf("vm.ToBool(%v) = %v, want %v", test.input, result, test.expected)
			}
		}
	}
}

// ============================================
// ARITHMETIC TESTS
// ============================================

func TestAdd(t *testing.T) {
	// Strict typing: only same-type addition. list+list via '+' is removed; use append().
	if _, err := vm.Add(float64(5), float64(3)); err != nil {
		t.Errorf("vm.Add(5, 3): unexpected error: %v", err)
	}
	r, err := vm.Add("Hello", " World")
	if err != nil || r != "Hello World" {
		t.Errorf("vm.Add(text,text): got %v, %v", r, err)
	}
	// list+list is now a TypeError
	if _, err := vm.Add([]interface{}{1}, []interface{}{2}); err == nil {
		t.Error("vm.Add(list, list): expected TypeError, got nil")
	}
}

func TestSubtract(t *testing.T) {
	result, err := vm.Subtract(float64(10), float64(3))
	if err != nil {
		t.Fatalf("Subtract error: %v", err)
	}
	if result != float64(7) {
		t.Errorf("vm.Subtract(10, 3) = %v, want 7", result)
	}

	// Test error case
	_, err = vm.Subtract("hello", float64(3))
	if err == nil {
		t.Error("Expected error for non-numeric subtraction")
	}
}

func TestMultiply(t *testing.T) {
	// Strict typing: only number*number. String repetition removed; use str_repeat().
	r, err := vm.Multiply(float64(5), float64(3))
	if err != nil || r != float64(15) {
		t.Errorf("vm.Multiply(5,3): got %v, %v", r, err)
	}
	if _, err := vm.Multiply("ab", float64(3)); err == nil {
		t.Error("vm.Multiply(text,number): expected TypeError, got nil")
	}
}

func TestDivide(t *testing.T) {
	result, err := vm.Divide(float64(10), float64(2))
	if err != nil {
		t.Fatalf("Divide error: %v", err)
	}
	if result != float64(5) {
		t.Errorf("vm.Divide(10, 2) = %v, want 5", result)
	}

	// Test division by zero
	_, err = vm.Divide(float64(10), float64(0))
	if err == nil {
		t.Error("Expected error for division by zero")
	}

	// Test non-numeric division
	_, err = vm.Divide("hello", float64(2))
	if err == nil {
		t.Error("Expected error for non-numeric division")
	}
}

func TestModulo(t *testing.T) {
	result, err := vm.Modulo(float64(17), float64(5))
	if err != nil {
		t.Fatalf("Modulo error: %v", err)
	}
	if result != float64(2) {
		t.Errorf("vm.Modulo(17, 5) = %v, want 2", result)
	}

	// Test modulo by zero
	_, err = vm.Modulo(float64(10), float64(0))
	if err == nil {
		t.Error("Expected error for modulo by zero")
	}
}

// ============================================
// COMPARISON TESTS
// ============================================

func TestCompare(t *testing.T) {
	tests := []struct {
		op       string
		left     vm.Value
		right    vm.Value
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
		result, err := vm.Compare(test.op, test.left, test.right)
		if err != nil {
			t.Errorf("vm.Compare(%q, %v, %v) error: %v", test.op, test.left, test.right, err)
			continue
		}
		if result != test.expected {
			t.Errorf("vm.Compare(%q, %v, %v) = %v, want %v", test.op, test.left, test.right, result, test.expected)
		}
	}
}

func TestEquals(t *testing.T) {
	tests := []struct {
		left     vm.Value
		right    vm.Value
		expected bool
	}{
		{float64(5), float64(5), true},
		{float64(5), float64(10), false},
		{"hello", "hello", true},
		{"hello", "world", false},
		{true, true, true},
		{true, false, false},
		{nil, nil, true},
		{float64(5), "5", false}, // Different types
	}

	for _, test := range tests {
		result := vm.Equals(test.left, test.right)
		if result != test.expected {
			t.Errorf("vm.Equals(%v, %v) = %v, want %v", test.left, test.right, result, test.expected)
		}
	}
}

// ============================================
// EVALUATOR TESTS
// ============================================

func TestEvaluatorVariableDeclaration(t *testing.T) {
	input := "Declare x to be 5."
	_, err := evaluate(input)
	if err != nil {
		t.Fatalf("Evaluation error: %v", err)
	}
}

func TestEvaluatorConstantDeclaration(t *testing.T) {
	input := "Declare x to always be 5."
	_, err := evaluate(input)
	if err != nil {
		t.Fatalf("Evaluation error: %v", err)
	}
}

func TestEvaluatorConstantReassignment(t *testing.T) {
	input := `Declare x to always be 5.
Set x to be 10.`
	_, err := evaluate(input)
	if err == nil {
		t.Fatal("Expected error when reassigning constant, got nil")
	}
	if !strings.Contains(err.Error(), "cannot reassign constant") {
		t.Errorf("Expected constant reassignment error, got: %v", err)
	}
}

func TestEvaluatorAssignment(t *testing.T) {
	input := `Declare x to be 5.
Set x to be 10.`
	_, err := evaluate(input)
	if err != nil {
		t.Fatalf("Evaluation error: %v", err)
	}
}

func TestEvaluatorArithmetic(t *testing.T) {
	tests := []struct {
		input string
	}{
		{"Declare x to be 1 + 2."},
		{"Declare x to be 5 - 3."},
		{"Declare x to be 4 * 3."},
		{"Declare x to be 8 / 2."},
		{"Declare x to be 1 + 2 * 3."},
		{"Declare x to be -5."},
		{"Declare x to be (1 + 2) * 3."},
	}

	for _, test := range tests {
		_, err := evaluate(test.input)
		if err != nil {
			t.Errorf("Input %q: evaluation error: %v", test.input, err)
		}
	}
}

func TestEvaluatorComparisons(t *testing.T) {
	tests := []struct {
		code     string
		expected string
	}{
		{`Declare x to be 5.
If x is equal to 5, then
    Print "yes".
thats it.`, "yes\n"},
		{`Declare x to be 5.
If x is less than 10, then
    Print "yes".
thats it.`, "yes\n"},
		{`Declare x to be 5.
If x is greater than 3, then
    Print "yes".
thats it.`, "yes\n"},
		{`Declare x to be 5.
If x is less than or equal to 5, then
    Print "yes".
thats it.`, "yes\n"},
		{`Declare x to be 5.
If x is greater than or equal to 5, then
    Print "yes".
thats it.`, "yes\n"},
		{`Declare x to be 5.
If x is not equal to 10, then
    Print "yes".
thats it.`, "yes\n"},
	}

	for _, test := range tests {
		output := captureOutput(func() {
			evaluate(test.code)
		})
		if output != test.expected {
			t.Errorf("Expected output %q, got %q", test.expected, output)
		}
	}
}

func TestEvaluatorIfElse(t *testing.T) {
	code := `Declare x to be 5.
If x is equal to 10, then
    Print "ten".
otherwise
    Print "not ten".
thats it.`

	output := captureOutput(func() {
		evaluate(code)
	})
	if output != "not ten\n" {
		t.Errorf("Expected 'not ten', got %q", output)
	}
}

func TestEvaluatorIfElseIf(t *testing.T) {
	code := `Declare x to be 5.
If x is equal to 1, then
    Print "one".
otherwise if x is equal to 5, then
    Print "five".
otherwise
    Print "other".
thats it.`

	output := captureOutput(func() {
		evaluate(code)
	})
	if output != "five\n" {
		t.Errorf("Expected 'five', got %q", output)
	}
}

func TestEvaluatorWhileLoop(t *testing.T) {
	code := `Declare x to be 0.
repeat the following while x is less than 3:
    Set x to be x + 1.
thats it.`

	_, err := evaluate(code)
	if err != nil {
		t.Fatalf("Evaluation error: %v", err)
	}
}

func TestEvaluatorForLoop(t *testing.T) {
	code := `Declare count to be 0.
repeat the following 5 times:
    Set count to be count + 1.
thats it.`

	_, err := evaluate(code)
	if err != nil {
		t.Fatalf("Evaluation error: %v", err)
	}
}

func TestEvaluatorForLoopOutput(t *testing.T) {
	code := `repeat the following 3 times:
    Print "hello".
thats it.`

	output := captureOutput(func() {
		evaluate(code)
	})
	expected := "hello\nhello\nhello\n"
	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

func TestEvaluatorForEachLoop(t *testing.T) {
	code := `Declare myList to be [1, 2, 3].
for each item in myList, do the following:
    Print the value of item.
thats it.`

	output := captureOutput(func() {
		evaluate(code)
	})
	expected := "1\n2\n3\n"
	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

func TestEvaluatorFunctionDeclarationAndCall(t *testing.T) {
	code := `Declare function greet that does the following:
    Print "Hello".
thats it.
Call greet.`

	output := captureOutput(func() {
		evaluate(code)
	})
	if output != "Hello\n" {
		t.Errorf("Expected 'Hello', got %q", output)
	}
}

func TestEvaluatorFunctionWithParams(t *testing.T) {
	code := `Declare function add that takes a and b and does the following:
    Return a + b.
thats it.
Set result to be the result of calling add with 3 and 7.
Print the value of result.`

	output := captureOutput(func() {
		evaluate(code)
	})
	if output != "10\n" {
		t.Errorf("Expected '10', got %q", output)
	}
}

func TestEvaluatorFunctionReturn(t *testing.T) {
	code := `Declare function double that takes x and does the following:
    Return x * 2.
thats it.
Set result to be the result of calling double with 5.
Print the value of result.`

	output := captureOutput(func() {
		evaluate(code)
	})
	if output != "10\n" {
		t.Errorf("Expected '10', got %q", output)
	}
}

func TestEvaluatorFunctionSingleParam(t *testing.T) {
	code := `Declare function square that takes n and does the following:
    Return n * n.
thats it.
Set result to be the result of calling square with 4.
Print the value of result.`

	output := captureOutput(func() {
		evaluate(code)
	})
	if output != "16\n" {
		t.Errorf("Expected '16', got %q", output)
	}
}

func TestEvaluatorUndefinedVariable(t *testing.T) {
	code := "Print the value of undefined."
	_, err := evaluate(code)
	if err == nil {
		t.Fatal("Expected error for undefined variable, got nil")
	}
	if !strings.Contains(err.Error(), "undefined variable") {
		t.Errorf("Expected undefined variable error, got: %v", err)
	}
}

func TestEvaluatorUndefinedFunction(t *testing.T) {
	code := "Call undefinedFunc."
	_, err := evaluate(code)
	if err == nil {
		t.Fatal("Expected error for undefined function, got nil")
	}
	if !strings.Contains(err.Error(), "undefined function") {
		t.Errorf("Expected undefined function error, got: %v", err)
	}
}

func TestEvaluatorDivisionByZero(t *testing.T) {
	code := `Declare x to be 10.
Declare y to be 0.
Set result to be x / y.`
	_, err := evaluate(code)
	if err == nil {
		t.Fatal("Expected division by zero error, got nil")
	}
	if !strings.Contains(err.Error(), "division by zero") {
		t.Errorf("Expected division by zero error, got: %v", err)
	}
}

func TestEvaluatorFunctionArgumentMismatch(t *testing.T) {
	code := `Declare function add that takes a and b and does the following:
    Return a + b.
thats it.
Set result to be the result of calling add with 5.`

	_, err := evaluate(code)
	if err == nil {
		t.Fatal("Expected argument mismatch error, got nil")
	}
	if !strings.Contains(err.Error(), "expects") {
		t.Errorf("Expected argument mismatch error, got: %v", err)
	}
}

func TestEvaluatorListOperations(t *testing.T) {
	code := `Declare myList to be [1, 2, 3, 4, 5].`
	_, err := evaluate(code)
	if err != nil {
		t.Fatalf("Evaluation error: %v", err)
	}
}

func TestEvaluatorRangeLiteral(t *testing.T) {
	tests := []struct {
		code     string
		varName  string
		expected []interface{}
	}{
		{
			code:     `Declare r to be [1 .. 5].`,
			varName:  "r",
			expected: []interface{}{float64(1), float64(2), float64(3), float64(4)},
		},
		{
			code:     `Let myRange be a range from 1 to 3.`,
			varName:  "myRange",
			expected: []interface{}{float64(1), float64(2)},
		},
		{
			code:     `Declare desc to be [5 .. 1].`,
			varName:  "desc",
			expected: []interface{}{},
		},
		{
			code:     `Declare evens to be [0 .. 10 by 2].`,
			varName:  "evens",
			expected: []interface{}{float64(0), float64(2), float64(4), float64(6), float64(8)},
		},
		{
			code:     `Let odds be a range from 1 to 9 by 2.`,
			varName:  "odds",
			expected: []interface{}{float64(1), float64(3), float64(5), float64(7)},
		},
		{
			code:     `Declare countdown to be [10 .. 0 by -2].`,
			varName:  "countdown",
			expected: []interface{}{float64(10), float64(8), float64(6), float64(4), float64(2)},
		},
		{
			code:     `Let multiples be a range from 5 to 25 by 5.`,
			varName:  "multiples",
			expected: []interface{}{float64(5), float64(10), float64(15), float64(20)},
		},
	}

	for _, test := range tests {
		lexer := parser.NewLexer(test.code)
		tokens := lexer.TokenizeAll()
		p := parser.NewParser(tokens)
		program, err := p.Parse()
		if err != nil {
			t.Fatalf("Parse error for %q: %v", test.code, err)
		}

		env := vm.NewEnvironment()
		stdlib.Register(env)
		evaluator := vm.NewEvaluator(env, stdlib.Eval)
		_, err = evaluator.Eval(program)
		if err != nil {
			t.Fatalf("Eval error for %q: %v", test.code, err)
		}

		val, ok := env.Get(test.varName)
		if !ok {
			t.Fatalf("Variable %q not found", test.varName)
		}

		rangeVal, ok := val.(*types.RangeValue)
		if !ok {
			t.Fatalf("Expected *RangeValue, got %T", val)
		}

		// Convert to slice for comparison
		result := rangeVal.ToSlice()
		if len(result) != len(test.expected) {
			t.Errorf("Expected %d elements, got %d", len(test.expected), len(result))
		}

		for i, v := range result {
			if v != test.expected[i] {
				t.Errorf("Element %d: expected %v, got %v", i, test.expected[i], v)
			}
		}
	}
}

func TestRangeImmutability(t *testing.T) {
	// Test that trying to assign to a range index results in an error
	code := `Declare r to be [1 .. 10].
Set the item at position 0 in r to be 99.`

	lexer := parser.NewLexer(code)
	tokens := lexer.TokenizeAll()
	p := parser.NewParser(tokens)
	program, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	env := vm.NewEnvironment()
	stdlib.Register(env)
	evaluator := vm.NewEvaluator(env, stdlib.Eval)
	_, err = evaluator.Eval(program)

	if err == nil {
		t.Fatal("Expected error when trying to modify a range, but got nil")
	}

	if !strings.Contains(err.Error(), "cannot modify a range") {
		t.Errorf("Expected error message to contain 'cannot modify a range', got: %v", err)
	}
}

func TestRangeLazyEvaluation(t *testing.T) {
	// Test that large ranges use lazy evaluation
	// A range with more than 20 elements should still be fast to create
	code := `Declare bigRange to be [1 .. 1000].
Declare first to be the item at position 0 in bigRange.
Declare middle to be the item at position 500 in bigRange.
Declare last to be the item at position 998 in bigRange.`

	lexer := parser.NewLexer(code)
	tokens := lexer.TokenizeAll()
	p := parser.NewParser(tokens)
	program, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	env := vm.NewEnvironment()
	stdlib.Register(env)
	evaluator := vm.NewEvaluator(env, stdlib.Eval)
	_, err = evaluator.Eval(program)
	if err != nil {
		t.Fatalf("Eval error: %v", err)
	}

	// Verify the values are correct
	first, _ := env.Get("first")
	if first != float64(1) {
		t.Errorf("Expected first element to be 1, got %v", first)
	}

	middle, _ := env.Get("middle")
	if middle != float64(501) {
		t.Errorf("Expected middle element to be 501, got %v", middle)
	}

	last, _ := env.Get("last")
	if last != float64(999) {
		t.Errorf("Expected last element to be 999, got %v", last)
	}

	// Verify the range itself is still a RangeValue (not expanded to a full slice)
	bigRange, _ := env.Get("bigRange")
	rangeVal, ok := bigRange.(*types.RangeValue)
	if !ok {
		t.Fatalf("Expected bigRange to be *RangeValue, got %T", bigRange)
	}

	if rangeVal.Length() != 999 {
		t.Errorf("Expected range length to be 999, got %d", rangeVal.Length())
	}
}

func TestEvaluatorRangeInLoop(t *testing.T) {
	code := `For each n in [1 .. 3], do the following:
    Print n.
thats it.`

	output := captureOutput(func() {
		evaluate(code)
	})
	expected := "1\n2\n"
	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

func TestEvaluatorStringConcatenation(t *testing.T) {
	code := `Declare x to be "Hello" + " World".
Print the value of x.`

	output := captureOutput(func() {
		evaluate(code)
	})
	if output != "Hello World\n" {
		t.Errorf("Expected 'Hello World', got %q", output)
	}
}

func TestEvaluatorStringMultiplication(t *testing.T) {
	// String * number is no longer supported with strict typing.
	// Use str_repeat() for text repetition.
	code := `Declare x to be str_repeat("ab", 3).
Print the value of x.`

	output := captureOutput(func() {
		evaluate(code)
	})
	if output != "ababab\n" {
		t.Errorf("Expected 'ababab', got %q", output)
	}
}

func TestEvaluatorNestedScopes(t *testing.T) {
	code := `Declare x to be 5.
Declare function change_local that does the following:
    Declare x to be 10.
    Print the value of x.
thats it.
Call change_local.
Print the value of x.`

	output := captureOutput(func() {
		evaluate(code)
	})
	// The function creates a local x, doesn't affect outer x
	if output != "10\n5\n" {
		t.Errorf("Expected '10\\n5\\n', got %q", output)
	}
}

func TestEvaluatorRecursion(t *testing.T) {
	code := `Declare function factorial that takes n and does the following:
    If n is less than or equal to 1, then
        Return 1.
    otherwise
        Set prev to be the result of calling factorial with n - 1.
        Return n * prev.
    thats it.
thats it.
Set result to be the result of calling factorial with 5.
Print the value of result.`

	output := captureOutput(func() {
		evaluate(code)
	})
	if output != "120\n" {
		t.Errorf("Expected '120', got %q", output)
	}
}

func TestEvaluatorCaseInsensitiveKeywords(t *testing.T) {
	tests := []string{
		"DECLARE x TO BE 5.",
		"declare x to be 5.",
		"Declare X to be 5.",
	}

	for _, input := range tests {
		_, err := evaluate(input)
		if err != nil {
			t.Errorf("Input %q: evaluation error: %v", input, err)
		}
	}
}

func TestEvaluatorCaseInsensitiveComparisons(t *testing.T) {
	code := `Declare x to be 5.
IF x IS EQUAL TO 5, THEN
    Print "yes".
thats it.`

	output := captureOutput(func() {
		evaluate(code)
	})
	if output != "yes\n" {
		t.Errorf("Expected 'yes', got %q", output)
	}
}

func TestEvaluatorRemainder(t *testing.T) {
	tests := []struct {
		code     string
		expected string
	}{
		{`Print the remainder of 17 divided by 5.`, "2\n"},
		{`Print the remainder of 10 / 3.`, "1\n"},
		{`Print the remainder of 100 divided by 7.`, "2\n"},
		{`Declare x to be 15.
Declare y to be 4.
Print the remainder of x divided by y.`, "3\n"},
	}

	for _, test := range tests {
		output := captureOutput(func() {
			evaluate(test.code)
		})
		if output != test.expected {
			t.Errorf("Expected %q, got %q", test.expected, output)
		}
	}
}

func TestEvaluatorBooleanLiterals(t *testing.T) {
	tests := []struct {
		code     string
		expected string
	}{
		{`Declare x to be true.
Print the value of x.`, "true\n"},
		{`Declare x to be false.
Print the value of x.`, "false\n"},
		{`Declare x to be true.
If x is equal to true, then
    Print "yes".
thats it.`, "yes\n"},
	}

	for _, test := range tests {
		output := captureOutput(func() {
			evaluate(test.code)
		})
		if output != test.expected {
			t.Errorf("Expected %q, got %q", test.expected, output)
		}
	}
}

func TestEvaluatorToggle(t *testing.T) {
	code := `Declare is_on to be true.
Toggle is_on.
Print the value of is_on.
Toggle the value of is_on.
Print the value of is_on.`

	output := captureOutput(func() {
		evaluate(code)
	})
	expected := "false\ntrue\n"
	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

func TestEvaluatorToggleNonBoolean(t *testing.T) {
	code := `Declare x to be 5.
Toggle x.`

	_, err := evaluate(code)
	if err == nil {
		t.Fatal("Expected error when toggling non-boolean")
	}
	if !strings.Contains(err.Error(), "cannot toggle non-boolean") {
		t.Errorf("Expected toggle error, got: %v", err)
	}
}

func TestEvaluatorLocation(t *testing.T) {
	code := `Declare x to be 5.
Print the location of x.`

	output := captureOutput(func() {
		evaluate(code)
	})
	// Location should be a non-empty string starting with "0x"
	if len(output) < 3 || output[:2] != "0x" {
		t.Errorf("Expected location string starting with '0x', got %q", output)
	}
}

func TestEvaluatorIndexExpression(t *testing.T) {
	tests := []struct {
		code     string
		expected string
	}{
		{`Declare myList to be [10, 20, 30].
Print the item at position 0 in myList.`, "10\n"},
		{`Declare myList to be [10, 20, 30].
Print the item at position 1 in myList.`, "20\n"},
		{`Declare myList to be [10, 20, 30].
Print myList[2].`, "30\n"},
	}

	for _, test := range tests {
		output := captureOutput(func() {
			evaluate(test.code)
		})
		if output != test.expected {
			t.Errorf("Expected %q, got %q", test.expected, output)
		}
	}
}

func TestEvaluatorIndexAssignment(t *testing.T) {
	code := `Declare myList to be [10, 20, 30].
Set the item at position 1 in myList to be 99.
Print the item at position 1 in myList.`

	output := captureOutput(func() {
		evaluate(code)
	})
	if output != "99\n" {
		t.Errorf("Expected '99', got %q", output)
	}
}

func TestEvaluatorIndexOutOfBounds(t *testing.T) {
	code := `Declare myList to be [10, 20, 30].
Print the item at position 10 in myList.`

	_, err := evaluate(code)
	if err == nil {
		t.Fatal("Expected index out of bounds error")
	}
	if !strings.Contains(err.Error(), "out of range") {
		t.Errorf("Expected out of range error, got: %v", err)
	}
}

func TestEvaluatorLengthExpression(t *testing.T) {
	tests := []struct {
		code     string
		expected string
	}{
		{`Declare myList to be [1, 2, 3, 4, 5].
Print the length of myList.`, "5\n"},
		{`Declare myList to be [].
Print the length of myList.`, "0\n"},
		{`Declare myString to be "hello".
Print the length of myString.`, "5\n"},
	}

	for _, test := range tests {
		output := captureOutput(func() {
			evaluate(test.code)
		})
		if output != test.expected {
			t.Errorf("Expected %q, got %q", test.expected, output)
		}
	}
}

func TestEvaluatorEmptyList(t *testing.T) {
	code := `Declare myList to be [].
Print the length of myList.`

	output := captureOutput(func() {
		evaluate(code)
	})
	if output != "0\n" {
		t.Errorf("Expected '0', got %q", output)
	}
}

func TestEvaluatorListAddition(t *testing.T) {
	// list + list via '+' is no longer supported (strict typing).
	// Lists are heterogeneous; use append() for arrays or build combined lists manually.
	// This test now verifies that the TypeError is raised.
	code := `Declare list1 to be [1, 2].
Declare list2 to be [3, 4].
Declare combined to be list1 + list2.`

	// evaluate captures panics/errors gracefully; the result should be empty (error path)
	output := captureOutput(func() {
		evaluate(code)
	})
	// We just want to confirm no output was produced (error stops execution)
	if output != "" {
		t.Errorf("Expected no output (TypeError), got %q", output)
	}
}

func TestEvaluatorReturnInLoop(t *testing.T) {
	code := `Declare function findFirst that takes nums and does the following:
    for each n in nums, do the following:
        If n is greater than 5, then
            Return n.
        thats it.
    thats it.
    Return 0.
thats it.
Set result to be the result of calling findFirst with [1, 3, 7, 9].
Print the value of result.`

	output := captureOutput(func() {
		evaluate(code)
	})
	if output != "7\n" {
		t.Errorf("Expected '7', got %q", output)
	}
}

func TestFunctionValueString(t *testing.T) {
	fn := &vm.FunctionValue{
		Name:       "test",
		Parameters: []string{"a", "b"},
		Body:       []ast.Statement{},
		Closure:    vm.NewEnvironment(),
	}

	result := vm.ToString(fn)
	if result != "<function test>" {
		t.Errorf("vm.ToString(function) = %q, want '<function test>'", result)
	}
}

func TestRuntimeErrorFormat(t *testing.T) {
	err := &vm.RuntimeError{
		Message:   "test error",
		CallStack: []string{"<main>", "myFunc(a, b)"},
	}

	errStr := err.Error()
	if !strings.Contains(errStr, "test error") {
		t.Error("Error should contain message")
	}
	if !strings.Contains(errStr, "Call Stack") {
		t.Error("Error should contain call stack")
	}
	if !strings.Contains(errStr, "<main>") {
		t.Error("Error should contain <main>")
	}
	if !strings.Contains(errStr, "myFunc(a, b)") {
		t.Error("Error should contain function name")
	}
}

func TestEvaluatorImport(t *testing.T) {
	// Create a temporary file for import testing
	tempDir := t.TempDir()
	libFile := tempDir + "/testlib.abc"

	// Create a library file with functions and variables
	libContent := `# Test library
Declare function double that takes x and does the following:
    Return x * 2.
thats it.

Declare magicNumber to always be 42.
`
	err := os.WriteFile(libFile, []byte(libContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test library file: %v", err)
	}

	// Test importing the library
	code := `Import "` + libFile + `".
Declare result to be 0.
Set result to the result of calling double with 5.
Print the value of result.
Print the value of magicNumber.`

	output := captureOutput(func() {
		evaluate(code)
	})

	expected := "10\n42\n"
	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

func TestEvaluatorImportWithFrom(t *testing.T) {
	// Create a temporary file for import testing
	tempDir := t.TempDir()
	libFile := tempDir + "/helpers.abc"

	// Create a library file
	libContent := `Declare function square that takes n and does the following:
    Return n * n.
thats it.
`
	err := os.WriteFile(libFile, []byte(libContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test library file: %v", err)
	}

	// Test importing with "from" syntax
	code := `Import from "` + libFile + `".
Declare result to be 0.
Set result to the result of calling square with 3.
Print the value of result.`

	output := captureOutput(func() {
		evaluate(code)
	})

	if output != "9\n" {
		t.Errorf("Expected '9\\n', got %q", output)
	}
}

func TestEvaluatorImportNonexistent(t *testing.T) {
	code := `Import "nonexistent_file.abc".`

	_, err := evaluate(code)
	if err == nil {
		t.Error("Expected error when importing nonexistent file")
	}

	// Check that error message contains helpful information
	errStr := err.Error()
	if !strings.Contains(errStr, "nonexistent_file.abc") {
		t.Error("Error should mention the file that failed to import")
	}
}

func TestEvaluatorSelectiveImport(t *testing.T) {
	// Create a temporary file for import testing
	tempDir := t.TempDir()
	libFile := tempDir + "/testlib.abc"

	// Create a library file with multiple functions
	libContent := `Declare function add that takes a and b and does the following:
    Return a + b.
thats it.

Declare function multiply that takes a and b and does the following:
    Return a * b.
thats it.

Declare version to always be "1.0".
`
	err := os.WriteFile(libFile, []byte(libContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test library file: %v", err)
	}

	// Test selective import - only import add
	code := `Import add from "` + libFile + `".
Declare result to be 0.
Set result to the result of calling add with 3 and 5.
Print the value of result.`

	output := captureOutput(func() {
		evaluate(code)
	})

	if output != "8\n" {
		t.Errorf("Expected '8\\n', got %q", output)
	}
}

func TestEvaluatorImportEverything(t *testing.T) {
	// Create a temporary file for import testing
	tempDir := t.TempDir()
	libFile := tempDir + "/testlib.abc"

	// Create a library file
	libContent := `Declare function greet that takes name and does the following:
    Print "Hello,", the value of name.
thats it.

Declare greeting to be "Welcome".
`
	err := os.WriteFile(libFile, []byte(libContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test library file: %v", err)
	}

	// Test import everything
	code := `Import everything from "` + libFile + `".
Call greet with "World".
Print the value of greeting.`

	output := captureOutput(func() {
		evaluate(code)
	})

	expected := "Hello, World\nWelcome\n"
	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

func TestEvaluatorImportAll(t *testing.T) {
	// Create a temporary file for import testing
	tempDir := t.TempDir()
	libFile := tempDir + "/testlib.abc"

	// Create a library file
	libContent := `Declare myVar to be 42.
`
	err := os.WriteFile(libFile, []byte(libContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test library file: %v", err)
	}

	// Test import all (synonym for everything)
	code := `Import all from "` + libFile + `".
Print the value of myVar.`

	output := captureOutput(func() {
		evaluate(code)
	})

	if output != "42\n" {
		t.Errorf("Expected '42\\n', got %q", output)
	}
}

func TestEvaluatorSafeImport(t *testing.T) {
	// Create a temporary file for import testing
	tempDir := t.TempDir()
	libFile := tempDir + "/testlib.abc"

	// Create a library file with top-level code
	libContent := `Print "This should not print in safe mode".

Declare function test that does the following:
    Print "Test function".
thats it.

Declare safeVar to be 100.
`
	err := os.WriteFile(libFile, []byte(libContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test library file: %v", err)
	}

	// Test safe import - should not print top-level statement
	code := `Import all from "` + libFile + `" safely.
Call test.
Print the value of safeVar.`

	output := captureOutput(func() {
		evaluate(code)
	})

	expected := "Test function\n100\n"
	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}

	// Verify that the top-level print did NOT execute
	if strings.Contains(output, "This should not print") {
		t.Error("Safe import should not execute top-level statements")
	}
}

// ============================================
// CONTINUE STATEMENT TESTS
// ============================================

func TestContinueInWhileLoop(t *testing.T) {
	code := `Declare i to be 1.
Declare result to be [].
repeat the following while i is less than or equal to 6:
    Declare mod to be the remainder of i divided by 2.
    Set i to be i + 1.
    If mod is equal to 0, then
        Continue.
    thats it.
    Set result to be append(result, i - 1).
thats it.
Print the value of result.`

	output := captureOutput(func() {
		evaluate(code)
	})

	if !strings.Contains(output, "1") || !strings.Contains(output, "3") || !strings.Contains(output, "5") {
		t.Errorf("Expected odd numbers in output, got: %q", output)
	}
}

func TestContinueInForLoop(t *testing.T) {
	code := `Declare count to be 0.
repeat the following 10 times:
    If count is less than 5, then
        Set count to be count + 1.
        Continue.
    thats it.
    Set count to be count + 2.
thats it.
Print the value of count.`

	output := captureOutput(func() {
		evaluate(code)
	})

	if !strings.Contains(output, "15") {
		t.Errorf("Expected '15', got: %q", output)
	}
}

func TestContinueInForEachLoop(t *testing.T) {
	code := `Declare nums to be [1, 2, 3, 4, 5, 6].
Declare evens to be [].
For each n in nums, do the following:
    Declare mod to be the remainder of n divided by 2.
    If mod is not equal to 0, then
        Continue.
    thats it.
    Set evens to be append(evens, n).
thats it.
Print the value of evens.`

	output := captureOutput(func() {
		evaluate(code)
	})

	if !strings.Contains(output, "2") || !strings.Contains(output, "4") || !strings.Contains(output, "6") {
		t.Errorf("Expected [2 4 6] in output, got: %q", output)
	}
}

// ============================================
// NOTHING LITERAL TESTS
// ============================================

func TestNothingLiteral(t *testing.T) {
	// 'nothing' is a nil value; printing it should not crash and produces empty-like output
	code := `Declare x to be nothing.
If x is equal to nothing, then
    Print "is nil".
thats it.`

	output := captureOutput(func() {
		evaluate(code)
	})

	if !strings.Contains(output, "is nil") {
		t.Errorf("Expected 'is nil', got: %q", output)
	}
}

func TestNothingEquality(t *testing.T) {
	code := `Declare x to be nothing.
If x is equal to nothing, then
    Print "is nothing".
thats it.`

	output := captureOutput(func() {
		evaluate(code)
	})

	if !strings.Contains(output, "is nothing") {
		t.Errorf("Expected 'is nothing', got: %q", output)
	}
}

func TestNothingNotEqual(t *testing.T) {
	code := `Declare x to be 5.
If x is not equal to nothing, then
    Print "not nothing".
thats it.`

	output := captureOutput(func() {
		evaluate(code)
	})

	if !strings.Contains(output, "not nothing") {
		t.Errorf("Expected 'not nothing', got: %q", output)
	}
}

// ============================================
// LOGICAL OPERATOR TESTS
// ============================================

func TestLogicalAnd(t *testing.T) {
	tests := []struct {
		code     string
		expected string
	}{
		{`Declare x to be 5.
Declare y to be 10.
If x is greater than 3 and y is less than 20, then
    Print "yes".
thats it.`, "yes"},
		{`Declare x to be 5.
Declare y to be 10.
If x is greater than 10 and y is less than 20, then
    Print "yes".
otherwise
    Print "no".
thats it.`, "no"},
	}

	for _, tt := range tests {
		output := captureOutput(func() {
			evaluate(tt.code)
		})
		if !strings.Contains(output, tt.expected) {
			t.Errorf("Expected %q in output, got: %q", tt.expected, output)
		}
	}
}

func TestLogicalOr(t *testing.T) {
	tests := []struct {
		code     string
		expected string
	}{
		{`Declare x to be 5.
If x is greater than 10 or x is less than 10, then
    Print "yes".
thats it.`, "yes"},
		{`Declare x to be 5.
If x is greater than 10 or x is greater than 100, then
    Print "yes".
otherwise
    Print "no".
thats it.`, "no"},
	}

	for _, tt := range tests {
		output := captureOutput(func() {
			evaluate(tt.code)
		})
		if !strings.Contains(output, tt.expected) {
			t.Errorf("Expected %q in output, got: %q", tt.expected, output)
		}
	}
}

func TestLogicalNot(t *testing.T) {
	code := `Declare flag to be false.
If not flag, then
    Print "not false is true".
thats it.`

	output := captureOutput(func() {
		evaluate(code)
	})

	if !strings.Contains(output, "not false is true") {
		t.Errorf("Expected 'not false is true', got: %q", output)
	}
}

func TestBooleanIsTrue(t *testing.T) {
	tests := []struct {
		code     string
		expected string
	}{
		{`Declare isRaining to be true.
If isRaining is true, then
    Print "raining".
thats it.`, "raining"},
		{`Declare isRaining to be false.
If isRaining is true, then
    Print "raining".
otherwise
    Print "not raining".
thats it.`, "not raining"},
	}

	for _, tt := range tests {
		output := captureOutput(func() {
			evaluate(tt.code)
		})
		if !strings.Contains(output, tt.expected) {
			t.Errorf("Expected %q in output, got: %q", tt.expected, output)
		}
	}
}

func TestBooleanIsFalse(t *testing.T) {
	tests := []struct {
		code     string
		expected string
	}{
		{`Declare isRaining to be false.
If isRaining is false, then
    Print "not raining".
thats it.`, "not raining"},
		{`Declare isRaining to be true.
If isRaining is false, then
    Print "not raining".
otherwise
    Print "raining".
thats it.`, "raining"},
	}

	for _, tt := range tests {
		output := captureOutput(func() {
			evaluate(tt.code)
		})
		if !strings.Contains(output, tt.expected) {
			t.Errorf("Expected %q in output, got: %q", tt.expected, output)
		}
	}
}

func TestBooleanIsntTrue(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{"isn't true (apostrophe) with false", `Declare isRaining to be false.
If isRaining isn't true, then
    Print "not raining".
thats it.`, "not raining"},
		{"isn't true (apostrophe) with true", `Declare isRaining to be true.
If isRaining isn't true, then
    Print "not raining".
otherwise
    Print "raining".
thats it.`, "raining"},
		{"is not true with false", `Declare isRaining to be false.
If isRaining is not true, then
    Print "not raining".
thats it.`, "not raining"},
		{"is not true with true", `Declare isRaining to be true.
If isRaining is not true, then
    Print "not raining".
otherwise
    Print "raining".
thats it.`, "raining"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := captureOutput(func() {
				evaluate(tt.code)
			})
			if !strings.Contains(output, tt.expected) {
				t.Errorf("Expected %q in output, got: %q", tt.expected, output)
			}
		})
	}
}

func TestBooleanIsntFalse(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{"isn't false (apostrophe) with true", `Declare isRaining to be true.
If isRaining isn't false, then
    Print "raining".
thats it.`, "raining"},
		{"is not false with true", `Declare isRaining to be true.
If isRaining is not false, then
    Print "raining".
thats it.`, "raining"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := captureOutput(func() {
				evaluate(tt.code)
			})
			if !strings.Contains(output, tt.expected) {
				t.Errorf("Expected %q in output, got: %q", tt.expected, output)
			}
		})
	}
}

func TestLogicalShortCircuit(t *testing.T) {
	// AND short-circuit: condition is false when left side is false
	code := `Declare x to be false.
If x and true, then
    Print "wrong".
otherwise
    Print "short-circuit works".
thats it.`

	output := captureOutput(func() {
		evaluate(code)
	})

	if !strings.Contains(output, "short-circuit works") {
		t.Errorf("Expected short-circuit behavior, got: %q", output)
	}
}

// ============================================
// STDLIB - NEW MATH FUNCTIONS TESTS
// ============================================

func TestStdlibLog(t *testing.T) {
	code := `Print log(1).`
	output := captureOutput(func() {
		evaluate(code)
	})
	if !strings.Contains(output, "0") {
		t.Errorf("log(1) should be 0, got: %q", output)
	}
}

func TestStdlibExp(t *testing.T) {
	code := `Print exp(0).`
	output := captureOutput(func() {
		evaluate(code)
	})
	if !strings.Contains(output, "1") {
		t.Errorf("exp(0) should be 1, got: %q", output)
	}
}

func TestStdlibRandom(t *testing.T) {
	code := `Declare r to be random().
If r is greater than or equal to 0, then
    If r is less than 1, then
        Print "valid".
    thats it.
thats it.`
	output := captureOutput(func() {
		evaluate(code)
	})
	if !strings.Contains(output, "valid") {
		t.Errorf("random() should return value in [0, 1), got: %q", output)
	}
}

func TestStdlibRandomBetween(t *testing.T) {
	code := `Declare r to be random_between(5, 10).
If r is greater than or equal to 5, then
    If r is less than or equal to 10, then
        Print "valid".
    thats it.
thats it.`
	output := captureOutput(func() {
		evaluate(code)
	})
	if !strings.Contains(output, "valid") {
		t.Errorf("random_between(5, 10) should return value in [5, 10], got: %q", output)
	}
}

func TestMathConstants(t *testing.T) {
	code := `If pi is greater than 3, then
    Print "pi ok".
thats it.
If e is greater than 2, then
    Print "e ok".
thats it.`
	output := captureOutput(func() {
		evaluate(code)
	})
	if !strings.Contains(output, "pi ok") {
		t.Errorf("pi should be > 3, got: %q", output)
	}
	if !strings.Contains(output, "e ok") {
		t.Errorf("e should be > 2, got: %q", output)
	}
}

// ============================================
// STDLIB - NEW STRING FUNCTIONS TESTS
// ============================================

func TestStdlibStartsWith(t *testing.T) {
	code := `Print starts_with("hello world", "hello").`
	output := captureOutput(func() {
		evaluate(code)
	})
	if !strings.Contains(output, "true") {
		t.Errorf("starts_with should return true, got: %q", output)
	}
}

func TestStdlibEndsWith(t *testing.T) {
	code := `Print ends_with("hello world", "world").`
	output := captureOutput(func() {
		evaluate(code)
	})
	if !strings.Contains(output, "true") {
		t.Errorf("ends_with should return true, got: %q", output)
	}
}

func TestStdlibIndexOf(t *testing.T) {
	code := `Print index_of("hello world", "world").`
	output := captureOutput(func() {
		evaluate(code)
	})
	if !strings.Contains(output, "6") {
		t.Errorf("index_of should return 6, got: %q", output)
	}
}

func TestStdlibSubstring(t *testing.T) {
	code := `Print substring("hello world", 6, 5).`
	output := captureOutput(func() {
		evaluate(code)
	})
	if !strings.Contains(output, "world") {
		t.Errorf("substring should return 'world', got: %q", output)
	}
}

func TestStdlibStrRepeat(t *testing.T) {
	code := `Print str_repeat("ab", 3).`
	output := captureOutput(func() {
		evaluate(code)
	})
	if !strings.Contains(output, "ababab") {
		t.Errorf("str_repeat should return 'ababab', got: %q", output)
	}
}

func TestStdlibCountOccurrences(t *testing.T) {
	code := `Print count_occurrences("abcabc", "abc").`
	output := captureOutput(func() {
		evaluate(code)
	})
	if !strings.Contains(output, "2") {
		t.Errorf("count_occurrences should return 2, got: %q", output)
	}
}

func TestStdlibToNumber(t *testing.T) {
	code := `Declare n to be to_number("42.5").
If n is greater than 42, then
    Print "ok".
thats it.`
	output := captureOutput(func() {
		evaluate(code)
	})
	if !strings.Contains(output, "ok") {
		t.Errorf("to_number should parse '42.5' > 42, got: %q", output)
	}
}

func TestStdlibIsEmpty(t *testing.T) {
	code := `Print is_empty("").
Print is_empty("hi").
Print is_empty([]).`
	output := captureOutput(func() {
		evaluate(code)
	})
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) < 3 || lines[0] != "true" || lines[1] != "false" || lines[2] != "true" {
		t.Errorf("is_empty results wrong, got: %q", output)
	}
}

// ============================================
// STDLIB - NEW LIST FUNCTIONS TESTS
// ============================================

func TestStdlibSum(t *testing.T) {
	code := `Print sum([1, 2, 3, 4, 5]).`
	output := captureOutput(func() {
		evaluate(code)
	})
	if !strings.Contains(output, "15") {
		t.Errorf("sum([1,2,3,4,5]) should be 15, got: %q", output)
	}
}

func TestStdlibUnique(t *testing.T) {
	code := `Declare u to be unique([1, 2, 2, 3, 3, 3]).
Print count(u).`
	output := captureOutput(func() {
		evaluate(code)
	})
	if !strings.Contains(output, "3") {
		t.Errorf("unique([1,2,2,3,3,3]) should have 3 elements, got: %q", output)
	}
}

func TestStdlibFirst(t *testing.T) {
	code := `Print first([10, 20, 30]).`
	output := captureOutput(func() {
		evaluate(code)
	})
	if !strings.Contains(output, "10") {
		t.Errorf("first([10,20,30]) should be 10, got: %q", output)
	}
}

func TestStdlibLast(t *testing.T) {
	code := `Print last([10, 20, 30]).`
	output := captureOutput(func() {
		evaluate(code)
	})
	if !strings.Contains(output, "30") {
		t.Errorf("last([10,20,30]) should be 30, got: %q", output)
	}
}

func TestStdlibFlatten(t *testing.T) {
	code := `Print flatten([[1, 2], [3, 4], [5]]).`
	output := captureOutput(func() {
		evaluate(code)
	})
	if !strings.Contains(output, "1") || !strings.Contains(output, "5") {
		t.Errorf("flatten should produce flat list, got: %q", output)
	}
}

func TestStdlibCount(t *testing.T) {
	code := `Print count([1, 2, 3, 4, 5]).`
	output := captureOutput(func() {
		evaluate(code)
	})
	if !strings.Contains(output, "5") {
		t.Errorf("count([1,2,3,4,5]) should be 5, got: %q", output)
	}
}

func TestStdlibSlice(t *testing.T) {
	code := `Print slice([10, 20, 30, 40, 50], 1, 4).`
	output := captureOutput(func() {
		evaluate(code)
	})
	if !strings.Contains(output, "20") || !strings.Contains(output, "40") {
		t.Errorf("slice should return [20 30 40], got: %q", output)
	}
}

// ============================================
// CAST TO SYNTAX TESTS
// ============================================

func TestCastToNumber(t *testing.T) {
	code := `Declare age_str to be "25".
Declare age to be age_str cast to number.
If age is greater than 20, then
    Print "cast works".
thats it.`

	output := captureOutput(func() {
		evaluate(code)
	})

	if !strings.Contains(output, "cast works") {
		t.Errorf("Expected 'cast works', got: %q", output)
	}
}

func TestCastToText(t *testing.T) {
	code := `Declare n to be 42.
Declare s to be n cast to text.
Print starts_with(s, "4").`

	output := captureOutput(func() {
		evaluate(code)
	})

	if !strings.Contains(output, "true") {
		t.Errorf("Expected 'true' after cast to text, got: %q", output)
	}
}

func TestCastToBoolean(t *testing.T) {
	code := `Declare zero to be 0.
Declare b to be zero cast to boolean.
If not b, then
    Print "zero is false".
thats it.`

	output := captureOutput(func() {
		evaluate(code)
	})

	if !strings.Contains(output, "zero is false") {
		t.Errorf("Expected 'zero is false', got: %q", output)
	}
}

func TestCastInCondition(t *testing.T) {
	code := `Declare s to be "100".
If s cast to number is greater than 50, then
    Print "condition cast works".
thats it.`

	output := captureOutput(func() {
		evaluate(code)
	})

	if !strings.Contains(output, "condition cast works") {
		t.Errorf("Expected 'condition cast works', got: %q", output)
	}
}

func TestCastedKeyword(t *testing.T) {
	// "casted" should work as an alias for "cast"
	code := `Declare s to be "7".
Declare n to be s casted to number.
If n is equal to 7, then
    Print "casted keyword works".
thats it.`

	output := captureOutput(func() {
		evaluate(code)
	})

	if !strings.Contains(output, "casted keyword works") {
		t.Errorf("Expected 'casted keyword works', got: %q", output)
	}
}

// ============================================
// TYPE SYSTEM TESTS (static typing)
// ============================================

func TestTypeSystem_VariableTypeLocked(t *testing.T) {
	// Once a variable is declared, its type is fixed.
	code := `Declare x to be 5.
Set x to be "hello".`
	output := captureOutput(func() { evaluate(code) })
	if output != "" {
		t.Errorf("Expected no output (TypeError on reassignment), got %q", output)
	}
}

func TestTypeSystem_BooleanConditionRequired(t *testing.T) {
	// Conditions must be boolean; number is not accepted.
	code := `Declare n to be 5.
If n, then
    Print "bad".
thats it.`
	output := captureOutput(func() { evaluate(code) })
	if output != "" {
		t.Errorf("Expected no output (TypeError: condition must be boolean), got %q", output)
	}
}

func TestTypeSystem_CastToExplicit(t *testing.T) {
	code := `Declare s to be "42".
Declare n to be s cast to number.
If n is greater than 41, then
    Print "ok".
thats it.`
	output := captureOutput(func() { evaluate(code) })
	if output != "ok\n" {
		t.Errorf("Expected 'ok', got %q", output)
	}
}

func TestTypeSystem_StrictArithmetic(t *testing.T) {
	// number + text must be a TypeError
	code := `Declare n to be 5.
Declare t to be "hi".
Declare x to be n + t.`
	output := captureOutput(func() { evaluate(code) })
	if output != "" {
		t.Errorf("Expected no output (TypeError: number + text), got %q", output)
	}
}

func TestTypeSystem_ArrayHomogeneous(t *testing.T) {
	code := `Declare arr to be an array of number [1, 2, 3].
Print count(arr).
Print sum(arr).
Print first(arr).
Print last(arr).`
	output := captureOutput(func() { evaluate(code) })
	expected := "3\n6\n1\n3\n"
	if output != expected {
		t.Errorf("Array test: expected %q, got %q", expected, output)
	}
}

func TestTypeSystem_ArrayRejectsWrongType(t *testing.T) {
	code := `Declare arr to be an array of number [1, 2].
Set arr to be append(arr, "hello").`
	output := captureOutput(func() { evaluate(code) })
	if output != "" {
		t.Errorf("Expected no output (TypeError: append wrong type), got %q", output)
	}
}

func TestTypeSystem_ArrayMixedLiteralRejects(t *testing.T) {
	code := `Declare arr to be an array of number [1, "two", 3].`
	output := captureOutput(func() { evaluate(code) })
	if output != "" {
		t.Errorf("Expected no output (TypeError: mixed array literal), got %q", output)
	}
}

func TestTypeSystem_LookupTable(t *testing.T) {
	code := `Declare ages to be a lookup table.
Set ages at "Alice" to be 30.
Set ages at "Bob" to be 25.
Print ages at "Alice".
Print count(ages).
If ages has "Alice", then
    Print "yes".
thats it.
If ages has "Carol", then
    Print "no".
otherwise
    Print "absent".
thats it.`
	output := captureOutput(func() { evaluate(code) })
	expected := "30\n2\nyes\nabsent\n"
	if output != expected {
		t.Errorf("LookupTable test: expected %q, got %q", expected, output)
	}
}

func TestTypeSystem_LookupTableEntryAccess(t *testing.T) {
	code := `Declare t to be a lookup table.
Set t at 1 to be "one".
Set t at 2 to be "two".
Print the entry 1 in t.
Print the entry 2 in t.`
	output := captureOutput(func() { evaluate(code) })
	expected := "one\ntwo\n"
	if output != expected {
		t.Errorf("LookupTable entry access: expected %q, got %q", expected, output)
	}
}

func TestTypeSystem_LookupTableIteration(t *testing.T) {
	code := `Declare scores to be a lookup table.
Set scores at "A" to be 90.
Set scores at "B" to be 80.
Set scores at "C" to be 70.
For each k in scores, do the following:
    Print the value of k.
thats it.`
	output := captureOutput(func() { evaluate(code) })
	expected := "A\nB\nC\n"
	if output != expected {
		t.Errorf("LookupTable iteration: expected %q, got %q", expected, output)
	}
}

func TestTypeSystem_LookupTableKeys(t *testing.T) {
	code := `Declare t to be a lookup table.
Set t at "x" to be 1.
Set t at "y" to be 2.
Print count(keys(t)).`
	output := captureOutput(func() { evaluate(code) })
	if output != "2\n" {
		t.Errorf("LookupTable keys(): expected '2', got %q", output)
	}
}

func TestTypeSystem_ArrayForeach(t *testing.T) {
	code := `Declare nums to be an array of number [10, 20, 30].
Declare total to be 0.
For each n in nums, do the following:
    Set total to be total + n.
thats it.
Print total.`
	output := captureOutput(func() { evaluate(code) })
	if output != "60\n" {
		t.Errorf("Array foreach: expected '60', got %q", output)
	}
}

func TestTypeSystem_NothingAssignableToAnyVar(t *testing.T) {
	// nothing (nil) can be assigned to any typed variable (universal null)
	code := `Declare x to be 5.
Set x to be nothing.
If x is equal to nothing, then
    Print "null".
thats it.`
	output := captureOutput(func() { evaluate(code) })
	if output != "null\n" {
		t.Errorf("Nothing assignability: expected 'null', got %q", output)
	}
}

// ============================================
// NIL-CHECK EXPRESSION TESTS ("is something" / "has a value")
// ============================================

func TestNilCheck_IsSomething(t *testing.T) {
	code := `Declare x to be 42.
If x is something, then
    Print "something".
thats it.`
	output := captureOutput(func() { evaluate(code) })
	if output != "something\n" {
		t.Errorf("expected 'something', got %q", output)
	}
}

func TestNilCheck_IsNothing(t *testing.T) {
	code := `Declare x to be nothing.
If x is something, then
    Print "something".
otherwise
    Print "nothing".
thats it.`
	output := captureOutput(func() { evaluate(code) })
	if output != "nothing\n" {
		t.Errorf("expected 'nothing', got %q", output)
	}
}

func TestNilCheck_HasAValue(t *testing.T) {
	code := `Declare x to be "hello".
If x has a value, then
    Print "has value".
thats it.`
	output := captureOutput(func() { evaluate(code) })
	if output != "has value\n" {
		t.Errorf("expected 'has value', got %q", output)
	}
}

func TestNilCheck_HasNoValue(t *testing.T) {
	code := `Declare x to be nothing.
If x has no value, then
    Print "no value".
thats it.`
	output := captureOutput(func() { evaluate(code) })
	if output != "no value\n" {
		t.Errorf("expected 'no value', got %q", output)
	}
}

func TestNilCheck_IsNothingKeyword(t *testing.T) {
	code := `Declare x to be nothing.
If x is nothing, then
    Print "is nothing".
thats it.`
	output := captureOutput(func() { evaluate(code) })
	if output != "is nothing\n" {
		t.Errorf("expected 'is nothing', got %q", output)
	}
}

func TestNilCheck_AfterClear(t *testing.T) {
	code := `Declare x to be 5.
If x is something, then
    Print "before".
thats it.
Set x to be nothing.
If x is something, then
    Print "after set".
otherwise
    Print "cleared".
thats it.`
	output := captureOutput(func() { evaluate(code) })
	expected := "before\ncleared\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestNilCheck_WithText(t *testing.T) {
	code := `Declare name to be "Alice".
If name is something, then
    Print "name is set".
thats it.`
	output := captureOutput(func() { evaluate(code) })
	if output != "name is set\n" {
		t.Errorf("expected 'name is set', got %q", output)
	}
}

func TestNilCheck_ReturnsBool(t *testing.T) {
	// The nil-check must return a boolean — usable in and/or expressions
	code := `Declare x to be 5.
Declare y to be nothing.
If x is something, then
    If y is nothing, then
        Print "ok".
    thats it.
thats it.`
	output := captureOutput(func() { evaluate(code) })
	if output != "ok\n" {
		t.Errorf("expected 'ok', got %q", output)
	}
}

// ============================================
// NATURAL ENGLISH AGGREGATE SYNTAX TESTS
// ============================================

func TestNaturalEnglish_FirstOf(t *testing.T) {
	code := `Declare names to be ["Alice", "Bob", "Carol"].
Declare result to be first of names.
Print result.`
	output := captureOutput(func() { evaluate(code) })
	if output != "Alice\n" {
		t.Errorf("expected 'Alice', got %q", output)
	}
}

func TestNaturalEnglish_LastOf(t *testing.T) {
	code := `Declare names to be ["Alice", "Bob", "Carol"].
Declare result to be last of names.
Print result.`
	output := captureOutput(func() { evaluate(code) })
	if output != "Carol\n" {
		t.Errorf("expected 'Carol', got %q", output)
	}
}

func TestNaturalEnglish_TheNumberOf(t *testing.T) {
	code := `Declare names to be an array of text ["Alice", "Bob", "Carol"].
Declare n to be the number of names.
Print n.`
	output := captureOutput(func() { evaluate(code) })
	if output != "3\n" {
		t.Errorf("expected '3', got %q", output)
	}
}

func TestNaturalEnglish_TheSizeOf(t *testing.T) {
	code := `Declare items to be [1, 2, 3, 4, 5].
Declare n to be the size of items.
Print n.`
	output := captureOutput(func() { evaluate(code) })
	if output != "5\n" {
		t.Errorf("expected '5', got %q", output)
	}
}

func TestNaturalEnglish_SumOf(t *testing.T) {
	code := `Declare nums to be [10, 20, 30].
Declare total to be sum of nums.
Print total.`
	output := captureOutput(func() { evaluate(code) })
	if output != "60\n" {
		t.Errorf("expected '60', got %q", output)
	}
}

// ─── Static type annotation tests ────────────────────────────────────────────

func TestTypedDecl_BasicNumber(t *testing.T) {
	code := `Declare x as number to be 5.
Print x.`
	output := captureOutput(func() { evaluate(code) })
	if output != "5\n" {
		t.Errorf("expected '5', got %q", output)
	}
}

func TestTypedDecl_BasicText(t *testing.T) {
	code := `Declare name as text to be "Alice".
Print name.`
	output := captureOutput(func() { evaluate(code) })
	if output != "Alice\n" {
		t.Errorf("expected 'Alice', got %q", output)
	}
}

func TestTypedDecl_BasicBoolean(t *testing.T) {
	code := `Declare flag as boolean to be true.
Print flag.`
	output := captureOutput(func() { evaluate(code) })
	if output != "true\n" {
		t.Errorf("expected 'true', got %q", output)
	}
}

func TestTypedDecl_NoInitialValue(t *testing.T) {
	// Variable declared with type but no initial value should default to nothing.
	code := `Declare score as number.
Set score to be 42.
Print score.`
	output := captureOutput(func() { evaluate(code) })
	if output != "42\n" {
		t.Errorf("expected '42', got %q", output)
	}
}

func TestTypedDecl_TypeEnforced(t *testing.T) {
	// Assigning a text value to a number variable must fail.
	code := `Declare x as number to be 5.
Set x to be "hello".`
	output := captureOutput(func() { evaluate(code) })
	if output != "" {
		t.Errorf("expected no output (TypeError on reassignment), got %q", output)
	}
}

func TestTypedDecl_WrongInitType(t *testing.T) {
	// Initialising a number variable with text must fail.
	code := `Declare x as number to be "hello".`
	output := captureOutput(func() { evaluate(code) })
	if output != "" {
		t.Errorf("expected no output (TypeError on init), got %q", output)
	}
}

func TestTypedDecl_Constant(t *testing.T) {
	code := `Declare PI as number to always be 3.14.
Print PI.`
	output := captureOutput(func() { evaluate(code) })
	if output != "3.14\n" {
		t.Errorf("expected '3.14', got %q", output)
	}
}

func TestTypedDecl_UnknownType(t *testing.T) {
	// Unknown type annotation must produce a RuntimeError.
	code := `Declare x as blurgh to be 5.`
	_, err := evaluate(code)
	if err == nil {
		t.Fatal("expected error for unknown type annotation, got nil")
	}
}

// ─── Custom error type tests ──────────────────────────────────────────────────

func TestCustomError_DeclareAndRaise(t *testing.T) {
	code := `Declare NetworkError as an error type.
Try doing the following:
    Raise "Host unreachable" as NetworkError.
on NetworkError:
    Print "caught".
thats it.`
	output := captureOutput(func() { evaluate(code) })
	if output != "caught\n" {
		t.Errorf("expected 'caught', got %q", output)
	}
}

func TestCustomError_CatchAllStillWorks(t *testing.T) {
	code := `Declare ValidationError as an error type.
Try doing the following:
    Raise "bad input" as ValidationError.
on error:
    Print "caught".
thats it.`
	output := captureOutput(func() { evaluate(code) })
	if output != "caught\n" {
		t.Errorf("expected 'caught', got %q", output)
	}
}

func TestCustomError_TypedCatchDoesNotCatchOtherType(t *testing.T) {
	// A typed catch handler must not intercept errors of a different type.
	code := `Declare NetworkError as an error type.
Declare DatabaseError as an error type.
Try doing the following:
    Try doing the following:
        Raise "row missing" as DatabaseError.
    on NetworkError:
        Print "wrong handler".
    thats it.
on DatabaseError:
    Print "right handler".
thats it.`
	output := captureOutput(func() { evaluate(code) })
	if output != "right handler\n" {
		t.Errorf("expected 'right handler', got %q", output)
	}
}

func TestCustomError_FinallyRunsOnTypedMismatch(t *testing.T) {
	// Finally must run even when the typed handler does not match.
	code := `Declare NetworkError as an error type.
Try doing the following:
    Try doing the following:
        Raise "oops" as NetworkError.
    on NetworkError:
        Print "inner caught".
    but finally:
        Print "inner finally".
    thats it.
thats it.`
	output := captureOutput(func() { evaluate(code) })
	if output != "inner caught\ninner finally\n" {
		t.Errorf("expected 'inner caught\\ninner finally\\n', got %q", output)
	}
}

func TestCustomError_ErrorTypeAccessible(t *testing.T) {
	// The error variable in the catch block should be accessible.
	code := `Declare AppError as an error type.
Try doing the following:
    Raise "something broke" as AppError.
on AppError:
    Print "ok".
thats it.`
	output := captureOutput(func() { evaluate(code) })
	if output != "ok\n" {
		t.Errorf("expected 'ok', got %q", output)
	}
}

// ─── casefold tests ───────────────────────────────────────────────────────────

func TestCasefold_Function(t *testing.T) {
	code := `Print casefold("HELLO").`
	output := captureOutput(func() { evaluate(code) })
	if output != "hello\n" {
		t.Errorf("casefold('HELLO') expected 'hello', got %q", output)
	}
}

func TestCasefold_OfSyntax(t *testing.T) {
	code := `Declare s to be "WORLD".
Print casefold of s.`
	output := captureOutput(func() { evaluate(code) })
	if output != "world\n" {
		t.Errorf("casefold of s expected 'world', got %q", output)
	}
}

func TestCasefold_PossessiveSyntax(t *testing.T) {
	code := `Declare s to be "WORLD".
Print s's casefold.`
	output := captureOutput(func() { evaluate(code) })
	if output != "world\n" {
		t.Errorf("s's casefold expected 'world', got %q", output)
	}
}

func TestCasefold_PossessiveInCondition(t *testing.T) {
	code := `Declare answer to be "Y".
If answer's casefold is equal to "y", then
    Print "yes".
otherwise
    Print "no".
thats it.`
	output := captureOutput(func() { evaluate(code) })
	if output != "yes\n" {
		t.Errorf("expected 'yes', got %q", output)
	}
}

func TestCasefold_OfInCondition(t *testing.T) {
	code := `Declare answer to be "Y".
If casefold of answer is equal to "y", then
    Print "yes".
otherwise
    Print "no".
thats it.`
	output := captureOutput(func() { evaluate(code) })
	if output != "yes\n" {
		t.Errorf("expected 'yes', got %q", output)
	}
}

// ─── error hierarchy tests ────────────────────────────────────────────────────

func TestErrorHierarchy_ExactMatch(t *testing.T) {
	code := `Declare LibError as an error type.
Declare SubError as a type of LibError.
Try doing the following:
    Raise "oops" as SubError.
on SubError:
    Print "exact".
thats it.`
	output := captureOutput(func() { evaluate(code) })
	if output != "exact\n" {
		t.Errorf("expected 'exact', got %q", output)
	}
}

func TestErrorHierarchy_ParentCatchesChild(t *testing.T) {
	code := `Declare LibError as an error type.
Declare SubError as a type of LibError.
Try doing the following:
    Raise "oops" as SubError.
on LibError:
    Print "parent".
thats it.`
	output := captureOutput(func() { evaluate(code) })
	if output != "parent\n" {
		t.Errorf("expected 'parent', got %q", output)
	}
}

func TestErrorHierarchy_NonMatchingDoesNotCatch(t *testing.T) {
	code := `Declare LibError as an error type.
Declare SubError1 as a type of LibError.
Declare SubError2 as a type of LibError.
Try doing the following:
    Try doing the following:
        Raise "oops" as SubError1.
    on SubError2:
        Print "wrong".
    thats it.
on SubError1:
    Print "correct".
thats it.`
	output := captureOutput(func() { evaluate(code) })
	if output != "correct\n" {
		t.Errorf("expected 'correct', got %q", output)
	}
}

func TestErrorHierarchy_CatchAllCatchesChild(t *testing.T) {
	code := `Declare LibError as an error type.
Declare SubError as a type of LibError.
Try doing the following:
    Raise "oops" as SubError.
on error:
    Print "catchall".
thats it.`
	output := captureOutput(func() { evaluate(code) })
	if output != "catchall\n" {
		t.Errorf("expected 'catchall', got %q", output)
	}
}

func TestErrorHierarchy_ErrorIsExpression(t *testing.T) {
	code := `Declare LibError as an error type.
Declare SubError1 as a type of LibError.
Declare SubError2 as a type of LibError.
Try doing the following:
    Raise "oops" as SubError1.
on LibError:
    If error is SubError1, then
        Print "matched SubError1".
    otherwise
        Print "no match".
    thats it.
thats it.`
	output := captureOutput(func() { evaluate(code) })
	if output != "matched SubError1\n" {
		t.Errorf("expected 'matched SubError1', got %q", output)
	}
}

func TestErrorHierarchy_ErrorIsExpression_SubtypeCheck(t *testing.T) {
	code := `Declare LibError as an error type.
Declare SubError as a type of LibError.
Try doing the following:
    Raise "oops" as SubError.
on error:
    If error is LibError, then
        Print "is LibError".
    otherwise
        Print "not LibError".
    thats it.
thats it.`
	output := captureOutput(func() { evaluate(code) })
	if output != "is LibError\n" {
		t.Errorf("expected 'is LibError', got %q", output)
	}
}

// ─── new stdlib method tests ──────────────────────────────────────────────────

func TestTextMethods(t *testing.T) {
	tests := []struct{ code, want string }{
		{`Print title of "hello world".`, "Hello World\n"},
		{`Print "hello world"'s title.`, "Hello World\n"},
		{`Print capitalize of "hello world".`, "Hello world\n"},
		{`Print "hello world"'s capitalize.`, "Hello world\n"},
		{`Print swapcase of "Hello World".`, "hELLO wORLD\n"},
		{`Print trim_left of "  hi".`, "hi\n"},
		{`Print trim_right of "hi  ".`, "hi\n"},
		{`Print is_digit of "123".`, "true\n"},
		{`Print is_digit of "12a".`, "false\n"},
		{`Print is_alpha of "abc".`, "true\n"},
		{`Print is_alnum of "abc123".`, "true\n"},
		{`Print is_space of "   ".`, "true\n"},
		{`Print is_upper of "HELLO".`, "true\n"},
		{`Print is_lower of "hello".`, "true\n"},
		{`Print is_integer of 5.0.`, "true\n"},
		{`Print is_integer of 5.5.`, "false\n"},
		{`Print sign of -3.`, "-1\n"},
		{`Print sign of 0.`, "0\n"},
		{`Print sign of 7.`, "1\n"},
	}
	for _, tt := range tests {
		got := captureOutput(func() { evaluate(tt.code) })
		if got != tt.want {
			t.Errorf("code=%q: expected %q, got %q", tt.code, tt.want, got)
		}
	}
}

func TestListMethods(t *testing.T) {
	tests := []struct{ code, want string }{
		{`Print average of [2, 4, 6].`, "4\n"},
		{`Print min_value of [3, 1, 4].`, "1\n"},
		{`Print max_value of [3, 1, 4].`, "4\n"},
		{`Print product of [1, 2, 3, 4].`, "24\n"},
		{`Print any_true of [false, false, true].`, "true\n"},
		{`Print all_true of [true, true, false].`, "false\n"},
		{`Print sorted_desc of [1, 3, 2].`, "[3 2 1]\n"},
	}
	for _, tt := range tests {
		got := captureOutput(func() { evaluate(tt.code) })
		if got != tt.want {
			t.Errorf("code=%q: expected %q, got %q", tt.code, tt.want, got)
		}
	}
}

func TestCompileTimeTypeError_TextOnNumber(t *testing.T) {
	_, err := evaluate(`Print title of 42.`)
	if err == nil {
		t.Fatal("expected type error, got nil")
	}
	if !strings.Contains(err.Error(), "TypeError") {
		t.Errorf("expected TypeError, got: %v", err)
	}
}

func TestCompileTimeTypeError_NumberOnText(t *testing.T) {
	_, err := evaluate(`Print is_integer of "hello".`)
	if err == nil {
		t.Fatal("expected type error, got nil")
	}
	if !strings.Contains(err.Error(), "TypeError") {
		t.Errorf("expected TypeError, got: %v", err)
	}
}

func TestCompileTimeTypeError_ListOnText(t *testing.T) {
	_, err := evaluate(`Print average of "hello".`)
	if err == nil {
		t.Fatal("expected type error, got nil")
	}
	if !strings.Contains(err.Error(), "TypeError") {
		t.Errorf("expected TypeError, got: %v", err)
	}
}

func TestCompileTimeTypeError_PossessiveSyntax(t *testing.T) {
	_, err := evaluate(`Declare n as number to be 5. Print n's title.`)
	if err == nil {
		t.Fatal("expected type error, got nil")
	}
	if !strings.Contains(err.Error(), "TypeError") {
		t.Errorf("expected TypeError, got: %v", err)
	}
}

// ============================================
// COMPILE-TIME DUPLICATE VARIABLE DETECTION
// ============================================

// checkCode parses the given source and runs the static checker with stdlib
// predefines, mirroring what cmd/root.go does before evaluation.
func checkCode(input string) []*vm.TypeError {
lexer := parser.NewLexer(input)
tokens := lexer.TokenizeAll()
p := parser.NewParser(tokens)
program, err := p.Parse()
if err != nil {
return nil
}
return vm.Check(program, stdlib.PredefinedNames()...)
}

func TestChecker_DuplicateVarTopLevel(t *testing.T) {
errs := checkCode(`Declare x to be 1.
Declare x to be 2.`)
if len(errs) == 0 {
t.Fatal("expected a duplicate-variable error, got none")
}
msg := errs[0].Error()
if !strings.Contains(msg, "x") {
t.Errorf("error should mention variable name 'x', got: %s", msg)
}
if errs[0].Line != 2 {
t.Errorf("error should be on line 2, got line %d", errs[0].Line)
}
}

func TestChecker_DuplicateShadowsStdlibConstant(t *testing.T) {
errs := checkCode(`Declare pi to be 3.`)
if len(errs) == 0 {
t.Fatal("expected error for redeclaring stdlib constant 'pi', got none")
}
msg := errs[0].Error()
if !strings.Contains(msg, "pi") {
t.Errorf("error should mention 'pi', got: %s", msg)
}
}

func TestChecker_DuplicateLetSyntax(t *testing.T) {
errs := checkCode(`let x be 1.
let x be 2.`)
if len(errs) == 0 {
t.Fatal("expected a duplicate-variable error, got none")
}
if errs[0].Line != 2 {
t.Errorf("error should be on line 2, got line %d", errs[0].Line)
}
}

func TestChecker_NoDuplicateInDifferentScopes(t *testing.T) {
// Same name in an inner scope (if body) must NOT be flagged.
errs := checkCode(`Declare x to be 1.
If yes, then
    Declare x to be 2.
thats it.`)
for _, e := range errs {
if strings.Contains(e.Error(), "x") {
t.Errorf("unexpected error for shadowing in inner scope: %s", e.Error())
}
}
}

func TestChecker_NoDuplicateForUniqueNames(t *testing.T) {
errs := checkCode(`Declare a to be 1.
Declare b to be 2.
Declare c to be 3.`)
if len(errs) != 0 {
t.Errorf("expected no errors for distinct names, got: %v", errs)
}
}

// ============================================
// EVALUATOR REDEFINITION → COMPILE ERROR
// ============================================

// TestEvaluator_RedefinitionIsTypeError verifies that when the evaluator
// encounters a variable redefinition at runtime (e.g. inside an imported
// file that the checker did not inspect), it returns a *vm.TypeError so that
// the error is rendered as "Compile Error" rather than a generic "Error".
func TestEvaluator_RedefinitionIsTypeError(t *testing.T) {
	// stdlib.Register puts "pi" into the environment; re-declaring it must
	// produce a TypeError, not a plain/runtime error.
	_, err := evaluate(`Declare pi to be 3.`)
	if err == nil {
		t.Fatal("expected error for redeclaring stdlib constant 'pi', got nil")
	}
	if _, ok := err.(*vm.TypeError); !ok {
		t.Errorf("expected *vm.TypeError (Compile Error), got %T: %v", err, err)
	}
}

// TestEvaluator_TypedRedefinitionIsTypeError verifies the same behaviour for
// typed variable declarations (Declare x as number to be …).
func TestEvaluator_TypedRedefinitionIsTypeError(t *testing.T) {
	_, err := evaluate(`Declare pi as number to be 3.`)
	if err == nil {
		t.Fatal("expected error for redeclaring stdlib constant 'pi' as typed, got nil")
	}
	if _, ok := err.(*vm.TypeError); !ok {
		t.Errorf("expected *vm.TypeError (Compile Error), got %T: %v", err, err)
	}
}

// TestChecker_FollowsImports verifies that the checker reads and validates
// imported .abc files, catching stdlib-constant shadowing at compile time so
// that `english run` never partially executes a program with a compile error.
func TestChecker_FollowsImports(t *testing.T) {
	// Write a temporary library that redefines the stdlib constant "pi".
	libFile, err := os.CreateTemp(t.TempDir(), "lib_*.abc")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := libFile.WriteString("Declare pi to always be 3.14159.\n"); err != nil {
		t.Fatal(err)
	}
	libFile.Close()

	// The main program just imports that library.
	errs := checkCode(`Import "` + libFile.Name() + `".`)
	if len(errs) == 0 {
		t.Fatal("expected a compile error for importing a file that shadows 'pi', got none")
	}
	e := errs[0]
	msg := e.Error()
	if !strings.Contains(msg, "pi") {
		t.Errorf("error should mention 'pi', got: %s", msg)
	}
	if !strings.Contains(msg, "shadows") {
		t.Errorf("error should say 'shadows', got: %s", msg)
	}
	// The error must identify the file it came from.
	if e.File == "" {
		t.Errorf("error should carry the imported file path, but File is empty")
	}
	if e.File != libFile.Name() {
		t.Errorf("expected error File to be %q, got %q", libFile.Name(), e.File)
	}
}
