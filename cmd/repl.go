package cmd

import (
	"english/parser"
	"english/vm"
	"fmt"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("86")).
			Background(lipgloss.Color("236")).
			Padding(0, 1)

	promptStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("42")).
			Bold(true)

	multilinePromptStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("220"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true)

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("42"))

	commentStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))

	keywordStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("69")).
			Bold(true)

	stringStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("42"))

	numberStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("51"))

	operatorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("201"))

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))

	// Panel styles for variable display
	panelTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("212")).
			Background(lipgloss.Color("236")).
			Padding(0, 1).
			MarginBottom(1)

	panelBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("62")).
				Padding(0, 1)

	varNameStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")).
			Bold(true)

	varValueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("251"))

	constStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("141")).
			Italic(true)

	funcStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("117"))
)

// Display constants for value formatting
const (
	maxValueDisplayLength = 20
	maxArrayDisplayItems  = 3
)

func getWelcomeMessage() string {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("86")).
		Bold(true).
		Render("✨ Welcome to the English Language Interpreter! Type :help for assistance.")
}

type model struct {
	input     string
	history   []string
	output    []string
	multiline bool
	buffer    []string
	env       *vm.Environment
	evaluator *vm.Evaluator
	cursorPos int
	width     int
	height    int
	quitting  bool
}

func initialModel() model {
	env := vm.NewEnvironment()
	return model{
		input:     "",
		history:   []string{},
		output:    []string{getWelcomeMessage()},
		multiline: false,
		buffer:    []string{},
		env:       env,
		evaluator: vm.NewEvaluator(env),
		cursorPos: 0,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			m.quitting = true
			return m, tea.Quit

		case tea.KeyEnter:
			return m.handleEnter()

		case tea.KeyBackspace:
			if len(m.input) > 0 && m.cursorPos > 0 {
				m.input = m.input[:m.cursorPos-1] + m.input[m.cursorPos:]
				m.cursorPos--
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

		default:
			if msg.Type == tea.KeyRunes {
				// Insert character at cursor position
				m.input = m.input[:m.cursorPos] + string(msg.Runes) + m.input[m.cursorPos:]
				m.cursorPos += len(msg.Runes)
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	return m, nil
}

func (m model) handleEnter() (tea.Model, tea.Cmd) {
	line := m.input
	m.input = ""
	m.cursorPos = 0

	trimmed := strings.TrimSpace(line)

	// Handle commands
	if strings.HasPrefix(trimmed, ":") {
		m.output = append(m.output, m.handleCommand(trimmed))
		return m, nil
	}

	// Handle exit
	if trimmed == "exit" || trimmed == "quit" {
		m.quitting = true
		return m, tea.Quit
	}

	// Empty line
	if trimmed == "" {
		return m, nil
	}

	// Add to buffer
	m.buffer = append(m.buffer, line)

	// Check for multiline start
	if strings.Contains(strings.ToLower(trimmed), "following:") {
		m.multiline = true
		return m, nil
	}

	// Check for multiline end
	if strings.Contains(strings.ToLower(trimmed), "thats it.") {
		m.multiline = false
		code := strings.Join(m.buffer, "\n")
		m.buffer = []string{}
		result, err := m.executeCode(code)
		if err != nil {
			m.output = append(m.output, errorStyle.Render("✗ ")+err.Error())
		} else {
			m.output = append(m.output, successStyle.Render(result))
		}
		return m, nil
	}

	// Execute if not in multiline and ends with period
	if !m.multiline && strings.HasSuffix(trimmed, ".") {
		code := strings.Join(m.buffer, "\n")
		m.buffer = []string{}
		result, err := m.executeCode(code)
		if err != nil {
			m.output = append(m.output, errorStyle.Render("✗ ")+err.Error())
		} else {
			m.output = append(m.output, successStyle.Render(result))
		}
	}

	return m, nil
}

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
	case ":exit", ":quit", ":q":
		m.quitting = true
		return "Goodbye!"
	default:
		return errorStyle.Render("Unknown command: "+cmd) + helpStyle.Render(" (try :help)")
	}
}

func (m *model) listVariables() string {
	vars := m.env.GetAllVariables()
	funcs := m.env.GetAllFunctions()

	if len(vars) == 0 && len(funcs) == 0 {
		return commentStyle.Render("No variables or functions defined yet.")
	}

	var result strings.Builder
	result.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("86")).Bold(true).Render("📦 Current State:") + "\n\n")

	// Variables
	if len(vars) > 0 {
		varNames := make([]string, 0, len(vars))
		for name := range vars {
			varNames = append(varNames, name)
		}
		sort.Strings(varNames)

		result.WriteString(lipgloss.NewStyle().Bold(true).Render("Variables:") + "\n")
		for _, name := range varNames {
			val := vars[name]
			valStr := formatValue(val)
			if m.env.IsConstant(name) {
				result.WriteString("  " + constStyle.Render("const ") + varNameStyle.Render(name) + " = " + varValueStyle.Render(valStr) + "\n")
			} else {
				result.WriteString("  " + varNameStyle.Render(name) + " = " + varValueStyle.Render(valStr) + "\n")
			}
		}
	}

	// Functions
	if len(funcs) > 0 {
		funcNames := make([]string, 0, len(funcs))
		for name := range funcs {
			funcNames = append(funcNames, name)
		}
		sort.Strings(funcNames)

		if len(vars) > 0 {
			result.WriteString("\n")
		}
		result.WriteString(lipgloss.NewStyle().Bold(true).Render("Functions:") + "\n")
		for _, name := range funcNames {
			fn := funcs[name]
			params := strings.Join(fn.Parameters, ", ")
			result.WriteString("  " + funcStyle.Render("ƒ "+name) + "(" + params + ")\n")
		}
	}

	return result.String()
}

