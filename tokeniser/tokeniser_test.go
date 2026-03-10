package tokeniser_test

import (
	"strings"
	"testing"

	"english/token"
	"english/tokeniser"
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
		`Is  equal  to`,                  // extra spacing preserved
		"If x isn't true then\n",        // contraction preserved
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
