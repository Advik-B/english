package cmd

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// TestInitialModel verifies the initial model state
func TestInitialModel(t *testing.T) {
	m := initialModel()

	if m.input != "" {
		t.Errorf("Expected empty input, got: %s", m.input)
	}
	if m.cursorPos != 0 {
		t.Errorf("Expected cursor at position 0, got: %d", m.cursorPos)
	}
	if m.multiline {
		t.Error("Expected multiline to be false")
	}
	if m.quitting {
		t.Error("Expected quitting to be false")
	}
	if m.env == nil {
		t.Error("Expected environment to be initialized")
	}
	if m.evaluator == nil {
		t.Error("Expected evaluator to be initialized")
	}
	if len(m.output) != 1 {
		t.Errorf("Expected 1 welcome message in output, got: %d", len(m.output))
	}
}

// TestModelUpdate_TextInput verifies text input handling
func TestModelUpdate_TextInput(t *testing.T) {
	m := initialModel()

	// Type some text
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("test")}
	newModel, _ := m.Update(msg)
	m = newModel.(model)

	if m.input != "test" {
		t.Errorf("Expected input 'test', got: %s", m.input)
	}
	if m.cursorPos != 4 {
		t.Errorf("Expected cursor at position 4, got: %d", m.cursorPos)
	}
}

// TestModelUpdate_Backspace verifies backspace handling
func TestModelUpdate_Backspace(t *testing.T) {
	m := initialModel()
	m.input = "test"
	m.cursorPos = 4

	// Press backspace
	msg := tea.KeyMsg{Type: tea.KeyBackspace}
	newModel, _ := m.Update(msg)
	m = newModel.(model)

	if m.input != "tes" {
		t.Errorf("Expected input 'tes', got: %s", m.input)
	}
	if m.cursorPos != 3 {
		t.Errorf("Expected cursor at position 3, got: %d", m.cursorPos)
	}

	// Backspace at position 0 should not change anything
	m.cursorPos = 0
	newModel, _ = m.Update(msg)
	m = newModel.(model)
	if m.input != "tes" {
		t.Errorf("Expected input unchanged, got: %s", m.input)
	}
}

// TestModelUpdate_CursorMovement verifies cursor movement
func TestModelUpdate_CursorMovement(t *testing.T) {
	m := initialModel()
	m.input = "testing"
	m.cursorPos = 4

	tests := []struct {
		name        string
		key         tea.KeyType
		expectedPos int
	}{
		{"Left", tea.KeyLeft, 3},
		{"Left again", tea.KeyLeft, 2},
		{"Right", tea.KeyRight, 3},
		{"Home", tea.KeyHome, 0},
		{"End", tea.KeyEnd, 7},
		{"Right at end", tea.KeyRight, 7},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := tea.KeyMsg{Type: tt.key}
			newModel, _ := m.Update(msg)
			m = newModel.(model)
			if m.cursorPos != tt.expectedPos {
				t.Errorf("Expected cursor at %d, got: %d", tt.expectedPos, m.cursorPos)
			}
		})
	}
}

// TestModelUpdate_CtrlC verifies Ctrl+C exits
func TestModelUpdate_CtrlC(t *testing.T) {
	m := initialModel()

	msg := tea.KeyMsg{Type: tea.KeyCtrlC}
	newModel, cmd := m.Update(msg)
	m = newModel.(model)

	if !m.quitting {
		t.Error("Expected quitting to be true after Ctrl+C")
	}
	if cmd == nil {
		t.Error("Expected quit command to be returned")
	}
}

// TestModelUpdate_Esc verifies Esc exits
func TestModelUpdate_Esc(t *testing.T) {
	m := initialModel()

	msg := tea.KeyMsg{Type: tea.KeyEsc}
	newModel, cmd := m.Update(msg)
	m = newModel.(model)

	if !m.quitting {
		t.Error("Expected quitting to be true after Esc")
	}
	if cmd == nil {
		t.Error("Expected quit command to be returned")
	}
}

