package parser

import (
	"strings"
	"testing"

	"github.com/Advik-B/english/ast"
	"github.com/Advik-B/english/token"
)

// Helper function to parse input
func parse(input string) (*ast.Program, error) {
	lexer := NewLexer(input)
	tokens := lexer.TokenizeAll()
	parser := NewParser(tokens)
	return parser.Parse()
}

// ============================================
// LEXER TESTS
// ============================================

func TestLexerBasicTokens(t *testing.T) {
	tests := []struct {
		input    string
		expected []token.Type
	}{
		{".", []token.Type{token.PERIOD, token.EOF}},
		{",", []token.Type{token.COMMA, token.EOF}},
		{":", []token.Type{token.COLON, token.EOF}},
		{"(", []token.Type{token.LPAREN, token.EOF}},
		{")", []token.Type{token.RPAREN, token.EOF}},
		{"[", []token.Type{token.LBRACKET, token.EOF}},
		{"]", []token.Type{token.RBRACKET, token.EOF}},
		{"+", []token.Type{token.PLUS, token.EOF}},
		{"-", []token.Type{token.MINUS, token.EOF}},
		{"*", []token.Type{token.STAR, token.EOF}},
		{"/", []token.Type{token.SLASH, token.EOF}},
	}

	for _, test := range tests {
		lexer := NewLexer(test.input)
		tokens := lexer.TokenizeAll()
		if len(tokens) != len(test.expected) {
			t.Errorf("Input %q: got %d tokens, want %d", test.input, len(tokens), len(test.expected))
			continue
		}
		for i, tok := range tokens {
			if tok.Type != test.expected[i] {
				t.Errorf("Input %q, token %d: got %v, want %v", test.input, i, tok.Type, test.expected[i])
			}
		}
	}
}

func TestLexerNumbers(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"42", "42"},
		{"3.14", "3.14"},
		{"0", "0"},
		{"123.456", "123.456"},
		{"0.5", "0.5"},
		{"1000000", "1000000"},
	}

	for _, test := range tests {
		lexer := NewLexer(test.input)
		tok := lexer.NextToken()
		if tok.Type != token.NUMBER {
			t.Errorf("Input %q: got type %v, want NUMBER", test.input, tok.Type)
		}
		if tok.Value != test.expected {
			t.Errorf("Input %q: got value %q, want %q", test.input, tok.Value, test.expected)
		}
	}
}

func TestLexerStrings(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`"hello"`, "hello"},
		{`'world'`, "world"},
		{`"Hello, World!"`, "Hello, World!"},
		{`""`, ""},
		{`"with spaces"`, "with spaces"},
		{`'single quotes'`, "single quotes"},
	}

	for _, test := range tests {
		lexer := NewLexer(test.input)
		tok := lexer.NextToken()
		if tok.Type != token.STRING {
			t.Errorf("Input %q: got type %v, want STRING", test.input, tok.Type)
		}
		if tok.Value != test.expected {
			t.Errorf("Input %q: got value %q, want %q", test.input, tok.Value, test.expected)
		}
	}
}

func TestLexerKeywords(t *testing.T) {
	tests := []struct {
		input    string
		expected token.Type
	}{
		{"declare", token.DECLARE},
		{"DECLARE", token.DECLARE},
		{"Declare", token.DECLARE},
		{"function", token.FUNCTION},
		{"set", token.SET},
		{"print", token.PRINT},
		{"if", token.IF},
		{"then", token.THEN},
		{"otherwise", token.OTHERWISE},
		{"repeat", token.REPEAT},
		{"while", token.WHILE},
		{"for", token.FOR},
		{"each", token.EACH},
		{"return", token.RETURN},
		{"to", token.TO},
		{"be", token.BE},
		{"always", token.ALWAYS},
		{"call", token.CALL},
		{"the", token.THE},
		{"value", token.VALUE},
		{"of", token.OF},
		{"with", token.WITH},
		{"and", token.AND},
		{"takes", token.TAKES},
		{"does", token.DOES},
		{"following", token.FOLLOWING},
		{"thats", token.THATS},
		{"it", token.IT},
		{"times", token.TIMES},
		{"in", token.IN},
		{"do", token.DO},
		{"calling", token.CALLING},
		{"that", token.THAT},
		{"true", token.TRUE},
		{"false", token.FALSE},
		{"toggle", token.TOGGLE},
		{"location", token.LOCATION},
		{"item", token.ITEM},
		{"at", token.AT},
		{"position", token.POSITION},
		{"length", token.LENGTH},
		{"remainder", token.REMAINDER},
		{"divided", token.DIVIDED},
		{"by", token.BY},
		{"import", token.IMPORT},
		{"from", token.FROM},
	}

	for _, test := range tests {
		lexer := NewLexer(test.input)
		tok := lexer.NextToken()
		if tok.Type != test.expected {
			t.Errorf("Input %q: got %v, want %v", test.input, tok.Type, test.expected)
		}
	}
}

