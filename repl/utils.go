package repl

import (
	"bytes"
	"errors"
	"io"
	"os"
)

// ErrExit is returned when the user requests to exit the REPL.
var ErrExit = errors.New("exit requested")

// captureStdout captures stdout during a function execution.
// If pipe creation fails, the function is still executed but output is not captured.
func captureStdout(f func()) string {
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		// If pipe creation fails, execute the function without capturing
		// This is acceptable for REPL usage where missing output is better than a crash
		f()
		return ""
	}
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	// Ignore copy errors - partial output is acceptable for REPL usage
	_, _ = io.Copy(&buf, r)
	r.Close()
	return buf.String()
}
