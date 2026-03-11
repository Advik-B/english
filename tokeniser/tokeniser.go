// Package tokeniser provides the shared lexer for the English programming
// language.  Both the compiler pipeline (via the parser package) and the
// syntax-highlighter (via the highlight package) use this lexer so that
// keyword recognition, operator phrases, possessive handling, and every
// other tokenisation rule have a single authoritative implementation.
package tokeniser

import (
	"english/token"
	"strings"
	"unicode"
)

// Lexer tokenizes English language source code.
type Lexer struct {
	input         string
	position      int
	line          int
	col           int
	readPosition  int
	ch            byte
	lastTokenType token.Type // type of the most recently emitted token
}

// NewLexer creates a new Lexer for the given input.
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

// peekCharN returns the character n positions ahead of the current position
// (n=1 is the same as peekChar).
func (l *Lexer) peekCharN(n int) byte {
	pos := l.readPosition + n - 1
	if pos >= len(l.input) {
		return 0
	}
	return l.input[pos]
}

// isIdentChar reports whether b can appear in an identifier (letter, digit, or underscore).
func isIdentChar(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9') || b == '_'
}

// isPossessiveContext reports whether a possessive 's can follow a token of the given type.
// Only tokens that end a value expression (identifiers, literals, closing delimiters)
// allow a following 's to be treated as the possessive marker.
func isPossessiveContext(t token.Type) bool {
	switch t {
	case token.IDENTIFIER, token.STRING, token.NUMBER,
		token.RPAREN, token.RBRACKET, token.TRUE, token.FALSE:
		return true
	}
	return false
}

func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\r' {
		l.readChar()
	}
}

func (l *Lexer) readComment() (token.Token, bool) {
	if l.ch != '#' {
		return token.Token{}, false
	}
	line := l.line
	col := l.col
	pos := l.position
	l.readChar() // skip '#'
	var sb strings.Builder
	for l.ch != '\n' && l.ch != 0 {
		sb.WriteByte(l.ch)
		l.readChar()
	}
	text := strings.TrimSpace(sb.String())
	return token.Token{Type: token.COMMENT, Value: text, Line: line, Col: col, Pos: pos}, true
}

func (l *Lexer) readString(quote byte) string {
	l.readChar() // skip opening quote
	var result strings.Builder
	for l.ch != quote && l.ch != 0 {
		if l.ch == '\\' {
			l.readChar() // skip backslash
			switch l.ch {
			case 'n':
				result.WriteByte('\n')
			case 't':
				result.WriteByte('\t')
			case 'r':
				result.WriteByte('\r')
			case '\\':
				result.WriteByte('\\')
			case '"':
				result.WriteByte('"')
			case '\'':
				result.WriteByte('\'')
			default:
				// For unrecognized escape sequences (e.g., \x), preserve the backslash
				// and the character as-is. This allows unknown sequences to pass through
				// without error, though they won't have special meaning.
				result.WriteByte('\\')
				result.WriteByte(l.ch)
			}
		} else {
			result.WriteByte(l.ch)
		}
		l.readChar()
	}
	l.readChar() // skip closing quote
	return result.String()
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

	// Check for possessive form: identifier's
	// If we see an apostrophe followed by 's', include it in the identifier
	if l.ch == '\'' && l.peekChar() == 's' {
		l.readChar() // consume apostrophe
		l.readChar() // consume 's'
	}

	return l.input[start:l.position]
}

