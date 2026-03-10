package parser

import (
	"english/ast"
	"english/token"
	"fmt"
	"strconv"
	"strings"
)

// Magic string constants used in parsing
const (
	resultKeyword = "result"
)

// Parser transforms tokens into an AST
type Parser struct {
	tokens    []token.Token
	position  int
	curToken  token.Token
	peekToken token.Token
}

// NewParser creates a new parser for the given tokens
func NewParser(tokens []token.Token) *Parser {
	p := &Parser{tokens: tokens, position: 0}
	p.nextToken()
	p.nextToken()
	return p
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	if p.position < len(p.tokens) {
		p.peekToken = p.tokens[p.position]
		p.position++
	} else {
		p.peekToken = token.Token{Type: token.EOF}
	}
}

func (p *Parser) expectToken(tokenType token.Type) error {
	if p.curToken.Type != tokenType {
		return p.makeExpectError(tokenType)
	}
	return nil
}

func (p *Parser) makeExpectError(expected token.Type) error {
	var hint string

	// Provide helpful hints based on context
	switch expected {
	case token.PERIOD:
		hint = "Every statement must end with a period (.). Try adding a period at the end of the line."
	case token.TO:
		if p.curToken.Type == token.BE {
			hint = "Did you mean 'to be'? For example: 'Declare x to be 5.'"
		} else {
			hint = "The word 'to' is missing here. For example: 'Declare x to be 5.' or 'Set x to be 10.'"
		}
	case token.BE:
		if p.curToken.Type == token.TO {
			hint = "The word 'be' is missing after 'to'. For example: 'Declare x to be 5.'"
		} else {
			hint = "The word 'be' is expected here. For example: 'Declare x to be 5.'"
		}
	case token.THATS:
		hint = "Every block must end with 'thats it.' — did you forget to close the block?"
	case token.IT:
		if p.curToken.Type == token.PERIOD {
			hint = "The word 'it' is missing before the period. Write 'thats it.' to close the block."
		}
	case token.IDENTIFIER:
		if p.curToken.Type == token.NUMBER || p.curToken.Type == token.STRING {
			hint = "A name (like a variable name or function name) is expected here, not a number or text value."
		}
	}

	msg := fmt.Sprintf("I expected %s here but found %s instead.",
		tokenFriendlyName(expected), tokenFriendlyValue(p.curToken.Type, p.curToken.Value))
	return &SyntaxError{
		Msg:  msg,
		Line: p.curToken.Line,
		Col:  p.curToken.Col,
		Hint: hint,
	}
}

// Parse parses the tokens and returns the AST
func (p *Parser) Parse() (*ast.Program, error) {
	program := &ast.Program{}

	for p.curToken.Type != token.EOF {
		stmt, err := p.parseStatement()
		if err != nil {
			return nil, err
		}
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
	}

	return program, nil
}

func (p *Parser) parseStatement() (ast.Statement, error) {
	switch p.curToken.Type {
	case token.COMMENT:
		stmt := &ast.CommentStatement{Text: p.curToken.Value}
		p.nextToken()
		return stmt, nil
	case token.IMPORT:
		return p.parseImport()
	case token.DECLARE:
		return p.parseDeclaration()
	case token.LET:
		return p.parseLetDeclaration()
	case token.BREAK:
		return p.parseBreak()
	case token.CONTINUE, token.SKIP:
		return p.parseContinue()
	case token.ASK:
		return p.parseAskStatement()
	case token.SET:
		return p.parseAssignment()
	case token.CALL:
		return p.parseCall()
	case token.IF:
		return p.parseIfStatement()
	case token.REPEAT:
		return p.parseRepeat()
	case token.FOR:
		return p.parseForEach()
	case token.PRINT:
		return p.parseOutput(true)
	case token.WRITE:
		return p.parseOutput(false)
	case token.RETURN:
		return p.parseReturn()
	case token.TOGGLE:
		return p.parseToggle()
	case token.TRY:
		return p.parseTryStatement()
	case token.RAISE:
		return p.parseRaiseStatement()
	case token.SWAP:
		return p.parseSwapStatement()
	default:
		switch p.curToken.Type {
		case token.IDENTIFIER:
			name := p.curToken.Value
			return nil, &SyntaxError{
				Msg:  fmt.Sprintf("'%s' cannot start a statement here.", name),
				Line: p.curToken.Line,
				Col:  p.curToken.Col,
				Hint: fmt.Sprintf("To change a variable, use 'Set %s to be <value>.' To show it, use 'Print %s.'", name, name),
			}
		case token.NUMBER:
			num := p.curToken.Value
			return nil, &SyntaxError{
				Msg:  fmt.Sprintf("The number %s cannot appear here on its own.", num),
				Line: p.curToken.Line,
				Col:  p.curToken.Col,
				Hint: "Numbers must be part of a statement. For example: 'Declare x to be 5.' or 'Print 42.'",
			}
		case token.STRING:
			return nil, &SyntaxError{
				Msg:  fmt.Sprintf("The text %q cannot appear here on its own.", p.curToken.Value),
				Line: p.curToken.Line,
				Col:  p.curToken.Col,
				Hint: "Text must be part of a statement. For example: 'Print \"Hello, world!\".' or 'Declare greeting to be \"Hello\".'",
			}
		case token.EOF:
			return nil, &SyntaxError{
				Msg:  "The program ended unexpectedly.",
				Line: p.curToken.Line,
				Col:  p.curToken.Col,
				Hint: "Check that every block (like an 'If' or 'Repeat') is closed with 'thats it.'",
			}
		default:
			return nil, &SyntaxError{
				Msg:  fmt.Sprintf("I do not know what to do with '%s' here.", p.curToken.Value),
				Line: p.curToken.Line,
				Col:  p.curToken.Col,
				Hint: "Check the spelling of the keyword. Every statement must start with a word like 'Declare', 'Set', 'Print', 'If', 'Repeat', or 'Call'.",
			}
		}
	}
}

// parseLetDeclaration parses various "let" syntax forms:
// - let x be 10.
// - let x be equal to 10.
// - let x always be 10.
// - let x be always 10.
// - let x = 10.
// - let x equal 10.
func (p *Parser) parseLetDeclaration() (ast.Statement, error) {
	if err := p.expectToken(token.LET); err != nil {
		return nil, err
	}
	p.nextToken()

	// Get variable name
	nameToken := p.curToken
	if p.curToken.Type != token.IDENTIFIER {
		return nil, p.syntaxErr(
			"I expected a variable name here.",
			"Variable names must start with a letter. For example: 'let count be 10.' or 'let name be \"Alice\".'",
		)
	}
	p.nextToken()

	isConstant := false

	// Handle different syntax forms
	switch p.curToken.Type {
	case token.ASSIGN:
		// let x = 10.
		p.nextToken()
	case token.EQUAL:
		// let x equal 10.
		p.nextToken()
	case token.ALWAYS:
		// let x always be 10.
		isConstant = true
		p.nextToken()
		if err := p.expectToken(token.BE); err != nil {
			return nil, err
		}
		p.nextToken()
	case token.BE:
		// let x be 10. OR let x be equal to 10. OR let x be always 10.
		p.nextToken()
		if p.curToken.Type == token.ALWAYS {
			isConstant = true
			p.nextToken()
		} else if p.curToken.Type == token.EQUAL {
			// let x be equal to 10.
			p.nextToken()
			if err := p.expectToken(token.TO); err != nil {
				return nil, err
			}
			p.nextToken()
		}
	default:
		return nil, p.syntaxErr(
			fmt.Sprintf("I expected 'be', '=', or 'always' after the variable name '%s', but found '%s' instead.", nameToken.Value, p.curToken.Value),
			"To declare a variable write: 'let "+nameToken.Value+" be <value>.' For a constant: 'let "+nameToken.Value+" always be <value>.'",
		)
	}

	value, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	if err := p.expectToken(token.PERIOD); err != nil {
		return nil, err
	}
	p.nextToken()

	return &ast.VariableDecl{
		Name:       nameToken.Value,
		IsConstant: isConstant,
		Value:      value,
		Line:       nameToken.Line,
	}, nil
}

