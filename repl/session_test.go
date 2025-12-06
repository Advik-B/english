package repl

import (
	"bytes"
	"english/vm"
	"strings"
	"testing"
)

// ============================================
// SESSION TESTS
// ============================================

func TestNewSession(t *testing.T) {
	session := NewSession()
	if session == nil {
		t.Fatal("NewSession returned nil")
	}
	if session.env == nil {
		t.Error("Session environment is nil")
	}
	if session.evaluator == nil {
		t.Error("Session evaluator is nil")
	}
	if len(session.history) != 0 {
		t.Error("Session history should be empty")
	}
}

func TestSessionExecuteSimple(t *testing.T) {
	session := NewSession()

	// Test variable declaration
	result := session.Execute("Declare x to be 5.")
	if result.Error != nil {
		t.Fatalf("Execute error: %v", result.Error)
	}
	if !result.IsComplete {
		t.Error("Result should be complete")
	}

	// Verify variable was set
	val, ok := session.env.Get("x")
	if !ok {
		t.Error("Variable 'x' not found in environment")
	}
	if val != float64(5) {
		t.Errorf("Expected x=5, got x=%v", val)
	}
}

func TestSessionResultValue(t *testing.T) {
	session := NewSession()

	// ExecuteMultiLine returns the Value from evaluation
	code := `Declare x to be 42.
Declare y to be 8.
Declare z to be x + y.`

	result := session.ExecuteMultiLine(code)
	if result.Error != nil {
		t.Fatalf("Execute error: %v", result.Error)
	}

	// Value should be set (though for assignments it may be nil)
	// Check that we can get a value from an expression
	// The last statement "Declare z to be x + y." returns nil for declarations
	// but the code ran successfully
	val, ok := session.env.Get("z")
	if !ok {
		t.Error("Variable 'z' not found")
	}
	if val != float64(50) {
		t.Errorf("Expected z=50, got z=%v", val)
	}
}

func TestSessionExecuteMultiLine(t *testing.T) {
	session := NewSession()

	// Start a multi-line function definition
	result := session.Execute("Declare function greet that does the following:")
	if result.Error != nil {
		t.Fatalf("Execute error: %v", result.Error)
	}
	if result.IsComplete {
		t.Error("Result should not be complete yet")
	}
	if !result.NeedsMoreInput {
		t.Error("Should need more input")
	}
	if !session.IsMultiline() {
		t.Error("Session should be in multiline mode")
	}

	// Continue with function body
	result = session.Execute("    Print \"Hello\".")
	if result.IsComplete {
		t.Error("Result should not be complete yet")
	}

	// End the function
	result = session.Execute("thats it.")
	if result.Error != nil {
		t.Fatalf("Execute error: %v", result.Error)
	}
	if !result.IsComplete {
		t.Error("Result should be complete")
	}
	if session.IsMultiline() {
		t.Error("Session should not be in multiline mode")
	}

	// Verify function was defined
	fn, ok := session.env.GetFunction("greet")
	if !ok {
		t.Error("Function 'greet' not found")
	}
	if fn.Name != "greet" {
		t.Errorf("Expected function name 'greet', got '%s'", fn.Name)
	}
}

func TestSessionExecuteWithOutput(t *testing.T) {
	session := NewSession()

	// Declare a variable
	session.Execute("Declare x to be 42.")

	// Print the value
	result := session.Execute("Print the value of x.")
	if result.Error != nil {
		t.Fatalf("Execute error: %v", result.Error)
	}
	if !strings.Contains(result.Output, "42") {
		t.Errorf("Expected output to contain '42', got '%s'", result.Output)
	}
}

func TestSessionExecuteError(t *testing.T) {
	session := NewSession()

	// Try to use undefined variable
	result := session.Execute("Print the value of undefined.")
	if result.Error == nil {
		t.Error("Expected error for undefined variable")
	}
	if !strings.Contains(result.Error.Error(), "undefined variable") {
		t.Errorf("Error should mention 'undefined variable', got: %v", result.Error)
	}
}

func TestSessionExecuteParseError(t *testing.T) {
	session := NewSession()

	// Try invalid syntax
	result := session.Execute("Declare to be 5.")
	if result.Error == nil {
		t.Error("Expected parse error")
	}
	if !strings.Contains(result.Error.Error(), "parse error") {
		t.Errorf("Error should mention 'parse error', got: %v", result.Error)
	}
}