// keywords maps lowercase keywords to their token types
var keywords = map[string]token.Type{
	"declare":    token.DECLARE,
	"let":        token.LET,
	"equal":      token.EQUAL,
	"function":   token.FUNCTION,
	"that":       token.THAT,
	"does":       token.DOES,
	"the":        token.THE,
	"following":  token.FOLLOWING,
	"thats":      token.THATS,
	"it":         token.IT,
	"to":         token.TO,
	"be":         token.BE,
	"always":     token.ALWAYS,
	"set":        token.SET,
	"call":       token.CALL,
	"return":     token.RETURN,
	"print":      token.PRINT,
	"if":         token.IF,
	"then":       token.THEN,
	"otherwise":  token.OTHERWISE,
	"repeat":     token.REPEAT,
	"while":      token.WHILE,
	"forever":    token.FOREVER,
	"break":      token.BREAK,
	"out":        token.OUT,
	"loop":       token.LOOP,
	"times":      token.TIMES,
	"for":        token.FOR,
	"each":       token.EACH,
	"in":         token.IN,
	"do":         token.DO,
	"takes":      token.TAKES,
	"and":        token.AND,
	"with":       token.WITH,
	"of":         token.OF,
	"calling":    token.CALLING,
	"value":      token.VALUE,
	"item":       token.ITEM,
	"at":         token.AT,
	"position":   token.POSITION,
	"length":     token.LENGTH,
	"remainder":  token.REMAINDER,
	"divided":    token.DIVIDED,
	"by":         token.BY,
	"true":       token.TRUE,
	"false":      token.FALSE,
	"toggle":     token.TOGGLE,
	"location":   token.LOCATION,
	"write":      token.WRITE,
	"as":         token.AS,
	"structure":  token.STRUCTURE,
	"struct":     token.STRUCT,
	"fields":     token.FIELDS,
	"field":      token.FIELD,
	"instance":   token.INSTANCE,
	"new":        token.NEW,
	"try":        token.TRY,
	"doing":      token.DOING,
	"on":         token.ON,
	"finally":    token.FINALLY,
	"raise":      token.RAISE,
	"reference":  token.REFERENCE,
	"copy":       token.COPY,
	"swap":       token.SWAP,
	"casted":     token.CASTED,
	"cast":       token.CASTED,
	"type":       token.TYPE,
	"which":      token.WHICH,
	"is":         token.IS,
	"from":       token.FROM,
	"unsigned":   token.UNSIGNED,
	"integer":    token.INTEGER,
	"default":    token.DEFAULT,
	"but":        token.BUT,
	"import":     token.IMPORT,
	"everything": token.EVERYTHING,
	"all":        token.ALL,
	"safely":     token.SAFELY,
	"continue":   token.CONTINUE,
	"skip":       token.SKIP,
	"nothing":    token.NOTHING,
	"none":       token.NOTHING,
	"null":       token.NOTHING,
	"not":        token.NOT,
	"or":         token.OR,
	"ask":        token.ASK,
	"array":      token.ARRAY,
	"lookup":     token.LOOKUP,
	"table":      token.TABLE,
	"has":        token.HAS,
	"entry":      token.ENTRY,
	// Politeness prefixes – consumed by the parser before any statement.
	"please": token.PLEASE,
	"kindly": token.PLEASE,
	// Sleep statement keyword.
	"sleep": token.SLEEP,
}

func (l *Lexer) lookupKeyword(word string) token.Type {
	if tokenType, ok := keywords[strings.ToLower(word)]; ok {
		return tokenType
	}
	return token.IDENTIFIER
}