// parseImport parses import statements with natural English syntax:
// - Import "file.abc".
// - Import from "file.abc".
// - Import func1, func2 and func3 from "file.abc".
// - Import everything from "file.abc".
// - Import all from "file.abc".
// - Import all from "file.abc" safely.
func (p *Parser) parseImport() (ast.Statement, error) {
	if err := p.expectToken(token.IMPORT); err != nil {
		return nil, err
	}
	p.nextToken()

	var items []string
	var importAll bool
	var isSafe bool

	// Handle optional "the" keyword
	if p.curToken.Type == token.THE {
		p.nextToken()
	}

	// Check for "everything" or "all"
	if p.curToken.Type == token.EVERYTHING || p.curToken.Type == token.ALL {
		importAll = true
		p.nextToken()
	} else if p.curToken.Type == token.IDENTIFIER {
		// Parse list of items to import
		// Support: func1, func2 and func3
		for {
			items = append(items, p.curToken.Value)
			p.nextToken()

			// Check for comma or "and"
			if p.curToken.Type == token.COMMA {
				p.nextToken()
				// Optional "and" after comma
				if p.curToken.Type == token.AND {
					p.nextToken()
				}
			} else if p.curToken.Type == token.AND {
				p.nextToken()
			} else {
				// No more items
				break
			}

			// Expect another identifier
			if p.curToken.Type != token.IDENTIFIER {
				break
			}
		}
	}

	// Handle optional "from" keyword
	if p.curToken.Type == token.FROM {
		p.nextToken()
	}

	// Expect a string with the file path
	if p.curToken.Type != token.STRING {
		return nil, p.syntaxErr(
			"The file path after 'Import' must be in quotes.",
			"For example: 'Import \"myfile.abc\".' or 'Import everything from \"utils.abc\".'",
		)
	}

	filePath := p.curToken.Value
	p.nextToken()

	// Check for "safely" keyword
	if p.curToken.Type == token.SAFELY {
		isSafe = true
		p.nextToken()
	}

	// Expect period to end the statement
	if err := p.expectToken(token.PERIOD); err != nil {
		return nil, err
	}
	p.nextToken()

	return &ast.ImportStatement{
		Path:      filePath,
		Items:     items,
		ImportAll: importAll,
		IsSafe:    isSafe,
	}, nil
}

func (p *Parser) parseDeclaration() (ast.Statement, error) {
	if err := p.expectToken(token.DECLARE); err != nil {
		return nil, err
	}
	p.nextToken()

	// Check if it's a function declaration
	if p.curToken.Type == token.FUNCTION {
		return p.parseFunctionDeclaration()
	}

	// Check if it's a declaration with "as": "Declare X as ..."
	// This covers structs, typed variables, and custom error types.
	if p.curToken.Type == token.IDENTIFIER && p.peekToken.Type == token.AS {
		return p.parseDeclareAs()
	}

	// Variable or constant declaration
	nameToken := p.curToken
	if p.curToken.Type != token.IDENTIFIER {
		return nil, p.syntaxErr(
			"I expected a variable name after 'Declare'.",
			"Variable names must start with a letter. For example: 'Declare score to be 0.' or 'Declare name to always be \"Alice\".'",
		)
	}
	p.nextToken()

	if err := p.expectToken(token.TO); err != nil {
		return nil, err
	}
	p.nextToken()

	// Check for "always" keyword (can appear before or after "be")
	isConstant := false
	if p.curToken.Type == token.ALWAYS {
		isConstant = true
		p.nextToken()
	}

	if err := p.expectToken(token.BE); err != nil {
		return nil, err
	}
	p.nextToken()

	// Check for "always" after "be" if not seen before
	if !isConstant && p.curToken.Type == token.ALWAYS {
		isConstant = true
		p.nextToken()
	}

	value, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	if err := p.expectToken(token.PERIOD); err != nil {
		return nil, err
	}
	p.nextToken()

	return &ast.VariableDecl{
		Name:       nameToken.Value,
		IsConstant: isConstant,
		Value:      value,
		Line:       nameToken.Line,
	}, nil
}

// parseDeclareAs dispatches "Declare X as ..." to the correct parser:
//   - "Declare X as a structure ..."       → struct declaration
//   - "Declare X as an error type."        → custom error type declaration
//   - "Declare X as a type of Y."          → error subtype declaration
//   - "Declare X as <typename> to be ..."  → typed variable declaration
func (p *Parser) parseDeclareAs() (ast.Statement, error) {
	// curToken is IDENTIFIER (name), peekToken is AS.
	// Look at the token two positions ahead (after AS) to decide.
	// p.position currently points to the token after peekToken (i.e., after AS).
	tokAfterAs := p.tokenAt(p.position)
	tok2AfterAs := p.tokenAt(p.position + 1)

	isArticle := tokAfterAs.Type == token.IDENTIFIER &&
		(strings.ToLower(tokAfterAs.Value) == "a" || strings.ToLower(tokAfterAs.Value) == "an")

	if isArticle {
		// "Declare X as a/an ..."
		switch {
		case tok2AfterAs.Type == token.STRUCTURE || tok2AfterAs.Type == token.STRUCT:
			return p.parseStructDeclaration()
		case strings.ToLower(tok2AfterAs.Value) == "error":
			// Might be "Declare X as an error type."
			tok3AfterAs := p.tokenAt(p.position + 2)
			if tok3AfterAs.Type == token.TYPE {
				return p.parseErrorTypeDecl()
			}
		case tok2AfterAs.Type == token.TYPE:
			// Might be "Declare X as a type of Y." — error subtype declaration
			tok3AfterAs := p.tokenAt(p.position + 2)
			if tok3AfterAs.Type == token.OF {
				return p.parseErrorSubtypeDecl()
			}
		}
	}

	// Fall through: typed variable declaration "Declare X as typename to be value."
	return p.parseTypedVariableDecl()
}

// tokenAt safely returns the token at the given absolute position in p.tokens.
func (p *Parser) tokenAt(pos int) token.Token {
	if pos >= 0 && pos < len(p.tokens) {
		return p.tokens[pos]
	}
	return token.Token{Type: token.EOF}
}

func (p *Parser) parseFunctionDeclaration() (ast.Statement, error) {
	if err := p.expectToken(token.FUNCTION); err != nil {
		return nil, err
	}
	p.nextToken()

	nameToken := p.curToken
	if p.curToken.Type != token.IDENTIFIER {
		return nil, p.syntaxErr(
			"I expected a function name here.",
			"Function names must start with a letter. For example: 'Declare function greet that does the following:'",
		)
	}
	p.nextToken()

	var parameters []string

	// Skip optional "that" before "takes" or "does"
	if p.curToken.Type == token.THAT {
		p.nextToken()
	}

	if p.curToken.Type == token.TAKES {
		p.nextToken()
		for {
			paramToken := p.curToken
			if p.curToken.Type != token.IDENTIFIER {
				return nil, p.syntaxErr(
					"I expected a parameter name here.",
					"Parameter names must start with a letter. For example: 'Declare function add that takes x and y and does the following:'",
				)
			}
			parameters = append(parameters, paramToken.Value)
			p.nextToken()

			if p.curToken.Type != token.AND {
				break
			}
			// Check if "and" is followed by "does" (end of params) or another param
			if p.peekToken.Type == token.DOES {
				break
			}
			p.nextToken()
		}
	}

	// Support "and does" syntax after parameters
	if p.curToken.Type == token.AND {
		p.nextToken()
	}

	if err := p.expectToken(token.DOES); err != nil {
		return nil, err
	}
	p.nextToken()

	if err := p.expectToken(token.THE); err != nil {
		return nil, err
	}
	p.nextToken()

	if err := p.expectToken(token.FOLLOWING); err != nil {
		return nil, err
	}
	p.nextToken()

	if err := p.expectToken(token.COLON); err != nil {
		return nil, err
	}
	p.nextToken()

	body, err := p.parseBlock()
	if err != nil {
		return nil, err
	}

	if p.curToken.Type == token.THATS {
		p.nextToken()
		if err := p.expectToken(token.IT); err != nil {
			return nil, err
		}
		p.nextToken()
		if err := p.expectToken(token.PERIOD); err != nil {
			return nil, err
		}
		p.nextToken()
	}

	return &ast.FunctionDecl{
		Name:       nameToken.Value,
		Parameters: parameters,
		Body:       body,
	}, nil
}

