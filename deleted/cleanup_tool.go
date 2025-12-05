package main

import (
	"fmt"
	"os"
)

func main() {
	files := []string{
		"/workspaces/codespaces-blank/tokens.go",
		"/workspaces/codespaces-blank/lexer.go",
		"/workspaces/codespaces-blank/ast.go",
		"/workspaces/codespaces-blank/parser.go",
		"/workspaces/codespaces-blank/evaluator.go",
		"/workspaces/codespaces-blank/builtins.go",
	}

	for _, file := range files {
		if err := os.Remove(file); err != nil {
			fmt.Printf("Error removing %s: %v\n", file, err)
		} else {
			fmt.Printf("Removed %s\n", file)
		}
	}
}