func TestLexerIdentifiers(t *testing.T) {
	tests := []string{
		"x", "myVar", "test123", "with_underscore", "_start",
		"camelCase", "snake_case", "MixedCase123",
	}

	for _, input := range tests {
		lexer := NewLexer(input)
		tok := lexer.NextToken()
		if tok.Type != token.IDENTIFIER {
			t.Errorf("Input %q: got %v, want IDENTIFIER", input, tok.Type)
		}
		if tok.Value != input {
			t.Errorf("Input %q: got value %q, want %q", input, tok.Value, input)
		}
	}
}

func TestLexerMultiWordComparisons(t *testing.T) {
	tests := []struct {
		input    string
		expected token.Type
	}{
		{"is equal to", token.IS_EQUAL_TO},
		{"IS EQUAL TO", token.IS_EQUAL_TO},
		{"Is Equal To", token.IS_EQUAL_TO},
		{"is less than", token.IS_LESS_THAN},
		{"IS LESS THAN", token.IS_LESS_THAN},
		{"is greater than", token.IS_GREATER_THAN},
		{"IS GREATER THAN", token.IS_GREATER_THAN},
		{"is less than or equal to", token.IS_LESS_EQUAL},
		{"is greater than or equal to", token.IS_GREATER_EQUAL},
		{"is not equal to", token.IS_NOT_EQUAL},
	}

	for _, test := range tests {
		lexer := NewLexer(test.input)
		tok := lexer.NextToken()
		if tok.Type != test.expected {
			t.Errorf("Input %q: got %v, want %v", test.input, tok.Type, test.expected)
		}
	}
}

func TestLexerComments(t *testing.T) {
	input := "# This is a comment\nDeclare"
	lexer := NewLexer(input)
	tokens := lexer.TokenizeAll()

	// Comments now produce a COMMENT token rather than being silently discarded.
	// Expected: COMMENT, DECLARE, EOF.
	if len(tokens) != 3 {
		t.Errorf("Expected 3 tokens (COMMENT, DECLARE, EOF), got %d", len(tokens))
	}
	if tokens[0].Type != token.COMMENT {
		t.Errorf("First token should be COMMENT, got %v", tokens[0].Type)
	}
	if tokens[0].Value != "This is a comment" {
		t.Errorf("COMMENT value should be %q, got %q", "This is a comment", tokens[0].Value)
	}
	if tokens[1].Type != token.DECLARE {
		t.Errorf("Second token should be DECLARE, got %v", tokens[1].Type)
	}
}

func TestLexerLineAndColumn(t *testing.T) {
	input := "Declare x to be 5."
	lexer := NewLexer(input)
	tok := lexer.NextToken()

	if tok.Line != 1 {
		t.Errorf("Expected line 1, got %d", tok.Line)
	}
	if tok.Col < 1 {
		t.Errorf("Expected col >= 1, got %d", tok.Col)
	}
}

func TestLexerCompleteStatement(t *testing.T) {
	input := "Declare x to be 5."
	expected := []token.Type{
		token.DECLARE, token.IDENTIFIER, token.TO, token.BE, token.NUMBER, token.PERIOD, token.EOF,
	}

	lexer := NewLexer(input)
	tokens := lexer.TokenizeAll()

	if len(tokens) != len(expected) {
		t.Errorf("Expected %d tokens, got %d", len(expected), len(tokens))
		return
	}

	for i, tok := range tokens {
		if tok.Type != expected[i] {
			t.Errorf("Token %d: got %v, want %v", i, tok.Type, expected[i])
		}
	}
}

func TestLexerWhitespaceHandling(t *testing.T) {
	input := "   Declare   x   to   be   5   .   "
	expected := []token.Type{
		token.DECLARE, token.IDENTIFIER, token.TO, token.BE, token.NUMBER, token.PERIOD, token.EOF,
	}

	lexer := NewLexer(input)
	tokens := lexer.TokenizeAll()

	if len(tokens) != len(expected) {
		t.Errorf("Expected %d tokens, got %d", len(expected), len(tokens))
		return
	}

	for i, tok := range tokens {
		if tok.Type != expected[i] {
			t.Errorf("Token %d: got %v, want %v", i, tok.Type, expected[i])
		}
	}
}

