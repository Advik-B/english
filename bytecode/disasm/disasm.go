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
	"os"
	"path/filepath"
	"strings"

	"github.com/Advik-B/english/ast"
	"github.com/Advik-B/english/bytecode"
	"github.com/Advik-B/english/parser"

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
//
// importDepth controls how many levels of imported files are also disassembled
// and appended to the output.  Pass 0 (the default) to show only the top-level
// file.  Pass -1 for unlimited depth.
//
// unrollDepth controls how many levels of nested function-call arguments are
// extracted into temporary declarations to improve readability.  Pass 0 (the
// default) for no unrolling.  Pass -1 for fully recursive unrolling.
func Disassemble(program *ast.Program, filename string, useColor, friendlyOps bool, importDepth, unrollDepth int) string {
	d := &disassembler{
		useColor:    useColor,
		friendlyOps: friendlyOps,
		unrollDepth: unrollDepth,
	}
	return d.run(program, filename, importDepth, nil)
}

// ─── Internal disassembler ────────────────────────────────────────────────────

type disassembler struct {
	useColor    bool
	friendlyOps bool
	unrollDepth int
	counter     int      // running instruction index
	depth       int      // indentation depth (increases inside bodies)
	tmpCounter  int      // counter for generated temporary variable names
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

// run disassembles one program, then optionally recurses into imported files.
// seenFiles guards against circular imports; it is nil on the first call.
func (d *disassembler) run(program *ast.Program, filename string, importDepth int, seenFiles map[string]bool) string {
	if seenFiles == nil {
		seenFiles = make(map[string]bool)
	}
	abs, err := filepath.Abs(filename)
	if err != nil {
		abs = filename // fall back to the raw path for deduplication
	}
	seenFiles[abs] = true

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
		d.stmtMaybeUnroll(stmt)
	}

	// Follow imports when requested.
	if importDepth != 0 {
		for _, stmt := range program.Statements {
			imp, ok := stmt.(*ast.ImportStatement)
			if !ok {
				continue
			}
			importPath := imp.Path
			absImport, err := filepath.Abs(importPath)
			if err != nil {
				absImport = importPath // fall back to the raw path
			}
			if seenFiles[absImport] {
				continue
			}
			nextDepth := importDepth
			if importDepth > 0 {
				nextDepth = importDepth - 1
			}
			importedProg, resolvedPath, loadErr := loadImportedFile(importPath)
			if loadErr != nil {
				d.out = append(d.out,
					"",
					d.s(styleHeaderMeta, fmt.Sprintf("  [import: could not load '%s': %v]", imp.Path, loadErr)),
				)
				continue
			}
			d.out = append(d.out, "")
			d.run(importedProg, resolvedPath, nextDepth, seenFiles)
		}
	}

	return strings.Join(d.out, "\n") + "\n"
}

// loadImportedFile tries to load an imported .abc file.
// It first checks for a valid bytecode cache (.101); if the cache is current it
// decodes that.  Otherwise it falls back to parsing the source file directly.
// It returns the decoded program and the resolved file path used to load it.
func loadImportedFile(path string) (*ast.Program, string, error) {
	// Try the bytecode cache first.
	cachePath := bytecode.GetCachePath(path)
	if bytecode.IsCacheValid(path, cachePath) {
		data, err := bytecode.ReadBytecodeCache(cachePath)
		if err == nil {
			dec := bytecode.NewDecoder(data)
			prog, err := dec.Decode()
			if err == nil {
				return prog, cachePath, nil
			}
		}
	}

	// Fall back to parsing the source file.
	src, err := os.ReadFile(path)
	if err != nil {
		return nil, "", err
	}
	lx := parser.NewLexer(string(src))
	tokens := lx.TokenizeAll()
	p := parser.NewParser(tokens)
	prog, err := p.Parse()
	if err != nil {
		return nil, "", err
	}
	return prog, path, nil
}

// ─── Unrolling ────────────────────────────────────────────────────────────────