func (p *Parser) parseAssignment() (ast.Statement, error) {
	if err := p.expectToken(token.SET); err != nil {
		return nil, err
	}
	setLine := p.curToken.Line
	p.nextToken()

	// Check for "Set the item at position X in Y to be Z"
	// or "Set the entry KEY in TABLE to be VALUE"
	if p.curToken.Type == token.THE {
		p.nextToken()
		if p.curToken.Type == token.ITEM {
			return p.parseIndexAssignment(setLine)
		}
		if p.curToken.Type == token.ENTRY {
			return p.parseLookupKeyAssignment(setLine)
		}
		return nil, p.syntaxErr(
			fmt.Sprintf("I do not understand 'Set the %s' here.", p.curToken.Value),
			"After 'Set the' you can write 'item at position N in myList to be VALUE' or 'entry KEY in myTable to be VALUE'.",
		)
	}

	nameToken := p.curToken
	if p.curToken.Type != token.IDENTIFIER {
		return nil, p.syntaxErr(
			"I expected a variable name after 'Set'.",
			"For example: 'Set score to be 10.' or 'Set name to be \"Alice\".'",
		)
	}
	p.nextToken()

	// "Set TABLE at KEY to be VALUE." — lookup table shorthand write
	if p.curToken.Type == token.AT {
		p.nextToken() // consume AT
		key, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		if p.curToken.Type != token.TO {
			return nil, p.syntaxErr(
				fmt.Sprintf("I expected 'to' after the key in 'Set %s at ...', but found '%s'.", nameToken.Value, p.curToken.Value),
				"The full form is: 'Set tableName at key to be value.'",
			)
		}
		p.nextToken()
		if p.curToken.Type == token.BE {
			p.nextToken()
		}
		value, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		if err := p.expectToken(token.PERIOD); err != nil {
			return nil, err
		}
		p.nextToken()
		return &ast.LookupKeyAssignment{TableName: nameToken.Value, Key: key, Value: value, Line: setLine}, nil
	}

	if err := p.expectToken(token.TO); err != nil {
		return nil, err
	}
	p.nextToken()

	// "be" is optional - "set x to 10" and "set x to be 10" are both valid
	if p.curToken.Type == token.BE {
		p.nextToken()
	}

	// Check for function call result: "the result of calling ..."
	if p.curToken.Type == token.THE && p.peekToken.Type == token.IDENTIFIER && strings.EqualFold(p.peekToken.Value, resultKeyword) {
		p.nextToken() // consume THE
		p.nextToken() // consume "result"
		if p.curToken.Type == token.OF {
			p.nextToken()
			if p.curToken.Type == token.CALLING {
				p.nextToken()
				funcName := p.curToken.Value
				if p.curToken.Type != token.IDENTIFIER {
					return nil, p.syntaxErr(
						"I expected the name of a function to call here.",
						"For example: 'Set result to be the result of calling square of x.'",
					)
				}
				p.nextToken()

				args, err := p.parseFunctionArguments()
				if err != nil {
					return nil, err
				}

				if err := p.expectToken(token.PERIOD); err != nil {
					return nil, err
				}
				p.nextToken()

				return &ast.Assignment{
					Name: nameToken.Value,
					Value: &ast.FunctionCall{
						Name:      funcName,
						Arguments: args,
					},
					Line: setLine,
				}, nil
			}
		}
	}

	value, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	if err := p.expectToken(token.PERIOD); err != nil {
		return nil, err
	}
	p.nextToken()

	return &ast.Assignment{
		Name:  nameToken.Value,
		Value: value,
		Line:  setLine,
	}, nil
}

// parseIndexAssignment parses "the item at position X in Y to be Z"
func (p *Parser) parseIndexAssignment(setLine int) (ast.Statement, error) {
	// Already consumed "Set the", now at "item"
	if err := p.expectToken(token.ITEM); err != nil {
		return nil, err
	}
	p.nextToken()

	if err := p.expectToken(token.AT); err != nil {
		return nil, err
	}
	p.nextToken()

	if err := p.expectToken(token.POSITION); err != nil {
		return nil, err
	}
	p.nextToken()

	index, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	if err := p.expectToken(token.IN); err != nil {
		return nil, err
	}
	p.nextToken()

	listName := p.curToken.Value
	if p.curToken.Type != token.IDENTIFIER {
		return nil, p.syntaxErr(
			"I expected the name of the list here.",
			"For example: 'Set the item at position 1 in myList to be 99.'",
		)
	}
	p.nextToken()

	if err := p.expectToken(token.TO); err != nil {
		return nil, err
	}
	p.nextToken()

	// "be" is optional - "set item to 10" and "set item to be 10" are both valid
	if p.curToken.Type == token.BE {
		p.nextToken()
	}

	value, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	if err := p.expectToken(token.PERIOD); err != nil {
		return nil, err
	}
	p.nextToken()

	return &ast.IndexAssignment{
		ListName: listName,
		Index:    index,
		Value:    value,
		Line:     setLine,
	}, nil
}

func (p *Parser) parseCall() (ast.Statement, error) {
	if err := p.expectToken(token.CALL); err != nil {
		return nil, err
	}
	callLine := p.curToken.Line
	p.nextToken()

	// First identifier could be:
	// 1. Function name: "call greet with args."
	// 2. Method name: "call talk from p2." or "call talk on p2."
	// 3. Object name with possessive: "call p2's talk." (p2's is a single token)

	firstIdent := p.curToken.Value
	if p.curToken.Type != token.IDENTIFIER {
		return nil, p.syntaxErr(
			"I expected a function or method name after 'Call'.",
			"For example: 'Call greet.' or 'Call add with 3 and 5.'",
		)
	}
	p.nextToken()

	// Check for possessive syntax: "call p2's talk."
	// The identifier will end with 's (e.g., "p2's")
	if len(firstIdent) > 2 && firstIdent[len(firstIdent)-2:] == "'s" {
		// This is possessive: extract object name (remove 's)
		objectName := firstIdent[:len(firstIdent)-2]

		if p.curToken.Type != token.IDENTIFIER {
			return nil, p.syntaxErr(
				fmt.Sprintf("I expected a method name after '%s's'.", objectName),
				"For example: 'Call person's greet.' or 'Call person's setName with \"Alice\".'",
			)
		}
		methodName := p.curToken.Value
		p.nextToken()

		// Parse optional arguments
		var args []ast.Expression
		if p.curToken.Type == token.WITH {
			p.nextToken()
			args = p.parseCallArguments()
		}

		if err := p.expectToken(token.PERIOD); err != nil {
			return nil, err
		}
		p.nextToken()

		// Return as method call
		return &ast.CallStatement{
			MethodCall: &ast.MethodCall{
				Object:     &ast.Identifier{Name: objectName},
				MethodName: methodName,
				Arguments:  args,
			},
			Line: callLine,
		}, nil
	}

	// Check for "from" or "on" (method call syntax)
	if p.curToken.Type == token.FROM || p.curToken.Type == token.ON {
		methodName := firstIdent
		p.nextToken() // skip FROM/ON

		// Get object
		if p.curToken.Type != token.IDENTIFIER {
			return nil, p.syntaxErr(
				fmt.Sprintf("I expected the object name after 'Call %s from/on'.", methodName),
				"For example: 'Call talk from person.' or 'Call greet on employee.'",
			)
		}
		objectName := p.curToken.Value
		p.nextToken()

		// Parse optional arguments
		var args []ast.Expression
		if p.curToken.Type == token.WITH {
			p.nextToken()
			args = p.parseCallArguments()
		}

		if err := p.expectToken(token.PERIOD); err != nil {
			return nil, err
		}
		p.nextToken()

		// Return as method call
		return &ast.CallStatement{
			MethodCall: &ast.MethodCall{
				Object:     &ast.Identifier{Name: objectName},
				MethodName: methodName,
				Arguments:  args,
			},
			Line: callLine,
		}, nil
	}

	// Regular function call: "call greet with args."
	funcName := firstIdent
	var args []ast.Expression

	if p.curToken.Type == token.WITH {
		p.nextToken()
		args = p.parseCallArguments()
	}

	if err := p.expectToken(token.PERIOD); err != nil {
		return nil, err
	}
	p.nextToken()

	return &ast.CallStatement{
		FunctionCall: &ast.FunctionCall{
			Name:      funcName,
			Arguments: args,
		},
		Line: callLine,
	}, nil
}

