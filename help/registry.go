package help

// loadDefaultEntries populates the registry with comprehensive help content
// covering all features of the English programming language.
func (r *Registry) loadDefaultEntries() {
	// ═══════════════════════════════════════════════════════════════════════════
	// REPL COMMANDS
	// ═══════════════════════════════════════════════════════════════════════════

	r.Register(&HelpEntry{
		Name:        "help",
		Description: "Display help information with fuzzy search",
		Category:    "command",
		LongDesc:    "The help command displays help information. Use 'help <topic>' to search for specific topics using fuzzy matching. Supports searching by name, keywords, aliases, and approximate spelling.",
		Examples: []string{
			"help",
			"help print",
			"help loop",
			"help variable",
		},
		Keywords: []string{"help", "documentation", "info", "assist", "search"},
		Aliases:  []string{"?"},
	})

	r.Register(&HelpEntry{
		Name:        "exit",
		Description: "Exit the REPL",
		Category:    "command",
		LongDesc:    "Exits the REPL and returns to the shell. You can also use 'quit' or end with a period.",
		Examples:    []string{"exit", "quit", "exit.", "quit."},
		Keywords:    []string{"quit", "leave", "close", "terminate"},
		Aliases:     []string{"quit"},
	})

	// ═══════════════════════════════════════════════════════════════════════════
	// VARIABLE DECLARATION & ASSIGNMENT
	// ═══════════════════════════════════════════════════════════════════════════

	r.Register(&HelpEntry{
		Name:        "declare",
		Description: "Declare a variable with optional type annotation",
		Category:    "keyword",
		LongDesc:    "Use 'Declare' to create a new variable. Variables can be typed or untyped. Use 'as <type>' for type annotations. Use 'always' to create constants.",
		Examples: []string{
			"Declare x to be 5.",
			"Declare name to be \"Alice\".",
			"Declare x as number to be 10.",
			"Declare items to be [1, 2, 3].",
			"Declare pi to always be 3.14159.",
			"Declare count as number.",
		},
		Keywords: []string{"variable", "assign", "set", "define", "create", "var", "let"},
		Aliases:  []string{"let"},
		SeeAlso:  []string{"set", "types", "always"},
	})

	r.Register(&HelpEntry{
		Name:        "set",
		Description: "Change the value of an existing variable",
		Category:    "keyword",
		LongDesc:    "Use 'Set' to modify the value of a previously declared variable. Cannot be used on constants.",
		Examples: []string{
			"Set x to 10.",
			"Set name to \"Bob\".",
			"Set items to [4, 5, 6].",
			"Set the item at position 0 in list to 99.",
			"Set the field name of person to \"Alice\".",
		},
		Keywords: []string{"assign", "change", "update", "modify", "reassign"},
		SeeAlso:  []string{"declare"},
	})

	r.Register(&HelpEntry{
		Name:        "always",
		Description: "Declare a constant (immutable variable)",
		Category:    "keyword",
		LongDesc:    "Use 'always' within a declare statement to create a constant that cannot be modified after initialization.",
		Examples: []string{
			"Declare pi to always be 3.14159.",
			"Declare max_attempts to always be 10.",
		},
		Keywords: []string{"constant", "immutable", "final", "readonly"},
		SeeAlso:  []string{"declare"},
	})

	// ═══════════════════════════════════════════════════════════════════════════
	// CONTROL FLOW - CONDITIONALS
	// ═══════════════════════════════════════════════════════════════════════════

	r.Register(&HelpEntry{
		Name:        "if",
		Description: "Conditional statement execution",
		Category:    "keyword",
		LongDesc:    "Use 'If' statements to execute code conditionally based on a boolean expression. Supports 'otherwise' (else) and 'otherwise if' (else if) clauses. End with 'thats it.'",
		Examples: []string{
			"If x is 5, then print \"x is five\". thats it.",
			"If x > 10, then do the following:\n    Print \"large\".\nthats it.",
			"If x < 0, then print \"negative\". otherwise if x is 0, then print \"zero\". otherwise print \"positive\". thats it.",
		},
		Keywords: []string{"conditional", "then", "condition", "branch"},
		SeeAlso:  []string{"otherwise", "comparison", "boolean"},
	})

	r.Register(&HelpEntry{
		Name:        "otherwise",
		Description: "Else clause in conditional statements",
		Category:    "keyword",
		LongDesc:    "Use 'otherwise' as the else clause in if statements. Can be combined with 'if' for else-if chains.",
		Examples: []string{
			"If x > 0, then print \"positive\". otherwise print \"not positive\". thats it.",
			"If x is 1, then print \"one\". otherwise if x is 2, then print \"two\". otherwise print \"other\". thats it.",
		},
		Keywords: []string{"else", "elif", "elseif", "alternative"},
		Aliases:  []string{"else"},
		SeeAlso:  []string{"if"},
	})

	r.Register(&HelpEntry{
		Name:        "then",
		Description: "Introduces the consequence of a conditional",
		Category:    "keyword",
		LongDesc:    "Used after 'if' or 'otherwise if' to introduce the code block that executes when the condition is true.",
		Examples: []string{
			"If x > 5, then print \"greater\".",
		},
		Keywords: []string{"consequence", "branch"},
		SeeAlso:  []string{"if"},
	})

	// ═══════════════════════════════════════════════════════════════════════════
	// CONTROL FLOW - LOOPS
	// ═══════════════════════════════════════════════════════════════════════════

	r.Register(&HelpEntry{
		Name:        "for each",
		Description: "Iterate over elements in a collection",
		Category:    "keyword",
		LongDesc:    "Use 'For each' to iterate over items in a list or range. Each iteration provides one element from the collection.",
		Examples: []string{
			"For each n in [1, 2, 3], do the following:\n    Print the value of n.\nthats it.",
			"For each item in items, print the value of item.",
			"For each n in [1 .. 10], print the value of n.",
			"For each name in names, do the following:\n    Print name.\nthats it.",
		},
		Keywords: []string{"loop", "iterate", "collection", "list", "array", "each"},
		Aliases:  []string{"foreach", "for"},
		SeeAlso:  []string{"repeat", "while", "range"},
	})

	r.Register(&HelpEntry{
		Name:        "repeat while",
		Description: "Loop while a condition is true",
		Category:    "keyword",
		LongDesc:    "Use 'repeat the following while' or 'repeat while' to create a while loop that continues as long as the condition evaluates to true.",
		Examples: []string{
			"Declare i to be 0.\nRepeat the following while i < 10:\n    Print the value of i.\n    Set i to i + 1.\nthats it.",
			"Repeat while x > 0:\n    Set x to x - 1.\nthats it.",
		},
		Keywords: []string{"while", "loop", "condition", "iteration"},
		Aliases:  []string{"while"},
		SeeAlso:  []string{"for each", "repeat"},
	})

	r.Register(&HelpEntry{
		Name:        "repeat",
		Description: "Repeat a block a fixed number of times or forever",
		Category:    "keyword",
		LongDesc:    "Use 'repeat' to execute a block multiple times. Can specify a number with 'times', or use 'forever' for infinite loops.",
		Examples: []string{
			"Repeat 5 times:\n    Print \"Hello\".\nthats it.",
			"Repeat forever:\n    Print \"Running\".\nthats it.",
		},
		Keywords: []string{"loop", "times", "iterate", "forever", "infinite"},
		SeeAlso:  []string{"for each", "repeat while", "break"},
	})

	r.Register(&HelpEntry{
		Name:        "break",
		Description: "Exit a loop early",
		Category:    "keyword",
		LongDesc:    "Use 'break' or 'break out of this loop' to immediately exit the innermost loop.",
		Examples: []string{
			"For each n in [1 .. 100], do the following:\n    If n is 50, then break out of this loop.\n    Print the value of n.\nthats it.",
			"Repeat forever:\n    If done, then break.\nthats it.",
		},
		Keywords: []string{"exit", "terminate", "stop", "loop control"},
		SeeAlso:  []string{"continue", "repeat"},
	})

	r.Register(&HelpEntry{
		Name:        "continue",
		Description: "Skip to the next iteration of a loop",
		Category:    "keyword",
		LongDesc:    "Use 'continue' or 'skip' to skip the rest of the current iteration and move to the next one.",
		Examples: []string{
			"For each n in [1 .. 10], do the following:\n    If n is 5, then continue.\n    Print the value of n.\nthats it.",
		},
		Keywords: []string{"skip", "next", "loop control"},
		Aliases:  []string{"skip"},
		SeeAlso:  []string{"break", "repeat"},
	})

	// ═══════════════════════════════════════════════════════════════════════════
	// FUNCTIONS
	// ═══════════════════════════════════════════════════════════════════════════

	r.Register(&HelpEntry{
		Name:        "function",
		Description: "Define a reusable function",
		Category:    "keyword",
		LongDesc:    "Declare functions using 'Declare function <name> that takes <params> and does the following:'. Functions can return values using 'return'.",
		Examples: []string{
			"Declare function greet that takes name and does the following:\n    Print \"Hello, \" + name + \"!\".\nthats it.",
			"Declare function add that takes a and b and does the following:\n    Return a + b.\nthats it.",
			"Declare function say_hello that does the following:\n    Print \"Hello!\".\nthats it.",
		},
		Keywords: []string{"procedure", "subroutine", "method", "def", "define"},
		SeeAlso:  []string{"return", "call"},
	})

	r.Register(&HelpEntry{
		Name:        "return",
		Description: "Return a value from a function",
		Category:    "keyword",
		LongDesc:    "Use 'return' to exit a function and optionally provide a return value.",
		Examples: []string{
			"Return x + y.",
			"Return true.",
			"Return.",
		},
		Keywords: []string{"exit", "yield", "output"},
		SeeAlso:  []string{"function"},
	})

	r.Register(&HelpEntry{
		Name:        "call",
		Description: "Invoke a function",
		Category:    "keyword",
		LongDesc:    "Use 'call' to invoke a function. Use 'with' to pass arguments. Use 'the result of calling' to capture the return value.",
		Examples: []string{
			"Call greet with \"Alice\".",
			"Declare result to be the result of calling add with 5 and 3.",
			"Call process.",
		},
		Keywords: []string{"invoke", "execute", "run"},
		SeeAlso:  []string{"function", "with"},
	})

	// ═══════════════════════════════════════════════════════════════════════════
	// I/O OPERATIONS
	// ═══════════════════════════════════════════════════════════════════════════

	r.Register(&HelpEntry{
		Name:        "print",
		Description: "Output text to console with newline",
		Category:    "function",
		LongDesc:    "Print outputs values to the console followed by a newline. Use 'the value of' to print variable values. Multiple arguments are concatenated.",
		Examples: []string{
			"Print \"Hello, World!\".",
			"Print the value of x.",
			"Print the result of 2 + 2.",
			"Print \"x is\", the value of x.",
		},
		Keywords: []string{"output", "display", "show", "console", "write"},
		SeeAlso:  []string{"write", "ask"},
	})

	r.Register(&HelpEntry{
		Name:        "write",
		Description: "Output text to console without newline",
		Category:    "function",
		LongDesc:    "Write outputs values to the console without adding a newline at the end. Useful for building output on one line.",
		Examples: []string{
			"Write \"Enter name: \".",
		},
		Keywords: []string{"output", "display", "show", "console"},
		SeeAlso:  []string{"print"},
	})

	r.Register(&HelpEntry{
		Name:        "ask",
		Description: "Get input from the user",
		Category:    "function",
		LongDesc:    "Use 'ask' to prompt the user for input and store it in a variable. The input is always returned as text.",
		Examples: []string{
			"Ask the user for a name and declare it as name.",
			"Declare age to be ask(\"Enter your age: \").",
		},
		Keywords: []string{"input", "prompt", "user input", "read", "stdin"},
		Aliases:  []string{"input", "read"},
		SeeAlso:  []string{"print"},
	})

	// ═══════════════════════════════════════════════════════════════════════════
	// DATA TYPES
	// ═══════════════════════════════════════════════════════════════════════════

	r.Register(&HelpEntry{
		Name:        "number",
		Description: "Numeric data type (64-bit float)",
		Category:    "type",
		LongDesc:    "Numbers can be integers or floating-point values. All numbers are stored as 64-bit floats. Supports arithmetic operations.",
		Examples: []string{
			"Declare x as number to be 5.",
			"Declare pi as number to be 3.14159.",
			"Print the result of 10 + 20.",
		},
		Keywords: []string{"integer", "float", "numeric", "math", "int", "double"},
		SeeAlso:  []string{"text", "boolean", "arithmetic"},
	})

	r.Register(&HelpEntry{
		Name:        "text",
		Description: "String/text data type",
		Category:    "type",
		LongDesc:    "Text values are strings enclosed in double quotes. Supports concatenation with + and various string operations. Escape sequences like \\n for newline are supported.",
		Examples: []string{
			"Declare name as text to be \"Alice\".",
			"Print \"Hello, \" + name + \"!\".",
			"Declare message to be \"Line 1\\nLine 2\".",
		},
		Keywords: []string{"string", "character", "word", "str"},
		Aliases:  []string{"string"},
		SeeAlso:  []string{"number", "boolean", "string functions"},
	})

	r.Register(&HelpEntry{
		Name:        "boolean",
		Description: "True/false data type",
		Category:    "type",
		LongDesc:    "Boolean values are either true or false. Used in conditions and logical operations. Supports 'and', 'or', 'not' operators.",
		Examples: []string{
			"Declare is_active as boolean to be true.",
			"If is_active, then print \"Active\". thats it.",
			"Declare flag to be true and false.",
		},
		Keywords: []string{"true", "false", "logical", "bool", "binary"},
		SeeAlso:  []string{"comparison", "if", "logical operators"},
	})

	r.Register(&HelpEntry{
		Name:        "nothing",
		Description: "Null/nil value",
		Category:    "type",
		LongDesc:    "Represents the absence of a value. Can be tested with 'is nothing' or 'is something'.",
		Examples: []string{
			"Declare x to be nothing.",
			"If x is nothing, then print \"x has no value\". thats it.",
			"If result is something, then print result. thats it.",
		},
		Keywords: []string{"null", "nil", "none", "void", "empty"},
		Aliases:  []string{"null", "none"},
		SeeAlso:  []string{"is nothing", "is something"},
	})

	r.Register(&HelpEntry{
		Name:        "list",
		Description: "Ordered collection of values",
		Category:    "type",
		LongDesc:    "Lists are ordered collections that can contain any type of values. Access elements by zero-based index using brackets. Can be declared with type constraints.",
		Examples: []string{
			"Declare items to be [1, 2, 3, 4, 5].",
			"Print the value of items[0].",
			"For each item in items, print the value of item.",
			"Declare names to be an array of text.",
		},
		Keywords: []string{"array", "collection", "sequence", "vector"},
		Aliases:  []string{"array"},
		SeeAlso:  []string{"for each", "range", "list functions"},
	})

	r.Register(&HelpEntry{
		Name:        "lookup table",
		Description: "Key-value dictionary/map",
		Category:    "type",
		LongDesc:    "Lookup tables store key-value pairs. Access values using 'table at key' syntax. Keys are typically text.",
		Examples: []string{
			"Declare scores to be a lookup table.",
			"Set scores at \"Alice\" to 95.",
			"Print scores at \"Alice\".",
			"If scores has \"Bob\", then print \"Found\". thats it.",
		},
		Keywords: []string{"dictionary", "map", "hash", "object", "dict", "hashmap"},
		Aliases:  []string{"dictionary", "map"},
		SeeAlso:  []string{"lookup table functions"},
	})

	// ═══════════════════════════════════════════════════════════════════════════
	// OPERATORS
	// ═══════════════════════════════════════════════════════════════════════════

	r.Register(&HelpEntry{
		Name:        "arithmetic",
		Description: "Mathematical operators",
		Category:    "operator",
		LongDesc:    "Arithmetic operators: + (addition), - (subtraction), * (multiplication), / (division). For modulo, use 'remainder of X divided by Y'.",
		Examples: []string{
			"Print the result of 5 + 3.",
			"Print the result of 10 - 4.",
			"Print the result of 6 * 7.",
			"Print the result of 15 / 3.",
			"Print the remainder of 10 divided by 3.",
		},
		Keywords: []string{"math", "addition", "subtraction", "multiplication", "division", "modulo", "plus", "minus"},
		SeeAlso:  []string{"number", "math functions"},
	})

	r.Register(&HelpEntry{
		Name:        "comparison",
		Description: "Compare values for equality and ordering",
		Category:    "operator",
		LongDesc:    "Comparison operators: 'is' or 'is equal to' (equality), 'is not' (inequality), 'is greater than' (>), 'is less than' (<), 'is greater or equal', 'is less or equal'.",
		Examples: []string{
			"If x is 5, then print \"equal\". thats it.",
			"If x is not 0, then print \"not zero\". thats it.",
			"If x is greater than 10, then print \"greater\". thats it.",
			"If x is less than 5, then print \"less\". thats it.",
		},
		Keywords: []string{"equal", "greater", "less", "compare", "relational"},
		SeeAlso:  []string{"if", "boolean"},
	})

	r.Register(&HelpEntry{
		Name:        "logical operators",
		Description: "Boolean logic operators",
		Category:    "operator",
		LongDesc:    "Logical operators: 'and' (both must be true), 'or' (at least one must be true), 'not' (negation).",
		Examples: []string{
			"If x > 0 and x < 10, then print \"in range\". thats it.",
			"If x is 5 or x is 10, then print \"match\". thats it.",
			"If not flag, then print \"flag is false\". thats it.",
		},
		Keywords: []string{"and", "or", "not", "boolean", "logic"},
		SeeAlso:  []string{"boolean", "if"},
	})

	r.Register(&HelpEntry{
		Name:        "possessive",
		Description: "Method call using possessive syntax",
		Category:    "operator",
		LongDesc:    "Use the possessive 's to call methods or access properties on values. Works with strings, numbers, and objects.",
		Examples: []string{
			"Print \"hello\"'s uppercase.",
			"Print 5.0's is_integer.",
			"Print person's name.",
		},
		Keywords: []string{"method", "property", "accessor", "dot"},
		SeeAlso:  []string{"call"},
	})

	// ═══════════════════════════════════════════════════════════════════════════
	// SPECIAL SYNTAX
	// ═══════════════════════════════════════════════════════════════════════════

	r.Register(&HelpEntry{
		Name:        "range",
		Description: "Create a sequence of consecutive numbers",
		Category:    "concept",
		LongDesc:    "Ranges create sequences of consecutive integers. Supports both ascending and descending ranges. Two syntaxes: bracket notation [start .. end] or English 'a range from start to end'.",
		Examples: []string{
			"Declare nums to be [1 .. 10].",
			"Declare nums to be a range from 1 to 30.",
			"For each n in [1 .. 5], print the value of n.",
			"Declare countdown to be [10 .. 1].",
		},
		Keywords: []string{"sequence", "numbers", "from", "to", "series"},
		SeeAlso:  []string{"list", "for each"},
	})

	r.Register(&HelpEntry{
		Name:        "do the following",
		Description: "Start a multi-line code block",
		Category:    "keyword",
		LongDesc:    "Use 'do the following:' or just 'do:' to start a multi-line block. Must be closed with 'thats it.'",
		Examples: []string{
			"If x > 0, then do the following:\n    Print \"positive\".\n    Print \"number\".\nthats it.",
			"Repeat 3 times, do:\n    Print \"Hello\".\nthats it.",
		},
		Keywords: []string{"block", "multiline", "group", "begin"},
		SeeAlso:  []string{"thats it"},
	})

	r.Register(&HelpEntry{
		Name:        "thats it",
		Description: "End a multi-line code block",
		Category:    "keyword",
		LongDesc:    "Use 'thats it.' (with a period) to close a block started with 'do the following:', function definitions, loops, or conditionals.",
		Examples: []string{
			"If x > 0, then do the following:\n    Print \"positive\".\nthats it.",
		},
		Keywords: []string{"end", "close", "finish", "done"},
		Aliases:  []string{"that's it"},
		SeeAlso:  []string{"do the following"},
	})

	r.Register(&HelpEntry{
		Name:        "politeness",
		Description: "Optional polite prefixes for statements",
		Category:    "concept",
		LongDesc:    "You can optionally prefix statements with 'Please', 'Kindly', 'Could you', or 'Would you kindly' for politeness. Use --polite or --minimum-politeness flags when running files to enforce politeness.",
		Examples: []string{
			"Please print \"Hello\".",
			"Kindly declare x to be 5.",
			"Could you print the value of x.",
			"Would you kindly set x to 10.",
		},
		Keywords: []string{"please", "kindly", "polite", "courteous", "courtesy"},
		SeeAlso:  []string{"run command"},
	})

	r.Register(&HelpEntry{
		Name:        "comments",
		Description: "Add explanatory notes to code",
		Category:    "concept",
		LongDesc:    "Use '#' for single-line comments. Everything after # on a line is ignored by the interpreter.",
		Examples: []string{
			"# This is a comment",
			"Print \"Hello\". # This prints a greeting",
			"# TODO: Add error handling here",
		},
		Keywords: []string{"note", "documentation", "remark", "annotation"},
	})

	// ═══════════════════════════════════════════════════════════════════════════
	// ERROR HANDLING
	// ═══════════════════════════════════════════════════════════════════════════

	r.Register(&HelpEntry{
		Name:        "try catch",
		Description: "Handle errors gracefully",
		Category:    "keyword",
		LongDesc:    "Use try/catch blocks to handle errors. Supports 'on error' to catch all errors, 'on <ErrorType>' for specific errors, and 'finally' for cleanup code that always runs.",
		Examples: []string{
			"Try the following:\n    Print the result of 10 / 0.\non error:\n    Print \"Cannot divide by zero\".\nthats it.",
			"Try the following:\n    Print \"risky\".\nfinally:\n    Print \"cleanup\".\nthats it.",
			"Try the following:\n    Raise \"Bad\" as NetworkError.\non NetworkError:\n    Print \"Network problem\".\nthats it.",
		},
		Keywords: []string{"error", "exception", "catch", "finally", "except"},
		Aliases:  []string{"try", "catch"},
		SeeAlso:  []string{"raise", "error types"},
	})

	r.Register(&HelpEntry{
		Name:        "raise",
		Description: "Throw an error",
		Category:    "keyword",
		LongDesc:    "Use 'raise' to throw an error. Optionally specify an error type with 'as <ErrorType>'.",
		Examples: []string{
			"Raise \"Something went wrong\".",
			"Raise \"Connection failed\" as NetworkError.",
		},
		Keywords: []string{"throw", "error", "exception"},
		SeeAlso:  []string{"try catch", "error types"},
	})

	r.Register(&HelpEntry{
		Name:        "error types",
		Description: "Define custom error hierarchies",
		Category:    "concept",
		LongDesc:    "Declare custom error types for structured error handling. Error types can inherit from other error types using 'as a type of'.",
		Examples: []string{
			"Declare NetworkError as an error type.",
			"Declare TimeoutError as a type of NetworkError.",
			"Try the following:\n    Raise \"timeout\" as TimeoutError.\non NetworkError:\n    Print \"Network issue\".\nthats it.",
		},
		Keywords: []string{"exception", "hierarchy", "custom"},
		SeeAlso:  []string{"try catch", "raise"},
	})

	// ═══════════════════════════════════════════════════════════════════════════
	// STRUCTURES
	// ═══════════════════════════════════════════════════════════════════════════

	r.Register(&HelpEntry{
		Name:        "struct",
		Description: "Define custom data structures",
		Category:    "concept",
		LongDesc:    "Structs allow you to create custom data types with named fields. Use 'Define a <name> with <fields>' syntax. Fields can have types and default values.",
		Examples: []string{
			"Define a person with a name and an age.",
			"Declare john as a person with name \"John\" and age 30.",
			"Print the name of john.",
			"Set the age of john to 31.",
		},
		Keywords: []string{"structure", "object", "type", "custom type", "class", "record"},
		SeeAlso:  []string{"types"},
	})

	// ═══════════════════════════════════════════════════════════════════════════
	// IMPORTS
	// ═══════════════════════════════════════════════════════════════════════════

	r.Register(&HelpEntry{
		Name:        "import",
		Description: "Import code from other files",
		Category:    "keyword",
		LongDesc:    "Use 'import' to include code from other .abc files. Can import everything or specific functions. Use 'safely' to prevent side effects.",
		Examples: []string{
			"Import \"utils.abc\".",
			"Import add and subtract from \"math.abc\".",
			"Import everything from \"lib.abc\".",
			"Import all from \"lib.abc\" safely.",
		},
		Keywords: []string{"include", "require", "module", "load"},
		SeeAlso:  []string{"function"},
	})

	// ═══════════════════════════════════════════════════════════════════════════
	// MATH FUNCTIONS
	// ═══════════════════════════════════════════════════════════════════════════

	r.Register(&HelpEntry{
		Name:        "sqrt",
		Description: "Calculate square root",
		Category:    "function",
		LongDesc:    "Returns the square root of a number.",
		Examples: []string{
			"Print sqrt(16).",
			"Declare root to be sqrt(25).",
		},
		Keywords: []string{"math", "root", "square"},
		SeeAlso:  []string{"pow", "math functions"},
	})

	r.Register(&HelpEntry{
		Name:        "pow",
		Description: "Calculate power (exponentiation)",
		Category:    "function",
		LongDesc:    "Returns base raised to the exponent power.",
		Examples: []string{
			"Print pow(2, 8).",
			"Declare result to be pow(3, 4).",
		},
		Keywords: []string{"math", "power", "exponent", "exponential"},
		SeeAlso:  []string{"sqrt"},
	})

	r.Register(&HelpEntry{
		Name:        "abs",
		Description: "Calculate absolute value",
		Category:    "function",
		LongDesc:    "Returns the absolute (positive) value of a number.",
		Examples: []string{
			"Print abs(-5).",
			"Print abs(3).",
		},
		Keywords: []string{"math", "absolute", "magnitude"},
		SeeAlso:  []string{"sign"},
	})

	r.Register(&HelpEntry{
		Name:        "floor",
		Description: "Round down to nearest integer",
		Category:    "function",
		LongDesc:    "Returns the largest integer less than or equal to the number.",
		Examples: []string{
			"Print floor(4.9).",
			"Print floor(-2.3).",
		},
		Keywords: []string{"math", "round", "integer"},
		SeeAlso:  []string{"ceil", "round"},
	})

	r.Register(&HelpEntry{
		Name:        "ceil",
		Description: "Round up to nearest integer",
		Category:    "function",
		LongDesc:    "Returns the smallest integer greater than or equal to the number.",
		Examples: []string{
			"Print ceil(4.1).",
			"Print ceil(-2.7).",
		},
		Keywords: []string{"math", "round", "integer", "ceiling"},
		SeeAlso:  []string{"floor", "round"},
	})

	r.Register(&HelpEntry{
		Name:        "round",
		Description: "Round to nearest integer",
		Category:    "function",
		LongDesc:    "Returns the nearest integer, rounding half values up.",
		Examples: []string{
			"Print round(4.5).",
			"Print round(4.4).",
		},
		Keywords: []string{"math", "rounding", "integer"},
		SeeAlso:  []string{"floor", "ceil"},
	})

	r.Register(&HelpEntry{
		Name:        "min",
		Description: "Find minimum of two numbers",
		Category:    "function",
		LongDesc:    "Returns the smaller of two numbers.",
		Examples: []string{
			"Print min(5, 3).",
			"Declare smallest to be min(x, y).",
		},
		Keywords: []string{"math", "minimum", "smaller", "least"},
		SeeAlso:  []string{"max", "min_value"},
	})

	r.Register(&HelpEntry{
		Name:        "max",
		Description: "Find maximum of two numbers",
		Category:    "function",
		LongDesc:    "Returns the larger of two numbers.",
		Examples: []string{
			"Print max(5, 3).",
			"Declare largest to be max(x, y).",
		},
		Keywords: []string{"math", "maximum", "larger", "greatest"},
		SeeAlso:  []string{"min", "max_value"},
	})

	r.Register(&HelpEntry{
		Name:        "sin",
		Description: "Calculate sine (trigonometry)",
		Category:    "function",
		LongDesc:    "Returns the sine of an angle in radians.",
		Examples: []string{
			"Print sin(0).",
			"Print sin(pi / 2).",
		},
		Keywords: []string{"math", "trigonometry", "sine"},
		SeeAlso:  []string{"cos", "tan", "pi"},
	})

	r.Register(&HelpEntry{
		Name:        "cos",
		Description: "Calculate cosine (trigonometry)",
		Category:    "function",
		LongDesc:    "Returns the cosine of an angle in radians.",
		Examples: []string{
			"Print cos(0).",
			"Print cos(pi).",
		},
		Keywords: []string{"math", "trigonometry", "cosine"},
		SeeAlso:  []string{"sin", "tan", "pi"},
	})

	r.Register(&HelpEntry{
		Name:        "tan",
		Description: "Calculate tangent (trigonometry)",
		Category:    "function",
		LongDesc:    "Returns the tangent of an angle in radians.",
		Examples: []string{
			"Print tan(0).",
			"Print tan(pi / 4).",
		},
		Keywords: []string{"math", "trigonometry", "tangent"},
		SeeAlso:  []string{"sin", "cos", "pi"},
	})

	r.Register(&HelpEntry{
		Name:        "log",
		Description: "Calculate natural logarithm",
		Category:    "function",
		LongDesc:    "Returns the natural logarithm (base e) of a number.",
		Examples: []string{
			"Print log(e).",
			"Print log(10).",
		},
		Keywords: []string{"math", "logarithm", "ln"},
		SeeAlso:  []string{"log10", "log2", "exp", "e"},
	})

	r.Register(&HelpEntry{
		Name:        "log10",
		Description: "Calculate base-10 logarithm",
		Category:    "function",
		LongDesc:    "Returns the base-10 logarithm of a number.",
		Examples: []string{
			"Print log10(100).",
			"Print log10(1000).",
		},
		Keywords: []string{"math", "logarithm", "log"},
		SeeAlso:  []string{"log", "log2"},
	})

	r.Register(&HelpEntry{
		Name:        "log2",
		Description: "Calculate base-2 logarithm",
		Category:    "function",
		LongDesc:    "Returns the base-2 logarithm of a number.",
		Examples: []string{
			"Print log2(8).",
			"Print log2(256).",
		},
		Keywords: []string{"math", "logarithm", "binary"},
		SeeAlso:  []string{"log", "log10"},
	})

	r.Register(&HelpEntry{
		Name:        "exp",
		Description: "Calculate e raised to a power",
		Category:    "function",
		LongDesc:    "Returns e raised to the given power.",
		Examples: []string{
			"Print exp(1).",
			"Print exp(2).",
		},
		Keywords: []string{"math", "exponential", "e"},
		SeeAlso:  []string{"log", "pow", "e"},
	})

	r.Register(&HelpEntry{
		Name:        "random",
		Description: "Generate random number between 0 and 1",
		Category:    "function",
		LongDesc:    "Returns a random floating-point number in the range [0, 1).",
		Examples: []string{
			"Print random().",
			"Declare r to be random().",
		},
		Keywords: []string{"math", "random", "rand"},
		SeeAlso:  []string{"random_between"},
	})

	r.Register(&HelpEntry{
		Name:        "random_between",
		Description: "Generate random number in a range",
		Category:    "function",
		LongDesc:    "Returns a random floating-point number between min (inclusive) and max (exclusive).",
		Examples: []string{
			"Print random_between(1, 10).",
			"Declare roll to be random_between(1, 7).",
		},
		Keywords: []string{"math", "random", "rand", "range"},
		SeeAlso:  []string{"random"},
	})

	r.Register(&HelpEntry{
		Name:        "is_nan",
		Description: "Check if a number is NaN (Not a Number)",
		Category:    "function",
		LongDesc:    "Returns true if the value is NaN, false otherwise.",
		Examples: []string{
			"If is_nan(result), then print \"Invalid\". thats it.",
		},
		Keywords: []string{"math", "nan", "validation"},
		SeeAlso:  []string{"is_infinite"},
	})

	r.Register(&HelpEntry{
		Name:        "is_infinite",
		Description: "Check if a number is infinite",
		Category:    "function",
		LongDesc:    "Returns true if the value is positive or negative infinity.",
		Examples: []string{
			"If is_infinite(result), then print \"Infinite\". thats it.",
		},
		Keywords: []string{"math", "infinity", "validation"},
		SeeAlso:  []string{"is_nan", "infinity"},
	})

	r.Register(&HelpEntry{
		Name:        "clamp",
		Description: "Clamp a value between minimum and maximum",
		Category:    "function",
		LongDesc:    "Returns value clamped to the range [min, max].",
		Examples: []string{
			"Print clamp(15, 0, 10).",
			"Declare bounded to be clamp(x, -100, 100).",
		},
		Keywords: []string{"math", "limit", "bound", "constrain"},
		SeeAlso:  []string{"min", "max"},
	})

	r.Register(&HelpEntry{
		Name:        "sign",
		Description: "Get the sign of a number",
		Category:    "function",
		LongDesc:    "Returns -1 for negative numbers, 0 for zero, and 1 for positive numbers.",
		Examples: []string{
			"Print sign(-5).",
			"Print sign(0).",
			"Print sign(42).",
		},
		Keywords: []string{"math", "positive", "negative", "signum"},
		SeeAlso:  []string{"abs"},
	})

	r.Register(&HelpEntry{
		Name:        "is_integer",
		Description: "Check if a number is an integer",
		Category:    "function",
		LongDesc:    "Returns true if the number has no fractional part.",
		Examples: []string{
			"If is_integer(x), then print \"whole number\". thats it.",
			"Print 5.0's is_integer.",
			"Print 5.5's is_integer.",
		},
		Keywords: []string{"math", "validation", "whole"},
		SeeAlso:  []string{"floor", "ceil"},
	})

	// Math constants
	r.Register(&HelpEntry{
		Name:        "pi",
		Description: "Mathematical constant π (3.14159...)",
		Category:    "constant",
		LongDesc:    "The ratio of a circle's circumference to its diameter.",
		Examples: []string{
			"Print pi.",
			"Declare circumference to be 2 * pi * radius.",
		},
		Keywords: []string{"math", "circle", "constant"},
		SeeAlso:  []string{"e"},
	})

	r.Register(&HelpEntry{
		Name:        "e",
		Description: "Mathematical constant e (2.71828...)",
		Category:    "constant",
		LongDesc:    "Euler's number, the base of natural logarithms.",
		Examples: []string{
			"Print e.",
			"Print exp(1).",
		},
		Keywords: []string{"math", "euler", "constant", "exponential"},
		SeeAlso:  []string{"pi", "exp", "log"},
	})

	r.Register(&HelpEntry{
		Name:        "infinity",
		Description: "Positive infinity constant",
		Category:    "constant",
		LongDesc:    "Represents positive infinity. Negative infinity is -infinity.",
		Examples: []string{
			"Print infinity.",
			"Print -infinity.",
		},
		Keywords: []string{"math", "infinite", "constant"},
		SeeAlso:  []string{"is_infinite"},
	})

	// ═══════════════════════════════════════════════════════════════════════════
	// STRING FUNCTIONS
	// ═══════════════════════════════════════════════════════════════════════════

	r.Register(&HelpEntry{
		Name:        "uppercase",
		Description: "Convert text to uppercase",
		Category:    "function",
		LongDesc:    "Returns the text with all letters converted to uppercase.",
		Examples: []string{
			"Print uppercase(\"hello\").",
			"Print \"hello\"'s uppercase.",
		},
		Keywords: []string{"string", "case", "upper", "caps"},
		SeeAlso:  []string{"lowercase", "title"},
	})

	r.Register(&HelpEntry{
		Name:        "lowercase",
		Description: "Convert text to lowercase",
		Category:    "function",
		LongDesc:    "Returns the text with all letters converted to lowercase.",
		Examples: []string{
			"Print lowercase(\"HELLO\").",
			"Print \"HELLO\"'s lowercase.",
		},
		Keywords: []string{"string", "case", "lower"},
		SeeAlso:  []string{"uppercase", "casefold"},
	})

	r.Register(&HelpEntry{
		Name:        "casefold",
		Description: "Convert to lowercase for case-insensitive comparison",
		Category:    "function",
		LongDesc:    "Similar to lowercase but more aggressive for case-insensitive comparisons.",
		Examples: []string{
			"Print casefold(\"Straße\").",
		},
		Keywords: []string{"string", "case", "comparison"},
		SeeAlso:  []string{"lowercase"},
	})

	r.Register(&HelpEntry{
		Name:        "title",
		Description: "Convert to title case",
		Category:    "function",
		LongDesc:    "Returns the text with the first letter of each word capitalized.",
		Examples: []string{
			"Print title(\"hello world\").",
			"Print \"hello world\"'s title.",
		},
		Keywords: []string{"string", "case", "capitalize"},
		SeeAlso:  []string{"capitalize", "uppercase"},
	})

	r.Register(&HelpEntry{
		Name:        "capitalize",
		Description: "Capitalize first letter",
		Category:    "function",
		LongDesc:    "Returns the text with the first letter capitalized and the rest lowercase.",
		Examples: []string{
			"Print capitalize(\"hello\").",
		},
		Keywords: []string{"string", "case"},
		SeeAlso:  []string{"title"},
	})

	r.Register(&HelpEntry{
		Name:        "swapcase",
		Description: "Swap uppercase and lowercase",
		Category:    "function",
		LongDesc:    "Returns the text with uppercase letters converted to lowercase and vice versa.",
		Examples: []string{
			"Print swapcase(\"HeLLo\").",
		},
		Keywords: []string{"string", "case", "toggle"},
		SeeAlso:  []string{"uppercase", "lowercase"},
	})

	r.Register(&HelpEntry{
		Name:        "trim",
		Description: "Remove leading and trailing whitespace",
		Category:    "function",
		LongDesc:    "Returns the text with whitespace removed from both ends.",
		Examples: []string{
			"Print trim(\"  hello  \").",
		},
		Keywords: []string{"string", "whitespace", "strip"},
		SeeAlso:  []string{"trim_left", "trim_right"},
	})

	r.Register(&HelpEntry{
		Name:        "trim_left",
		Description: "Remove leading whitespace",
		Category:    "function",
		LongDesc:    "Returns the text with whitespace removed from the left side.",
		Examples: []string{
			"Print trim_left(\"  hello\").",
		},
		Keywords: []string{"string", "whitespace", "strip"},
		SeeAlso:  []string{"trim", "trim_right"},
	})

	r.Register(&HelpEntry{
		Name:        "trim_right",
		Description: "Remove trailing whitespace",
		Category:    "function",
		LongDesc:    "Returns the text with whitespace removed from the right side.",
		Examples: []string{
			"Print trim_right(\"hello  \").",
		},
		Keywords: []string{"string", "whitespace", "strip"},
		SeeAlso:  []string{"trim", "trim_left"},
	})

	r.Register(&HelpEntry{
		Name:        "split",
		Description: "Split text into a list",
		Category:    "function",
		LongDesc:    "Splits the text at each occurrence of the delimiter and returns a list of parts.",
		Examples: []string{
			"Print split(\"a,b,c\", \",\").",
			"Declare words to be split(text, \" \").",
		},
		Keywords: []string{"string", "parse", "tokenize"},
		SeeAlso:  []string{"join"},
	})

	r.Register(&HelpEntry{
		Name:        "join",
		Description: "Join a list into text",
		Category:    "function",
		LongDesc:    "Joins the elements of a list into a single string with the separator between elements.",
		Examples: []string{
			"Print join([\"a\", \"b\", \"c\"], \",\").",
			"Declare text to be join(words, \" \").",
		},
		Keywords: []string{"string", "concatenate", "combine"},
		SeeAlso:  []string{"split"},
	})

	r.Register(&HelpEntry{
		Name:        "replace",
		Description: "Replace all occurrences of substring",
		Category:    "function",
		LongDesc:    "Returns the text with all occurrences of 'old' replaced with 'new'.",
		Examples: []string{
			"Print replace(\"hello world\", \"world\", \"there\").",
		},
		Keywords: []string{"string", "substitute", "change"},
		SeeAlso:  []string{"contains"},
	})

	r.Register(&HelpEntry{
		Name:        "contains",
		Description: "Check if text contains substring",
		Category:    "function",
		LongDesc:    "Returns true if the text contains the substring.",
		Examples: []string{
			"If contains(text, \"hello\"), then print \"Found\". thats it.",
		},
		Keywords: []string{"string", "search", "find", "includes"},
		SeeAlso:  []string{"starts_with", "ends_with", "index_of"},
	})

	r.Register(&HelpEntry{
		Name:        "starts_with",
		Description: "Check if text starts with prefix",
		Category:    "function",
		LongDesc:    "Returns true if the text starts with the prefix.",
		Examples: []string{
			"If starts_with(text, \"Hello\"), then print \"Greeting\". thats it.",
		},
		Keywords: []string{"string", "prefix", "begin"},
		SeeAlso:  []string{"ends_with", "contains"},
	})

	r.Register(&HelpEntry{
		Name:        "ends_with",
		Description: "Check if text ends with suffix",
		Category:    "function",
		LongDesc:    "Returns true if the text ends with the suffix.",
		Examples: []string{
			"If ends_with(filename, \".txt\"), then print \"Text file\". thats it.",
		},
		Keywords: []string{"string", "suffix", "extension"},
		SeeAlso:  []string{"starts_with", "contains"},
	})

	r.Register(&HelpEntry{
		Name:        "index_of",
		Description: "Find index of substring",
		Category:    "function",
		LongDesc:    "Returns the index of the first occurrence of substring, or -1 if not found.",
		Examples: []string{
			"Print index_of(\"hello world\", \"world\").",
		},
		Keywords: []string{"string", "search", "find", "position"},
		SeeAlso:  []string{"contains"},
	})

	r.Register(&HelpEntry{
		Name:        "substring",
		Description: "Extract substring",
		Category:    "function",
		LongDesc:    "Returns a substring starting at 'start' with the specified 'length'.",
		Examples: []string{
			"Print substring(\"hello\", 1, 3).",
		},
		Keywords: []string{"string", "extract", "slice"},
		SeeAlso:  []string{"slice"},
	})

	r.Register(&HelpEntry{
		Name:        "str_repeat",
		Description: "Repeat text multiple times",
		Category:    "function",
		LongDesc:    "Returns the text repeated n times.",
		Examples: []string{
			"Print str_repeat(\"ha\", 3).",
		},
		Keywords: []string{"string", "repeat", "multiply"},
		SeeAlso:  []string{"join"},
	})

	r.Register(&HelpEntry{
		Name:        "count_occurrences",
		Description: "Count occurrences of substring",
		Category:    "function",
		LongDesc:    "Returns the number of times substring appears in the text.",
		Examples: []string{
			"Print count_occurrences(\"hello hello\", \"hello\").",
		},
		Keywords: []string{"string", "count", "search"},
		SeeAlso:  []string{"contains"},
	})

	r.Register(&HelpEntry{
		Name:        "pad_left",
		Description: "Pad text on the left",
		Category:    "function",
		LongDesc:    "Returns the text padded on the left with the specified character to reach the width.",
		Examples: []string{
			"Print pad_left(\"42\", 5, \" \").",
		},
		Keywords: []string{"string", "padding", "align"},
		SeeAlso:  []string{"pad_right", "center", "zfill"},
	})

	r.Register(&HelpEntry{
		Name:        "pad_right",
		Description: "Pad text on the right",
		Category:    "function",
		LongDesc:    "Returns the text padded on the right with the specified character to reach the width.",
		Examples: []string{
			"Print pad_right(\"42\", 5, \" \").",
		},
		Keywords: []string{"string", "padding", "align"},
		SeeAlso:  []string{"pad_left", "center"},
	})

	r.Register(&HelpEntry{
		Name:        "center",
		Description: "Center text with padding",
		Category:    "function",
		LongDesc:    "Returns the text centered with the specified character on both sides to reach the width.",
		Examples: []string{
			"Print center(\"Title\", 20, \"=\").",
		},
		Keywords: []string{"string", "padding", "align"},
		SeeAlso:  []string{"pad_left", "pad_right"},
	})

	r.Register(&HelpEntry{
		Name:        "zfill",
		Description: "Pad number with zeros",
		Category:    "function",
		LongDesc:    "Returns the text padded with zeros on the left to reach the width.",
		Examples: []string{
			"Print zfill(\"42\", 5).",
		},
		Keywords: []string{"string", "padding", "zeros", "format"},
		SeeAlso:  []string{"pad_left"},
	})

	r.Register(&HelpEntry{
		Name:        "to_number",
		Description: "Convert text to number",
		Category:    "function",
		LongDesc:    "Parses the text as a number. Returns an error if the text is not a valid number.",
		Examples: []string{
			"Declare num to be to_number(\"42\").",
			"Declare pi_val to be to_number(\"3.14\").",
		},
		Keywords: []string{"string", "parse", "convert", "cast"},
		SeeAlso:  []string{"to_string"},
	})

	r.Register(&HelpEntry{
		Name:        "to_string",
		Description: "Convert value to text",
		Category:    "function",
		LongDesc:    "Converts any value to its string representation.",
		Examples: []string{
			"Print to_string(42).",
			"Declare text to be to_string(3.14).",
		},
		Keywords: []string{"string", "convert", "cast", "format"},
		SeeAlso:  []string{"to_number"},
	})

	r.Register(&HelpEntry{
		Name:        "is_empty",
		Description: "Check if text is empty",
		Category:    "function",
		LongDesc:    "Returns true if the text has zero length.",
		Examples: []string{
			"If is_empty(name), then print \"No name\". thats it.",
		},
		Keywords: []string{"string", "validation", "check"},
		SeeAlso:  []string{"trim"},
	})

	r.Register(&HelpEntry{
		Name:        "is_digit",
		Description: "Check if all characters are digits",
		Category:    "function",
		LongDesc:    "Returns true if all characters in the text are digits.",
		Examples: []string{
			"If is_digit(input), then print \"Valid number\". thats it.",
		},
		Keywords: []string{"string", "validation", "numeric"},
		SeeAlso:  []string{"is_alpha", "is_alnum"},
	})

	r.Register(&HelpEntry{
		Name:        "is_alpha",
		Description: "Check if all characters are letters",
		Category:    "function",
		LongDesc:    "Returns true if all characters in the text are letters.",
		Examples: []string{
			"If is_alpha(input), then print \"Valid name\". thats it.",
		},
		Keywords: []string{"string", "validation", "letters"},
		SeeAlso:  []string{"is_digit", "is_alnum"},
	})

	r.Register(&HelpEntry{
		Name:        "is_alnum",
		Description: "Check if all characters are alphanumeric",
		Category:    "function",
		LongDesc:    "Returns true if all characters are letters or digits.",
		Examples: []string{
			"If is_alnum(username), then print \"Valid\". thats it.",
		},
		Keywords: []string{"string", "validation", "alphanumeric"},
		SeeAlso:  []string{"is_digit", "is_alpha"},
	})

	r.Register(&HelpEntry{
		Name:        "is_space",
		Description: "Check if all characters are whitespace",
		Category:    "function",
		LongDesc:    "Returns true if all characters are whitespace.",
		Examples: []string{
			"If is_space(text), then print \"Empty\". thats it.",
		},
		Keywords: []string{"string", "validation", "whitespace"},
		SeeAlso:  []string{"trim"},
	})

	r.Register(&HelpEntry{
		Name:        "is_upper",
		Description: "Check if all letters are uppercase",
		Category:    "function",
		LongDesc:    "Returns true if all letter characters are uppercase.",
		Examples: []string{
			"If is_upper(text), then print \"SHOUTING\". thats it.",
		},
		Keywords: []string{"string", "validation", "case"},
		SeeAlso:  []string{"is_lower", "uppercase"},
	})

	r.Register(&HelpEntry{
		Name:        "is_lower",
		Description: "Check if all letters are lowercase",
		Category:    "function",
		LongDesc:    "Returns true if all letter characters are lowercase.",
		Examples: []string{
			"If is_lower(text), then print \"lowercase\". thats it.",
		},
		Keywords: []string{"string", "validation", "case"},
		SeeAlso:  []string{"is_upper", "lowercase"},
	})

	// ═══════════════════════════════════════════════════════════════════════════
	// LIST FUNCTIONS
	// ═══════════════════════════════════════════════════════════════════════════

	r.Register(&HelpEntry{
		Name:        "append",
		Description: "Add element to end of list",
		Category:    "function",
		LongDesc:    "Adds an item to the end of the list. Modifies the list in place.",
		Examples: []string{
			"Call append(items, 4).",
		},
		Keywords: []string{"list", "add", "push", "insert"},
		SeeAlso:  []string{"insert", "remove"},
	})

	r.Register(&HelpEntry{
		Name:        "remove",
		Description: "Remove element at index from list",
		Category:    "function",
		LongDesc:    "Removes and returns the element at the specified index. Modifies the list in place.",
		Examples: []string{
			"Declare removed to be remove(items, 0).",
		},
		Keywords: []string{"list", "delete", "pop"},
		SeeAlso:  []string{"append", "insert"},
	})

	r.Register(&HelpEntry{
		Name:        "insert",
		Description: "Insert element at index in list",
		Category:    "function",
		LongDesc:    "Inserts an item at the specified index. Modifies the list in place.",
		Examples: []string{
			"Call insert(items, 1, \"new\").",
		},
		Keywords: []string{"list", "add", "splice"},
		SeeAlso:  []string{"append", "remove"},
	})

	r.Register(&HelpEntry{
		Name:        "sort",
		Description: "Sort list in ascending order",
		Category:    "function",
		LongDesc:    "Sorts the list in ascending order. Modifies the list in place.",
		Examples: []string{
			"Call sort(numbers).",
		},
		Keywords: []string{"list", "order", "arrange"},
		SeeAlso:  []string{"sorted_desc", "reverse"},
	})

	r.Register(&HelpEntry{
		Name:        "sorted_desc",
		Description: "Sort list in descending order",
		Category:    "function",
		LongDesc:    "Sorts the list in descending order. Modifies the list in place.",
		Examples: []string{
			"Call sorted_desc(numbers).",
		},
		Keywords: []string{"list", "order", "arrange"},
		SeeAlso:  []string{"sort", "reverse"},
	})

	r.Register(&HelpEntry{
		Name:        "reverse",
		Description: "Reverse the order of list elements",
		Category:    "function",
		LongDesc:    "Reverses the list in place.",
		Examples: []string{
			"Call reverse(items).",
		},
		Keywords: []string{"list", "flip", "invert"},
		SeeAlso:  []string{"sort"},
	})

	r.Register(&HelpEntry{
		Name:        "slice",
		Description: "Extract sublist",
		Category:    "function",
		LongDesc:    "Returns a new list containing elements from start index to end index (exclusive).",
		Examples: []string{
			"Declare sub to be slice(items, 1, 4).",
		},
		Keywords: []string{"list", "extract", "substring"},
		SeeAlso:  []string{"substring"},
	})

	r.Register(&HelpEntry{
		Name:        "count",
		Description: "Get number of elements in list",
		Category:    "function",
		LongDesc:    "Returns the number of elements in the list.",
		Examples: []string{
			"Print count(items).",
			"Declare size to be count(numbers).",
		},
		Keywords: []string{"list", "length", "size"},
		SeeAlso:  []string{"is_empty"},
	})

	r.Register(&HelpEntry{
		Name:        "sum",
		Description: "Calculate sum of numeric list",
		Category:    "function",
		LongDesc:    "Returns the sum of all numbers in the list.",
		Examples: []string{
			"Print sum([1, 2, 3, 4, 5]).",
			"Declare total to be sum(numbers).",
		},
		Keywords: []string{"list", "math", "total", "add"},
		SeeAlso:  []string{"average", "product"},
	})

	r.Register(&HelpEntry{
		Name:        "average",
		Description: "Calculate average of numeric list",
		Category:    "function",
		LongDesc:    "Returns the arithmetic mean of all numbers in the list.",
		Examples: []string{
			"Print average([1, 2, 3, 4, 5]).",
			"Declare mean to be average(scores).",
		},
		Keywords: []string{"list", "math", "mean"},
		SeeAlso:  []string{"sum"},
	})

	r.Register(&HelpEntry{
		Name:        "min_value",
		Description: "Find minimum value in list",
		Category:    "function",
		LongDesc:    "Returns the smallest value in the list.",
		Examples: []string{
			"Print min_value([5, 2, 8, 1]).",
		},
		Keywords: []string{"list", "math", "minimum", "smallest"},
		SeeAlso:  []string{"max_value", "min"},
	})

	r.Register(&HelpEntry{
		Name:        "max_value",
		Description: "Find maximum value in list",
		Category:    "function",
		LongDesc:    "Returns the largest value in the list.",
		Examples: []string{
			"Print max_value([5, 2, 8, 1]).",
		},
		Keywords: []string{"list", "math", "maximum", "largest"},
		SeeAlso:  []string{"min_value", "max"},
	})

	r.Register(&HelpEntry{
		Name:        "product",
		Description: "Calculate product of numeric list",
		Category:    "function",
		LongDesc:    "Returns the product of all numbers in the list (multiplies them together).",
		Examples: []string{
			"Print product([2, 3, 4]).",
		},
		Keywords: []string{"list", "math", "multiply"},
		SeeAlso:  []string{"sum"},
	})

	r.Register(&HelpEntry{
		Name:        "first",
		Description: "Get first element of list",
		Category:    "function",
		LongDesc:    "Returns the first element of the list.",
		Examples: []string{
			"Print first(items).",
		},
		Keywords: []string{"list", "head", "start"},
		SeeAlso:  []string{"last"},
	})

	r.Register(&HelpEntry{
		Name:        "last",
		Description: "Get last element of list",
		Category:    "function",
		LongDesc:    "Returns the last element of the list.",
		Examples: []string{
			"Print last(items).",
		},
		Keywords: []string{"list", "tail", "end"},
		SeeAlso:  []string{"first"},
	})

	r.Register(&HelpEntry{
		Name:        "unique",
		Description: "Remove duplicates from list",
		Category:    "function",
		LongDesc:    "Returns a new list with duplicate values removed.",
		Examples: []string{
			"Declare distinct to be unique([1, 2, 2, 3, 1]).",
		},
		Keywords: []string{"list", "deduplicate", "distinct"},
		SeeAlso:  []string{"sort"},
	})

	r.Register(&HelpEntry{
		Name:        "flatten",
		Description: "Flatten nested lists",
		Category:    "function",
		LongDesc:    "Flattens nested lists into a single flat list.",
		Examples: []string{
			"Declare flat to be flatten([[1, 2], [3, 4]]).",
		},
		Keywords: []string{"list", "nested", "unnest"},
		SeeAlso:  []string{"zip_with"},
	})

	r.Register(&HelpEntry{
		Name:        "any_true",
		Description: "Check if any element is true",
		Category:    "function",
		LongDesc:    "Returns true if at least one element in the list is true.",
		Examples: []string{
			"If any_true(flags), then print \"At least one is true\". thats it.",
		},
		Keywords: []string{"list", "boolean", "or"},
		SeeAlso:  []string{"all_true"},
	})

	r.Register(&HelpEntry{
		Name:        "all_true",
		Description: "Check if all elements are true",
		Category:    "function",
		LongDesc:    "Returns true if all elements in the list are true.",
		Examples: []string{
			"If all_true(checks), then print \"All passed\". thats it.",
		},
		Keywords: []string{"list", "boolean", "and"},
		SeeAlso:  []string{"any_true"},
	})

	r.Register(&HelpEntry{
		Name:        "zip_with",
		Description: "Pair elements from two lists",
		Category:    "function",
		LongDesc:    "Returns a list of pairs by combining elements from two lists.",
		Examples: []string{
			"Declare pairs to be zip_with([1, 2], [\"a\", \"b\"]).",
		},
		Keywords: []string{"list", "combine", "pair"},
		SeeAlso:  []string{"flatten"},
	})

	// ═══════════════════════════════════════════════════════════════════════════
	// LOOKUP TABLE FUNCTIONS
	// ═══════════════════════════════════════════════════════════════════════════

	r.Register(&HelpEntry{
		Name:        "keys",
		Description: "Get all keys from lookup table",
		Category:    "function",
		LongDesc:    "Returns a list of all keys in the lookup table.",
		Examples: []string{
			"Declare all_keys to be keys(table).",
			"For each k in keys(scores), print k.",
		},
		Keywords: []string{"dictionary", "map", "lookup"},
		SeeAlso:  []string{"values", "table_has"},
	})

	r.Register(&HelpEntry{
		Name:        "values",
		Description: "Get all values from lookup table",
		Category:    "function",
		LongDesc:    "Returns a list of all values in the lookup table.",
		Examples: []string{
			"Declare all_values to be values(table).",
		},
		Keywords: []string{"dictionary", "map", "lookup"},
		SeeAlso:  []string{"keys"},
	})

	r.Register(&HelpEntry{
		Name:        "table_remove",
		Description: "Remove key from lookup table",
		Category:    "function",
		LongDesc:    "Removes the specified key and its value from the lookup table.",
		Examples: []string{
			"Call table_remove(scores, \"Alice\").",
		},
		Keywords: []string{"dictionary", "map", "lookup", "delete"},
		SeeAlso:  []string{"table_has"},
	})

	r.Register(&HelpEntry{
		Name:        "table_has",
		Description: "Check if lookup table has key",
		Category:    "function",
		LongDesc:    "Returns true if the lookup table contains the specified key.",
		Examples: []string{
			"If table_has(scores, \"Bob\"), then print \"Found\". thats it.",
		},
		Keywords: []string{"dictionary", "map", "lookup", "contains"},
		Aliases:  []string{"has"},
		SeeAlso:  []string{"keys"},
	})

	r.Register(&HelpEntry{
		Name:        "merge",
		Description: "Merge two lookup tables",
		Category:    "function",
		LongDesc:    "Returns a new lookup table with all key-value pairs from both tables. Values from the second table overwrite values from the first for duplicate keys.",
		Examples: []string{
			"Declare combined to be merge(table1, table2).",
		},
		Keywords: []string{"dictionary", "map", "lookup", "combine"},
		SeeAlso:  []string{"keys", "values"},
	})

	r.Register(&HelpEntry{
		Name:        "get_or_default",
		Description: "Get value with fallback default",
		Category:    "function",
		LongDesc:    "Returns the value for the key, or the default value if the key doesn't exist.",
		Examples: []string{
			"Declare score to be get_or_default(scores, \"Unknown\", 0).",
		},
		Keywords: []string{"dictionary", "map", "lookup", "default", "safe"},
		SeeAlso:  []string{"table_has"},
	})

	// ═══════════════════════════════════════════════════════════════════════════
	// TIME FUNCTIONS
	// ═══════════════════════════════════════════════════════════════════════════

	r.Register(&HelpEntry{
		Name:        "current_time",
		Description: "Get current date and time",
		Category:    "function",
		LongDesc:    "Returns the current date and time as a text string.",
		Examples: []string{
			"Print current_time().",
			"Declare now to be current_time().",
		},
		Keywords: []string{"time", "date", "now", "timestamp"},
		SeeAlso:  []string{"elapsed_time", "sleep"},
	})

	r.Register(&HelpEntry{
		Name:        "elapsed_time",
		Description: "Get elapsed time since program start",
		Category:    "function",
		LongDesc:    "Returns the number of seconds elapsed since the program started.",
		Examples: []string{
			"Print elapsed_time().",
		},
		Keywords: []string{"time", "duration", "timer"},
		SeeAlso:  []string{"current_time"},
	})

	r.Register(&HelpEntry{
		Name:        "sleep",
		Description: "Pause execution for a duration",
		Category:    "function",
		LongDesc:    "Pauses program execution for the specified number of seconds. Can also use 'Sleep for N <unit>' syntax.",
		Examples: []string{
			"Call sleep(2).",
			"Sleep for 1 second.",
			"Sleep for 500 milliseconds.",
			"Wait for 2 seconds.",
		},
		Keywords: []string{"time", "wait", "delay", "pause"},
		Aliases:  []string{"wait"},
		SeeAlso:  []string{"current_time"},
	})

	// ═══════════════════════════════════════════════════════════════════════════
	// SPECIAL OPERATIONS
	// ═══════════════════════════════════════════════════════════════════════════

	r.Register(&HelpEntry{
		Name:        "toggle",
		Description: "Flip a boolean variable",
		Category:    "keyword",
		LongDesc:    "Toggles a boolean variable between true and false.",
		Examples: []string{
			"Toggle flag.",
		},
		Keywords: []string{"boolean", "flip", "invert", "negate"},
		SeeAlso:  []string{"boolean"},
	})

	r.Register(&HelpEntry{
		Name:        "swap",
		Description: "Swap values of two variables",
		Category:    "keyword",
		LongDesc:    "Swaps the values of two variables in a single operation.",
		Examples: []string{
			"Swap a and b.",
		},
		Keywords: []string{"exchange", "switch"},
		SeeAlso:  []string{"set"},
	})
}
