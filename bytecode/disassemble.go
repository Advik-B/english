// Package bytecode – disassemble.go
//
// Provides Disassemble(), which walks a decoded *ast.Program and produces a
// colourised, instruction-level listing of a .101 bytecode file – similar to
// the output of Python's `dis` module.  The output never shows English source
// code; it shows raw opcodes and their operands so that compiled programs can
// be inspected without running or re-transpiling them.
package bytecode

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"english/ast"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

// ─── Colour palette (Dracula-inspired, consistent with highlight package) ─────

var disasmRenderer = func() *lipgloss.Renderer {
	r := lipgloss.NewRenderer(os.Stdout)
	r.SetColorProfile(termenv.TrueColor)
	return r
}()

var (
	// Instruction-name styles – grouped by category so the eye can quickly
	// distinguish declarations, control flow, I/O, assignments, and closers.

	styleOpcodeDecl    = disasmRenderer.NewStyle().Foreground(lipgloss.Color("#BD93F9")).Bold(true) // purple  – DECLARE_VAR / FUNC_DECL / IMPORT
	styleOpcodeControl = disasmRenderer.NewStyle().Foreground(lipgloss.Color("#FF79C6")).Bold(true) // pink    – IF / WHILE / FOR / RETURN / BREAK
	styleOpcodeIO      = disasmRenderer.NewStyle().Foreground(lipgloss.Color("#50FA7B")).Bold(true) // green   – OUTPUT_PRINT / OUTPUT_WRITE
	styleOpcodeAssign  = disasmRenderer.NewStyle().Foreground(lipgloss.Color("#FFB86C")).Bold(true) // orange  – ASSIGN / INDEX_ASSIGN / TOGGLE
	styleOpcodeCall    = disasmRenderer.NewStyle().Foreground(lipgloss.Color("#8BE9FD")).Bold(true) // cyan    – CALL
	styleOpcodeEnd     = disasmRenderer.NewStyle().Foreground(lipgloss.Color("#6272A4"))            // muted   – END_* markers

	// Operand styles
	styleIdx    = disasmRenderer.NewStyle().Foreground(lipgloss.Color("#6272A4"))            // muted   – instruction index numbers
	styleIdent  = disasmRenderer.NewStyle().Foreground(lipgloss.Color("#F8F8F2"))            // white   – variable / parameter names
	styleLabel  = disasmRenderer.NewStyle().Foreground(lipgloss.Color("#8BE9FD")).Italic(true) // cyan italic – function / import labels
	styleStr    = disasmRenderer.NewStyle().Foreground(lipgloss.Color("#50FA7B"))            // green   – string literals
	styleNum    = disasmRenderer.NewStyle().Foreground(lipgloss.Color("#F1FA8C"))            // yellow  – number literals
	styleBool   = disasmRenderer.NewStyle().Foreground(lipgloss.Color("#FFB86C"))            // orange  – boolean literals
	styleNull   = disasmRenderer.NewStyle().Foreground(lipgloss.Color("#6272A4")).Italic(true) // muted italic – null / nothing
	styleOp     = disasmRenderer.NewStyle().Foreground(lipgloss.Color("#FF5555"))            // red     – operators
	styleArrow  = disasmRenderer.NewStyle().Foreground(lipgloss.Color("#FF79C6"))            // pink    – ← assignment arrow
	stylePunct  = disasmRenderer.NewStyle().Foreground(lipgloss.Color("#6272A4"))            // muted   – ( ) [ ] ,
	styleConst  = disasmRenderer.NewStyle().Foreground(lipgloss.Color("#FFD700"))            // gold    – [const] tag
	styleMeta   = disasmRenderer.NewStyle().Foreground(lipgloss.Color("#6272A4")).Italic(true) // muted italic – type tags, metadata
	styleType   = disasmRenderer.NewStyle().Foreground(lipgloss.Color("#8BE9FD")).Italic(true) // cyan italic – type annotations

	// Header
	styleHeaderTitle = disasmRenderer.NewStyle().Foreground(lipgloss.Color("#BD93F9")).Bold(true)
	styleHeaderMeta  = disasmRenderer.NewStyle().Foreground(lipgloss.Color("#6272A4")).Italic(true)
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
	counter     int    // running instruction index
	depth       int    // indentation depth (increases inside bodies)
	out         []string
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
			"  "+d.s(styleHeaderMeta, fmt.Sprintf("(format v%d · %d %s)", FormatVersion, n, noun)),
		"",
	)

	for _, stmt := range program.Statements {
		d.stmt(stmt)
	}

	return strings.Join(d.out, "\n") + "\n"
}

// ─── Statement handlers ───────────────────────────────────────────────────────

