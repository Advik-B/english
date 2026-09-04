package repl

import (
	"strings"

	"github.com/Advik-B/english/parser"
)

// needsMoreInput reports whether what has been typed so far is the beginning
// of a block rather than a finished statement or a mistake.
//
// It asks the parser, which is the only thing that knows. The answer used to
// come from searching the text of each line: a line containing "following" and
// ending in ":" opened a block, a line ending in " then" opened a block, and a
// line containing "thats it." closed one. None of those can tell code from the
// inside of a string or a comment, so
//
//	Print "thats it.".
//
// inside a loop ended the block early and sent half a loop to be executed,
// while
//
//	# do the following:
//
// opened a block that no "thats it." would ever close. The heuristic also had
// to carry its own exception list for "otherwise", "on ...:" and "but
// finally:", which are continuations rather than openers — a second grammar,
// maintained by hand, next to the real one.
//
// A statement that is merely unfinished, like "Declare x to be", is not
// treated as more-to-come: it is reported straight away rather than leaving
// the prompt waiting for a line that cannot help.
func needsMoreInput(code string) bool {
	if strings.TrimSpace(code) == "" {
		return false
	}
	lexer := parser.NewLexer(code)
	_, err := parser.NewParser(lexer.TokenizeAll()).Parse()
	return parser.IsTruncated(err)
}
