package parser

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/Advik-B/english/ast"
	"github.com/Advik-B/english/token"
	"github.com/Advik-B/english/tokeniser"
	"github.com/Advik-B/english/types"
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
	// blockAtEOF records that a block body ran into the end of the input
	// without being closed. It answers one question, for an interactive
	// prompt: is more of this program still to come? See markTruncated.
	blockAtEOF bool
}

// NewParser creates a new parser for the given tokens
func NewParser(tokens []token.Token) *Parser {
	p := &Parser{tokens: tokens, position: 0}
	p.nextToken()
	p.nextToken()
	return p
}

// isWord reports whether the current token is the given filler word, matched
// case-insensitively. English has a number of words that read naturally but
// are not keywords — "being", "gives", "back", "a" — and they were previously
// compared with a mix of ==, strings.ToLower and strings.EqualFold, so some
// were accidentally case-sensitive.
func (p *Parser) isWord(word string) bool {
	return p.curToken.Type == token.IDENTIFIER && strings.EqualFold(p.curToken.Value, word)
}

// skipWord consumes the current token if it is the given filler word.
func (p *Parser) skipWord(word string) bool {
	if p.isWord(word) {
		p.nextToken()
		return true
	}
	return false
}

// peekWord reports whether the lookahead token is the given filler word.
func (p *Parser) peekWord(word string) bool {
	return p.peekToken.Type == token.IDENTIFIER && strings.EqualFold(p.peekToken.Value, word)
}

// skipOptional consumes the current token if it has the given type.
func (p *Parser) skipOptional(t token.Type) bool {
	if p.curToken.Type == t {
		p.nextToken()
		return true
	}
	return false
}