// ============================================
// PARSER TESTS
// ============================================

func TestParserVariableDeclaration(t *testing.T) {
	tests := []struct {
		input      string
		name       string
		isConstant bool
	}{
		{"Declare x to be 5.", "x", false},
		{"Declare y to always be 10.", "y", true},
		{"Declare z to be always 20.", "z", true},
		{"Declare myVar to be 100.", "myVar", false},
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

		varDecl, ok := program.Statements[0].(*ast.VariableDecl)
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

	assignment, ok := program.Statements[0].(*ast.Assignment)
	if !ok {
		t.Fatalf("Expected Assignment, got %T", program.Statements[0])
	}

	if assignment.Name != "x" {
		t.Errorf("Expected name 'x', got %q", assignment.Name)
	}
}

func TestParserAssignmentWithoutBe(t *testing.T) {
	// Test "set x to 10" syntax (without "be")
	input := "Set x to 15."
	program, err := parse(input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	if len(program.Statements) != 1 {
		t.Fatalf("Expected 1 statement, got %d", len(program.Statements))
	}

	assignment, ok := program.Statements[0].(*ast.Assignment)
	if !ok {
		t.Fatalf("Expected Assignment, got %T", program.Statements[0])
	}

	if assignment.Name != "x" {
		t.Errorf("Expected name 'x', got %q", assignment.Name)
	}

	// Verify the value is correct
	numLit, ok := assignment.Value.(*ast.NumberLiteral)
	if !ok {
		t.Fatalf("Expected NumberLiteral, got %T", assignment.Value)
	}
	if numLit.Value != 15 {
		t.Errorf("Expected value 15, got %v", numLit.Value)
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

	funcDecl, ok := program.Statements[0].(*ast.FunctionDecl)
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

	funcDecl, ok := program.Statements[0].(*ast.FunctionDecl)
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

func TestParserFunctionDeclarationSingleParam(t *testing.T) {
	input := `Declare function double that takes x and does the following:
    Return x * 2.
thats it.`

	program, err := parse(input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	funcDecl, ok := program.Statements[0].(*ast.FunctionDecl)
	if !ok {
		t.Fatalf("Expected FunctionDecl, got %T", program.Statements[0])
	}

	if len(funcDecl.Parameters) != 1 {
		t.Errorf("Expected 1 parameter, got %d", len(funcDecl.Parameters))
	}

	if funcDecl.Parameters[0] != "x" {
		t.Errorf("Expected parameter 'x', got %q", funcDecl.Parameters[0])
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

	ifStmt, ok := program.Statements[0].(*ast.IfStatement)
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

	ifStmt, ok := program.Statements[0].(*ast.IfStatement)
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

func TestParserIfElseIfStatement(t *testing.T) {
	input := `If x is equal to 1, then
    Print "one".
otherwise if x is equal to 2, then
    Print "two".
otherwise
    Print "other".
thats it.`

	program, err := parse(input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	ifStmt, ok := program.Statements[0].(*ast.IfStatement)
	if !ok {
		t.Fatalf("Expected IfStatement, got %T", program.Statements[0])
	}

	if len(ifStmt.ElseIf) != 1 {
		t.Errorf("Expected 1 else-if part, got %d", len(ifStmt.ElseIf))
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

	whileLoop, ok := program.Statements[0].(*ast.WhileLoop)
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

	forLoop, ok := program.Statements[0].(*ast.ForLoop)
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

	forEachLoop, ok := program.Statements[0].(*ast.ForEachLoop)
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

func TestParserForEachLoopCustomName(t *testing.T) {
	input := `for each x in myList, do the following:
    Print the value of x.
thats it.`

	program, err := parse(input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	forEachLoop, ok := program.Statements[0].(*ast.ForEachLoop)
	if !ok {
		t.Fatalf("Expected ForEachLoop, got %T", program.Statements[0])
	}

	if forEachLoop.Item != "x" {
		t.Errorf("Expected item 'x', got %q", forEachLoop.Item)
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

		_, ok := program.Statements[0].(*ast.OutputStatement)
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

	returnStmt, ok := program.Statements[0].(*ast.ReturnStatement)
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

	callStmt, ok := program.Statements[0].(*ast.CallStatement)
	if !ok {
		t.Fatalf("Expected CallStatement, got %T", program.Statements[0])
	}

	if callStmt.FunctionCall.Name != "myFunction" {
		t.Errorf("Expected function name 'myFunction', got %q", callStmt.FunctionCall.Name)
	}
}

func TestParserImportStatement(t *testing.T) {
	tests := []struct {
		input         string
		expectedPath  string
		expectedItems []string
		expectedAll   bool
		expectedSafe  bool
	}{
		{`Import "library.abc".`, "library.abc", nil, false, false},
		{`Import from "utils.abc".`, "utils.abc", nil, false, false},
		{`Import square from "math.abc".`, "math.abc", []string{"square"}, false, false},
		{`Import square and cube from "math.abc".`, "math.abc", []string{"square", "cube"}, false, false},
		{`Import add, multiply and divide from "ops.abc".`, "ops.abc", []string{"add", "multiply", "divide"}, false, false},
		{`Import everything from "lib.abc".`, "lib.abc", nil, true, false},
		{`Import all from "lib.abc".`, "lib.abc", nil, true, false},
		{`Import all from "lib.abc" safely.`, "lib.abc", nil, true, true},
		{`Import square and cube from "math.abc" safely.`, "math.abc", []string{"square", "cube"}, false, true},
	}

	for _, test := range tests {
		program, err := parse(test.input)
		if err != nil {
			t.Errorf("Input %q: parse error: %v", test.input, err)
			continue
		}

		importStmt, ok := program.Statements[0].(*ast.ImportStatement)
		if !ok {
			t.Errorf("Input %q: expected ImportStatement, got %T", test.input, program.Statements[0])
			continue
		}

		if importStmt.Path != test.expectedPath {
			t.Errorf("Input %q: expected path %q, got %q", test.input, test.expectedPath, importStmt.Path)
		}

		if len(importStmt.Items) != len(test.expectedItems) {
			t.Errorf("Input %q: expected %d items, got %d", test.input, len(test.expectedItems), len(importStmt.Items))
		} else {
			for i, item := range test.expectedItems {
				if importStmt.Items[i] != item {
					t.Errorf("Input %q: expected item[%d] %q, got %q", test.input, i, item, importStmt.Items[i])
				}
			}
		}

		if importStmt.ImportAll != test.expectedAll {
			t.Errorf("Input %q: expected ImportAll %v, got %v", test.input, test.expectedAll, importStmt.ImportAll)
		}

		if importStmt.IsSafe != test.expectedSafe {
			t.Errorf("Input %q: expected IsSafe %v, got %v", test.input, test.expectedSafe, importStmt.IsSafe)
		}
	}
}

func TestParserToggleStatement(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Toggle isEnabled.", "isEnabled"},
		{"Toggle the value of flag.", "flag"},
	}

	for _, test := range tests {
		program, err := parse(test.input)
		if err != nil {
			t.Errorf("Input %q: parse error: %v", test.input, err)
			continue
		}

		toggleStmt, ok := program.Statements[0].(*ast.ToggleStatement)
		if !ok {
			t.Errorf("Input %q: expected ToggleStatement, got %T", test.input, program.Statements[0])
			continue
		}

		if toggleStmt.Name != test.expected {
			t.Errorf("Input %q: expected name %q, got %q", test.input, test.expected, toggleStmt.Name)
		}
	}
}

func TestParserArithmeticExpressions(t *testing.T) {
	tests := []string{
		"Declare x to be 1 + 2.",
		"Declare x to be 3 - 1.",
		"Declare x to be 2 * 3.",
		"Declare x to be 6 / 2.",
		"Declare x to be 1 + 2 * 3.",
		"Declare x to be (1 + 2) * 3.",
	}

	for _, input := range tests {
		_, err := parse(input)
		if err != nil {
			t.Errorf("Input %q: parse error: %v", input, err)
		}
	}
}

func TestParserUnaryExpression(t *testing.T) {
	input := "Declare x to be -5."
	program, err := parse(input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	varDecl := program.Statements[0].(*ast.VariableDecl)
	_, ok := varDecl.Value.(*ast.UnaryExpression)
	if !ok {
		t.Errorf("Expected UnaryExpression, got %T", varDecl.Value)
	}
}

func TestParserListLiteral(t *testing.T) {
	tests := []struct {
		input    string
		elements int
	}{
		{"Declare x to be [1, 2, 3].", 3},
		{"Declare x to be [].", 0},
		{"Declare x to be [1].", 1},
		{`Declare x to be ["a", "b"].`, 2},
	}

	for _, test := range tests {
		program, err := parse(test.input)
		if err != nil {
			t.Errorf("Input %q: parse error: %v", test.input, err)
			continue
		}

		varDecl, ok := program.Statements[0].(*ast.VariableDecl)
		if !ok {
			t.Errorf("Input %q: expected VariableDecl, got %T", test.input, program.Statements[0])
			continue
		}

		list, ok := varDecl.Value.(*ast.ListLiteral)
		if !ok {
			t.Errorf("Input %q: expected ListLiteral, got %T", test.input, varDecl.Value)
			continue
		}

		if len(list.Elements) != test.elements {
			t.Errorf("Input %q: expected %d elements, got %d", test.input, test.elements, len(list.Elements))
		}
	}
}

func TestParserRangeLiteral(t *testing.T) {
	tests := []struct {
		input       string
		description string
	}{
		{"Declare x to be [1 .. 10].", "programmer-style range"},
		{"Let r be [0 .. 100].", "let statement with range"},
		{"Declare r to be a range from 5 to 20.", "natural English range"},
		{"Let myRange be a range from 1 to 30.", "natural English with let"},
		{"Declare evens to be [0 .. 10 by 2].", "programmer-style range with step"},
		{"Let odds be a range from 1 to 9 by 2.", "natural English range with step"},
		{"Declare countdown to be [10 .. 0 by -2].", "programmer-style descending range with step"},
		{"Let multiples be a range from 5 to 25 by 5.", "natural English range with custom step"},
	}

	for _, test := range tests {
		program, err := parse(test.input)
		if err != nil {
			t.Errorf("Input %q (%s): parse error: %v", test.input, test.description, err)
			continue
		}

		varDecl, ok := program.Statements[0].(*ast.VariableDecl)
		if !ok {
			t.Errorf("Input %q (%s): expected VariableDecl, got %T", test.input, test.description, program.Statements[0])
			continue
		}

		_, ok = varDecl.Value.(*ast.RangeLiteral)
		if !ok {
			t.Errorf("Input %q (%s): expected RangeLiteral, got %T", test.input, test.description, varDecl.Value)
			continue
		}
	}
}

func TestLexerDotDotOperator(t *testing.T) {
	tests := []struct {
		input    string
		expected []token.Type
	}{
		{"..", []token.Type{token.DOTDOT, token.EOF}},
		{"[1 .. 10]", []token.Type{token.LBRACKET, token.NUMBER, token.DOTDOT, token.NUMBER, token.RBRACKET, token.EOF}},
		{"...", []token.Type{token.DOTDOT, token.PERIOD, token.EOF}},
	}

	for _, test := range tests {
		lexer := NewLexer(test.input)
		tokens := lexer.TokenizeAll()
		if len(tokens) != len(test.expected) {
			t.Errorf("Input %q: got %d tokens, want %d", test.input, len(tokens), len(test.expected))
			for i, tok := range tokens {
				t.Logf("  Token %d: %v (%q)", i, tok.Type, tok.Value)
			}
			continue
		}
		for i, tok := range tokens {
			if tok.Type != test.expected[i] {
				t.Errorf("Input %q, token %d: got %v, want %v", test.input, i, tok.Type, test.expected[i])
			}
		}
	}
}

func TestParserFunctionCallResult(t *testing.T) {
	input := "Set result to be the result of calling add with 5 and 10."
	program, err := parse(input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	assignment, ok := program.Statements[0].(*ast.Assignment)
	if !ok {
		t.Fatalf("Expected Assignment, got %T", program.Statements[0])
	}

	funcCall, ok := assignment.Value.(*ast.FunctionCall)
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

func TestParserIndexExpression(t *testing.T) {
	tests := []string{
		"Print the item at position 0 in myList.",
		"Print myList[0].",
	}

	for _, input := range tests {
		program, err := parse(input)
		if err != nil {
			t.Errorf("Input %q: parse error: %v", input, err)
			continue
		}

		output, ok := program.Statements[0].(*ast.OutputStatement)
		if !ok {
			t.Errorf("Input %q: expected OutputStatement, got %T", input, program.Statements[0])
			continue
		}

		_, ok = output.Values[0].(*ast.IndexExpression)
		if !ok {
			t.Errorf("Input %q: expected IndexExpression, got %T", input, output.Values[0])
		}
	}
}

func TestParserIndexAssignment(t *testing.T) {
	input := "Set the item at position 0 in myList to be 42."
	program, err := parse(input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	indexAssign, ok := program.Statements[0].(*ast.IndexAssignment)
	if !ok {
		t.Fatalf("Expected IndexAssignment, got %T", program.Statements[0])
	}

	if indexAssign.ListName != "myList" {
		t.Errorf("Expected list name 'myList', got %q", indexAssign.ListName)
	}
}

func TestParserLengthExpression(t *testing.T) {
	input := "Print the length of myList."
	program, err := parse(input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	output, ok := program.Statements[0].(*ast.OutputStatement)
	if !ok {
		t.Fatalf("Expected OutputStatement, got %T", program.Statements[0])
	}

	_, ok = output.Values[0].(*ast.LengthExpression)
	if !ok {
		t.Errorf("Expected LengthExpression, got %T", output.Values[0])
	}
}

func TestParserLocationExpression(t *testing.T) {
	input := "Print the location of x."
	program, err := parse(input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	output, ok := program.Statements[0].(*ast.OutputStatement)
	if !ok {
		t.Fatalf("Expected OutputStatement, got %T", program.Statements[0])
	}

	loc, ok := output.Values[0].(*ast.LocationExpression)
	if !ok {
		t.Errorf("Expected LocationExpression, got %T", output.Values[0])
	}

	if loc.Name != "x" {
		t.Errorf("Expected name 'x', got %q", loc.Name)
	}
}

func TestParserRemainderExpression(t *testing.T) {
	tests := []string{
		"Print the remainder of 17 divided by 5.",
		"Print the remainder of x / y.",
	}

	for _, input := range tests {
		program, err := parse(input)
		if err != nil {
			t.Errorf("Input %q: parse error: %v", input, err)
			continue
		}

		output, ok := program.Statements[0].(*ast.OutputStatement)
		if !ok {
			t.Errorf("Input %q: expected OutputStatement, got %T", input, program.Statements[0])
			continue
		}

		binary, ok := output.Values[0].(*ast.BinaryExpression)
		if !ok {
			t.Errorf("Input %q: expected BinaryExpression, got %T", input, output.Values[0])
			continue
		}

		if binary.Operator != "%" {
			t.Errorf("Input %q: expected operator '%%', got %q", input, binary.Operator)
		}
	}
}

func TestParserBooleanLiterals(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"Declare x to be true.", true},
		{"Declare x to be false.", false},
	}

	for _, test := range tests {
		program, err := parse(test.input)
		if err != nil {
			t.Errorf("Input %q: parse error: %v", test.input, err)
			continue
		}

		varDecl := program.Statements[0].(*ast.VariableDecl)
		boolLit, ok := varDecl.Value.(*ast.BooleanLiteral)
		if !ok {
			t.Errorf("Input %q: expected BooleanLiteral, got %T", test.input, varDecl.Value)
			continue
		}

		if boolLit.Value != test.expected {
			t.Errorf("Input %q: expected %v, got %v", test.input, test.expected, boolLit.Value)
		}
	}
}

func TestParserErrors(t *testing.T) {
	tests := []struct {
		input       string
		expectError bool
	}{
		{"Declare x to be", true},    // Missing value and period
		{"Set x to be 5", true},      // Missing period
		{"If x is equal to 5", true}, // Incomplete if statement
		{"Declare to be 5.", true},   // Missing variable name
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

func TestParserMultipleStatements(t *testing.T) {
	input := `Declare x to be 5.
Declare y to be 10.
Set x to be x + y.`

	program, err := parse(input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	if len(program.Statements) != 3 {
		t.Errorf("Expected 3 statements, got %d", len(program.Statements))
	}
}

func TestParserFunctionCallWithParens(t *testing.T) {
	input := "Print add(5, 10)."
	program, err := parse(input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	output, ok := program.Statements[0].(*ast.OutputStatement)
	if !ok {
		t.Fatalf("Expected OutputStatement, got %T", program.Statements[0])
	}

	funcCall, ok := output.Values[0].(*ast.FunctionCall)
	if !ok {
		t.Errorf("Expected FunctionCall, got %T", output.Values[0])
	}

	if funcCall.Name != "add" {
		t.Errorf("Expected function name 'add', got %q", funcCall.Name)
	}

	if len(funcCall.Arguments) != 2 {
		t.Errorf("Expected 2 arguments, got %d", len(funcCall.Arguments))
	}
}

func TestParserAskStatement(t *testing.T) {
	tests := []struct {
		input   string
		varName string
	}{
		{`Ask "What is your age?" and store it in userAge.`, "userAge"},
		{`Ask "What is your age?" and store the answer in userAge.`, "userAge"},
		{`Ask "What is your age?" and store the result in userAge.`, "userAge"},
		{`Ask "What is your age?" as userAge.`, "userAge"},
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

		assign, ok := program.Statements[0].(*ast.Assignment)
		if !ok {
			t.Errorf("Input %q: expected *ast.Assignment, got %T", test.input, program.Statements[0])
			continue
		}

		if assign.Name != test.varName {
			t.Errorf("Input %q: expected variable name %q, got %q", test.input, test.varName, assign.Name)
		}

		if _, ok := assign.Value.(*ast.AskExpression); !ok {
			t.Errorf("Input %q: expected *ast.AskExpression as value, got %T", test.input, assign.Value)
		}
	}
}

// TestUnterminatedStringLiteral covers the lexer bug where a string with no
// closing quote was accepted silently: readString fell out of its loop at end
// of input and then advanced past the "closing" quote regardless, swallowing
// the rest of the file into the literal instead of reporting an error.
func TestUnterminatedStringLiteral(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"statement position", `Print "hello.`},
		{"expression position", `Declare x to be "unclosed.`},
		{"single quotes", `Print 'hello.`},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := parse(test.input)
			if err == nil {
				t.Fatalf("expected an error for %q, got none", test.input)
			}
			if !strings.Contains(err.Error(), "never closed") {
				t.Errorf("error does not describe an unterminated literal: %v", err)
			}
		})
	}
}

// TestTerminatedStringStillParses guards the fix above against over-reach.
func TestTerminatedStringStillParses(t *testing.T) {
	for _, src := range []string{
		`Print "hello".`,
		`Print 'hello'.`,
		`Print "with \"escaped\" quotes".`,
		`Print "".`,
	} {
		if _, err := parse(src); err != nil {
			t.Errorf("%q should parse, got: %v", src, err)
		}
	}
}

// TestIllegalCharacterReported covers the lexer's ERROR token, which the parser
// previously had no case for and reported with a generic message.
func TestIllegalCharacterReported(t *testing.T) {
	_, err := parse("@ this is nonsense.")
	if err == nil {
		t.Fatal("expected an error for an illegal character, got none")
	}
	if !strings.Contains(err.Error(), "do not recognise the character") {
		t.Errorf("error does not name the illegal character: %v", err)
	}
}

// TestTypeAnnotationForms covers the three annotation positions that used to be
// served by three incompatible parsers. Keyword-named types ("integer",
// "array", "lookup table") were syntax errors in the typed-declaration form
// even though types.Parse accepts them, and a leading article was accepted for
// struct fields but consumed as the type name in declarations.
func TestTypeAnnotationForms(t *testing.T) {
	for _, src := range []string{
		`Declare a as number to be 1.`,
		`Declare a as a number to be 1.`,
		`Declare a as an integer to be 1.`,
		`Declare a as text to be "x".`,
		`Declare a as a text to be "x".`,
		`Declare a as boolean to be true.`,
		`Declare a as integer to be 1.`,
		`Declare a as number.`,
		`Declare a as a number.`,
		`Print 1 cast to text.`,
		`Print "1" cast to number.`,
		`Print 1 cast to a number.`,
	} {
		if _, err := parse(src); err != nil {
			t.Errorf("%q should parse, got: %v", src, err)
		}
	}
}

// TestTypeAnnotationRejectsNonTypes covers the old parseTypeName default branch,
// which accepted *any* token as a type name — so "cast x to 5" parsed cleanly
// and only failed at run time.
func TestTypeAnnotationRejectsNonTypes(t *testing.T) {
	for _, src := range []string{
		`Print 1 cast to 5.`,
		`Print 1 cast to "number".`,
		`Declare y as 42 to be 1.`,
		`Declare y as "number" to be 1.`,
	} {
		_, err := parse(src)
		if err == nil {
			t.Errorf("%q should be rejected, but parsed", src)
			continue
		}
		if !strings.Contains(err.Error(), "type name") {
			t.Errorf("%q: error should mention a type name, got: %v", src, err)
		}
	}
}

// TestBlockEndIsMandatory covers the silent-swallow bug: parseBlock stopped at
// EOF without complaint and every closer in parser.go was optional, so the
// statements after an unclosed block were absorbed into it.
func TestBlockEndIsMandatory(t *testing.T) {
	for _, src := range []string{
		"Declare function f that does the following:\n    Print 1.\nPrint 2.",
		"If true, then\n    Print 1.",
		"Declare x to be 0.\nRepeat the following while x is less than 1:\n    Set x to be 1.",
	} {
		if _, err := parse(src); err == nil {
			t.Errorf("expected an error for an unclosed block:\n%s", src)
		}
	}
}

// TestComparisonsAreFirstClass covers the grammar bug where parseExpression
// started below the relational and logical layers, so a comparison could not
// appear anywhere a value was expected despite boolean being a declared type.
func TestComparisonsAreFirstClass(t *testing.T) {
	for _, src := range []string{
		`Declare x to be 1.
Declare b to be x is greater than 0.`,
		`Declare function f that takes n and does the following:
    Return n is equal to 1.
thats it.`,
		`Declare x to be 1.
Declare b to be x is greater than 0 and x is less than 5.`,
		`Declare x to be 1.
Print x is equal to 1.`,
	} {
		if _, err := parse(src); err != nil {
			t.Errorf("%q should parse, got: %v", src, err)
		}
	}
}

// TestAndStillSeparatesArguments guards the change above against over-reach:
// "and" must remain an argument separator, not become an operator, in call
// argument lists and after an "ask" prompt.
func TestAndStillSeparatesArguments(t *testing.T) {
	prog, err := parse(`Declare function add that takes a and b and does the following:
    Return a + b.
thats it.
Set s to be the result of calling add with 5 and 7.`)
	if err != nil {
		t.Fatalf("should parse, got: %v", err)
	}
	// The last statement must call add with two arguments, not one "5 and 7".
	last := prog.Statements[len(prog.Statements)-1]
	asg, ok := last.(*ast.Assignment)
	if !ok {
		t.Fatalf("last statement is %T, want *ast.Assignment", last)
	}
	call, ok := asg.Value.(*ast.FunctionCall)
	if !ok {
		t.Fatalf("assigned value is %T, want *ast.FunctionCall", asg.Value)
	}
	if len(call.Arguments) != 2 {
		t.Errorf("add was called with %d argument(s), want 2", len(call.Arguments))
	}
}

// TestLogicalPrecedence pins the fix for "and"/"or" sharing one precedence
// level, which made "a or b and c" parse as "(a or b) and c".
func TestLogicalPrecedence(t *testing.T) {
	prog, err := parse(`Declare r to be true or false and false.`)
	if err != nil {
		t.Fatalf("should parse, got: %v", err)
	}
	decl := prog.Statements[0].(*ast.VariableDecl)
	top, ok := decl.Value.(*ast.BinaryExpression)
	if !ok {
		t.Fatalf("value is %T, want *ast.BinaryExpression", decl.Value)
	}
	if top.Operator != "or" {
		t.Errorf("top-level operator is %q, want \"or\" (and binds tighter)", top.Operator)
	}
	right, ok := top.Right.(*ast.BinaryExpression)
	if !ok {
		t.Fatalf("right operand is %T, want *ast.BinaryExpression", top.Right)
	}
	if right.Operator != "and" {
		t.Errorf("right operand operator is %q, want \"and\"", right.Operator)
	}
}

// TestNotPrecedence pins the fix for "not" living in parsePrimary, which bound
// it tighter than comparison so "not x is equal to y" meant "(not x) is equal
// to y".
func TestNotPrecedence(t *testing.T) {
	prog, err := parse(`Declare x to be 1.
Declare r to be not x is equal to 2.`)
	if err != nil {
		t.Fatalf("should parse, got: %v", err)
	}
	decl := prog.Statements[1].(*ast.VariableDecl)
	un, ok := decl.Value.(*ast.UnaryExpression)
	if !ok {
		t.Fatalf("value is %T, want *ast.UnaryExpression (not should be outermost)", decl.Value)
	}
	if un.Operator != "not" {
		t.Errorf("operator is %q, want \"not\"", un.Operator)
	}
	if _, ok := un.Right.(*ast.BinaryExpression); !ok {
		t.Errorf("operand of not is %T, want the comparison", un.Right)
	}
}
