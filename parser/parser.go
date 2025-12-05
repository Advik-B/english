package parser

import (
	"english/ast"
	"english/token"
	"fmt"
	"strconv"
	"strings"
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
	case token.DECLARE:
		return p.parseDeclaration()
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
		return p.parseOutput()
	case token.RETURN:
		return p.parseReturn()
	case token.TOGGLE:
		return p.parseToggle()
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

func (p *Parser) parseDeclaration() (ast.Statement, error) {
	if err := p.expectToken(token.DECLARE); err != nil {
		return nil, err
	}
	p.nextToken()

	// Check if it's a function declaration
	if p.curToken.Type == token.FUNCTION {
		return p.parseFunctionDeclaration()
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

	if err := p.expectToken(token.BE); err != nil {
		return nil, err
	}
	p.nextToken()

	// Check for function call result
	if p.curToken.Type == token.THE {
		p.nextToken()
		if p.curToken.Type == token.IDENTIFIER && strings.EqualFold(p.curToken.Value, "result") {
			p.nextToken()
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

	if err := p.expectToken(token.BE); err != nil {
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

	funcName := p.curToken.Value
	if p.curToken.Type != token.IDENTIFIER {
		return nil, fmt.Errorf("expected function name after 'Call'")
	}
	p.nextToken()

	if err := p.expectToken(token.PERIOD); err != nil {
		return nil, err
	}
	p.nextToken()

	return &ast.CallStatement{
		FunctionCall: &ast.FunctionCall{
			Name:      funcName,
			Arguments: []ast.Expression{},
		},
	}, nil
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

func (p *Parser) parseOutput() (ast.Statement, error) {
	if err := p.expectToken(token.PRINT); err != nil {
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

	return &ast.OutputStatement{
		Value: value,
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

func (p *Parser) parseBlock() ([]ast.Statement, error) {
	var statements []ast.Statement

	for p.curToken.Type != token.THATS && p.curToken.Type != token.OTHERWISE && p.curToken.Type != token.EOF {
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
		// Handle "the item at position X in Y" or "the length of X" or "the remainder of X divided by Y" or "the location of X"
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
