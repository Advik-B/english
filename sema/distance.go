package sema

import "strings"

// editDistance returns the Levenshtein distance between a and b, compared
// case-insensitively, and is used to suggest a name the programmer may have
// meant.
//
// It is deterministic: the caller picks the lowest distance and breaks ties
// alphabetically. The evaluator's equivalent returns the first candidate
// within a threshold while iterating a Go map, so its suggestions vary between
// runs for the same program.
func editDistance(a, b string) int {
	x := []rune(strings.ToLower(a))
	y := []rune(strings.ToLower(b))
	if len(x) == 0 {
		return len(y)
	}
	if len(y) == 0 {
		return len(x)
	}

	// One row of the matrix is enough; prev[j] is the distance for y[:j].
	prev := make([]int, len(y)+1)
	cur := make([]int, len(y)+1)
	for j := range prev {
		prev[j] = j
	}

	for i := 1; i <= len(x); i++ {
		cur[0] = i
		for j := 1; j <= len(y); j++ {
			cost := 1
			if x[i-1] == y[j-1] {
				cost = 0
			}
			cur[j] = min3(cur[j-1]+1, prev[j]+1, prev[j-1]+cost)
		}
		prev, cur = cur, prev
	}
	return prev[len(y)]
}

func min3(a, b, c int) int {
	m := a
	if b < m {
		m = b
	}
	if c < m {
		m = c
	}
	return m
}