func TestSessionExit(t *testing.T) {
	tests := []string{"exit", "quit", "exit()", "quit()"}

	for _, cmd := range tests {
		session := NewSession()
		result := session.Execute(cmd)
		if result.Error != ErrExit {
			t.Errorf("Command '%s' should return ErrExit, got: %v", cmd, result.Error)
		}
	}
}

func TestSessionCommands(t *testing.T) {
	session := NewSession()

	// Test :help
	result := session.Execute(":help")
	if !strings.Contains(result.Output, "Commands:") {
		t.Error(":help should show commands")
	}

	// Test :vars when empty
	result = session.Execute(":vars")
	if !strings.Contains(result.Output, "No variables") {
		t.Errorf(":vars should show 'No variables', got: %s", result.Output)
	}

	// Add a variable
	session.Execute("Declare x to be 5.")

	// Test :vars with variable
	result = session.Execute(":vars")
	if !strings.Contains(result.Output, "x") {
		t.Errorf(":vars should show 'x', got: %s", result.Output)
	}
	if !strings.Contains(result.Output, "5") {
		t.Errorf(":vars should show '5', got: %s", result.Output)
	}
}

func TestSessionHistory(t *testing.T) {
	session := NewSession()

	// Execute some commands
	session.Execute("Declare x to be 5.")
	session.Execute("Declare y to be 10.")
	session.Execute("Print the value of x.")

	history := session.GetHistory()
	if len(history) != 3 {
		t.Errorf("Expected 3 history items, got %d", len(history))
	}

	// Test :history command
	result := session.Execute(":history")
	if !strings.Contains(result.Output, "Declare x to be 5.") {
		t.Errorf(":history should show commands, got: %s", result.Output)
	}
}

func TestSessionReset(t *testing.T) {
	session := NewSession()

	// Add some state
	session.Execute("Declare x to be 5.")
	session.Execute("Declare function test that does the following:")
	session.Execute("    Print \"test\".")
	session.Execute("thats it.")

	// Verify state exists
	_, ok := session.env.Get("x")
	if !ok {
		t.Error("Variable 'x' should exist before reset")
	}

	// Reset session
	session.Reset()

	// Verify state is cleared
	_, ok = session.env.Get("x")
	if ok {
		t.Error("Variable 'x' should not exist after reset")
	}

	history := session.GetHistory()
	if len(history) != 0 {
		t.Error("History should be empty after reset")
	}
}

func TestSessionExecuteMultiLineCode(t *testing.T) {
	session := NewSession()

	code := `Declare x to be 5.
Declare y to be 10.
Declare z to be x + y.
Print the value of z.`

	result := session.ExecuteMultiLine(code)
	if result.Error != nil {
		t.Fatalf("ExecuteMultiLine error: %v", result.Error)
	}
	if !strings.Contains(result.Output, "15") {
		t.Errorf("Expected output to contain '15', got '%s'", result.Output)
	}
}

func TestSessionWithExistingEnvironment(t *testing.T) {
	// Create environment with pre-existing variable
	env := vm.NewEnvironment()
	env.Define("preExisting", float64(100), false)

	// Create session with this environment
	session := NewSessionWithEnv(env)

	// Verify we can access pre-existing variable
	result := session.Execute("Print the value of preExisting.")
	if result.Error != nil {
		t.Fatalf("Execute error: %v", result.Error)
	}
	if !strings.Contains(result.Output, "100") {
		t.Errorf("Expected output to contain '100', got '%s'", result.Output)
	}
}

func TestSessionPrompt(t *testing.T) {
	session := NewSession()

	// Normal mode
	if session.GetPrompt() != ">>> " {
		t.Errorf("Expected '>>> ', got '%s'", session.GetPrompt())
	}

	// Enter multiline mode
	session.Execute("Declare function test that does the following:")
	if session.GetPrompt() != "... " {
		t.Errorf("Expected '... ', got '%s'", session.GetPrompt())
	}

	// Exit multiline mode
	session.Execute("    Print \"test\".")
	session.Execute("thats it.")
	if session.GetPrompt() != ">>> " {
		t.Errorf("Expected '>>> ', got '%s'", session.GetPrompt())
	}
}

func TestSessionFunctionWithParams(t *testing.T) {
	session := NewSession()

	// Define a function with parameters
	code := `Declare function add that takes a and b and does the following:
    Return a + b.
thats it.`

	result := session.ExecuteMultiLine(code)
	if result.Error != nil {
		t.Fatalf("Function definition error: %v", result.Error)
	}

	// Call the function
	code = `Set result to be the result of calling add with 3 and 7.
Print the value of result.`

	result = session.ExecuteMultiLine(code)
	if result.Error != nil {
		t.Fatalf("Function call error: %v", result.Error)
	}
	if !strings.Contains(result.Output, "10") {
		t.Errorf("Expected output '10', got '%s'", result.Output)
	}
}