func (d *disassembler) stmt(node ast.Statement) {
	switch s := node.(type) {

	case *ast.VariableDecl:
		name := d.s(styleIdent, s.Name)
		constTag := ""
		if s.IsConstant {
			constTag = " " + d.s(styleConst, "[const]")
		}
		arrow := d.s(styleArrow, "←")
		d.emit(styleOpcodeDecl, "DECLARE_VAR",
			name+constTag+"  "+arrow+"  "+d.expr(s.Value))

	case *ast.TypedVariableDecl:
		name := d.s(styleIdent, s.Name)
		typeTag := d.s(styleType, ":"+s.TypeName)
		constTag := ""
		if s.IsConstant {
			constTag = " " + d.s(styleConst, "[const]")
		}
		arrow := d.s(styleArrow, "←")
		d.emit(styleOpcodeDecl, "DECLARE_VAR",
			name+typeTag+constTag+"  "+arrow+"  "+d.expr(s.Value))

	case *ast.ErrorTypeDecl:
		name := d.s(styleIdent, s.Name)
		parentPart := ""
		if s.ParentType != "" {
			parentPart = "  " + d.s(styleOp, "extends") + "  " + d.s(styleIdent, s.ParentType)
		}
		d.emit(styleOpcodeDecl, "DECL_ERROR_TYPE", name+parentPart)

	case *ast.Assignment:
		name := d.s(styleIdent, s.Name)
		arrow := d.s(styleArrow, "←")
		d.emit(styleOpcodeAssign, "ASSIGN", name+"  "+arrow+"  "+d.expr(s.Value))

	case *ast.FunctionDecl:
		params := make([]string, len(s.Parameters))
		for i, p := range s.Parameters {
			params[i] = d.s(styleIdent, p)
		}
		paramStr := d.s(stylePunct, "(") +
			strings.Join(params, d.s(stylePunct, ", ")) +
			d.s(stylePunct, ")")
		d.emit(styleOpcodeDecl, "FUNC_DECL",
			d.s(styleLabel, s.Name)+"  "+paramStr)
		d.depth++
		for _, child := range s.Body {
			d.stmt(child)
		}
		d.depth--
		d.emitLabel(styleOpcodeEnd,
			fmt.Sprintf("%-18s", "END_FUNC"),
			d.s(styleMeta, s.Name))

	case *ast.CallStatement:
		if s.FunctionCall != nil {
			d.emit(styleOpcodeCall, "CALL",
				d.s(styleLabel, s.FunctionCall.Name)+d.argList(s.FunctionCall.Arguments))
		}

	case *ast.IfStatement:
		d.emit(styleOpcodeControl, "IF", d.expr(s.Condition))
		d.depth++
		for _, child := range s.Then {
			d.stmt(child)
		}
		d.depth--
		for _, ei := range s.ElseIf {
			d.emitLabel(styleOpcodeControl, fmt.Sprintf("%-18s", "ELSE_IF"), d.expr(ei.Condition))
			d.depth++
			for _, child := range ei.Body {
				d.stmt(child)
			}
			d.depth--
		}
		if len(s.Else) > 0 {
			d.emitLabel(styleOpcodeControl, fmt.Sprintf("%-18s", "ELSE"), "")
			d.depth++
			for _, child := range s.Else {
				d.stmt(child)
			}
			d.depth--
		}
		d.emitLabel(styleOpcodeEnd, fmt.Sprintf("%-18s", "END_IF"), "")

	case *ast.WhileLoop:
		d.emit(styleOpcodeControl, "WHILE", d.expr(s.Condition))
		d.depth++
		for _, child := range s.Body {
			d.stmt(child)
		}
		d.depth--
		d.emitLabel(styleOpcodeEnd, fmt.Sprintf("%-18s", "END_WHILE"), "")

	case *ast.ForLoop:
		d.emit(styleOpcodeControl, "FOR_LOOP", d.expr(s.Count)+"  "+d.s(styleMeta, "times"))
		d.depth++
		for _, child := range s.Body {
			d.stmt(child)
		}
		d.depth--
		d.emitLabel(styleOpcodeEnd, fmt.Sprintf("%-18s", "END_FOR_LOOP"), "")

	case *ast.ForEachLoop:
		item := d.s(styleIdent, s.Item)
		list := d.expr(s.List)
		d.emit(styleOpcodeControl, "FOR_EACH",
			item+"  "+d.s(styleOp, "in")+"  "+list)
		d.depth++
		for _, child := range s.Body {
			d.stmt(child)
		}
		d.depth--
		d.emitLabel(styleOpcodeEnd, fmt.Sprintf("%-18s", "END_FOR_EACH"), "")

	case *ast.IndexAssignment:
		listName := d.s(styleIdent, s.ListName)
		idxPart := d.s(stylePunct, "[") + d.expr(s.Index) + d.s(stylePunct, "]")
		arrow := d.s(styleArrow, "←")
		d.emit(styleOpcodeAssign, "INDEX_ASSIGN",
			listName+idxPart+"  "+arrow+"  "+d.expr(s.Value))

	case *ast.ReturnStatement:
		d.emit(styleOpcodeControl, "RETURN", d.expr(s.Value))

	case *ast.OutputStatement:
		opcode := "OUTPUT_PRINT"
		if !s.Newline {
			opcode = "OUTPUT_WRITE"
		}
		vals := make([]string, len(s.Values))
		for i, v := range s.Values {
			vals[i] = d.expr(v)
		}
		d.emit(styleOpcodeIO, opcode,
			strings.Join(vals, d.s(stylePunct, ", ")))

	case *ast.ToggleStatement:
		d.emit(styleOpcodeAssign, "TOGGLE", d.s(styleIdent, s.Name))

	case *ast.BreakStatement:
		d.emit(styleOpcodeControl, "BREAK", "")

	case *ast.ImportStatement:
		path := d.s(styleStr, `"`+s.Path+`"`)
		detail := ""
		switch {
		case s.ImportAll:
			detail = "  " + d.s(styleMeta, "(import all)")
		case len(s.Items) > 0:
			items := make([]string, len(s.Items))
			for i, it := range s.Items {
				items[i] = d.s(styleIdent, it)
			}
			detail = "  " + d.s(stylePunct, "(") +
				strings.Join(items, d.s(stylePunct, ", ")) +
				d.s(stylePunct, ")")
		}
		if s.IsSafe {
			detail += "  " + d.s(styleConst, "[safe]")
		}
		d.emit(styleOpcodeDecl, "IMPORT", path+detail)
	}
}

