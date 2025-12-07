package cmd

import (
	"english/parser"
	"english/vm"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Styles for the TUI
var (
	// Title bar style
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1)

	// Prompt styles
	promptStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#04B575")).
			Bold(true)

	multilinePromptStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFA500")).
				Bold(true)

	// Output styles
	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#04B575"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000")).
			Bold(true)

	infoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7D56F4"))

	// Syntax highlighting
	keywordStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#5C7FFF")).
			Bold(true)

	stringStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#04B575"))

	numberStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00D7FF"))

	// Panel styles
	panelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7D56F4")).
			Padding(0, 1).
			MarginTop(1)

	varNameStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFB86C")).
			Bold(true)

	varValueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F8F8F2"))

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6272A4")).
			Italic(true)
)

// model represents the TUI state
type model struct {
	// Input state
	input     string
	cursorPos int

	// REPL state
	env       *vm.Environment
	evaluator *vm.Evaluator
	history   []string
	output    []string

	// Multiline support
	multiline bool
	buffer    []string

	// UI state
	width    int
	height   int
	quitting bool
}

// initialModel creates the initial TUI model
func initialModel() model {
	env := vm.NewEnvironment()
	return model{
		input:     "",
		cursorPos: 0,
		env:       env,
		evaluator: vm.NewEvaluator(env),
		history:   []string{},
		output: []string{
			titleStyle.Render("✨ English Language Interpreter ✨"),
			helpStyle.Render("Type :help for commands • Ctrl+C to exit"),
			"",
		},
		multiline: false,
		buffer:    []string{},
		width:     80,
		height:    24,
		quitting:  false,
	}
}

// Init initializes the TUI
func (m model) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the model
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyPress(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	return m, nil
}

// handleKeyPress processes keyboard input
func (m model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyCtrlC, tea.KeyEsc:
		m.quitting = true
		return m, tea.Quit

	case tea.KeyEnter:
		return m.handleEnter()

	case tea.KeyBackspace:
		if m.cursorPos > 0 {
			m.input = m.input[:m.cursorPos-1] + m.input[m.cursorPos:]
			m.cursorPos--
		}

	case tea.KeyDelete:
		if m.cursorPos < len(m.input) {
			m.input = m.input[:m.cursorPos] + m.input[m.cursorPos+1:]
		}

	case tea.KeyLeft:
		if m.cursorPos > 0 {
			m.cursorPos--
		}

	case tea.KeyRight:
		if m.cursorPos < len(m.input) {
			m.cursorPos++
		}

	case tea.KeyHome:
		m.cursorPos = 0

	case tea.KeyEnd:
		m.cursorPos = len(m.input)

	case tea.KeyRunes:
		// Insert text at cursor position
		m.input = m.input[:m.cursorPos] + string(msg.Runes) + m.input[m.cursorPos:]
		m.cursorPos += len(msg.Runes)
	}

	return m, nil
}

// handleEnter processes the Enter key
func (m model) handleEnter() (tea.Model, tea.Cmd) {
	line := m.input
	m.input = ""
	m.cursorPos = 0

	trimmed := strings.TrimSpace(line)

	// Handle empty input
	if trimmed == "" {
		if m.multiline {
			m.buffer = append(m.buffer, line)
		}
		return m, nil
	}

	// Handle commands (don't add to history)
	if strings.HasPrefix(trimmed, ":") {
		output := m.handleCommand(trimmed)
		if output != "" {
			m.output = append(m.output, output)
		}
		return m, nil
	}

	// Handle exit (don't add to history)
	if trimmed == "exit" || trimmed == "quit" {
		m.quitting = true
		return m, tea.Quit
	}

	// Add to buffer for execution
	m.buffer = append(m.buffer, line)

	lower := strings.ToLower(trimmed)

	// Check for multiline start
	if strings.Contains(lower, "following:") || strings.Contains(lower, ", then") {
		m.multiline = true
		return m, nil
	}

	// Check for multiline end
	if strings.Contains(lower, "thats it.") {
		m.multiline = false
		code := strings.Join(m.buffer, "\n")
		m.buffer = []string{}
		// Add complete multiline code to history
		m.history = append(m.history, code)
		m.executeCode(code)
		return m, nil
	}

	// Execute single-line if not in multiline and ends with period
	if !m.multiline && strings.HasSuffix(trimmed, ".") {
		code := strings.Join(m.buffer, "\n")
		m.buffer = []string{}
		// Add complete single-line code to history
		m.history = append(m.history, code)
		m.executeCode(code)
	} else if !m.multiline {
		// If not in multiline and doesn't end with period, clear buffer to prevent accumulation
		m.buffer = []string{}
		m.output = append(m.output, helpStyle.Render("(Incomplete statement - statements must end with '.')"))
	}

	return m, nil
}

