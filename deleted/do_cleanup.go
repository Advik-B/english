package main

import (
	"fmt"
	"os"
)

func main() {
	files := []string{
		"tokens.go",
		"lexer.go",
		"ast.go",
		"parser.go",
		"evaluator.go",
		"builtins.go",
		"cleanup_tool.go",
	}

	for _, file := range files {
		if err := os.Remove(file); err == nil {
			fmt.Printf("Removed %s\n", file)
		}
	}

	// Remove directories
	os.RemoveAll("repl")
	os.RemoveAll("internal")
	
	fmt.Println("Cleanup complete!")
}
