package repl

import (
	"fmt"
	"strings"
)

// printHelp writes a help summary to the REPL output writer.
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
