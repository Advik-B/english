# English Language Interpreter

A programming language interpreter with natural English syntax, built using Go with a beautiful TUI powered by Charm's Bubble Tea and Cobra CLI framework.

## 🌟 Features

- **Natural English Syntax**: Write code using English keywords and natural language constructs
- **Case-Insensitive Keywords**: Keywords like `declare`, `say`, `if` work in any case  
- **Rich Error Messages**: Helpful error messages with suggestions (e.g., "perhaps you meant...")
- **Stack Traces**: Full call stack information for debugging runtime errors
- **Interactive REPL**: Beautiful terminal UI with syntax highlighting powered by Bubble Tea
- **Flexible Syntax**: Multiple ways to express the same thing (e.g., `to be always` or `to always be`)

## 🚀 Quick Start

### Installation

```bash
# Build the interpreter
go build -o english .
```

### Usage

```bash
# Interactive REPL with beautiful TUI
./english

# Run a source file
./english run program.abc
# or simply
./english program.abc

# Show version
./english version

# Show help
./english --help
```

## 📖 Language Guide

### Variables

```english
Declare x to be 5.
Declare name to be "John".
Declare pi to always be 3.14159.  # Constant (immutable)
Set x to 10.
```

### Arithmetic

```english
# Basic operations
Declare sum to be 5 + 3.
Declare difference to be 10 - 4.
Declare product to be 6 * 7.
Declare quotient to be 20 / 4.

# Remainder (modulo)
Say the remainder of 17 divided by 5.   # Outputs: 2
Say the remainder of 10 / 3.            # Alternative syntax
```

### Output

```english
Say "Hello, World!".
Say the value of x.
Say the value of x plus 5.
```

### Functions

```english
Declare function greet that takes name and does the following:
    Say "Hello".
    Say the value of name.
thats it.

Call greet with "Alice".

# Function with return value
Declare function add that takes a and b and does the following:
    Return the result of a plus b.
thats it.

Declare result to be the result of calling add with 5 and 3.
Say the value of result.
```

### Conditionals

```english
If x is greater than 10, then do the following:
    Say "x is large".
otherwise do the following:
    Say "x is small".
thats it.

# Comparison operators:
# is equal to, is not equal to
# is less than, is greater than
# is less than or equal to, is greater than or equal to
```

### Loops

```english
# While loop
Repeat while x is less than 10, do the following:
    Set x to x plus 1.
    Say the value of x.
thats it.

# For loop (repeat N times)
Repeat 5 times, do the following:
    Say "Hello".
thats it.

# For-each loop
Declare mylist to be [1, 2, 3, 4, 5].
For each item in mylist, do the following:
    Say the value of item.
thats it.
```

### Lists/Arrays

```english
Declare mylist to be [1, 2, 3, 4, 5].
Declare names to be ["Alice", "Bob", "Charlie"].
```

## 🎨 REPL Features

The interactive REPL (Read-Eval-Print Loop) features:

- **Syntax Highlighting**: Keywords, strings, numbers, and operators are color-coded
- **Multi-line Support**: Automatically detects multi-line blocks (ends with `thats it.`)
- **Command History**: Navigate through previous commands
- **Commands**:
  - `:help` or `:h` - Show help
  - `:clear` or `:cls` - Clear screen
  - `:exit` or `:quit` or `:q` - Exit REPL
  - `Ctrl+C` or `Esc` - Exit REPL

## 📁 Project Structure

```
.
├── main.go              # Entry point - calls cmd.Execute()
├── cmd/
│   ├── root.go          # Cobra CLI setup with subcommands
│   └── repl.go          # Bubble Tea REPL implementation
├── interpreter/
│   ├── tokens.go        # Token type definitions
│   ├── lexer.go         # Lexical analyzer (tokenizer)
│   ├── ast.go           # Abstract Syntax Tree node types
│   ├── parser.go        # Recursive descent parser
│   ├── evaluator.go     # Tree-walking interpreter with stack traces
│   └── builtins.go      # Built-in functions and value system
├── Makefile             # Build automation
├── go.mod               # Go module definition
└── *.abc                # Example/test source files
```

## 🛠️ Development

### Building

```bash
# Using make
make build

# Or directly with go
go build -o english .
```

### Cleaning

```bash
# Remove duplicate files from previous versions
make clean
```

### Testing

```bash
# Run example programs
./english syntax.abc
./english test_simple.abc
./english test_case_insensitive.abc
./english test_errors.abc
```

## 📦 Dependencies

- **[Cobra](https://github.com/spf13/cobra)** v1.10.2 - CLI framework
- **[Bubble Tea](https://github.com/charmbracelet/bubbletea)** v1.3.10 - TUI framework  
- **[Lipgloss](https://github.com/charmbracelet/lipgloss)** v1.1.0 - Terminal styling

Install dependencies:
```bash
go mod download
```

## 💡 Examples

### Hello World

```english
Say "Hello, World!".
```

### FizzBuzz

```english
Declare i to be 1.
Repeat while i is less than or equal to 100, do the following:
    If i divided by 15 is equal to 0, then do the following:
        Say "FizzBuzz".
    otherwise if i divided by 3 is equal to 0, then do the following:
        Say "Fizz".
    otherwise if i divided by 5 is equal to 0, then do the following:
        Say "Buzz".
    otherwise do the following:
        Say the value of i.
    thats it.
    Set i to i plus 1.
thats it.
```

### Factorial Function

```english
Declare function factorial that takes n and does the following:
    If n is less than or equal to 1, then do the following:
        Return 1.
    otherwise do the following:
        Declare smaller to be the result of calling factorial with n minus 1.
        Return the result of n times smaller.
    thats it.
thats it.

Declare result to be the result of calling factorial with 5.
Say the value of result.  # Outputs: 120
```

## 🎯 Language Features

### Case Insensitivity

Keywords are case-insensitive:
```english
Declare x to be 5.
DECLARE y to be 10.
declare z to be 15.
```

### Flexible Constant Syntax

Both forms work:
```english
Declare pi to be always 3.14159.
Declare e to always be 2.71828.
```

### Error Messages

The interpreter provides helpful error messages:

- **Parse Errors**: Show exact location and context
- **Runtime Errors**: Include full call stack traces
- **Suggestions**: "Perhaps you meant X?" for undefined variables/functions
- **Type Errors**: Clear explanations of type mismatches

## 📄 License

[Your License Here]

## 🤝 Contributing

[Your Contributing Guidelines Here]

## 🙏 Acknowledgments

Built with the amazing [Charm](https://charm.sh/) libraries for beautiful terminal UIs.
