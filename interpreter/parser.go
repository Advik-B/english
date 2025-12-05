package interpreter

import (
	"fmt"
	"strconv"
	"strings"
)

type Parser struct {
	tokens    []Token
	position  int
	curToken  Token
	peekToken Token
}

func NewParser(tokens []Token) *Parser {
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
		p.peekToken = Token{Type: TOKEN_EOF}
	}
}

func (p *Parser) expectToken(tokenType TokenType) error {
	if p.curToken.Type != tokenType {
		return p.makeExpectError(tokenType)
	}
	return nil
}

func (p *Parser) makeExpectError(expected TokenType) error {
	var suggestion string

	// Provide helpful suggestions based on context
	switch expected {
	case TOKEN_PERIOD:
		suggestion = "\n  Perhaps you forgot to end the statement with a period (.)"
	case TOKEN_TO:
		if p.curToken.Type == TOKEN_BE {
			suggestion = "\n  Perhaps you meant: 'to be' instead of just 'be'"
		}
	case TOKEN_BE:
		if p.curToken.Type == TOKEN_TO {
			suggestion = "\n  Perhaps you meant: 'to be' (you have 'to' but missing 'be')"
		}
	case TOKEN_THATS:
		suggestion = "\n  Perhaps you forgot to end the block with 'thats it.'"
	case TOKEN_IT:
		if p.curToken.Type == TOKEN_PERIOD {
			suggestion = "\n  Perhaps you meant: 'thats it.' (missing 'it' before the period)"
		}
	case TOKEN_IDENTIFIER:
		if p.curToken.Type == TOKEN_NUMBER || p.curToken.Type == TOKEN_STRING {
			suggestion = "\n  A variable name (identifier) is expected here, not a literal value"
		}
	}

	return fmt.Errorf("expected %v, got %v at line %d, column %d%s",
		expected, p.curToken.Type, p.curToken.Line, p.curToken.Col, suggestion)
}

