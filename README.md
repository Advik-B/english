# English Language Interpreter

A programming language interpreter with natural English syntax, built using Go with a beautiful TUI powered by Charm's Bubble Tea and Cobra CLI framework.

## 🌟 Features

- **Natural English Syntax**: Write code using English keywords and natural language constructs
- **Case-Insensitive Keywords**: Keywords like `declare`, `print`, `if` work in any case  
- **Rich Error Messages**: Helpful error messages with suggestions (e.g., "perhaps you meant...")
- **Stack Traces**: Full call stack information for debugging runtime errors
- **Interactive REPL**: Beautiful terminal UI with syntax highlighting powered by Bubble Tea
- **Flexible Syntax**: Multiple ways to express the same thing (e.g., `to be always` or `to always be`)
- **Bytecode Compilation**: Compile source files to binary format for faster loading
- **Automatic Bytecode Caching**: Imported files are automatically cached in `__engcache__/` for faster loading
- **Python Transpiler**: Convert English programs to human-readable Python with the `transpile` command

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

# Compile to bytecode (.101 format)
./english compile program.abc           # Creates program.101
./english compile program.abc -o out.101  # Custom output name

# Run bytecode directly (no parsing needed)
./english run program.101

# Transpile to Python (validates the program first)
./english transpile program.abc         # Creates program.abc.py
./english transpile program.101         # Creates program.101.py (from bytecode)

# Show version
./english version

# Show help
./english --help
```

## 📦 Bytecode Format

The English interpreter supports compiling source files to a binary bytecode format (`.101` files). This format:

- **Faster Loading**: No parsing required - the AST is serialized directly
- **Smaller Distribution**: Binary format is typically smaller than source
- **File Extension**: `.101` (because human-readable is `.abc`)

```bash
# Compile source to bytecode
./english compile myprogram.abc

# Run the compiled bytecode
./english run myprogram.101
```

The bytecode format uses a protobuf-like binary serialization of the AST, with:
- Magic bytes (`0x10, 0x1E, 0x4E, 0x47`) for file identification
- Version byte for format compatibility
- Binary-encoded AST nodes with type tags

## 🐍 Python Transpiler

The `transpile` command converts an English program to human-readable Python source code.
The program is validated (parsed and type-checked) before any output is written.

```bash
# Transpile a source file
./english transpile myprogram.abc       # Creates myprogram.abc.py

# Transpile a compiled bytecode file
./english transpile myprogram.101       # Creates myprogram.101.py
```

**What gets translated:**

| English                                      | Python                          |
|----------------------------------------------|---------------------------------|
| `Declare x to be 5.`                         | `x = 5`                         |
| `Declare pi to always be 3.14.`              | `pi = 3.14  # constant`         |
| `Print "hello".`                             | `print("hello")`                |
| `Write "hello".`                             | `print("hello", end="")`        |
| `If x is greater than 5, then ...`           | `if x > 5:`                     |
| `repeat the following while x < 10:`         | `while x < 10:`                 |
| `repeat the following 5 times:`              | `for _ in range(5):`            |
| `for each item in list:`                     | `for item in list:`             |
| `repeat forever:`                            | `while True:`                   |
| `Declare function foo that takes a ...`      | `def foo(a):`                   |
| `Return x.`                                  | `return x`                      |
| `Try doing the following: ... on error: ...` | `try: ... except Exception: ...`|
| `Raise "msg" as NetworkError.`               | `raise NetworkError("msg")`     |
| `Declare NetworkError as an error type.`     | `class NetworkError(Exception): pass` |
| `Declare ages to be a lookup table.`         | `ages = {}`                     |
| `Toggle flag.`                               | `flag = not flag`               |
| `Swap x and y.`                              | `x, y = y, x`                   |

Stdlib function calls are translated to their Python equivalents (e.g. `sqrt(x)` → `math.sqrt(x)`, `keys(table)` → `list(table.keys())`). A small set of helper functions is injected at the top of the file for operations without a direct single-expression Python equivalent.

## 📖 Language Guide

### Variables

```english
Declare x to be 5.
Declare name to be "John".
Declare pi to always be 3.14159.  # Constant (immutable)
Set x to 10.

# Alternative 'let' syntax
let y be 10.
let y be equal to 10.
let y = 10.
let y equal 10.
let constant always be 100.  # Constant
let constant be always 100.  # Constant (alternative)
```

### Scoped Variables

Variables can be declared inside blocks, loops, and functions. Each iteration of a loop creates a new scope:

```english
repeat the following 3 times:
    let temp be 42.  # 'temp' is scoped to each iteration
    Print temp.
thats it.
```

### Boolean Values

```english
# Boolean literals
Declare is_raining to be true.
Declare is_sunny to be false.

# Comparisons with booleans
If is_raining is equal to true, then
    Print "Bring an umbrella!".
thats it.

# Toggle booleans
Toggle is_raining.                    # Flips true to false or false to true
Toggle the value of is_raining.       # Alternative syntax
```

### Memory Location

```english
# Inspect variable memory location
Print the location of x.              # Outputs: 0x...:x (memory address)
```

### Arithmetic

```english
# Basic operations
Declare sum to be 5 + 3.
Declare difference to be 10 - 4.
Declare product to be 6 * 7.
Declare quotient to be 20 / 4.

# Remainder (modulo)
Print the remainder of 17 divided by 5.   # Outputs: 2
Print the remainder of 10 / 3.            # Alternative syntax
```

### Output

