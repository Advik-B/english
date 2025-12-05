package interpreter

import "testing"

func TestTokenTypeString(t *testing.T) {
	tests := []struct {
		tokenType TokenType
		expected  string
	}{
		{TOKEN_EOF, "EOF"},
		{TOKEN_ERROR, "ERROR"},
		{TOKEN_NUMBER, "NUMBER"},
		{TOKEN_STRING, "STRING"},
		{TOKEN_IDENTIFIER, "IDENTIFIER"},
		{TOKEN_DECLARE, "DECLARE"},
		{TOKEN_FUNCTION, "FUNCTION"},
		{TOKEN_SET, "SET"},
		{TOKEN_SAY, "SAY"},
		{TOKEN_IF, "IF"},
		{TOKEN_THEN, "THEN"},
		{TOKEN_OTHERWISE, "OTHERWISE"},
		{TOKEN_REPEAT, "REPEAT"},
		{TOKEN_WHILE, "WHILE"},
		{TOKEN_FOR, "FOR"},
		{TOKEN_RETURN, "RETURN"},
		{TOKEN_IS_EQUAL_TO, "IS_EQUAL_TO"},
		{TOKEN_IS_LESS_THAN, "IS_LESS_THAN"},
		{TOKEN_IS_GREATER_THAN, "IS_GREATER_THAN"},
		{TOKEN_PERIOD, "PERIOD"},
		{TOKEN_COMMA, "COMMA"},
		{TOKEN_COLON, "COLON"},
		{TOKEN_PLUS, "PLUS"},
		{TOKEN_MINUS, "MINUS"},
		{TOKEN_STAR, "STAR"},
		{TOKEN_SLASH, "SLASH"},
	}

	for _, test := range tests {
		result := test.tokenType.String()
		if result != test.expected {
			t.Errorf("TokenType(%d).String() = %q, want %q", test.tokenType, result, test.expected)
		}
	}
}

func TestUnknownTokenType(t *testing.T) {
	unknownType := TokenType(999)
	result := unknownType.String()
	if result != "UNKNOWN" {
		t.Errorf("Unknown TokenType.String() = %q, want %q", result, "UNKNOWN")
	}
}