func (p *Parser) Parse() (*Program, error) {
	program := &Program{}

	for p.curToken.Type != TOKEN_EOF {
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

func (p *Parser) parseStatement() (Statement, error) {
	switch p.curToken.Type {
	case TOKEN_DECLARE:
		return p.parseDeclaration()
	case TOKEN_SET:
		return p.parseAssignment()
	case TOKEN_CALL:
		return p.parseCall()
	case TOKEN_IF:
		return p.parseIfStatement()
	case TOKEN_REPEAT:
		return p.parseRepeat()
	case TOKEN_FOR:
		return p.parseForEach()
	case TOKEN_PRINT:
		return p.parseOutput()
	case TOKEN_RETURN:
		return p.parseReturn()
	case TOKEN_TOGGLE:
		return p.parseToggle()
	default:
		suggestion := ""
		switch p.curToken.Type {
		case TOKEN_IDENTIFIER:
			suggestion = "\n  Hint: To use a variable, you need 'Set', 'Print', or another statement keyword"
		case TOKEN_NUMBER, TOKEN_STRING:
			suggestion = "\n  Hint: Literal values must be part of a statement (e.g., 'Print \"text\".' or 'Declare x to be 5.')"
		case TOKEN_EOF:
			suggestion = "\n  Hint: Unexpected end of file - check if you have unclosed blocks"
		}
		return nil, fmt.Errorf("unexpected token: %v (value: '%s') at line %d, column %d%s",
			p.curToken.Type, p.curToken.Value, p.curToken.Line, p.curToken.Col, suggestion)
	}
}

func (p *Parser) parseDeclaration() (Statement, error) {
	if err := p.expectToken(TOKEN_DECLARE); err != nil {
		return nil, err
	}
	p.nextToken()

	// Check if it's a function declaration
	if p.curToken.Type == TOKEN_FUNCTION {
		return p.parseFunctionDeclaration()
	}

	// Variable or constant declaration
	nameToken := p.curToken
	if p.curToken.Type != TOKEN_IDENTIFIER {
		return nil, fmt.Errorf("expected identifier after 'Declare', got %v at line %d", p.curToken.Type, p.curToken.Line)
	}
	p.nextToken()

	if err := p.expectToken(TOKEN_TO); err != nil {
		return nil, err
	}
	p.nextToken()

	// Check for "always" keyword (can appear before or after "be")
	isConstant := false
	if p.curToken.Type == TOKEN_ALWAYS {
		isConstant = true
		p.nextToken()
	}

	if err := p.expectToken(TOKEN_BE); err != nil {
		return nil, err
	}
	p.nextToken()

	// Check for "always" after "be" if not seen before
	if !isConstant && p.curToken.Type == TOKEN_ALWAYS {
		isConstant = true
		p.nextToken()
	}

	value, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	if err := p.expectToken(TOKEN_PERIOD); err != nil {
		return nil, err
	}
	p.nextToken()

	return &VariableDecl{
		Name:       nameToken.Value,
		IsConstant: isConstant,
		Value:      value,
	}, nil
}

func (p *Parser) parseFunctionDeclaration() (Statement, error) {
	if err := p.expectToken(TOKEN_FUNCTION); err != nil {
		return nil, err
	}
	p.nextToken()

	nameToken := p.curToken
	if p.curToken.Type != TOKEN_IDENTIFIER {
		return nil, fmt.Errorf("expected function name, got %v", p.curToken.Type)
	}
	p.nextToken()

	var parameters []string

	// Skip optional "that" before "takes" or "does"
	if p.curToken.Type == TOKEN_THAT {
		p.nextToken()
	}

	if p.curToken.Type == TOKEN_TAKES {
		p.nextToken()
		for {
			paramToken := p.curToken
			if p.curToken.Type != TOKEN_IDENTIFIER {
				return nil, fmt.Errorf("expected parameter name")
			}
			parameters = append(parameters, paramToken.Value)
			p.nextToken()

			if p.curToken.Type != TOKEN_AND {
				break
			}
			// Check if "and" is followed by "does" (end of params) or another param
			if p.peekToken.Type == TOKEN_DOES {
				break
			}
			p.nextToken()
		}
	}

	// Support "and does" syntax after parameters
	if p.curToken.Type == TOKEN_AND {
		p.nextToken()
	}

	if err := p.expectToken(TOKEN_DOES); err != nil {
		return nil, err
	}
	p.nextToken()

	if err := p.expectToken(TOKEN_THE); err != nil {
		return nil, err
	}
	p.nextToken()

	if err := p.expectToken(TOKEN_FOLLOWING); err != nil {
		return nil, err
	}
	p.nextToken()

	if err := p.expectToken(TOKEN_COLON); err != nil {
		return nil, err
	}
	p.nextToken()

	body, err := p.parseBlock()
	if err != nil {
		return nil, err
	}

	if p.curToken.Type == TOKEN_THATS {
		p.nextToken()
		if err := p.expectToken(TOKEN_IT); err != nil {
			return nil, err
		}
		p.nextToken()
		if err := p.expectToken(TOKEN_PERIOD); err != nil {
			return nil, err
		}
		p.nextToken()
	}

	return &FunctionDecl{
		Name:       nameToken.Value,
		Parameters: parameters,
		Body:       body,
	}, nil
}

func (p *Parser) parseAssignment() (Statement, error) {
	if err := p.expectToken(TOKEN_SET); err != nil {
		return nil, err
	}
	p.nextToken()

	// Check for "Set the item at position X in Y to be Z"
	if p.curToken.Type == TOKEN_THE {
		p.nextToken()
		if p.curToken.Type == TOKEN_ITEM {
			return p.parseIndexAssignment()
		}
		return nil, fmt.Errorf("unexpected token after 'Set the': %v", p.curToken.Type)
	}

	nameToken := p.curToken
	if p.curToken.Type != TOKEN_IDENTIFIER {
		return nil, fmt.Errorf("expected identifier after 'Set'")
	}
	p.nextToken()

	if err := p.expectToken(TOKEN_TO); err != nil {
		return nil, err
	}
	p.nextToken()

	if err := p.expectToken(TOKEN_BE); err != nil {
		return nil, err
	}
	p.nextToken()

	// Check for function call result
	if p.curToken.Type == TOKEN_THE {
		p.nextToken()
		if p.curToken.Type == TOKEN_IDENTIFIER && strings.EqualFold(p.curToken.Value, "result") {
			p.nextToken()
			if p.curToken.Type == TOKEN_OF {
				p.nextToken()
				if p.curToken.Type == TOKEN_CALLING {
					p.nextToken()
					funcName := p.curToken.Value
					if p.curToken.Type != TOKEN_IDENTIFIER {
						return nil, fmt.Errorf("expected function name")
					}
					p.nextToken()

					args, err := p.parseFunctionArguments()
					if err != nil {
						return nil, err
					}

					if err := p.expectToken(TOKEN_PERIOD); err != nil {
						return nil, err
					}
					p.nextToken()

					return &Assignment{
						Name: nameToken.Value,
						Value: &FunctionCall{
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

	if err := p.expectToken(TOKEN_PERIOD); err != nil {
		return nil, err
	}
	p.nextToken()

	return &Assignment{
		Name:  nameToken.Value,
		Value: value,
	}, nil
}

// parseIndexAssignment parses "the item at position X in Y to be Z"
func (p *Parser) parseIndexAssignment() (Statement, error) {
	// Already consumed "Set the", now at "item"
	if err := p.expectToken(TOKEN_ITEM); err != nil {
		return nil, err
	}
	p.nextToken()

	if err := p.expectToken(TOKEN_AT); err != nil {
		return nil, err
	}
	p.nextToken()

	if err := p.expectToken(TOKEN_POSITION); err != nil {
		return nil, err
	}
	p.nextToken()

	index, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	if err := p.expectToken(TOKEN_IN); err != nil {
		return nil, err
	}
	p.nextToken()

	listName := p.curToken.Value
	if p.curToken.Type != TOKEN_IDENTIFIER {
		return nil, fmt.Errorf("expected list name")
	}
	p.nextToken()

	if err := p.expectToken(TOKEN_TO); err != nil {
		return nil, err
	}
	p.nextToken()

	if err := p.expectToken(TOKEN_BE); err != nil {
		return nil, err
	}
	p.nextToken()

	value, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	if err := p.expectToken(TOKEN_PERIOD); err != nil {
		return nil, err
	}
	p.nextToken()

	return &IndexAssignment{
		ListName: listName,
		Index:    index,
		Value:    value,
	}, nil
}

func (p *Parser) parseCall() (Statement, error) {
	if err := p.expectToken(TOKEN_CALL); err != nil {
		return nil, err
	}
	p.nextToken()

	funcName := p.curToken.Value
	if p.curToken.Type != TOKEN_IDENTIFIER {
		return nil, fmt.Errorf("expected function name after 'Call'")
	}
	p.nextToken()

	if err := p.expectToken(TOKEN_PERIOD); err != nil {
		return nil, err
	}
	p.nextToken()

	return &CallStatement{
		FunctionCall: &FunctionCall{
			Name:      funcName,
			Arguments: []Expression{},
		},
	}, nil
}

func (p *Parser) parseIfStatement() (Statement, error) {
	if err := p.expectToken(TOKEN_IF); err != nil {
		return nil, err
	}
	p.nextToken()

	condition, err := p.parseComparison()
	if err != nil {
		return nil, err
	}

	if err := p.expectToken(TOKEN_COMMA); err != nil {
		return nil, err
	}
	p.nextToken()

	if err := p.expectToken(TOKEN_THEN); err != nil {
		return nil, err
	}
	p.nextToken()

	thenBody, err := p.parseBlock()
	if err != nil {
		return nil, err
	}

	var elseIfParts []*ElseIfPart
	var elseBody []Statement

	for p.curToken.Type == TOKEN_OTHERWISE {
		p.nextToken()
		if p.curToken.Type == TOKEN_IF {
			p.nextToken()
			eifCond, err := p.parseComparison()
			if err != nil {
				return nil, err
			}
			if err := p.expectToken(TOKEN_COMMA); err != nil {
				return nil, err
			}
			p.nextToken()
			if err := p.expectToken(TOKEN_THEN); err != nil {
				return nil, err
			}
			p.nextToken()
			eifBody, err := p.parseBlock()
			if err != nil {
				return nil, err
			}
			elseIfParts = append(elseIfParts, &ElseIfPart{
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

	if p.curToken.Type == TOKEN_THATS {
		p.nextToken()
		if err := p.expectToken(TOKEN_IT); err != nil {
			return nil, err
		}
		p.nextToken()
		if err := p.expectToken(TOKEN_PERIOD); err != nil {
			return nil, err
		}
		p.nextToken()
	}

	return &IfStatement{
		Condition: condition,
		Then:      thenBody,
		ElseIf:    elseIfParts,
		Else:      elseBody,
	}, nil
}

func (p *Parser) parseRepeat() (Statement, error) {
	if err := p.expectToken(TOKEN_REPEAT); err != nil {
		return nil, err
	}
	p.nextToken()

	if err := p.expectToken(TOKEN_THE); err != nil {
		return nil, err
	}
	p.nextToken()

	if err := p.expectToken(TOKEN_FOLLOWING); err != nil {
		return nil, err
	}
	p.nextToken()

	// Check if it's a while loop or for loop
	if p.curToken.Type == TOKEN_WHILE {
		p.nextToken()
		condition, err := p.parseComparison()
		if err != nil {
			return nil, err
		}

		if err := p.expectToken(TOKEN_COLON); err != nil {
			return nil, err
		}
		p.nextToken()

		body, err := p.parseBlock()
		if err != nil {
			return nil, err
		}

		if p.curToken.Type == TOKEN_THATS {
			p.nextToken()
			if err := p.expectToken(TOKEN_IT); err != nil {
				return nil, err
			}
			p.nextToken()
			if err := p.expectToken(TOKEN_PERIOD); err != nil {
				return nil, err
			}
			p.nextToken()
		}

		return &WhileLoop{
			Condition: condition,
			Body:      body,
		}, nil
	}

	// For loop (N times)
	countExpr, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	if err := p.expectToken(TOKEN_TIMES); err != nil {
		return nil, err
	}
	p.nextToken()

	if err := p.expectToken(TOKEN_COLON); err != nil {
		return nil, err
	}
	p.nextToken()

	body, err := p.parseBlock()
	if err != nil {
		return nil, err
	}

	if p.curToken.Type == TOKEN_THATS {
		p.nextToken()
		if err := p.expectToken(TOKEN_IT); err != nil {
			return nil, err
		}
		p.nextToken()
		if err := p.expectToken(TOKEN_PERIOD); err != nil {
			return nil, err
		}
		p.nextToken()
	}

	return &ForLoop{
		Count: countExpr,
		Body:  body,
	}, nil
}

func (p *Parser) parseForEach() (Statement, error) {
	if err := p.expectToken(TOKEN_FOR); err != nil {
		return nil, err
	}
	p.nextToken()

	if err := p.expectToken(TOKEN_EACH); err != nil {
		return nil, err
	}
	p.nextToken()

	itemToken := p.curToken
	// Allow both IDENTIFIER and ITEM keyword as the loop variable name
	if p.curToken.Type != TOKEN_IDENTIFIER && p.curToken.Type != TOKEN_ITEM {
		return nil, fmt.Errorf("expected item identifier in for-each")
	}
	// Get the value, treating TOKEN_ITEM as "item" string
	itemName := itemToken.Value
	if itemToken.Type == TOKEN_ITEM {
		itemName = "item"
	}
	p.nextToken()

	if err := p.expectToken(TOKEN_IN); err != nil {
		return nil, err
	}
	p.nextToken()

	listExpr, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	if err := p.expectToken(TOKEN_COMMA); err != nil {
		return nil, err
	}
	p.nextToken()

	if err := p.expectToken(TOKEN_DO); err != nil {
		return nil, err
	}
	p.nextToken()

	if err := p.expectToken(TOKEN_THE); err != nil {
		return nil, err
	}
	p.nextToken()

	if err := p.expectToken(TOKEN_FOLLOWING); err != nil {
		return nil, err
	}
	p.nextToken()

	if err := p.expectToken(TOKEN_COLON); err != nil {
		return nil, err
	}
	p.nextToken()

	body, err := p.parseBlock()
	if err != nil {
		return nil, err
	}

	if p.curToken.Type == TOKEN_THATS {
		p.nextToken()
		if err := p.expectToken(TOKEN_IT); err != nil {
			return nil, err
		}
		p.nextToken()
		if err := p.expectToken(TOKEN_PERIOD); err != nil {
			return nil, err
		}
		p.nextToken()
	}

	return &ForEachLoop{
		Item: itemName,
		List: listExpr,
		Body: body,
	}, nil
}

func (p *Parser) parseOutput() (Statement, error) {
	if err := p.expectToken(TOKEN_PRINT); err != nil {
		return nil, err
	}
	p.nextToken()

	// Check for "the value of" pattern - handled by parseExpression now via parsePrimary
	// Just parse expression directly
	value, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	if err := p.expectToken(TOKEN_PERIOD); err != nil {
		return nil, err
	}
	p.nextToken()

	return &OutputStatement{
		Value: value,
	}, nil
}

func (p *Parser) parseReturn() (Statement, error) {
	if err := p.expectToken(TOKEN_RETURN); err != nil {
		return nil, err
	}
	p.nextToken()

	value, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	if err := p.expectToken(TOKEN_PERIOD); err != nil {
		return nil, err
	}
	p.nextToken()

	return &ReturnStatement{
		Value: value,
	}, nil
}

func (p *Parser) parseBlock() ([]Statement, error) {
	var statements []Statement

	for p.curToken.Type != TOKEN_THATS && p.curToken.Type != TOKEN_OTHERWISE && p.curToken.Type != TOKEN_EOF {
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

func (p *Parser) parseComparison() (Expression, error) {
	left, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	switch p.curToken.Type {
	case TOKEN_IS_EQUAL_TO, TOKEN_IS_LESS_THAN, TOKEN_IS_GREATER_THAN,
		TOKEN_IS_LESS_EQUAL, TOKEN_IS_GREATER_EQUAL, TOKEN_IS_NOT_EQUAL:
		op := p.curToken.Value
		p.nextToken()
		right, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		return &BinaryExpression{
			Left:     left,
			Operator: op,
			Right:    right,
		}, nil
	}

	return left, nil
}

func (p *Parser) parseExpression() (Expression, error) {
	return p.parseAdditive()
}

func (p *Parser) parseAdditive() (Expression, error) {
	left, err := p.parseMultiplicative()
	if err != nil {
		return nil, err
	}

	for p.curToken.Type == TOKEN_PLUS || p.curToken.Type == TOKEN_MINUS {
		op := "+"
		if p.curToken.Type == TOKEN_MINUS {
			op = "-"
		}
		p.nextToken()
		right, err := p.parseMultiplicative()
		if err != nil {
			return nil, err
		}
		left = &BinaryExpression{
			Left:     left,
			Operator: op,
			Right:    right,
		}
	}

	return left, nil
}

func (p *Parser) parseMultiplicative() (Expression, error) {
	left, err := p.parsePrimary()
	if err != nil {
		return nil, err
	}

	for p.curToken.Type == TOKEN_STAR || p.curToken.Type == TOKEN_SLASH {
		op := "*"
		if p.curToken.Type == TOKEN_SLASH {
			op = "/"
		}
		p.nextToken()
		right, err := p.parsePrimary()
		if err != nil {
			return nil, err
		}
		left = &BinaryExpression{
			Left:     left,
			Operator: op,
			Right:    right,
		}
	}

	return left, nil
}

func (p *Parser) parsePrimary() (Expression, error) {
	switch p.curToken.Type {
	case TOKEN_NUMBER:
		value, _ := strconv.ParseFloat(p.curToken.Value, 64)
		p.nextToken()
		return &NumberLiteral{Value: value}, nil

	case TOKEN_STRING:
		value := p.curToken.Value
		p.nextToken()
		return &StringLiteral{Value: value}, nil

	case TOKEN_TRUE:
		p.nextToken()
		return &BooleanLiteral{Value: true}, nil

	case TOKEN_FALSE:
		p.nextToken()
		return &BooleanLiteral{Value: false}, nil

	case TOKEN_LBRACKET:
		return p.parseList()

	case TOKEN_THE:
		// Handle "the item at position X in Y" or "the length of X" or "the remainder of X divided by Y" or "the location of X"
		p.nextToken()
		if p.curToken.Type == TOKEN_ITEM {
			return p.parseIndexExpression()
		}
		if p.curToken.Type == TOKEN_LENGTH {
			return p.parseLengthExpression()
		}
		if p.curToken.Type == TOKEN_REMAINDER {
			return p.parseRemainderExpression()
		}
		if p.curToken.Type == TOKEN_LOCATION {
			return p.parseLocationExpression()
		}
		// Fall back to treating "the" as part of other constructs
		// Put back THE token context - this is for "the value of" pattern
		if p.curToken.Type == TOKEN_VALUE {
			p.nextToken()
			if p.curToken.Type == TOKEN_OF {
				p.nextToken()
			}
			return p.parseExpression()
		}
		return nil, fmt.Errorf("unexpected token after 'the': %v at line %d", p.curToken.Type, p.curToken.Line)

	case TOKEN_ITEM:
		// "item" used as a variable name (not "the item at position")
		p.nextToken()
		return &Identifier{Name: "item"}, nil

	case TOKEN_IDENTIFIER:
		name := p.curToken.Value
		p.nextToken()

		// Check if it's a function call
		if p.curToken.Type == TOKEN_LPAREN {
			p.nextToken()
			args, err := p.parseFunctionCallArgs()
			if err != nil {
				return nil, err
			}
			if err := p.expectToken(TOKEN_RPAREN); err != nil {
				return nil, err
			}
			p.nextToken()
			return &FunctionCall{
				Name:      name,
				Arguments: args,
			}, nil
		}

		// Check if it's array indexing with brackets: list[0]
		if p.curToken.Type == TOKEN_LBRACKET {
			p.nextToken()
			index, err := p.parseExpression()
			if err != nil {
				return nil, err
			}
			if err := p.expectToken(TOKEN_RBRACKET); err != nil {
				return nil, err
			}
			p.nextToken()
			return &IndexExpression{
				List:  &Identifier{Name: name},
				Index: index,
			}, nil
		}

		return &Identifier{Name: name}, nil

	case TOKEN_LPAREN:
		p.nextToken()
		expr, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		if err := p.expectToken(TOKEN_RPAREN); err != nil {
			return nil, err
		}
		p.nextToken()
		return expr, nil

	case TOKEN_MINUS:
		p.nextToken()
		expr, err := p.parsePrimary()
		if err != nil {
			return nil, err
		}
		return &UnaryExpression{
			Operator: "-",
			Right:    expr,
		}, nil

	default:
		return nil, fmt.Errorf("unexpected token in expression: %v at line %d", p.curToken.Type, p.curToken.Line)
	}
}

// parseIndexExpression parses "item at position X in Y"
func (p *Parser) parseIndexExpression() (Expression, error) {
	// Already consumed "the", now at "item"
	if err := p.expectToken(TOKEN_ITEM); err != nil {
		return nil, err
	}
	p.nextToken()

	if err := p.expectToken(TOKEN_AT); err != nil {
		return nil, err
	}
	p.nextToken()

	if err := p.expectToken(TOKEN_POSITION); err != nil {
		return nil, err
	}
	p.nextToken()

	index, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	if err := p.expectToken(TOKEN_IN); err != nil {
		return nil, err
	}
	p.nextToken()

	list, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	return &IndexExpression{
		List:  list,
		Index: index,
	}, nil
}

// parseLengthExpression parses "length of X"
func (p *Parser) parseLengthExpression() (Expression, error) {
	// Already consumed "the", now at "length"
	if err := p.expectToken(TOKEN_LENGTH); err != nil {
		return nil, err
	}
	p.nextToken()

	if err := p.expectToken(TOKEN_OF); err != nil {
		return nil, err
	}
	p.nextToken()

	list, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	return &LengthExpression{
		List: list,
	}, nil
}

// parseRemainderExpression parses "remainder of X divided by Y" or "remainder of X / Y"
func (p *Parser) parseRemainderExpression() (Expression, error) {
	// Already consumed "the", now at "remainder"
	if err := p.expectToken(TOKEN_REMAINDER); err != nil {
		return nil, err
	}
	p.nextToken()

	if err := p.expectToken(TOKEN_OF); err != nil {
		return nil, err
	}
	p.nextToken()

	// Parse the dividend (left operand)
	left, err := p.parsePrimary()
	if err != nil {
		return nil, err
	}

	// Expect "divided by" or "/"
	if p.curToken.Type == TOKEN_DIVIDED {
		p.nextToken()
		if err := p.expectToken(TOKEN_BY); err != nil {
			return nil, err
		}
		p.nextToken()
	} else if p.curToken.Type == TOKEN_SLASH {
		p.nextToken()
	} else {
		return nil, fmt.Errorf("expected 'divided by' or '/' after remainder operand, got %v", p.curToken.Type)
	}

	// Parse the divisor (right operand)
	right, err := p.parsePrimary()
	if err != nil {
		return nil, err
	}

	return &BinaryExpression{
		Left:     left,
		Operator: "%",
		Right:    right,
	}, nil
}

// parseLocationExpression parses "location of X"
func (p *Parser) parseLocationExpression() (Expression, error) {
	// Already consumed "the", now at "location"
	if err := p.expectToken(TOKEN_LOCATION); err != nil {
		return nil, err
	}
	p.nextToken()

	if err := p.expectToken(TOKEN_OF); err != nil {
		return nil, err
	}
	p.nextToken()

	// Get the variable name
	if p.curToken.Type != TOKEN_IDENTIFIER {
		return nil, fmt.Errorf("expected variable name after 'the location of', got %v", p.curToken.Type)
	}
	name := p.curToken.Value
	p.nextToken()

	return &LocationExpression{
		Name: name,
	}, nil
}

// parseToggle parses "Toggle x." or "Toggle the value of x."
func (p *Parser) parseToggle() (Statement, error) {
	if err := p.expectToken(TOKEN_TOGGLE); err != nil {
		return nil, err
	}
	p.nextToken()

	// Handle "toggle the value of x"
	if p.curToken.Type == TOKEN_THE {
		p.nextToken()
		if p.curToken.Type == TOKEN_VALUE {
			p.nextToken()
			if p.curToken.Type == TOKEN_OF {
				p.nextToken()
			}
		}
	}

	// Get the variable name
	if p.curToken.Type != TOKEN_IDENTIFIER {
		return nil, fmt.Errorf("expected variable name after 'Toggle', got %v", p.curToken.Type)
	}
	name := p.curToken.Value
	p.nextToken()

	if err := p.expectToken(TOKEN_PERIOD); err != nil {
		return nil, err
	}
	p.nextToken()

	return &ToggleStatement{
		Name: name,
	}, nil
}

func (p *Parser) parseList() (Expression, error) {
	if err := p.expectToken(TOKEN_LBRACKET); err != nil {
		return nil, err
	}
	p.nextToken()

	var elements []Expression

	if p.curToken.Type != TOKEN_RBRACKET {
		for {
			elem, err := p.parseExpression()
			if err != nil {
				return nil, err
			}
			elements = append(elements, elem)

			if p.curToken.Type != TOKEN_COMMA {
				break
			}
			p.nextToken()
		}
	}

	if err := p.expectToken(TOKEN_RBRACKET); err != nil {
		return nil, err
	}
	p.nextToken()

	return &ListLiteral{Elements: elements}, nil
}

func (p *Parser) parseFunctionArguments() ([]Expression, error) {
	var args []Expression

	if p.curToken.Type == TOKEN_WITH {
		p.nextToken()
		for {
			arg, err := p.parseExpression()
			if err != nil {
				return nil, err
			}
			args = append(args, arg)

			if p.curToken.Type != TOKEN_AND {
				break
			}
			p.nextToken()
		}
	}

	return args, nil
}

func (p *Parser) parseFunctionCallArgs() ([]Expression, error) {
	var args []Expression

	if p.curToken.Type != TOKEN_RPAREN {
		for {
			arg, err := p.parseExpression()
			if err != nil {
				return nil, err
			}
			args = append(args, arg)

			if p.curToken.Type != TOKEN_COMMA {
				break
			}
			p.nextToken()
		}
	}

	return args, nil
}
