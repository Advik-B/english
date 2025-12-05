package interpreter

import (
	"testing"
)

func parse(input string) (*Program, error) {
	lexer := NewLexer(input)
	tokens := lexer.TokenizeAll()
	parser := NewParser(tokens)
	return parser.Parse()
}

func TestParserVariableDeclaration(t *testing.T) {
	tests := []struct {
		input      string
		name       string
		isConstant bool
	}{
		{"Declare x to be 5.", "x", false},
		{"Declare y to always be 10.", "y", true},
		{"Declare z to be always 20.", "z", true},
	}

	for _, test := range tests {
		program, err := parse(test.input)
		if err != nil {
			t.Errorf("Input %q: parse error: %v", test.input, err)
			continue
		}

		if len(program.Statements) != 1 {
			t.Errorf("Input %q: expected 1 statement, got %d", test.input, len(program.Statements))
			continue
		}

		varDecl, ok := program.Statements[0].(*VariableDecl)
		if !ok {
			t.Errorf("Input %q: expected VariableDecl, got %T", test.input, program.Statements[0])
			continue
		}

		if varDecl.Name != test.name {
			t.Errorf("Input %q: expected name %q, got %q", test.input, test.name, varDecl.Name)
		}
		if varDecl.IsConstant != test.isConstant {
			t.Errorf("Input %q: expected isConstant %v, got %v", test.input, test.isConstant, varDecl.IsConstant)
		}
	}
}