// parseCallArguments parses comma-separated call arguments
func (p *Parser) parseCallArguments() []ast.Expression {
	var args []ast.Expression

	for {
		arg, err := p.parseExpression()
		if err != nil {
			break
		}
		args = append(args, arg)

		if p.curToken.Type != token.AND && p.curToken.Type != token.COMMA {
			break
		}
		p.nextToken()
	}

	return args
}

func (p *Parser) parseIfStatement() (ast.Statement, error) {
	if err := p.expectToken(token.IF); err != nil {
		return nil, err
	}
	startLine := p.curToken.Line
	p.nextToken()

	condition, err := p.parseComparison()
	if err != nil {
		return nil, err
	}

	if err := p.expectToken(token.COMMA); err != nil {
		return nil, err
	}
	p.nextToken()

	if err := p.expectToken(token.THEN); err != nil {
		return nil, err
	}
	p.nextToken()

	thenBody, err := p.parseBlock()
	if err != nil {
		return nil, err
	}

	var elseIfParts []*ast.ElseIfPart
	var elseBody []ast.Statement

	for p.curToken.Type == token.OTHERWISE {
		p.nextToken()
		if p.curToken.Type == token.IF {
			p.nextToken()
			eifCond, err := p.parseComparison()
			if err != nil {
				return nil, err
			}
			if err := p.expectToken(token.COMMA); err != nil {
				return nil, err
			}
			p.nextToken()
			if err := p.expectToken(token.THEN); err != nil {
				return nil, err
			}
			p.nextToken()
			eifBody, err := p.parseBlock()
			if err != nil {
				return nil, err
			}
			elseIfParts = append(elseIfParts, &ast.ElseIfPart{
				Condition: eifCond,
				Body:      eifBody,
			})
		} else {
			elseBody, err = p.parseBlock()
			if err != nil {
				return nil, err
			}
			break
		}
	}

	if p.curToken.Type == token.THATS {
		p.nextToken()
		if err := p.expectToken(token.IT); err != nil {
			return nil, err
		}
		p.nextToken()
		if err := p.expectToken(token.PERIOD); err != nil {
			return nil, err
		}
		p.nextToken()
	}

	return &ast.IfStatement{
		Condition: condition,
		Then:      thenBody,
		ElseIf:    elseIfParts,
		Else:      elseBody,
		Line:      startLine,
	}, nil
}

func (p *Parser) parseRepeat() (ast.Statement, error) {
	if err := p.expectToken(token.REPEAT); err != nil {
		return nil, err
	}
	startLine := p.curToken.Line
	p.nextToken()

	// Check for "repeat forever" syntax
	if p.curToken.Type == token.FOREVER {
		p.nextToken()

		if err := p.expectToken(token.COLON); err != nil {
			return nil, err
		}
		p.nextToken()

		body, err := p.parseBlock()
		if err != nil {
			return nil, err
		}

		if p.curToken.Type == token.THATS {
			p.nextToken()
			if err := p.expectToken(token.IT); err != nil {
				return nil, err
			}
			p.nextToken()
			if err := p.expectToken(token.PERIOD); err != nil {
				return nil, err
			}
			p.nextToken()
		}

		return &ast.WhileLoop{
			Condition: &ast.BooleanLiteral{Value: true},
			Body:      body,
			Line:      startLine,
		}, nil
	}

	if err := p.expectToken(token.THE); err != nil {
		return nil, err
	}
	p.nextToken()

	if err := p.expectToken(token.FOLLOWING); err != nil {
		return nil, err
	}
	p.nextToken()

	// Check if it's a while loop or for loop
	if p.curToken.Type == token.WHILE {
		p.nextToken()
		condition, err := p.parseComparison()
		if err != nil {
			return nil, err
		}

		if err := p.expectToken(token.COLON); err != nil {
			return nil, err
		}
		p.nextToken()

		body, err := p.parseBlock()
		if err != nil {
			return nil, err
		}

		if p.curToken.Type == token.THATS {
			p.nextToken()
			if err := p.expectToken(token.IT); err != nil {
				return nil, err
			}
			p.nextToken()
			if err := p.expectToken(token.PERIOD); err != nil {
				return nil, err
			}
			p.nextToken()
		}

		return &ast.WhileLoop{
			Condition: condition,
			Body:      body,
			Line:      startLine,
		}, nil
	}

	// For loop (N times)
	countExpr, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	if err := p.expectToken(token.TIMES); err != nil {
		return nil, err
	}
	p.nextToken()

	if err := p.expectToken(token.COLON); err != nil {
		return nil, err
	}
	p.nextToken()

	body, err := p.parseBlock()
	if err != nil {
		return nil, err
	}

	if p.curToken.Type == token.THATS {
		p.nextToken()
		if err := p.expectToken(token.IT); err != nil {
			return nil, err
		}
		p.nextToken()
		if err := p.expectToken(token.PERIOD); err != nil {
			return nil, err
		}
		p.nextToken()
	}

	return &ast.ForLoop{
		Count: countExpr,
		Body:  body,
		Line:  startLine,
	}, nil
}

func (p *Parser) parseForEach() (ast.Statement, error) {
	if err := p.expectToken(token.FOR); err != nil {
		return nil, err
	}
	p.nextToken()

	if err := p.expectToken(token.EACH); err != nil {
		return nil, err
	}
	p.nextToken()

	itemToken := p.curToken
	// Allow both IDENTIFIER and ITEM keyword as the loop variable name
	if p.curToken.Type != token.IDENTIFIER && p.curToken.Type != token.ITEM {
		return nil, p.syntaxErr(
			"I expected a loop variable name here.",
			"For example: 'For each item in myList:' or 'For each number in scores:'",
		)
	}
	// Get the value, treating token.ITEM as "item" string
	itemName := itemToken.Value
	if itemToken.Type == token.ITEM {
		itemName = "item"
	}
	p.nextToken()

	if err := p.expectToken(token.IN); err != nil {
		return nil, err
	}
	p.nextToken()

	listExpr, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	if err := p.expectToken(token.COMMA); err != nil {
		return nil, err
	}
	p.nextToken()

	if err := p.expectToken(token.DO); err != nil {
		return nil, err
	}
	p.nextToken()

	if err := p.expectToken(token.THE); err != nil {
		return nil, err
	}
	p.nextToken()

	if err := p.expectToken(token.FOLLOWING); err != nil {
		return nil, err
	}
	p.nextToken()

	if err := p.expectToken(token.COLON); err != nil {
		return nil, err
	}
	p.nextToken()

	body, err := p.parseBlock()
	if err != nil {
		return nil, err
	}

	if p.curToken.Type == token.THATS {
		p.nextToken()
		if err := p.expectToken(token.IT); err != nil {
			return nil, err
		}
		p.nextToken()
		if err := p.expectToken(token.PERIOD); err != nil {
			return nil, err
		}
		p.nextToken()
	}

	return &ast.ForEachLoop{
		Item: itemName,
		List: listExpr,
		Body: body,
		Line: itemToken.Line,
	}, nil
}

