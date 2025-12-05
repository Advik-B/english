package token

import "testing"

func TestTypeString(t *testing.T) {
	tests := []struct {
		tokenType Type
		expected  string
	}{
		{EOF, "EOF"},
		{ERROR, "ERROR"},
		{NUMBER, "NUMBER"},
		{STRING, "STRING"},
		{IDENTIFIER, "IDENTIFIER"},
		{DECLARE, "DECLARE"},
		{FUNCTION, "FUNCTION"},
		{SET, "SET"},
		{PRINT, "PRINT"},
		{IF, "IF"},
		{THEN, "THEN"},
		{OTHERWISE, "OTHERWISE"},
		{REPEAT, "REPEAT"},
		{WHILE, "WHILE"},
		{FOR, "FOR"},
		{RETURN, "RETURN"},
		{IS_EQUAL_TO, "IS_EQUAL_TO"},
		{IS_LESS_THAN, "IS_LESS_THAN"},
		{IS_GREATER_THAN, "IS_GREATER_THAN"},
		{PERIOD, "PERIOD"},
		{COMMA, "COMMA"},
		{COLON, "COLON"},
		{PLUS, "PLUS"},
		{MINUS, "MINUS"},
		{STAR, "STAR"},
		{SLASH, "SLASH"},
		{TRUE, "TRUE"},
		{FALSE, "FALSE"},
		{TOGGLE, "TOGGLE"},
		{LOCATION, "LOCATION"},
		{NEWLINE, "NEWLINE"},
		{THAT, "THAT"},
		{DOES, "DOES"},
		{FOLLOWING, "FOLLOWING"},
		{THATS, "THATS"},
		{IT, "IT"},
		{TO, "TO"},
		{BE, "BE"},
		{ALWAYS, "ALWAYS"},
		{CALL, "CALL"},
		{TIMES, "TIMES"},
		{EACH, "EACH"},
		{IN, "IN"},
		{DO, "DO"},
		{TAKES, "TAKES"},
		{AND, "AND"},
		{WITH, "WITH"},
		{THE, "THE"},
		{OF, "OF"},
		{CALLING, "CALLING"},
		{VALUE, "VALUE"},
		{ITEM, "ITEM"},
		{AT, "AT"},
		{POSITION, "POSITION"},
		{LENGTH, "LENGTH"},
		{REMAINDER, "REMAINDER"},
		{DIVIDED, "DIVIDED"},
		{BY, "BY"},
		{LPAREN, "LPAREN"},
		{RPAREN, "RPAREN"},
		{LBRACKET, "LBRACKET"},
		{RBRACKET, "RBRACKET"},
		{IS_LESS_EQUAL, "IS_LESS_EQUAL"},
		{IS_GREATER_EQUAL, "IS_GREATER_EQUAL"},
		{IS_NOT_EQUAL, "IS_NOT_EQUAL"},
	}

	for _, test := range tests {
		result := test.tokenType.String()
		if result != test.expected {
			t.Errorf("Type(%d).String() = %q, want %q", test.tokenType, result, test.expected)
		}
	}
}

func TestUnknownType(t *testing.T) {
	unknownType := Type(999)
	result := unknownType.String()
	if result != "UNKNOWN" {
		t.Errorf("Unknown Type.String() = %q, want %q", result, "UNKNOWN")
	}
}

func TestTokenStruct(t *testing.T) {
	tok := Token{
		Type:  NUMBER,
		Value: "42",
		Line:  1,
		Col:   5,
	}

	if tok.Type != NUMBER {
		t.Errorf("Token.Type = %v, want NUMBER", tok.Type)
	}
	if tok.Value != "42" {
		t.Errorf("Token.Value = %q, want \"42\"", tok.Value)
	}
	if tok.Line != 1 {
		t.Errorf("Token.Line = %d, want 1", tok.Line)
	}
	if tok.Col != 5 {
		t.Errorf("Token.Col = %d, want 5", tok.Col)
	}
}
