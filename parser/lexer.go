// Package parser provides the lexer and parser for the English programming language.
package parser

import "github.com/Advik-B/english/tokeniser"

// Lexer is a type alias for the shared tokeniser.Lexer so that all existing
// callers (cmd, vm, lsp, transpiler …) continue to work without change while
// both the parser and the highlight package share a single tokenisation
// implementation.
type Lexer = tokeniser.Lexer

// NewLexer creates a new Lexer for the given input.
func NewLexer(input string) *Lexer {
	return tokeniser.NewLexer(input)
}
