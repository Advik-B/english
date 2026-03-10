// Package parser – human-readable error messages and hint strings.
//
// Every message that can be shown to a programmer when their code contains a
// syntax error lives here.  Edit this file to improve wording, add extra
// examples, or translate the text – no other file needs to change.
//
// Convention
//   - msg*  constants are the primary error description ("what went wrong").
//   - hint* constants give corrective guidance ("how to fix it").
//   - msgFmt* / hintFmt* constants are fmt.Sprintf format strings whose %s / %q
//     placeholders are filled in by the call site using runtime token values.
package parser

// ─── Hints ───────────────────────────────────────────────────────────────────
// Hints are shown below the primary message and suggest how to fix the problem.

const (
	// General structural hints.
	hintEndWithPeriod  = "Every statement must end with a period (.). Try adding a period at the end of the line."
	hintCloseThatsWith = "Every block must end with 'thats it.' — did you forget to close the block?"
	hintMissingIt      = "The word 'it' is missing before the period. Write 'thats it.' to close the block."
	hintExpectedToBe   = "The word 'to' is missing here. For example: 'Declare x to be 5.' or 'Set x to be 10.'"
	hintExpectedBe     = "The word 'be' is expected here. For example: 'Declare x to be 5.'"
	hintMissingBeAfterTo = "The word 'be' is missing after 'to'. For example: 'Declare x to be 5.'"
	hintNameNotLiteral = "A name (like a variable name or function name) is expected here, not a number or text value."

	// Statement-level hints.
	hintNumberAsStatement = "Numbers must be part of a statement. For example: 'Declare x to be 5.' or 'Print 42.'"
	hintStringAsStatement = `Text must be part of a statement. For example: 'Print "Hello, world!".' or 'Declare greeting to be "Hello".'`
	hintUnexpectedEOF     = "Check that every block (like an 'If' or 'Repeat') is closed with 'thats it.'"
	hintUnknownKeyword    = "Check the spelling of the keyword. Every statement must start with a word like 'Declare', 'Set', 'Print', 'If', 'Repeat', or 'Call'."

	// Variable / constant declarations.
	hintVarNameAfterLet     = "Variable names must start with a letter. For example: 'let count be 10.' or 'let name be \"Alice\".'"
	hintVarNameAfterDeclare = "Variable names must start with a letter. For example: 'Declare score to be 0.' or 'Declare name to always be \"Alice\".'"
	hintFunctionName        = "Function names must start with a letter. For example: 'Declare function greet that does the following:'"
	hintParameterName       = "Parameter names must start with a letter. For example: 'Declare function add that takes x and y and does the following:'"
	hintImportPath          = "For example: 'Import \"myfile.abc\".' or 'Import everything from \"utils.abc\".'"

	// Assignment / Set statements.
	hintSetVarName    = "For example: 'Set score to be 10.' or 'Set name to be \"Alice\".'"
	hintSetTheFull    = "After 'Set the' you can write 'item at position N in myList to be VALUE' or 'entry KEY in myTable to be VALUE'."
	hintSetCallResult = "For example: 'Set result to be the result of calling square of x.'"
	hintSetListName   = "For example: 'Set the item at position 1 in myList to be 99.'"
	hintSetTableFull  = "The full form is: 'Set tableName at key to be value.'"

	// Call statements.
	hintCallName = "For example: 'Call greet.' or 'Call add with 3 and 5.'"

	// Control-flow statements.
	hintForEachVar = "For example: 'For each item in myList:' or 'For each number in scores:'"
	hintBreakLoop  = "For example: 'Break out of the loop.' or 'Break out of this loop.'"

	// Output.
	hintPrintOrWrite = "To show output, use 'Print \"Hello\".' or 'Write \"No newline\".'"

	// Ask (user input).
	hintAskAs   = "For example: 'Ask \"What is your name?\" as userName.'"
	hintAskAnd  = "For example: 'Ask \"What is your age?\" and store it in userAge.'"
	hintAskFull = "For example: 'Ask \"What is your name?\" as myName.' or 'Ask \"Enter a number:\" and store it in num.'"

	// Type cast / possessive / error-type check.
	hintCastType       = "Valid type names are: number, text, boolean, integer. For example: 'x cast to number'."
	hintPossessive     = "For example: 'myText's length' or 'myText's upper'."
	hintErrorTypeCheck = "For example: 'error is NetworkError' or 'error is RuntimeError'."

	// Expressions.
	hintTheExpression     = "After 'the' you can use: 'the value of x', 'the length of myList', 'the remainder of a divided by b', 'the result of calling myFunction', or a field name like 'the age of person'."
	hintExpressionValue   = "Values can be numbers (like 42), text (like \"hello\"), variables (like myScore), function calls, or expressions in parentheses."
	hintIndexInOrOf       = "For example: 'the item at position 1 in myList' or 'the item at position 1 of myList'."
	hintRemainderDividedBy = "For example: 'the remainder of 10 divided by 3' or 'the remainder of a divided by b'."
	hintLocationOf        = "For example: 'the location of myVariable'."
	hintReferenceTo       = "For example: 'a reference to myVariable'."
	hintToggle            = "For example: 'Toggle isRunning.' or 'Toggle the value of isActive.'"

	// Arrays and lookup tables.
	hintArrayLiteral         = "For example: 'an array of [1, 2, 3]' or 'an array of number [1, 2, 3]'."
	hintArrayCloseBracket    = "Make sure every '[' is matched by a closing ']'. For example: 'an array of [1, 2, 3]'."
	hintLookupEntryIn        = "For example: 'the entry \"name\" in myTable'."
	hintLookupSetEntry       = "For example: 'Set the entry \"name\" in myTable to be \"Alice\".'"
	hintLookupTableName      = "For example: 'Set the entry \"name\" in myTable to be \"Alice\".'"

	// Error handling.
	hintOnError      = "For example: 'on error:' to catch all errors, or 'on NetworkError:' to catch a specific type."
	hintRaiseAs      = "For example: 'raise \"Something went wrong\" as NetworkError.'"
	hintSwapVars     = "For example: 'swap a and b.' swaps the values of a and b."

	// Custom error type declarations.
	hintErrorTypeDecl    = "For example: 'Declare NetworkError as an error type.'"
	hintErrorSubtypeDecl = "For example: 'Declare TimeoutError as a type of NetworkError.'"

	// Structure declarations.
	hintStructName         = "For example: 'Declare Person as a structure with the following fields:'"
	hintFieldName          = "Field names must start with a letter. For example: 'name is a text.'"
	hintFieldType          = "Valid types are: text, number, boolean, integer. For example: 'name is a text.' or 'age is an integer.'"
	hintMethodName         = "Method names must start with a letter. For example: 'let greet be a function that does the following:'"
	hintMethodParam        = "Parameter names must start with a letter. For example: 'let add be a function that takes x and y and does the following:'"
	hintNewInstanceOf      = "For example: 'a new instance of Person' or 'new instance of Car'."
	hintNewInstanceFields  = "For example: 'a new instance of Person with the following fields:'"
	hintFieldAssignment    = "Each field must be set like: 'name is \"Alice\".' or 'age is 30.'"
	hintTypedVarName       = "For example: 'Declare score as number to be 0.' or 'Declare name as text.'"
	hintTypedVarType       = "Valid types are: number, text, boolean, integer. For example: 'Declare score as number to be 0.'"
)

