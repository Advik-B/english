package cmd

import (
	"english/parser"
	"english/vm"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "english",
	Short: "English Language Interpreter",
	Long: `A programming language interpreter with natural English syntax.
Write code using English keywords and natural language constructs.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			StartREPL()
		} else {
			RunFile(args[0])
		}
	},
}

var replCmd = &cobra.Command{
	Use:   "repl",
	Short: "Start the interactive REPL",
	Long:  "Start the Read-Eval-Print Loop for interactive programming",
	Run: func(cmd *cobra.Command, args []string) {
		StartREPL()
	},
}

var runCmd = &cobra.Command{
	Use:   "run [file]",
	Short: "Run an English source file",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		RunFile(args[0])
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("English Language Interpreter v1.0.0")
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(replCmd)
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(versionCmd)
}

// RunFile executes an English source file
func RunFile(filename string) {
	content, err := os.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
		os.Exit(1)
	}

	env := vm.NewEnvironment()
	lexer := parser.NewLexer(string(content))
	tokens := lexer.TokenizeAll()

	p := parser.NewParser(tokens)
	program, err := p.Parse()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Parse error: %v\n", err)
		os.Exit(1)
	}

	evaluator := vm.NewEvaluator(env)
	_, err = evaluator.Eval(program)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Runtime error: %v\n", err)
		os.Exit(1)
	}
}
