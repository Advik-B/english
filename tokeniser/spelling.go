package tokeniser

import (
	"sort"
	"sync"

	"github.com/Advik-B/english/token"
)

// canonicalSpelling pins the preferred source spelling for the few token types
// that several words map to. Without it the choice would be alphabetical, which
// picks "kindly" over "please" and "none" over "nothing".
var canonicalSpelling = map[token.Type]string{
	token.PLEASE:  "please",
	token.NOTHING: "nothing",
	token.CASTED:  "cast",
	token.SLEEP:   "sleep",
}

// multiWordSpelling covers the comparison phrases, which the lexer assembles in
// tryMultiWordComparison rather than looking up in the keywords table.
var multiWordSpelling = map[token.Type]string{
	token.IS_EQUAL_TO:      "is equal to",
	token.IS_LESS_THAN:     "is less than",
	token.IS_GREATER_THAN:  "is greater than",
	token.IS_LESS_EQUAL:    "is less than or equal to",
	token.IS_GREATER_EQUAL: "is greater than or equal to",
	token.IS_NOT_EQUAL:     "is not equal to",
	token.IS_SOMETHING:     "is something",
	token.IS_NOTHING_OP:    "is nothing",
	token.IS_TRUE:          "is true",
	token.IS_FALSE:         "is false",
	token.ISNT_TRUE:        "isn't true",
	token.ISNT_FALSE:       "isn't false",
	token.POSSESSIVE:       "'s",
}

var (
	spellingOnce sync.Once
	spellings    map[token.Type]string
)

// buildSpellings reverses the keywords table so that a token type can be
// rendered back as the word a user would actually type.
func buildSpellings() {
	spellings = make(map[token.Type]string, len(keywords)+len(multiWordSpelling))

	// Group every spelling by token type so the choice is deterministic
	// regardless of Go's map iteration order.
	byType := make(map[token.Type][]string, len(keywords))
	for word, t := range keywords {
		byType[t] = append(byType[t], word)
	}
	for t, words := range byType {
		if pinned, ok := canonicalSpelling[t]; ok {
			spellings[t] = pinned
			continue
		}
		sort.Strings(words)
		spellings[t] = words[0]
	}
	for t, phrase := range multiWordSpelling {
		spellings[t] = phrase
	}
}

// Spelling returns the canonical source spelling of a keyword or operator token
// and whether one exists.
//
// It exists so that diagnostics never have to hard-code a list of token names.
// Error messages previously fell back to the Go constant name, so a missing
// "then" produced "I expected 'THEN' here" — and 23 of the token types the
// parser can demand had no friendly name at all.
func Spelling(t token.Type) (string, bool) {
	spellingOnce.Do(buildSpellings)
	s, ok := spellings[t]
	return s, ok
}
