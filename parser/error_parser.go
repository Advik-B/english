package parser

import (
	"english/ast"
	"english/token"
	"fmt"
)

// parseTryStatement parses a try/on error/but finally block
// Syntax: try doing the following:
//           print 10 / 0.
//         on error:
//           print error.
//         but finally:
//           print "this is always executed".
//         thats it.
func (p *Parser) parseTryStatement() (ast.Statement, error) {
	// Skip "try"
	p.nextToken()

	// Expect "doing"
	if err := p.expectToken(token.DOING); err != nil {
		return nil, err
	}
	p.nextToken()

	// Skip optional "the"
	if p.curToken.Type == token.THE {
		p.nextToken()
	}

	// Expect "following"
	if err := p.expectToken(token.FOLLOWING); err != nil {
		return nil, err
	}
	p.nextToken()

	// Expect ":"
	if err := p.expectToken(token.COLON); err != nil {
		return nil, err
	}
	p.nextToken()

	// Parse try body
	tryBody, err := p.parseBlock()
	if err != nil {
		return nil, err
	}

	// The parseBlock consumes up to "thats" but not including it
	// We need to check if we have "on error:" next
	
	var errorBody []ast.Statement
	var finallyBody []ast.Statement
	errorVar := "error" // Default error variable name

	// Check for "on error:"
	if p.curToken.Type == token.ON {
		p.nextToken()

		if err := p.expectToken(token.ONERROR); err != nil {
			return nil, err
		}
		p.nextToken()

		// Expect ":"
		if err := p.expectToken(token.COLON); err != nil {
			return nil, err
		}
		p.nextToken()

		// Parse error handling body
		errorBody, err = p.parseBlock()
		if err != nil {
			return nil, err
		}
	}

	// Check for "but finally:"
	if p.curToken.Type == token.BUT {
		p.nextToken()

		if err := p.expectToken(token.FINALLY); err != nil {
			return nil, err
		}
		p.nextToken()

		// Expect ":"
		if err := p.expectToken(token.COLON); err != nil {
			return nil, err
		}
		p.nextToken()

		// Parse finally body
		finallyBody, err = p.parseBlock()
		if err != nil {
			return nil, err
		}
	}

	// Expect "thats it."
	if err := p.expectToken(token.THATS); err != nil {
		return nil, err
	}
	p.nextToken()

	if err := p.expectToken(token.IT); err != nil {
		return nil, err
	}
	p.nextToken()

	if err := p.expectToken(token.PERIOD); err != nil {
		return nil, err
	}
	p.nextToken()

	return &ast.TryStatement{
		TryBody:     tryBody,
		ErrorVar:    errorVar,
		ErrorBody:   errorBody,
		FinallyBody: finallyBody,
	}, nil
}

// parseRaiseStatement parses a raise statement
// Syntax: raise "10 / 0 SHOULD NOT COMPUTE" as RuntimeError.
//         raise "error message".
func (p *Parser) parseRaiseStatement() (ast.Statement, error) {
	// Skip "raise"
	p.nextToken()

	// Parse error message expression
	message, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	errorType := "RuntimeError" // Default

	// Check for "as ErrorType"
	if p.curToken.Type == token.AS {
		p.nextToken()

		if p.curToken.Type != token.IDENTIFIER {
			return nil, fmt.Errorf("expected error type after 'as', got %v", p.curToken.Type)
		}

		errorType = p.curToken.Value
		p.nextToken()
	}

	// Expect period
	if err := p.expectToken(token.PERIOD); err != nil {
		return nil, err
	}
	p.nextToken()

	return &ast.RaiseStatement{
		Message:   message,
		ErrorType: errorType,
	}, nil
}

// parseSwapStatement parses a swap statement
// Syntax: swap a and b.
func (p *Parser) parseSwapStatement() (ast.Statement, error) {
	// Skip "swap"
	p.nextToken()

	if p.curToken.Type != token.IDENTIFIER {
		return nil, fmt.Errorf("expected variable name after 'swap', got %v", p.curToken.Type)
	}
	name1 := p.curToken.Value
	p.nextToken()

	// Expect "and"
	if err := p.expectToken(token.AND); err != nil {
		return nil, err
	}
	p.nextToken()

	if p.curToken.Type != token.IDENTIFIER {
		return nil, fmt.Errorf("expected variable name after 'and', got %v", p.curToken.Type)
	}
	name2 := p.curToken.Value
	p.nextToken()

	// Expect period
	if err := p.expectToken(token.PERIOD); err != nil {
		return nil, err
	}
	p.nextToken()

	return &ast.SwapStatement{
		Name1: name1,
		Name2: name2,
	}, nil
}