// stmtMaybeUnroll emits a statement, applying expression unrolling first when
// unrollDepth != 0.  When unrollDepth is 0 it falls through directly to stmt().
//
// The top-level expression in each statement is never extracted into a
// temporary – only its nested function-call arguments are.  This preserves the
// original variable/assignment name on the final emitted line, e.g.
//
//	Declare x to be f1(f2(f3(0))).   --unroll-depth 1
//	  →  __tmp0 = f2(f3(0))
//	  →  x = f1(__tmp0)
func (d *disassembler) stmtMaybeUnroll(node ast.Statement) {
	if d.unrollDepth == 0 {
		d.stmt(node)
		return
	}
	// For statement types that contain a top-level value expression, extract
	// nested function calls from the *arguments* of the outermost call into
	// temporary declarations first.
	switch s := node.(type) {
	case *ast.VariableDecl:
		extras, newVal := d.unrollTopExpr(s.Value)
		for _, tmp := range extras {
			d.stmt(tmp)
		}
		d.stmt(&ast.VariableDecl{Name: s.Name, Value: newVal, IsConstant: s.IsConstant})
	case *ast.TypedVariableDecl:
		extras, newVal := d.unrollTopExpr(s.Value)
		for _, tmp := range extras {
			d.stmt(tmp)
		}
		d.stmt(&ast.TypedVariableDecl{Name: s.Name, TypeName: s.TypeName, Value: newVal, IsConstant: s.IsConstant})
	case *ast.Assignment:
		extras, newVal := d.unrollTopExpr(s.Value)
		for _, tmp := range extras {
			d.stmt(tmp)
		}
		d.stmt(&ast.Assignment{Name: s.Name, Value: newVal})
	case *ast.ReturnStatement:
		extras, newVal := d.unrollTopExpr(s.Value)
		for _, tmp := range extras {
			d.stmt(tmp)
		}
		d.stmt(&ast.ReturnStatement{Value: newVal})
	case *ast.OutputStatement:
		var allExtras []ast.Statement
		newVals := make([]ast.Expression, len(s.Values))
		for i, v := range s.Values {
			extras, newV := d.unrollTopExpr(v)
			allExtras = append(allExtras, extras...)
			newVals[i] = newV
		}
		for _, tmp := range allExtras {
			d.stmt(tmp)
		}
		d.stmt(&ast.OutputStatement{Values: newVals, Newline: s.Newline})
	default:
		d.stmt(node)
	}
}

// unrollTopExpr processes the top-level expression for unrolling.
// If the expression is a FunctionCall its arguments are extracted via
// extractFuncCalls; the outer call itself is kept intact so the original
// statement's variable name is preserved.  Non-FunctionCall expressions are
// returned unchanged.
func (d *disassembler) unrollTopExpr(expr ast.Expression) ([]ast.Statement, ast.Expression) {
	fc, ok := expr.(*ast.FunctionCall)
	if !ok {
		return nil, expr
	}
	var allExtras []ast.Statement
	newArgs := make([]ast.Expression, len(fc.Arguments))
	for i, arg := range fc.Arguments {
		extras, newArg := d.extractFuncCalls(arg, d.unrollDepth)
		allExtras = append(allExtras, extras...)
		newArgs[i] = newArg
	}
	return allExtras, &ast.FunctionCall{Name: fc.Name, Arguments: newArgs}
}

// extractFuncCalls recursively extracts a nested function-call expression into
// temporary VariableDecl statements.  It returns the list of generated
// declarations (in order) and an Identifier referencing the final temporary.
//
// depth controls how many levels to extract: 1 wraps only this call, 2 also
// wraps the function-call arguments, and so on.  -1 means fully recursive.
// depth==0 is a no-op (the expression is returned as-is).
func (d *disassembler) extractFuncCalls(expr ast.Expression, depth int) ([]ast.Statement, ast.Expression) {
	if depth == 0 {
		return nil, expr
	}
	fc, ok := expr.(*ast.FunctionCall)
	if !ok {
		return nil, expr
	}

	// Recurse into each argument first so inner calls are emitted before outer.
	nextDepth := depth
	if depth > 0 {
		nextDepth = depth - 1
	}
	var allExtras []ast.Statement
	newArgs := make([]ast.Expression, len(fc.Arguments))
	for i, arg := range fc.Arguments {
		extras, newArg := d.extractFuncCalls(arg, nextDepth)
		allExtras = append(allExtras, extras...)
		newArgs[i] = newArg
	}

	// Create a temporary for this function call.
	tmpName := fmt.Sprintf("__tmp%d", d.tmpCounter)
	d.tmpCounter++
	newCall := &ast.FunctionCall{Name: fc.Name, Arguments: newArgs}
	allExtras = append(allExtras, &ast.VariableDecl{Name: tmpName, Value: newCall})

	return allExtras, &ast.Identifier{Name: tmpName}
}