// NextToken returns the next token from the input.
func (l *Lexer) NextToken() token.Token {
	l.skipWhitespace()

	// Emit COMMENT token so the parser (and transpiler) can carry comments through.
	if tok, ok := l.readComment(); ok {
		l.lastTokenType = tok.Type
		return tok
	}
	l.skipWhitespace()

	line := l.line
	col := l.col
	pos := l.position

	if l.ch == 0 {
		return token.Token{Type: token.EOF, Line: line, Col: col, Pos: pos}
	}

	// Check for multi-word comparison operators (case-insensitive)
	// Triggers for "is ..." (e.g. "is equal to") and "has ..." (e.g. "has a value")
	if (l.ch == 'i' || l.ch == 'I') && l.position+1 < len(l.input) &&
		strings.ToLower(l.input[l.position:l.position+2]) == "is" {
		return l.tryMultiWordComparison(line, col, pos)
	}
	if (l.ch == 'h' || l.ch == 'H') && l.position+2 < len(l.input) &&
		strings.ToLower(l.input[l.position:l.position+3]) == "has" {
		return l.tryMultiWordComparison(line, col, pos)
	}

	var tok token.Token

	switch l.ch {
	case '.':
		tok = token.Token{Type: token.PERIOD, Value: ".", Line: line, Col: col, Pos: pos}
		l.readChar()
	case ',':
		tok = token.Token{Type: token.COMMA, Value: ",", Line: line, Col: col, Pos: pos}
		l.readChar()
	case ':':
		tok = token.Token{Type: token.COLON, Value: ":", Line: line, Col: col, Pos: pos}
		l.readChar()
	case '(':
		tok = token.Token{Type: token.LPAREN, Value: "(", Line: line, Col: col, Pos: pos}
		l.readChar()
	case ')':
		tok = token.Token{Type: token.RPAREN, Value: ")", Line: line, Col: col, Pos: pos}
		l.readChar()
	case '[':
		tok = token.Token{Type: token.LBRACKET, Value: "[", Line: line, Col: col, Pos: pos}
		l.readChar()
	case ']':
		tok = token.Token{Type: token.RBRACKET, Value: "]", Line: line, Col: col, Pos: pos}
		l.readChar()
	case '+':
		tok = token.Token{Type: token.PLUS, Value: "+", Line: line, Col: col, Pos: pos}
		l.readChar()
	case '-':
		tok = token.Token{Type: token.MINUS, Value: "-", Line: line, Col: col, Pos: pos}
		l.readChar()
	case '*':
		tok = token.Token{Type: token.STAR, Value: "*", Line: line, Col: col, Pos: pos}
		l.readChar()
	case '/':
		tok = token.Token{Type: token.SLASH, Value: "/", Line: line, Col: col, Pos: pos}
		l.readChar()
	case '=':
		tok = token.Token{Type: token.ASSIGN, Value: "=", Line: line, Col: col, Pos: pos}
		l.readChar()
	case '\'':
		// Emit a POSSESSIVE token when the apostrophe-s follows an expression
		// (identifier, string literal, number, closing paren/bracket), and is
		// immediately followed by a non-identifier character.  This enables:
		//   "hello"'s title   →   MethodCall{Object:"hello", MethodName:"title"}
		//   42's is_integer   →   MethodCall{Object:42,       MethodName:"is_integer"}
		// Single-quoted strings ('hello') are still lexed normally when they appear
		// at the start of an expression (i.e. the previous token was not a value).
		if l.peekChar() == 's' && !isIdentChar(l.peekCharN(2)) &&
			isPossessiveContext(l.lastTokenType) {
			l.readChar() // consume '
			l.readChar() // consume s
			tok = token.Token{Type: token.POSSESSIVE, Value: "'s", Line: line, Col: col, Pos: pos}
		} else {
			str := l.readString(l.ch)
			tok = token.Token{Type: token.STRING, Value: str, Line: line, Col: col, Pos: pos}
		}
	case '"':
		str := l.readString(l.ch)
		tok = token.Token{Type: token.STRING, Value: str, Line: line, Col: col, Pos: pos}
	case '\n':
		tok = token.Token{Type: token.NEWLINE, Value: "\n", Line: line, Col: col, Pos: pos}
		l.readChar()
	default:
		if unicode.IsDigit(rune(l.ch)) {
			num := l.readNumber()
			tok := token.Token{Type: token.NUMBER, Value: num, Line: line, Col: col, Pos: pos}
			l.lastTokenType = tok.Type
			return tok
		} else if unicode.IsLetter(rune(l.ch)) || l.ch == '_' {
			ident := l.readIdentifier()
			lower := strings.ToLower(ident)

			// Detect multi-word politeness prefixes: "could you" and
			// "would you kindly". These must start at the beginning of a
			// statement (i.e. after NEWLINE / EOF / at position 0), but
			// checking that strictly would be complex; instead we simply
			// look ahead for the expected following words and emit PLEASE.
			if lower == "could" {
				savPos, savReadPos, savLine, savCol, savCh := l.position, l.readPosition, l.line, l.col, l.ch
				l.skipWhitespace()
				next := l.readIdentifier()
				if strings.ToLower(next) == "you" {
					tok := token.Token{Type: token.PLEASE, Value: "could you", Line: line, Col: col, Pos: pos}
					l.lastTokenType = tok.Type
					return tok
				}
				// Rollback
				l.position, l.readPosition, l.line, l.col, l.ch = savPos, savReadPos, savLine, savCol, savCh
			} else if lower == "would" {
				savPos, savReadPos, savLine, savCol, savCh := l.position, l.readPosition, l.line, l.col, l.ch
				l.skipWhitespace()
				next1 := l.readIdentifier()
				if strings.ToLower(next1) == "you" {
					l.skipWhitespace()
					next2 := l.readIdentifier()
					if strings.ToLower(next2) == "kindly" {
						tok := token.Token{Type: token.PLEASE, Value: "would you kindly", Line: line, Col: col, Pos: pos}
						l.lastTokenType = tok.Type
						return tok
					}
					// "would you" without "kindly" - rollback to after "would"
					l.position, l.readPosition, l.line, l.col, l.ch = savPos, savReadPos, savLine, savCol, savCh
				} else {
					// Rollback
					l.position, l.readPosition, l.line, l.col, l.ch = savPos, savReadPos, savLine, savCol, savCh
				}
			}

			tokenType := l.lookupKeyword(ident)
			tok := token.Token{Type: tokenType, Value: ident, Line: line, Col: col, Pos: pos}
			l.lastTokenType = tok.Type
			return tok
		}
		tok = token.Token{Type: token.ERROR, Value: string(l.ch), Line: line, Col: col, Pos: pos}
		l.readChar()
	}

	l.lastTokenType = tok.Type
	return tok
}

