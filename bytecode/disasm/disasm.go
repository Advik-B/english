// Package disasm – disasm.go
//
// Provides Disassemble(), which walks a decoded *ast.Program and produces a
// colourised, instruction-level listing of a .101 bytecode file – similar to
// the output of Python's `dis` module.  The output never shows English source
// code; it shows raw opcodes and their operands so that compiled programs can
// be inspected without running or re-transpiling them.
//
// The package is split into four files:
//
//	disasm.go      – public API, disassembler struct, core helpers, run()
//	statements.go  – stmt() handler (all statement AST node types)
//	expressions.go – expr() handler (all expression AST node types)
//	styles.go      – lipgloss renderer and colour-style variables
package disasm

import (
	"fmt"
	"path/filepath"
	"strings"

	"english/ast"
	"english/bytecode"

	"github.com/charmbracelet/lipgloss"
)

// ─── Public API ───────────────────────────────────────────────────────────────

// Disassemble produces a colorised, instruction-level listing of a decoded
// .101 bytecode program.  filename is used only for the header line.
// When useColor is false the output contains no ANSI escape codes.
// When friendlyOps is true, comparison/logical operators are shown as English
// prose (e.g. "is less than or equal to") instead of their symbolic equivalents
// (e.g. "<=").  friendlyOps has no effect on arithmetic operators (+, -, *, /,
// %) which are always shown symbolically.
func Disassemble(program *ast.Program, filename string, useColor, friendlyOps bool) string {
	d := &disassembler{useColor: useColor, friendlyOps: friendlyOps}
	return d.run(program, filename)
}

// ─── Internal disassembler ────────────────────────────────────────────────────

type disassembler struct {
	useColor    bool
	friendlyOps bool
	counter     int      // running instruction index
	depth       int      // indentation depth (increases inside bodies)
	out         []string // accumulated output lines
}

// s applies a lipgloss style only when colour is enabled.
func (d *disassembler) s(style lipgloss.Style, text string) string {
	if d.useColor {
		return style.Render(text)
	}
	return text
}

// opStr returns the display form of an operator.
// When friendlyOps is true the original English-prose form is kept
// (e.g. "is less than or equal to"); otherwise symOp() converts it to a
// conventional symbol (e.g. "<=").  Arithmetic operators (+, -, *, /, %)
// are always returned unchanged by symOp, so they are never affected.
func (d *disassembler) opStr(op string) string {
	if d.friendlyOps {
		return op
	}
	return symOp(op)
}

// emit appends a formatted instruction line to the output.
// opcodeStyle selects the colour for the opcode name.
// operands is the (already-rendered) operand string – may be empty.
func (d *disassembler) emit(opcodeStyle lipgloss.Style, opcode, operands string) {
	idx := d.s(styleIdx, fmt.Sprintf("%4d", d.counter))
	d.counter++
	indent := strings.Repeat("    ", d.depth)
	op := d.s(opcodeStyle, fmt.Sprintf("%-18s", opcode))
	// Index is always at column 0; indentation comes after the "idx  " prefix.
	line := idx + "  " + indent + op
	if operands != "" {
		line += "  " + operands
	}
	d.out = append(d.out, line)
}

// emitLabel appends a line that is not counted as an instruction (e.g. END_*).
func (d *disassembler) emitLabel(style lipgloss.Style, label string, extra ...string) {
	indent := strings.Repeat("    ", d.depth)
	// Six spaces occupy the same width as "NNN  " so the opcode column stays aligned.
	// Indentation comes after that placeholder, mirroring the emit layout.
	text := "      " + indent + d.s(style, label)
	if len(extra) > 0 && extra[0] != "" {
		text += "  " + extra[0]
	}
	d.out = append(d.out, text)
}

// ─── Top-level ────────────────────────────────────────────────────────────────

func (d *disassembler) run(program *ast.Program, filename string) string {
	n := len(program.Statements)
	noun := "statements"
	if n == 1 {
		noun = "statement"
	}
	base := filepath.Base(filename)

	d.out = append(d.out,
		d.s(styleHeaderTitle, "=== Disassembly of "+base+" ===")+
			"  "+d.s(styleHeaderMeta, fmt.Sprintf("(format v%d · %d %s)", bytecode.FormatVersion, n, noun)),
		"",
	)

	for _, stmt := range program.Statements {
		d.stmt(stmt)
	}

	return strings.Join(d.out, "\n") + "\n"
}
