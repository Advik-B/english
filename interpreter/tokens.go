package interpreter

// TokenType represents different token types in the English language
type TokenType int

const (
	// Special tokens
	TOKEN_EOF TokenType = iota
	TOKEN_ERROR
	TOKEN_NEWLINE

	// Literals
	TOKEN_NUMBER
	TOKEN_STRING
	TOKEN_IDENTIFIER

	// Keywords
	TOKEN_DECLARE
	TOKEN_FUNCTION
	TOKEN_THAT
	TOKEN_DOES
	TOKEN_FOLLOWING
	TOKEN_THATS
	TOKEN_IT
	TOKEN_TO
	TOKEN_BE
	TOKEN_ALWAYS
	TOKEN_SET
	TOKEN_CALL
	TOKEN_RETURN
	TOKEN_PRINT
	TOKEN_IF
	TOKEN_THEN
	TOKEN_OTHERWISE
	TOKEN_REPEAT
	TOKEN_WHILE
	TOKEN_TIMES
	TOKEN_FOR
	TOKEN_EACH
	TOKEN_IN
	TOKEN_DO
	TOKEN_TAKES
	TOKEN_AND
	TOKEN_WITH
	TOKEN_THE
	TOKEN_OF
	TOKEN_CALLING
	TOKEN_VALUE
	TOKEN_ITEM
	TOKEN_AT
	TOKEN_POSITION
	TOKEN_LENGTH
	TOKEN_REMAINDER
	TOKEN_DIVIDED
	TOKEN_BY
	TOKEN_TRUE
	TOKEN_FALSE
	TOKEN_TOGGLE
	TOKEN_LOCATION

	// Operators and Punctuation
	TOKEN_PERIOD
	TOKEN_COMMA
	TOKEN_COLON
	TOKEN_LPAREN
	TOKEN_RPAREN
	TOKEN_LBRACKET
	TOKEN_RBRACKET
	TOKEN_PLUS
	TOKEN_MINUS
	TOKEN_STAR
	TOKEN_SLASH

	// Comparison operators (multi-word)
	TOKEN_IS_EQUAL_TO
	TOKEN_IS_LESS_THAN
	TOKEN_IS_GREATER_THAN
	TOKEN_IS_LESS_EQUAL
	TOKEN_IS_GREATER_EQUAL
	TOKEN_IS_NOT_EQUAL
)

// Token represents a single token from the lexer
type Token struct {
	Type  TokenType
	Value string
	Line  int
	Col   int
}

// String representation of token type
func (t TokenType) String() string {
	switch t {
	case TOKEN_EOF:
		return "EOF"
	case TOKEN_ERROR:
		return "ERROR"
	case TOKEN_NEWLINE:
		return "NEWLINE"
	case TOKEN_NUMBER:
		return "NUMBER"
	case TOKEN_STRING:
		return "STRING"
	case TOKEN_IDENTIFIER:
		return "IDENTIFIER"
	case TOKEN_DECLARE:
		return "DECLARE"
	case TOKEN_FUNCTION:
		return "FUNCTION"
	case TOKEN_THAT:
		return "THAT"
	case TOKEN_DOES:
		return "DOES"
	case TOKEN_FOLLOWING:
		return "FOLLOWING"
	case TOKEN_THATS:
		return "THATS"
	case TOKEN_IT:
		return "IT"
	case TOKEN_TO:
		return "TO"
	case TOKEN_BE:
		return "BE"
	case TOKEN_ALWAYS:
		return "ALWAYS"
	case TOKEN_SET:
		return "SET"
	case TOKEN_CALL:
		return "CALL"
	case TOKEN_RETURN:
		return "RETURN"
	case TOKEN_PRINT:
		return "PRINT"
	case TOKEN_IF:
		return "IF"
	case TOKEN_THEN:
		return "THEN"
	case TOKEN_OTHERWISE:
		return "OTHERWISE"
	case TOKEN_REPEAT:
		return "REPEAT"
	case TOKEN_WHILE:
		return "WHILE"
	case TOKEN_TIMES:
		return "TIMES"
	case TOKEN_FOR:
		return "FOR"
	case TOKEN_EACH:
		return "EACH"
	case TOKEN_IN:
		return "IN"
	case TOKEN_DO:
		return "DO"
	case TOKEN_TAKES:
		return "TAKES"
	case TOKEN_AND:
		return "AND"
	case TOKEN_WITH:
		return "WITH"
	case TOKEN_THE:
		return "THE"
	case TOKEN_OF:
		return "OF"
	case TOKEN_CALLING:
		return "CALLING"
	case TOKEN_VALUE:
		return "VALUE"
	case TOKEN_ITEM:
		return "ITEM"
	case TOKEN_AT:
		return "AT"
	case TOKEN_POSITION:
		return "POSITION"
	case TOKEN_LENGTH:
		return "LENGTH"
	case TOKEN_REMAINDER:
		return "REMAINDER"
	case TOKEN_DIVIDED:
		return "DIVIDED"
	case TOKEN_BY:
		return "BY"
	case TOKEN_TRUE:
		return "TRUE"
	case TOKEN_FALSE:
		return "FALSE"
	case TOKEN_TOGGLE:
		return "TOGGLE"
	case TOKEN_LOCATION:
		return "LOCATION"
	case TOKEN_PERIOD:
		return "PERIOD"
	case TOKEN_COMMA:
		return "COMMA"
	case TOKEN_COLON:
		return "COLON"
	case TOKEN_LPAREN:
		return "LPAREN"
	case TOKEN_RPAREN:
		return "RPAREN"
	case TOKEN_LBRACKET:
		return "LBRACKET"
	case TOKEN_RBRACKET:
		return "RBRACKET"
	case TOKEN_PLUS:
		return "PLUS"
	case TOKEN_MINUS:
		return "MINUS"
	case TOKEN_STAR:
		return "STAR"
	case TOKEN_SLASH:
		return "SLASH"
	case TOKEN_IS_EQUAL_TO:
		return "IS_EQUAL_TO"
	case TOKEN_IS_LESS_THAN:
		return "IS_LESS_THAN"
	case TOKEN_IS_GREATER_THAN:
		return "IS_GREATER_THAN"
	case TOKEN_IS_LESS_EQUAL:
		return "IS_LESS_EQUAL"
	case TOKEN_IS_GREATER_EQUAL:
		return "IS_GREATER_EQUAL"
	case TOKEN_IS_NOT_EQUAL:
		return "IS_NOT_EQUAL"
	default:
		return "UNKNOWN"
	}
}