func TestSessionLoops(t *testing.T) {
	session := NewSession()

	// Test for loop
	code := `repeat the following 3 times:
    Print "hello".
thats it.`

	result := session.ExecuteMultiLine(code)
	if result.Error != nil {
		t.Fatalf("Loop error: %v", result.Error)
	}

	// Count "hello" occurrences
	count := strings.Count(result.Output, "hello")
	if count != 3 {
		t.Errorf("Expected 3 'hello', got %d", count)
	}
}

func TestSessionConditionals(t *testing.T) {
	session := NewSession()

	// Test if statement
	code := `Declare x to be 10.
If x is greater than 5, then
    Print "big".
otherwise
    Print "small".
thats it.`

	result := session.ExecuteMultiLine(code)
	if result.Error != nil {
		t.Fatalf("Conditional error: %v", result.Error)
	}
	if !strings.Contains(result.Output, "big") {
		t.Errorf("Expected 'big', got '%s'", result.Output)
	}
}

func TestSessionConstants(t *testing.T) {
	session := NewSession()

	// Declare constant
	session.Execute("Declare PI to always be 3.14.")

	// Try to reassign
	result := session.Execute("Set PI to be 3.0.")
	if result.Error == nil {
		t.Error("Expected error when reassigning constant")
	}
	if !strings.Contains(result.Error.Error(), "cannot reassign constant") {
		t.Errorf("Error should mention constant, got: %v", result.Error)
	}

	// Verify :vars shows const marker
	result = session.Execute(":vars")
	if !strings.Contains(result.Output, "const") {
		t.Errorf(":vars should show 'const', got: %s", result.Output)
	}
}

func TestSessionLists(t *testing.T) {
	session := NewSession()

	// Create a list
	session.Execute("Declare myList to be [1, 2, 3, 4, 5].")

	// Access element
	result := session.Execute("Print the item at position 2 in myList.")
	if result.Error != nil {
		t.Fatalf("List access error: %v", result.Error)
	}
	if !strings.Contains(result.Output, "3") {
		t.Errorf("Expected '3', got '%s'", result.Output)
	}

	// Get length
	result = session.Execute("Print the length of myList.")
	if !strings.Contains(result.Output, "5") {
		t.Errorf("Expected '5', got '%s'", result.Output)
	}
}

func TestSessionUnknownCommand(t *testing.T) {
	session := NewSession()

	result := session.Execute(":unknown")
	if !strings.Contains(result.Output, "Unknown command") {
		t.Errorf("Should show 'Unknown command', got: %s", result.Output)
	}
}

func TestSessionClearCommand(t *testing.T) {
	session := NewSession()

	// Add some history
	session.Execute("Declare x to be 5.")
	session.Execute("Declare y to be 10.")

	// Start multiline
	session.Execute("Declare function test that does the following:")

	// Clear
	result := session.Execute(":clear")
	if !strings.Contains(result.Output, "Cleared") {
		t.Errorf("Expected 'Cleared', got: %s", result.Output)
	}

	// Verify state
	if len(session.history) != 0 {
		t.Error("History should be empty after :clear")
	}
	if len(session.buffer) != 0 {
		t.Error("Buffer should be empty after :clear")
	}
	if session.multiline {
		t.Error("Should not be in multiline mode after :clear")
	}
}

// ============================================
// CONSOLE TESTS
// ============================================

func TestNewConsole(t *testing.T) {
	console := NewConsole()
	if console == nil {
		t.Fatal("NewConsole returned nil")
	}
	if console.session == nil {
		t.Error("Console session is nil")
	}
}

func TestConsoleWithCustomIO(t *testing.T) {
	input := strings.NewReader("Declare x to be 5.\nexit\n")
	output := &bytes.Buffer{}

	console := NewConsoleWithIO(input, output)
	err := console.Start()
	if err != nil {
		t.Fatalf("Console error: %v", err)
	}

	// Check output contains welcome message
	if !strings.Contains(output.String(), "English Language Interpreter") {
		t.Error("Output should contain welcome message")
	}
}

func TestConsoleSession(t *testing.T) {
	session := NewSession()
	session.Execute("Declare preSet to be 100.")

	console := NewConsoleWithSession(session)
	if console.GetSession() != session {
		t.Error("GetSession should return the provided session")
	}

	// Verify pre-existing state
	val, ok := console.GetSession().env.Get("preSet")
	if !ok || val != float64(100) {
		t.Error("Pre-existing variable should be accessible")
	}
}