func (m *model) getHelp() string {
	help := lipgloss.NewStyle().Foreground(lipgloss.Color("86")).Bold(true).Render("Available Commands:") + "\n"
	help += "  " + successStyle.Render(":help") + "   - Show this help\n"
	help += "  " + successStyle.Render(":vars") + "   - List all variables and functions\n"
	help += "  " + successStyle.Render(":clear") + "  - Clear the screen\n"
	help += "  " + successStyle.Render(":exit") + "   - Exit REPL\n\n"
	help += lipgloss.NewStyle().Foreground(lipgloss.Color("86")).Bold(true).Render("Quick Start:") + "\n"
	help += "  " + commentStyle.Render("# Declare a variable") + "\n"
	help += "  " + m.highlightSyntax("Declare x to be 5.") + "\n"
	help += "  " + m.highlightSyntax("Print the value of x.") + "\n"
	help += "\n" + commentStyle.Render("Variables panel shows on the right →")
	return help
}

func (m model) executeCode(code string) (string, error) {
	lexer := parser.NewLexer(code)
	tokens := lexer.TokenizeAll()

	p := parser.NewParser(tokens)
	program, err := p.Parse()
	if err != nil {
		return "", fmt.Errorf("parse error: %v", err)
	}

	_, err = m.evaluator.Eval(program)
	if err != nil {
		return "", fmt.Errorf("runtime error: %v", err)
	}

	return "✓ Executed successfully", nil
}

func (m *model) highlightSyntax(code string) string {
	// Simple syntax highlighting
	keywords := map[string]bool{
		"declare": true, "function": true, "that": true, "does": true,
		"the": true, "following": true, "thats": true, "it": true,
		"to": true, "be": true, "always": true, "set": true,
		"call": true, "return": true, "print": true, "if": true,
		"then": true, "otherwise": true, "repeat": true, "while": true,
		"forever": true, "break": true, "out": true, "loop": true,
		"times": true, "for": true, "each": true, "in": true,
		"do": true, "takes": true, "and": true, "with": true,
		"result": true, "of": true, "calling": true, "value": true,
		"true": true, "false": true, "toggle": true, "location": true,
	}

	words := strings.Fields(code)
	result := []string{}

	for _, word := range words {
		lower := strings.ToLower(strings.Trim(word, ".,;:"))
		if keywords[lower] {
			result = append(result, keywordStyle.Render(word))
		} else if strings.HasPrefix(word, "\"") || strings.HasPrefix(word, "'") {
			result = append(result, stringStyle.Render(word))
		} else if len(word) > 0 && word[0] >= '0' && word[0] <= '9' {
			result = append(result, numberStyle.Render(word))
		} else {
			result = append(result, word)
		}
	}

	return strings.Join(result, " ")
}