// TestModelUpdate_WindowSize verifies window resize handling
func TestModelUpdate_WindowSize(t *testing.T) {
	m := initialModel()

	msg := tea.WindowSizeMsg{Width: 100, Height: 50}
	newModel, _ := m.Update(msg)
	m = newModel.(model)

	if m.width != 100 {
		t.Errorf("Expected width 100, got: %d", m.width)
	}
	if m.height != 50 {
		t.Errorf("Expected height 50, got: %d", m.height)
	}
}

// TestHandleEnter_Command verifies command handling
func TestHandleEnter_Command(t *testing.T) {
	m := initialModel()
	outputLen := len(m.output)

	// Test :help command
	m.input = ":help"
	m.cursorPos = len(m.input)
	newModel, _ := m.handleEnter()
	m = newModel.(model)

	if len(m.output) <= outputLen {
		t.Error("Expected output to be added after :help command")
	}
	if !strings.Contains(m.output[len(m.output)-1], "Commands") {
		t.Error("Expected help text to contain 'Commands'")
	}
	if m.input != "" {
		t.Errorf("Expected input to be cleared, got: %s", m.input)
	}
	if m.cursorPos != 0 {
		t.Errorf("Expected cursor at position 0, got: %d", m.cursorPos)
	}
}

// TestHandleEnter_Exit verifies exit handling
func TestHandleEnter_Exit(t *testing.T) {
	tests := []string{"exit", "quit"}
	for _, cmd := range tests {
		t.Run(cmd, func(t *testing.T) {
			m := initialModel()
			m.input = cmd
			m.cursorPos = len(m.input)
			newModel, _ := m.handleEnter()
			m = newModel.(model)

			if !m.quitting {
				t.Errorf("Expected quitting to be true after '%s'", cmd)
			}
		})
	}
}

// TestHandleEnter_EmptyInput verifies empty input handling
func TestHandleEnter_EmptyInput(t *testing.T) {
	m := initialModel()
	outputLen := len(m.output)

	m.input = "   "
	m.cursorPos = len(m.input)
	newModel, _ := m.handleEnter()
	m = newModel.(model)

	if len(m.output) != outputLen {
		t.Error("Expected no output for empty input")
	}
	if m.input != "" {
		t.Error("Expected input to be cleared")
	}
}

// TestHandleEnter_CodeExecution verifies code execution
func TestHandleEnter_CodeExecution(t *testing.T) {
	m := initialModel()

	// Execute a simple declaration
	m.input = "Declare x to be 5."
	m.cursorPos = len(m.input)
	newModel, _ := m.handleEnter()
	m = newModel.(model)

	// Check that code was executed successfully
	if len(m.output) == 0 {
		t.Error("Expected output after code execution")
	}

	// Verify variable was set in environment
	val, ok := m.env.Get("x")
	if !ok {
		t.Error("Expected variable 'x' to be defined")
	}
	if val != float64(5) {
		t.Errorf("Expected x=5, got: %v", val)
	}
}

// TestHandleEnter_MultilineStart verifies multiline mode activation
func TestHandleEnter_MultilineStart(t *testing.T) {
	m := initialModel()

	// Start a function definition
	m.input = "Declare function test that does the following:"
	m.cursorPos = len(m.input)
	newModel, _ := m.handleEnter()
	m = newModel.(model)

	if !m.multiline {
		t.Error("Expected multiline mode to be activated")
	}
	if len(m.buffer) != 1 {
		t.Errorf("Expected 1 line in buffer, got: %d", len(m.buffer))
	}
}

// TestHandleEnter_MultilineEnd verifies multiline mode deactivation
func TestHandleEnter_MultilineEnd(t *testing.T) {
	m := initialModel()

	// Start multiline
	m.input = "Declare function mytest that does the following:"
	m.cursorPos = len(m.input)
	newModel, _ := m.handleEnter()
	m = newModel.(model)

	// Continue multiline
	m.input = "    Print \"hello\"."
	m.cursorPos = len(m.input)
	newModel, _ = m.handleEnter()
	m = newModel.(model)

	// End multiline
	m.input = "thats it."
	m.cursorPos = len(m.input)
	newModel, _ = m.handleEnter()
	m = newModel.(model)

	if m.multiline {
		t.Error("Expected multiline mode to be deactivated")
	}
	if len(m.buffer) != 0 {
		t.Errorf("Expected buffer to be cleared, got: %d lines", len(m.buffer))
	}

	// Verify function was defined
	_, ok := m.env.GetFunction("mytest")
	if !ok {
		t.Error("Expected function 'mytest' to be defined")
	}
}

