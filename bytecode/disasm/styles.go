// Package disasm – styles.go
//
// Dracula-inspired lipgloss colour palette shared across all disassembler
// files.  A single renderer is created once at package init time so that
// TrueColor is always requested regardless of the caller's environment.
// Individual helpers then use d.s() to conditionally apply styles at
// run-time based on whether the caller enabled colour output.
package disasm

import (
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

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
	styleIdx   = disasmRenderer.NewStyle().Foreground(lipgloss.Color("#6272A4"))              // muted        – instruction index numbers
	styleIdent = disasmRenderer.NewStyle().Foreground(lipgloss.Color("#F8F8F2"))              // white        – variable / parameter names
	styleLabel = disasmRenderer.NewStyle().Foreground(lipgloss.Color("#8BE9FD")).Italic(true) // cyan italic  – function / import labels
	styleStr   = disasmRenderer.NewStyle().Foreground(lipgloss.Color("#50FA7B"))              // green        – string literals
	styleNum   = disasmRenderer.NewStyle().Foreground(lipgloss.Color("#F1FA8C"))              // yellow       – number literals
	styleBool  = disasmRenderer.NewStyle().Foreground(lipgloss.Color("#FFB86C"))              // orange       – boolean literals
	styleNull  = disasmRenderer.NewStyle().Foreground(lipgloss.Color("#6272A4")).Italic(true) // muted italic – null / nothing
	styleOp    = disasmRenderer.NewStyle().Foreground(lipgloss.Color("#FF5555"))              // red          – operators
	styleArrow = disasmRenderer.NewStyle().Foreground(lipgloss.Color("#FF79C6"))              // pink         – ← assignment arrow
	stylePunct = disasmRenderer.NewStyle().Foreground(lipgloss.Color("#6272A4"))              // muted        – ( ) [ ] ,
	styleConst = disasmRenderer.NewStyle().Foreground(lipgloss.Color("#FFD700"))              // gold         – [const] tag
	styleMeta  = disasmRenderer.NewStyle().Foreground(lipgloss.Color("#6272A4")).Italic(true) // muted italic – type tags, metadata
	styleType  = disasmRenderer.NewStyle().Foreground(lipgloss.Color("#8BE9FD")).Italic(true) // cyan italic  – type annotations

	// Header
	styleHeaderTitle = disasmRenderer.NewStyle().Foreground(lipgloss.Color("#BD93F9")).Bold(true)
	styleHeaderMeta  = disasmRenderer.NewStyle().Foreground(lipgloss.Color("#6272A4")).Italic(true)
)
