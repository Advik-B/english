package parser

import (
	"strings"
	"testing"

	"github.com/Advik-B/english/token"
)

// looksLikeGoConstant reports whether s is an upper-snake-case identifier — the
// shape of a Go constant name, which must never reach a user-facing message.
func looksLikeGoConstant(s string) bool {
	s = strings.Trim(s, "'")
	if s == "" {
		return false
	}
	for _, r := range s {
		if !(r >= 'A' && r <= 'Z') && r != '_' && !(r >= '0' && r <= '9') {
			return false
		}
	}
	return true
}

// TestTokenFriendlyNameNeverLeaksConstants guards the whole class of bug where
// a token type the parser can demand had no friendly rendering and fell through
// to its Go constant name — a missing "then" produced "I expected 'THEN' here".
//
// It sweeps every token type the enum defines, so a token added later is
// covered automatically rather than needing a new case here.
func TestTokenFriendlyNameNeverLeaksConstants(t *testing.T) {
	for tt := token.Type(0); tt <= token.COMMENT; tt++ {
		name := tokenFriendlyName(tt)
		if looksLikeGoConstant(name) {
			t.Errorf("token %s renders as %q, which leaks the Go constant name", tt, name)
		}
		if name == "" {
			t.Errorf("token %s renders as the empty string", tt)
		}
	}
}

// TestKeywordsRenderAsWords checks the derived path produces readable wording
// for keywords rather than punctuation phrasing.
func TestKeywordsRenderAsWords(t *testing.T) {
	cases := map[token.Type]string{
		token.THEN:      "the word 'then'",
		token.AS:        "the word 'as'",
		token.LET:       "the word 'let'",
		token.EACH:      "the word 'each'",
		token.FINALLY:   "the word 'finally'",
		token.REMAINDER: "the word 'remainder'",
		token.NOTHING:   "the word 'nothing'",
		token.PLEASE:    "the word 'please'",
		token.CASTED:    "the word 'cast'",
		token.SLEEP:     "the word 'sleep'",
	}
	for tt, want := range cases {
		if got := tokenFriendlyName(tt); got != want {
			t.Errorf("tokenFriendlyName(%s) = %q, want %q", tt, got, want)
		}
	}
}

// TestEveryDemandedTokenHasAName asserts that every token type expectToken can
// be called with renders readably. The list mirrors the expectToken call sites.
func TestEveryDemandedTokenHasAName(t *testing.T) {
	demanded := []token.Type{
		token.AND, token.AS, token.AT, token.BE, token.BREAK, token.BY,
		token.CALL, token.COLON, token.COMMA, token.COPY, token.DECLARE,
		token.DEFAULT, token.DO, token.DOES, token.DOING, token.EACH,
		token.FINALLY, token.FOLLOWING, token.FOR, token.FROM, token.FUNCTION,
		token.IF, token.IMPORT, token.IN, token.INSTANCE, token.IS, token.IT,
		token.ITEM, token.LBRACKET, token.LENGTH, token.LET, token.LOCATION,
		token.LOOP, token.OF, token.OUT, token.PERIOD, token.POSITION,
		token.RANGE, token.RBRACKET, token.REFERENCE, token.REMAINDER,
		token.REPEAT, token.RETURN, token.RPAREN, token.SET, token.THATS,
		token.THE, token.THEN, token.TIMES, token.TO, token.TOGGLE,
		token.TYPE, token.WITH,
	}
	for _, tt := range demanded {
		name := tokenFriendlyName(tt)
		if looksLikeGoConstant(name) {
			t.Errorf("token %s is demanded by expectToken but renders as %q", tt, name)
		}
	}
}
