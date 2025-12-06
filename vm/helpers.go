package vm

import "strings"

// levenshteinDistance calculates the Levenshtein distance between two strings
func levenshteinDistance(s1, s2 string) int {
	if len(s1) == 0 {
		return len(s2)
	}
	if len(s2) == 0 {
		return len(s1)
	}

	if len(s1) > len(s2) {
		s1, s2 = s2, s1
	}

	lenS1 := len(s1)
	lenS2 := len(s2)

	distances := make([]int, lenS1+1)
	for i := range distances {
		distances[i] = i
	}

	for i := 1; i <= lenS2; i++ {
		prev := i
		for j := 1; j <= lenS1; j++ {
			current := distances[j-1]
			if s2[i-1] != s1[j-1] {
				current = min(min(distances[j-1]+1, distances[j]+1), prev+1)
			}
			distances[j-1] = prev
			prev = current
		}
		distances[lenS1] = prev
	}

	return distances[lenS1]
}

// findSimilarName finds a similar name from a list of candidates
func findSimilarName(name string, candidates []string) string {
	name = strings.ToLower(name)

	// Simple similarity check (case-insensitive match or one-char difference)
	for _, candidate := range candidates {
		if strings.ToLower(candidate) == name {
			return candidate
		}
		if levenshteinDistance(strings.ToLower(candidate), name) <= 2 {
			return candidate
		}
	}

	return ""
}