// ─── Static messages ─────────────────────────────────────────────────────────
// These are the primary error descriptions when no runtime token value is
// needed to build the message.

const (
	msgVarNameExpected      = "I expected a variable name here."
	msgFunctionNameExpected = "I expected a function name here."
	msgParameterName        = "I expected a parameter name here."
	msgImportPath           = "The file path after 'Import' must be in quotes."
	msgDeclareVarName       = "I expected a variable name after 'Declare'."
	msgSetVarName           = "I expected a variable name after 'Set'."
	msgSetCallFuncName      = "I expected the name of a function to call here."
	msgSetListName          = "I expected the name of the list here."
	msgCallName             = "I expected a function or method name after 'Call'."
	msgForEachVar           = "I expected a loop variable name here."
	msgPrintOrWrite         = "I expected 'Print' or 'Write' here."
	msgAskVarAs             = "I expected a variable name after 'as' to store the answer."
	msgAskVarAnd            = "I expected a variable name to store the answer in."
	msgErrorTypeOnName      = "I expected an error type name or 'error' after 'on'."
	msgRaiseErrorType       = "I expected an error type name after 'as'."
	msgSwapFirstVar         = "I expected the first variable name after 'swap'."
	msgSwapSecondVar        = "I expected the second variable name after 'and'."
	msgErrorTypeName        = "I expected the name of the new error type."
	msgErrorSubtypeName     = "I expected the name of the error subtype."
	msgErrorParentType      = "I expected the parent error type name after 'of'."
	msgStructName           = "I expected the name of the structure after 'Declare'."
	msgFieldName            = "I expected the name of the field."
	msgMethodName           = "I expected the method name."
	msgMethodParam          = "I expected a parameter name."
	msgNewInstanceName      = "I expected the structure name after 'of'."
	msgNewInstanceField     = "I expected a field name here."
	msgTypedVarName         = "I expected a variable name after 'Declare'."
	msgLocationVar          = "I expected a variable name after 'the location of'."
	msgReferenceVar         = "I expected a variable name after 'reference to'."
	msgToggleVar            = "I expected a variable name after 'Toggle'."
	msgPossessive           = "I expected a method name after the possessive ('s)."
	msgErrorTypeIsName      = "I expected an error type name after 'is'."
	msgLookupTableName      = "I expected the table name after 'in'."
	msgArrayNeedsOf         = "I expected 'of' after 'array'."
	msgArrayNeedsCloseBrkt  = "I expected ']' to close the array, but reached the end of the file."
	msgStructMethodParam    = "I expected a parameter name."
)

