package stacktraces_test

import (
	"errors"
	"regexp"
	"strings"
	"testing"

	"github.com/Advik-B/english/stacktraces"
)

// stripANSI removes ANSI escape sequences from s so that tests can check
// the plain-text content of coloured output independently of styling.
var ansiRe = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func stripANSI(s string) string { return ansiRe.ReplaceAllString(s, "") }

// fakeRuntimeError is a test double that satisfies stacktraces.RuntimeError.
type fakeRuntimeError struct {
	msg   string
	stack []string
	line  int
}

func (e *fakeRuntimeError) Error() string          { return "Runtime Error: " + e.msg }
func (e *fakeRuntimeError) RuntimeMessage() string { return e.msg }
func (e *fakeRuntimeError) RuntimeLine() int       { return e.line }
func (e *fakeRuntimeError) RuntimeCallStack() []string {
	if e.stack == nil {
		return []string{}
	}
	return e.stack
}

// fakeCompileError is a test double that satisfies stacktraces.CompileError.
type fakeCompileError struct {
	msg  string
	line int
}

func (e *fakeCompileError) Error() string       { return e.msg }
func (e *fakeCompileError) CompileMessage() string { return e.msg }
func (e *fakeCompileError) CompileLine() int       { return e.line }

// ─── HasColor ────────────────────────────────────────────────────────────────

func TestHasColor_NoColorEnv(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	if stacktraces.HasColor() {
		t.Fatal("expected HasColor() == false when NO_COLOR is set")
	}
}

// ─── Render – plain text (NO_COLOR) ──────────────────────────────────────────

func TestRender_NilError(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	got := stacktraces.Render(nil)
	if got != "" {
		t.Fatalf("expected empty string for nil error, got %q", got)
	}
}

func TestRender_PlainRuntimeError(t *testing.T) {
	t.Setenv("NO_COLOR", "1")

	re := &fakeRuntimeError{
		msg:   "division by zero",
		stack: []string{"<main>", "myFunc"},
	}

	got := stacktraces.Render(re)

	if !strings.Contains(got, "division by zero") {
		t.Errorf("expected message in output, got:\n%s", got)
	}
	if !strings.Contains(got, "<main>") {
		t.Errorf("expected <main> frame in output, got:\n%s", got)
	}
	if !strings.Contains(got, "myFunc") {
		t.Errorf("expected myFunc frame in output, got:\n%s", got)
	}
	if !strings.Contains(got, "Call Stack") {
		t.Errorf("expected 'Call Stack' section in output, got:\n%s", got)
	}
}

func TestRender_PlainRuntimeError_EmptyStack(t *testing.T) {
	t.Setenv("NO_COLOR", "1")

	re := &fakeRuntimeError{msg: "oops", stack: []string{}}
	got := stacktraces.Render(re)

	if !strings.Contains(got, "oops") {
		t.Errorf("expected message in output, got:\n%s", got)
	}
	if strings.Contains(got, "Call Stack") {
		t.Errorf("did not expect 'Call Stack' for empty stack, got:\n%s", got)
	}
}

func TestRender_PlainGenericError(t *testing.T) {
	t.Setenv("NO_COLOR", "1")

	err := errors.New("something went wrong")
	got := stacktraces.Render(err)

	if !strings.Contains(got, "something went wrong") {
		t.Errorf("expected error message in output, got:\n%s", got)
	}
}

// ─── Render – coloured output ────────────────────────────────────────────────

// TestRenderWithColor_RuntimeError explicitly exercises the coloured rendering
// path by calling RenderWithColor with color=true, regardless of the terminal
// environment.
func TestRenderWithColor_RuntimeError(t *testing.T) {
	re := &fakeRuntimeError{
		msg:   "index out of range",
		stack: []string{"<main>", "process", "<stdlib>"},
	}

	got := stacktraces.RenderWithColor(re, true)

	// The coloured output must still contain the key text.
	for _, want := range []string{"index out of range", "<main>", "process", "<stdlib>"} {
		if !strings.Contains(got, want) {
			t.Errorf("expected %q in coloured output, got:\n%s", want, got)
		}
	}
	// Coloured output should contain at least one ANSI escape sequence.
	if !strings.Contains(got, "\x1b[") {
		t.Errorf("expected ANSI escape codes in coloured output, got:\n%q", got)
	}
}

func TestRenderWithColor_GenericError(t *testing.T) {
	err := errors.New("type mismatch")
	got := stacktraces.RenderWithColor(err, true)

	if !strings.Contains(got, "type mismatch") {
		t.Errorf("expected error message in coloured output, got:\n%s", got)
	}
	if !strings.Contains(got, "\x1b[") {
		t.Errorf("expected ANSI escape codes in coloured output, got:\n%q", got)
	}
}

