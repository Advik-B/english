package parser

import (
	"english/ast"
	"english/token"
	"fmt"
	"strings"
)

// parseTryStatement parses a try/on error/but finally block
// Syntax: try doing the following:
//
//	  print 10 / 0.
//	on error:
//	  print error.
//	but finally:
//	  print "this is always executed".
//	thats it.
func (p *Parser) parseTryStatement() (ast.Statement, error) {
	// Skip "try"
	startLine := p.curToken.Line
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
	errorType := ""     // Empty means catch all errors

	// Check for "on <error|TypeName>:"
	// - "on error:"        catches all errors (backward-compatible)
	// - "on NetworkError:" catches only errors with ErrorType == "NetworkError"
	if p.curToken.Type == token.ON {
		p.nextToken()

		// Accept any identifier: "error" (catch-all) or a specific type name
		if p.curToken.Type != token.IDENTIFIER {
			return nil, p.syntaxErr(
				msgErrorTypeOnName,
				hintOnError,
			)
		}
		handlerName := p.curToken.Value
		p.nextToken()

		if strings.ToLower(handlerName) != "error" {
			// Type-specific catch: only catch errors whose ErrorType matches
			errorType = handlerName
		}

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
		ErrorType:   errorType,
		ErrorBody:   errorBody,
		FinallyBody: finallyBody,
		Line:        startLine,
	}, nil
}

// parseRaiseStatement parses a raise statement
// Syntax: raise "10 / 0 SHOULD NOT COMPUTE" as RuntimeError.
//
//	raise "error message".
func (p *Parser) parseRaiseStatement() (ast.Statement, error) {
	// Skip "raise"
	startLine := p.curToken.Line
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
			return nil, p.syntaxErr(
				msgRaiseErrorType,
				hintRaiseAs,
			)
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
		Line:      startLine,
	}, nil
}

// parseSwapStatement parses a swap statement
// Syntax: swap a and b.
func (p *Parser) parseSwapStatement() (ast.Statement, error) {
	// Skip "swap"
	startLine := p.curToken.Line
	p.nextToken()

	if p.curToken.Type != token.IDENTIFIER {
		return nil, p.syntaxErr(
			msgSwapFirstVar,
			hintSwapVars,
		)
	}
	name1 := p.curToken.Value
	p.nextToken()

	// Expect "and"
	if err := p.expectToken(token.AND); err != nil {
		return nil, err
	}
	p.nextToken()

	if p.curToken.Type != token.IDENTIFIER {
		return nil, p.syntaxErr(
			msgSwapSecondVar,
			hintSwapVars,
		)
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
		Line:  startLine,
	}, nil
}

// parseErrorTypeDecl parses a root custom error type declaration.
// Syntax: Declare NetworkError as an error type.
// This is called from parseDeclareAs when "an/a error type" is detected.
func (p *Parser) parseErrorTypeDecl() (ast.Statement, error) {
	nameToken := p.curToken
	if nameToken.Type != token.IDENTIFIER {
		return nil, p.syntaxErr(
			msgErrorTypeName,
			hintErrorTypeDecl,
		)
	}
	p.nextToken() // consume name

	// Consume "as"
	if err := p.expectToken(token.AS); err != nil {
		return nil, err
	}
	p.nextToken()

	// Consume "an" or "a"
	if p.curToken.Type != token.IDENTIFIER || (strings.ToLower(p.curToken.Value) != "a" && strings.ToLower(p.curToken.Value) != "an") {
		return nil, p.syntaxErr(
			fmt.Sprintf(msgFmtArticleAfterAs, p.curToken.Value),
			hintErrorTypeDecl,
		)
	}
	p.nextToken()

	// Consume "error"
	if p.curToken.Type != token.IDENTIFIER || strings.ToLower(p.curToken.Value) != "error" {
		return nil, p.syntaxErr(
			fmt.Sprintf(msgFmtExpectedErrorWord, p.curToken.Value),
			hintErrorTypeDecl,
		)
	}
	p.nextToken()

	// Consume "type"
	if err := p.expectToken(token.TYPE); err != nil {
		return nil, err
	}
	p.nextToken()

	// Expect period
	if err := p.expectToken(token.PERIOD); err != nil {
		return nil, err
	}
	p.nextToken()

	return &ast.ErrorTypeDecl{
		Name: nameToken.Value,
	}, nil
}

// parseErrorSubtypeDecl parses an error subtype declaration.
// Syntax: Declare CustomErr1 as a type of CustomLibError.
// This is called from parseDeclareAs when "a type of" is detected.
func (p *Parser) parseErrorSubtypeDecl() (ast.Statement, error) {
	nameToken := p.curToken
	if nameToken.Type != token.IDENTIFIER {
		return nil, p.syntaxErr(
			msgErrorSubtypeName,
			hintErrorSubtypeDecl,
		)
	}
	p.nextToken() // consume name

	// Consume "as"
	if err := p.expectToken(token.AS); err != nil {
		return nil, err
	}
	p.nextToken()

	// Consume "a" or "an"
	if p.curToken.Type != token.IDENTIFIER || (strings.ToLower(p.curToken.Value) != "a" && strings.ToLower(p.curToken.Value) != "an") {
		return nil, p.syntaxErr(
			fmt.Sprintf(msgFmtArticleAfterAs, p.curToken.Value),
			hintErrorSubtypeDecl,
		)
	}
	p.nextToken()

	// Consume "type"
	if err := p.expectToken(token.TYPE); err != nil {
		return nil, err
	}
	p.nextToken()

	// Consume "of"
	if err := p.expectToken(token.OF); err != nil {
		return nil, err
	}
	p.nextToken()

	// Consume parent type name
	if p.curToken.Type != token.IDENTIFIER {
		return nil, p.syntaxErr(
			msgErrorParentType,
			hintErrorSubtypeDecl,
		)
	}
	parentName := p.curToken.Value
	p.nextToken()

	// Expect period
	if err := p.expectToken(token.PERIOD); err != nil {
		return nil, err
	}
	p.nextToken()

	return &ast.ErrorTypeDecl{
		Name:       nameToken.Value,
		ParentType: parentName,
	}, nil
}