// ─── Format-string messages ───────────────────────────────────────────────────
// Call fmt.Sprintf(msgFmt*, ...) at the call site to interpolate token values.

const (
	// "I expected 'be', '=', or 'always' after the variable name '<name>', but found '<tok>'."
	msgFmtLetAfterName = "I expected 'be', '=', or 'always' after the variable name '%s', but found '%s' instead."

	// "I do not understand 'Set the <tok>' here."
	msgFmtSetThe = "I do not understand 'Set the %s' here."

	// "I expected 'to' after the key in 'Set <name> at ...', but found '<tok>'."
	msgFmtSetAtTo = "I expected 'to' after the key in 'Set %s at ...', but found '%s'."

	// "I expected a method name after '<object>'s'."
	msgFmtCallPossessive = "I expected a method name after '%s's'."

	// "I expected the object name after 'Call <method> from/on'."
	msgFmtCallFromOn = "I expected the object name after 'Call %s from/on'."

	// "I expected 'the' or 'this' here, but found '<tok>'."
	msgFmtBreakTheThis = "I expected 'the' or 'this' here, but found '%s'."

	// "I expected 'as' or 'and' after the question text, but found '<tok>'."
	msgFmtAskAfter = "I expected 'as' or 'and' after the question text, but found '%s'."

	// "I expected a type name after 'cast to', but found '<tok>'."
	msgFmtCastTypeName = "I expected a type name after 'cast to', but found '%s'."

	// "I do not understand 'the <tok>' here."
	msgFmtTheUnknown = "I do not understand 'the %s' here."

	// "I do not know how to use '<tok>' as a value here."
	msgFmtExprUnknown = "I do not know how to use '%s' as a value here."

	// "I expected 'in' or 'of' after the index number, but found '<tok>'."
	msgFmtIndexAfter = "I expected 'in' or 'of' after the index number, but found '%s'."

	// "I expected 'divided by' after the first number, but found '<tok>'."
	msgFmtRemainderAfter = "I expected 'divided by' after the first number, but found '%s'."

	// "I expected '[' to open the array, but found '<tok>'."
	msgFmtArrayOpenBracket = "I expected '[' to open the array, but found '%s'."

	// "I expected 'in' after the key, but found '<tok>'."
	msgFmtLookupEntryIn = "I expected 'in' after the key, but found '%s'."

	// "I expected 'to' after the table name '<name>', but found '<tok>'."
	msgFmtLookupTableTo = "I expected 'to' after the table name '%s', but found '%s'."

	// "I expected 'structure' or 'struct' after 'as', but found '<tok>'."
	msgFmtStructOrStruct = "I expected 'structure' or 'struct' after 'as', but found '%s'."

	// "I expected 'fields' after 'following', but found '<tok>'."
	msgFmtFieldsAfter = "I expected 'fields' after 'following', but found '%s'."

	// "I expected a type name for field '<field>', but found '<tok>'."
	msgFmtFieldTypeName = "I expected a type name for field '%s', but found '%s'."

	// "I expected 'a' or 'an' after 'as', but found '<tok>'."
	msgFmtArticleAfterAs = "I expected 'a' or 'an' after 'as', but found '%s'."

	// "I expected the word 'error' here, but found '<tok>'."
	msgFmtExpectedErrorWord = "I expected the word 'error' here, but found '%s'."

	// "I expected a type name after 'as', but found '<tok>'."
	msgFmtTypedVarType = "I expected a type name after 'as', but found '%s'."

	// Statement-level messages that embed the token value.
	// "'<name>' cannot start a statement here."
	msgFmtIdentifierStatement = "'%s' cannot start a statement here."

	// "The number <n> cannot appear here on its own."
	msgFmtNumberStatement = "The number %s cannot appear here on its own."

	// "The text <s> cannot appear here on its own."
	msgFmtStringStatement = "The text %q cannot appear here on its own."

	// "I do not know what to do with '<tok>' here."
	msgFmtUnknownToken = "I do not know what to do with '%s' here."
)

// ─── Format-string hints ─────────────────────────────────────────────────────
// Hints that embed runtime token values.

const (
	// "To change a variable, use 'Set <name> to be <value>.' To show it, use 'Print <name>.'"
	hintFmtIdentifierStatement = "To change a variable, use 'Set %s to be <value>.' To show it, use 'Print %s.'"

	// "To declare a variable write: 'let <name> be <value>.' For a constant: 'let <name> always be <value>.'"
	hintFmtLetDeclaration = "To declare a variable write: 'let %s be <value>.' For a constant: 'let %s always be <value>.'"

	// "After 'array of <type>' I expected '[' to open the list of values."  (same as msgFmtArrayAfterType above)
	hintFmtArrayAfterType = "After 'array of %s' I expected '[' to open the list of values."
)
