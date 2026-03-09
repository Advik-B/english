package cmd

import (
	"english/bytecode"
	"english/parser"
	"english/transpiler"
	"english/vm"
	"english/vm/stdlib"
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
	Long:  "Start the Read-Eval-Print Loop with beautiful TUI interface for interactive programming",
	Run: func(cmd *cobra.Command, args []string) {
		StartREPL()
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

var transpileCmd = &cobra.Command{
	Use:   "transpile [file]",
	Short: "Transpile an English source or bytecode file to Python",
	Long: `Transpile an English source file (.abc) or bytecode file (.101) to human-readable Python.
The program is validated (parsed and type-checked) before transpilation.
The output file will have the same name with .py added as a suffix.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		TranspileFile(args[0])
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
	rootCmd.AddCommand(transpileCmd)
	rootCmd.AddCommand(versionCmd)

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
	stdlib.Register(env)
	lexer := parser.NewLexer(string(content))
	tokens := lexer.TokenizeAll()

	p := parser.NewParser(tokens)
	program, err := p.Parse()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Parse error: %v\n", err)
		os.Exit(1)
	}

	typeErrs := vm.Check(program)
	if len(typeErrs) > 0 {
		for _, e := range typeErrs {
			fmt.Fprintln(os.Stderr, e.Error())
		}
		os.Exit(1)
	}

	evaluator := vm.NewEvaluator(env, stdlib.Eval)
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

	typeErrs := vm.Check(program)
	if len(typeErrs) > 0 {
		for _, e := range typeErrs {
			fmt.Fprintln(os.Stderr, e.Error())
		}
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

// TranspileFile validates an English source or bytecode file and transpiles it to Python.
// The output is written to the same filename with ".py" appended as a suffix.
func TranspileFile(filename string) {
	ext := strings.ToLower(filepath.Ext(filename))

	// Parse or decode the input file into an AST, then validate it.
	var pySource string
	if ext == ".101" {
		data, err := os.ReadFile(filename)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
			os.Exit(1)
		}

		decoder := bytecode.NewDecoder(data)
		prog, err := decoder.Decode()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Bytecode error: %v\n", err)
			os.Exit(1)
		}

		// Validate: run the type checker
		typeErrs := vm.Check(prog)
		if len(typeErrs) > 0 {
			for _, e := range typeErrs {
				fmt.Fprintln(os.Stderr, e.Error())
			}
			os.Exit(1)
		}

		pySource = transpiler.NewTranspiler().Transpile(prog)
	} else {
		content, err := os.ReadFile(filename)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
			os.Exit(1)
		}

		lexer := parser.NewLexer(string(content))
		tokens := lexer.TokenizeAll()

		p := parser.NewParser(tokens)
		prog, err := p.Parse()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Parse error: %v\n", err)
			os.Exit(1)
		}

		// Validate: run the type checker
		typeErrs := vm.Check(prog)
		if len(typeErrs) > 0 {
			for _, e := range typeErrs {
				fmt.Fprintln(os.Stderr, e.Error())
			}
			os.Exit(1)
		}

		pySource = transpiler.NewTranspiler().Transpile(prog)
	}

	output := filename + ".py"
	err := os.WriteFile(output, []byte(pySource), 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing Python file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Transpiled %s -> %s\n", filename, output)
}
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
	stdlib.Register(env)
	evaluator := vm.NewEvaluator(env, stdlib.Eval)
	_, err = evaluator.Eval(program)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Runtime error: %v\n", err)
		os.Exit(1)
	}
}
