package repl

import (
	"fmt"
	"io"
	"runtime"
	"strings"
	"time"

	"github.com/Advik-B/english/highlight"
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
		// Show the appropriate prompt and remember which one was printed so
		// we can redraw the line with syntax highlighting after input arrives.
		var prompt string
		if depth == 0 && len(buffer) == 0 {
			prompt = PrimaryPrompt
		} else {
			prompt = ContinuationPrompt
		}
		fmt.Fprint(r.out, prompt)

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

		// When color is enabled, overwrite the just-typed line with a
		// syntax-highlighted version.  The sequence:
		//   \033[1A  – move cursor up one line (to the line the user typed on)
		//   \r       – carriage-return to the beginning of that line
		//   \033[2K  – erase the entire line
		// …then we reprint prompt + highlighted code and move to the next line.
		if r.useColor && line != "" {
			fmt.Fprintf(r.out, "\033[1A\r\033[2K%s%s\n",
				prompt, highlight.Highlight(line, true))
		}

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
