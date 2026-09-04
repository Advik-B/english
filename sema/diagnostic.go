// Package sema performs semantic analysis on a parsed English program: name
// resolution, type inference and type checking.
//
// It is the single source of truth for what a well-typed program is. The
// runtime type checks that remain in the execution engines are assertions
// against the same rules, not a second implementation of them.
package sema

import (
	"fmt"
	"sort"

	"github.com/Advik-B/english/ast"
)

// Diagnostic is one problem found during analysis.
//
// It carries a full position rather than just a line, which the previous
// checker could not do: expressions had no position, so every argument-type
// error was reported at "line 0".
type Diagnostic struct {
	Pos ast.Position
	// File is the source file the problem is in. It is empty for the file
	// being compiled and set for problems inside an imported file.
	File    string
	Message string
	// Hint is optional guidance towards a fix.
	Hint string
}

func (d *Diagnostic) Error() string {
	switch {
	case d.File != "" && d.Pos.IsKnown():
		return fmt.Sprintf("TypeError in '%s' at line %d, column %d: %s", d.File, d.Pos.Line, d.Pos.Col, d.Message)
	case d.Pos.IsKnown():
		return fmt.Sprintf("TypeError at line %d, column %d: %s", d.Pos.Line, d.Pos.Col, d.Message)
	case d.File != "":
		return fmt.Sprintf("TypeError in '%s': %s", d.File, d.Message)
	default:
		return "TypeError: " + d.Message
	}
}

// CompileMessage implements stacktraces.CompileError.
func (d *Diagnostic) CompileMessage() string { return d.Message }

// CompileLine implements stacktraces.CompileError.
func (d *Diagnostic) CompileLine() int { return d.Pos.Line }

// CompileCol implements stacktraces.CompileColumnError.
func (d *Diagnostic) CompileCol() int { return d.Pos.Col }

// CompileFile implements stacktraces.CompileFileError.
func (d *Diagnostic) CompileFile() string { return d.File }

// CompileHint implements stacktraces.CompileHintError.
func (d *Diagnostic) CompileHint() string { return d.Hint }

// sortDiagnostics orders problems by position so that the first thing reported
// is the first thing wrong in the file. Analysis visits declarations before
// bodies, so the natural order is not source order.
func sortDiagnostics(ds []*Diagnostic) {
	sort.SliceStable(ds, func(i, j int) bool {
		a, b := ds[i], ds[j]
		if a.File != b.File {
			return a.File < b.File
		}
		if a.Pos.Line != b.Pos.Line {
			return a.Pos.Line < b.Pos.Line
		}
		return a.Pos.Col < b.Pos.Col
	})
}
