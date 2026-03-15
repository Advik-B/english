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
	"fmt"
	"io"
	"runtime"
	"strings"
	"time"

	vm "english/astvm"
	"english/parser"
	"english/stacktraces"
	"english/stdlib"
)

// Version is the language version string shown in the REPL banner.
// cmd/repl.go overrides this with the value of cmd.Version at startup.
var Version = "1.2.1"

const (
	// PrimaryPrompt is displayed when the REPL is ready for a new statement.
	PrimaryPrompt = ">>> "
	// ContinuationPrompt is displayed when more input is needed to complete a
	// multi-line block.
	ContinuationPrompt = "... "
)

// REPL is an interactive Read-Eval-Print Loop for the English language.
type REPL struct {
	env       *vm.Environment
	evaluator *vm.Evaluator
	in        *bufio.Reader
	out       io.Writer
	useColor  bool
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
		env:       env,
		evaluator: ev,
		in:        bufio.NewReader(in),
		out:       out,
		useColor:  useColor,
	}
}

// Run prints the startup banner and then enters the interactive loop.
// It returns when the user types "exit" / "quit" or when the input reader
// reaches EOF.
func (r *REPL) Run() {
	now := time.Now().Format("Jan 2 2006")
	fmt.Fprintf(r.out, "English %s (%s) on %s\n", Version, now, runtime.GOOS)
	fmt.Fprintf(r.out, "Type \"exit\" to exit, \"help\" for help.\n")
	r.Loop()
}

// Loop runs the REPL input loop without printing the startup banner.
// This is useful for testing or for embedding the REPL in a larger program.
func (r *REPL) Loop() {
	var buffer []string
	depth := 0

	for {
		// Show the appropriate prompt.
		if depth == 0 && len(buffer) == 0 {
			fmt.Fprint(r.out, PrimaryPrompt)
		} else {
			fmt.Fprint(r.out, ContinuationPrompt)
		}

		// Read one line of input.
		line, err := r.in.ReadString('\n')
		if err == io.EOF {
			if line == "" {
				// Clean EOF with no pending input – exit the loop.
				fmt.Fprintln(r.out)
				break
			}
			// Final line without a trailing newline – process it normally.
		} else if err != nil {
			break
		}

		line = strings.TrimRight(line, "\r\n")
		trimmed := strings.TrimSpace(line)

		// ── Special top-level commands (only at the primary prompt) ─────────
		if depth == 0 && len(buffer) == 0 {
			switch trimmed {
			case "exit", "exit.", "quit", "quit.":
				return
			case "help", "help.":
				r.printHelp()
				continue
			case "":
				continue
			}
		}

		// ── Ignore blank lines outside of a block ────────────────────────────
		if trimmed == "" {
			if depth > 0 {
				buffer = append(buffer, line)
			}
			continue
		}

		lower := strings.ToLower(trimmed)

		// ── Update block depth ───────────────────────────────────────────────
		if isBlockOpener(lower) {
			depth++
		}
		if isBlockCloser(lower) {
			depth--
			if depth < 0 {
				depth = 0
			}
		}

		buffer = append(buffer, line)

		// ── Execute when the block (or single statement) is complete ─────────
		if depth <= 0 && len(buffer) > 0 {
			depth = 0
			code := strings.Join(buffer, "\n")
			buffer = nil
			r.execute(code)
		}
	}
}

// ── Block-depth helpers ──────────────────────────────────────────────────────

// isBlockOpener reports whether a trimmed, lower-cased line opens a new block.
//
// Block-openers:
//   - Any line that contains "following" and ends with ":" covers all forms:
//     "do the following:", "does the following:",
//     "repeat the following while …:", "Try doing the following:", etc.
//   - Any line that ends with " then" (covers "If …, then")
//
// Exception: lines that start with "otherwise" are continuations of an
// existing if-else chain and do not open a new depth level, even when they
// themselves end with " then" (e.g. "otherwise if …, then").
// Similarly, "on ErrorType:" and "but finally:" are catch/finally clauses
// inside an existing try block and do not affect depth.
func isBlockOpener(lower string) bool {
	// Continuation branches never open a new block level.
	if strings.HasPrefix(lower, "otherwise") {
		return false
	}
	if strings.HasPrefix(lower, "on ") && strings.HasSuffix(lower, ":") {
		return false
	}
	if strings.HasPrefix(lower, "but finally") {
		return false
	}

	// "do the following:", "repeat the following while …:", etc.
	if strings.Contains(lower, "following") && strings.HasSuffix(lower, ":") {
		return true
	}
	if strings.HasSuffix(lower, " then") {
		return true
	}
	return false
}

// isBlockCloser reports whether a trimmed, lower-cased line closes a block.
func isBlockCloser(lower string) bool {
	return strings.Contains(lower, "thats it.") || strings.Contains(lower, "that's it.")
}

// ── Execution ────────────────────────────────────────────────────────────────

// execute parses code and evaluates it.  Any output produced by Print
// statements is written directly to r.out via the evaluator's output writer.
// Parse and runtime errors are rendered and written to r.out as well.
func (r *REPL) execute(code string) {
	// Parse
	lexer := parser.NewLexer(code)
	tokens := lexer.TokenizeAll()
	p := parser.NewParser(tokens)
	program, parseErr := p.Parse()
	if parseErr != nil {
		fmt.Fprint(r.out, stacktraces.RenderWithColor(parseErr, r.useColor))
		return
	}

	// Evaluate – Print output goes directly to r.out via ev.out.
	_, execErr := r.evaluator.Eval(program)
	if execErr != nil {
		fmt.Fprint(r.out, stacktraces.RenderWithColor(execErr, r.useColor))
	}
}

// ── Help ─────────────────────────────────────────────────────────────────────

func (r *REPL) printHelp() {
	fmt.Fprintln(r.out, "English REPL")
	fmt.Fprintln(r.out, strings.Repeat("─", 50))
	fmt.Fprintln(r.out, "Commands:")
	fmt.Fprintln(r.out, "  exit / quit     Exit the REPL")
	fmt.Fprintln(r.out, "  help            Show this help message")
	fmt.Fprintln(r.out, "")
	fmt.Fprintln(r.out, "Language tips:")
	fmt.Fprintln(r.out, "  · Statements must end with a period (.)")
	fmt.Fprintln(r.out, "  · Multi-line blocks open with 'do the following:'")
	fmt.Fprintln(r.out, "    and close with 'thats it.'")
	fmt.Fprintln(r.out, "  · If/else: 'If …, then … otherwise … thats it.'")
	fmt.Fprintln(r.out, "  · Loops:   'repeat the following while …: … thats it.'")
	fmt.Fprintln(r.out, "")
	fmt.Fprintln(r.out, "Examples:")
	fmt.Fprintln(r.out, "  >>> Declare x to be 5.")
	fmt.Fprintln(r.out, "  >>> Print the value of x.")
	fmt.Fprintln(r.out, "  5")
	fmt.Fprintln(r.out, "  >>> For each n in [1, 2, 3], do the following:")
	fmt.Fprintln(r.out, "  ...     Print the value of n.")
	fmt.Fprintln(r.out, "  ... thats it.")
}