func TestRenderWithColor_PlainVsColored_DifferentOutput(t *testing.T) {
	re := &fakeRuntimeError{msg: "boom", stack: []string{"<main>"}}
	plain := stacktraces.RenderWithColor(re, false)
	colored := stacktraces.RenderWithColor(re, true)

	if plain == colored {
		t.Error("expected plain and coloured renders to differ")
	}
}

// ─── Print (smoke test – just verifies no panic) ─────────────────────────────

func TestPrint_DoesNotPanic(t *testing.T) {
	t.Setenv("NO_COLOR", "1")

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Print panicked: %v", r)
		}
	}()

	stacktraces.Print(errors.New("test error"))
	stacktraces.Print(nil)
}

// ─── Compile error rendering ──────────────────────────────────────────────────

func TestRender_PlainCompileError_WithLine(t *testing.T) {
	t.Setenv("NO_COLOR", "1")

	ce := &fakeCompileError{msg: "variable 'x' is already declared at line 3", line: 7}
	got := stacktraces.Render(ce)

	if !strings.Contains(got, "Compile Error") {
		t.Errorf("expected 'Compile Error' header in plain output, got:\n%s", got)
	}
	if !strings.Contains(got, "line 7") {
		t.Errorf("expected 'line 7' in plain output, got:\n%s", got)
	}
	if !strings.Contains(got, ce.msg) {
		t.Errorf("expected message in plain output, got:\n%s", got)
	}
}

func TestRender_PlainCompileError_NoLine(t *testing.T) {
	t.Setenv("NO_COLOR", "1")

	ce := &fakeCompileError{msg: "unknown type 'foo'", line: 0}
	got := stacktraces.Render(ce)

	if !strings.Contains(got, "Compile Error") {
		t.Errorf("expected 'Compile Error' in plain output, got:\n%s", got)
	}
	if !strings.Contains(got, "unknown type 'foo'") {
		t.Errorf("expected message in plain output, got:\n%s", got)
	}
}

func TestRenderWithColor_CompileError(t *testing.T) {
	ce := &fakeCompileError{msg: "variable 'pi' shadows a predefined constant", line: 37}
	got := stacktraces.RenderWithColor(ce, true)

	for _, want := range []string{"Compile Error", "37", "pi"} {
		if !strings.Contains(got, want) {
			t.Errorf("expected %q in coloured compile error output, got:\n%s", want, got)
		}
	}
	if !strings.Contains(got, "\x1b[") {
		t.Errorf("expected ANSI escape codes in coloured output, got:\n%q", got)
	}
}

func TestRender_CompileVsRuntime_DifferentHeaders(t *testing.T) {
	t.Setenv("NO_COLOR", "1")

	ce := &fakeCompileError{msg: "type mismatch", line: 5}
	re := &fakeRuntimeError{msg: "division by zero", stack: []string{}}

	compileOut := stacktraces.Render(ce)
	runtimeOut := stacktraces.Render(re)

	if !strings.Contains(compileOut, "Compile Error") {
		t.Errorf("expected 'Compile Error' in compile output, got:\n%s", compileOut)
	}
	if !strings.Contains(runtimeOut, "Runtime Error") {
		t.Errorf("expected 'Runtime Error' in runtime output, got:\n%s", runtimeOut)
	}
	if strings.Contains(compileOut, "Runtime Error") {
		t.Errorf("compile output should not contain 'Runtime Error', got:\n%s", compileOut)
	}
	if strings.Contains(runtimeOut, "Compile Error") {
		t.Errorf("runtime output should not contain 'Compile Error', got:\n%s", runtimeOut)
	}
}

// ─── Runtime error line number ────────────────────────────────────────────────

func TestRender_PlainRuntimeError_WithLine(t *testing.T) {
	t.Setenv("NO_COLOR", "1")

	re := &fakeRuntimeError{msg: "division by zero", stack: []string{"<main>"}, line: 7}
	got := stacktraces.Render(re)

	if !strings.Contains(got, "line 7") {
		t.Errorf("expected 'line 7' in plain output, got:\n%s", got)
	}
	if !strings.Contains(got, "division by zero") {
		t.Errorf("expected message in plain output, got:\n%s", got)
	}
}

func TestRender_PlainRuntimeError_NoLine(t *testing.T) {
	t.Setenv("NO_COLOR", "1")

	re := &fakeRuntimeError{msg: "something went wrong", stack: []string{"<main>"}, line: 0}
	got := stacktraces.Render(re)

	if strings.Contains(got, "line 0") {
		t.Errorf("should not show 'line 0' when line is unknown, got:\n%s", got)
	}
	if !strings.Contains(got, "Runtime Error") {
		t.Errorf("expected 'Runtime Error' in output, got:\n%s", got)
	}
}

