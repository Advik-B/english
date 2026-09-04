package repl

import (
	"fmt"
	"os"
	"runtime/debug"

	"github.com/Advik-B/english/parser"
	"github.com/Advik-B/english/stacktraces"
)

// execute parses code and evaluates it.  Any output produced by Print
// statements is written directly to r.out via the evaluator's output writer.
// Parse and runtime errors are rendered and written to r.out as well.
//
// A panic anywhere in the parser or evaluator is contained here so that an
// interpreter bug ends the current statement rather than the whole interactive
// session, which would discard everything the user has defined so far.
func (r *REPL) execute(code string) {
	defer func() {
		if rec := recover(); rec != nil {
			fmt.Fprintf(r.out, "Internal error: %v\n", rec)
			fmt.Fprintf(r.out, "This is a bug in the interpreter, not in your program.\n")
			fmt.Fprintf(r.out, "Your session is still alive; previously defined names are intact.\n")
			// Set ENGLISH_DEBUG to see the Go stack behind an internal error.
			if os.Getenv("ENGLISH_DEBUG") != "" {
				fmt.Fprintf(r.out, "\n%s\n", debug.Stack())
			}
		}
	}()

	// Parse
	lexer := parser.NewLexer(code)
	tokens := lexer.TokenizeAll()
	p := parser.NewParser(tokens)
	program, parseErr := p.Parse()
	if parseErr != nil {
		fmt.Fprint(r.out, stacktraces.RenderWithColor(parseErr, r.useColor))
		return
	}

	// Analyse before evaluating, so the REPL applies the same rules as
	// `english run`. It previously skipped analysis entirely, so a program
	// the compiler rejects was accepted here.
	if diags := r.analyzer.CheckNext(program); len(diags) > 0 {
		for _, d := range diags {
			fmt.Fprint(r.out, stacktraces.RenderWithColor(d, r.useColor))
		}
		return
	}

	// Evaluate – Print output goes directly to r.out via ev.out.
	_, execErr := r.evaluator.Eval(program)
	if execErr != nil {
		fmt.Fprint(r.out, stacktraces.RenderWithColor(execErr, r.useColor))
	}
}
