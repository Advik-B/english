package parser

import (
	"english/ast"
	"english/token"
	"fmt"
)

// parseStructDeclaration parses a struct declaration
// Syntax: declare Person as a structure with the following fields:
//           name is a string.
//           age is an unsigned integer with 18 being the default.
//           let talk be a function that does the following:
//               print "hello, my name is", name.
//           thats it.
//         thats it.
func (p *Parser) parseStructDeclaration() (ast.Statement, error) {
	nameToken := p.curToken
	if p.curToken.Type != token.IDENTIFIER {
		return nil, fmt.Errorf("expected struct name after 'Declare', got %v", p.curToken.Type)
	}
	p.nextToken()

	// Expect "as"
	if err := p.expectToken(token.AS); err != nil {
		return nil, err
	}
	p.nextToken()

	// Skip optional "a" or "an"
	if p.curToken.Type == token.IDENTIFIER && (p.curToken.Value == "a" || p.curToken.Value == "an") {
		p.nextToken()
	}

	// Expect "structure" or "struct"
	if p.curToken.Type != token.STRUCTURE && p.curToken.Type != token.STRUCT {
		return nil, fmt.Errorf("expected 'structure' or 'struct', got %v", p.curToken.Type)
	}
	p.nextToken()

	// Expect "with"
	if err := p.expectToken(token.WITH); err != nil {
		return nil, err
	}
	p.nextToken()

	// Skip "the"
	if p.curToken.Type == token.THE {
		p.nextToken()
	}

	// Expect "following"
	if err := p.expectToken(token.FOLLOWING); err != nil {
		return nil, err
	}
	p.nextToken()

	// Expect "fields" or "field"
	if p.curToken.Type != token.FIELDS && p.curToken.Type != token.FIELD {
		return nil, fmt.Errorf("expected 'fields' or 'field', got %v", p.curToken.Type)
	}
	p.nextToken()

	// Expect ":"
	if err := p.expectToken(token.COLON); err != nil {
		return nil, err
	}
	p.nextToken()

	// Skip optional newline
	if p.curToken.Type == token.NEWLINE {
		p.nextToken()
	}

	// Parse fields and methods
	var fields []*ast.StructField
	var methods []*ast.FunctionDecl

	for p.curToken.Type != token.THATS && p.curToken.Type != token.EOF {
		// Skip newlines and indentation
		for p.curToken.Type == token.NEWLINE {
			p.nextToken()
		}

		if p.curToken.Type == token.THATS {
			break
		}

		// Check if it's a method (let functionName be a function...)
		if p.curToken.Type == token.LET {
			method, err := p.parseStructMethod()
			if err != nil {
				return nil, err
			}
			methods = append(methods, method)
		} else {
			// Parse field
			field, err := p.parseStructField()
			if err != nil {
				return nil, err
			}
			fields = append(fields, field)
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

	return &ast.StructDecl{
		Name:    nameToken.Value,
		Fields:  fields,
		Methods: methods,
	}, nil
}

// parseStructField parses a field declaration within a struct
// Syntax: name is a string.
//         age is an unsigned integer with 18 being the default.
func (p *Parser) parseStructField() (*ast.StructField, error) {
	nameToken := p.curToken
	if p.curToken.Type != token.IDENTIFIER {
		return nil, fmt.Errorf("expected field name, got %v", p.curToken.Type)
	}
	p.nextToken()

	// Expect "is"
	if err := p.expectToken(token.IS); err != nil {
		return nil, err
	}
	p.nextToken()

	// Skip optional "a" or "an"
	if p.curToken.Type == token.IDENTIFIER && (p.curToken.Value == "a" || p.curToken.Value == "an") {
		p.nextToken()
	}

	// Check for "unsigned"
	isUnsigned := false
	if p.curToken.Type == token.UNSIGNED {
		isUnsigned = true
		p.nextToken()
	}

	// Get type name
	typeToken := p.curToken
	if p.curToken.Type != token.IDENTIFIER && p.curToken.Type != token.INTEGER {
		return nil, fmt.Errorf("expected type name, got %v", p.curToken.Type)
	}
	typeName := typeToken.Value
	if p.curToken.Type == token.INTEGER {
		typeName = "integer"
	}
	p.nextToken()

	var defaultValue ast.Expression

	// Check for default value: "with X being the default"
	if p.curToken.Type == token.WITH {
		p.nextToken()

		// Parse default value expression
		val, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		defaultValue = val

		// Expect "being"
		if p.curToken.Type == token.IDENTIFIER && p.curToken.Value == "being" {
			p.nextToken()
		}

		// Skip "the"
		if p.curToken.Type == token.THE {
			p.nextToken()
		}

		// Expect "default"
		if err := p.expectToken(token.DEFAULT); err != nil {
			return nil, err
		}
		p.nextToken()
	}

	// Expect period
	if err := p.expectToken(token.PERIOD); err != nil {
		return nil, err
	}
	p.nextToken()

	return &ast.StructField{
		Name:         nameToken.Value,
		TypeName:     typeName,
		DefaultValue: defaultValue,
		IsUnsigned:   isUnsigned,
	}, nil
}

// parseStructMethod parses a method within a struct definition
func (p *Parser) parseStructMethod() (*ast.FunctionDecl, error) {
	// Skip "let"
	p.nextToken()

	nameToken := p.curToken
	if p.curToken.Type != token.IDENTIFIER {
		return nil, fmt.Errorf("expected method name, got %v", p.curToken.Type)
	}
	p.nextToken()

	// Expect "be" or "to be"
	if p.curToken.Type == token.TO {
		p.nextToken()
	}

	if err := p.expectToken(token.BE); err != nil {
		return nil, err
	}
	p.nextToken()

	// Skip optional "a"
	if p.curToken.Type == token.IDENTIFIER && p.curToken.Value == "a" {
		p.nextToken()
	}

	// Expect "function"
	if err := p.expectToken(token.FUNCTION); err != nil {
		return nil, err
	}
	p.nextToken()

	var parameters []string

	// Check for "that takes" for parameters
	if p.curToken.Type == token.THAT {
		p.nextToken()
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
	}

	// Support "and does" syntax after parameters
	if p.curToken.Type == token.AND {
		p.nextToken()
	}

	// Expect "that does" or just "does"
	if p.curToken.Type == token.THAT {
		p.nextToken()
	}

	if err := p.expectToken(token.DOES); err != nil {
		return nil, err
	}
	p.nextToken()

	// Skip "the"
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

	// Parse function body
	body, err := p.parseBlock()
	if err != nil {
		return nil, err
	}

	return &ast.FunctionDecl{
		Name:       nameToken.Value,
		Parameters: parameters,
		Body:       body,
	}, nil
}
