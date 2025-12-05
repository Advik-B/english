package repl

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

// Console provides an interactive terminal interface for the REPL.
// It wraps a Session and provides line editing, prompt display, etc.
type Console struct {
	session *Session
	input   io.Reader
	output  io.Writer
	running bool
}

// NewConsole creates a new interactive console with a fresh session.
func NewConsole() *Console {
	return &Console{
		session: NewSession(),
		input:   os.Stdin,
		output:  os.Stdout,
		running: false,
	}
}

// NewConsoleWithSession creates a console with an existing session.
func NewConsoleWithSession(session *Session) *Console {
	return &Console{
		session: session,
		input:   os.Stdin,
		output:  os.Stdout,
		running: false,
	}
}

// NewConsoleWithIO creates a console with custom input/output streams.
// This is useful for testing.
func NewConsoleWithIO(input io.Reader, output io.Writer) *Console {
	return &Console{
		session: NewSession(),
		input:   input,
		output:  output,
		running: false,
	}
}

// GetSession returns the underlying session.
func (c *Console) GetSession() *Session {
	return c.session
}

// Start begins the interactive REPL loop.
// It returns when the user exits or an error occurs.
func (c *Console) Start() error {
	c.running = true
	c.printWelcome()

	scanner := bufio.NewScanner(c.input)

	for c.running {
		// Display prompt
		c.print(c.session.GetPrompt())

		// Read input
		if !scanner.Scan() {
			// EOF or error
			break
		}

		line := scanner.Text()
		result := c.session.Execute(line)

		// Handle output
		if result.Output != "" {
			c.print(result.Output)
			// Add newline if output doesn't end with one
			if !strings.HasSuffix(result.Output, "\n") {
				c.print("\n")
			}
		}

		// Handle errors
		if result.Error != nil {
			if result.Error == ErrExit {
				c.println("Goodbye!")
				c.running = false
				break
			}
			c.printf("Error: %v\n", result.Error)
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

// Stop signals the console to stop running.
func (c *Console) Stop() {
	c.running = false
}

// IsRunning returns whether the console is currently running.
func (c *Console) IsRunning() bool {
	return c.running
}

// printWelcome displays the welcome message.
func (c *Console) printWelcome() {
	c.println("English Language Interpreter v1.0.0")
	c.println("Type :help for assistance, 'exit' to quit.")
	c.println("")
}

// Helper print methods
func (c *Console) print(s string) {
	fmt.Fprint(c.output, s)
}

func (c *Console) println(s string) {
	fmt.Fprintln(c.output, s)
}

func (c *Console) printf(format string, args ...interface{}) {
	fmt.Fprintf(c.output, format, args...)
}

// InteractiveConsole provides a more feature-rich interactive experience
// similar to Python's REPL with readline-like functionality.
// This is a simplified version; for full readline support, consider
// integrating with a library like github.com/chzyer/readline.
type InteractiveConsole struct {
	*Console
	historyFile string
}

// NewInteractiveConsole creates a new interactive console with history support.
func NewInteractiveConsole() *InteractiveConsole {
	return &InteractiveConsole{
		Console:     NewConsole(),
		historyFile: "",
	}
}

// SetHistoryFile sets the file path for persisting command history.
func (ic *InteractiveConsole) SetHistoryFile(path string) {
	ic.historyFile = path
}

// LoadHistory loads command history from the history file.
func (ic *InteractiveConsole) LoadHistory() error {
	if ic.historyFile == "" {
		return nil
	}

	file, err := os.Open(ic.historyFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // History file doesn't exist yet
		}
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		ic.session.history = append(ic.session.history, scanner.Text())
	}

	return scanner.Err()
}

// SaveHistory saves command history to the history file.
func (ic *InteractiveConsole) SaveHistory() error {
	if ic.historyFile == "" {
		return nil
	}

	file, err := os.Create(ic.historyFile)
	if err != nil {
		return err
	}
	defer file.Close()

	for _, cmd := range ic.session.history {
		fmt.Fprintln(file, cmd)
	}

	return nil
}
