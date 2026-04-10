package cmd

import (
	"os"

	"github.com/Advik-B/english/repl"
	"github.com/Advik-B/english/stacktraces"
)

// StartREPL starts the interactive Read-Eval-Print Loop using the repl package.
// Color is automatically enabled when the terminal supports ANSI codes (TTY,
// no NO_COLOR env var).
func StartREPL() {
	r := repl.New(os.Stdin, os.Stdout, stacktraces.HasColor())
	r.Run()
}