// tryMultiWordComparison handles multi-word operators like "is equal to".
// line, col, and pos are the position of the first character of the phrase,
// already captured by the caller.
func (l *Lexer) tryMultiWordComparison(line, col, pos int) token.Token {
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

		// Handle "isn't" contraction: after reading "isn", consume "'t" to form "isn't"
		if strings.ToLower(word) == "isn" && l.ch == '\'' && (l.peekChar() == 't' || l.peekChar() == 'T') {
			l.readChar() // consume '
			l.readChar() // consume t
			word = "isn't"
		}

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
		case "is something", "has a value":
			tokenType = token.IS_SOMETHING
		case "is nothing", "has no value":
			tokenType = token.IS_NOTHING_OP
		case "is true":
			tokenType = token.IS_TRUE
		case "is false":
			tokenType = token.IS_FALSE
		case "isn't true", "is not true":
			tokenType = token.ISNT_TRUE
		case "isn't false", "is not false":
			tokenType = token.ISNT_FALSE
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
		return token.Token{Type: bestMatchType, Value: bestMatch, Line: line, Col: col, Pos: pos}
	}

	// Restore position if not a comparison operator
	l.position = savePos
	l.readPosition = saveReadPos
	l.line = saveLine
	l.col = saveCol
	l.ch = saveCh
	word := l.readIdentifier()
	return token.Token{Type: l.lookupKeyword(word), Value: word, Line: line, Col: col, Pos: pos}
}

// Offset returns the current byte position in the input.  After a call to
// NextToken, Offset() returns the position of the first byte that has not yet
// been consumed — i.e. the exclusive end of the just-returned token in the
// source string.  This is used by TokenizeForHighlight to locate the raw
// source bytes for each token.
func (l *Lexer) Offset() int {
	return l.position
}

// TokenizeAll returns all tokens from the input, skipping NEWLINE tokens so
// that the parser receives a flat, newline-free token stream.
func (l *Lexer) TokenizeAll() []token.Token {
	var tokens []token.Token
	for {
		tok := l.NextToken()
		// Update lastTokenType for every non-whitespace token so the possessive
		// detector has accurate context on the next call.
		if tok.Type != token.NEWLINE {
			l.lastTokenType = tok.Type
			tokens = append(tokens, tok)
		}
		if tok.Type == token.EOF {
			break
		}
	}
	return tokens
}

// TokenizeForHighlight tokenizes source and returns a token stream suitable
// for syntax highlighting.  Unlike TokenizeAll it:
//
//   - preserves NEWLINE tokens
//   - inserts WHITESPACE tokens for the horizontal whitespace that the lexer
//     normally discards between semantic tokens
//   - sets each token's Value to the exact bytes from source so that the
//     original text can be reconstructed verbatim (including original casing,
//     spacing inside multi-word operators, quote characters around strings,
//     and the leading '#' of comments)
//
// The returned slice, when its Value fields are concatenated in order,
// reproduces source exactly.
func TokenizeForHighlight(source string) []token.Token {
	l := NewLexer(source)
	var tokens []token.Token
	cursor := 0

	for {
		tok := l.NextToken()
		end := l.Offset()

		// Emit a WHITESPACE token for any horizontal whitespace skipped before
		// this token (l.skipWhitespace consumed it without emitting anything).
		if tok.Pos > cursor {
			tokens = append(tokens, token.Token{
				Type:  token.WHITESPACE,
				Value: source[cursor:tok.Pos],
				Pos:   cursor,
			})
		}

		if tok.Type == token.EOF {
			break
		}

		// Override Value with the verbatim source bytes so that highlighting
		// can render strings with their quotes, comments with their '#', and
		// comparison operators with their original spacing and casing.
		rawTok := tok
		rawTok.Value = source[tok.Pos:end]
		tokens = append(tokens, rawTok)
		cursor = end
	}
	return tokens
}
