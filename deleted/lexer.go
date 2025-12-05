// DEPRECATED: This file has been moved to interpreter/lexer.go
// This file should be deleted
package deleted

import (
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

func (l *Lexer) peekWord() string {
	start := l.readPosition
	end := start
	for end < len(l.input) && (unicode.IsLetter(rune(l.input[end])) || unicode.IsDigit(rune(l.input[end])) || l.input[end] == '_') {
		end++
	}
	return l.input[start:end]
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

var keywords = map[string]TokenType{
	"declare":   TOKEN_DECLARE,
	"function":  TOKEN_FUNCTION,
	"that":      TOKEN_THAT,
	"does":      TOKEN_DOES,
	"the":       TOKEN_THE,
	"following": TOKEN_FOLLOWING,
	"thats":     TOKEN_THATS,
	"it":        TOKEN_IT,
	"to":        TOKEN_TO,
	"be":        TOKEN_BE,
	"always":    TOKEN_ALWAYS,
	"set":       TOKEN_SET,
	"call":      TOKEN_CALL,
	"return":    TOKEN_RETURN,
	"say":       TOKEN_SAY,
	"if":        TOKEN_IF,
	"then":      TOKEN_THEN,
	"otherwise": TOKEN_OTHERWISE,
	"repeat":    TOKEN_REPEAT,
	"while":     TOKEN_WHILE,
	"times":     TOKEN_TIMES,
	"for":       TOKEN_FOR,
	"each":      TOKEN_EACH,
	"in":        TOKEN_IN,
	"do":        TOKEN_DO,
	"takes":     TOKEN_TAKES,
	"and":       TOKEN_AND,
	"with":      TOKEN_WITH,
	"result":    TOKEN_RESULT,
	"of":        TOKEN_OF,
	"calling":   TOKEN_CALLING,
	"value":     TOKEN_VALUE,
}

func (l *Lexer) lookupKeyword(word string) TokenType {
	if tokenType, ok := keywords[strings.ToLower(word)]; ok {
		return tokenType
	}
	return TOKEN_IDENTIFIER
}

// NextToken returns the next token
func (l *Lexer) NextToken() Token {
	l.skipWhitespace()
	l.skipComment()
	l.skipWhitespace()

	line := l.line
	col := l.col

	if l.ch == 0 {
		return Token{Type: TOKEN_EOF, Line: line, Col: col}
	}

	// Check for multi-word comparison operators
	if l.ch == 'i' && l.input[l.position:l.position+2] == "is" {
		return l.tryMultiWordComparison()
	}

	var tok Token

	switch l.ch {
	case '.':
		tok = Token{Type: TOKEN_PERIOD, Value: ".", Line: line, Col: col}
		l.readChar()
	case ',':
		tok = Token{Type: TOKEN_COMMA, Value: ",", Line: line, Col: col}
		l.readChar()
	case ':':
		tok = Token{Type: TOKEN_COLON, Value: ":", Line: line, Col: col}
		l.readChar()
	case '(':
		tok = Token{Type: TOKEN_LPAREN, Value: "(", Line: line, Col: col}
		l.readChar()
	case ')':
		tok = Token{Type: TOKEN_RPAREN, Value: ")", Line: line, Col: col}
		l.readChar()
	case '[':
		tok = Token{Type: TOKEN_LBRACKET, Value: "[", Line: line, Col: col}
		l.readChar()
	case ']':
		tok = Token{Type: TOKEN_RBRACKET, Value: "]", Line: line, Col: col}
		l.readChar()
	case '+':
		tok = Token{Type: TOKEN_PLUS, Value: "+", Line: line, Col: col}
		l.readChar()
	case '-':
		tok = Token{Type: TOKEN_MINUS, Value: "-", Line: line, Col: col}
		l.readChar()
	case '*':
		tok = Token{Type: TOKEN_STAR, Value: "*", Line: line, Col: col}
		l.readChar()
	case '/':
		tok = Token{Type: TOKEN_SLASH, Value: "/", Line: line, Col: col}
		l.readChar()
	case '"', '\'':
		quote := l.ch
		str := l.readString(quote)
		tok = Token{Type: TOKEN_STRING, Value: str, Line: line, Col: col}
	case '\n':
		tok = Token{Type: TOKEN_NEWLINE, Value: "\n", Line: line, Col: col}
		l.readChar()
	default:
		if unicode.IsDigit(rune(l.ch)) {
			num := l.readNumber()
			return Token{Type: TOKEN_NUMBER, Value: num, Line: line, Col: col}
		} else if unicode.IsLetter(rune(l.ch)) || l.ch == '_' {
			ident := l.readIdentifier()
			tokenType := l.lookupKeyword(ident)
			return Token{Type: tokenType, Value: ident, Line: line, Col: col}
		}
		tok = Token{Type: TOKEN_ERROR, Value: string(l.ch), Line: line, Col: col}
		l.readChar()
	}

	return tok
}

// tryMultiWordComparison handles multi-word operators like "is equal to"
func (l *Lexer) tryMultiWordComparison() Token {
	line := l.line
	col := l.col

	// Read "is" and following words
	savePos := l.position
	saveReadPos := l.readPosition
	saveLine := l.line
	saveCol := l.col
	saveCh := l.ch

	// Try to read the full operator
	phrase := ""
	words := []string{}
	for {
		l.skipWhitespace()
		if !unicode.IsLetter(rune(l.ch)) {
			break
		}
		word := l.readIdentifier()
		words = append(words, strings.ToLower(word))
		l.skipWhitespace()
		if l.ch == ',' || l.ch == ':' || l.ch == 0 {
			break
		}
	}

	phrase = strings.Join(words, " ")

	var tokenType TokenType
	switch phrase {
	case "is equal to":
		tokenType = TOKEN_IS_EQUAL_TO
	case "is less than":
		tokenType = TOKEN_IS_LESS_THAN
	case "is greater than":
		tokenType = TOKEN_IS_GREATER_THAN
	case "is less than or equal to":
		tokenType = TOKEN_IS_LESS_EQUAL
	case "is greater than or equal to":
		tokenType = TOKEN_IS_GREATER_EQUAL
	case "is not equal to":
		tokenType = TOKEN_IS_NOT_EQUAL
	default:
		// Restore position if not a comparison operator
		l.position = savePos
		l.readPosition = saveReadPos
		l.line = saveLine
		l.col = saveCol
		l.ch = saveCh
		word := l.readIdentifier()
		return Token{Type: l.lookupKeyword(word), Value: word, Line: line, Col: col}
	}

	return Token{Type: tokenType, Value: phrase, Line: line, Col: col}
}

// TokenizeAll returns all tokens from the input
func (l *Lexer) TokenizeAll() []Token {
	var tokens []Token
	for {
		tok := l.NextToken()
		if tok.Type != TOKEN_NEWLINE { // Skip newlines for now
			tokens = append(tokens, tok)
		}
		if tok.Type == TOKEN_EOF {
			break
		}
	}
	return tokens
}
