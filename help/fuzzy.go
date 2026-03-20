package help

// levenshteinDistance computes the edit distance between two strings.
// This is the minimum number of single-character edits (insertions, deletions, or substitutions)
// required to change one string into the other.
func levenshteinDistance(s1, s2 string) int {
	// Handle empty strings
	if len(s1) == 0 {
		return len(s2)
	}
	if len(s2) == 0 {
		return len(s1)
	}

	// Create a 2D matrix for dynamic programming
	rows := len(s1) + 1
	cols := len(s2) + 1
	matrix := make([][]int, rows)
	for i := range matrix {
		matrix[i] = make([]int, cols)
	}

	// Initialize first row and column
	for i := 0; i < rows; i++ {
		matrix[i][0] = i
	}
	for j := 0; j < cols; j++ {
		matrix[0][j] = j
	}

	// Fill in the matrix
	for i := 1; i < rows; i++ {
		for j := 1; j < cols; j++ {
			cost := 1
			if s1[i-1] == s2[j-1] {
				cost = 0
			}

			// Take minimum of:
			// - deletion: matrix[i-1][j] + 1
			// - insertion: matrix[i][j-1] + 1
			// - substitution: matrix[i-1][j-1] + cost
			matrix[i][j] = min(
				matrix[i-1][j]+1,
				min(matrix[i][j-1]+1, matrix[i-1][j-1]+cost),
			)
		}
	}

	return matrix[rows-1][cols-1]
}

// min returns the minimum of two integers.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
