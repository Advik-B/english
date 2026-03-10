package highlight_test

import (
	"regexp"
	"strings"
	"testing"

	"english/highlight"
)

// stripANSI removes ANSI escape sequences from s so that tests can check
// plain-text content independently of styling.
var ansiRe = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func stripANSI(s string) string { return ansiRe.ReplaceAllString(s, "") }

// ─── Highlight – plain text (no colour) ──────────────────────────────────────

func TestHighlight_NoColor_ReturnsSourceUnchanged(t *testing.T) {
	source := "# Hello world\nDeclare x to be 5.\nPrint \"Hello\"."
	got := highlight.Highlight(source, false)
	if got != source {
		t.Errorf("expected source unchanged when useColor=false, got:\n%q", got)
	}
}

// ─── Highlight – colour output ────────────────────────────────────────────────

func TestHighlight_Color_ContainsANSI(t *testing.T) {
	source := `Declare x to be 42.`
	got := highlight.Highlight(source, true)
	if !strings.Contains(got, "\x1b[") {
		t.Errorf("expected ANSI codes in coloured output, got:\n%q", got)
	}
}

func TestHighlight_Color_PreservesText(t *testing.T) {
	source := "Declare score to be 100.\nPrint \"Hello, World!\".\n# a comment"
	got := highlight.Highlight(source, true)
	for _, fragment := range []string{"Declare", "score", "100", `"Hello, World!"`, "# a comment"} {
		if !strings.Contains(stripANSI(got), fragment) {
			t.Errorf("expected %q in stripped output, got:\n%s", fragment, stripANSI(got))
		}
	}
}

func TestHighlight_Color_Keywords(t *testing.T) {
	cases := []string{
		"if", "repeat", "while", "break", "return",
		"declare", "let", "function", "set", "call",
		"print", "true", "false", "nothing",
	}
	for _, kw := range cases {
		source := kw + " "
		got := stripANSI(highlight.Highlight(source, true))
		if !strings.Contains(got, kw) {
			t.Errorf("keyword %q not found in stripped output %q", kw, got)
		}
	}
}

func TestHighlight_Color_StringLiteral(t *testing.T) {
	source := `Print "hello world".`
	got := stripANSI(highlight.Highlight(source, true))
	if !strings.Contains(got, `"hello world"`) {
		t.Errorf("expected string literal in output, got:\n%s", got)
	}
}

func TestHighlight_Color_NumberLiteral(t *testing.T) {
	source := `Declare x to be 3.14.`
	got := stripANSI(highlight.Highlight(source, true))
	if !strings.Contains(got, "3.14") {
		t.Errorf("expected number literal in output, got:\n%s", got)
	}
}

func TestHighlight_Color_Comment(t *testing.T) {
	source := `# This is a comment`
	got := stripANSI(highlight.Highlight(source, true))
	if !strings.Contains(got, "# This is a comment") {
		t.Errorf("expected comment text in output, got:\n%s", got)
	}
}

func TestHighlight_Color_MultiWordComparison(t *testing.T) {
	source := `If x is equal to 5 then`
	got := stripANSI(highlight.Highlight(source, true))
	if !strings.Contains(got, "is equal to") {
		t.Errorf("expected comparison operator in output, got:\n%s", got)
	}
}

func TestHighlight_Color_NewlinesPreserved(t *testing.T) {
	source := "line one\nline two\nline three\n"
	got := highlight.Highlight(source, true)
	if strings.Count(got, "\n") != strings.Count(source, "\n") {
		t.Errorf("newline count changed: want %d, got %d",
			strings.Count(source, "\n"), strings.Count(got, "\n"))
	}
}

func TestHighlight_InvalidSyntax_GracefulFallback(t *testing.T) {
	source := "Declare x to be 5.\n@@invalid@@\n"
	got := highlight.Highlight(source, true)
	plain := stripANSI(got)
	if !strings.Contains(plain, "@@invalid@@") {
		t.Errorf("expected invalid chars preserved in output, got:\n%s", plain)
	}
	if !strings.Contains(plain, "Declare") {
		t.Errorf("expected valid code before error in output, got:\n%s", plain)
	}
}

// ─── HighlightInline ──────────────────────────────────────────────────────────

func TestHighlightInline_NoColor_ReturnsUnchanged(t *testing.T) {
	text := "Try 'Declare x to be 5.' for example."
	got := highlight.HighlightInline(text, false)
	if got != text {
		t.Errorf("expected text unchanged when useColor=false, got:\n%q", got)
	}
}

func TestHighlightInline_Color_HighlightsSnippets(t *testing.T) {
	text := "Use 'Set pi to be 3.14.' to update the value."
	got := highlight.HighlightInline(text, true)
	if !strings.Contains(got, "\x1b[") {
		t.Errorf("expected ANSI codes in coloured output, got:\n%q", got)
	}
	plain := stripANSI(got)
	if !strings.Contains(plain, "Set") {
		t.Errorf("expected 'Set' in stripped output, got:\n%s", plain)
	}
	if !strings.Contains(plain, "pi") {
		t.Errorf("expected 'pi' in stripped output, got:\n%s", plain)
	}
}

func TestHighlightInline_Color_PreservesSurroundingProse(t *testing.T) {
	text := "Before 'Declare x to be 5.' after"
	got := stripANSI(highlight.HighlightInline(text, true))
	if !strings.Contains(got, "Before") {
		t.Errorf("expected 'Before' in output, got:\n%s", got)
	}
	if !strings.Contains(got, "after") {
		t.Errorf("expected 'after' in output, got:\n%s", got)
	}
}

func TestHighlightInline_Color_NoSnippets_NoChange(t *testing.T) {
	text := "There are no code snippets here."
	got := stripANSI(highlight.HighlightInline(text, true))
	if got != text {
		t.Errorf("expected prose unchanged when no snippets, got:\n%q", got)
	}
}