// ============================================
// INTEGRATION TESTS
// ============================================

func TestCompleteREPLScenario(t *testing.T) {
	session := NewSession()

	// Define a function
	session.Execute("Declare function factorial that takes n and does the following:")
	session.Execute("    If n is less than or equal to 1, then")
	session.Execute("        Return 1.")
	session.Execute("    otherwise")
	session.Execute("        Set prev to be the result of calling factorial with n - 1.")
	session.Execute("        Return n * prev.")
	session.Execute("    thats it.")
	result := session.Execute("thats it.")

	if result.Error != nil {
		t.Fatalf("Function definition error: %v", result.Error)
	}

	// Call factorial
	result = session.Execute("Set result to be the result of calling factorial with 5.")
	if result.Error != nil {
		t.Fatalf("Function call error: %v", result.Error)
	}

	// Print result
	result = session.Execute("Print the value of result.")
	if result.Error != nil {
		t.Fatalf("Print error: %v", result.Error)
	}

	if !strings.Contains(result.Output, "120") {
		t.Errorf("Expected factorial(5) = 120, got: %s", result.Output)
	}
}

func TestStringOperations(t *testing.T) {
	session := NewSession()

	// String concatenation
	session.Execute("Declare greeting to be \"Hello\" + \" \" + \"World\".")
	result := session.Execute("Print the value of greeting.")
	if !strings.Contains(result.Output, "Hello World") {
		t.Errorf("Expected 'Hello World', got: %s", result.Output)
	}

	// String multiplication
	session.Execute("Declare repeated to be \"ab\" * 3.")
	result = session.Execute("Print the value of repeated.")
	if !strings.Contains(result.Output, "ababab") {
		t.Errorf("Expected 'ababab', got: %s", result.Output)
	}
}

func TestArithmeticOperations(t *testing.T) {
	session := NewSession()

	tests := []struct {
		code     string
		expected string
	}{
		{"Declare x to be 10 + 5.", ""},
		{"Print the value of x.", "15"},
		{"Set x to be 10 - 3.", ""},
		{"Print the value of x.", "7"},
		{"Set x to be 4 * 3.", ""},
		{"Print the value of x.", "12"},
		{"Set x to be 15 / 3.", ""},
		{"Print the value of x.", "5"},
		{"Print the remainder of 17 divided by 5.", "2"},
	}

	for _, test := range tests {
		result := session.Execute(test.code)
		if result.Error != nil {
			t.Fatalf("Code '%s' error: %v", test.code, result.Error)
		}
		if test.expected != "" && !strings.Contains(result.Output, test.expected) {
			t.Errorf("Code '%s': expected '%s', got '%s'", test.code, test.expected, result.Output)
		}
	}
}

func TestToggle(t *testing.T) {
	session := NewSession()

	session.Execute("Declare flag to be true.")
	result := session.Execute("Print the value of flag.")
	if !strings.Contains(result.Output, "true") {
		t.Errorf("Expected 'true', got: %s", result.Output)
	}

	session.Execute("Toggle flag.")
	result = session.Execute("Print the value of flag.")
	if !strings.Contains(result.Output, "false") {
		t.Errorf("Expected 'false', got: %s", result.Output)
	}
}

func TestForEachLoop(t *testing.T) {
	session := NewSession()

	code := `Declare numbers to be [10, 20, 30].
for each n in numbers, do the following:
    Print the value of n.
thats it.`

	result := session.ExecuteMultiLine(code)
	if result.Error != nil {
		t.Fatalf("For-each error: %v", result.Error)
	}

	if !strings.Contains(result.Output, "10") ||
		!strings.Contains(result.Output, "20") ||
		!strings.Contains(result.Output, "30") {
		t.Errorf("Expected 10, 20, 30 in output, got: %s", result.Output)
	}
}

func TestWhileLoop(t *testing.T) {
	session := NewSession()

	code := `Declare counter to be 0.
repeat the following while counter is less than 3:
    Print the value of counter.
    Set counter to be counter + 1.
thats it.`

	result := session.ExecuteMultiLine(code)
	if result.Error != nil {
		t.Fatalf("While loop error: %v", result.Error)
	}

	if !strings.Contains(result.Output, "0") ||
		!strings.Contains(result.Output, "1") ||
		!strings.Contains(result.Output, "2") {
		t.Errorf("Expected 0, 1, 2 in output, got: %s", result.Output)
	}
}
