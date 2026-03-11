# 🗣️ English — A Natural-Language Programming Language

> Write code the way you speak. **English** is a fully-featured, Turing-complete programming language where your source code reads like plain English.

```english
Declare function greet that takes name and does the following:
    Print "Hello, ", the value of name, "!".
thats it.

Call greet with "World".
```

---

## ✨ Features

| Feature | Description |
|---|---|
| 🔤 **Natural Syntax** | Keywords and constructs read like plain English sentences |
| 🔡 **Case-Insensitive** | `Declare`, `DECLARE`, and `declare` all work equally |
| 🎨 **Interactive REPL** | Syntax-highlighted REPL powered by Bubble Tea |
| 💬 **Helpful Errors** | Rich error messages with "did you mean…?" suggestions and full stack traces |
| 📦 **Bytecode Compiler** | Compile `.abc` source files to fast-loading `.101` bytecode |
| ⚡ **Auto-Caching** | Imported files are automatically cached (like Python's `__pycache__`) |
| 🐍 **Python Transpiler** | Convert any English program to readable Python |
| 🔒 **Strict Typing** | Variables are type-locked at declaration; no silent coercions |
| 🏗️ **Structs & Methods** | Define custom data structures with fields and possessive accessor syntax |
| ⚠️ **Custom Errors** | Declare named error types and catch them selectively |

---

## 🚀 Quick Start

### Build

```bash
git clone https://github.com/Advik-B/english
cd english
go build -o english .
```

### Run

```bash
# Start the interactive REPL
./english

# Run a source file
./english run program.abc
./english program.abc          # shorthand

# Show help
./english --help
```

---

## 📖 Language Guide

This guide walks through every feature of English from the ground up.

### Step 1 — Hello World

Every English statement ends with a period (`.`). The simplest possible program:

```english
Print "Hello, World!".
```

Run it:

```bash
./english run hello_world.abc
# Hello, World!
```

---

### Step 2 — Variables & Constants

Declare variables with `Declare … to be …`:

```english
Declare name to be "Alice".
Declare age  to be 30.
Declare pi   to always be 3.14159.   # constant — cannot be reassigned
```

Reassign with `Set … to …`:

```english
Set age to 31.
```

**Alternative `let` syntax** — all of the following are equivalent:

```english
let score be 100.
let score be equal to 100.
let score = 100.
let score equal 100.
```

**Constants with `let`:**

```english
let max_size always be 256.
let max_size be always 256.
```

> **Note:** Keywords are case-insensitive. `Declare`, `DECLARE`, and `declare` all work.

---

### Step 3 — Data Types

English has five built-in value types:

| Type | Examples | Notes |
|---|---|---|
| `number` | `42`, `3.14`, `-7` | 64-bit float |
| `text` | `"hello"`, `"line1\nline2"` | supports `\n`, `\t`, `\\`, `\"` |
| `boolean` | `true`, `false` | |
| `nothing` | `nothing` | equivalent to null/nil |
| list | `[1, 2, 3]` | ordered, mixed-type allowed |

Declare a variable with an **explicit type annotation**:

```english
Declare count    as number  to be 0.
Declare greeting as text    to be "Hi".
Declare active   as boolean to be true.
Declare pending  as number.              # no initial value — starts as nothing
```

Types are **locked at declaration**. Assigning a value of the wrong type is a `TypeError`.

---

### Step 4 — Arithmetic & Expressions

```english
Declare a to be 10.
Declare b to be 3.

Print a + b.         # 13
Print a - b.         # 7
Print a * b.         # 30
Print a / b.         # 3.3333...

# Modulo (remainder)
Print the remainder of 17 divided by 5.   # 2
Print the remainder of 10 / 3.            # 1
```

Arithmetic on strings performs **concatenation**:

```english
Declare full_name to be "Jane" + " " + "Doe".
Print full_name.    # Jane Doe
```

---

### Step 5 — Output

`Print` outputs a line followed by a newline. `Write` outputs without a trailing newline.

```english
Print "Hello, World!".          # Hello, World!\n
Write "Hello, ".                # Hello, (no newline)
Write "World!\n".               # World!\n

# Print accepts multiple arguments separated by commas
Print "Name:", the value of name.      # Name: Alice
Print "Sum:", 5 + 3.                   # Sum: 8
```

String **escape sequences**:

| Sequence | Meaning |
|---|---|
| `\n` | newline |
| `\t` | tab |
| `\\` | literal backslash |
| `\"` | literal double-quote |

---

### Step 6 — Conditionals

```english
If age is greater than 17, then
    Print "Adult".
otherwise if age is greater than 12, then
    Print "Teenager".
otherwise
    Print "Child".
thats it.
```

Every `If` block is closed with `thats it.`

**Comparison operators:**

| English | Meaning |
|---|---|
| `is equal to` | `==` |
| `is not equal to` | `!=` |
| `is less than` | `<` |
| `is greater than` | `>` |
| `is less than or equal to` | `<=` |
| `is greater than or equal to` | `>=` |

**Logical operators:** `and`, `or`, `not`

```english
If age is greater than or equal to 18 and has_license is equal to true, then
    Print "Can drive.".
thats it.

If not is_tired, then
    Print "Driver is alert.".
thats it.
```

---

### Step 7 — Loops

#### While loop

```english
Declare i to be 1.
Repeat while i is less than or equal to 5, do the following:
    Print the value of i.
    Set i to i + 1.
thats it.
```

#### Counted loop (repeat N times)

```english
Repeat 3 times, do the following:
    Print "Hello!".
thats it.
```

#### Infinite loop with `break`

```english
Declare x to be 0.
Repeat forever:
    Set x to x + 1.
    If x is equal to 5, then
        Break out of this loop.
    thats it.
thats it.
```

#### For-each loop

```english
Declare colors to be ["red", "green", "blue"].
For each color in colors, do the following:
    Print the value of color.
thats it.
```

#### `Continue` (skip to next iteration)

```english
Repeat while i is less than 10, do the following:
    Set i to i + 1.
    If the remainder of i divided by 2 is equal to 0, then
        Continue.
    thats it.
    Print the value of i.   # only odd numbers
thats it.
```

---

### Step 8 — Functions

Declare a function with `Declare function … that does the following:` and close it with `thats it.`

```english
Declare function say_hello that does the following:
    Print "Hello from a function!".
thats it.

Call say_hello.
```

**Parameters** use `that takes … and does the following:`

```english
Declare function greet that takes name and does the following:
    Print "Hello, ", the value of name, "!".
thats it.

Call greet with "Alice".
```

**Multiple parameters** are separated with `and`:

```english
Declare function add that takes a and b and does the following:
    Return a + b.
thats it.

Declare result to be the result of calling add with 5 and 3.
Print the value of result.   # 8
```

**Recursive functions** work naturally:

```english
Declare function factorial that takes n and does the following:
    If n is less than or equal to 1, then
        Return 1.
    thats it.
    Return n * the result of calling factorial with n - 1.
thats it.

Print the result of calling factorial with 6.   # 720
```

---

### Step 9 — Lists (Arrays)

Create a list with square brackets:

```english
Declare numbers to be [10, 20, 30, 40, 50].
Declare names   to be ["Alice", "Bob", "Charlie"].
```

**Access elements** by position (0-indexed):

```english
Print the item at position 0 in numbers.    # 10
Print the item at position 2 in numbers.    # 30
```

**Modify elements:**

```english
Set the item at position 2 in numbers to be 99.
```

**Iterate with for-each:**

```english
For each n in numbers, do the following:
    Print the value of n.
thats it.
```

**Useful built-in list functions:**

```english
append(numbers, 60).                    # add to end
remove(numbers, 0).                     # remove at index
Print count(numbers).                   # length
Print sum(numbers).                     # sum of elements
Print reverse(numbers).                 # reversed copy
Print sort(numbers).                    # sorted copy
Print slice(numbers, 1, 3).             # sub-list
Print first(numbers).                   # first element
Print last(numbers).                    # last element
Print unique(numbers).                  # deduplicated copy
Print average(numbers).                 # arithmetic mean
```

---

### Step 10 — Lookup Tables (Dictionaries)

A **lookup table** is an ordered key-value dictionary. Keys may be numbers, text, or booleans.

```english
Declare ages to be a lookup table.

Set ages at "Alice" to be 30.
Set ages at "Bob"   to be 25.

Print ages at "Alice".          # 30
Print count(ages).              # 2
```

**Check for key membership:**

```english
If ages has "Alice", then
    Print "Alice is in the table.".
thats it.
```

**Iterate over keys:**

```english
For each name in ages, do the following:
    Print the value of name.
thats it.
```

**Useful built-in lookup functions:**

```english
Print keys(ages).                            # list of keys
Print values(ages).                          # list of values
Set ages to be table_remove(ages, "Bob").    # remove a key
Print get_or_default(ages, "Dave", 0).       # safe access with fallback
```

---

### Step 11 — Nothing (Null)

The `nothing` literal represents the absence of a value (like `null` or `nil`).

```english
Declare result to be nothing.

If result is nothing, then
    Print "No result yet.".
thats it.

# "is something" checks that a value is NOT nothing
If result is something, then
    Print "Got a result!".
thats it.
```

**Alternative syntax:**

```english
If result has a value,  then ...   # same as "is something"
If result has no value, then ...   # same as "is nothing"
```

---

### Step 12 — Booleans & Toggle

```english
Declare is_raining to be true.
Declare is_sunny   to be false.

# Toggle flips a boolean variable in place
Toggle is_raining.                    # now false
Toggle the value of is_raining.       # now true again
```

Use `is true` / `is false` as shorthand comparisons:

```english
If is_raining is true, then
    Print "Bring an umbrella!".
thats it.
```

---

### Step 13 — String Operations

```english
Declare s to be "Hello, World!".

Print uppercase(s).                     # HELLO, WORLD!
Print lowercase(s).                     # hello, world!
Print title(s).                         # Hello, World!
Print trim("  hello  ").                # hello
Print replace(s, "World", "English").   # Hello, English!
Print contains(s, "World").             # true
Print starts_with(s, "Hello").          # true
Print ends_with(s, "!").                # true
Print substring(s, 7, 5).              # World
Print split("a,b,c", ",").             # [a, b, c]
Print join(["a", "b", "c"], "-").       # a-b-c
Print to_number("42").                  # 42
Print to_string(42).                    # 42
Print is_empty("").                     # true
Print index_of(s, "World").            # 7
Print count_occurrences(s, "l").        # 3
Print str_repeat("ab", 3).             # ababab
```

**Possessive syntax** — call methods directly on a value using `'s`:

```english
Print "hello"'s uppercase.             # HELLO
Print "hello"'s title.                 # Hello
Print 5.0's is_integer.                # true
```

---

### Step 14 — Type Casting

Convert between types with `cast to` or `casted to`:

```english
Declare age_str to be "25".
Declare age     to be age_str cast to number.
Print age + 5.              # 30

Declare score to be 98.
Declare label to be score cast to text.
Print label + " points".    # 98 points

Declare flag to be 1 cast to boolean.
Print the value of flag.    # true
```

Both `cast to` and `casted to` are accepted:

```english
Declare temp to be "98.6" casted to number.
```

---

### Step 15 — Structures (Custom Types)

Define a struct with named fields:

```english
Declare structure Person with fields name as text and age as number.
```

Create an instance with `new`:

```english
Declare alice to be a new Person with name "Alice" and age 30.
```

Access and modify fields using the **possessive `'s`** syntax:

```english
Print alice's name.          # Alice
Print alice's age.           # 30

Set alice's age to 31.
```

---

### Step 16 — Error Handling

Wrap risky code in a `Try` block. The `on error:` clause catches any error. Use `but finally:` for cleanup that always runs.

```english
Try doing the following:
    Print the result of calling risky_function with 0.
on error:
    Print "Something went wrong:", error.
but finally:
    Print "Cleanup complete.".
thats it.
```

#### Custom Error Types

Declare named error types and catch them selectively:

```english
Declare NetworkError    as an error type.
Declare ValidationError as an error type.

Try doing the following:
    Raise "Host unreachable" as NetworkError.
on NetworkError:
    Print "Network error:", error.
on ValidationError:
    Print "Validation error:", error.
on error:
    Print "Unknown error:", error.
but finally:
    Print "Done.".
thats it.
```

Check the type of an error at runtime:

```english
on error:
    If error is NetworkError, then
        Print "It was a network error.".
    thats it.
```

**Error type hierarchy:**

```english
Declare AppError     as an error type.
Declare NetworkError as a type of AppError.   # NetworkError is a subtype of AppError

Try doing the following:
    Raise "timeout" as NetworkError.
on AppError:
    Print "Caught as AppError (caught subtype too).".
thats it.
```

---

### Step 17 — Importing Files

Reuse code from other `.abc` files.

```english
# Import everything (also runs top-level code in the file)
Import "math_library.abc".

# Import with explicit "from"
Import from "utilities.abc".

# Selective import (only named items)
Import square and cube from "math_library.abc".
Import add, multiply from "helpers.abc".

# Import everything, explicitly
Import everything from "library.abc".
Import all from "library.abc".

# Safe import — loads declarations without running top-level code
Import all from "library.abc" safely.
Import square from "math.abc" safely.
```

Imported files are **automatically cached** as bytecode in `__engcache__/` (similar to Python's `__pycache__`). The cache is invalidated when the source file changes.

Example library (`math_library.abc`):

```english
Print "Library loaded.".       # runs on normal import; skipped on safe import

Declare function square that takes x and does the following:
    Return x * x.
thats it.

Declare pi to always be 3.14159.
```

---

### Step 18 — Advanced Features

#### References & Copies

By default, primitive values are **copied** on assignment. Use `a reference to` for an alias:

```english
Declare x to be 10.
Declare ref_x to be a reference to x.

Set x to 99.
Print the value of ref_x.     # 99 — ref_x mirrors x

# Explicit copy
Declare my_list  to be [1, 2, 3].
Declare my_copy  to be a copy of my_list.
```

#### Swap

```english
Declare a to be 1.
Declare b to be 2.
Swap a and b.
Print the value of a.    # 2
Print the value of b.    # 1
```

#### Memory Location

```english
Print the location of x.    # Outputs: 0x...:x
```

#### User Input

```english
Ask "What is your name?" and store in name.
Print "Hello, ", the value of name, "!".
```

Or as an expression:

```english
Declare answer to be ask("Enter a number: ").
Print answer cast to number + 1.
```

---

## 📚 Standard Library Reference

### Math

| Function | Description |
|---|---|
| `sqrt(x)` | square root |
| `pow(x, y)` | x to the power of y |
| `abs(x)` | absolute value |
| `floor(x)` | round down |
| `ceil(x)` | round up |
| `round(x)` | round to nearest integer |
| `min(a, b)` | smaller of two values |
| `max(a, b)` | larger of two values |
| `sin(x)` / `cos(x)` / `tan(x)` | trigonometry (radians) |
| `log(x)` / `log10(x)` / `log2(x)` | logarithms |
| `exp(x)` | e^x |
| `random()` | random float in [0, 1) |
| `random_between(a, b)` | random float in [a, b] |
| `is_nan(x)` | true if x is NaN |
| `is_infinite(x)` | true if x is ±infinity |
| `clamp(x, lo, hi)` | clamp x to [lo, hi] |
| `sign(x)` | -1, 0, or 1 |
| `is_integer(x)` | true if x has no fractional part |

**Built-in constants:** `pi`, `e`, `infinity`

### Strings

| Function | Description |
|---|---|
| `uppercase(s)` | UPPER CASE |
| `lowercase(s)` | lower case |
| `title(s)` | Title Case |
| `capitalize(s)` | First letter upper |
| `swapcase(s)` | sWAP cASE |
| `casefold(s)` | aggressive lowercase for comparison |
| `trim(s)` | strip leading/trailing whitespace |
| `trim_left(s)` | strip leading whitespace |
| `trim_right(s)` | strip trailing whitespace |
| `replace(s, old, new)` | replace all occurrences |
| `split(s, delim)` | split into list |
| `join(list, sep)` | join list with separator |
| `contains(s, sub)` | true if sub is in s |
| `starts_with(s, prefix)` | prefix check |
| `ends_with(s, suffix)` | suffix check |
| `substring(s, start, len)` | extract sub-string |
| `index_of(s, sub)` | first index of sub, or -1 |
| `count_occurrences(s, sub)` | count of sub in s |
| `str_repeat(s, n)` | repeat s n times |
| `pad_left(s, n)` | left-pad to width n |
| `pad_right(s, n)` | right-pad to width n |
| `center(s, n)` | center to width n |
| `zfill(s, n)` | zero-pad to width n |
| `to_number(s)` | parse string as number |
| `to_string(v)` | convert any value to string |
| `is_empty(s)` | true if s is `""` |
| `is_digit(s)` | true if all chars are digits |
| `is_alpha(s)` | true if all chars are letters |
| `is_alnum(s)` | true if all chars are alphanumeric |
| `is_space(s)` | true if all chars are whitespace |
| `is_upper(s)` | true if all chars are uppercase |
| `is_lower(s)` | true if all chars are lowercase |

### Lists

| Function | Description |
|---|---|
| `append(list, item)` | add item to end |
| `remove(list, index)` | remove element at index |
| `insert(list, index, item)` | insert at position |
| `sort(list)` | sorted copy (ascending) |
| `sorted_desc(list)` | sorted copy (descending) |
| `reverse(list)` | reversed copy |
| `slice(list, start, end)` | sub-list |
| `count(list)` | number of elements |
| `sum(list)` | sum of numeric elements |
| `average(list)` | arithmetic mean |
| `min_value(list)` | smallest element |
| `max_value(list)` | largest element |
| `product(list)` | product of all elements |
| `first(list)` | first element |
| `last(list)` | last element |
| `unique(list)` | deduplicated copy |
| `flatten(list)` | flatten nested lists |
| `any_true(list)` | true if any element is true |
| `all_true(list)` | true if all elements are true |
| `zip_with(list1, list2)` | list of `[a, b]` pairs |

### Lookup Tables

| Function | Description |
|---|---|
| `keys(table)` | list of all keys |
| `values(table)` | list of all values |
| `count(table)` | number of entries |
| `table_remove(table, key)` | remove key, return new table |
| `table_has(table, key)` | true if key exists |
| `merge(t1, t2)` | merge two tables (t2 wins on conflict) |
| `get_or_default(table, key, default)` | safe access with fallback |

---

## 🖥️ CLI Reference

```bash
# Run a source file
./english run program.abc
./english program.abc           # shorthand

# Start the interactive REPL
./english

# Compile to bytecode
./english compile program.abc           # creates program.101
./english compile program.abc -o out.101

# Run compiled bytecode
./english run program.101

# Transpile to Python
./english transpile program.abc         # creates program.abc.py
./english transpile program.101         # creates program.101.py

# Show version
./english version

# Help
./english --help
./english run --help
```

---

## 🎨 Interactive REPL

Start the REPL with no arguments:

```bash
./english
```

Features:
- **Syntax highlighting** — keywords, strings, numbers, and operators are color-coded
- **Multi-line blocks** — automatically detects incomplete blocks until you type `thats it.`
- **Command history** — use arrow keys to navigate previous inputs

REPL commands:

| Command | Description |
|---|---|
| `:help` / `:h` | show help |
| `:clear` / `:cls` | clear screen |
| `:exit` / `:quit` / `:q` | exit the REPL |
| `Ctrl+C` / `Esc` | exit the REPL |

---

## 📦 Bytecode Compilation

English can compile source files to a binary `.101` format for faster loading (no parsing required at runtime).

```bash
# Compile
./english compile myprogram.abc          # creates myprogram.101

# Run bytecode directly
./english run myprogram.101
```

The `.101` format:
- Uses magic bytes `0x10 0x1E 0x4E 0x47` for identification
- Includes a version byte for format compatibility
- Stores a binary-encoded AST (protobuf-style serialization)

Imported files are **automatically compiled and cached** in `__engcache__/`. The cache is invalidated by a SipHash (PEP 552-style) of the source file content.

---

## 🐍 Python Transpiler

Convert any English program to readable Python:

```bash
./english transpile myprogram.abc        # creates myprogram.abc.py
./english transpile myprogram.101        # works on bytecode too
```

Quick translation reference:

| English | Python |
|---|---|
| `Declare x to be 5.` | `x = 5` |
| `Declare pi to always be 3.14.` | `pi = 3.14  # constant` |
| `Print "hello".` | `print("hello")` |
| `Write "hello".` | `print("hello", end="")` |
| `If x is greater than 5, then …` | `if x > 5:` |
| `Repeat while x is less than 10, …` | `while x < 10:` |
| `Repeat 5 times, …` | `for _ in range(5):` |
| `For each item in list, …` | `for item in list:` |
| `Repeat forever:` | `while True:` |
| `Declare function foo that takes a …` | `def foo(a):` |
| `Return x.` | `return x` |
| `Try doing the following: … on error: …` | `try: … except Exception: …` |
| `Raise "msg" as NetworkError.` | `raise NetworkError("msg")` |
| `Declare NetworkError as an error type.` | `class NetworkError(Exception): pass` |
| `Declare ages to be a lookup table.` | `ages = {}` |
| `Toggle flag.` | `flag = not flag` |
| `Swap x and y.` | `x, y = y, x` |

Standard library calls are mapped to their Python equivalents (e.g. `sqrt(x)` → `math.sqrt(x)`). A small set of helper functions is injected at the top of the generated file for operations without a direct Python equivalent.

---

## 💡 Example Programs

### Hello World

```english
Print "Hello, World!".
```

### FizzBuzz

```english
Declare i to be 1.
Repeat while i is less than or equal to 100, do the following:
    If the remainder of i divided by 15 is equal to 0, then
        Print "FizzBuzz".
    otherwise if the remainder of i divided by 3 is equal to 0, then
        Print "Fizz".
    otherwise if the remainder of i divided by 5 is equal to 0, then
        Print "Buzz".
    otherwise
        Print the value of i.
    thats it.
    Set i to i + 1.
thats it.
```

### Fibonacci Sequence

```english
Declare function fib that takes n and does the following:
    If n is less than or equal to 1, then
        Return n.
    thats it.
    Return the result of calling fib with n - 1
         + the result of calling fib with n - 2.
thats it.

Declare i to be 0.
Repeat while i is less than 10, do the following:
    Print the result of calling fib with i.
    Set i to i + 1.
thats it.
```

### Bubble Sort

```english
Declare arr to be [64, 34, 25, 12, 22, 11, 90].
Declare n to be 7.
Declare i to be 0.

Repeat while i is less than n - 1:
    Declare j to be 0.
    Repeat while j is less than n - i - 1:
        Declare a to be the item at position j in arr.
        Declare b to be the item at position j + 1 in arr.
        If a is greater than b, then
            Set the item at position j in arr to be b.
            Set the item at position j + 1 in arr to be a.
        thats it.
        Set j to j + 1.
    thats it.
    Set i to i + 1.
thats it.

Print the value of arr.
```

### Custom Structs

```english
Declare structure Point with fields x as number and y as number.

Declare function distance that takes p and does the following:
    Return sqrt(p's x * p's x + p's y * p's y).
thats it.

Declare origin to be a new Point with x 3 and y 4.
Print the result of calling distance with origin.   # 5
```

### Error Handling

```english
Declare NetworkError as an error type.

Declare function fetch that takes url and does the following:
    If url is equal to "", then
        Raise "URL must not be empty" as NetworkError.
    thats it.
    Print "Fetching:", the value of url.
thats it.

Try doing the following:
    Call fetch with "".
on NetworkError:
    Print "Network error:", error.
but finally:
    Print "Done.".
thats it.
```

---

## 📁 Project Structure

```
english/
├── main.go                  # Entry point
├── cmd/
│   ├── root.go              # Cobra CLI & subcommands
│   └── repl.go              # Bubble Tea REPL
├── token/
│   └── token.go             # Token type definitions
├── tokeniser/
│   └── tokeniser.go         # Shared lexer
├── ast/
│   └── ast.go               # AST node types (50+ nodes)
├── parser/
│   ├── lexer.go             # Tokenizer wrapper
│   ├── parser.go            # Recursive-descent parser
│   ├── messages.go          # Error message strings
│   └── syntax_error.go      # SyntaxError type
├── astvm/
│   └── vm.go                # AST tree-walk interpreter
├── vm/
│   ├── evaluator.go         # Statement evaluator
│   ├── environment.go       # Scoped variable store
│   ├── checker.go           # Compile-time type checker
│   ├── values.go            # Value types & errors
│   ├── stdlib/
│   │   └── stdlib.go        # Standard library functions
│   └── types/
│       └── kind.go          # TypeKind enum
├── bytecode/
│   └── bytecode.go          # AST ↔ .101 serialization
├── transpiler/
│   └── transpiler.go        # AST → Python
├── highlight/
│   └── highlight.go         # Syntax highlighting
├── stacktraces/
│   └── stacktraces.go       # Error rendering
└── examples/                # 60+ example programs
    ├── hello_world.abc
    ├── fibonacci.abc
    ├── fizzbuzz.abc
    ├── factorial.abc
    ├── bubble_sort.abc
    ├── error_types.abc
    ├── strict_types.abc
    ├── lookup_table_demo.abc
    ├── turing_machine.abc   # Proves Turing completeness
    └── ...
```

---

## 🛠️ Development

### Build

```bash
go build -o english .
```

### Test

```bash
# All tests
go test ./...

# Verbose output
go test ./... -v

# Single package
go test ./parser/... -v
go test ./vm/...    -v
```

### Run Examples

```bash
./english run examples/hello_world.abc
./english run examples/fibonacci.abc
./english run examples/turing_machine.abc
```

### Compile & Run Bytecode

```bash
./english compile examples/fibonacci.abc
./english run     examples/fibonacci.101
```

---

## 📦 Dependencies

| Package | Version | Purpose |
|---|---|---|
| [cobra](https://github.com/spf13/cobra) | v1.10.2 | CLI framework |
| [bubbletea](https://github.com/charmbracelet/bubbletea) | v1.3.10 | REPL TUI |
| [lipgloss](https://github.com/charmbracelet/lipgloss) | v1.1.0 | Terminal styling |
| [siphash](https://github.com/dchest/siphash) | v1.2.3 | Bytecode cache hashing |
| [go-isatty](https://github.com/mattn/go-isatty) | v0.0.20 | TTY detection |

```bash
go mod download
```

---

## 📄 License

MIT License

## 🙏 Acknowledgments

Built with the amazing [Charm](https://charm.sh/) terminal libraries.
