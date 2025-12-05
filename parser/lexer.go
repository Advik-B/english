// Package parser provides the lexer and parser for the English programming language.
package parser

import (
	"english/token"
	"strings"
	"unicode"
)

// Lexer tokenizes English language source code
type Lexer struct {
	input        string
	position     int
	line         int
	col          int
	readPosition int
	ch           byte
}

// NewLexer creates a new lexer for the given input
func NewLexer(input string) *Lexer {
	l := &Lexer{input: input, line: 1, col: 0}
	l.readChar()
	return l
}

func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition++
	l.col++
	if l.ch == '\n' {
		l.line++
		l.col = 0
	}
}

func (l *Lexer) peekChar() byte {
	if l.readPosition >= len(l.input) {
		return 0
	}
	return l.input[l.readPosition]
}

func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\r' {
		l.readChar()
	}
}

func (l *Lexer) skipComment() {
	if l.ch == '#' {
		for l.ch != '\n' && l.ch != 0 {
			l.readChar()
		}
	}
}

func (l *Lexer) readString(quote byte) string {
	l.readChar() // skip opening quote
	start := l.position
	for l.ch != quote && l.ch != 0 {
		l.readChar()
	}
	result := l.input[start:l.position]
	l.readChar() // skip closing quote
	return result
}

func (l *Lexer) readNumber() string {
	start := l.position
	for unicode.IsDigit(rune(l.ch)) {
		l.readChar()
	}
	// Handle decimal numbers
	if l.ch == '.' && unicode.IsDigit(rune(l.peekChar())) {
		l.readChar() // skip dot
		for unicode.IsDigit(rune(l.ch)) {
			l.readChar()
		}
	}
	return l.input[start:l.position]
}

func (l *Lexer) readIdentifier() string {
	start := l.position
	for unicode.IsLetter(rune(l.ch)) || unicode.IsDigit(rune(l.ch)) || l.ch == '_' {
		l.readChar()
	}
	return l.input[start:l.position]
}

// keywords maps lowercase keywords to their token types
var keywords = map[string]token.Type{
	"declare":   token.DECLARE,
	"function":  token.FUNCTION,
	"that":      token.THAT,
	"does":      token.DOES,
	"the":       token.THE,
	"following": token.FOLLOWING,
	"thats":     token.THATS,
	"it":        token.IT,
	"to":        token.TO,
	"be":        token.BE,
	"always":    token.ALWAYS,
	"set":       token.SET,
	"call":      token.CALL,
	"return":    token.RETURN,
	"print":     token.PRINT,
	"if":        token.IF,
	"then":      token.THEN,
	"otherwise": token.OTHERWISE,
	"repeat":    token.REPEAT,
	"while":     token.WHILE,
	"times":     token.TIMES,
	"for":       token.FOR,
	"each":      token.EACH,
	"in":        token.IN,
	"do":        token.DO,
	"takes":     token.TAKES,
	"and":       token.AND,
	"with":      token.WITH,
	"of":        token.OF,
	"calling":   token.CALLING,
	"value":     token.VALUE,
	"item":      token.ITEM,
	"at":        token.AT,
	"position":  token.POSITION,
	"length":    token.LENGTH,
	"remainder": token.REMAINDER,
	"divided":   token.DIVIDED,
	"by":        token.BY,
	"true":      token.TRUE,
	"false":     token.FALSE,
	"toggle":    token.TOGGLE,
	"location":  token.LOCATION,
}

func (l *Lexer) lookupKeyword(word string) token.Type {
	if tokenType, ok := keywords[strings.ToLower(word)]; ok {
		return tokenType
	}
	return token.IDENTIFIER
}