func TestRenderWithColor_RuntimeError_WithLine(t *testing.T) {
	re := &fakeRuntimeError{
		msg:   "index out of range",
		stack: []string{"<main>", "processItems"},
		line:  42,
	}
	got := stacktraces.RenderWithColor(re, true)

	for _, want := range []string{"index out of range", "42", "<main>", "processItems"} {
		if !strings.Contains(got, want) {
			t.Errorf("expected %q in coloured output with line, got:\n%s", want, got)
		}
	}
}

// ─── Syntax error rendering ───────────────────────────────────────────────────

// fakeSyntaxError is a test double that satisfies stacktraces.SyntaxError.
type fakeSyntaxError struct {
	msg  string
	line int
	col  int
	hint string
}

func (e *fakeSyntaxError) Error() string       { return "Syntax Error: " + e.msg }
func (e *fakeSyntaxError) SyntaxMessage() string { return e.msg }
func (e *fakeSyntaxError) SyntaxLine() int       { return e.line }
func (e *fakeSyntaxError) SyntaxCol() int        { return e.col }
func (e *fakeSyntaxError) SyntaxHint() string    { return e.hint }

func TestRender_PlainSyntaxError_WithLine(t *testing.T) {
	t.Setenv("NO_COLOR", "1")

	se := &fakeSyntaxError{msg: "'y' cannot start a statement here.", line: 2, col: 1, hint: "Use 'Set y to be ...' instead."}
	got := stacktraces.Render(se)

	if !strings.Contains(got, "Syntax Error") {
		t.Errorf("expected 'Syntax Error' in output, got:\n%s", got)
	}
	if !strings.Contains(got, "line 2") {
		t.Errorf("expected 'line 2' in output, got:\n%s", got)
	}
	if !strings.Contains(got, "column 1") {
		t.Errorf("expected 'column 1' in output, got:\n%s", got)
	}
	if !strings.Contains(got, se.msg) {
		t.Errorf("expected message in output, got:\n%s", got)
	}
	if !strings.Contains(got, se.hint) {
		t.Errorf("expected hint in output, got:\n%s", got)
	}
}

func TestRender_PlainSyntaxError_NoHint(t *testing.T) {
	t.Setenv("NO_COLOR", "1")

	se := &fakeSyntaxError{msg: "unexpected end of file", line: 10, col: 5, hint: ""}
	got := stacktraces.Render(se)

	if !strings.Contains(got, "Syntax Error") {
		t.Errorf("expected 'Syntax Error' in output, got:\n%s", got)
	}
	if strings.Contains(got, "Hint:") {
		t.Errorf("should not show 'Hint:' when hint is empty, got:\n%s", got)
	}
}

func TestRender_PlainSyntaxError_NoLine(t *testing.T) {
	t.Setenv("NO_COLOR", "1")

	se := &fakeSyntaxError{msg: "something is wrong", line: 0, col: 0, hint: ""}
	got := stacktraces.Render(se)

	if !strings.Contains(got, "Syntax Error") {
		t.Errorf("expected 'Syntax Error' in output, got:\n%s", got)
	}
	if strings.Contains(got, "line 0") {
		t.Errorf("should not show 'line 0' when line is unknown, got:\n%s", got)
	}
}

func TestRenderWithColor_SyntaxError(t *testing.T) {
	se := &fakeSyntaxError{
		msg:  "'pi' cannot start a statement here.",
		line: 5,
		col:  3,
		hint: "Use 'Set pi to be ...' instead.",
	}
	got := stacktraces.RenderWithColor(se, true)

	for _, want := range []string{"Syntax Error", "'pi' cannot start", "5", "3", "Set pi"} {
		if !strings.Contains(stripANSI(got), want) {
			t.Errorf("expected %q in coloured syntax error output, got:\n%s", want, got)
		}
	}
	if !strings.Contains(got, "\x1b[") {
		t.Errorf("expected ANSI escape codes in coloured output, got:\n%q", got)
	}
}

func TestRender_SyntaxVsRuntime_DifferentHeaders(t *testing.T) {
	t.Setenv("NO_COLOR", "1")

	se := &fakeSyntaxError{msg: "unexpected token", line: 3, col: 1, hint: ""}
	re := &fakeRuntimeError{msg: "division by zero", stack: []string{}}

	syntaxOut := stacktraces.Render(se)
	runtimeOut := stacktraces.Render(re)

	if !strings.Contains(syntaxOut, "Syntax Error") {
		t.Errorf("expected 'Syntax Error' in syntax output, got:\n%s", syntaxOut)
	}
	if strings.Contains(syntaxOut, "Runtime Error") {
		t.Errorf("syntax output should not contain 'Runtime Error', got:\n%s", syntaxOut)
	}
	if !strings.Contains(runtimeOut, "Runtime Error") {
		t.Errorf("expected 'Runtime Error' in runtime output, got:\n%s", runtimeOut)
	}
	if strings.Contains(runtimeOut, "Syntax Error") {
		t.Errorf("runtime output should not contain 'Syntax Error', got:\n%s", runtimeOut)
	}
}
