package tokeniser_test

import (
	"strings"
	"testing"

	"github.com/Advik-B/english/token"
	"github.com/Advik-B/english/tokeniser"
)

// ─── TokenizeForHighlight ─────────────────────────────────────────────────────

// TestTokenizeForHighlight_ReconstructsSource checks that concatenating the
// Value field of every token returned by TokenizeForHighlight reproduces the
// original source exactly.
func TestTokenizeForHighlight_ReconstructsSource(t *testing.T) {
	cases := []string{
		`Declare x to be 5.`,
		"# comment\nDeclare x to be 5.\n",
		`If x is equal to 5 then`,
		`Print "Hello, World!".`,
		`Declare score to be 3.14.`,
		"line one\nline two\n",
		`Declare x to be 5.` + "\n@@invalid@@\n",
		`"hello"'s length`,
		`Is  equal  to`,          // extra spacing preserved
		"If x isn't true then\n", // contraction preserved
		`If x is greater than or equal to 10 then`,
	}

	for _, src := range cases {
		t.Run(src, func(t *testing.T) {
			toks := tokeniser.TokenizeForHighlight(src)
			var sb strings.Builder
			for _, tok := range toks {
				sb.WriteString(tok.Value)
			}
			got := sb.String()
			if got != src {
				t.Errorf("reconstruction mismatch\nwant: %q\n got: %q", src, got)
			}
		})
	}
}

