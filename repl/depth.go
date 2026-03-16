package repl

import "strings"

// isBlockOpener reports whether a trimmed, lower-cased line opens a new block.
//
// Block-openers:
//   - Any line that contains "following" and ends with ":" covers all forms:
//     "do the following:", "does the following:",
//     "repeat the following while …:", "Try doing the following:", etc.
//   - Any line that ends with " then" (covers "If …, then")
//
// Exception: lines that start with "otherwise" are continuations of an
// existing if-else chain and do not open a new depth level, even when they
// themselves end with " then" (e.g. "otherwise if …, then").
// Similarly, "on ErrorType:" and "but finally:" are catch/finally clauses
// inside an existing try block and do not affect depth.
func isBlockOpener(lower string) bool {
	// Continuation branches never open a new block level.
	if strings.HasPrefix(lower, "otherwise") {
		return false
	}
	if strings.HasPrefix(lower, "on ") && strings.HasSuffix(lower, ":") {
		return false
	}
	if strings.HasPrefix(lower, "but finally") {
		return false
	}

	// "do the following:", "repeat the following while …:", etc.
	if strings.Contains(lower, "following") && strings.HasSuffix(lower, ":") {
		return true
	}
	if strings.HasSuffix(lower, " then") {
		return true
	}
	return false
}

// isBlockCloser reports whether a trimmed, lower-cased line closes a block.
func isBlockCloser(lower string) bool {
	return strings.Contains(lower, "thats it.") || strings.Contains(lower, "that's it.")
}
