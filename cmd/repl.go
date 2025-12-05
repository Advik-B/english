package cmd

import (
	"english/interpreter"
	"fmt"
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
)

func getWelcomeMessage() string {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("86")).
		Bold(true).
		Render("✨ Welcome to the English Language Interpreter! Type :help for assistance.")
}

type model struct {
	input       string
	history     []string
	output      []string
	multiline   bool
	buffer      []string
	env         *interpreter.Environment
	evaluator   *interpreter.Evaluator
	cursorPos   int
	width       int
	height      int
	quitting    bool
}

func initialModel() model {
	env := interpreter.NewEnvironment()
	return model{
		input:     "",
		history:   []string{},
		output:    []string{getWelcomeMessage()},
		multiline: false,
		buffer:    []string{},
		env:       env,
		evaluator: interpreter.NewEvaluator(env),
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
	case ":exit", ":quit", ":q":
		m.quitting = true
		return "Goodbye!"
	default:
		return errorStyle.Render("Unknown command: "+cmd) + helpStyle.Render(" (try :help)")
	}
}

func (m *model) getHelp() string {
	help := lipgloss.NewStyle().Foreground(lipgloss.Color("86")).Bold(true).Render("Available Commands:") + "\n"
	help += "  " + successStyle.Render(":help") + "   - Show this help\n"
	help += "  " + successStyle.Render(":clear") + "  - Clear the screen\n"
	help += "  " + successStyle.Render(":exit") + "   - Exit REPL\n\n"
	help += lipgloss.NewStyle().Foreground(lipgloss.Color("86")).Bold(true).Render("Quick Start:") + "\n"
	help += "  " + commentStyle.Render("# Declare a variable") + "\n"
	help += "  " + m.highlightSyntax("Declare x to be 5.") + "\n"
	help += "  " + m.highlightSyntax("Say the value of x.") + "\n"
	return help
}

func (m model) executeCode(code string) (string, error) {
	lexer := interpreter.NewLexer(code)
	tokens := lexer.TokenizeAll()

	parser := interpreter.NewParser(tokens)
	program, err := parser.Parse()
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
		"call": true, "return": true, "say": true, "if": true,
		"then": true, "otherwise": true, "repeat": true, "while": true,
		"times": true, "for": true, "each": true, "in": true,
		"do": true, "takes": true, "and": true, "with": true,
		"result": true, "of": true, "calling": true, "value": true,
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

func (m model) View() string {
	if m.quitting {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("86")).
			Bold(true).
			Render("\n✨ Goodbye! Happy coding!\n\n")
	}

	var s strings.Builder

	// Title
	title := titleStyle.Render(" English Language Interpreter v1.0.0 ")
	s.WriteString(title + "\n")
	s.WriteString(helpStyle.Render("Type :help for commands, Ctrl+C to exit") + "\n\n")

	// Output history
	maxLines := m.height - 6
	if len(m.output) > maxLines {
		for _, line := range m.output[len(m.output)-maxLines:] {
			s.WriteString(line + "\n")
		}
	} else {
		for _, line := range m.output {
			s.WriteString(line + "\n")
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

	s.WriteString("\n" + prompt + inputDisplay)

	return s.String()
}

func StartREPL() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}
