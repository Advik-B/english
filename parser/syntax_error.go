package parser

import (
	"errors"
	"fmt"
	"strings"

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
	// Truncated reports that the parser ran out of input rather than finding
	// something it did not expect: the text so far is the beginning of
	// something valid. An interactive prompt reads this to decide between
	// asking for another line and reporting a mistake.
	Truncated bool
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

// markTruncated records whether the parser stopped because the input ended
// part-way through a block.
//
// Two things have to hold. The parser returns without advancing when it fails,
// so the current token being the end of input means it ran out of tokens
// rather than finding the wrong one. And a block body has to have reached that
// end unclosed, which is what parseBlock records: without it, a plainly
// incomplete statement like "Declare x to be" would read as "more to come"
// and an interactive prompt would wait instead of reporting it.
//
// Stamping this in one place keeps every error site from having to know.
func (p *Parser) markTruncated(err error) error {
	var syntaxErr *SyntaxError
	if errors.As(err, &syntaxErr) && p.curToken.Type == token.EOF && p.blockAtEOF {
		syntaxErr.Truncated = true
	}
	return err
}

// SyntaxErrors is every syntax error one parse found, reported together.
//
// The parser used to stop at the first, so a file with two typos took two runs
// to fix, and the editor — which shows one diagnostic per parse and then gives
// up before extracting any symbols — told you nothing else about a file until
// its last syntax error was gone.
type SyntaxErrors []*SyntaxError

func (e SyntaxErrors) Error() string {
	parts := make([]string, 0, len(e))
	for _, err := range e {
		parts = append(parts, err.Error())
	}
	return strings.Join(parts, "\n\n")
}

// Unwrap returns the first error, so that anything asking "is this a syntax
// error, and where?" keeps working unchanged.
func (e SyntaxErrors) Unwrap() error {
	if len(e) == 0 {
		return nil
	}
	return e[0]
}

// Errors returns every syntax error behind a parse failure, which is one error
// for most callers and the whole set for anything that shows them all.
func Errors(err error) []*SyntaxError {
	var many SyntaxErrors
	if errors.As(err, &many) {
		return many
	}
	var one *SyntaxError
	if errors.As(err, &one) {
		return []*SyntaxError{one}
	}
	return nil
}

// IsTruncated reports whether a parse failed only because the input ended
// part-way through something valid.
//
// This is what an interactive prompt needs in order to decide between asking
// for another line and reporting a mistake. The REPL used to answer it with a
// text search for "thats it." and a check for a line ending in "then", so
// printing the text "thats it." inside a loop ended the block, a comment
// mentioning "do the following:" opened one, and neither could be told from
// the real thing.
func IsTruncated(err error) bool {
	var syntaxErr *SyntaxError
	return errors.As(err, &syntaxErr) && syntaxErr.Truncated
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
