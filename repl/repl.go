// Package repl provides a Python-inspired Read-Eval-Print Loop for the
// English programming language.
//
// The REPL reads one statement at a time from an [io.Reader], evaluates it
// using the AST VM, and writes output to an [io.Writer].  Because all I/O is
// injected the loop is trivially testable without a real terminal.
//
// Primary prompt:      ">>> "
// Continuation prompt: "... "
//
// Multi-line blocks are detected by counting block-openers ("following:", "if
// …, then") and the corresponding block-closer ("thats it.").  When the depth
// is greater than zero the continuation prompt is shown until the block is
// complete.
package repl

import (
	"bufio"
	"io"

	vm "github.com/Advik-B/english/astvm"
	"github.com/Advik-B/english/help"
	"github.com/Advik-B/english/stdlib"
)

// Version is the language version string shown in the REPL banner.
// cmd/repl.go overrides this with the value of version.Version at startup.

const (
	// PrimaryPrompt is displayed when the REPL is ready for a new statement.
	PrimaryPrompt = ">>> "
	// ContinuationPrompt is displayed when more input is needed to complete a
	// multi-line block.
	ContinuationPrompt = "... "
)

// REPL is an interactive Read-Eval-Print Loop for the English language.
type REPL struct {
	env          *vm.Environment
	evaluator    *vm.Evaluator
	in           *bufio.Reader
	out          io.Writer
	useColor     bool
	helpRegistry *help.Registry
}

// New creates a REPL that reads from in and writes all output (prompts,
// program output, and error messages) to out.
//
// When useColor is true, error messages are rendered with ANSI colour codes.
// Pass stacktraces.HasColor() for automatic TTY detection.
func New(in io.Reader, out io.Writer, useColor bool) *REPL {
	env := vm.NewEnvironment()
	stdlib.Register(env)
	ev := vm.NewEvaluator(env, stdlib.Eval)
	ev.SetOutput(out)
	return &REPL{
		env:          env,
		evaluator:    ev,
		in:           bufio.NewReader(in),
		out:          out,
		useColor:     useColor,
		helpRegistry: help.NewRegistry(),
	}
}
