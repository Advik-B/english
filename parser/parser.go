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
	var suggestion string

	// Provide helpful suggestions based on context
	switch expected {
	case token.PERIOD:
		suggestion = "\n  Perhaps you forgot to end the statement with a period (.)"
	case token.TO:
		if p.curToken.Type == token.BE {
			suggestion = "\n  Perhaps you meant: 'to be' instead of just 'be'"
		}
	case token.BE:
		if p.curToken.Type == token.TO {
			suggestion = "\n  Perhaps you meant: 'to be' (you have 'to' but missing 'be')"
		}
	case token.THATS:
		suggestion = "\n  Perhaps you forgot to end the block with 'thats it.'"
	case token.IT:
		if p.curToken.Type == token.PERIOD {
			suggestion = "\n  Perhaps you meant: 'thats it.' (missing 'it' before the period)"
		}
	case token.IDENTIFIER:
		if p.curToken.Type == token.NUMBER || p.curToken.Type == token.STRING {
			suggestion = "\n  A variable name (identifier) is expected here, not a literal value"
		}
	}

	return fmt.Errorf("expected %v, got %v at line %d, column %d%s",
		expected, p.curToken.Type, p.curToken.Line, p.curToken.Col, suggestion)
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
	case token.IMPORT:
		return p.parseImport()
	case token.DECLARE:
		return p.parseDeclaration()
	case token.LET:
		return p.parseLetDeclaration()
	case token.BREAK:
		return p.parseBreak()
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
		suggestion := ""
		switch p.curToken.Type {
		case token.IDENTIFIER:
			suggestion = "\n  Hint: To use a variable, you need 'Set', 'Print', or another statement keyword"
		case token.NUMBER, token.STRING:
			suggestion = "\n  Hint: Literal values must be part of a statement (e.g., 'Print \"text\".' or 'Declare x to be 5.')"
		case token.EOF:
			suggestion = "\n  Hint: Unexpected end of file - check if you have unclosed blocks"
		}
		return nil, fmt.Errorf("unexpected token: %v (value: '%s') at line %d, column %d%s",
			p.curToken.Type, p.curToken.Value, p.curToken.Line, p.curToken.Col, suggestion)
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
		return nil, fmt.Errorf("expected identifier after 'let', got %v at line %d", p.curToken.Type, p.curToken.Line)
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
		return nil, fmt.Errorf("expected 'be', '=', 'equal', or 'always' after variable name, got %v at line %d", p.curToken.Type, p.curToken.Line)
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
	}, nil
}

// parseImport parses import statements with natural English syntax:
// - Import code from "file.abc".
// - Import "utilities.abc".
func (p *Parser) parseImport() (ast.Statement, error) {
	if err := p.expectToken(token.IMPORT); err != nil {
		return nil, err
	}
	p.nextToken()

	// Handle optional "code" or "the" keywords for natural language
	// "Import code from file.abc" or "Import the code from file.abc"
	if p.curToken.Type == token.THE {
		p.nextToken()
	}
	
	// Skip optional "code" keyword
	if p.curToken.Type == token.IDENTIFIER && strings.ToLower(p.curToken.Value) == "code" {
		p.nextToken()
	}

	// Handle optional "from" keyword
	if p.curToken.Type == token.FROM {
		p.nextToken()
	}

	// Expect a string with the file path
	if p.curToken.Type != token.STRING {
		return nil, fmt.Errorf("expected file path (string) after 'Import', got %v at line %d", p.curToken.Type, p.curToken.Line)
	}

	filePath := p.curToken.Value
	p.nextToken()

	// Expect period to end the statement
	if err := p.expectToken(token.PERIOD); err != nil {
		return nil, err
	}
	p.nextToken()

	return &ast.ImportStatement{
		Path: filePath,
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

	// Check if it's a struct declaration: "Declare Person as a structure..."
	// We need to peek ahead to see if we have "as" followed by "structure"/"struct"
	if p.curToken.Type == token.IDENTIFIER && p.peekToken.Type == token.AS {
		// Save position and try struct parsing
		return p.parseStructDeclaration()
	}

	// Variable or constant declaration
	nameToken := p.curToken
	if p.curToken.Type != token.IDENTIFIER {
		return nil, fmt.Errorf("expected identifier after 'Declare', got %v at line %d", p.curToken.Type, p.curToken.Line)
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
	}, nil
}

func (p *Parser) parseFunctionDeclaration() (ast.Statement, error) {
	if err := p.expectToken(token.FUNCTION); err != nil {
		return nil, err
	}
	p.nextToken()

	nameToken := p.curToken
	if p.curToken.Type != token.IDENTIFIER {
		return nil, fmt.Errorf("expected function name, got %v", p.curToken.Type)
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
				return nil, fmt.Errorf("expected parameter name")
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
	p.nextToken()

	// Check for "Set the item at position X in Y to be Z"
	if p.curToken.Type == token.THE {
		p.nextToken()
		if p.curToken.Type == token.ITEM {
			return p.parseIndexAssignment()
		}
		return nil, fmt.Errorf("unexpected token after 'Set the': %v", p.curToken.Type)
	}

	nameToken := p.curToken
	if p.curToken.Type != token.IDENTIFIER {
		return nil, fmt.Errorf("expected identifier after 'Set'")
	}
	p.nextToken()

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
					return nil, fmt.Errorf("expected function name")
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
	}, nil
}

// parseIndexAssignment parses "the item at position X in Y to be Z"
func (p *Parser) parseIndexAssignment() (ast.Statement, error) {
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
		return nil, fmt.Errorf("expected list name")
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
	}, nil
}