func (p *Parser) parseOutput(newline bool) (ast.Statement, error) {
	// Accept either PRINT or WRITE token
	if p.curToken.Type != token.PRINT && p.curToken.Type != token.WRITE {
		return nil, p.syntaxErr(
			"I expected 'Print' or 'Write' here.",
			"To show output, use 'Print \"Hello\".' or 'Write \"No newline\".'",
		)
	}
	startLine := p.curToken.Line
	p.nextToken()

	var values []ast.Expression

	// Parse first expression
	value, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	values = append(values, value)

	// Parse additional comma-separated expressions
	for p.curToken.Type == token.COMMA {
		p.nextToken() // consume comma
		value, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		values = append(values, value)
	}

	if err := p.expectToken(token.PERIOD); err != nil {
		return nil, err
	}
	p.nextToken()

	return &ast.OutputStatement{
		Values:  values,
		Newline: newline,
		Line:    startLine,
	}, nil
}

func (p *Parser) parseReturn() (ast.Statement, error) {
	if err := p.expectToken(token.RETURN); err != nil {
		return nil, err
	}
	startLine := p.curToken.Line
	p.nextToken()

	value, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	if err := p.expectToken(token.PERIOD); err != nil {
		return nil, err
	}
	p.nextToken()

	return &ast.ReturnStatement{
		Value: value,
		Line:  startLine,
	}, nil
}

func (p *Parser) parseBreak() (ast.Statement, error) {
	if err := p.expectToken(token.BREAK); err != nil {
		return nil, err
	}
	p.nextToken()

	if err := p.expectToken(token.OUT); err != nil {
		return nil, err
	}
	p.nextToken()

	if err := p.expectToken(token.OF); err != nil {
		return nil, err
	}
	p.nextToken()

	// Accept "the" or "this" (as IDENTIFIER)
	if p.curToken.Type == token.THE {
		p.nextToken()
	} else if p.curToken.Type == token.IDENTIFIER && strings.EqualFold(p.curToken.Value, "this") {
		p.nextToken()
	} else {
		return nil, p.syntaxErr(
			fmt.Sprintf("I expected 'the' or 'this' here, but found '%s'.", p.curToken.Value),
			"For example: 'Break out of the loop.' or 'Break out of this loop.'",
		)
	}

	if err := p.expectToken(token.LOOP); err != nil {
		return nil, err
	}
	p.nextToken()

	if err := p.expectToken(token.PERIOD); err != nil {
		return nil, err
	}
	p.nextToken()

	return &ast.BreakStatement{}, nil
}

// parseContinue parses a continue statement:
//   - "Continue." or "Skip."
//   - "Continue the loop." or "Skip the loop."
func (p *Parser) parseContinue() (ast.Statement, error) {
	p.nextToken() // consume CONTINUE or SKIP

	// Optional "the loop"
	if p.curToken.Type == token.THE {
		p.nextToken() // consume THE
		if p.curToken.Type == token.LOOP {
			p.nextToken() // consume LOOP
		}
	}

	if err := p.expectToken(token.PERIOD); err != nil {
		return nil, err
	}
	p.nextToken()

	return &ast.ContinueStatement{}, nil
}

// parseAskStatement parses an ask statement for user input:
//   - "Ask "prompt" as varname."   (create or set variable)
//   - "Ask "prompt" and store it in varname."
func (p *Parser) parseAskStatement() (ast.Statement, error) {
	p.nextToken() // consume ASK

	// Parse the prompt expression
	prompt, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	// Determine target variable name
	var varName string

	if p.curToken.Type == token.AS {
		// "Ask "prompt" as varname."
		p.nextToken() // consume AS
		if p.curToken.Type != token.IDENTIFIER {
			return nil, p.syntaxErr(
				"I expected a variable name after 'as' to store the answer.",
				"For example: 'Ask \"What is your name?\" as userName.'",
			)
		}
		varName = p.curToken.Value
		p.nextToken()
	} else if p.curToken.Type == token.AND {
		// "Ask "prompt" and store it in varname." or "Ask "prompt" and save it in varname."
		p.nextToken() // consume AND
		// Skip "store"/"save"/"put" identifier if present
		if p.curToken.Type == token.IDENTIFIER {
			p.nextToken()
		}
		// Skip "it", "the", "answer", "result", "response" identifiers/tokens
		for p.curToken.Type == token.IDENTIFIER || p.curToken.Type == token.THE {
			if p.curToken.Type == token.IN {
				break
			}
			p.nextToken()
		}
		// Expect "in"
		if p.curToken.Type == token.IN {
			p.nextToken() // consume IN
		}
		if p.curToken.Type != token.IDENTIFIER {
			return nil, p.syntaxErr(
				"I expected a variable name to store the answer in.",
				"For example: 'Ask \"What is your age?\" and store it in userAge.'",
			)
		}
		varName = p.curToken.Value
		p.nextToken()
	} else {
		return nil, p.syntaxErr(
			fmt.Sprintf("I expected 'as' or 'and' after the question text, but found '%s'.", p.curToken.Value),
			"For example: 'Ask \"What is your name?\" as myName.' or 'Ask \"Enter a number:\" and store it in num.'",
		)
	}

	if err := p.expectToken(token.PERIOD); err != nil {
		return nil, err
	}
	p.nextToken()

	// Use Assignment so it works whether the variable already exists or not.
	// The Assignment evaluator calls Set(), which creates the variable if it doesn't exist.
	return &ast.Assignment{
		Name:  varName,
		Value: &ast.AskExpression{Prompt: prompt},
	}, nil
}

func (p *Parser) parseBlock() ([]ast.Statement, error) {
	var statements []ast.Statement

	for p.curToken.Type != token.THATS &&
		p.curToken.Type != token.OTHERWISE &&
		p.curToken.Type != token.ON &&
		p.curToken.Type != token.BUT &&
		p.curToken.Type != token.EOF {
		stmt, err := p.parseStatement()
		if err != nil {
			return nil, err
		}
		if stmt != nil {
			statements = append(statements, stmt)
		}
	}

	return statements, nil
}

func (p *Parser) parseComparison() (ast.Expression, error) {
	return p.parseLogical()
}

// parseLogical handles logical "and" / "or" operators (lowest precedence above comparison)
func (p *Parser) parseLogical() (ast.Expression, error) {
	left, err := p.parseRelational()
	if err != nil {
		return nil, err
	}

	for p.curToken.Type == token.AND || p.curToken.Type == token.OR {
		op := "and"
		if p.curToken.Type == token.OR {
			op = "or"
		}
		p.nextToken()
		right, err := p.parseRelational()
		if err != nil {
			return nil, err
		}
		left = &ast.BinaryExpression{
			Left:     left,
			Operator: op,
			Right:    right,
		}
	}

	return left, nil
}

// parseRelational handles comparison operators like "is equal to", "is less than", etc.
func (p *Parser) parseRelational() (ast.Expression, error) {
	left, err := p.parseCast()
	if err != nil {
		return nil, err
	}

	switch p.curToken.Type {
	case token.IS_EQUAL_TO, token.IS_LESS_THAN, token.IS_GREATER_THAN,
		token.IS_LESS_EQUAL, token.IS_GREATER_EQUAL, token.IS_NOT_EQUAL:
		op := p.curToken.Value
		p.nextToken()
		right, err := p.parseCast()
		if err != nil {
			return nil, err
		}
		return &ast.BinaryExpression{
			Left:     left,
			Operator: op,
			Right:    right,
		}, nil

	case token.IS_SOMETHING:
		// "x is something" / "x has a value" — postfix nil check (not nil)
		p.nextToken()
		return &ast.NilCheckExpression{Value: left, IsSomethingCheck: true}, nil

	case token.IS_NOTHING_OP:
		// "x is nothing" / "x has no value" — postfix nil check (is nil)
		p.nextToken()
		return &ast.NilCheckExpression{Value: left, IsSomethingCheck: false}, nil

	case token.IS:
		// "error is TypeName" — error type check (exact or inherited match)
		p.nextToken()
		if p.curToken.Type != token.IDENTIFIER {
			return nil, p.syntaxErr(
				"I expected an error type name after 'is'.",
				"For example: 'error is NetworkError' or 'error is RuntimeError'.",
			)
		}
		typeName := p.curToken.Value
		p.nextToken()
		return &ast.ErrorTypeCheckExpression{Value: left, TypeName: typeName}, nil
	}

	return left, nil
}