// TestHandleCommand_Vars verifies :vars command
func TestHandleCommand_Vars(t *testing.T) {
	m := initialModel()

	// Test with no user-defined variables (but there are built-in functions)
	output := m.handleCommand(":vars")
	// Should show built-in functions
	if !strings.Contains(output, "Functions") {
		t.Errorf("Expected output to show built-in functions, got: %s", output)
	}

	// Add a user variable
	m.env.Define("x", float64(5), false)

	// Test with variables
	output = m.handleCommand(":vars")
	if !strings.Contains(output, "x") {
		t.Errorf("Expected output to contain 'x', got: %s", output)
	}
	if !strings.Contains(output, "5") {
		t.Errorf("Expected output to contain '5', got: %s", output)
	}
}

// TestHandleCommand_Clear verifies :clear command
func TestHandleCommand_Clear(t *testing.T) {
	m := initialModel()

	// Add some output and history
	m.output = append(m.output, "line 1", "line 2", "line 3")
	m.history = []string{"cmd1", "cmd2"}

	output := m.handleCommand(":clear")
	if output != "" {
		t.Errorf("Expected empty output, got: %s", output)
	}
	if len(m.output) != 0 {
		t.Errorf("Expected output to be cleared, got: %d lines", len(m.output))
	}
	if len(m.history) != 0 {
		t.Errorf("Expected history to be cleared, got: %d lines", len(m.history))
	}
}

// TestHandleCommand_Help verifies :help command
func TestHandleCommand_Help(t *testing.T) {
	m := initialModel()

	output := m.handleCommand(":help")
	if !strings.Contains(output, "Available Commands") {
		t.Error("Expected help to contain 'Available Commands'")
	}
	if !strings.Contains(output, ":vars") {
		t.Error("Expected help to mention :vars command")
	}
}

// TestHandleCommand_Unknown verifies unknown command handling
func TestHandleCommand_Unknown(t *testing.T) {
	m := initialModel()

	output := m.handleCommand(":unknown")
	if !strings.Contains(output, "Unknown command") {
		t.Errorf("Expected 'Unknown command', got: %s", output)
	}
}

// TestListVariables verifies variable listing
func TestListVariables(t *testing.T) {
	m := initialModel()

	// Add variables
	m.env.Define("x", float64(10), false)
	m.env.Define("PI", 3.14, true)

	output := m.listVariables()

	if !strings.Contains(output, "x") {
		t.Error("Expected output to contain variable 'x'")
	}
	if !strings.Contains(output, "10") {
		t.Error("Expected output to contain value '10'")
	}
	if !strings.Contains(output, "PI") {
		t.Error("Expected output to contain constant 'PI'")
	}
	if !strings.Contains(output, "const") {
		t.Error("Expected output to mark PI as const")
	}
}

// TestExecuteCode_Success verifies successful code execution
func TestExecuteCode_Success(t *testing.T) {
	m := initialModel()

	code := "Declare x to be 42."
	result, err := m.executeCode(code)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if !strings.Contains(result, "successfully") {
		t.Errorf("Expected success message, got: %s", result)
	}

	val, ok := m.env.Get("x")
	if !ok {
		t.Error("Expected variable 'x' to be defined")
	}
	if val != float64(42) {
		t.Errorf("Expected x=42, got: %v", val)
	}
}

// TestExecuteCode_ParseError verifies parse error handling
func TestExecuteCode_ParseError(t *testing.T) {
	m := initialModel()

	code := "Declare to be invalid."
	_, err := m.executeCode(code)

	if err == nil {
		t.Error("Expected parse error")
	}
	if !strings.Contains(err.Error(), "parse error") {
		t.Errorf("Expected parse error message, got: %v", err)
	}
}