// NextToken returns the next token from the input
func (l *Lexer) NextToken() token.Token {
	l.skipWhitespace()
	l.skipComment()
	l.skipWhitespace()

	line := l.line
	col := l.col

	if l.ch == 0 {
		return token.Token{Type: token.EOF, Line: line, Col: col}
	}

	// Check for multi-word comparison operators (case-insensitive)
	if (l.ch == 'i' || l.ch == 'I') && l.position+1 < len(l.input) &&
		strings.ToLower(l.input[l.position:l.position+2]) == "is" {
		return l.tryMultiWordComparison()
	}

	var tok token.Token

	switch l.ch {
	case '.':
		tok = token.Token{Type: token.PERIOD, Value: ".", Line: line, Col: col}
		l.readChar()
	case ',':
		tok = token.Token{Type: token.COMMA, Value: ",", Line: line, Col: col}
		l.readChar()
	case ':':
		tok = token.Token{Type: token.COLON, Value: ":", Line: line, Col: col}
		l.readChar()
	case '(':
		tok = token.Token{Type: token.LPAREN, Value: "(", Line: line, Col: col}
		l.readChar()
	case ')':
		tok = token.Token{Type: token.RPAREN, Value: ")", Line: line, Col: col}
		l.readChar()
	case '[':
		tok = token.Token{Type: token.LBRACKET, Value: "[", Line: line, Col: col}
		l.readChar()
	case ']':
		tok = token.Token{Type: token.RBRACKET, Value: "]", Line: line, Col: col}
		l.readChar()
	case '+':
		tok = token.Token{Type: token.PLUS, Value: "+", Line: line, Col: col}
		l.readChar()
	case '-':
		tok = token.Token{Type: token.MINUS, Value: "-", Line: line, Col: col}
		l.readChar()
	case '*':
		tok = token.Token{Type: token.STAR, Value: "*", Line: line, Col: col}
		l.readChar()
	case '/':
		tok = token.Token{Type: token.SLASH, Value: "/", Line: line, Col: col}
		l.readChar()
	case '"', '\'':
		quote := l.ch
		str := l.readString(quote)
		tok = token.Token{Type: token.STRING, Value: str, Line: line, Col: col}
	case '\n':
		tok = token.Token{Type: token.NEWLINE, Value: "\n", Line: line, Col: col}
		l.readChar()
	default:
		if unicode.IsDigit(rune(l.ch)) {
			num := l.readNumber()
			return token.Token{Type: token.NUMBER, Value: num, Line: line, Col: col}
		} else if unicode.IsLetter(rune(l.ch)) || l.ch == '_' {
			ident := l.readIdentifier()
			tokenType := l.lookupKeyword(ident)
			return token.Token{Type: tokenType, Value: ident, Line: line, Col: col}
		}
		tok = token.Token{Type: token.ERROR, Value: string(l.ch), Line: line, Col: col}
		l.readChar()
	}

	return tok
}

// tryMultiWordComparison handles multi-word operators like "is equal to"
func (l *Lexer) tryMultiWordComparison() token.Token {
	line := l.line
	col := l.col

	// Save current state for potential rollback
	savePos := l.position
	saveReadPos := l.readPosition
	saveLine := l.line
	saveCol := l.col
	saveCh := l.ch

	// Try to read the full operator - read words one at a time and check for valid phrases
	words := []string{}
	bestMatch := ""
	bestMatchType := token.ERROR
	bestMatchPos := l.position
	bestMatchReadPos := l.readPosition
	bestMatchLine := l.line
	bestMatchCol := l.col
	bestMatchCh := l.ch

	for {
		l.skipWhitespace()
		if !unicode.IsLetter(rune(l.ch)) {
			break
		}
		word := l.readIdentifier()
		words = append(words, strings.ToLower(word))

		phrase := strings.Join(words, " ")

		// Check if current phrase matches a comparison operator
		var tokenType token.Type
		switch phrase {
		case "is equal to":
			tokenType = token.IS_EQUAL_TO
		case "is less than":
			tokenType = token.IS_LESS_THAN
		case "is greater than":
			tokenType = token.IS_GREATER_THAN
		case "is less than or equal to":
			tokenType = token.IS_LESS_EQUAL
		case "is greater than or equal to":
			tokenType = token.IS_GREATER_EQUAL
		case "is not equal to":
			tokenType = token.IS_NOT_EQUAL
		default:
			tokenType = token.ERROR
		}

		// If we found a match, save it (we want the longest match)
		if tokenType != token.ERROR {
			bestMatch = phrase
			bestMatchType = tokenType
			bestMatchPos = l.position
			bestMatchReadPos = l.readPosition
			bestMatchLine = l.line
			bestMatchCol = l.col
			bestMatchCh = l.ch
		}

		l.skipWhitespace()
		if l.ch == ',' || l.ch == ':' || l.ch == 0 {
			break
		}
	}

	// If we found a valid comparison operator, use it
	if bestMatchType != token.ERROR {
		// Restore position to after the matched phrase
		l.position = bestMatchPos
		l.readPosition = bestMatchReadPos
		l.line = bestMatchLine
		l.col = bestMatchCol
		l.ch = bestMatchCh
		return token.Token{Type: bestMatchType, Value: bestMatch, Line: line, Col: col}
	}

	// Restore position if not a comparison operator
	l.position = savePos
	l.readPosition = saveReadPos
	l.line = saveLine
	l.col = saveCol
	l.ch = saveCh
	word := l.readIdentifier()
	return token.Token{Type: l.lookupKeyword(word), Value: word, Line: line, Col: col}
}

// TokenizeAll returns all tokens from the input
func (l *Lexer) TokenizeAll() []token.Token {
	var tokens []token.Token
	for {
		tok := l.NextToken()
		if tok.Type != token.NEWLINE { // Skip newlines for now
			tokens = append(tokens, tok)
		}
		if tok.Type == token.EOF {
			break
		}
	}
	return tokens
}
