package parser

import (
	"english/token"
	"fmt"
)

// SyntaxError is a structured parse-time error.
// It satisfies the stacktraces.SyntaxError interface so the renderer can
// display it with a dedicated "Syntax Error" header, line/column information,
// and an optional user-friendly hint.
type SyntaxError struct {
	Msg  string // human-readable description of what went wrong
	Line int    // 1-based source line (0 = unknown)
	Col  int    // 1-based source column (0 = unknown)
	Hint string // optional guidance for the programmer
}

// Error implements the standard error interface.
func (e *SyntaxError) Error() string {
	if e.Line > 0 {
		if e.Hint != "" {
			return fmt.Sprintf("Syntax Error at line %d, column %d: %s\nHint: %s", e.Line, e.Col, e.Msg, e.Hint)
		}
		return fmt.Sprintf("Syntax Error at line %d, column %d: %s", e.Line, e.Col, e.Msg)
	}
	return "Syntax Error: " + e.Msg
}

// SyntaxMessage implements stacktraces.SyntaxError.
func (e *SyntaxError) SyntaxMessage() string { return e.Msg }

// SyntaxLine implements stacktraces.SyntaxError.
func (e *SyntaxError) SyntaxLine() int { return e.Line }

// SyntaxCol implements stacktraces.SyntaxError.
func (e *SyntaxError) SyntaxCol() int { return e.Col }

// SyntaxHint implements stacktraces.SyntaxError.
func (e *SyntaxError) SyntaxHint() string { return e.Hint }

// syntaxErr is a convenience constructor used inside the parser.
func (p *Parser) syntaxErr(msg string, hint string) *SyntaxError {
	return &SyntaxError{
		Msg:  msg,
		Line: p.curToken.Line,
		Col:  p.curToken.Col,
		Hint: hint,
	}
}

// tokenFriendlyName returns a human-readable name for the expected token type.
func tokenFriendlyName(t token.Type) string {
	switch t {
	case token.PERIOD:
		return "a period (.)"
	case token.COMMA:
		return "a comma (,)"
	case token.COLON:
		return "a colon (:)"
	case token.IDENTIFIER:
		return "a name"
	case token.NUMBER:
		return "a number"
	case token.STRING:
		return "some text (in quotes)"
	case token.BE:
		return "the word 'be'"
	case token.TO:
		return "the word 'to'"
	case token.THATS:
		return "the word 'thats'"
	case token.IT:
		return "the word 'it'"
	case token.FUNCTION:
		return "the word 'function'"
	case token.DOES:
		return "the word 'does'"
	case token.FOLLOWING:
		return "the word 'following'"
	case token.TIMES:
		return "the word 'times'"
	case token.IN:
		return "the word 'in'"
	case token.AND:
		return "the word 'and'"
	case token.WITH:
		return "the word 'with'"
	case token.THE:
		return "the word 'the'"
	case token.OF:
		return "the word 'of'"
	case token.AT:
		return "the word 'at'"
	case token.POSITION:
		return "the word 'position'"
	case token.ITEM:
		return "the word 'item'"
	case token.DOING:
		return "the word 'doing'"
	case token.WHILE:
		return "the word 'while'"
	case token.RETURN:
		return "the word 'return'"
	case token.IMPORT:
		return "the word 'import'"
	case token.DECLARE:
		return "the word 'declare'"
	case token.SET:
		return "the word 'set'"
	case token.PRINT:
		return "the word 'print'"
	case token.CALL:
		return "the word 'call'"
	case token.IF:
		return "the word 'if'"
	case token.REPEAT:
		return "the word 'repeat'"
	case token.LBRACKET:
		return "an opening bracket ([)"
	case token.RBRACKET:
		return "a closing bracket (])"
	case token.LPAREN:
		return "an opening parenthesis (()"
	case token.RPAREN:
		return "a closing parenthesis ())"
	default:
		return fmt.Sprintf("'%s'", t)
	}
}

// tokenFriendlyValue returns a human-readable description of a token type + value.
func tokenFriendlyValue(t token.Type, value string) string {
	switch t {
	case token.IDENTIFIER:
		return fmt.Sprintf("the name '%s'", value)
	case token.NUMBER:
		return fmt.Sprintf("the number %s", value)
	case token.STRING:
		return fmt.Sprintf("the text %q", value)
	case token.EOF:
		return "the end of the file"
	case token.NEWLINE:
		return "a new line"
	case token.PERIOD:
		return "a period (.)"
	case token.COMMA:
		return "a comma (,)"
	case token.COLON:
		return "a colon (:)"
	default:
		if value != "" && value != t.String() {
			return fmt.Sprintf("'%s'", value)
		}
		return fmt.Sprintf("'%s'", t)
	}
}