// ─── Operator symbol mapping ─────────────────────────────────────────────────

// symOp converts English-phrased operator names (as stored in the AST) to their
// conventional symbolic equivalents so that the disassembly does not contain
// English prose.
func symOp(op string) string {
	switch op {
	case "is equal to":
		return "=="
	case "is not equal to":
		return "!="
	case "is less than":
		return "<"
	case "is greater than":
		return ">"
	case "is less than or equal to":
		return "<="
	case "is greater than or equal to":
		return ">="
	case "and":
		return "&&"
	case "or":
		return "||"
	case "not":
		return "!"
	default:
		// +, -, *, /, % and any future operators pass through unchanged.
		return op
	}
}

// ─── Expression renderer ──────────────────────────────────────────────────────

// expr renders an expression as a compact, operator-notation string.
// It is intentionally brief because expressions appear as operands on a single
// disassembly line; complex sub-expressions are shown in full but stay inline.
func (d *disassembler) expr(node ast.Expression) string {
	if node == nil {
		return d.s(styleNull, "null")
	}
	switch ex := node.(type) {

	case *ast.NumberLiteral:
		return d.s(styleNum, fmt.Sprintf("%g", ex.Value))

	case *ast.StringLiteral:
		return d.s(styleStr, `"`+ex.Value+`"`)

	case *ast.BooleanLiteral:
		if ex.Value {
			return d.s(styleBool, "true")
		}
		return d.s(styleBool, "false")

	case *ast.NothingLiteral:
		return d.s(styleNull, "null")

	case *ast.ListLiteral:
		elems := make([]string, len(ex.Elements))
		for i, e := range ex.Elements {
			elems[i] = d.expr(e)
		}
		return d.s(stylePunct, "[") +
			strings.Join(elems, d.s(stylePunct, ", ")) +
			d.s(stylePunct, "]")

	case *ast.Identifier:
		return d.s(styleIdent, ex.Name)

	case *ast.BinaryExpression:
		left := d.expr(ex.Left)
		right := d.expr(ex.Right)
		op := d.s(styleOp, d.opStr(ex.Operator))
		return d.s(stylePunct, "(") + left + " " + op + " " + right + d.s(stylePunct, ")")

	case *ast.UnaryExpression:
		return d.s(styleOp, d.opStr(ex.Operator)) + d.expr(ex.Right)

	case *ast.FunctionCall:
		return d.s(styleLabel, ex.Name) + d.argList(ex.Arguments)

	case *ast.IndexExpression:
		return d.expr(ex.List) +
			d.s(stylePunct, "[") + d.expr(ex.Index) + d.s(stylePunct, "]")

	case *ast.LengthExpression:
		return d.s(styleOpcodeCall, "len") +
			d.s(stylePunct, "(") + d.expr(ex.List) + d.s(stylePunct, ")")

	case *ast.LocationExpression:
		return d.s(styleOp, "&") + d.s(styleIdent, ex.Name)

	case *ast.ErrorTypeCheckExpression:
		return d.expr(ex.Value) + " " + d.s(styleOp, "is") + " " + d.s(styleIdent, ex.TypeName)

	default:
		return d.s(styleNull, "?")
	}
}

// argList renders a function-call argument list as "(a, b, c)".
func (d *disassembler) argList(args []ast.Expression) string {
	if len(args) == 0 {
		return d.s(stylePunct, "()")
	}
	parts := make([]string, len(args))
	for i, a := range args {
		parts[i] = d.expr(a)
	}
	return d.s(stylePunct, "(") +
		strings.Join(parts, d.s(stylePunct, ", ")) +
		d.s(stylePunct, ")")
}
