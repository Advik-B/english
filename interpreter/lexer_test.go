package interpreter

import (
	"testing"
)

func TestLexerBasicTokens(t *testing.T) {
	tests := []struct {
		input    string
		expected []TokenType
	}{
		{".", []TokenType{TOKEN_PERIOD, TOKEN_EOF}},
		{",", []TokenType{TOKEN_COMMA, TOKEN_EOF}},
		{":", []TokenType{TOKEN_COLON, TOKEN_EOF}},
		{"(", []TokenType{TOKEN_LPAREN, TOKEN_EOF}},
		{")", []TokenType{TOKEN_RPAREN, TOKEN_EOF}},
		{"[", []TokenType{TOKEN_LBRACKET, TOKEN_EOF}},
		{"]", []TokenType{TOKEN_RBRACKET, TOKEN_EOF}},
		{"+", []TokenType{TOKEN_PLUS, TOKEN_EOF}},
		{"-", []TokenType{TOKEN_MINUS, TOKEN_EOF}},
		{"*", []TokenType{TOKEN_STAR, TOKEN_EOF}},
		{"/", []TokenType{TOKEN_SLASH, TOKEN_EOF}},
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
	}

	for _, test := range tests {
		lexer := NewLexer(test.input)
		tok := lexer.NextToken()
		if tok.Type != TOKEN_NUMBER {
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
	}

	for _, test := range tests {
		lexer := NewLexer(test.input)
		tok := lexer.NextToken()
		if tok.Type != TOKEN_STRING {
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
		expected TokenType
	}{
		{"declare", TOKEN_DECLARE},
		{"DECLARE", TOKEN_DECLARE},
		{"Declare", TOKEN_DECLARE},
		{"function", TOKEN_FUNCTION},
		{"set", TOKEN_SET},
		{"say", TOKEN_SAY},
		{"if", TOKEN_IF},
		{"then", TOKEN_THEN},
		{"otherwise", TOKEN_OTHERWISE},
		{"repeat", TOKEN_REPEAT},
		{"while", TOKEN_WHILE},
		{"for", TOKEN_FOR},
		{"each", TOKEN_EACH},
		{"return", TOKEN_RETURN},
		{"to", TOKEN_TO},
		{"be", TOKEN_BE},
		{"always", TOKEN_ALWAYS},
		{"call", TOKEN_CALL},
		{"the", TOKEN_THE},
		{"value", TOKEN_VALUE},
		{"of", TOKEN_OF},
		{"with", TOKEN_WITH},
		{"and", TOKEN_AND},
		{"takes", TOKEN_TAKES},
		{"does", TOKEN_DOES},
		{"following", TOKEN_FOLLOWING},
		{"thats", TOKEN_THATS},
		{"it", TOKEN_IT},
		{"times", TOKEN_TIMES},
		{"in", TOKEN_IN},
		{"do", TOKEN_DO},
		{"calling", TOKEN_CALLING},
		{"that", TOKEN_THAT},
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
	}

	for _, input := range tests {
		lexer := NewLexer(input)
		tok := lexer.NextToken()
		if tok.Type != TOKEN_IDENTIFIER {
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
		expected TokenType
	}{
		{"is equal to", TOKEN_IS_EQUAL_TO},
		{"IS EQUAL TO", TOKEN_IS_EQUAL_TO},
		{"Is Equal To", TOKEN_IS_EQUAL_TO},
		{"is less than", TOKEN_IS_LESS_THAN},
		{"IS LESS THAN", TOKEN_IS_LESS_THAN},
		{"is greater than", TOKEN_IS_GREATER_THAN},
		{"IS GREATER THAN", TOKEN_IS_GREATER_THAN},
		{"is less than or equal to", TOKEN_IS_LESS_EQUAL},
		{"is greater than or equal to", TOKEN_IS_GREATER_EQUAL},
		{"is not equal to", TOKEN_IS_NOT_EQUAL},
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
	if tokens[0].Type != TOKEN_DECLARE {
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
	// The column is tracked after reading the token
	if tok.Col < 1 {
		t.Errorf("Expected col >= 1, got %d", tok.Col)
	}
}

func TestLexerCompleteStatement(t *testing.T) {
	input := "Declare x to be 5."
	expected := []TokenType{
		TOKEN_DECLARE, TOKEN_IDENTIFIER, TOKEN_TO, TOKEN_BE, TOKEN_NUMBER, TOKEN_PERIOD, TOKEN_EOF,
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
	expected := []TokenType{
		TOKEN_DECLARE, TOKEN_IDENTIFIER, TOKEN_TO, TOKEN_BE, TOKEN_NUMBER, TOKEN_PERIOD, TOKEN_EOF,
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
