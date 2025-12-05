# English Language Interpreter

A complete interpreter for the English programming language, written in Go. The interpreter supports variables, constants, functions, loops, conditionals, and list operations with natural language syntax.

## Files Overview

### Core Components

1. **tokens.go** - Token definitions and types
   - Defines `TokenType` enum for all token types
   - Defines `Token` struct with type, value, line, and column information
   - Maps token types to human-readable strings

2. **lexer.go** - Tokenization/Scanning
   - `Lexer` struct that converts source code into tokens
   - Handles single and multi-character operators
   - Special handling for multi-word operators like "is equal to", "the result of calling"
   - Automatic keyword recognition
   - Comment skipping with `#` prefix
   - String and number literal parsing

3. **ast.go** - Abstract Syntax Tree (AST) Node Definitions
   - `Program` - root node
   - Statements: `VariableDecl`, `Assignment`, `FunctionDecl`, `CallStatement`, `OutputStatement`, `ReturnStatement`, `IfStatement`, `WhileLoop`, `ForLoop`, `ForEachLoop`
   - Expressions: `NumberLiteral`, `StringLiteral`, `ListLiteral`, `Identifier`, `BinaryExpression`, `UnaryExpression`, `FunctionCall`

4. **parser.go** - Recursive Descent Parser
   - `Parser` struct that converts tokens into AST
   - Implements parsing for all language constructs
   - Handles operator precedence for arithmetic and logical expressions
   - Recursive parsing for nested structures

5. **builtins.go** - Value System and Built-in Operations
   - `Value` type interface for runtime values
   - Type conversion functions: `toNumber()`, `toString()`, `toBool()`
   - Comparison operations: `compare()`, `equals()`
   - Arithmetic operations: `add()`, `subtract()`, `multiply()`, `divide()`
   - `FunctionValue` struct for user-defined functions

6. **evaluator.go** - Tree-Walking Interpreter
   - `Environment` struct for scope management (variables, functions, parent scope)
   - `Evaluator` struct that executes the AST
   - Recursive evaluation of all node types
   - Support for function calls with parameter binding
   - Return value handling using `ReturnValue` wrapper
   - Enhanced error messages with call stack traces

7. **main.go** - Entry Point and Enhanced REPL
   - File execution mode: `./english program.abc`
   - **Interactive REPL mode with:**
     - **Syntax highlighting** - Keywords, strings, numbers, operators color-coded
     - **Multi-line support** - Automatic detection of multi-line blocks
     - **Smart prompts** - `>>>` for normal mode, `...` for multi-line mode
     - **Colored output** - Errors, successes, and code beautifully formatted
     - **REPL commands** - `:help`, `:clear`, `:exit`
     - **Persistent environment** - Variables and functions persist across statements
   - Orchestrates the lexer → parser → evaluator pipeline

8. **repl/repl.go** - REPL Module (separate package)
   - Standalone REPL implementation with syntax highlighting
   - Color management and terminal control
   - Helper functions for interactive experience

### Test Files

- **syntax.abc** - Reference syntax documentation with all language features
- **test_simple.abc** - Simple test program demonstrating basic features
- **test_case_insensitive.abc** - Demonstrates case-insensitive keyword support
- **test_errors.abc** - Examples showcasing enhanced error messages

## Language Features

### Variable Declaration
```
Declare x to be 5.
Declare y to be always 10.
Declare z to always be "hello".
```
Both `to be always` and `to always be` syntax are supported for constants.

### Assignment
```
Set x to be 15.
Set x to be x + 1.
```

### Functions
```
Declare function add that takes a and b and does the following:
	Return a + b.
thats it.

Call function_name.
Set result to be the result of calling add with 3 and 7.
```

### Control Flow - If/Else
```
If x is equal to 20, then
	Say "x is 20".
otherwise
	Say "x is not 20".
thats it.

If x is less than 10, then
	Say "small".
otherwise if x is less than 20, then
	Say "medium".
otherwise
	Say "large".
thats it.
```

### Loops - For (Count)
```
repeat the following 5 times:
	Say "Hello".
thats it.
```

### Loops - While
```
repeat the following while x is less than 20:
	Say x.
	Set x to be x + 1.
thats it.
```

### Loops - For Each
```
Declare myList to be [1, 2, 3, 4, 5].
for each item in myList, do the following:
	Say the value of item.
thats it.
```

### Output
```
Say "Hello, World!".
Say 42.
Say the value of x.
```

### Lists
```
Declare myList to be [1, 2, 3, 4, 5].
```

### Comparisons
- `is equal to`
- `is less than`
- `is greater than`
- `is less than or equal to`
- `is greater than or equal to`
- `is not equal to`

### Arithmetic
- `+`, `-`, `*`, `/`

### Comments
```
# This is a comment
Declare x to be 5.  # Inline comment
```

## Architecture

```
Source Code (.abc file)
    ↓
[Lexer] → Tokens
    ↓
[Parser] → AST
    ↓
[Evaluator] → Output
```

### Execution Flow

1. **Lexical Analysis**: Source code is tokenized by the lexer
2. **Syntax Analysis**: Tokens are parsed into an AST by the parser
3. **Evaluation**: The evaluator walks the AST and executes statements
4. **Scope Management**: Each function call creates a new child environment with its own variable scope

### Scope Rules

- Variables are function-scoped
- Child scopes can access parent scope variables
- Constants cannot be reassigned
- Functions are stored globally and can be called from any scope
- Function parameters shadow outer scope variables of the same name

## Building and Running

### Build
```bash
go build -o english
```

### Run a File
```bash
./english program.abc
```

### Interactive REPL
```bash
./english
```

## Implementation Details

### Multi-word Operators
The lexer has special handling for natural language comparison operators:
- When it encounters "is", it looks ahead to build compound operators
- Examples: "is equal to", "is less than", "the result of calling"

### Type System
The interpreter uses a dynamically typed system with the following types:
- **Numbers** (float64) - supports arithmetic operations
- **Strings** - text values
- **Lists** ([]interface{}) - ordered collections
- **Functions** - user-defined or built-in
- **Boolean** (implicit from toBool) - used in conditions

### Error Handling
- Parse errors include line and column information
- Runtime errors report the operation that failed
- Type conversion errors are caught and reported

### Block Terminators
- Simple statements end with a period (`.`)
- Block statements end with `thats it.`
  - Examples: functions, if statements, loops

## Notes

- **Keywords are case-insensitive** - `DECLARE`, `Declare`, and `declare` are all equivalent
- **Identifiers (variable/function names) are case-sensitive** - `myVar` and `MyVar` are different variables
- **File extension** - English language source files use the `.abc` extension
- Whitespace handling: indentation is used for readability but not enforced
- Comments can appear anywhere with `#` prefix and continue to end of line
- The interpreter uses tree-walking evaluation (direct AST execution) rather than bytecode compilation

## Future Enhancements

Possible features not yet implemented:
- Array indexing and slicing
- String interpolation
- Object/struct types
- More built-in functions (length, type checking, etc.)
- Break/continue for loops
- Try/catch error handling
- Import/module system