func (p *Parser) parseExpression() (ast.Expression, error) {
	return p.parseCast()
}

// parseCast handles postfix operators on any expression:
//   - "cast to <type>" / "casted to <type>" — explicit type conversion
//   - "has <key>"                            — lookup table key check
func (p *Parser) parseCast() (ast.Expression, error) {
	expr, err := p.parseAdditive()
	if err != nil {
		return nil, err
	}

	// Postfix "cast to <type>" or "casted to <type>"
	if p.curToken.Type == token.CASTED {
		p.nextToken() // consume "cast"/"casted"
		if p.curToken.Type == token.TO {
			p.nextToken()
		}
		typeName := p.parseTypeName()
		if typeName == "" {
			return nil, p.syntaxErr(
				fmt.Sprintf("I expected a type name after 'cast to', but found '%s'.", p.curToken.Value),
				"Valid type names are: number, text, boolean, integer. For example: 'x cast to number'.",
			)
		}
		return &ast.CastExpression{Value: expr, TypeName: typeName}, nil
	}

	// Postfix "has <key>" — lookup table membership test
	if p.curToken.Type == token.HAS {
		p.nextToken() // consume HAS
		key, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		return &ast.HasExpression{Table: expr, Key: key}, nil
	}

	// Postfix "at <key>" — lookup table / array access
	if p.curToken.Type == token.AT {
		// Peek: if followed by POSITION it is the existing list index expression
		// (handled in parsePrimary when THE·ITEM·AT·POSITION is already consumed).
		// Here we only handle the new "identifier at key" shorthand.
		p.nextToken() // consume AT
		key, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		return &ast.LookupKeyAccess{Table: expr, Key: key}, nil
	}

	// Postfix possessive: expr's method
	// Handles string literals and other non-identifier expressions:
	//   "hello"'s title   →   MethodCall{Object: "hello", MethodName: "title"}
	if p.curToken.Type == token.POSSESSIVE {
		p.nextToken() // consume 's
		if p.curToken.Type != token.IDENTIFIER {
			return nil, p.syntaxErr(
				"I expected a method name after the possessive ('s).",
				"For example: 'myText's length' or 'myText's upper'.",
			)
		}
		methodName := p.curToken.Value
		p.nextToken()
		var args []ast.Expression
		if p.curToken.Type == token.WITH {
			p.nextToken()
			args = p.parseCallArguments()
		}
		return &ast.MethodCall{Object: expr, MethodName: methodName, Arguments: args}, nil
	}

	return expr, nil
}

// parseTypeName parses a type name (single word or "unsigned integer")
func (p *Parser) parseTypeName() string {
	// Handle "unsigned integer"
	if p.curToken.Type == token.UNSIGNED {
		p.nextToken()
		if p.curToken.Type == token.INTEGER {
			p.nextToken()
			return "unsigned integer"
		}
		return "unsigned"
	}

	name := ""
	switch p.curToken.Type {
	case token.IDENTIFIER:
		name = strings.ToLower(p.curToken.Value)
	case token.INTEGER:
		name = "integer"
	case token.TYPE:
		name = strings.ToLower(p.curToken.Value)
	default:
		// Try interpreting keyword token values as type names
		name = strings.ToLower(p.curToken.Value)
	}
	if name != "" {
		p.nextToken()
	}
	return name
}

func (p *Parser) parseAdditive() (ast.Expression, error) {
	left, err := p.parseMultiplicative()
	if err != nil {
		return nil, err
	}

	for p.curToken.Type == token.PLUS || p.curToken.Type == token.MINUS {
		op := "+"
		if p.curToken.Type == token.MINUS {
			op = "-"
		}
		p.nextToken()
		right, err := p.parseMultiplicative()
		if err != nil {
			return nil, err
		}
		left = &ast.BinaryExpression{
			Left:     left,
			Operator: op,
			Right:    right,
		}
	}

	return left, nil
}

func (p *Parser) parseMultiplicative() (ast.Expression, error) {
	left, err := p.parsePrimary()
	if err != nil {
		return nil, err
	}

	for p.curToken.Type == token.STAR || p.curToken.Type == token.SLASH {
		op := "*"
		if p.curToken.Type == token.SLASH {
			op = "/"
		}
		p.nextToken()
		right, err := p.parsePrimary()
		if err != nil {
			return nil, err
		}
		left = &ast.BinaryExpression{
			Left:     left,
			Operator: op,
			Right:    right,
		}
	}

	return left, nil
}