// handleCommand processes REPL commands
func (m *model) handleCommand(cmd string) string {
	switch cmd {
	case ":help", ":h":
		return m.getHelp()

	case ":clear", ":cls":
		m.output = []string{}
		m.history = []string{}
		return ""

	case ":vars", ":v":
		return m.listVariables()

	case ":history", ":hist":
		return m.getHistory()

	case ":reset":
		m.env = vm.NewEnvironment()
		m.evaluator = vm.NewEvaluator(m.env)
		return successStyle.Render("✓ Session reset")

	case ":exit", ":quit", ":q":
		m.quitting = true
		return ""

	default:
		return errorStyle.Render("✗ Unknown command: " + cmd) + " " + helpStyle.Render("(try :help)")
	}
}

// getHelp returns help text
func (m *model) getHelp() string {
	var help strings.Builder

	help.WriteString(infoStyle.Render("📖 Commands:") + "\n")
	help.WriteString("  " + keywordStyle.Render(":help") + "    Show this help\n")
	help.WriteString("  " + keywordStyle.Render(":vars") + "    List variables and functions\n")
	help.WriteString("  " + keywordStyle.Render(":clear") + "   Clear screen\n")
	help.WriteString("  " + keywordStyle.Render(":reset") + "   Reset environment\n")
	help.WriteString("  " + keywordStyle.Render(":history") + " Show command history\n")
	help.WriteString("  " + keywordStyle.Render(":exit") + "    Exit REPL\n\n")

	help.WriteString(infoStyle.Render("📚 Quick Examples:") + "\n")
	help.WriteString("  " + helpStyle.Render("Declare x to be 5.") + "\n")
	help.WriteString("  " + helpStyle.Render("Print the value of x.") + "\n")
	help.WriteString("  " + helpStyle.Render("Declare y to always be 10.") + " " + helpStyle.Render("(constant)") + "\n")

	return help.String()
}

// formatVariableEntry formats a single variable for display
func (m *model) formatVariableEntry(name string, value vm.Value) string {
	valStr := formatValue(value)
	if m.env.IsConstant(name) {
		return "  " + varNameStyle.Render(name) + " = " + varValueStyle.Render(valStr) + " " + helpStyle.Render("(const)")
	}
	return "  " + varNameStyle.Render(name) + " = " + varValueStyle.Render(valStr)
}

// listVariables returns formatted list of variables and functions
func (m *model) listVariables() string {
	vars := m.env.GetAllVariables()
	funcs := m.env.GetAllFunctions()

	if len(vars) == 0 && len(funcs) == 0 {
		return helpStyle.Render("No variables or functions defined yet.")
	}

	var result strings.Builder

	// List variables
	if len(vars) > 0 {
		varNames := make([]string, 0, len(vars))
		for name := range vars {
			varNames = append(varNames, name)
		}
		sort.Strings(varNames)

		result.WriteString(infoStyle.Render("📦 Variables:") + "\n")
		for _, name := range varNames {
			result.WriteString(m.formatVariableEntry(name, vars[name]) + "\n")
		}
	}

	// List functions
	if len(funcs) > 0 {
		funcNames := make([]string, 0, len(funcs))
		for name := range funcs {
			funcNames = append(funcNames, name)
		}
		sort.Strings(funcNames)

		if len(vars) > 0 {
			result.WriteString("\n")
		}
		result.WriteString(infoStyle.Render("⚡ Functions:") + "\n")
		for _, name := range funcNames {
			fn := funcs[name]
			params := strings.Join(fn.Parameters, ", ")
			result.WriteString("  " + keywordStyle.Render(name) + "(" + params + ")\n")
		}
	}

	return result.String()
}