// TestExecuteCode_RuntimeError verifies runtime error handling
func TestExecuteCode_RuntimeError(t *testing.T) {
	m := initialModel()

	code := "Print the value of undefined."
	_, err := m.executeCode(code)

	if err == nil {
		t.Error("Expected runtime error")
	}
	if !strings.Contains(err.Error(), "runtime error") {
		t.Errorf("Expected runtime error message, got: %v", err)
	}
}

// TestFormatValue verifies value formatting
func TestFormatValue(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		expected string
	}{
		{"integer", float64(42), "42"},
		{"float", 3.14159, "3.142"},
		{"string", "hello", "\"hello\""},
		{"true", true, "true"},
		{"false", false, "false"},
		{"nil", nil, "nil"},
		{"small array", []interface{}{float64(1), float64(2), float64(3)}, "[1, 2, 3]"},
		{"large array", []interface{}{1, 2, 3, 4, 5}, "[5 items]"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatValue(tt.value)
			if !strings.Contains(result, tt.expected) {
				t.Errorf("Expected result to contain '%s', got: %s", tt.expected, result)
			}
		})
	}
}

// TestView verifies view rendering
func TestView(t *testing.T) {
	m := initialModel()
	m.width = 100
	m.height = 30

	view := m.View()

	if view == "" {
		t.Error("Expected non-empty view")
	}
	if !strings.Contains(view, "English Language Interpreter") {
		t.Error("Expected view to contain title")
	}
	if !strings.Contains(view, ">>>") {
		t.Error("Expected view to contain prompt")
	}
}

// TestView_Quitting verifies quit view
func TestView_Quitting(t *testing.T) {
	m := initialModel()
	m.quitting = true

	view := m.View()

	if !strings.Contains(view, "Goodbye") {
		t.Error("Expected view to contain goodbye message")
	}
}

// TestView_WithVariables verifies variable panel rendering
func TestView_WithVariables(t *testing.T) {
	m := initialModel()
	m.width = 100
	m.height = 30
	m.env.Define("x", float64(5), false)

	view := m.View()

	if !strings.Contains(view, "Variables") {
		t.Error("Expected view to contain Variables panel")
	}
}

// TestGetWelcomeMessage verifies welcome message
func TestGetWelcomeMessage(t *testing.T) {
	msg := getWelcomeMessage()
	if msg == "" {
		t.Error("Expected non-empty welcome message")
	}
	if !strings.Contains(msg, "Welcome") {
		t.Error("Expected welcome message to contain 'Welcome'")
	}
}

// TestHighlightSyntax verifies syntax highlighting
func TestHighlightSyntax(t *testing.T) {
	m := initialModel()

	tests := []struct {
		name  string
		input string
	}{
		{"declaration", "Declare x to be 5."},
		{"function", "function test that does the following:"},
		{"print", "Print the value of x."},
		{"string", "Print \"hello world\"."},
		{"number", "Declare x to be 42."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := m.highlightSyntax(tt.input)
			if result == "" {
				t.Error("Expected non-empty highlighted output")
			}
			// The result should contain ANSI color codes from lipgloss
			// We just verify it returns something
		})
	}
}

// TestRenderVariablesPanel verifies variables panel rendering
func TestRenderVariablesPanel(t *testing.T) {
	m := initialModel()

	// Test with no user-defined variables (but has built-in functions)
	panel := m.renderVariablesPanel(30)
	if panel == "" {
		t.Error("Expected non-empty panel")
	}
	if !strings.Contains(panel, "Functions") {
		t.Error("Expected panel to show built-in functions")
	}

	// Test with variables and functions
	m.env.Define("x", float64(10), false)
	m.env.Define("PI", 3.14, true)

	panel = m.renderVariablesPanel(30)
	if !strings.Contains(panel, "x") {
		t.Error("Expected panel to contain variable 'x'")
	}
	if !strings.Contains(panel, "PI") {
		t.Error("Expected panel to contain constant 'PI'")
	}
}