// at converts a token's location into a node position. Every AST node embeds
// ast.Base, so this is the single place the parser records where a node starts.
func at(t token.Token) ast.Base {
	return ast.At(t.Line, t.Col, t.Pos)
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
		hint = hintEndWithPeriod
	case token.TO:
		if p.curToken.Type == token.BE {
			hint = "Did you mean 'to be'? For example: 'Declare x to be 5.'"
		} else {
			hint = hintExpectedToBe
		}
	case token.BE:
		if p.curToken.Type == token.TO {
			hint = hintMissingBeAfterTo
		} else {
			hint = hintExpectedBe
		}
	case token.THATS:
		hint = hintCloseThatsWith
	case token.IT:
		if p.curToken.Type == token.PERIOD {
			hint = hintMissingIt
		}
	case token.IDENTIFIER:
		if p.curToken.Type == token.NUMBER || p.curToken.Type == token.STRING {
			hint = hintNameNotLiteral
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

// maxReportedSyntaxErrors caps a single parse's report. Past a certain point
// the later errors are consequences of the earlier ones rather than separate
// mistakes, and a wall of them is less use than a handful.
const maxReportedSyntaxErrors = 20

// Parse parses the tokens and returns the AST.
//
// It reports every syntax error it can find, not just the first. Stopping at
// the first meant a file with two typos took two runs to fix, and the editor —
// which shows one diagnostic per parse and gives up before extracting any
// symbols — showed nothing else about a file until the last syntax error in it
// was gone.
func (p *Parser) Parse() (*ast.Program, error) {
	program := &ast.Program{}
	var errs SyntaxErrors

	for p.curToken.Type != token.EOF {
		// Check whether the upcoming statement starts with a PLEASE prefix.
		// Consume it here so we can reliably read the actual statement keyword's
		// line number (curToken.Line after the consume) for the politeness tally.
		polite := p.curToken.Type == token.PLEASE
		if polite {
			p.nextToken() // consume PLEASE; curToken is now the statement keyword
		}
		stmtStartLine := p.curToken.Line

		// parseStatement() also handles PLEASE (for inner blocks), but since we
		// already consumed it above, curToken is no longer PLEASE here.
		stmt, err := p.parseStatement()
		if err != nil {
			err = p.markTruncated(err)

			// Input that simply ran out is not something to recover from:
			// there is nothing after it to resynchronise on, and an
			// interactive prompt needs to see it as "more is coming".
			var syntaxErr *SyntaxError
			if IsTruncated(err) || !errors.As(err, &syntaxErr) {
				return nil, err
			}

			errs = append(errs, syntaxErr)
			if len(errs) >= maxReportedSyntaxErrors {
				return nil, errs
			}
			p.synchronise()
			continue
		}
		program.Statements = append(program.Statements, stmt)
		// Comments don't count toward the politeness tally.
		if _, isComment := stmt.(*ast.CommentStatement); !isComment {
			program.TotalCount++
			if polite {
				program.PoliteCount++
			} else {
				line := stmt.Pos().Line
				if line == 0 {
					line = stmtStartLine
				}
				program.ImpoliteLines = append(program.ImpoliteLines, line)
			}
		}
	}

	if len(errs) > 0 {
		return nil, errs
	}
	return program, nil
}

// synchronise skips to where the next statement plausibly begins, after an
// error, so that one mistake does not hide every later one.
//
// A statement ends with a period, so the token after one is the natural place
// to resume; failing that, a keyword that can only begin a statement will do.
// It always consumes at least one token, or a parser that failed on a
// statement keyword would fail on it again forever.
func (p *Parser) synchronise() {
	p.nextToken()
	for p.curToken.Type != token.EOF {
		if p.curToken.Type == token.PERIOD {
			p.nextToken()
			return
		}
		if startsStatement(p.curToken.Type) {
			return
		}
		p.nextToken()
	}
}

// startsStatement reports whether a token can only appear at the start of a
// statement, which makes it a safe place to resume parsing after an error.
func startsStatement(t token.Type) bool {
	switch t {
	case token.PLEASE, token.COMMENT, token.IMPORT, token.DECLARE, token.LET,
		token.BREAK, token.CONTINUE, token.SKIP, token.ASK, token.SET,
		token.CALL, token.IF, token.REPEAT, token.FOR, token.PRINT,
		token.WRITE, token.RETURN, token.TOGGLE, token.TRY, token.RAISE,
		token.SWAP, token.SLEEP:
		return true
	}
	return false
}

// errorTokenErr turns a lexer ERROR token into a syntax error. The lexer emits
// one for an unterminated text literal or an unrecognised character; before
// this the parser had no case for it and reported a confusing generic message.
func (p *Parser) errorTokenErr() error {
	if p.curToken.Value == tokeniser.UnterminatedString {
		return p.syntaxErr(msgUnterminatedText, hintUnterminatedText)
	}
	return p.syntaxErr(
		fmt.Sprintf(msgFmtIllegalChar, p.curToken.Value),
		hintIllegalChar,
	)
}

// expectBlockEnd consumes the "thats it." that closes a block.
//
// This epilogue was previously written out ten times: six copies in this file
// guarded by "if p.curToken.Type == token.THATS", which made the closer
// *optional*, and four mandatory copies elsewhere. Because parseBlock also
// stops at EOF without complaint, a missing "thats it." was not an error at
// all — the statements that followed were silently absorbed into the block:
//
//	Declare function f that does the following:
//	    Print 1.
//	Print 2.            <- silently became part of f's body
//
// It is now required everywhere, in one place.
func (p *Parser) expectBlockEnd() error {
	if err := p.expectBlockEndNoPeriod(); err != nil {
		return err
	}
	if err := p.expectToken(token.PERIOD); err != nil {
		return err
	}
	p.nextToken()
	return nil
}

// expectBlockEndNoPeriod consumes "thats it" without the trailing period.
// Used where the block closes an expression — a struct instantiation — and the
// period belongs to the enclosing statement.
func (p *Parser) expectBlockEndNoPeriod() error {
	if err := p.expectToken(token.THATS); err != nil {
		return err
	}
	p.nextToken()
	if err := p.expectToken(token.IT); err != nil {
		return err
	}
	p.nextToken()
	return nil
}

// parseStatement parses one statement and records where it started.
//
// Stamping here covers every statement kind from one place, including the ones
// that carried no position at all: imports, struct and error-type
// declarations, break, continue and comments.
func (p *Parser) parseStatement() (ast.Statement, error) {
	start := at(p.curToken)
	stmt, err := p.parseStatementInner()
	if err != nil {
		return nil, err
	}
	ast.SetPosIfUnknown(stmt, start.Position)
	return stmt, nil
}

func (p *Parser) parseStatementInner() (ast.Statement, error) {

	// A politeness prefix (please / kindly / could you / would you kindly) may
	// appear inside blocks (loops, function bodies, if-branches) as well as at
	// the top level.  Consuming it here means every call site automatically
	// supports polite statements without extra plumbing.  The prefix has no
	// semantic effect on execution; it is only counted at the top level in
	// Parse() for --minimum-politeness enforcement.
	if p.curToken.Type == token.PLEASE {
		p.nextToken()
	}

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
	case token.SLEEP:
		return p.parseSleepStatement()
	default:
		switch p.curToken.Type {
		case token.ERROR:
			return nil, p.errorTokenErr()
		case token.IDENTIFIER:
			name := p.curToken.Value
			return nil, &SyntaxError{
				Msg:  fmt.Sprintf(msgFmtIdentifierStatement, name),
				Line: p.curToken.Line,
				Col:  p.curToken.Col,
				Hint: fmt.Sprintf(hintFmtIdentifierStatement, name, name),
			}
		case token.NUMBER:
			return nil, &SyntaxError{
				Msg:  fmt.Sprintf(msgFmtNumberStatement, p.curToken.Value),
				Line: p.curToken.Line,
				Col:  p.curToken.Col,
				Hint: hintNumberAsStatement,
			}
		case token.STRING:
			return nil, &SyntaxError{
				Msg:  fmt.Sprintf(msgFmtStringStatement, p.curToken.Value),
				Line: p.curToken.Line,
				Col:  p.curToken.Col,
				Hint: hintStringAsStatement,
			}
		case token.EOF:
			return nil, &SyntaxError{
				Msg:  "The program ended unexpectedly.",
				Line: p.curToken.Line,
				Col:  p.curToken.Col,
				Hint: hintUnexpectedEOF,
			}
		default:
			return nil, &SyntaxError{
				Msg:  fmt.Sprintf(msgFmtUnknownToken, p.curToken.Value),
				Line: p.curToken.Line,
				Col:  p.curToken.Col,
				Hint: hintUnknownKeyword,
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
			msgVarNameExpected,
			hintVarNameAfterLet,
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
			fmt.Sprintf(msgFmtLetAfterName, nameToken.Value, p.curToken.Value),
			fmt.Sprintf(hintFmtLetDeclaration, nameToken.Value, nameToken.Value),
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
		Base:       at(nameToken),
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
			msgImportPath,
			hintImportPath,
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
			msgDeclareVarName,
			hintVarNameAfterDeclare,
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
		Base:       at(nameToken),
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
	funcPos := at(p.curToken)
	p.nextToken()

	nameToken := p.curToken
	if p.curToken.Type != token.IDENTIFIER {
		return nil, p.syntaxErr(
			msgFunctionNameExpected,
			hintFunctionName,
		)
	}
	p.nextToken()

	var parameters []ast.Param

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
					msgParameterName,
					hintParameterName,
				)
			}
			param := ast.Param{Base: at(paramToken), Name: paramToken.Value}
			p.nextToken()

			// Optional annotation: "takes x as number".
			if p.curToken.Type == token.AS {
				p.nextToken()
				paramType, err := p.parseTypeName()
				if err != nil {
					return nil, err
				}
				param.Type = paramType
			}
			parameters = append(parameters, param)

			if p.curToken.Type != token.AND {
				break
			}
			// "and" here either joins another parameter or introduces the rest
			// of the declaration ("and gives back …", "and does …").
			if p.peekToken.Type == token.DOES || p.peekWord("gives") {
				break
			}
			p.nextToken()
		}
	}

	// Support "and does" / "and gives back" syntax after parameters
	p.skipOptional(token.COMMA)
	p.skipOptional(token.AND)

	// Optional return type: "gives back a number".
	var returnType *ast.TypeExpr
	if p.skipWord("gives") {
		if !p.skipWord("back") {
			return nil, p.syntaxErr(msgGivesNeedsBack, hintReturnType)
		}
		var err error
		returnType, err = p.parseTypeName()
		if err != nil {
			return nil, err
		}
		p.skipOptional(token.COMMA)
		p.skipOptional(token.AND)
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

	if err := p.expectBlockEnd(); err != nil {
		return nil, err
	}

	return &ast.FunctionDecl{
		Base:       funcPos,
		Name:       nameToken.Value,
		Params:     parameters,
		ReturnType: returnType,
		Body:       body,
	}, nil
}

func (p *Parser) parseAssignment() (ast.Statement, error) {
	if err := p.expectToken(token.SET); err != nil {
		return nil, err
	}
	setPos := at(p.curToken)
	p.nextToken()

	// Check for "Set the item at position X in Y to be Z"
	// or "Set the entry KEY in TABLE to be VALUE"
	if p.curToken.Type == token.THE {
		p.nextToken()
		if p.curToken.Type == token.ITEM {
			return p.parseIndexAssignment(setPos)
		}
		if p.curToken.Type == token.ENTRY {
			return p.parseLookupKeyAssignment(setPos)
		}
		return nil, p.syntaxErr(
			fmt.Sprintf(msgFmtSetThe, p.curToken.Value),
			hintSetTheFull,
		)
	}

	nameToken := p.curToken
	if p.curToken.Type != token.IDENTIFIER {
		return nil, p.syntaxErr(
			msgSetVarName,
			hintSetVarName,
		)
	}
	p.nextToken()

	// "Set PERSON's name to be VALUE." — write a struct field.
	//
	// ast.FieldAssignment was implemented by both engines, the transpiler and
	// the disassembler, and the parser never built one, so there was no way to
	// change a field from source at all: a structure could be created and read
	// but not modified.
	if p.curToken.Type == token.POSSESSIVE {
		p.nextToken() // consume 's
		if p.curToken.Type != token.IDENTIFIER && !token.IsKeyword(p.curToken.Type) {
			return nil, p.syntaxErr(msgSetFieldName, hintSetField)
		}
		fieldName := p.curToken.Value
		p.nextToken()

		if err := p.expectToken(token.TO); err != nil {
			return nil, err
		}
		p.nextToken()
		p.skipOptional(token.BE)

		value, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		if err := p.expectToken(token.PERIOD); err != nil {
			return nil, err
		}
		p.nextToken()

		return &ast.FieldAssignment{
			ObjectName: nameToken.Value,
			Field:      fieldName,
			Value:      value,
			Base:       setPos,
		}, nil
	}

	// "Set TABLE at KEY to be VALUE." — lookup table shorthand write
	if p.curToken.Type == token.AT {
		p.nextToken() // consume AT
		key, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		if p.curToken.Type != token.TO {
			return nil, p.syntaxErr(
				fmt.Sprintf(msgFmtSetAtTo, nameToken.Value, p.curToken.Value),
				hintSetTableFull,
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
		return &ast.LookupKeyAssignment{TableName: nameToken.Value, Key: key, Value: value, Base: setPos}, nil
	}

	if err := p.expectToken(token.TO); err != nil {
		return nil, err
	}
	p.nextToken()

	// "be" is optional - "set x to 10" and "set x to be 10" are both valid
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

	return &ast.Assignment{
		Name:  nameToken.Value,
		Value: value,
		Base:  setPos,
	}, nil
}

// parseIndexAssignment parses "the item at position X in Y to be Z"
func (p *Parser) parseIndexAssignment(setPos ast.Base) (ast.Statement, error) {
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
			msgSetListName,
			hintSetListName,
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
		Base:     setPos,
	}, nil
}

func (p *Parser) parseCall() (ast.Statement, error) {
	if err := p.expectToken(token.CALL); err != nil {
		return nil, err
	}
	callPos := at(p.curToken)
	p.nextToken()

	// First identifier could be:
	// 1. Function name: "call greet with args."
	// 2. Method name: "call talk from p2." or "call talk on p2."
	// 3. Object name with possessive: "call p2's talk."

	firstIdent := p.curToken.Value
	if p.curToken.Type != token.IDENTIFIER {
		return nil, p.syntaxErr(
			msgCallName,
			hintCallName,
		)
	}
	p.nextToken()

	// Possessive syntax: "call p2's talk."
	//
	// This used to test whether the identifier's text ended in "'s", because
	// the lexer folded the possessive into the name here while emitting a
	// POSSESSIVE token everywhere else. The two spellings had drifted: this
	// path required a plain name for the method where the expression path
	// accepted a keyword too, so "Print x's length." worked and
	// "Call x's length." did not.
	if p.curToken.Type == token.POSSESSIVE {
		p.nextToken() // consume 's
		methodCall, err := p.parseMethodAfterPossessive(
			&ast.Identifier{Base: callPos, Name: firstIdent})
		if err != nil {
			return nil, err
		}

		if err := p.expectToken(token.PERIOD); err != nil {
			return nil, err
		}
		p.nextToken()

		return &ast.CallStatement{MethodCall: methodCall, Base: callPos}, nil
	}

	// Check for "from" or "on" (method call syntax)
	if p.curToken.Type == token.FROM || p.curToken.Type == token.ON {
		methodName := firstIdent
		p.nextToken() // skip FROM/ON

		// Get object
		if p.curToken.Type != token.IDENTIFIER {
			return nil, p.syntaxErr(
				fmt.Sprintf(msgFmtCallFromOn, methodName),
				hintCallName,
			)
		}
		objectName := p.curToken.Value
		p.nextToken()

		args, err := p.parseWithArguments()
		if err != nil {
			return nil, err
		}

		if err := p.expectToken(token.PERIOD); err != nil {
			return nil, err
		}
		p.nextToken()

		// Return as method call
		return &ast.CallStatement{
			MethodCall: &ast.MethodCall{
				Object:     &ast.Identifier{Base: callPos, Name: objectName},
				MethodName: methodName,
				Arguments:  args,
			},
			Base: callPos,
		}, nil
	}

	// Regular function call: "call greet with args."
	funcName := firstIdent
	args, err := p.parseWithArguments()
	if err != nil {
		return nil, err
	}

	if err := p.expectToken(token.PERIOD); err != nil {
		return nil, err
	}
	p.nextToken()

	return &ast.CallStatement{
		Base: callPos,
		FunctionCall: &ast.FunctionCall{
			Base:      callPos,
			Name:      funcName,
			Arguments: args,
		},
	}, nil
}

// parseArgumentList reads the arguments of a call written in English, after
// the "with": "with a and b", or "with a, b".
//
// There were two of these for the same construct, with different separator
// rules — one accepted a comma and the other only "and" — and different error
// handling. This one discarded the error and returned whatever it had parsed
// so far, so "Call f with ." became "Call f." and the mistake disappeared.
// The parenthesised form f(a, b) stays separate, because inside parentheses
// "and" is an operator rather than a separator.
func (p *Parser) parseArgumentList() ([]ast.Expression, error) {
	var args []ast.Expression

	for {
		arg, err := p.parseArgument()
		if err != nil {
			return nil, err
		}
		args = append(args, arg)

		if p.curToken.Type != token.AND && p.curToken.Type != token.COMMA {
			return args, nil
		}
		p.nextToken()
	}
}

// parseWithArguments reads an optional "with …" argument list.
func (p *Parser) parseWithArguments() ([]ast.Expression, error) {
	if p.curToken.Type != token.WITH {
		return nil, nil
	}
	p.nextToken()
	return p.parseArgumentList()
}

func (p *Parser) parseIfStatement() (ast.Statement, error) {
	if err := p.expectToken(token.IF); err != nil {
		return nil, err
	}
	startPos := at(p.curToken)
	p.nextToken()

	condition, err := p.parseExpression()
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
			eifCond, err := p.parseExpression()
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

	if err := p.expectBlockEnd(); err != nil {
		return nil, err
	}

	return &ast.IfStatement{
		Condition: condition,
		Then:      thenBody,
		ElseIf:    elseIfParts,
		Else:      elseBody,
		Base:      startPos,
	}, nil
}

func (p *Parser) parseRepeat() (ast.Statement, error) {
	if err := p.expectToken(token.REPEAT); err != nil {
		return nil, err
	}
	startPos := at(p.curToken)
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

		if err := p.expectBlockEnd(); err != nil {
			return nil, err
		}

		return &ast.WhileLoop{
			Condition: &ast.BooleanLiteral{Value: true},
			Body:      body,
			Base:      startPos,
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
		condition, err := p.parseExpression()
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

		if err := p.expectBlockEnd(); err != nil {
			return nil, err
		}

		return &ast.WhileLoop{
			Condition: condition,
			Body:      body,
			Base:      startPos,
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

	if err := p.expectBlockEnd(); err != nil {
		return nil, err
	}

	return &ast.ForLoop{
		Count: countExpr,
		Body:  body,
		Base:  startPos,
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
			msgForEachVar,
			hintForEachVar,
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

	if err := p.expectBlockEnd(); err != nil {
		return nil, err
	}

	return &ast.ForEachLoop{
		Item: itemName,
		List: listExpr,
		Body: body,
		Base: at(itemToken),
	}, nil
}

func (p *Parser) parseOutput(newline bool) (ast.Statement, error) {
	// Accept either PRINT or WRITE token
	if p.curToken.Type != token.PRINT && p.curToken.Type != token.WRITE {
		return nil, p.syntaxErr(
			msgPrintOrWrite,
			hintPrintOrWrite,
		)
	}
	startPos := at(p.curToken)
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
		Base:    startPos,
	}, nil
}

func (p *Parser) parseReturn() (ast.Statement, error) {
	if err := p.expectToken(token.RETURN); err != nil {
		return nil, err
	}
	startPos := at(p.curToken)
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
		Base:  startPos,
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
			fmt.Sprintf(msgFmtBreakTheThis, p.curToken.Value),
			hintBreakLoop,
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
//   - "Ask "prompt" as varname."
//   - "Ask "prompt" and store it in varname."
//   - "Ask "prompt" and store the answer in varname."
//   - "Ask "prompt" and store the result in varname."
func (p *Parser) parseAskStatement() (ast.Statement, error) {
	askPos := at(p.curToken)
	p.nextToken() // consume ASK

	// Parse the prompt expression
	prompt, err := p.parseArgument()
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
				msgAskVarAs,
				hintAskAs,
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
		// Skip filler words: "it" (token.IT), "the" (token.THE), or plain identifiers
		// like "answer", "result", "response"
		for p.curToken.Type == token.IDENTIFIER || p.curToken.Type == token.THE || p.curToken.Type == token.IT {
			p.nextToken()
		}
		// Expect "in"
		if p.curToken.Type == token.IN {
			p.nextToken() // consume IN
		}
		if p.curToken.Type != token.IDENTIFIER {
			return nil, p.syntaxErr(
				msgAskVarAnd,
				hintAskAnd,
			)
		}
		varName = p.curToken.Value
		p.nextToken()
	} else {
		return nil, p.syntaxErr(
			fmt.Sprintf(msgFmtAskAfter, p.curToken.Value),
			hintAskFull,
		)
	}

	if err := p.expectToken(token.PERIOD); err != nil {
		return nil, err
	}
	p.nextToken()

	// Use Assignment so it works whether the variable already exists or not.
	// The Assignment evaluator calls Set(), which creates the variable if it doesn't exist.
	// "Ask … as name." introduces name, so it is a declaration rather than
	// an assignment to something that already exists.
	return &ast.VariableDecl{
		Base:  askPos,
		Name:  varName,
		Value: &ast.AskExpression{Base: askPos, Prompt: prompt},
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
		statements = append(statements, stmt)
	}

	// Every block body goes through here, so this is where an unclosed one is
	// visible. The caller reports the missing "thats it."; recording it lets an
	// interactive prompt tell "there is more to come" from "that is wrong".
	if p.curToken.Type == token.EOF {
		p.blockAtEOF = true
	}

	return statements, nil
}

// parseOr handles "or", the loosest-binding operator.
//
// "and" used to share this level with "or", which made "a or b and c" parse as
// "(a or b) and c" instead of the conventional "a or (b and c)".
func (p *Parser) parseOr() (ast.Expression, error) {
	left, err := p.parseAnd()
	if err != nil {
		return nil, err
	}

	for p.curToken.Type == token.OR {
		opPos := at(p.curToken)
		p.nextToken()
		right, err := p.parseAnd()
		if err != nil {
			return nil, err
		}
		left = &ast.BinaryExpression{Base: opPos, Left: left, Operator: "or", Right: right}
	}

	return left, nil
}

// parseAnd handles "and", which binds tighter than "or".
func (p *Parser) parseAnd() (ast.Expression, error) {
	left, err := p.parseNot()
	if err != nil {
		return nil, err
	}

	for p.curToken.Type == token.AND {
		opPos := at(p.curToken)
		p.nextToken()
		right, err := p.parseNot()
		if err != nil {
			return nil, err
		}
		left = &ast.BinaryExpression{Base: opPos, Left: left, Operator: "and", Right: right}
	}

	return left, nil
}

// parseNot handles the "not" prefix operator.
//
// It sits above the relational layer so that "not x is equal to y" means
// "not (x is equal to y)". It used to live in parsePrimary, which bound it
// tighter than both arithmetic and comparison, so the same phrase parsed as
// "(not x) is equal to y".
func (p *Parser) parseNot() (ast.Expression, error) {
	if p.curToken.Type == token.NOT {
		opPos := at(p.curToken)
		p.nextToken()
		right, err := p.parseNot()
		if err != nil {
			return nil, err
		}
		return &ast.UnaryExpression{Base: opPos, Operator: "not", Right: right}, nil
	}
	return p.parseRelational()
}

// parseRelational handles comparison operators like "is equal to", "is less than", etc.
// parseRelational parses a comparison and records where it started.
//
// The nodes built here — comparisons, the "is true"/"is nothing" sugar, and
// error-type checks — are created after the left operand has been consumed, so
// they are stamped from the wrapper rather than inline.
func (p *Parser) parseRelational() (ast.Expression, error) {
	start := at(p.curToken)
	expr, err := p.parseRelationalExpr()
	if err != nil {
		return nil, err
	}
	ast.SetPosIfUnknown(expr, start.Position)
	return expr, nil
}

func (p *Parser) parseRelationalExpr() (ast.Expression, error) {
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

	case token.IS_TRUE:
		// "x is true" — check that x is boolean true
		p.nextToken()
		return &ast.BinaryExpression{
			Left:     left,
			Operator: "is equal to",
			Right:    &ast.BooleanLiteral{Value: true},
		}, nil

	case token.IS_FALSE:
		// "x is false" — check that x is boolean false
		p.nextToken()
		return &ast.BinaryExpression{
			Left:     left,
			Operator: "is equal to",
			Right:    &ast.BooleanLiteral{Value: false},
		}, nil

	case token.ISNT_TRUE:
		// "x isn't true" — check that x is not boolean true
		p.nextToken()
		return &ast.BinaryExpression{
			Left:     left,
			Operator: "is not equal to",
			Right:    &ast.BooleanLiteral{Value: true},
		}, nil

	case token.ISNT_FALSE:
		// "x isn't false" — check that x is not boolean false
		p.nextToken()
		return &ast.BinaryExpression{
			Left:     left,
			Operator: "is not equal to",
			Right:    &ast.BooleanLiteral{Value: false},
		}, nil

	case token.IS:
		// "error is TypeName" — error type check (exact or inherited match)
		p.nextToken()
		if p.curToken.Type != token.IDENTIFIER {
			return nil, p.syntaxErr(
				msgErrorTypeIsName,
				hintErrorTypeCheck,
			)
		}
		typeName := p.curToken.Value
		p.nextToken()
		return &ast.ErrorTypeCheckExpression{Value: left, TypeName: typeName}, nil
	}

	return left, nil
}

// parseExpression parses a complete expression, including comparisons and
// "and"/"or".
//
// It used to start at parseCast, which sits *below* the relational and logical
// layers, so a comparison could not appear anywhere a value was expected:
// "Declare b to be x is greater than 5." was a syntax error even though
// boolean is a first-class declared type. Only conditions and parenthesised
// expressions reached the comparison layer.
func (p *Parser) parseExpression() (ast.Expression, error) {
	return p.parseOr()
}

// parseArgument parses an expression in a position where "and" is a separator
// rather than an operator — argument lists ("with 5 and 7"), "X of Y" call
// arguments, and the "ask" prompt.
//
// It stops below the "and"/"or" layer so those keywords stay available as
// separators; write parentheses to use them as operators in such a position,
// as in "Call f with (a and b).".
func (p *Parser) parseArgument() (ast.Expression, error) {
	return p.parseNot()
}

// parseOfArgument parses the operand of an "X of Y" phrase — "casefold of name",
// "the length of items", "the value of x".
//
// It binds tightest of the three entry points, because such a phrase is itself
// the left operand of any following comparison: "casefold of answer is equal to
// \"y\"" must group as "(casefold of answer) is equal to \"y\"", not as
// "casefold of (answer is equal to \"y\")".
func (p *Parser) parseOfArgument() (ast.Expression, error) {
	return p.parseCast()
}

// isPossessiveMethodNameToken reports whether token t can legally appear as a
// method name after a possessive ('s). We allow plain identifiers and keyword
// tokens only, and reject literals/operators/punctuation by default.
func isPossessiveMethodNameToken(t token.Type) bool {
	return t == token.IDENTIFIER || token.IsKeyword(t)
}

// parseCast handles postfix operators on any expression:
//   - "cast to <type>" / "casted to <type>" — explicit type conversion
//   - "has <key>"                            — lookup table key check
func (p *Parser) parseCast() (ast.Expression, error) {
	start := at(p.curToken)
	expr, err := p.parseCastExpr(start)
	if err != nil {
		return nil, err
	}
	ast.SetPosIfUnknown(expr, start.Position)
	return expr, nil
}

func (p *Parser) parseCastExpr(start ast.Base) (ast.Expression, error) {
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
		typeName, err := p.parseTypeName()
		if err != nil {
			return nil, err
		}
		return &ast.CastExpression{Value: expr, Type: typeName}, nil
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
	//   x's length       →   MethodCall{Object: x,       MethodName: "length"}
	//   "hello"'s title  →   MethodCall{Object: "hello", MethodName: "title"}
	if p.curToken.Type == token.POSSESSIVE {
		p.nextToken() // consume 's
		return p.parseMethodAfterPossessive(expr)
	}

	return expr, nil
}

// parseMethodAfterPossessive reads the method name and arguments that follow a
// consumed "'s", for any object expression.
//
// This is the one place the construct is parsed. There used to be three, one
// per way the possessive could reach the parser, and they disagreed about what
// may follow the apostrophe and whether arguments are allowed.
func (p *Parser) parseMethodAfterPossessive(object ast.Expression) (*ast.MethodCall, error) {
	if !isPossessiveMethodNameToken(p.curToken.Type) {
		return nil, p.syntaxErr(msgPossessive, hintPossessive)
	}
	methodName := p.curToken.Value
	p.nextToken()

	args, err := p.parseWithArguments()
	if err != nil {
		return nil, err
	}
	return &ast.MethodCall{Object: object, MethodName: methodName, Arguments: args}, nil
}

// canNameAType reports whether a token may begin a type annotation.
//
// Several type names are lexed as keywords rather than identifiers — "integer",
// "array", "table", "range", "type" — which is why the typed-declaration form
// used to reject "Declare x as integer to be 5." even though types.Parse
// accepts "integer" and the error hint advertised it.
func canNameAType(t token.Type) bool {
	switch t {
	case token.IDENTIFIER, token.INTEGER, token.UNSIGNED,
		token.ARRAY, token.LOOKUP, token.TABLE, token.RANGE, token.TYPE:
		return true
	}
	return false
}

// parseTypeName parses a type annotation and returns its name.
//
// This is the single entry point for every annotation position — typed
// declarations, struct fields, "cast to", and array element types. Those four
// used to be served by three separate implementations that disagreed: one
// accepted *any* token at all (so "cast x to 5" parsed cleanly and only failed
// at run time), one accepted only bare identifiers (so keyword-named types were
// syntax errors), and they differed on case handling and on whether a leading
// article was allowed.
//
// A name that matches a built-in type is normalised to lower case; anything
// else keeps its original spelling, because it may name a struct — which only
// the type checker can resolve.
func (p *Parser) parseTypeName() (*ast.TypeExpr, error) {
	pos := at(p.curToken)

	// An article reads naturally here: "Declare x as a number to be 5."
	if p.curToken.Type == token.IDENTIFIER &&
		(strings.EqualFold(p.curToken.Value, "a") || strings.EqualFold(p.curToken.Value, "an")) {
		p.nextToken()
		pos = at(p.curToken)
	}

	if !canNameAType(p.curToken.Type) {
		return nil, p.syntaxErr(
			fmt.Sprintf(msgFmtTypeNameExpected, tokenFriendlyValue(p.curToken.Type, p.curToken.Value)),
			fmt.Sprintf(hintFmtTypeName, strings.Join(types.UserTypeNames(), ", ")),
		)
	}

	// Multi-word built-in names.
	switch p.curToken.Type {
	case token.UNSIGNED:
		p.nextToken()
		if p.curToken.Type != token.INTEGER {
			return nil, p.syntaxErr(
				msgUnsignedNeedsInteger,
				hintUnsignedInteger,
			)
		}
		p.nextToken()
		return typeExprAt(pos, "unsigned integer"), nil
	case token.LOOKUP:
		p.nextToken()
		if p.curToken.Type == token.TABLE {
			p.nextToken()
		}
		return typeExprAt(pos, "lookup table"), nil
	case token.INTEGER:
		p.nextToken()
		return typeExprAt(pos, "integer"), nil
	}

	name := p.curToken.Value
	p.nextToken()
	// Normalise built-in spellings; leave anything else alone so that a struct
	// name keeps the case it was declared with.
	if types.Parse(name) != types.TypeUnknown {
		name = strings.ToLower(name)
	}
	return typeExprAt(pos, name), nil
}

// typeExprAt builds a positioned type annotation, resolving its built-in kind
// once so that nothing downstream has to re-parse the name at run time.
func typeExprAt(pos ast.Base, name string) *ast.TypeExpr {
	return &ast.TypeExpr{Base: pos, Name: name, Kind: types.Parse(name)}
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
		opPos := at(p.curToken)
		p.nextToken()
		right, err := p.parseMultiplicative()
		if err != nil {
			return nil, err
		}
		left = &ast.BinaryExpression{
			Base:     opPos,
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
		opPos := at(p.curToken)
		p.nextToken()
		right, err := p.parsePrimary()
		if err != nil {
			return nil, err
		}
		left = &ast.BinaryExpression{
			Base:     opPos,
			Left:     left,
			Operator: op,
			Right:    right,
		}
	}

	return left, nil
}

// parsePrimary parses a primary expression and records where it started.
//
// The stamping happens here, once, rather than at each of the ~40 node
// constructions inside parsePrimaryExpr.
func (p *Parser) parsePrimary() (ast.Expression, error) {
	start := at(p.curToken)
	expr, err := p.parsePrimaryExpr()
	if err != nil {
		return nil, err
	}
	ast.SetPosIfUnknown(expr, start.Position)
	return expr, nil
}

func (p *Parser) parsePrimaryExpr() (ast.Expression, error) {
	switch p.curToken.Type {
	case token.NUMBER:
		// A literal too large for a 64-bit float used to become +Inf here,
		// silently, because the error was discarded: a program full of digits
		// ran and computed with infinity.
		value, err := strconv.ParseFloat(p.curToken.Value, 64)
		if err != nil {
			return nil, p.syntaxErr(
				fmt.Sprintf(msgFmtNumberOutOfRange, p.curToken.Value),
				hintNumberOutOfRange,
			)
		}
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
		prompt, err := p.parseArgument()
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
		// "the result of calling f with …" — a call used as a value.
		//
		// This was previously recognised only after "Set", and by consuming
		// "the" and "result" before checking what followed, so the phrase was
		// a syntax error anywhere else and "Set x to be the result of f."
		// silently dropped the words it had already eaten. The three-token
		// lookahead here needs no rollback.
		if p.isWord(resultKeyword) &&
			p.peekToken.Type == token.OF &&
			p.tokenAt(p.position).Type == token.CALLING {
			return p.parseCallResult()
		}
		// Check for field access: "the name of person"
		if p.curToken.Type == token.IDENTIFIER {
			fieldName := p.curToken.Value
			lowerFieldName := strings.ToLower(fieldName)
			p.nextToken()
			if p.curToken.Type == token.OF {
				p.nextToken()
				obj, err := p.parseOfArgument()
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
			return p.parseOfArgument()
		}
		return nil, p.syntaxErr(
			fmt.Sprintf(msgFmtTheUnknown, p.curToken.Value),
			hintTheExpression,
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
			if p.curToken.Type == token.RANGE {
				// "a range from X to Y"
				return p.parseRangeExpression()
			}
			// Not a special phrase, treat "a"/"an" as identifier
			return &ast.Identifier{Name: name}, nil
		}

		p.nextToken()

		// A possessive after a name — "x's length" — is handled by the postfix
		// layer, which sees the POSSESSIVE token after this returns the plain
		// identifier. This used to test the identifier's own text for a "'s"
		// suffix, because the lexer folded the possessive into the name.

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
			arg, err := p.parseOfArgument()
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

	case token.ERROR:
		return nil, p.errorTokenErr()

	default:
		return nil, p.syntaxErr(
			fmt.Sprintf(msgFmtExprUnknown, p.curToken.Value),
			hintExpressionValue,
		)
	}
}

// parseIndexExpression parses "item at position X in/of Y"
// parseCallResult parses "result of calling f with …" with "the" already
// consumed, producing the call as an ordinary expression.
func (p *Parser) parseCallResult() (ast.Expression, error) {
	pos := at(p.curToken)
	p.nextToken() // consume "result"
	p.nextToken() // consume "of"
	p.nextToken() // consume "calling"

	if p.curToken.Type != token.IDENTIFIER {
		return nil, p.syntaxErr(msgSetCallFuncName, hintSetCallResult)
	}
	funcName := p.curToken.Value
	p.nextToken()

	args, err := p.parseFunctionArguments()
	if err != nil {
		return nil, err
	}
	return &ast.FunctionCall{Base: pos, Name: funcName, Arguments: args}, nil
}

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
			fmt.Sprintf(msgFmtIndexAfter, p.curToken.Value),
			hintIndexInOrOf,
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
			fmt.Sprintf(msgFmtRemainderAfter, p.curToken.Value),
			hintRemainderDividedBy,
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
			msgLocationVar,
			hintLocationOf,
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
			msgReferenceVar,
			hintReferenceTo,
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
	startPos := at(p.curToken)
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
			msgToggleVar,
			hintToggle,
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
		Base: startPos,
	}, nil
}

func (p *Parser) parseList() (ast.Expression, error) {
	if err := p.expectToken(token.LBRACKET); err != nil {
		return nil, err
	}
	p.nextToken()

	// Empty list
	if p.curToken.Type == token.RBRACKET {
		p.nextToken()
		return &ast.ListLiteral{Elements: []ast.Expression{}}, nil
	}

	// Parse first expression
	firstExpr, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	// Check if it's a range [start .. end] or [start .. end by step]
	if p.curToken.Type == token.DOTDOT {
		p.nextToken() // consume ".."
		endExpr, err := p.parseExpression()
		if err != nil {
			return nil, err
		}

		// Check for optional "by step" clause
		var stepExpr ast.Expression
		if p.curToken.Type == token.BY {
			p.nextToken() // consume "by"
			stepExpr, err = p.parseExpression()
			if err != nil {
				return nil, err
			}
		}

		if err := p.expectToken(token.RBRACKET); err != nil {
			return nil, err
		}
		p.nextToken()
		return &ast.RangeLiteral{Start: firstExpr, End: endExpr, Step: stepExpr}, nil
	}

	// Otherwise it's a regular list
	var elements []ast.Expression
	elements = append(elements, firstExpr)

	for p.curToken.Type == token.COMMA {
		p.nextToken()
		elem, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		elements = append(elements, elem)
	}

	if err := p.expectToken(token.RBRACKET); err != nil {
		return nil, err
	}
	p.nextToken()

	return &ast.ListLiteral{Elements: elements}, nil
}

func (p *Parser) parseRangeExpression() (ast.Expression, error) {
	// Expects: RANGE FROM <expr> TO <expr> [BY <expr>]
	if err := p.expectToken(token.RANGE); err != nil {
		return nil, err
	}
	p.nextToken() // consume RANGE

	if err := p.expectToken(token.FROM); err != nil {
		return nil, err
	}
	p.nextToken() // consume FROM

	startExpr, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	if err := p.expectToken(token.TO); err != nil {
		return nil, err
	}
	p.nextToken() // consume TO

	endExpr, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	// Check for optional "by step" or "stepping by step" clause
	var stepExpr ast.Expression
	if p.curToken.Type == token.BY {
		p.nextToken() // consume "by"
		stepExpr, err = p.parseExpression()
		if err != nil {
			return nil, err
		}
	}

	return &ast.RangeLiteral{Start: startExpr, End: endExpr, Step: stepExpr}, nil
}

func (p *Parser) parseFunctionArguments() ([]ast.Expression, error) {
	return p.parseWithArguments()
}

// parseFunctionCallArgs reads the arguments of the parenthesised form, f(a, b).
//
// Separate from parseArgumentList on purpose: inside parentheses "and" is the
// boolean operator, so only a comma separates arguments and each one is a full
// expression rather than one that stops at "and".
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
			msgArrayNeedsOf,
			hintArrayLiteral,
		)
	}
	p.nextToken() // consume OF

	// Optional element type hint before the bracket
	var elementType *ast.TypeExpr
	if p.curToken.Type != token.LBRACKET {
		var err error
		elementType, err = p.parseTypeName()
		if err != nil {
			return nil, err
		}
	}

	if p.curToken.Type != token.LBRACKET {
		hint := hintArrayLiteral
		if elementType != nil {
			hint = fmt.Sprintf(hintFmtArrayAfterType, elementType.Name)
		}
		return nil, p.syntaxErr(
			fmt.Sprintf(msgFmtArrayOpenBracket, p.curToken.Value),
			hint,
		)
	}
	p.nextToken() // consume [

	// Elements are separated by commas, as they are in a list. The comma used
	// to be optional here, so "an array of number [1 2 3]" was a three-element
	// array — the same text that is a syntax error one line up in a list.
	var elements []ast.Expression
	for p.curToken.Type != token.RBRACKET && p.curToken.Type != token.EOF {
		if len(elements) > 0 {
			if p.curToken.Type != token.COMMA {
				return nil, p.syntaxErr(
					fmt.Sprintf(msgFmtArraySeparator, p.curToken.Value),
					hintArraySeparator,
				)
			}
			p.nextToken()
		}
		elem, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		elements = append(elements, elem)
	}
	if p.curToken.Type != token.RBRACKET {
		return nil, p.syntaxErr(
			msgArrayNeedsCloseBrkt,
			hintArrayCloseBracket,
		)
	}
	p.nextToken() // consume ]

	return &ast.ArrayLiteral{ElemType: elementType, Elements: elements}, nil
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
			fmt.Sprintf(msgFmtLookupEntryIn, p.curToken.Value),
			hintLookupEntryIn,
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
func (p *Parser) parseLookupKeyAssignment(setPos ast.Base) (ast.Statement, error) {
	p.nextToken() // consume ENTRY

	key, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	if p.curToken.Type != token.IN {
		return nil, p.syntaxErr(
			fmt.Sprintf(msgFmtLookupEntryIn, p.curToken.Value),
			hintLookupSetEntry,
		)
	}
	p.nextToken() // consume IN

	if p.curToken.Type != token.IDENTIFIER {
		return nil, p.syntaxErr(
			msgLookupTableName,
			hintLookupSetEntry,
		)
	}
	tableName := p.curToken.Value
	p.nextToken()

	if p.curToken.Type != token.TO {
		return nil, p.syntaxErr(
			fmt.Sprintf(msgFmtLookupTableTo, tableName, p.curToken.Value),
			hintLookupSetEntry,
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

	return &ast.LookupKeyAssignment{TableName: tableName, Key: key, Value: value, Base: setPos}, nil
}

// parseSleepStatement parses "Sleep for <duration>." and "Wait for <duration>."
// where duration is either:
//   - <number><unit>  e.g. 500ms, 2s, 1m, 1h
//   - a second / a minute / an hour  (natural-language shorthands)
//
// Accepted unit names:
//
//	ms / millisecond / milliseconds
//	s  / second      / seconds
//	m  / minute      / minutes
//	h  / hour        / hours
//
// The statement desugars into a CallStatement that calls the stdlib "sleep"
// function with the duration already converted to seconds.
//
// Examples:
//
//	Sleep for 500ms.
//	Wait for 2 seconds.
//	Please sleep for 1 minute.
//	Would you kindly wait for a second.
func (p *Parser) parseSleepStatement() (ast.Statement, error) {
	sleepPos := at(p.curToken)
	p.nextToken() // consume SLEEP / WAIT

	if p.curToken.Type != token.FOR {
		return nil, &SyntaxError{
			Msg:  "Expected 'for' after 'sleep' or 'wait'.",
			Line: p.curToken.Line,
			Col:  p.curToken.Col,
			Hint: "Use the form: 'Sleep for 500ms.' or 'Wait for 2 seconds.'",
		}
	}
	p.nextToken() // consume FOR

	var seconds float64

	// Handle the natural-English shorthands: "a second", "an hour", etc.
	isArticle := p.curToken.Type == token.IDENTIFIER &&
		(strings.ToLower(p.curToken.Value) == "a" || strings.ToLower(p.curToken.Value) == "an")
	if isArticle {
		p.nextToken() // consume article

		if p.curToken.Type != token.IDENTIFIER {
			return nil, &SyntaxError{
				Msg:  "Expected a time unit after article (a/an).",
				Line: p.curToken.Line,
				Col:  p.curToken.Col,
				Hint: "Use the form: 'Sleep for a second.' or 'Wait for an hour.'",
			}
		}
		unit := strings.ToLower(p.curToken.Value)
		switch unit {
		case "millisecond", "milliseconds", "ms":
			seconds = 0.001
		case "second", "seconds", "s":
			seconds = 1.0
		case "minute", "minutes", "m":
			seconds = 60.0
		case "hour", "hours", "h":
			seconds = 3600.0
		default:
			return nil, &SyntaxError{
				Msg:  fmt.Sprintf("Unknown time unit %q after article. Use second, minute, or hour.", unit),
				Line: p.curToken.Line,
				Col:  p.curToken.Col,
				Hint: "Use the form: 'Sleep for a second.' or 'Wait for an hour.'",
			}
		}
		p.nextToken() // consume unit
	} else {
		if p.curToken.Type != token.NUMBER {
			return nil, &SyntaxError{
				Msg:  fmt.Sprintf("Expected a number or 'a'/'an' after 'for', got %s.", tokenFriendlyValue(p.curToken.Type, p.curToken.Value)),
				Line: p.curToken.Line,
				Col:  p.curToken.Col,
				Hint: "Use the form: 'Sleep for 500ms.' or 'Wait for 2 seconds.'",
			}
		}
		numVal, err := strconv.ParseFloat(p.curToken.Value, 64)
		if err != nil {
			return nil, &SyntaxError{
				Msg:  fmt.Sprintf("Invalid duration number: %s", p.curToken.Value),
				Line: p.curToken.Line,
				Col:  p.curToken.Col,
			}
		}
		p.nextToken() // consume NUMBER

		// Read the unit identifier.
		if p.curToken.Type != token.IDENTIFIER {
			return nil, &SyntaxError{
				Msg:  fmt.Sprintf("Expected a time unit after the number, got %s.", tokenFriendlyValue(p.curToken.Type, p.curToken.Value)),
				Line: p.curToken.Line,
				Col:  p.curToken.Col,
				Hint: "Use the form: 'Sleep for 500ms.' or 'Wait for 2 seconds.'",
			}
		}
		unit := strings.ToLower(p.curToken.Value)
		var multiplier float64
		switch unit {
		case "ms", "millisecond", "milliseconds":
			multiplier = 0.001
		case "s", "second", "seconds":
			multiplier = 1.0
		case "m", "minute", "minutes":
			multiplier = 60.0
		case "h", "hour", "hours":
			multiplier = 3600.0
		default:
			return nil, &SyntaxError{
				Msg:  fmt.Sprintf("Unknown time unit %q. Use ms, s, m, or h (or milliseconds, seconds, minutes, hours).", unit),
				Line: p.curToken.Line,
				Col:  p.curToken.Col,
				Hint: "Use the form: 'Sleep for 500ms.' or 'Wait for 2 seconds.'",
			}
		}
		p.nextToken() // consume unit
		seconds = numVal * multiplier
	}

	if err := p.expectToken(token.PERIOD); err != nil {
		return nil, err
	}
	p.nextToken() // consume PERIOD

	return &ast.CallStatement{
		FunctionCall: &ast.FunctionCall{
			Base:      sleepPos,
			Name:      "sleep",
			Arguments: []ast.Expression{&ast.NumberLiteral{Base: sleepPos, Value: seconds}},
		},
		Base: sleepPos,
	}, nil
}
