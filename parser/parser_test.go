package parser

import (
	"english/ast"
	"english/token"
	"testing"
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

	// Should have DECLARE and EOF (comment is skipped)
	if len(tokens) != 2 {
		t.Errorf("Expected 2 tokens, got %d", len(tokens))
	}
	if tokens[0].Type != token.DECLARE {
		t.Errorf("First token should be DECLARE, got %v", tokens[0].Type)
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
