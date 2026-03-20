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
	AS
	STRUCTURE
	STRUCT
	FIELDS
	FIELD
	INSTANCE
	NEW
	TRY
	DOING
	ON
	ONERROR
	FINALLY
	RAISE
	REFERENCE
	COPY
	SWAP
	CASTED
	TYPE
	WHICH
	IS
	FROM
	UNSIGNED
	INTEGER
	DEFAULT
	BUT
	IMPORT
	EVERYTHING
	ALL
	SAFELY

	// New keywords
	CONTINUE
	SKIP
	NOTHING
	NOT
	OR
	ASK
	ARRAY
	LOOKUP
	TABLE
	HAS
	ENTRY
	RANGE

	// PLEASE is emitted for politeness prefixes: "please", "kindly",
	// "could you", and "would you kindly". When --minimum-politeness is
	// active the parser counts PLEASE-prefixed statements toward the
	// politeness quota.
	PLEASE

	// SLEEP is emitted for the "sleep" and "wait" keywords used in
	// "Sleep for <duration>." / "Wait for <duration>." statements.
	SLEEP

	// WHITESPACE is emitted by TokenizeForHighlight to represent horizontal
	// whitespace (spaces / tabs) that was skipped by the lexer between tokens.
	// It is never produced by NextToken or TokenizeAll; it exists solely for
	// reconstructing the original source text during syntax highlighting.
	WHITESPACE

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
	DOTDOT // ".." range operator

	// Comparison operators (multi-word)
	IS_EQUAL_TO
	IS_LESS_THAN
	IS_GREATER_THAN
	IS_LESS_EQUAL
	IS_GREATER_EQUAL
	IS_NOT_EQUAL
	IS_SOMETHING  // "is something" / "has a value" — not-nil check
	IS_NOTHING_OP // "is nothing" / "has no value" — nil check
	IS_TRUE       // "is true"  — boolean true check
	IS_FALSE      // "is false" — boolean false check
	ISNT_TRUE     // "isn't true"  — boolean not-true check
	ISNT_FALSE    // "isn't false" — boolean not-false check
	POSSESSIVE    // standalone 's — postfix possessive operator (e.g. "hello"'s title)
	COMMENT       // # … — source comment; Value holds the text after '#' (trimmed)
)

// Token represents a single token from the lexer
type Token struct {
	Type  Type
	Value string
	Line  int
	Col   int
	Pos   int // byte offset of the token's first character in the source string
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
	case AS:
		return "AS"
	case STRUCTURE:
		return "STRUCTURE"
	case STRUCT:
		return "STRUCT"
	case FIELDS:
		return "FIELDS"
	case FIELD:
		return "FIELD"
	case INSTANCE:
		return "INSTANCE"
	case NEW:
		return "NEW"
	case TRY:
		return "TRY"
	case DOING:
		return "DOING"
	case ON:
		return "ON"
	case ONERROR:
		return "ONERROR"
	case FINALLY:
		return "FINALLY"
	case RAISE:
		return "RAISE"
	case REFERENCE:
		return "REFERENCE"
	case COPY:
		return "COPY"
	case SWAP:
		return "SWAP"
	case CASTED:
		return "CASTED"
	case TYPE:
		return "TYPE"
	case WHICH:
		return "WHICH"
	case IS:
		return "IS"
	case FROM:
		return "FROM"
	case UNSIGNED:
		return "UNSIGNED"
	case INTEGER:
		return "INTEGER"
	case DEFAULT:
		return "DEFAULT"
	case BUT:
		return "BUT"
	case IMPORT:
		return "IMPORT"
	case EVERYTHING:
		return "EVERYTHING"
	case ALL:
		return "ALL"
	case SAFELY:
		return "SAFELY"
	case CONTINUE:
		return "CONTINUE"
	case SKIP:
		return "SKIP"
	case NOTHING:
		return "NOTHING"
	case NOT:
		return "NOT"
	case OR:
		return "OR"
	case ASK:
		return "ASK"
	case ARRAY:
		return "ARRAY"
	case LOOKUP:
		return "LOOKUP"
	case TABLE:
		return "TABLE"
	case HAS:
		return "HAS"
	case ENTRY:
		return "ENTRY"
	case RANGE:
		return "RANGE"
	case PLEASE:
		return "PLEASE"
	case SLEEP:
		return "SLEEP"
	case WHITESPACE:
		return "WHITESPACE"
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
	case DOTDOT:
		return "DOTDOT"
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
	case IS_SOMETHING:
		return "IS_SOMETHING"
	case IS_NOTHING_OP:
		return "IS_NOTHING_OP"
	case IS_TRUE:
		return "IS_TRUE"
	case IS_FALSE:
		return "IS_FALSE"
	case ISNT_TRUE:
		return "ISNT_TRUE"
	case ISNT_FALSE:
		return "ISNT_FALSE"
	case POSSESSIVE:
		return "POSSESSIVE"
	case COMMENT:
		return "COMMENT"
	default:
		return "UNKNOWN"
	}
}