// TestTokenizeForHighlight_WhitespaceTokens verifies that WHITESPACE tokens
// are emitted for horizontal whitespace between semantic tokens.
func TestTokenizeForHighlight_WhitespaceTokens(t *testing.T) {
	src := "Declare x to be 5."
	toks := tokeniser.TokenizeForHighlight(src)
	for _, tok := range toks {
		if tok.Type == token.WHITESPACE && tok.Value == "" {
			t.Errorf("empty WHITESPACE token emitted")
		}
	}

	// There should be at least one WHITESPACE token (between "Declare" and "x").
	found := false
	for _, tok := range toks {
		if tok.Type == token.WHITESPACE {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected at least one WHITESPACE token in %q", src)
	}
}

// TestTokenizeForHighlight_NewlineTokens verifies that NEWLINE tokens are kept.
func TestTokenizeForHighlight_NewlineTokens(t *testing.T) {
	src := "line one\nline two\n"
	toks := tokeniser.TokenizeForHighlight(src)
	count := 0
	for _, tok := range toks {
		if tok.Type == token.NEWLINE {
			count++
		}
	}
	if count != 2 {
		t.Errorf("expected 2 NEWLINE tokens, got %d", count)
	}
}

// TestTokenizeForHighlight_CommentRawText verifies that the '#' is included in
// the raw Value of COMMENT tokens.
func TestTokenizeForHighlight_CommentRawText(t *testing.T) {
	src := "# hello world\n"
	toks := tokeniser.TokenizeForHighlight(src)
	for _, tok := range toks {
		if tok.Type == token.COMMENT {
			if !strings.HasPrefix(tok.Value, "#") {
				t.Errorf("COMMENT token Value should start with '#', got %q", tok.Value)
			}
			return
		}
	}
	t.Errorf("no COMMENT token found in %q", src)
}

// TestTokenizeForHighlight_StringRawText verifies that string tokens include
// their surrounding quote characters.
func TestTokenizeForHighlight_StringRawText(t *testing.T) {
	src := `Print "hello".`
	toks := tokeniser.TokenizeForHighlight(src)
	for _, tok := range toks {
		if tok.Type == token.STRING {
			if tok.Value != `"hello"` {
				t.Errorf("STRING token Value should be %q, got %q", `"hello"`, tok.Value)
			}
			return
		}
	}
	t.Errorf("no STRING token found in %q", src)
}

// TestTokenizeForHighlight_ComparisonOriginalSpacing verifies that multi-word
// comparison tokens preserve the original spacing from the source.
func TestTokenizeForHighlight_ComparisonOriginalSpacing(t *testing.T) {
	// Two spaces between words – the raw Value must match the source exactly.
	src := "x is  equal  to y"
	toks := tokeniser.TokenizeForHighlight(src)
	for _, tok := range toks {
		if tok.Type == token.IS_EQUAL_TO {
			if tok.Value != "is  equal  to" {
				t.Errorf("IS_EQUAL_TO raw Value should be %q, got %q", "is  equal  to", tok.Value)
			}
			return
		}
	}
	t.Errorf("no IS_EQUAL_TO token found in %q", src)
}

// ─── Lexer / TokenizeAll (existing behaviour unchanged) ──────────────────────

// TestNewLexer_TokenizeAll_BasicDeclaration checks the semantic token stream
// is unchanged after refactoring.
func TestNewLexer_TokenizeAll_BasicDeclaration(t *testing.T) {
	src := "Declare x to be 5."
	l := tokeniser.NewLexer(src)
	toks := l.TokenizeAll()

	want := []token.Type{
		token.DECLARE, token.IDENTIFIER, token.TO, token.BE, token.NUMBER,
		token.PERIOD, token.EOF,
	}
	if len(toks) != len(want) {
		t.Fatalf("want %d tokens, got %d: %v", len(want), len(toks), toks)
	}
	for i, tt := range want {
		if toks[i].Type != tt {
			t.Errorf("token[%d]: want %s, got %s", i, tt, toks[i].Type)
		}
	}
}

// TestTokenizeForHighlight_UnterminatedString verifies that unterminated
// strings don't cause a panic due to slice bounds errors.
// This is a regression test for the bug where the lexer position could
// exceed the source string length.
func TestTokenizeForHighlight_UnterminatedString(t *testing.T) {
	cases := []string{
		`print rt's casefold'.`, // From bug report
		`'`,                     // Single quote at end
		`"`,                     // Double quote at end
		`'hello`,                // Unterminated single-quoted string
		`"hello`,                // Unterminated double-quoted string
		`x'`,                    // Single char followed by quote
		`''`,                    // Two quotes
	}

	for _, src := range cases {
		t.Run(src, func(t *testing.T) {
			// Should not panic
			toks := tokeniser.TokenizeForHighlight(src)
			// Verify we got some tokens
			if len(toks) == 0 {
				t.Errorf("expected at least one token for %q", src)
			}
			// Verify reconstruction doesn't panic and doesn't exceed source length
			var sb strings.Builder
			for _, tok := range toks {
				sb.WriteString(tok.Value)
			}
			got := sb.String()
			if len(got) > len(src) {
				t.Errorf("reconstruction longer than source\nsrc: %q (len=%d)\ngot: %q (len=%d)",
					src, len(src), got, len(got))
			}
		})
	}
}

// ─── Word boundaries, numbers and non-ASCII names ────────────────────────────

// tokenTypes lexes source and returns the token types, EOF included.
func tokenTypes(src string) []token.Type {
	var out []token.Type
	for _, tok := range tokeniser.NewLexer(src).TokenizeAll() {
		out = append(out, tok.Type)
	}
	return out
}

// TestComparisonNeedsAWholeWord covers the multi-word comparison scan, which
// triggered on any text starting with the letters "is" or "has" — "island",
// "hash", "is_digit" — and then read forward to the end of the line looking
// for a phrase before rolling back, which is quadratic on a line full of such
// words and can never match.
func TestComparisonNeedsAWholeWord(t *testing.T) {
	for _, src := range []string{"island", "hash", "is_digit", "issue", "haskell"} {
		got := tokenTypes(src)
		if len(got) != 2 || got[0] != token.IDENTIFIER {
			t.Errorf("%q lexed as %v, want one identifier", src, got)
		}
	}

	// The real phrases still lex as one token each.
	for src, want := range map[string]token.Type{
		"is equal to":                 token.IS_EQUAL_TO,
		"is greater than or equal to": token.IS_GREATER_EQUAL,
		"has a value":                 token.IS_SOMETHING,
		"isn't true":                  token.ISNT_TRUE,
	} {
		got := tokenTypes(src)
		if len(got) != 2 || got[0] != want {
			t.Errorf("%q lexed as %v, want %v", src, got, want)
		}
	}
}

// TestScientificNotation covers exponents, which the lexer did not read at
// all: "1e10" became the number 1 followed by the name "e10", which parsed as
// two things and meant neither.
func TestScientificNotation(t *testing.T) {
	for _, src := range []string{"1e10", "2.5E-3", "6e+23", "1E0"} {
		got := tokenTypes(src)
		if len(got) != 2 || got[0] != token.NUMBER {
			t.Errorf("%q lexed as %v, want one number", src, got)
		}
		if value := tokeniser.NewLexer(src).TokenizeAll()[0].Value; value != src {
			t.Errorf("%q lexed with the value %q", src, value)
		}
	}

	// An "e" that is not an exponent is left where it belongs.
	got := tokenTypes("1 exp")
	if len(got) != 3 || got[0] != token.NUMBER || got[1] != token.IDENTIFIER {
		t.Errorf("\"1 exp\" lexed as %v, want a number and a name", got)
	}
}

// TestNonASCIINames covers identifiers with a non-ASCII letter. The lexer
// scans bytes, and testing one byte of a multi-byte character with
// unicode.IsLetter asks the wrong question: the first byte of "é" is a letter
// as a rune and its second is not, so "café" lexed as a name, a one-byte name
// and an unrecognised character.
func TestNonASCIINames(t *testing.T) {
	got := tokenTypes("café")
	if len(got) != 2 || got[0] != token.IDENTIFIER {
		t.Errorf("\"café\" lexed as %v, want one identifier", got)
	}
	if value := tokeniser.NewLexer("café").TokenizeAll()[0].Value; value != "café" {
		t.Errorf("the name lexed as %q", value)
	}
	for _, tok := range tokeniser.NewLexer(`Declare naïve to be "résumé".`).TokenizeAll() {
		if tok.Type == token.ERROR {
			t.Errorf("a non-ASCII letter produced an error token: %q", tok.Value)
		}
	}
}