func (p *Parser) parseCall() (ast.Statement, error) {
	if err := p.expectToken(token.CALL); err != nil {
		return nil, err
	}
	p.nextToken()

	// First identifier could be:
	// 1. Function name: "call greet with args."
	// 2. Method name: "call talk from p2." or "call talk on p2."
	// 3. Object name with possessive: "call p2's talk." (p2's is a single token)
	
	firstIdent := p.curToken.Value
	if p.curToken.Type != token.IDENTIFIER {
		return nil, fmt.Errorf("expected identifier after 'Call'")
	}
	p.nextToken()

	// Check for possessive syntax: "call p2's talk."
	// The identifier will end with 's (e.g., "p2's")
	if len(firstIdent) > 2 && firstIdent[len(firstIdent)-2:] == "'s" {
		// This is possessive: extract object name (remove 's)
		objectName := firstIdent[:len(firstIdent)-2]
		
		if p.curToken.Type != token.IDENTIFIER {
			return nil, fmt.Errorf("expected method name after possessive")
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
		}, nil
	}

	// Check for "from" or "on" (method call syntax)
	if p.curToken.Type == token.FROM || p.curToken.Type == token.ON {
		methodName := firstIdent
		p.nextToken() // skip FROM/ON
		
		// Get object
		if p.curToken.Type != token.IDENTIFIER {
			return nil, fmt.Errorf("expected object name after 'from'/'on'")
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
	}, nil
}

func (p *Parser) parseRepeat() (ast.Statement, error) {
	if err := p.expectToken(token.REPEAT); err != nil {
		return nil, err
	}
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
		return nil, fmt.Errorf("expected item identifier in for-each")
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
	}, nil
}

func (p *Parser) parseOutput(newline bool) (ast.Statement, error) {
	// Accept either PRINT or WRITE token
	if p.curToken.Type != token.PRINT && p.curToken.Type != token.WRITE {
		return nil, fmt.Errorf("expected %v or %v, got %v", token.PRINT, token.WRITE, p.curToken.Type)
	}
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
	}, nil
}

func (p *Parser) parseReturn() (ast.Statement, error) {
	if err := p.expectToken(token.RETURN); err != nil {
		return nil, err
	}
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
		return nil, fmt.Errorf("expected 'the' or 'this', got %v at line %d, column %d",
			p.curToken.Type, p.curToken.Line, p.curToken.Col)
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
	left, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	switch p.curToken.Type {
	case token.IS_EQUAL_TO, token.IS_LESS_THAN, token.IS_GREATER_THAN,
		token.IS_LESS_EQUAL, token.IS_GREATER_EQUAL, token.IS_NOT_EQUAL:
		op := p.curToken.Value
		p.nextToken()
		right, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		return &ast.BinaryExpression{
			Left:     left,
			Operator: op,
			Right:    right,
		}, nil
	}

	return left, nil
}

func (p *Parser) parseExpression() (ast.Expression, error) {
	return p.parseAdditive()
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

	case token.LBRACKET:
		return p.parseList()

	case token.THE:
		// Handle "the item at position X in Y" or "the length of X" or "the remainder of X divided by Y" or "the location of X" or "the type of X" or "the name of person" (field access)
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
		// Check for field access: "the name of person"
		if p.curToken.Type == token.IDENTIFIER {
			fieldName := p.curToken.Value
			p.nextToken()
			if p.curToken.Type == token.OF {
				p.nextToken()
				// Parse the object expression
				obj, err := p.parseExpression()
				if err != nil {
					return nil, err
				}
				return &ast.FieldAccess{
					Object: obj,
					Field:  fieldName,
				}, nil
			}
			// Not field access, restore identifier
			return &ast.Identifier{Name: fieldName}, nil
		}
		// Fall back to treating "the" as part of other constructs
		// Put back THE token context - this is for "the value of" pattern
		if p.curToken.Type == token.VALUE {
			p.nextToken()
			if p.curToken.Type == token.OF {
				p.nextToken()
			}
			return p.parseExpression()
		}
		return nil, fmt.Errorf("unexpected token after 'the': %v at line %d", p.curToken.Type, p.curToken.Line)

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
			// Not a special phrase, treat "a"/"an" as identifier
			return &ast.Identifier{Name: name}, nil
		}
		
		p.nextToken()

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
		expr, err := p.parseExpression()
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
		return nil, fmt.Errorf("unexpected token in expression: %v at line %d", p.curToken.Type, p.curToken.Line)
	}
}

// parseIndexExpression parses "item at position X in Y"
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

	if err := p.expectToken(token.IN); err != nil {
		return nil, err
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
		return nil, fmt.Errorf("expected 'divided by' or '/' after remainder operand, got %v", p.curToken.Type)
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
		return nil, fmt.Errorf("expected variable name after 'the location of', got %v", p.curToken.Type)
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
		return nil, fmt.Errorf("expected variable name after 'reference to', got %v", p.curToken.Type)
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
		return nil, fmt.Errorf("expected variable name after 'Toggle', got %v", p.curToken.Type)
	}
	name := p.curToken.Value
	p.nextToken()

	if err := p.expectToken(token.PERIOD); err != nil {
		return nil, err
	}
	p.nextToken()

	return &ast.ToggleStatement{
		Name: name,
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