// renderVariablesPanel creates the right-side panel showing variables and functions
func (m model) renderVariablesPanel(height int) string {
	var content strings.Builder

	// Get variables and functions
	vars := m.env.GetAllVariables()
	funcs := m.env.GetAllFunctions()

	// Panel title
	content.WriteString(panelTitleStyle.Render("📦 Variables & Functions") + "\n\n")

	// Sort variable names for consistent display
	varNames := make([]string, 0, len(vars))
	for name := range vars {
		varNames = append(varNames, name)
	}
	sort.Strings(varNames)

	// Display variables
	if len(varNames) > 0 {
		content.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("86")).Render("Variables:") + "\n")
		for _, name := range varNames {
			val := vars[name]
			valStr := formatValue(val)
			// Truncate long values
			if len(valStr) > maxValueDisplayLength {
				valStr = valStr[:maxValueDisplayLength-3] + "..."
			}
			if m.env.IsConstant(name) {
				content.WriteString("  " + constStyle.Render("const ") + varNameStyle.Render(name) + " = " + varValueStyle.Render(valStr) + "\n")
			} else {
				content.WriteString("  " + varNameStyle.Render(name) + " = " + varValueStyle.Render(valStr) + "\n")
			}
		}
		content.WriteString("\n")
	}

	// Sort function names
	funcNames := make([]string, 0, len(funcs))
	for name := range funcs {
		funcNames = append(funcNames, name)
	}
	sort.Strings(funcNames)

	// Display functions
	if len(funcNames) > 0 {
		content.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("117")).Render("Functions:") + "\n")
		for _, name := range funcNames {
			fn := funcs[name]
			params := strings.Join(fn.Parameters, ", ")
			content.WriteString("  " + funcStyle.Render("ƒ "+name) + "(" + params + ")\n")
		}
	}

	if len(varNames) == 0 && len(funcNames) == 0 {
		content.WriteString(commentStyle.Render("  No variables yet.\n  Try: Declare x to be 5."))
	}

	// Create the panel with border
	panelContent := content.String()
	panel := panelBorderStyle.Width(28).Height(height - 4).Render(panelContent)

	return panel
}

// formatValue converts a Value to a display string
func formatValue(v vm.Value) string {
	switch val := v.(type) {
	case float64:
		if val == float64(int64(val)) {
			return fmt.Sprintf("%d", int64(val))
		}
		return fmt.Sprintf("%.4g", val)
	case string:
		return fmt.Sprintf("\"%s\"", val)
	case []interface{}:
		if len(val) > maxArrayDisplayItems {
			return fmt.Sprintf("[%d items]", len(val))
		}
		parts := make([]string, len(val))
		for i, item := range val {
			parts[i] = formatValue(item)
		}
		return "[" + strings.Join(parts, ", ") + "]"
	case bool:
		if val {
			return "true"
		}
		return "false"
	case nil:
		return "nil"
	default:
		return fmt.Sprintf("%v", val)
	}
}

func (m model) View() string {
	if m.quitting {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("86")).
			Bold(true).
			Render("\n✨ Goodbye! Happy coding!\n\n")
	}

	// Calculate panel widths
	panelWidth := 32
	mainWidth := m.width - panelWidth - 3
	if mainWidth < 40 {
		mainWidth = 40
	}

	var mainContent strings.Builder

	// Title
	title := titleStyle.Render(" English Language Interpreter v1.0.0 ")
	mainContent.WriteString(title + "\n")
	mainContent.WriteString(helpStyle.Render("Type :help for commands, Ctrl+C to exit") + "\n\n")

	// Output history
	maxLines := m.height - 8
	if maxLines < 5 {
		maxLines = 5
	}
	if len(m.output) > maxLines {
		for _, line := range m.output[len(m.output)-maxLines:] {
			mainContent.WriteString(line + "\n")
		}
	} else {
		for _, line := range m.output {
			mainContent.WriteString(line + "\n")
		}
	}

	// Prompt
	var prompt string
	if m.multiline {
		prompt = multilinePromptStyle.Render("... ")
	} else {
		prompt = promptStyle.Render(">>> ")
	}

	// Input with cursor
	inputDisplay := m.input
	if m.cursorPos < len(m.input) {
		inputDisplay = m.input[:m.cursorPos] +
			lipgloss.NewStyle().Reverse(true).Render(string(m.input[m.cursorPos])) +
			m.input[m.cursorPos+1:]
	} else {
		inputDisplay = m.input + lipgloss.NewStyle().Reverse(true).Render(" ")
	}

	mainContent.WriteString("\n" + prompt + inputDisplay)

	// Create the main content panel
	mainPanel := lipgloss.NewStyle().Width(mainWidth).Render(mainContent.String())

	// Create the variables panel
	varPanel := m.renderVariablesPanel(m.height)

	// Join panels horizontally
	return lipgloss.JoinHorizontal(lipgloss.Top, mainPanel, "  ", varPanel)
}

func StartREPL() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}