func (p *Parser) parsePrimary() (ast.Expression, error) {
	switch p.curToken.Type {
	case token.NUMBER:
		value, _ := strconv.ParseFloat(p.curToken.Value, 64)
		p.nextToken()
		return &ast.NumberLiteral{Value: value}, nil

	case token.STRING:
		value := p.curToken.Value
		p.nextToken()
		return &ast.StringLiteral{Value: value}, nil

	case token.TRUE:
		p.nextToken()
		return &ast.BooleanLiteral{Value: true}, nil

	case token.FALSE:
		p.nextToken()
		return &ast.BooleanLiteral{Value: false}, nil

	case token.NOTHING:
		p.nextToken()
		return &ast.NothingLiteral{}, nil

	case token.NOT:
		// Logical NOT unary operator: "not <expression>"
		p.nextToken()
		expr, err := p.parsePrimary()
		if err != nil {
			return nil, err
		}
		return &ast.UnaryExpression{
			Operator: "not",
			Right:    expr,
		}, nil

	case token.ASK:
		// "ask(<prompt>)" or "ask" used as expression
		p.nextToken() // consume ASK
		if p.curToken.Type == token.LPAREN {
			p.nextToken() // consume (
			var prompt ast.Expression
			if p.curToken.Type != token.RPAREN {
				var err error
				prompt, err = p.parseExpression()
				if err != nil {
					return nil, err
				}
			}
			if err := p.expectToken(token.RPAREN); err != nil {
				return nil, err
			}
			p.nextToken()
			return &ast.AskExpression{Prompt: prompt}, nil
		}
		// "ask" with a string directly (no parentheses)
		prompt, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		return &ast.AskExpression{Prompt: prompt}, nil

	case token.LBRACKET:
		return p.parseList()

	case token.THE:
		// Handle "the item at position X in Y", "the length of X", "the entry KEY in TABLE", etc.
		p.nextToken()
		if p.curToken.Type == token.ITEM {
			return p.parseIndexExpression()
		}
		if p.curToken.Type == token.LENGTH {
			return p.parseLengthExpression()
		}
		if p.curToken.Type == token.REMAINDER {
			return p.parseRemainderExpression()
		}
		if p.curToken.Type == token.LOCATION {
			return p.parseLocationExpression()
		}
		if p.curToken.Type == token.TYPE {
			return p.parseTypeExpression()
		}
		// "the entry KEY in TABLE" — lookup table access
		if p.curToken.Type == token.ENTRY {
			return p.parseLookupKeyAccess()
		}
		// Check for field access: "the name of person"
		if p.curToken.Type == token.IDENTIFIER {
			fieldName := p.curToken.Value
			lowerFieldName := strings.ToLower(fieldName)
			p.nextToken()
			if p.curToken.Type == token.OF {
				p.nextToken()
				obj, err := p.parseExpression()
				if err != nil {
					return nil, err
				}
				// Map natural-English aggregate phrases to stdlib function calls.
				// e.g. "the number of names" → count(names)
				aggregateFuncs := map[string]string{
					"number": "count",
					"size":   "count",
					"sum":    "sum",
				}
				if funcName, ok := aggregateFuncs[lowerFieldName]; ok {
					return &ast.FunctionCall{Name: funcName, Arguments: []ast.Expression{obj}}, nil
				}
				return &ast.FieldAccess{
					Object: obj,
					Field:  fieldName,
				}, nil
			}
			return &ast.Identifier{Name: fieldName}, nil
		}
		if p.curToken.Type == token.VALUE {
			p.nextToken()
			if p.curToken.Type == token.OF {
				p.nextToken()
			}
			return p.parseExpression()
		}
		return nil, p.syntaxErr(
			fmt.Sprintf("I do not understand 'the %s' here.", p.curToken.Value),
			"After 'the' you can use: 'the value of x', 'the length of myList', 'the remainder of a divided by b', 'the result of calling myFunction', or a field name like 'the age of person'.",
		)

	case token.ITEM:
		// "item" used as a variable name (not "the item at position")
		p.nextToken()
		return &ast.Identifier{Name: "item"}, nil

	case token.IDENTIFIER:
		name := p.curToken.Value

		// Check for special identifier phrases
		if name == "a" || name == "an" {
			p.nextToken()
			if p.curToken.Type == token.NEW {
				// "a new instance of Person"
				return p.parseStructInstantiation()
			}
			if p.curToken.Type == token.REFERENCE {
				// "a reference to x"
				return p.parseReferenceExpression()
			}
			if p.curToken.Type == token.COPY {
				// "a copy of x"
				return p.parseCopyExpression()
			}
			if p.curToken.Type == token.LOOKUP {
				// "a lookup table"
				p.nextToken() // consume LOOKUP
				if p.curToken.Type == token.TABLE {
					p.nextToken() // consume TABLE
				}
				return &ast.LookupTableLiteral{}, nil
			}
			if p.curToken.Type == token.ARRAY {
				// "an array of [elements]" or "an array of number [elements]"
				return p.parseArrayLiteral()
			}
			// Not a special phrase, treat "a"/"an" as identifier
			return &ast.Identifier{Name: name}, nil
		}

		p.nextToken()

		// Possessive expression: "x's method" → MethodCall{Object: x, MethodName: method}
		// e.g. "her_love_txt's casefold" → casefold applied to her_love_txt
		if len(name) > 2 && name[len(name)-2:] == "'s" {
			objectName := name[:len(name)-2]
			if p.curToken.Type == token.IDENTIFIER {
				methodName := p.curToken.Value
				p.nextToken()
				var args []ast.Expression
				if p.curToken.Type == token.WITH {
					p.nextToken()
					args = p.parseCallArguments()
				}
				return &ast.MethodCall{
					Object:     &ast.Identifier{Name: objectName},
					MethodName: methodName,
					Arguments:  args,
				}, nil
			}
			// No method name after possessive — treat as plain identifier
			name = objectName
		}

		// Check if it's a function call
		if p.curToken.Type == token.LPAREN {
			p.nextToken()
			args, err := p.parseFunctionCallArgs()
			if err != nil {
				return nil, err
			}
			if err := p.expectToken(token.RPAREN); err != nil {
				return nil, err
			}
			p.nextToken()
			return &ast.FunctionCall{
				Name:      name,
				Arguments: args,
			}, nil
		}

		// "first of X", "last of X", "sum of X" etc. — natural-English function-call syntax.
		// Any bare identifier followed by "of" is treated as a single-argument function call.
		// The name is passed as-is (original case) to match how all other function calls work.
		if p.curToken.Type == token.OF {
			p.nextToken()
			arg, err := p.parseExpression()
			if err != nil {
				return nil, err
			}
			return &ast.FunctionCall{
				Name:      name,
				Arguments: []ast.Expression{arg},
			}, nil
		}

		// Check if it's array indexing with brackets: list[0]
		if p.curToken.Type == token.LBRACKET {
			p.nextToken()
			index, err := p.parseExpression()
			if err != nil {
				return nil, err
			}
			if err := p.expectToken(token.RBRACKET); err != nil {
				return nil, err
			}
			p.nextToken()
			return &ast.IndexExpression{
				List:  &ast.Identifier{Name: name},
				Index: index,
			}, nil
		}

		return &ast.Identifier{Name: name}, nil

	case token.LPAREN:
		p.nextToken()
		// Allow logical operators (and/or) inside parentheses
		expr, err := p.parseComparison()
		if err != nil {
			return nil, err
		}
		if err := p.expectToken(token.RPAREN); err != nil {
			return nil, err
		}
		p.nextToken()
		return expr, nil

	case token.MINUS:
		p.nextToken()
		expr, err := p.parsePrimary()
		if err != nil {
			return nil, err
		}
		return &ast.UnaryExpression{
			Operator: "-",
			Right:    expr,
		}, nil

	case token.NEW:
		// "new instance of Person" (without "a")
		return p.parseStructInstantiation()

	default:
		return nil, p.syntaxErr(
			fmt.Sprintf("I do not know how to use '%s' as a value here.", p.curToken.Value),
			"Values can be numbers (like 42), text (like \"hello\"), variables (like myScore), function calls, or expressions in parentheses.",
		)
	}
}

// parseIndexExpression parses "item at position X in/of Y"
func (p *Parser) parseIndexExpression() (ast.Expression, error) {
	// Already consumed "the", now at "item"
	if err := p.expectToken(token.ITEM); err != nil {
		return nil, err
	}
	p.nextToken()

	if err := p.expectToken(token.AT); err != nil {
		return nil, err
	}
	p.nextToken()

	if err := p.expectToken(token.POSITION); err != nil {
		return nil, err
	}
	p.nextToken()

	index, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	// Accept both "in" and "of": "the item at position 0 in list" / "of list"
	if p.curToken.Type != token.IN && p.curToken.Type != token.OF {
		return nil, p.syntaxErr(
			fmt.Sprintf("I expected 'in' or 'of' after the index number, but found '%s'.", p.curToken.Value),
			"For example: 'the item at position 1 in myList' or 'the item at position 1 of myList'.",
		)
	}
	p.nextToken()

	list, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	return &ast.IndexExpression{
		List:  list,
		Index: index,
	}, nil
}

// parseLengthExpression parses "length of X"
func (p *Parser) parseLengthExpression() (ast.Expression, error) {
	// Already consumed "the", now at "length"
	if err := p.expectToken(token.LENGTH); err != nil {
		return nil, err
	}
	p.nextToken()

	if err := p.expectToken(token.OF); err != nil {
		return nil, err
	}
	p.nextToken()

	list, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	return &ast.LengthExpression{
		List: list,
	}, nil
}

// parseRemainderExpression parses "remainder of X divided by Y" or "remainder of X / Y"
func (p *Parser) parseRemainderExpression() (ast.Expression, error) {
	// Already consumed "the", now at "remainder"
	if err := p.expectToken(token.REMAINDER); err != nil {
		return nil, err
	}
	p.nextToken()

	if err := p.expectToken(token.OF); err != nil {
		return nil, err
	}
	p.nextToken()

	// Parse the dividend (left operand)
	left, err := p.parsePrimary()
	if err != nil {
		return nil, err
	}

	// Expect "divided by" or "/"
	if p.curToken.Type == token.DIVIDED {
		p.nextToken()
		if err := p.expectToken(token.BY); err != nil {
			return nil, err
		}
		p.nextToken()
	} else if p.curToken.Type == token.SLASH {
		p.nextToken()
	} else {
		return nil, p.syntaxErr(
			fmt.Sprintf("I expected 'divided by' after the first number, but found '%s'.", p.curToken.Value),
			"For example: 'the remainder of 10 divided by 3' or 'the remainder of a divided by b'.",
		)
	}

	// Parse the divisor (right operand)
	right, err := p.parsePrimary()
	if err != nil {
		return nil, err
	}

	return &ast.BinaryExpression{
		Left:     left,
		Operator: "%",
		Right:    right,
	}, nil
}

// parseLocationExpression parses "location of X"
func (p *Parser) parseLocationExpression() (ast.Expression, error) {
	// Already consumed "the", now at "location"
	if err := p.expectToken(token.LOCATION); err != nil {
		return nil, err
	}
	p.nextToken()

	if err := p.expectToken(token.OF); err != nil {
		return nil, err
	}
	p.nextToken()

	// Get the variable name
	if p.curToken.Type != token.IDENTIFIER {
		return nil, p.syntaxErr(
			"I expected a variable name after 'the location of'.",
			"For example: 'the location of myVariable'.",
		)
	}
	name := p.curToken.Value
	p.nextToken()

	return &ast.LocationExpression{
		Name: name,
	}, nil
}

