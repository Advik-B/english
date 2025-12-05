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
func captureStdout(f func()) string {
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		return ""
	}
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}
