package cmd

import (
	"english/bytecode"
	"english/parser"
	"english/repl"
	"english/vm"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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
	Long: `Start the Read-Eval-Print Loop for interactive programming.
By default, starts with a beautiful TUI interface using Charm Bracelet libraries.
Use --simple flag for a plain REPL suitable for pipes, scripts, or automation.`,
	Run: func(cmd *cobra.Command, args []string) {
		simple, err := cmd.Flags().GetBool("simple")
		if err != nil {
			// This should never happen for a boolean flag defined in init()
			fmt.Fprintf(os.Stderr, "Error reading simple flag: %v\n", err)
			os.Exit(1)
		}
		if simple {
			StartSimpleREPL()
		} else {
			StartREPL()
		}
	},
}

var runCmd = &cobra.Command{
	Use:   "run [file]",
	Short: "Run an English source file (.abc) or bytecode file (.101)",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		filename := args[0]
		ext := strings.ToLower(filepath.Ext(filename))
		if ext == ".101" {
			RunBytecode(filename)
		} else {
			RunFile(filename)
		}
	},
}

var compileCmd = &cobra.Command{
	Use:   "compile [file]",
	Short: "Compile an English source file (.abc) to bytecode (.101)",
	Long: `Compile an English source file to binary bytecode format.
The output file will have the same name with .101 extension.
Bytecode files can be executed directly without parsing.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		output, _ := cmd.Flags().GetString("output")
		CompileFile(args[0], output)
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
	rootCmd.AddCommand(compileCmd)
	rootCmd.AddCommand(versionCmd)

	replCmd.Flags().BoolP("simple", "s", false, "Start simple REPL (no TUI) for pipes, scripts, or automation")
	compileCmd.Flags().StringP("output", "o", "", "Output file name (default: input file with .101 extension)")
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

// CompileFile compiles an English source file to bytecode
func CompileFile(filename string, output string) {
	content, err := os.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
		os.Exit(1)
	}

	lexer := parser.NewLexer(string(content))
	tokens := lexer.TokenizeAll()

	p := parser.NewParser(tokens)
	program, err := p.Parse()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Parse error: %v\n", err)
		os.Exit(1)
	}

	encoder := bytecode.NewEncoder()
	data, err := encoder.Encode(program)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Compile error: %v\n", err)
		os.Exit(1)
	}

	// Determine output filename
	if output == "" {
		ext := filepath.Ext(filename)
		if ext == "" {
			output = filename + ".101"
		} else {
			output = filename[:len(filename)-len(ext)] + ".101"
		}
	}

	err = os.WriteFile(output, data, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing bytecode file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Compiled %s -> %s (%d bytes)\n", filename, output, len(data))
}

// RunBytecode executes a compiled bytecode file
func RunBytecode(filename string) {
	data, err := os.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
		os.Exit(1)
	}

	decoder := bytecode.NewDecoder(data)
	program, err := decoder.Decode()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Bytecode error: %v\n", err)
		os.Exit(1)
	}

	env := vm.NewEnvironment()
	evaluator := vm.NewEvaluator(env)
	_, err = evaluator.Eval(program)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Runtime error: %v\n", err)
		os.Exit(1)
	}
}

// StartSimpleREPL starts the simple non-TUI REPL
func StartSimpleREPL() {
	console := repl.NewConsole()
	if err := console.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "REPL error: %v\n", err)
		os.Exit(1)
	}
}
