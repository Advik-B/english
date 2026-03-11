// Package highlight provides syntax highlighting for English source code.
//
// It exposes two main entry points:
//   - Highlight(source, useColor)       – colour an entire source file
//   - HighlightInline(text, useColor)   – highlight English code snippets
//     embedded inside prose text (e.g. error-message hints that contain
//     single-quoted examples like `'Declare x to be 5.'`)
//
// When useColor is false both functions return plain text with no ANSI codes.
//
// Both functions use the shared tokeniser package so that keyword recognition,
// multi-word operators, possessive handling, and every other tokenisation rule
// are identical to those used by the actual compiler.
package highlight

import (
	"os"
	"regexp"
	"strings"

	"english/token"
	"english/tokeniser"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

// ─── Colour palette (Dracula-inspired) ───────────────────────────────────────

var colorRenderer = func() *lipgloss.Renderer {
	r := lipgloss.NewRenderer(os.Stdout)
	r.SetColorProfile(termenv.TrueColor)
	return r
}()

var (
	// Keywords: general built-in words
	styleKeyword = colorRenderer.NewStyle().Foreground(lipgloss.Color("#BD93F9")) // purple

	// Control-flow keywords: if, repeat, while, break, return, continue …
	styleControlFlow = colorRenderer.NewStyle().Foreground(lipgloss.Color("#FF79C6")) // pink

	// Declaration keywords: declare, let, function, set, call
	styleDeclaration = colorRenderer.NewStyle().Foreground(lipgloss.Color("#8BE9FD")) // cyan

	// String literals
	styleString = colorRenderer.NewStyle().Foreground(lipgloss.Color("#50FA7B")) // green

	// Number literals
	styleNumber = colorRenderer.NewStyle().Foreground(lipgloss.Color("#F1FA8C")) // yellow

	// Boolean literals: true, false
	styleBool = colorRenderer.NewStyle().Foreground(lipgloss.Color("#FFB86C")) // orange

	// Null literals: nothing, none, null
	styleNull = colorRenderer.NewStyle().Foreground(lipgloss.Color("#6272A4")) // muted blue-grey

	// Identifiers (variables, function names)
	styleIdentifier = colorRenderer.NewStyle().Foreground(lipgloss.Color("#F8F8F2")) // white

	// Arithmetic operators: +  -  *  /  =
	styleOperator = colorRenderer.NewStyle().Foreground(lipgloss.Color("#FF5555")) // red

	// Multi-word comparison operators: is equal to, is less than …
	styleComparison = colorRenderer.NewStyle().Foreground(lipgloss.Color("#FFD700")) // gold

	// Possessive operator: 's
	stylePossessive = colorRenderer.NewStyle().Foreground(lipgloss.Color("#8BE9FD")) // cyan

	// Punctuation: . , : ( ) [ ]
	stylePunctuation = colorRenderer.NewStyle().Foreground(lipgloss.Color("#6272A4")) // muted

	// Comments: # …
	styleComment = colorRenderer.NewStyle().Foreground(lipgloss.Color("#6272A4")).Italic(true)

	// Invalid / remainder after a syntax error
	styleInvalid = colorRenderer.NewStyle().Foreground(lipgloss.Color("#44475A")) // dim
)

// ─── Token-kind mapping ───────────────────────────────────────────────────────

// tokenKind is the highlight-level category used to select a colour style.
// It is a coarser grouping than the full token.Type set used by the compiler.
type tokenKind int

const (
	kindWhitespace  tokenKind = iota // spaces / tabs (not newlines)
	kindNewline                      // \n
	kindComment                      // # …
	kindControlFlow                  // if, repeat, while, break, return …
	kindDeclaration                  // declare, let, function, set, call
	kindKeyword                      // other built-in keywords
	kindString                       // "…" or '…' string literals
	kindNumber                       // numeric literals
	kindBool                         // true, false
	kindNull                         // nothing, none, null
	kindOperator                     // + - * / =
	kindComparison                   // multi-word comparison phrases
	kindPossessive                   // 's
	kindPunctuation                  // . , : ( ) [ ]
	kindIdentifier                   // variable / function names
	kindInvalid                      // unrecognised characters / after error
)

// chunk is a single coloured segment of source text.
type chunk struct {
	kind tokenKind
	text string // exact source bytes (including quotes, whitespace in comparisons, etc.)
}

// mapTokenKind maps a compiler token.Type to the coarser highlight tokenKind.
func mapTokenKind(t token.Type) tokenKind {
	switch t {
	// ── Whitespace ────────────────────────────────────────────────────────────
	case token.WHITESPACE:
		return kindWhitespace
	case token.NEWLINE:
		return kindNewline

	// ── Literals ──────────────────────────────────────────────────────────────
	case token.STRING:
		return kindString
	case token.NUMBER:
		return kindNumber
	case token.TRUE, token.FALSE:
		return kindBool
	case token.NOTHING:
		return kindNull

	// ── Operators ─────────────────────────────────────────────────────────────
	case token.PLUS, token.MINUS, token.STAR, token.SLASH, token.ASSIGN:
		return kindOperator

	// ── Comparison operators ──────────────────────────────────────────────────
	case token.IS_EQUAL_TO, token.IS_LESS_THAN, token.IS_GREATER_THAN,
		token.IS_LESS_EQUAL, token.IS_GREATER_EQUAL, token.IS_NOT_EQUAL,
		token.IS_SOMETHING, token.IS_NOTHING_OP,
		token.IS_TRUE, token.IS_FALSE, token.ISNT_TRUE, token.ISNT_FALSE:
		return kindComparison

	// ── Possessive ────────────────────────────────────────────────────────────
	case token.POSSESSIVE:
		return kindPossessive

	// ── Punctuation ───────────────────────────────────────────────────────────
	case token.PERIOD, token.COMMA, token.COLON,
		token.LPAREN, token.RPAREN, token.LBRACKET, token.RBRACKET:
		return kindPunctuation

	// ── Comment ───────────────────────────────────────────────────────────────
	case token.COMMENT:
		return kindComment

	// ── Plain identifier ──────────────────────────────────────────────────────
	case token.IDENTIFIER:
		return kindIdentifier

	// ── Error / unknown ───────────────────────────────────────────────────────
	case token.ERROR:
		return kindInvalid

	// ── Control-flow keywords ─────────────────────────────────────────────────
	case token.IF, token.THEN, token.OTHERWISE,
		token.REPEAT, token.WHILE, token.FOREVER,
		token.BREAK, token.OUT, token.LOOP, token.TIMES,
		token.FOR, token.EACH, token.DO,
		token.RETURN, token.CONTINUE, token.SKIP,
		token.SLEEP:
		return kindControlFlow

	// ── Declaration keywords ──────────────────────────────────────────────────
	case token.DECLARE, token.LET, token.FUNCTION, token.SET, token.CALL,
		token.PLEASE:
		return kindDeclaration

	// ── Everything else is a general keyword ──────────────────────────────────
	default:
		return kindKeyword
	}
}

// tokenize converts source into a slice of coloured chunks using the shared
// tokeniser so that highlighting always reflects the compiler's view of the
// source text.
func tokenize(source string) []chunk {
	tokens := tokeniser.TokenizeForHighlight(source)
	chunks := make([]chunk, 0, len(tokens))
	for _, tok := range tokens {
		chunks = append(chunks, chunk{kind: mapTokenKind(tok.Type), text: tok.Value})
	}
	return chunks
}

// ─── Colour rendering ─────────────────────────────────────────────────────────

// applyStyle colours text using the lipgloss style corresponding to kind.
func applyStyle(kind tokenKind, text string) string {
	switch kind {
	case kindControlFlow:
		return styleControlFlow.Render(text)
	case kindDeclaration:
		return styleDeclaration.Render(text)
	case kindKeyword:
		return styleKeyword.Render(text)
	case kindString:
		return styleString.Render(text)
	case kindNumber:
		return styleNumber.Render(text)
	case kindBool:
		return styleBool.Render(text)
	case kindNull:
		return styleNull.Render(text)
	case kindOperator:
		return styleOperator.Render(text)
	case kindComparison:
		return styleComparison.Render(text)
	case kindPossessive:
		return stylePossessive.Render(text)
	case kindPunctuation:
		return stylePunctuation.Render(text)
	case kindComment:
		return styleComment.Render(text)
	case kindIdentifier:
		return styleIdentifier.Render(text)
	case kindInvalid:
		return styleInvalid.Render(text)
	default:
		return text // whitespace / newline – pass through unchanged
	}
}

// render converts a slice of chunks into a string.
// If useColor is false, only the raw text is emitted (no ANSI codes).
func render(chunks []chunk, useColor bool) string {
	if !useColor {
		var sb strings.Builder
		for _, c := range chunks {
			sb.WriteString(c.text)
		}
		return sb.String()
	}

	var sb strings.Builder
	for _, c := range chunks {
		switch c.kind {
		case kindWhitespace, kindNewline:
			sb.WriteString(c.text)
		default:
			sb.WriteString(applyStyle(c.kind, c.text))
		}
	}
	return sb.String()
}

// ─── Public API ───────────────────────────────────────────────────────────────

// Highlight returns source with ANSI syntax highlighting applied when
// useColor is true.  If the source contains unrecognised characters the
// remainder is coloured in a dim grey so that as much as possible is
// highlighted while invalid sections are still clearly distinguishable.
func Highlight(source string, useColor bool) string {
	if !useColor {
		return source
	}
	return render(tokenize(source), true)
}

// ─── Inline snippet highlighter ──────────────────────────────────────────────

// snippetRe matches single-quoted English code examples embedded in prose
// text.  The pattern requires the content to start with an upper or
// lower-case letter and end before the closing quote so that possessive
// apostrophes in variable names are not mistakenly consumed.
//
// Examples matched:
//
//	'Declare x to be 5.'
//	'Set score to be 10.'
//	'Print "Hello".'
var snippetRe = regexp.MustCompile(`'([A-Za-z][^']*)'`)

// HighlightInline scans text for English code snippets delimited by single
// quotes (e.g. the hint strings produced by parser/messages.go) and applies
// syntax highlighting to each snippet.  Surrounding prose is returned
// unchanged.
//
// When useColor is false the text is returned unmodified.
func HighlightInline(text string, useColor bool) string {
	if !useColor {
		return text
	}
	return snippetRe.ReplaceAllStringFunc(text, func(match string) string {
		// match is the full 'code' including the surrounding single quotes.
		inner := match[1 : len(match)-1]
		highlighted := Highlight(inner, true)
		// Wrap in dim single-quote delimiters so the quoting is still visible.
		quoteStyle := colorRenderer.NewStyle().Foreground(lipgloss.Color("#6272A4"))
		return quoteStyle.Render("'") + highlighted + quoteStyle.Render("'")
	})
}
