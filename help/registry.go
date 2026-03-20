package help

// loadDefaultEntries populates the registry with comprehensive help content.
func (r *Registry) loadDefaultEntries() {
	// REPL Commands
	r.Register(&HelpEntry{
		Name:        "help",
		Description: "Display help information",
		Category:    "command",
		LongDesc:    "The help command displays help information. Use 'help <topic>' to search for specific topics using fuzzy matching.",
		Examples: []string{
			"help",
			"help print",
			"help loop",
		},
		Keywords: []string{"help", "documentation", "info", "assist"},
		Aliases:  []string{"?"},
	})

	r.Register(&HelpEntry{
		Name:        "exit",
		Description: "Exit the REPL",
		Category:    "command",
		LongDesc:    "Exits the REPL and returns to the shell. You can also use 'quit'.",
		Examples:    []string{"exit", "quit"},
		Keywords:    []string{"quit", "leave", "close"},
		Aliases:     []string{"quit"},
	})

	// Variables and Declarations
	r.Register(&HelpEntry{
		Name:        "declare",
		Description: "Declare a variable",
		Category:    "keyword",
		LongDesc:    "Use 'Declare' to create a variable. Variables can be typed or untyped.",
		Examples: []string{
			"Declare x to be 5.",
			"Declare name to be \"Alice\".",
			"Declare x as number to be 10.",
			"Declare items to be [1, 2, 3].",
		},
		Keywords: []string{"variable", "assign", "set", "define"},
		SeeAlso:  []string{"set", "types"},
	})

	r.Register(&HelpEntry{
		Name:        "set",
		Description: "Change the value of an existing variable",
		Category:    "keyword",
		LongDesc:    "Use 'Set' to change the value of a previously declared variable.",
		Examples: []string{
			"Set x to 10.",
			"Set name to \"Bob\".",
			"Set items to [4, 5, 6].",
		},
		Keywords: []string{"assign", "change", "update", "modify"},
		SeeAlso:  []string{"declare"},
	})

	// Control Flow
	r.Register(&HelpEntry{
		Name:        "if",
		Description: "Conditional execution",
		Category:    "keyword",
		LongDesc:    "Use 'If' statements to execute code conditionally. Supports 'otherwise' (else) and 'otherwise if' (else if) clauses.",
		Examples: []string{
			"If x is 5, then print \"x is five\". thats it.",
			"If x > 10, then do the following:\n    Print \"large\".\nthats it.",
			"If x < 0, then print \"negative\". otherwise if x is 0, then print \"zero\". otherwise print \"positive\". thats it.",
		},
		Keywords: []string{"conditional", "then", "otherwise", "else"},
		Aliases:  []string{"conditional"},
		SeeAlso:  []string{"comparison", "boolean"},
	})

	r.Register(&HelpEntry{
		Name:        "for each",
		Description: "Iterate over a collection",
		Category:    "keyword",
		LongDesc:    "Use 'For each' to iterate over items in a list or range.",
		Examples: []string{
			"For each n in [1, 2, 3], do the following:\n    Print the value of n.\nthats it.",
			"For each item in items, print the value of item.",
			"For each n in a range from 1 to 10, print the value of n.",
		},
		Keywords: []string{"loop", "iterate", "collection", "list", "array"},
		Aliases:  []string{"foreach", "for"},
		SeeAlso:  []string{"repeat", "while", "range"},
	})

	r.Register(&HelpEntry{
		Name:        "repeat while",
		Description: "Loop while a condition is true",
		Category:    "keyword",
		LongDesc:    "Use 'repeat the following while' to create a while loop that continues as long as the condition is true.",
		Examples: []string{
			"Declare i to be 0.\nRepeat the following while i < 10:\n    Print the value of i.\n    Set i to i + 1.\nthats it.",
		},
		Keywords: []string{"while", "loop", "condition"},
		Aliases:  []string{"while"},
		SeeAlso:  []string{"for each", "repeat"},
	})

	r.Register(&HelpEntry{
		Name:        "repeat",
		Description: "Repeat a block a fixed number of times",
		Category:    "keyword",
		LongDesc:    "Use 'repeat' to execute a block multiple times.",
		Examples: []string{
			"Repeat 5 times:\n    Print \"Hello\".\nthats it.",
		},
		Keywords: []string{"loop", "times", "iterate"},
		SeeAlso:  []string{"for each", "repeat while"},
	})

	// Functions
	r.Register(&HelpEntry{
		Name:        "print",
		Description: "Output text to the console",
		Category:    "function",
		LongDesc:    "Print outputs values to the console. Use 'the value of' to print variable values.",
		Examples: []string{
			"Print \"Hello, World!\".",
			"Print the value of x.",
			"Print the result of 2 + 2.",
		},
		Keywords: []string{"output", "display", "show", "console"},
		SeeAlso:  []string{"input"},
	})

	r.Register(&HelpEntry{
		Name:        "ask",
		Description: "Get input from the user",
		Category:    "function",
		LongDesc:    "Use 'Ask' to prompt the user for input and store it in a variable.",
		Examples: []string{
			"Ask the user for a name and declare it as name.",
			"Ask the user for a number and declare it as age.",
		},
		Keywords: []string{"input", "prompt", "user input", "read"},
		Aliases:  []string{"input", "read"},
		SeeAlso:  []string{"print"},
	})

	// Data Types
	r.Register(&HelpEntry{
		Name:        "number",
		Description: "Numeric data type",
		Category:    "type",
		LongDesc:    "Numbers can be integers or floating-point values. Supports arithmetic operations.",
		Examples: []string{
			"Declare x as number to be 5.",
			"Declare pi as number to be 3.14159.",
			"Print the result of 10 + 20.",
		},
		Keywords: []string{"integer", "float", "numeric", "math"},
		SeeAlso:  []string{"text", "boolean", "list"},
	})

	r.Register(&HelpEntry{
		Name:        "text",
		Description: "String/text data type",
		Category:    "type",
		LongDesc:    "Text values are strings enclosed in quotes. Supports concatenation and various string operations.",
		Examples: []string{
			"Declare name as text to be \"Alice\".",
			"Print \"Hello, \" + name + \"!\".",
		},
		Keywords: []string{"string", "character", "word"},
		Aliases:  []string{"string"},
		SeeAlso:  []string{"number", "boolean"},
	})

	r.Register(&HelpEntry{
		Name:        "boolean",
		Description: "True/false data type",
		Category:    "type",
		LongDesc:    "Boolean values are either true or false. Used in conditions and logic.",
		Examples: []string{
			"Declare is_active as boolean to be true.",
			"If is_active, then print \"Active\". thats it.",
		},
		Keywords: []string{"true", "false", "logical", "bool"},
		SeeAlso:  []string{"comparison", "if"},
	})

	r.Register(&HelpEntry{
		Name:        "list",
		Description: "Ordered collection of values",
		Category:    "type",
		LongDesc:    "Lists are ordered collections that can contain any type of values. Access elements by index.",
		Examples: []string{
			"Declare items to be [1, 2, 3, 4, 5].",
			"Print the value of items[0].",
			"For each item in items, print the value of item.",
		},
		Keywords: []string{"array", "collection", "sequence"},
		Aliases:  []string{"array"},
		SeeAlso:  []string{"for each", "range"},
	})

	r.Register(&HelpEntry{
		Name:        "range",
		Description: "Create a sequence of numbers",
		Category:    "concept",
		LongDesc:    "Ranges create sequences of consecutive integers. Supports both ascending and descending ranges.",
		Examples: []string{
			"Declare nums to be [1 .. 10].",
			"Declare nums to be a range from 1 to 30.",
			"For each n in [1 .. 5], print the value of n.",
		},
		Keywords: []string{"sequence", "numbers", "from", "to"},
		SeeAlso:  []string{"list", "for each"},
	})

	// Operators
	r.Register(&HelpEntry{
		Name:        "arithmetic",
		Description: "Mathematical operations",
		Category:    "operator",
		LongDesc:    "Arithmetic operators: + (addition), - (subtraction), * (multiplication), / (division), % (modulo)",
		Examples: []string{
			"Print the result of 5 + 3.",
			"Print the result of 10 - 4.",
			"Print the result of 6 * 7.",
			"Print the result of 15 / 3.",
			"Print the result of 10 % 3.",
		},
		Keywords: []string{"math", "addition", "subtraction", "multiplication", "division", "modulo"},
		SeeAlso:  []string{"number"},
	})

	r.Register(&HelpEntry{
		Name:        "comparison",
		Description: "Compare values",
		Category:    "operator",
		LongDesc:    "Comparison operators: is (equal), is not (not equal), > (greater than), < (less than), >= (greater or equal), <= (less or equal)",
		Examples: []string{
			"If x is 5, then print \"equal\". thats it.",
			"If x is not 0, then print \"not zero\". thats it.",
			"If x > 10, then print \"greater\". thats it.",
		},
		Keywords: []string{"equal", "greater", "less", "compare"},
		SeeAlso:  []string{"if", "boolean"},
	})

	// Error Handling
	r.Register(&HelpEntry{
		Name:        "try catch",
		Description: "Handle errors gracefully",
		Category:    "keyword",
		LongDesc:    "Use try/catch blocks to handle errors. Supports 'on error' to catch specific error types and 'finally' for cleanup.",
		Examples: []string{
			"Try the following:\n    Print the result of 10 / 0.\non error:\n    Print \"Cannot divide by zero\".\nthats it.",
			"Try the following:\n    Print \"risky\".\nfinally:\n    Print \"cleanup\".\nthats it.",
		},
		Keywords: []string{"error", "exception", "catch", "finally"},
		Aliases:  []string{"try", "catch"},
		SeeAlso:  []string{"error types"},
	})

	// Structs
	r.Register(&HelpEntry{
		Name:        "struct",
		Description: "Define custom data structures",
		Category:    "concept",
		LongDesc:    "Structs allow you to create custom data types with named fields.",
		Examples: []string{
			"Define a person with a name and an age.",
			"Declare john as a person with name \"John\" and age 30.",
		},
		Keywords: []string{"structure", "object", "type", "custom type"},
		SeeAlso:  []string{"types"},
	})

	// Comments
	r.Register(&HelpEntry{
		Name:        "comments",
		Description: "Add explanatory notes to code",
		Category:    "concept",
		LongDesc:    "Use '#' for single-line comments. Comments are ignored by the interpreter.",
		Examples: []string{
			"# This is a comment",
			"Print \"Hello\". # This prints a greeting",
		},
		Keywords: []string{"note", "documentation", "remark"},
	})

	// Multi-line blocks
	r.Register(&HelpEntry{
		Name:        "do the following",
		Description: "Start a multi-line block",
		Category:    "keyword",
		LongDesc:    "Use 'do the following:' to start a multi-line block. End it with 'thats it.'",
		Examples: []string{
			"If x > 0, then do the following:\n    Print \"positive\".\n    Print \"number\".\nthats it.",
		},
		Keywords: []string{"block", "multiline", "group"},
		SeeAlso:  []string{"thats it"},
	})

	r.Register(&HelpEntry{
		Name:        "thats it",
		Description: "End a multi-line block",
		Category:    "keyword",
		LongDesc:    "Use 'thats it.' to close a block started with 'do the following:'",
		Keywords:    []string{"end", "close", "finish"},
		Aliases:     []string{"that's it"},
		SeeAlso:     []string{"do the following"},
	})

	// Politeness
	r.Register(&HelpEntry{
		Name:        "politeness",
		Description: "Optional polite prefixes for statements",
		Category:    "concept",
		LongDesc:    "You can optionally prefix statements with 'Please', 'Kindly', 'Could you', or 'Would you kindly' for politeness. Use --polite flag to require all statements to be polite.",
		Examples: []string{
			"Please print \"Hello\".",
			"Kindly declare x to be 5.",
			"Could you print the value of x.",
		},
		Keywords: []string{"please", "kindly", "polite", "courteous"},
		SeeAlso:  []string{"run command"},
	})
}