// getHistory returns command history
func (m *model) getHistory() string {
	if len(m.history) == 0 {
		return helpStyle.Render("No command history.")
	}

	var result strings.Builder
	result.WriteString(infoStyle.Render("📜 History:") + "\n")
	for i, cmd := range m.history {
		result.WriteString(fmt.Sprintf("  %3d  %s\n", i+1, cmd))
	}
	return result.String()
}

// executeCode parses and executes code
func (m *model) executeCode(code string) {
	// Parse
	lexer := parser.NewLexer(code)
	tokens := lexer.TokenizeAll()

	p := parser.NewParser(tokens)
	program, err := p.Parse()
	if err != nil {
		m.output = append(m.output, errorStyle.Render("✗ Parse error: ")+err.Error())
		return
	}

	// Capture stdout during execution
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		// Fallback to execution without capture if pipe fails
		_, execErr := m.evaluator.Eval(program)
		if execErr != nil {
			m.output = append(m.output, errorStyle.Render("✗ Runtime error: ")+execErr.Error())
		} else {
			m.output = append(m.output, successStyle.Render("✓ Success"))
		}
		return
	}
	
	// Set up cleanup to ensure stdout is always restored
	os.Stdout = w
	defer func() {
		w.Close()
		os.Stdout = oldStdout
	}()

	// Execute
	execErr := func() error {
		_, err := m.evaluator.Eval(program)
		return err
	}()

	// Close writer to signal completion
	w.Close()
	os.Stdout = oldStdout

	// Read captured output using io.Copy
	var capturedOutput strings.Builder
	_, _ = io.Copy(&capturedOutput, r)
	r.Close()

	// Handle execution errors
	if execErr != nil {
		m.output = append(m.output, errorStyle.Render("✗ Runtime error: ")+execErr.Error())
		return
	}

	// Show captured output or success message
	output := strings.TrimSpace(capturedOutput.String())
	if output != "" {
		m.output = append(m.output, output)
	} else {
		m.output = append(m.output, successStyle.Render("✓ Success"))
	}
}

// formatValue converts a value to display string
func formatValue(v vm.Value) string {
	switch val := v.(type) {
	case float64:
		if val == float64(int64(val)) {
			return fmt.Sprintf("%d", int64(val))
		}
		return fmt.Sprintf("%.4g", val)
	case string:
		return fmt.Sprintf("%q", val)
	case bool:
		return fmt.Sprintf("%t", val)
	case []interface{}:
		if len(val) == 0 {
			return "[]"
		}
		if len(val) > 3 {
			return fmt.Sprintf("[%d items]", len(val))
		}
		parts := make([]string, len(val))
		for i, item := range val {
			parts[i] = formatValue(item)
		}
		return "[" + strings.Join(parts, ", ") + "]"
	case nil:
		return "nil"
	default:
		return fmt.Sprintf("%v", val)
	}
}

// View renders the TUI
func (m model) View() string {
	if m.quitting {
		return titleStyle.Render("👋 Goodbye! Thanks for using English!\n")
	}

	var view strings.Builder

	// Title
	view.WriteString(titleStyle.Render(" English REPL "))
	view.WriteString("\n")
	view.WriteString(helpStyle.Render("Type :help for commands"))
	view.WriteString("\n\n")

	// Output history (show last N lines based on height)
	maxOutputLines := m.height - 8
	if maxOutputLines < 5 {
		maxOutputLines = 5
	}

	startIdx := 0
	if len(m.output) > maxOutputLines {
		startIdx = len(m.output) - maxOutputLines
	}

	for i := startIdx; i < len(m.output); i++ {
		view.WriteString(m.output[i])
		view.WriteString("\n")
	}

	// Prompt
	view.WriteString("\n")
	if m.multiline {
		view.WriteString(multilinePromptStyle.Render("... "))
	} else {
		view.WriteString(promptStyle.Render(">>> "))
	}

	// Input with cursor
	if m.cursorPos < len(m.input) {
		view.WriteString(m.input[:m.cursorPos])
		view.WriteString(lipgloss.NewStyle().Reverse(true).Render(string(m.input[m.cursorPos])))
		view.WriteString(m.input[m.cursorPos+1:])
	} else {
		view.WriteString(m.input)
		view.WriteString(lipgloss.NewStyle().Reverse(true).Render(" "))
	}

	return view.String()
}

// StartREPL starts the TUI REPL
func StartREPL() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error starting REPL: %v\n", err)
		os.Exit(1)
	}
}
