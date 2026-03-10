// Package stacktraces provides pretty-printed, coloured runtime error output
// for the English language interpreter.
//
// It detects whether the terminal supports ANSI colour and renders errors with
// full call-stack information in a consistent, human-readable format.
package stacktraces

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-isatty"
	"github.com/muesli/termenv"
)

// ─── Colour-support detection ────────────────────────────────────────────────

// HasColor reports whether stderr supports ANSI colour output.
//
// Colour is disabled when any of the following conditions holds:
//   - The NO_COLOR environment variable is set to any value, including the
//     empty string (https://no-color.org/ – "when set, even to empty").
//   - stderr is not a TTY (e.g. piped or redirected).
func HasColor() bool {
	if _, ok := os.LookupEnv("NO_COLOR"); ok {
		return false
	}
	fd := os.Stderr.Fd()
	return isatty.IsTerminal(fd) || isatty.IsCygwinTerminal(fd)
}

// ─── Styles ──────────────────────────────────────────────────────────────────

// colorRenderer forces TrueColor output. It is used by renderColored when the
// caller has already established that colour is supported (or explicitly
// requested via RenderWithColor). This ensures ANSI codes are always emitted
// even in non-TTY contexts such as tests.
var colorRenderer = func() *lipgloss.Renderer {
	r := lipgloss.NewRenderer(os.Stderr)
	r.SetColorProfile(termenv.TrueColor)
	return r
}()

var (
	// Header bar across the top of the error block.
	headerStyle = colorRenderer.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FF5555")).
			Background(lipgloss.Color("#3D0000")).
			Padding(0, 1)

	// Error type label (e.g. "RuntimeError", "ParseError").
	labelStyle = colorRenderer.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FF5555"))

	// Primary error message.
	messageStyle = colorRenderer.NewStyle().
			Foreground(lipgloss.Color("#FFB8B8"))

	// "Call Stack" section header.
	stackHeaderStyle = colorRenderer.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#FFD700"))

	// Frame number inside the call stack.
	frameNumberStyle = colorRenderer.NewStyle().
				Foreground(lipgloss.Color("#6272A4"))

	// Frame name inside the call stack.
	frameNameStyle = colorRenderer.NewStyle().
			Foreground(lipgloss.Color("#F8F8F2"))

	// Dimmed decorative separator.
	separatorStyle = colorRenderer.NewStyle().
			Foreground(lipgloss.Color("#44475A"))
)

// ─── Error kinds ─────────────────────────────────────────────────────────────

// RuntimeError is the interface satisfied by vm.RuntimeError.
// Using a local interface avoids an import cycle between this package and vm.
type RuntimeError interface {
	error
	RuntimeMessage() string
	RuntimeCallStack() []string
}

// ─── Public API ──────────────────────────────────────────────────────────────

// Render formats err as a pretty, colour-aware string.
//
// When colour is not supported the output is plain text but still contains the
// full call-stack so that information is never lost.
func Render(err error) string {
	return RenderWithColor(err, HasColor())
}

// RenderWithColor formats err as a pretty string using the explicit colour
// flag. Pass HasColor() for normal use; pass true/false in tests or when the
// caller has already determined colour support.
func RenderWithColor(err error, color bool) string {
	if err == nil {
		return ""
	}
	if !color {
		return renderPlain(err)
	}
	return renderColored(err)
}

// Print writes the formatted error to stderr.
func Print(err error) {
	fmt.Fprint(os.Stderr, Render(err))
}

// ─── Plain-text renderer ─────────────────────────────────────────────────────

func renderPlain(err error) string {
	var sb strings.Builder

	if re, ok := err.(RuntimeError); ok {
		sb.WriteString("Runtime Error: ")
		sb.WriteString(re.RuntimeMessage())
		sb.WriteString("\n")

		stack := re.RuntimeCallStack()
		if len(stack) > 0 {
			sb.WriteString("\nCall Stack (most recent first):\n")
			for i, frame := range stack {
				sb.WriteString(fmt.Sprintf("  %d. %s\n", i+1, frame))
			}
		}
		return sb.String()
	}

	// Generic error (parse error, type error, etc.)
	sb.WriteString("Error: ")
	sb.WriteString(err.Error())
	sb.WriteString("\n")
	return sb.String()
}

// ─── Coloured renderer ───────────────────────────────────────────────────────

func renderColored(err error) string {
	var sb strings.Builder

	if re, ok := err.(RuntimeError); ok {
		renderRuntimeError(&sb, re)
		return sb.String()
	}

	renderGenericError(&sb, err)
	return sb.String()
}

func renderRuntimeError(sb *strings.Builder, re RuntimeError) {
	sep := separatorStyle.Render(strings.Repeat("─", 50))

	sb.WriteString("\n")
	sb.WriteString(headerStyle.Render(" ✖  Runtime Error "))
	sb.WriteString("\n")
	sb.WriteString(sep)
	sb.WriteString("\n\n")

	sb.WriteString("  ")
	sb.WriteString(labelStyle.Render("Message: "))
	sb.WriteString(messageStyle.Render(re.RuntimeMessage()))
	sb.WriteString("\n")

	stack := re.RuntimeCallStack()
	if len(stack) > 0 {
		sb.WriteString("\n")
		sb.WriteString("  ")
		sb.WriteString(stackHeaderStyle.Render("Call Stack") + separatorStyle.Render(" (most recent first)"))
		sb.WriteString("\n")
		sb.WriteString(sep)
		sb.WriteString("\n")
		for i, frame := range stack {
			sb.WriteString("  ")
			sb.WriteString(frameNumberStyle.Render(fmt.Sprintf("%2d.", i+1)))
			sb.WriteString("  ")
			sb.WriteString(frameNameStyle.Render(frame))
			sb.WriteString("\n")
		}
	}

	sb.WriteString(sep)
	sb.WriteString("\n\n")
}

func renderGenericError(sb *strings.Builder, err error) {
	sep := separatorStyle.Render(strings.Repeat("─", 50))

	sb.WriteString("\n")
	sb.WriteString(headerStyle.Render(" ✖  Error "))
	sb.WriteString("\n")
	sb.WriteString(sep)
	sb.WriteString("\n\n")

	sb.WriteString("  ")
	sb.WriteString(messageStyle.Render(err.Error()))
	sb.WriteString("\n")

	sb.WriteString(sep)
	sb.WriteString("\n\n")
}
