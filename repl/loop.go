package repl

import (
	"fmt"
	"io"
	"runtime"
	"strings"
	"time"
)

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
