package vm

import (
	"bytes"
	"english/ast"
	"english/parser"
	"io"
	"os"
	"strings"
	"testing"
)

// Helper function to evaluate code
func evaluate(input string) (Value, error) {
	lexer := parser.NewLexer(input)
	tokens := lexer.TokenizeAll()
	p := parser.NewParser(tokens)
	program, err := p.Parse()
	if err != nil {
		return nil, err
	}
	env := NewEnvironment()
	evaluator := NewEvaluator(env)
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
	env := NewEnvironment()
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
	env := NewEnvironment()
	err := env.Define("PI", float64(3.14159), true)
	if err != nil {
		t.Fatalf("Define error: %v", err)
	}

	if !env.IsConstant("PI") {
		t.Error("PI should be a constant")
	}
}

func TestEnvironmentSet(t *testing.T) {
	env := NewEnvironment()
	env.Define("x", float64(5), false)
	env.Set("x", float64(10))

	val, _ := env.Get("x")
	if val != float64(10) {
		t.Errorf("Expected 10, got %v", val)
	}
}

func TestEnvironmentSetConstantError(t *testing.T) {
	env := NewEnvironment()
	env.Define("PI", float64(3.14159), true)
	err := env.Set("PI", float64(3.14))

	if err == nil {
		t.Error("Expected error when reassigning constant")
	}
}

func TestEnvironmentChildScope(t *testing.T) {
	parent := NewEnvironment()
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
	env := NewEnvironment()
	fn := &FunctionValue{Name: "test", Parameters: []string{}, Body: []ast.Statement{}}
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
	env := NewEnvironment()
	env.Define("x", float64(5), false)
	env.Define("y", float64(10), false)

	vars := env.GetAllVariables()
	if len(vars) != 2 {
		t.Errorf("Expected 2 variables, got %d", len(vars))
	}
}

func TestEnvironmentGetAllFunctions(t *testing.T) {
	env := NewEnvironment()
	fn1 := &FunctionValue{Name: "fn1", Parameters: []string{}, Body: []ast.Statement{}}
	fn2 := &FunctionValue{Name: "fn2", Parameters: []string{}, Body: []ast.Statement{}}
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
		result, err := ToNumber(test.input)
		if test.hasError {
			if err == nil {
				t.Errorf("Expected error for %v, got nil", test.input)
			}
		} else {
			if err != nil {
				t.Errorf("Unexpected error for %v: %v", test.input, err)
			}
			if result != test.expected {
				t.Errorf("ToNumber(%v) = %v, want %v", test.input, result, test.expected)
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
		{true, "true"},
		{false, "false"},
		{[]interface{}{float64(1), float64(2), float64(3)}, "[1 2 3]"},
	}

	for _, test := range tests {
		result := ToString(test.input)
		if result != test.expected {
			t.Errorf("ToString(%v) = %q, want %q", test.input, result, test.expected)
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
		result := ToBool(test.input)
		if result != test.expected {
			t.Errorf("ToBool(%v) = %v, want %v", test.input, result, test.expected)
		}
	}
}

// ============================================
// ARITHMETIC TESTS
// ============================================

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
		result, err := Add(test.left, test.right)
		if err != nil {
			t.Errorf("Add(%v, %v) error: %v", test.left, test.right, err)
			continue
		}
		switch expected := test.expected.(type) {
		case float64:
			if result != expected {
				t.Errorf("Add(%v, %v) = %v, want %v", test.left, test.right, result, expected)
			}
		case string:
			if result != expected {
				t.Errorf("Add(%v, %v) = %v, want %v", test.left, test.right, result, expected)
			}
		}
	}
}

func TestSubtract(t *testing.T) {
	result, err := Subtract(float64(10), float64(3))
	if err != nil {
		t.Fatalf("Subtract error: %v", err)
	}
	if result != float64(7) {
		t.Errorf("Subtract(10, 3) = %v, want 7", result)
	}

	// Test error case
	_, err = Subtract("hello", float64(3))
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
		result, err := Multiply(test.left, test.right)
		if err != nil {
			t.Errorf("Multiply(%v, %v) error: %v", test.left, test.right, err)
			continue
		}
		if result != test.expected {
			t.Errorf("Multiply(%v, %v) = %v, want %v", test.left, test.right, result, test.expected)
		}
	}
}

func TestDivide(t *testing.T) {
	result, err := Divide(float64(10), float64(2))
	if err != nil {
		t.Fatalf("Divide error: %v", err)
	}
	if result != float64(5) {
		t.Errorf("Divide(10, 2) = %v, want 5", result)
	}

	// Test division by zero
	_, err = Divide(float64(10), float64(0))
	if err == nil {
		t.Error("Expected error for division by zero")
	}

	// Test non-numeric division
	_, err = Divide("hello", float64(2))
	if err == nil {
		t.Error("Expected error for non-numeric division")
	}
}

func TestModulo(t *testing.T) {
	result, err := Modulo(float64(17), float64(5))
	if err != nil {
		t.Fatalf("Modulo error: %v", err)
	}
	if result != float64(2) {
		t.Errorf("Modulo(17, 5) = %v, want 2", result)
	}

	// Test modulo by zero
	_, err = Modulo(float64(10), float64(0))
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
		result, err := Compare(test.op, test.left, test.right)
		if err != nil {
			t.Errorf("Compare(%q, %v, %v) error: %v", test.op, test.left, test.right, err)
			continue
		}
		if result != test.expected {
			t.Errorf("Compare(%q, %v, %v) = %v, want %v", test.op, test.left, test.right, result, test.expected)
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
		{true, true, true},
		{true, false, false},
		{nil, nil, true},
		{float64(5), "5", false}, // Different types
	}

	for _, test := range tests {
		result := Equals(test.left, test.right)
		if result != test.expected {
			t.Errorf("Equals(%v, %v) = %v, want %v", test.left, test.right, result, test.expected)
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
	code := `Declare x to be "ab" * 3.
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
	code := `Declare list1 to be [1, 2].
Declare list2 to be [3, 4].
Declare combined to be list1 + list2.
Print the length of combined.`

	output := captureOutput(func() {
		evaluate(code)
	})
	if output != "4\n" {
		t.Errorf("Expected '4', got %q", output)
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
	fn := &FunctionValue{
		Name:       "test",
		Parameters: []string{"a", "b"},
		Body:       []ast.Statement{},
		Closure:    NewEnvironment(),
	}

	result := ToString(fn)
	if result != "<function test>" {
		t.Errorf("ToString(function) = %q, want '<function test>'", result)
	}
}

func TestRuntimeErrorFormat(t *testing.T) {
	err := &RuntimeError{
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