func TestParserAssignment(t *testing.T) {
	input := "Set x to be 15."
	program, err := parse(input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	if len(program.Statements) != 1 {
		t.Fatalf("Expected 1 statement, got %d", len(program.Statements))
	}

	assignment, ok := program.Statements[0].(*Assignment)
	if !ok {
		t.Fatalf("Expected Assignment, got %T", program.Statements[0])
	}

	if assignment.Name != "x" {
		t.Errorf("Expected name 'x', got %q", assignment.Name)
	}
}

func TestParserFunctionDeclarationWithoutParams(t *testing.T) {
	input := `Declare function greet that does the following:
    Print "Hello".
thats it.`

	program, err := parse(input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	if len(program.Statements) != 1 {
		t.Fatalf("Expected 1 statement, got %d", len(program.Statements))
	}

	funcDecl, ok := program.Statements[0].(*FunctionDecl)
	if !ok {
		t.Fatalf("Expected FunctionDecl, got %T", program.Statements[0])
	}

	if funcDecl.Name != "greet" {
		t.Errorf("Expected function name 'greet', got %q", funcDecl.Name)
	}

	if len(funcDecl.Parameters) != 0 {
		t.Errorf("Expected 0 parameters, got %d", len(funcDecl.Parameters))
	}
}

func TestParserFunctionDeclarationWithParams(t *testing.T) {
	input := `Declare function add that takes a and b and does the following:
    Return a + b.
thats it.`

	program, err := parse(input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	funcDecl, ok := program.Statements[0].(*FunctionDecl)
	if !ok {
		t.Fatalf("Expected FunctionDecl, got %T", program.Statements[0])
	}

	if funcDecl.Name != "add" {
		t.Errorf("Expected function name 'add', got %q", funcDecl.Name)
	}

	if len(funcDecl.Parameters) != 2 {
		t.Errorf("Expected 2 parameters, got %d", len(funcDecl.Parameters))
	}

	if funcDecl.Parameters[0] != "a" || funcDecl.Parameters[1] != "b" {
		t.Errorf("Expected parameters [a, b], got %v", funcDecl.Parameters)
	}
}

func TestParserIfStatement(t *testing.T) {
	input := `If x is equal to 5, then
    Print "yes".
thats it.`

	program, err := parse(input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	ifStmt, ok := program.Statements[0].(*IfStatement)
	if !ok {
		t.Fatalf("Expected IfStatement, got %T", program.Statements[0])
	}

	if ifStmt.Condition == nil {
		t.Error("Expected condition, got nil")
	}

	if len(ifStmt.Then) != 1 {
		t.Errorf("Expected 1 then statement, got %d", len(ifStmt.Then))
	}
}

func TestParserIfElseStatement(t *testing.T) {
	input := `If x is equal to 5, then
    Print "yes".
otherwise
    Print "no".
thats it.`

	program, err := parse(input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	ifStmt, ok := program.Statements[0].(*IfStatement)
	if !ok {
		t.Fatalf("Expected IfStatement, got %T", program.Statements[0])
	}

	if len(ifStmt.Then) != 1 {
		t.Errorf("Expected 1 then statement, got %d", len(ifStmt.Then))
	}

	if len(ifStmt.Else) != 1 {
		t.Errorf("Expected 1 else statement, got %d", len(ifStmt.Else))
	}
}

func TestParserWhileLoop(t *testing.T) {
	input := `repeat the following while x is less than 10:
    Print "loop".
thats it.`

	program, err := parse(input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	whileLoop, ok := program.Statements[0].(*WhileLoop)
	if !ok {
		t.Fatalf("Expected WhileLoop, got %T", program.Statements[0])
	}

	if whileLoop.Condition == nil {
		t.Error("Expected condition, got nil")
	}

	if len(whileLoop.Body) != 1 {
		t.Errorf("Expected 1 body statement, got %d", len(whileLoop.Body))
	}
}

func TestParserForLoop(t *testing.T) {
	input := `repeat the following 5 times:
    Print "loop".
thats it.`

	program, err := parse(input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	forLoop, ok := program.Statements[0].(*ForLoop)
	if !ok {
		t.Fatalf("Expected ForLoop, got %T", program.Statements[0])
	}

	if forLoop.Count == nil {
		t.Error("Expected count, got nil")
	}

	if len(forLoop.Body) != 1 {
		t.Errorf("Expected 1 body statement, got %d", len(forLoop.Body))
	}
}

func TestParserForEachLoop(t *testing.T) {
	input := `for each item in myList, do the following:
    Print the value of item.
thats it.`

	program, err := parse(input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	forEachLoop, ok := program.Statements[0].(*ForEachLoop)
	if !ok {
		t.Fatalf("Expected ForEachLoop, got %T", program.Statements[0])
	}

	if forEachLoop.Item != "item" {
		t.Errorf("Expected item 'item', got %q", forEachLoop.Item)
	}

	if len(forEachLoop.Body) != 1 {
		t.Errorf("Expected 1 body statement, got %d", len(forEachLoop.Body))
	}
}

func TestParserOutputStatement(t *testing.T) {
	tests := []string{
		`Print "Hello".`,
		`Print 42.`,
		`Print the value of x.`,
	}

	for _, input := range tests {
		program, err := parse(input)
		if err != nil {
			t.Errorf("Input %q: parse error: %v", input, err)
			continue
		}

		_, ok := program.Statements[0].(*OutputStatement)
		if !ok {
			t.Errorf("Input %q: expected OutputStatement, got %T", input, program.Statements[0])
		}
	}
}

func TestParserReturnStatement(t *testing.T) {
	input := "Return 5."
	program, err := parse(input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	returnStmt, ok := program.Statements[0].(*ReturnStatement)
	if !ok {
		t.Fatalf("Expected ReturnStatement, got %T", program.Statements[0])
	}

	if returnStmt.Value == nil {
		t.Error("Expected return value, got nil")
	}
}

func TestParserCallStatement(t *testing.T) {
	input := "Call myFunction."
	program, err := parse(input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	callStmt, ok := program.Statements[0].(*CallStatement)
	if !ok {
		t.Fatalf("Expected CallStatement, got %T", program.Statements[0])
	}

	if callStmt.FunctionCall.Name != "myFunction" {
		t.Errorf("Expected function name 'myFunction', got %q", callStmt.FunctionCall.Name)
	}
}

func TestParserArithmeticExpressions(t *testing.T) {
	tests := []string{
		"Declare x to be 1 + 2.",
		"Declare x to be 3 - 1.",
		"Declare x to be 2 * 3.",
		"Declare x to be 6 / 2.",
		"Declare x to be 1 + 2 * 3.",
	}

	for _, input := range tests {
		_, err := parse(input)
		if err != nil {
			t.Errorf("Input %q: parse error: %v", input, err)
		}
	}
}

func TestParserListLiteral(t *testing.T) {
	input := "Declare x to be [1, 2, 3]."
	program, err := parse(input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	varDecl, ok := program.Statements[0].(*VariableDecl)
	if !ok {
		t.Fatalf("Expected VariableDecl, got %T", program.Statements[0])
	}

	list, ok := varDecl.Value.(*ListLiteral)
	if !ok {
		t.Fatalf("Expected ListLiteral, got %T", varDecl.Value)
	}

	if len(list.Elements) != 3 {
		t.Errorf("Expected 3 elements, got %d", len(list.Elements))
	}
}

func TestParserFunctionCallResult(t *testing.T) {
	input := "Set result to be the result of calling add with 5 and 10."
	program, err := parse(input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	assignment, ok := program.Statements[0].(*Assignment)
	if !ok {
		t.Fatalf("Expected Assignment, got %T", program.Statements[0])
	}

	funcCall, ok := assignment.Value.(*FunctionCall)
	if !ok {
		t.Fatalf("Expected FunctionCall, got %T", assignment.Value)
	}

	if funcCall.Name != "add" {
		t.Errorf("Expected function name 'add', got %q", funcCall.Name)
	}

	if len(funcCall.Arguments) != 2 {
		t.Errorf("Expected 2 arguments, got %d", len(funcCall.Arguments))
	}
}

func TestParserCaseInsensitivity(t *testing.T) {
	tests := []string{
		"DECLARE x TO BE 5.",
		"declare x to be 5.",
		"Declare X TO be 5.",
	}

	for _, input := range tests {
		_, err := parse(input)
		if err != nil {
			t.Errorf("Input %q: parse error: %v", input, err)
		}
	}
}

func TestParserErrors(t *testing.T) {
	tests := []struct {
		input       string
		expectError bool
	}{
		{"Declare x to be", true},       // Missing value and period
		{"Set x to be 5", true},         // Missing period
		{"If x is equal to 5", true},    // Incomplete if statement
		{"Declare to be 5.", true},      // Missing variable name
	}

	for _, test := range tests {
		_, err := parse(test.input)
		if test.expectError && err == nil {
			t.Errorf("Input %q: expected error, got nil", test.input)
		}
		if !test.expectError && err != nil {
			t.Errorf("Input %q: unexpected error: %v", test.input, err)
		}
	}
}
