package vm

import "fmt"

// TypeError is a type problem discovered while executing.
//
// Semantic analysis catches these before a program runs, so reaching one at
// run time means the analyser could not prove the problem statically: a
// declaration inside an imported file the analyser could not read, for
// example. It is reported as a compile error because that is what it is.
type TypeError struct {
	Line    int
	Message string
	// File is the source file in which the error occurred. It is empty for the
	// main file and populated for errors inside an imported file.
	File string
}

func (te *TypeError) Error() string {
	if te.File != "" && te.Line > 0 {
		return fmt.Sprintf("TypeError in '%s' at line %d: %s", te.File, te.Line, te.Message)
	}
	if te.Line > 0 {
		return fmt.Sprintf("TypeError at line %d: %s", te.Line, te.Message)
	}
	return fmt.Sprintf("TypeError: %s", te.Message)
}

// CompileMessage implements stacktraces.CompileError.
func (te *TypeError) CompileMessage() string { return te.Message }

// CompileLine implements stacktraces.CompileError.
func (te *TypeError) CompileLine() int { return te.Line }

// CompileFile implements stacktraces.CompileFileError.
func (te *TypeError) CompileFile() string { return te.File }
