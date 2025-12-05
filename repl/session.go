// Package repl provides an interactive Read-Eval-Print Loop for the English
// programming language. It is designed similar to Python's REPL, supporting
// persistent state, multi-line input, command history, and programmatic access
// for testing.
package repl

import (
	"english/parser"
	"english/vm"
	"fmt"
	"strings"
)

// Result represents the result of executing a line of code.
type Result struct {
	// Output contains any printed output from the code
	Output string
	// Value contains the result value (if any)
	Value vm.Value
	// Error contains any error that occurred
	Error error
	// IsComplete indicates whether the input was a complete statement
	IsComplete bool
	// NeedsMoreInput indicates the REPL needs more input (multi-line mode)
	NeedsMoreInput bool
}

// Session represents an interactive REPL session with persistent state.
// It can be used programmatically for testing or with an interactive shell.
type Session struct {
	env         *vm.Environment
	evaluator   *vm.Evaluator
	history     []string
	buffer      []string
	multiline   bool
	nestingDepth int  // Track nesting depth for multi-line blocks
	output      *strings.Builder
}

// NewSession creates a new REPL session with a fresh environment.
func NewSession() *Session {
	env := vm.NewEnvironment()
	return &Session{
		env:          env,
		evaluator:    vm.NewEvaluator(env),
		history:      make([]string, 0),
		buffer:       make([]string, 0),
		multiline:    false,
		nestingDepth: 0,
		output:       &strings.Builder{},
	}
}

// NewSessionWithEnv creates a new REPL session with an existing environment.
// This is useful for testing scenarios where you want to pre-populate variables.
func NewSessionWithEnv(env *vm.Environment) *Session {
	return &Session{
		env:          env,
		evaluator:    vm.NewEvaluator(env),
		history:      make([]string, 0),
		buffer:       make([]string, 0),
		multiline:    false,
		nestingDepth: 0,
		output:       &strings.Builder{},
	}
}

// Execute processes a line of input and returns the result.
// This is the main method for interacting with the REPL programmatically.
func (s *Session) Execute(line string) Result {
	trimmed := strings.TrimSpace(line)

	// Handle empty input
	if trimmed == "" && !s.multiline {
		return Result{IsComplete: true}
	}

	// Handle built-in commands
	if strings.HasPrefix(trimmed, ":") {
		cmdResult := s.handleCommand(trimmed)
		return Result{
			Output:     cmdResult,
			IsComplete: true,
		}
	}

	// Handle exit commands
	if trimmed == "exit" || trimmed == "quit" || trimmed == "exit()" || trimmed == "quit()" {
		return Result{
			Output:     "Goodbye!",
			IsComplete: true,
			Error:      ErrExit,
		}
	}

	// Add to history if non-empty
	if trimmed != "" {
		s.history = append(s.history, line)
	}

	// Add to buffer
	s.buffer = append(s.buffer, line)

	lower := strings.ToLower(trimmed)

	// Check for multiline start (function definitions, loops, etc.)
	// These constructs start a new nesting level
	if strings.Contains(lower, "following:") || strings.Contains(lower, ", then") {
		s.nestingDepth++
		if s.nestingDepth == 1 {
			s.multiline = true
		}
		return Result{
			IsComplete:     false,
			NeedsMoreInput: true,
		}
	}

	// Check for multiline end
	if strings.Contains(lower, "thats it.") {
		s.nestingDepth--
		if s.nestingDepth < 0 {
			s.nestingDepth = 0
		}
		// Only execute when we're back to top level
		if s.nestingDepth == 0 {
			s.multiline = false
			code := strings.Join(s.buffer, "\n")
			s.buffer = s.buffer[:0]
			return s.executeCode(code)
		}
		// Still nested, wait for more input
		return Result{
			IsComplete:     false,
			NeedsMoreInput: true,
		}
	}

	// If in multiline mode, wait for more input
	if s.multiline {
		return Result{
			IsComplete:     false,
			NeedsMoreInput: true,
		}
	}

	// Execute single line if it ends with a period
	if strings.HasSuffix(trimmed, ".") {
		code := strings.Join(s.buffer, "\n")
		s.buffer = s.buffer[:0]
		return s.executeCode(code)
	}

	// Otherwise, wait for more input
	return Result{
		IsComplete:     false,
		NeedsMoreInput: true,
	}
}

// ExecuteMultiLine executes multiple lines of code at once.
// This is useful for running scripts or test cases.
// Unlike Execute, this parses the entire code block at once rather than
// line by line, which is more appropriate for complete programs.
func (s *Session) ExecuteMultiLine(code string) Result {
	// For complete code blocks, parse and execute directly
	// This handles multi-line constructs properly
	return s.executeCode(code)
}

