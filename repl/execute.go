package repl

import (
	"fmt"

	"github.com/Advik-B/english/parser"
	"github.com/Advik-B/english/stacktraces"
)

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
