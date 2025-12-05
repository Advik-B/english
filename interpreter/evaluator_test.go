package interpreter

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func evaluate(input string) (Value, error) {
	lexer := NewLexer(input)
	tokens := lexer.TokenizeAll()
	parser := NewParser(tokens)
	program, err := parser.Parse()
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
		input    string
		expected float64
	}{
		{"Declare x to be 1 + 2.", 0},
		{"Declare x to be 5 - 3.", 0},
		{"Declare x to be 4 * 3.", 0},
		{"Declare x to be 8 / 2.", 0},
		{"Declare x to be 1 + 2 * 3.", 0},
		{"Declare x to be -5.", 0},
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
