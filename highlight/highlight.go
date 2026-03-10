// Package highlight provides syntax highlighting for English source code.
//
// It exposes two main entry points:
//   - Highlight(source, useColor)       – colour an entire source file
//   - HighlightInline(text, useColor)   – highlight English code snippets
//     embedded inside prose text (e.g. error-message hints that contain
//     single-quoted examples like `'Declare x to be 5.'`)
//
// When useColor is false both functions return plain text with no ANSI codes.
package highlight

import (
	"os"
	"regexp"
	"strings"
	"unicode"

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

// ─── Token categories ─────────────────────────────────────────────────────────

type tokenKind int

const (
	kindWhitespace tokenKind = iota // spaces / tabs (not newlines)
	kindNewline                     // \n
	kindComment                     // # …
	kindControlFlow                 // if, repeat, while, break, return …
	kindDeclaration                 // declare, let, function, set, call
	kindKeyword                     // other built-in keywords
	kindString                      // "…" or '…' string literals
	kindNumber                      // numeric literals
	kindBool                        // true, false
	kindNull                        // nothing, none, null
	kindOperator                    // + - * / =
	kindComparison                  // multi-word comparison phrases
	kindPossessive                  // 's
	kindPunctuation                 // . , : ( ) [ ]
	kindIdentifier                  // variable / function names
	kindInvalid                     // unrecognised characters / after error
)

// chunk is a single coloured segment of source text.
type chunk struct {
	kind tokenKind
	text string // exact source bytes (including quotes, whitespace in comparisons, etc.)
}

// ─── Keyword classification ───────────────────────────────────────────────────

// controlFlowWords is the set of keywords that direct program flow.
var controlFlowWords = map[string]bool{
	"if":        true,
	"then":      true,
	"otherwise": true,
	"repeat":    true,
	"while":     true,
	"forever":   true,
	"break":     true,
	"out":       true,
	"loop":      true,
	"times":     true,
	"for":       true,
	"each":      true,
	"do":        true,
	"return":    true,
	"continue":  true,
	"skip":      true,
	"thats":     true,
	"it":        true,
	// Ask-statement helpers: "store it in", "store the answer in", "store the result in"
	"store":  true,
	"answer": true,
	"result": true,
}

// declarationWords are keywords used when introducing names/functions.
var declarationWords = map[string]bool{
	"declare":  true,
	"let":      true,
	"function": true,
	"set":      true,
	"call":     true,
}

// allKeywords is every reserved word in the language (all lowercase).
var allKeywords = map[string]bool{
	"declare": true, "let": true, "equal": true, "function": true,
	"that": true, "does": true, "the": true, "following": true,
	"thats": true, "it": true, "to": true, "be": true, "always": true,
	"set": true, "call": true, "return": true, "print": true,
	"if": true, "then": true, "otherwise": true, "repeat": true,
	"while": true, "forever": true, "break": true, "out": true,
	"loop": true, "times": true, "for": true, "each": true,
	"in": true, "do": true, "takes": true, "and": true, "with": true,
	"of": true, "calling": true, "value": true, "item": true,
	"at": true, "position": true, "length": true, "remainder": true,
	"divided": true, "by": true, "toggle": true, "location": true,
	"write": true, "as": true, "structure": true, "struct": true,
	"fields": true, "field": true, "instance": true, "new": true,
	"try": true, "doing": true, "on": true, "finally": true,
	"raise": true, "reference": true, "copy": true, "swap": true,
	"casted": true, "cast": true, "type": true, "which": true,
	"is": true, "from": true, "unsigned": true, "integer": true,
	"default": true, "but": true, "import": true, "everything": true,
	"all": true, "safely": true, "continue": true, "skip": true,
	"not": true, "or": true, "ask": true, "array": true,
	"lookup": true, "table": true, "has": true, "entry": true,
	"error": true, "onerror": true,
}

// classifyWord returns the tokenKind for a lower-cased keyword.
func classifyWord(lower string) tokenKind {
	switch lower {
	case "true", "false":
		return kindBool
	case "nothing", "none", "null":
		return kindNull
	}
	if controlFlowWords[lower] {
		return kindControlFlow
	}
	if declarationWords[lower] {
		return kindDeclaration
	}
	if allKeywords[lower] {
		return kindKeyword
	}
	return kindIdentifier
}

// ─── Scanner ──────────────────────────────────────────────────────────────────

// scanner walks through source bytes and emits chunks.
type scanner struct {
	src string
	pos int

	// lastWasValue tracks whether the previous non-whitespace token was a
	// value-ending token so the possessive 's can be recognised.
	lastWasValue bool
}

func newScanner(src string) *scanner { return &scanner{src: src} }

func (s *scanner) done() bool     { return s.pos >= len(s.src) }
func (s *scanner) peek() byte     { return s.peekAt(0) }
func (s *scanner) peekAt(n int) byte {
	i := s.pos + n
	if i >= len(s.src) {
		return 0
	}
	return s.src[i]
}

func (s *scanner) advance() byte {
	ch := s.src[s.pos]
	s.pos++
	return ch
}

// readWhile consumes bytes while pred returns true, returns the consumed text.
func (s *scanner) readWhile(pred func(byte) bool) string {
	start := s.pos
	for !s.done() && pred(s.src[s.pos]) {
		s.pos++
	}
	return s.src[start:s.pos]
}

func isIdentByte(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') ||
		(b >= '0' && b <= '9') || b == '_'
}

// readWord reads an ASCII word (letters, digits, _).
func (s *scanner) readWord() string {
	start := s.pos
	for !s.done() && isIdentByte(s.src[s.pos]) {
		s.pos++
	}
	return s.src[start:s.pos]
}

// readSpaces reads horizontal whitespace (spaces/tabs) only.
func (s *scanner) readSpaces() string {
	return s.readWhile(func(b byte) bool { return b == ' ' || b == '\t' || b == '\r' })
}

// peekWordAt returns the lowercase word starting at src[offset] without advancing.
func (s *scanner) peekWordAt(offset int) string {
	i := offset
	for i < len(s.src) && isIdentByte(s.src[i]) {
		i++
	}
	return strings.ToLower(s.src[offset:i])
}

// tryMultiWordComparison attempts to match a multi-word comparison phrase
// starting at the current position (which must be the start of "is" or "has").
// Returns the raw source text and true on success, or ("", false) on failure.
func (s *scanner) tryMultiWordComparison() (string, bool) {
	save := s.pos

	// Collect words separated by horizontal whitespace, building the phrase.
	var rawBuf strings.Builder
	var phraseBuf strings.Builder

	var bestRaw string

	multiPhrases := map[string]bool{
		"is equal to":                 true,
		"is less than":                true,
		"is greater than":             true,
		"is less than or equal to":    true,
		"is greater than or equal to": true,
		"is not equal to":             true,
		"is something":                true,
		"is nothing":                  true,
		"has a value":                 true,
		"has no value":                true,
		"is true":                     true,
		"is false":                    true,
		"isn't true":                  true,
		"isn't false":                 true,
		"is not true":                 true,
		"is not false":                true,
	}

	for {
		sp := s.readSpaces()
		if !s.done() && unicode.IsLetter(rune(s.peek())) {
			word := s.readWord()

			// Handle "isn't" contraction: after reading "isn", consume "'t" to form "isn't"
			if strings.ToLower(word) == "isn" && !s.done() && s.peek() == '\'' &&
				(s.peekAt(1) == 't' || s.peekAt(1) == 'T') {
				s.advance() // consume '
				s.advance() // consume t
				word = "isn't"
			}

			rawBuf.WriteString(sp)
			rawBuf.WriteString(word)
			if phraseBuf.Len() > 0 {
				phraseBuf.WriteByte(' ')
			}
			phraseBuf.WriteString(strings.ToLower(word))

			phrase := phraseBuf.String()
			if multiPhrases[phrase] {
				bestRaw = rawBuf.String()
			}
			// Stop if next non-space char is punctuation or end
			peek := s.peek()
			if peek == 0 || peek == ',' || peek == ':' || peek == '.' ||
				peek == '\n' || peek == ')' || peek == ']' {
				break
			}
		} else {
			// Whitespace leads to non-letter; put spaces back and stop
			s.pos -= len(sp)
			break
		}
	}

	if bestRaw != "" {
		return bestRaw, true
	}
	s.pos = save
	return "", false
}

// readString reads a quoted string literal starting at the current quote char.
// Returns the full raw text including the surrounding quotes.
func (s *scanner) readString(quote byte) string {
	start := s.pos
	s.pos++ // skip opening quote
	for !s.done() {
		ch := s.src[s.pos]
		if ch == '\\' {
			s.pos += 2 // skip escape sequence
			continue
		}
		if ch == quote {
			s.pos++ // skip closing quote
			break
		}
		if ch == '\n' {
			break // unterminated string – stop at newline
		}
		s.pos++
	}
	return s.src[start:s.pos]
}

// isPossessiveContext returns true if the previous non-whitespace token was
// a value-ending token (identifier, literal, closing delimiter).
func (s *scanner) isPossessiveContext() bool {
	return s.lastWasValue
}

// tokenize converts source into a slice of coloured chunks.
// On an unrecognised character the rest of the source is emitted as kindInvalid.
func tokenize(source string) []chunk {
	s := newScanner(source)
	var chunks []chunk

	for !s.done() {
		ch := s.peek()

		// ── Newline ───────────────────────────────────────────────────
		if ch == '\n' {
			s.advance()
			chunks = append(chunks, chunk{kindNewline, "\n"})
			continue
		}

		// ── Horizontal whitespace ─────────────────────────────────────
		if ch == ' ' || ch == '\t' || ch == '\r' {
			sp := s.readSpaces()
			chunks = append(chunks, chunk{kindWhitespace, sp})
			continue
		}

		// ── Comment ───────────────────────────────────────────────────
		if ch == '#' {
			start := s.pos
			s.advance() // '#'
			for !s.done() && s.peek() != '\n' {
				s.advance()
			}
			chunks = append(chunks, chunk{kindComment, s.src[start:s.pos]})
			s.lastWasValue = false
			continue
		}

		// ── Multi-word comparisons starting with 'is', 'isn't', or 'has' ──
		// Attempt when the peeked word is exactly "is", "isn" (for "isn't"), or "has".
		if ch == 'i' || ch == 'I' || ch == 'h' || ch == 'H' {
			word := s.peekWordAt(s.pos)
			if word == "is" || word == "isn" || word == "has" {
				if raw, ok := s.tryMultiWordComparison(); ok {
					chunks = append(chunks, chunk{kindComparison, raw})
					s.lastWasValue = false
					continue
				}
			}
		}

		// ── Possessive 's ─────────────────────────────────────────────
		if ch == '\'' && s.peekAt(1) == 's' && !isIdentByte(s.peekAt(2)) &&
			s.isPossessiveContext() {
			s.advance() // '
			s.advance() // s
			chunks = append(chunks, chunk{kindPossessive, "'s"})
			s.lastWasValue = true
			continue
		}

		// ── String literals ───────────────────────────────────────────
		if ch == '"' || ch == '\'' {
			raw := s.readString(ch)
			chunks = append(chunks, chunk{kindString, raw})
			s.lastWasValue = true
			continue
		}

		// ── Number literals ───────────────────────────────────────────
		if ch >= '0' && ch <= '9' {
			start := s.pos
			for !s.done() && s.src[s.pos] >= '0' && s.src[s.pos] <= '9' {
				s.pos++
			}
			if !s.done() && s.src[s.pos] == '.' {
				next := s.peekAt(1)
				if next >= '0' && next <= '9' {
					s.pos++ // dot
					for !s.done() && s.src[s.pos] >= '0' && s.src[s.pos] <= '9' {
						s.pos++
					}
				}
			}
			chunks = append(chunks, chunk{kindNumber, s.src[start:s.pos]})
			s.lastWasValue = true
			continue
		}

		// ── Words (keywords + identifiers) ────────────────────────────
		if unicode.IsLetter(rune(ch)) || ch == '_' {
			word := s.readWord()
			kind := classifyWord(strings.ToLower(word))
			chunks = append(chunks, chunk{kind, word})
			// Identifiers and boolean/null literals end a value expression.
			s.lastWasValue = kind == kindIdentifier || kind == kindBool || kind == kindNull
			continue
		}

		// ── Operators ─────────────────────────────────────────────────
		switch ch {
		case '+', '-', '*', '/', '=':
			s.advance()
			chunks = append(chunks, chunk{kindOperator, string(ch)})
			s.lastWasValue = false
			continue
		}

		// ── Punctuation ───────────────────────────────────────────────
		switch ch {
		case '.', ',', ':', '(', ')', '[', ']':
			s.advance()
			isClose := ch == ')' || ch == ']'
			chunks = append(chunks, chunk{kindPunctuation, string(ch)})
			s.lastWasValue = isClose
			continue
		}

		// ── Unrecognised – emit rest as invalid ───────────────────────
		rest := s.src[s.pos:]
		chunks = append(chunks, chunk{kindInvalid, rest})
		s.pos = len(s.src)
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
