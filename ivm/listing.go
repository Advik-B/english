package ivm

// Listing produces a human-readable, instruction-level listing of a Chunk,
// similar to Python's dis module output.  When useColor is true, ANSI escape
// codes are added using a Dracula-inspired palette.
//
// Format:
//
//	<index>   <opcode>          <operand annotation>
//
// Nested sub-chunks (functions, struct methods, struct default-value
// expressions) are printed recursively with a header line.

import (
	"fmt"
	"math"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

// listingRenderer is created once so TrueColor is always requested.
var listingRenderer = func() *lipgloss.Renderer {
	r := lipgloss.NewRenderer(os.Stdout)
	r.SetColorProfile(termenv.TrueColor)
	return r
}()

// colour styles (Dracula palette)
var (
	lsHeader  = listingRenderer.NewStyle().Foreground(lipgloss.Color("#BD93F9")).Bold(true) // purple  – chunk header
	lsIndex   = listingRenderer.NewStyle().Foreground(lipgloss.Color("#6272A4"))            // muted   – index
	lsOpData  = listingRenderer.NewStyle().Foreground(lipgloss.Color("#50FA7B")).Bold(true) // green   – data opcodes
	lsOpCtrl  = listingRenderer.NewStyle().Foreground(lipgloss.Color("#FF79C6")).Bold(true) // pink    – control flow
	lsOpIO    = listingRenderer.NewStyle().Foreground(lipgloss.Color("#8BE9FD")).Bold(true) // cyan    – i/o
	lsOpCall  = listingRenderer.NewStyle().Foreground(lipgloss.Color("#FFB86C")).Bold(true) // orange  – calls
	lsOpMisc  = listingRenderer.NewStyle().Foreground(lipgloss.Color("#F8F8F2"))            // white   – everything else
	lsName    = listingRenderer.NewStyle().Foreground(lipgloss.Color("#F8F8F2")).Italic(true)
	lsStr     = listingRenderer.NewStyle().Foreground(lipgloss.Color("#50FA7B"))
	lsNum     = listingRenderer.NewStyle().Foreground(lipgloss.Color("#F1FA8C"))
	lsBool    = listingRenderer.NewStyle().Foreground(lipgloss.Color("#FFB86C"))
	lsComment = listingRenderer.NewStyle().Foreground(lipgloss.Color("#6272A4")).Italic(true)
)

// Listing produces an opcode-level listing of chunk and all its sub-chunks.
// title is used in the header (e.g. "<module>", "function add", etc.).
func Listing(chunk *Chunk, title string, useColor bool) string {
	var sb strings.Builder
	printChunk(&sb, chunk, title, useColor, 0)
	return sb.String()
}

func printChunk(sb *strings.Builder, chunk *Chunk, title string, color bool, depth int) {
	indent := strings.Repeat("  ", depth)

	// ── Header ────────────────────────────────────────────────────────────────
	header := fmt.Sprintf("%s=== %s ===", indent, title)
	sb.WriteString(applyStyle(color, lsHeader, header))
	sb.WriteString("\n")

	// ── Constant pool ────────────────────────────────────────────────────────
	if len(chunk.Constants) > 0 {
		sb.WriteString(indent)
		sb.WriteString(applyStyle(color, lsComment, "Constants:"))
		sb.WriteString("\n")
		for i, c := range chunk.Constants {
			valStr := formatConst(c)
			var valStyle lipgloss.Style
			switch c.(type) {
			case string:
				valStyle = lsStr
			case float64:
				valStyle = lsNum
			case bool:
				valStyle = lsBool
			default:
				valStyle = lsComment
			}
			idxPart := applyStyle(color, lsComment, fmt.Sprintf("%s  [%d] ", indent, i))
			valPart := applyStyle(color, valStyle, valStr)
			sb.WriteString(idxPart)
			sb.WriteString(valPart)
			sb.WriteString("\n")
		}
	}

	// ── Name pool ────────────────────────────────────────────────────────────
	if len(chunk.Names) > 0 {
		sb.WriteString(indent)
		sb.WriteString(applyStyle(color, lsComment, "Names:"))
		sb.WriteString("\n")
		for i, n := range chunk.Names {
			idxPart := applyStyle(color, lsComment, fmt.Sprintf("%s  [%d] ", indent, i))
			namePart := applyStyle(color, lsName, fmt.Sprintf("%q", n))
			sb.WriteString(idxPart)
			sb.WriteString(namePart)
			sb.WriteString("\n")
		}
	}

	sb.WriteString(indent)
	sb.WriteString(applyStyle(color, lsComment, "Code:"))
	sb.WriteString("\n")

	// ── Instructions ─────────────────────────────────────────────────────────
	for i, instr := range chunk.Code {
		annotation := formatOperand(chunk, instr)
		opStr := OpName(instr.Op)
		idxStr := fmt.Sprintf("%s%4d  ", indent, i)

		sb.WriteString(applyStyle(color, lsIndex, idxStr))
		sb.WriteString(applyStyle(color, opStyle(instr.Op), fmt.Sprintf("%-24s", opStr)))
		if annotation != "" {
			sb.WriteString("  ")
			sb.WriteString(applyStyle(color, lsComment, annotation))
		}
		sb.WriteString("\n")
	}

	// ── Function sub-chunks ───────────────────────────────────────────────────
	for _, fc := range chunk.Funcs {
		params := strings.Join(fc.Params, ", ")
		sb.WriteString("\n")
		printChunk(sb, fc.Body, fmt.Sprintf("function %s(%s)", fc.Name, params), color, depth+1)
	}

	// ── Struct sub-chunks ─────────────────────────────────────────────────────
	for _, sd := range chunk.StructDefs {
		sb.WriteString("\n")
		sb.WriteString(indent)
		sb.WriteString(applyStyle(color, lsComment, fmt.Sprintf("struct %s:", sd.Name)))
		sb.WriteString("\n")
		for _, f := range sd.Fields {
			line := fmt.Sprintf("%s  field %s : %s", indent, f.Name, f.TypeName)
			sb.WriteString(applyStyle(color, lsComment, line))
			sb.WriteString("\n")
			if f.DefaultExprChunk != nil {
				printChunk(sb, f.DefaultExprChunk,
					fmt.Sprintf("default(%s.%s)", sd.Name, f.Name), color, depth+2)
			}
		}
		for _, m := range sd.Methods {
			params := strings.Join(m.Params, ", ")
			sb.WriteString("\n")
			printChunk(sb, m.Body,
				fmt.Sprintf("method %s.%s(%s)", sd.Name, m.Name, params), color, depth+2)
		}
	}
}

// ─── operand annotation ───────────────────────────────────────────────────────

func formatOperand(chunk *Chunk, instr Instruction) string {
	op := instr.Op
	operand := instr.Operand

	name := func(idx uint32) string {
		if int(idx) < len(chunk.Names) {
			return fmt.Sprintf("%q (names[%d])", chunk.Names[idx], idx)
		}
		return fmt.Sprintf("names[%d]", idx)
	}
	constVal := func(idx uint32) string {
		if int(idx) < len(chunk.Constants) {
			return fmt.Sprintf("%s (consts[%d])", formatConst(chunk.Constants[idx]), idx)
		}
		return fmt.Sprintf("consts[%d]", idx)
	}

	switch op {
	case OP_LOAD_CONST:
		return constVal(operand)
	case OP_LOAD_VAR, OP_DEFINE_VAR, OP_DEFINE_CONST, OP_STORE_VAR, OP_TOGGLE_VAR,
		OP_MAKE_REFERENCE, OP_LOCATION:
		return name(operand)
	case OP_DEFINE_TYPED, OP_DEFINE_TYPED_CONST:
		return name(operand)
	case OP_JUMP:
		return fmt.Sprintf("-> %d", operand)
	case OP_JUMP_IF_FALSE, OP_JUMP_IF_TRUE:
		return fmt.Sprintf("-> %d", operand)
	case OP_BINARY_OP:
		return BinOp(operand).String()
	case OP_UNARY_OP:
		switch UnaryOp(operand) {
		case UnaryNeg:
			return "neg (-)"
		case UnaryNot:
			return "not"
		}
	case OP_CALL:
		argc := operand >> 16
		nameIdx := operand & 0xFFFF
		return fmt.Sprintf("%s argc=%d", name(nameIdx), argc)
	case OP_CALL_METHOD:
		argc := operand >> 16
		methIdx := operand & 0xFFFF
		return fmt.Sprintf("%s argc=%d", name(methIdx), argc)
	case OP_DEFINE_FUNC:
		if int(operand) < len(chunk.Funcs) {
			fc := chunk.Funcs[operand]
			return fmt.Sprintf("%q (funcs[%d])", fc.Name, operand)
		}
		return fmt.Sprintf("funcs[%d]", operand)
	case OP_DEFINE_STRUCT:
		if int(operand) < len(chunk.StructDefs) {
			return fmt.Sprintf("%q (structs[%d])", chunk.StructDefs[operand].Name, operand)
		}
		return fmt.Sprintf("structs[%d]", operand)
	case OP_NEW_STRUCT:
		fieldCount := operand >> 16
		snIdx := operand & 0xFFFF
		return fmt.Sprintf("%s fields=%d", name(snIdx), fieldCount)
	case OP_GET_FIELD, OP_SET_FIELD:
		return name(operand)
	case OP_CAST:
		return name(operand)
	case OP_ERROR_TYPE_CHECK:
		return name(operand)
	case OP_NIL_CHECK:
		if operand == 1 {
			return "is_something"
		}
		return "is_nothing"
	case OP_ASK:
		if operand == 1 {
			return "with_prompt"
		}
		return "no_prompt"
	case OP_PRINT:
		count := operand >> 1
		newline := (operand & 1) == 1
		return fmt.Sprintf("count=%d newline=%v", count, newline)
	case OP_BUILD_LIST, OP_BUILD_ARRAY:
		return fmt.Sprintf("count=%d", operand)
	case OP_INDEX_SET:
		return name(operand)
	case OP_LOOKUP_SET:
		return name(operand)
	case OP_RAISE:
		if operand == 0 {
			return "generic"
		}
		return name(operand - 1)
	case OP_TRY_BEGIN:
		return fmt.Sprintf("catch_offset=%d", operand)
	case OP_TRY_END:
		return fmt.Sprintf("end_offset=%d", operand)
	case OP_CATCH:
		errVarIdx := operand >> 16
		errTypeIdx := operand & 0xFFFF
		varStr := "var=_"
		if errVarIdx > 0 && int(errVarIdx) < len(chunk.Names) {
			varStr = fmt.Sprintf("var=%q", chunk.Names[errVarIdx])
		} else if errVarIdx > 0 {
			varStr = fmt.Sprintf("var=names[%d]", errVarIdx)
		}
		typeStr := "type=any"
		if errTypeIdx > 0 {
			typeStr = fmt.Sprintf("type=%s", name(errTypeIdx-1))
		}
		return fmt.Sprintf("%s %s", varStr, typeStr)
	case OP_DEFINE_ERROR_TYPE:
		nameIdx := operand >> 16
		parentIdx := operand & 0xFFFF
		n := name(nameIdx)
		if parentIdx == 0 {
			return fmt.Sprintf("%s parent=error", n)
		}
		return fmt.Sprintf("%s parent=%s", n, name(parentIdx-1))
	case OP_SWAP_VARS:
		n1 := operand >> 16
		n2 := operand & 0xFFFF
		return fmt.Sprintf("%s <-> %s", name(n1), name(n2))
	case OP_IMPORT:
		flags := operand
		parts := []string{}
		if flags&4 != 0 {
			parts = append(parts, "all")
		}
		if flags&2 != 0 {
			parts = append(parts, "safe")
		}
		if flags&1 != 0 {
			parts = append(parts, "selective")
		}
		return strings.Join(parts, " ")
	case OP_SET_LINE:
		return fmt.Sprintf("line=%d", operand)
	}
	return ""
}

// ─── helpers ──────────────────────────────────────────────────────────────────

// String method for BinOp for the listing.
func (b BinOp) String() string {
	switch b {
	case BinAdd:
		return "+"
	case BinSub:
		return "-"
	case BinMul:
		return "*"
	case BinDiv:
		return "/"
	case BinMod:
		return "%"
	case BinEq:
		return "=="
	case BinNeq:
		return "!="
	case BinLt:
		return "<"
	case BinLte:
		return "<="
	case BinGt:
		return ">"
	case BinGte:
		return ">="
	default:
		return fmt.Sprintf("binop(%d)", int(b))
	}
}

func formatConst(v interface{}) string {
	switch val := v.(type) {
	case float64:
		if math.IsInf(val, 1) {
			return "+inf"
		}
		if math.IsInf(val, -1) {
			return "-inf"
		}
		if math.IsNaN(val) {
			return "NaN"
		}
		return fmt.Sprintf("%g", val)
	case string:
		return fmt.Sprintf("%q", val)
	case bool:
		if val {
			return "true"
		}
		return "false"
	case nil:
		return "nothing"
	case []interface{}:
		parts := make([]string, len(val))
		for i, item := range val {
			parts[i] = formatConst(item)
		}
		return "[" + strings.Join(parts, ", ") + "]"
	default:
		return fmt.Sprintf("%v", v)
	}
}

func opStyle(op Opcode) lipgloss.Style {
	switch op {
	case OP_DEFINE_VAR, OP_DEFINE_CONST, OP_DEFINE_TYPED, OP_DEFINE_TYPED_CONST,
		OP_DEFINE_FUNC, OP_DEFINE_STRUCT, OP_DEFINE_ERROR_TYPE,
		OP_LOAD_CONST, OP_LOAD_NOTHING, OP_LOAD_VAR,
		OP_BUILD_LIST, OP_BUILD_ARRAY, OP_BUILD_LOOKUP,
		OP_NEW_STRUCT, OP_IMPORT:
		return lsOpData
	case OP_JUMP, OP_JUMP_IF_FALSE, OP_JUMP_IF_TRUE, OP_RETURN,
		OP_TRY_BEGIN, OP_TRY_END, OP_CATCH, OP_RAISE,
		OP_PUSH_SCOPE, OP_POP_SCOPE:
		return lsOpCtrl
	case OP_PRINT, OP_ASK:
		return lsOpIO
	case OP_CALL, OP_CALL_METHOD:
		return lsOpCall
	default:
		return lsOpMisc
	}
}

func applyStyle(color bool, style lipgloss.Style, s string) string {
	if !color {
		return s
	}
	return style.Render(s)
}
