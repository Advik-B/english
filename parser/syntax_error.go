package parser

import (
	"fmt"

	"github.com/Advik-B/english/token"
	"github.com/Advik-B/english/tokeniser"
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
// tokenFriendlyName renders a token type the way a user would read it.
//
// Punctuation and literals get hand-written wording; every keyword and
// multi-word operator is derived from the lexer's own spelling table, so a
// token can never again fall through to its Go constant name (a missing "then"
// used to produce "I expected 'THEN' here").
func tokenFriendlyName(t token.Type) string {
	switch t {
	case token.PERIOD:
		return "a period (.)"
	case token.COMMA:
		return "a comma (,)"
	case token.COLON:
		return "a colon (:)"
	case token.LBRACKET:
		return "an opening bracket ([)"
	case token.RBRACKET:
		return "a closing bracket (])"
	case token.LPAREN:
		return "an opening parenthesis (()"
	case token.RPAREN:
		return "a closing parenthesis ())"
	case token.PLUS:
		return "a plus sign (+)"
	case token.MINUS:
		return "a minus sign (-)"
	case token.STAR:
		return "a multiplication sign (*)"
	case token.SLASH:
		return "a division sign (/)"
	case token.ASSIGN:
		return "an equals sign (=)"
	case token.DOTDOT:
		return "a range operator (..)"
	case token.IDENTIFIER:
		return "a name"
	case token.NUMBER:
		return "a number"
	case token.STRING:
		return "some text (in quotes)"
	case token.COMMENT:
		return "a comment"
	case token.NEWLINE:
		return "a new line"
	case token.EOF:
		return "the end of the file"
	case token.ERROR:
		return "an unrecognised character"
	case token.WHITESPACE:
		return "whitespace"
	}
	if word, ok := tokeniser.Spelling(t); ok {
		if token.IsKeyword(t) {
			return fmt.Sprintf("the word '%s'", word)
		}
		return fmt.Sprintf("'%s'", word)
	}
	return fmt.Sprintf("'%s'", t)
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