// parseTypeExpression parses "the type of x"
func (p *Parser) parseTypeExpression() (ast.Expression, error) {
	// Already consumed "the", now at "type"
	if err := p.expectToken(token.TYPE); err != nil {
		return nil, err
	}
	p.nextToken()

	if err := p.expectToken(token.OF); err != nil {
		return nil, err
	}
	p.nextToken()

	// Parse the expression whose type we want
	expr, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	return &ast.TypeExpression{Value: expr}, nil
}

// parseReferenceExpression parses "a reference to x"
func (p *Parser) parseReferenceExpression() (ast.Expression, error) {
	// Already consumed "a", now at "reference"
	if err := p.expectToken(token.REFERENCE); err != nil {
		return nil, err
	}
	p.nextToken()

	if err := p.expectToken(token.TO); err != nil {
		return nil, err
	}
	p.nextToken()

	if p.curToken.Type != token.IDENTIFIER {
		return nil, p.syntaxErr(
			"I expected a variable name after 'reference to'.",
			"For example: 'a reference to myVariable'.",
		)
	}
	name := p.curToken.Value
	p.nextToken()

	return &ast.ReferenceExpression{Name: name}, nil
}

// parseCopyExpression parses "a copy of x"
func (p *Parser) parseCopyExpression() (ast.Expression, error) {
	// Already consumed "a", now at "copy"
	if err := p.expectToken(token.COPY); err != nil {
		return nil, err
	}
	p.nextToken()

	if err := p.expectToken(token.OF); err != nil {
		return nil, err
	}
	p.nextToken()

	// Parse the expression to copy
	expr, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	return &ast.CopyExpression{Value: expr}, nil
}

// parseToggle parses "Toggle x." or "Toggle the value of x."
func (p *Parser) parseToggle() (ast.Statement, error) {
	if err := p.expectToken(token.TOGGLE); err != nil {
		return nil, err
	}
	startLine := p.curToken.Line
	p.nextToken()

	// Handle "toggle the value of x"
	if p.curToken.Type == token.THE {
		p.nextToken()
		if p.curToken.Type == token.VALUE {
			p.nextToken()
			if p.curToken.Type == token.OF {
				p.nextToken()
			}
		}
	}

	// Get the variable name
	if p.curToken.Type != token.IDENTIFIER {
		return nil, p.syntaxErr(
			"I expected a variable name after 'Toggle'.",
			"For example: 'Toggle isRunning.' or 'Toggle the value of isActive.'",
		)
	}
	name := p.curToken.Value
	p.nextToken()

	if err := p.expectToken(token.PERIOD); err != nil {
		return nil, err
	}
	p.nextToken()

	return &ast.ToggleStatement{
		Name: name,
		Line: startLine,
	}, nil
}

func (p *Parser) parseList() (ast.Expression, error) {
	if err := p.expectToken(token.LBRACKET); err != nil {
		return nil, err
	}
	p.nextToken()

	var elements []ast.Expression

	if p.curToken.Type != token.RBRACKET {
		for {
			elem, err := p.parseExpression()
			if err != nil {
				return nil, err
			}
			elements = append(elements, elem)

			if p.curToken.Type != token.COMMA {
				break
			}
			p.nextToken()
		}
	}

	if err := p.expectToken(token.RBRACKET); err != nil {
		return nil, err
	}
	p.nextToken()

	return &ast.ListLiteral{Elements: elements}, nil
}

func (p *Parser) parseFunctionArguments() ([]ast.Expression, error) {
	var args []ast.Expression

	if p.curToken.Type == token.WITH {
		p.nextToken()
		for {
			arg, err := p.parseExpression()
			if err != nil {
				return nil, err
			}
			args = append(args, arg)

			if p.curToken.Type != token.AND {
				break
			}
			p.nextToken()
		}
	}

	return args, nil
}

func (p *Parser) parseFunctionCallArgs() ([]ast.Expression, error) {
	var args []ast.Expression

	if p.curToken.Type != token.RPAREN {
		for {
			arg, err := p.parseExpression()
			if err != nil {
				return nil, err
			}
			args = append(args, arg)

			if p.curToken.Type != token.COMMA {
				break
			}
			p.nextToken()
		}
	}

	return args, nil
}

// parseArrayLiteral parses "an array of [TYPE] [elements]"
// Cursor is on ARRAY token when called.
func (p *Parser) parseArrayLiteral() (ast.Expression, error) {
	p.nextToken() // consume ARRAY

	if p.curToken.Type != token.OF {
		return nil, p.syntaxErr(
			"I expected 'of' after 'array'.",
			"For example: 'an array of [1, 2, 3]' or 'an array of number [1, 2, 3]'.",
		)
	}
	p.nextToken() // consume OF

	// Optional element type hint before the bracket
	elementType := ""
	if p.curToken.Type != token.LBRACKET {
		elementType = p.parseTypeName()
	}

	if p.curToken.Type != token.LBRACKET {
		hint := "For example: 'an array of [1, 2, 3]' or 'an array of number [1, 2, 3]'."
		if elementType != "" {
			hint = fmt.Sprintf("After 'array of %s' I expected '[' to open the list of values.", elementType)
		}
		return nil, p.syntaxErr(
			fmt.Sprintf("I expected '[' to open the array, but found '%s'.", p.curToken.Value),
			hint,
		)
	}
	p.nextToken() // consume [

	var elements []ast.Expression
	for p.curToken.Type != token.RBRACKET && p.curToken.Type != token.EOF {
		elem, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		elements = append(elements, elem)
		if p.curToken.Type == token.COMMA {
			p.nextToken()
		}
	}
	if p.curToken.Type != token.RBRACKET {
		return nil, p.syntaxErr(
			"I expected ']' to close the array, but reached the end of the file.",
			"Make sure every '[' is matched by a closing ']'. For example: 'an array of [1, 2, 3]'.",
		)
	}
	p.nextToken() // consume ]

	return &ast.ArrayLiteral{ElementType: elementType, Elements: elements}, nil
}

// parseLookupKeyAccess parses "the entry KEY in TABLE".
// Cursor is on ENTRY when called.
func (p *Parser) parseLookupKeyAccess() (ast.Expression, error) {
	p.nextToken() // consume ENTRY

	key, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	if p.curToken.Type != token.IN {
		return nil, p.syntaxErr(
			fmt.Sprintf("I expected 'in' after the key, but found '%s'.", p.curToken.Value),
			"For example: 'the entry \"name\" in myTable'.",
		)
	}
	p.nextToken() // consume IN

	table, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	return &ast.LookupKeyAccess{Table: table, Key: key}, nil
}

// parseLookupKeyAssignment parses "the entry KEY in TABLE to be VALUE."
// Cursor is on ENTRY when called (parseAssignment has already consumed "Set the").
func (p *Parser) parseLookupKeyAssignment(setLine int) (ast.Statement, error) {
	p.nextToken() // consume ENTRY

	key, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	if p.curToken.Type != token.IN {
		return nil, p.syntaxErr(
			fmt.Sprintf("I expected 'in' after the key, but found '%s'.", p.curToken.Value),
			"For example: 'Set the entry \"name\" in myTable to be \"Alice\".'",
		)
	}
	p.nextToken() // consume IN

	if p.curToken.Type != token.IDENTIFIER {
		return nil, p.syntaxErr(
			"I expected the table name after 'in'.",
			"For example: 'Set the entry \"name\" in myTable to be \"Alice\".'",
		)
	}
	tableName := p.curToken.Value
	p.nextToken()

	if p.curToken.Type != token.TO {
		return nil, p.syntaxErr(
			fmt.Sprintf("I expected 'to' after the table name '%s', but found '%s'.", tableName, p.curToken.Value),
			"For example: 'Set the entry \"name\" in myTable to be \"Alice\".'",
		)
	}
	p.nextToken()
	if p.curToken.Type == token.BE {
		p.nextToken()
	}

	value, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	if err := p.expectToken(token.PERIOD); err != nil {
		return nil, err
	}
	p.nextToken()

	return &ast.LookupKeyAssignment{TableName: tableName, Key: key, Value: value, Line: setLine}, nil
}