// executeCode parses and evaluates code, capturing output.
func (s *Session) executeCode(code string) Result {
	// Parse the code
	lexer := parser.NewLexer(code)
	tokens := lexer.TokenizeAll()

	p := parser.NewParser(tokens)
	program, err := p.Parse()
	if err != nil {
		return Result{
			Error:      fmt.Errorf("parse error: %v", err),
			IsComplete: true,
		}
	}

	// Capture stdout for output
	capturedOutput := captureStdout(func() {
		_, err = s.evaluator.Eval(program)
	})

	if err != nil {
		return Result{
			Output:     capturedOutput,
			Error:      fmt.Errorf("runtime error: %v", err),
			IsComplete: true,
		}
	}

	return Result{
		Output:     capturedOutput,
		IsComplete: true,
	}
}

// handleCommand processes REPL commands (starting with ':')
func (s *Session) handleCommand(cmd string) string {
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return ""
	}

	switch parts[0] {
	case ":help", ":h", ":?":
		return s.getHelp()
	case ":clear", ":cls":
		s.history = s.history[:0]
		s.buffer = s.buffer[:0]
		s.multiline = false
		s.nestingDepth = 0
		return "Cleared history and buffer."
	case ":reset":
		s.Reset()
		return "Session reset. All variables and functions cleared."
	case ":vars", ":v":
		return s.listVariables()
	case ":funcs", ":f":
		return s.listFunctions()
	case ":history", ":hist":
		return s.formatHistory()
	case ":exit", ":quit", ":q":
		return "Use 'exit' or 'quit' to exit the REPL."
	default:
		return fmt.Sprintf("Unknown command: %s (try :help)", parts[0])
	}
}

// getHelp returns the help text for REPL commands.
func (s *Session) getHelp() string {
	return `English REPL - Interactive Programming Environment

Commands:
  :help, :h, :?    - Show this help message
  :vars, :v        - List all defined variables
  :funcs, :f       - List all defined functions
  :history, :hist  - Show command history
  :clear, :cls     - Clear history and input buffer
  :reset           - Reset session (clear all variables and functions)
  :exit, :quit, :q - Exit the REPL

Quick Start:
  Declare x to be 5.             # Declare a variable
  Declare y to always be 10.     # Declare a constant
  Set x to be 15.                # Reassign a variable
  Print the value of x.          # Print a value
  Print "Hello, World!".         # Print a string

Multi-line (functions, loops, conditionals):
  Declare function add that takes a and b and does the following:
      Return a + b.
  thats it.

  If x is greater than 5, then
      Print "x is big".
  thats it.

  repeat the following 3 times:
      Print "Hello".
  thats it.

Exit: Type 'exit', 'quit', or use :exit`
}

// listVariables returns a formatted list of all variables.
func (s *Session) listVariables() string {
	vars := s.env.GetAllVariables()
	if len(vars) == 0 {
		return "No variables defined."
	}

	var sb strings.Builder
	sb.WriteString("Variables:\n")
	for name, value := range vars {
		constMark := ""
		if s.env.IsConstant(name) {
			constMark = " (const)"
		}
		sb.WriteString(fmt.Sprintf("  %s = %s%s\n", name, vm.ToString(value), constMark))
	}
	return sb.String()
}

// listFunctions returns a formatted list of all functions.
func (s *Session) listFunctions() string {
	funcs := s.env.GetAllFunctions()
	if len(funcs) == 0 {
		return "No functions defined."
	}

	var sb strings.Builder
	sb.WriteString("Functions:\n")
	for name, fn := range funcs {
		params := strings.Join(fn.Parameters, ", ")
		sb.WriteString(fmt.Sprintf("  %s(%s)\n", name, params))
	}
	return sb.String()
}

// formatHistory returns a formatted command history.
func (s *Session) formatHistory() string {
	if len(s.history) == 0 {
		return "No command history."
	}

	var sb strings.Builder
	sb.WriteString("Command History:\n")
	for i, cmd := range s.history {
		sb.WriteString(fmt.Sprintf("  %d: %s\n", i+1, cmd))
	}
	return sb.String()
}

// Reset clears the session state, including all variables and functions.
func (s *Session) Reset() {
	s.env = vm.NewEnvironment()
	s.evaluator = vm.NewEvaluator(s.env)
	s.history = s.history[:0]
	s.buffer = s.buffer[:0]
	s.multiline = false
	s.nestingDepth = 0
}

// GetEnvironment returns the current environment.
// Useful for inspecting state in tests.
func (s *Session) GetEnvironment() *vm.Environment {
	return s.env
}

// GetHistory returns the command history.
func (s *Session) GetHistory() []string {
	result := make([]string, len(s.history))
	copy(result, s.history)
	return result
}

// IsMultiline returns whether the session is in multiline mode.
func (s *Session) IsMultiline() bool {
	return s.multiline
}

// GetPrompt returns the appropriate prompt string for the current state.
func (s *Session) GetPrompt() string {
	if s.multiline {
		return "... "
	}
	return ">>> "
}