```english
Print "Hello, World!".
Print the value of x.
Print the value of x plus 5.

# Multiple arguments (space-separated)
Print "Hello", "World".           # Outputs: Hello World
Print "x =", the value of x.      # Outputs: x = 10

# Write (no newline)
Write "Hello ".
Write "World".
Write "\n".                       # Manual newline

# Escape sequences in strings
Print "Line1\nLine2".             # \n = newline
Print "Tab\tSeparated".           # \t = tab
```

### Import Files

Import code from other files to reuse functions and variables. The English import system supports several natural syntax variations and modes:

```english
# Simple import (imports everything from the file)
Import "math_library.abc".

# Import with "from" syntax
Import from "utilities.abc".

# Selective imports (import only specific items)
Import square, cube and isEven from "math_library.abc".
Import add, multiply from "helpers.abc".  # Comma without "and" also works

# Import everything explicitly
Import everything from "library.abc".
Import all from "library.abc".

# Safe imports (don't run top-level code, only load declarations)
Import all from "library.abc" safely.
Import square and cube from "math.abc" safely.

# Once imported, all functions and variables from the file are available
Call myFunction with 10.
```

**Import Modes:**

- **Normal Import**: Executes all code in the imported file, including top-level statements
- **Selective Import**: Runs the file in an isolated environment and imports only specified items
- **Safe Import**: Only loads declarations (functions, variables) without executing top-level statements

**Bytecode Caching:**

Imported files are automatically compiled to bytecode and cached in the `__engcache__/` directory for faster loading. The cache:
- Automatically regenerates when source files are modified
- Uses SipHash (as per PEP 552) for fast, efficient hashing
- Works transparently - no manual compilation needed
- Similar to Python's `__pycache__` behavior

Example library file (`math_library.abc`):
```english
# Math Library

Print "Initializing math library...".  # This runs on normal import, not on safe import

Declare function square that takes x and does the following:
    Return x * x.
thats it.

Declare pi to always be 3.14159.
```

Example usage:
```english
# Selective import - only gets what you specify
Import square from "math_library.abc".

Declare result to be 0.
Set result to the result of calling square with 5.
Print "Result:", the value of result.  # Outputs: Result: 25

# Safe import - doesn't print "Initializing math library..."
Import all from "math_library.abc" safely.
Print "Pi:", the value of pi.          # Outputs: Pi: 3.14159
```

### Functions

```english
Declare function greet that takes name and does the following:
    Print "Hello".
    Print the value of name.
thats it.

Call greet with "Alice".

# Function with return value
Declare function add that takes a and b and does the following:
    Return the result of a plus b.
thats it.

Declare result to be the result of calling add with 5 and 3.
Print the value of result.
```

### Conditionals

```english
If x is greater than 10, then do the following:
    Print "x is large".
otherwise do the following:
    Print "x is small".
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
    Print the value of x.
thats it.

# Infinite loop with break
repeat forever:
    Set x to x plus 1.
    If x is equal to 10, then
        break out of this loop.
    thats it.
thats it.

# For loop (repeat N times)
Repeat 5 times, do the following:
    Print "Hello".
thats it.

# For-each loop
Declare mylist to be [1, 2, 3, 4, 5].
For each item in mylist, do the following:
    Print the value of item.
thats it.

# Break statement works in all loops
For each item in mylist, do the following:
    If item is equal to 3, then
        break out of this loop.
    thats it.
    Print item.
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
├── token/
│   ├── token.go         # Token type definitions
│   └── token_test.go    # Token tests
├── ast/
│   ├── ast.go           # Abstract Syntax Tree node types
│   └── ast_test.go      # AST tests
├── parser/
│   ├── lexer.go         # Lexical analyzer (tokenizer)
│   ├── parser.go        # Recursive descent parser
│   └── parser_test.go   # Lexer and parser tests
├── vm/
│   ├── vm.go            # Virtual machine (evaluator) and runtime
│   └── vm_test.go       # VM and evaluator tests
├── bytecode/
│   ├── bytecode.go      # Binary serialization of AST
│   └── bytecode_test.go # Bytecode tests
├── transpiler/
│   └── transpiler.go    # AST → Python transpiler
├── examples/            # Example programs
│   ├── hello_world.abc  # Simple hello world
│   ├── fibonacci.abc    # Fibonacci sequence
│   ├── fizzbuzz.abc     # Classic FizzBuzz challenge
│   ├── factorial.abc    # Recursive factorial
│   ├── arrays.abc       # Array operations
│   ├── conditionals.abc # If/else examples
│   ├── loops.abc        # Loop constructs
│   ├── functions.abc    # Function examples
│   ├── turing_machine.abc # Turing completeness demo
│   └── ...              # And many more!
├── go.mod               # Go module definition
└── *.101                # Compiled bytecode files
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
# Run all tests
go test ./...

# Run tests with verbose output
go test ./... -v

# Run tests for specific package
go test ./token/... -v
go test ./ast/... -v  
go test ./parser/... -v
go test ./vm/... -v
go test ./bytecode/... -v

# Run example programs
./english run examples/hello_world.abc
./english run examples/fibonacci.abc
./english run examples/turing_machine.abc

# Compile and run bytecode
./english compile examples/fibonacci.abc
./english run examples/fibonacci.101
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
Print "Hello, World!".
```

### FizzBuzz

```english
Declare i to be 1.
Repeat while i is less than or equal to 100, do the following:
    If i divided by 15 is equal to 0, then do the following:
        Print "FizzBuzz".
    otherwise if i divided by 3 is equal to 0, then do the following:
        Print "Fizz".
    otherwise if i divided by 5 is equal to 0, then do the following:
        Print "Buzz".
    otherwise do the following:
        Print the value of i.
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
Print the value of result.  # Outputs: 120
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

MIT License

## 🙏 Acknowledgments

Built with the amazing [Charm](https://charm.sh/) libraries for beautiful terminal UIs.
