package stacktraces_test

import (
	"errors"
	"strings"
	"testing"

	"english/stacktraces"
)

// fakeRuntimeError is a test double that satisfies stacktraces.RuntimeError.
type fakeRuntimeError struct {
	msg   string
	stack []string
}

func (e *fakeRuntimeError) Error() string          { return "Runtime Error: " + e.msg }
func (e *fakeRuntimeError) RuntimeMessage() string { return e.msg }
func (e *fakeRuntimeError) RuntimeCallStack() []string {
	if e.stack == nil {
		return []string{}
	}
	return e.stack
}

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
