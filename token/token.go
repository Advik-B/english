// Package token defines the token types used by the lexer and parser
// for the English programming language.
package token

// Type represents different token types in the English language
type Type int

const (
	// Special tokens
	EOF Type = iota
	ERROR
	NEWLINE

	// Literals
	NUMBER
	STRING
	IDENTIFIER

	// Keywords
	DECLARE
	LET
	EQUAL
	FUNCTION
	THAT
	DOES
	FOLLOWING
	THATS
	IT
	TO
	BE
	ALWAYS
	SET
	CALL
	RETURN
	PRINT
	IF
	THEN
	OTHERWISE
	REPEAT
	WHILE
	FOREVER
	BREAK
	OUT
	LOOP
	TIMES
	FOR
	EACH
	IN
	DO
	TAKES
	AND
	WITH
	THE
	OF
	CALLING
	VALUE
	ITEM
	AT
	POSITION
	LENGTH
	REMAINDER
	DIVIDED
	BY
	TRUE
	FALSE
	TOGGLE
	LOCATION
	WRITE

	// Operators and Punctuation
	PERIOD
	COMMA
	COLON
	LPAREN
	RPAREN
	LBRACKET
	RBRACKET
	PLUS
	MINUS
	STAR
	SLASH
	ASSIGN

	// Comparison operators (multi-word)
	IS_EQUAL_TO
	IS_LESS_THAN
	IS_GREATER_THAN
	IS_LESS_EQUAL
	IS_GREATER_EQUAL
	IS_NOT_EQUAL
)

// Token represents a single token from the lexer
type Token struct {
	Type  Type
	Value string
	Line  int
	Col   int
}

// String representation of token type
func (t Type) String() string {
	switch t {
	case EOF:
		return "EOF"
	case ERROR:
		return "ERROR"
	case NEWLINE:
		return "NEWLINE"
	case NUMBER:
		return "NUMBER"
	case STRING:
		return "STRING"
	case IDENTIFIER:
		return "IDENTIFIER"
	case DECLARE:
		return "DECLARE"
	case LET:
		return "LET"
	case EQUAL:
		return "EQUAL"
	case FUNCTION:
		return "FUNCTION"
	case THAT:
		return "THAT"
	case DOES:
		return "DOES"
	case FOLLOWING:
		return "FOLLOWING"
	case THATS:
		return "THATS"
	case IT:
		return "IT"
	case TO:
		return "TO"
	case BE:
		return "BE"
	case ALWAYS:
		return "ALWAYS"
	case SET:
		return "SET"
	case CALL:
		return "CALL"
	case RETURN:
		return "RETURN"
	case PRINT:
		return "PRINT"
	case IF:
		return "IF"
	case THEN:
		return "THEN"
	case OTHERWISE:
		return "OTHERWISE"
	case REPEAT:
		return "REPEAT"
	case WHILE:
		return "WHILE"
	case FOREVER:
		return "FOREVER"
	case BREAK:
		return "BREAK"
	case OUT:
		return "OUT"
	case LOOP:
		return "LOOP"
	case TIMES:
		return "TIMES"
	case FOR:
		return "FOR"
	case EACH:
		return "EACH"
	case IN:
		return "IN"
	case DO:
		return "DO"
	case TAKES:
		return "TAKES"
	case AND:
		return "AND"
	case WITH:
		return "WITH"
	case THE:
		return "THE"
	case OF:
		return "OF"
	case CALLING:
		return "CALLING"
	case VALUE:
		return "VALUE"
	case ITEM:
		return "ITEM"
	case AT:
		return "AT"
	case POSITION:
		return "POSITION"
	case LENGTH:
		return "LENGTH"
	case REMAINDER:
		return "REMAINDER"
	case DIVIDED:
		return "DIVIDED"
	case BY:
		return "BY"
	case TRUE:
		return "TRUE"
	case FALSE:
		return "FALSE"
	case TOGGLE:
		return "TOGGLE"
	case LOCATION:
		return "LOCATION"
	case WRITE:
		return "WRITE"
	case PERIOD:
		return "PERIOD"
	case COMMA:
		return "COMMA"
	case COLON:
		return "COLON"
	case LPAREN:
		return "LPAREN"
	case RPAREN:
		return "RPAREN"
	case LBRACKET:
		return "LBRACKET"
	case RBRACKET:
		return "RBRACKET"
	case PLUS:
		return "PLUS"
	case MINUS:
		return "MINUS"
	case STAR:
		return "STAR"
	case SLASH:
		return "SLASH"
	case ASSIGN:
		return "ASSIGN"
	case IS_EQUAL_TO:
		return "IS_EQUAL_TO"
	case IS_LESS_THAN:
		return "IS_LESS_THAN"
	case IS_GREATER_THAN:
		return "IS_GREATER_THAN"
	case IS_LESS_EQUAL:
		return "IS_LESS_EQUAL"
	case IS_GREATER_EQUAL:
		return "IS_GREATER_EQUAL"
	case IS_NOT_EQUAL:
		return "IS_NOT_EQUAL"
	default:
		return "UNKNOWN"
	}
}
