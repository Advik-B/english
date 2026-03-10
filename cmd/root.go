package cmd

import (
	"english/ast"
	"english/bytecode"
	"english/ivm"
	"english/parser"
	"english/stacktraces"
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
		vmFlag, _ := cmd.Flags().GetString("vm")
		ext := strings.ToLower(filepath.Ext(filename))
		if ext == ".101" {
			RunBytecode(filename)
		} else {
			if strings.EqualFold(vmFlag, "ast") {
				RunFileAST(filename)
			} else {
				RunFileIVM(filename)
			}
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
The output file has the source filename with its extension replaced by ".py".

By default each imported .abc file is transpiled to its own .py file and
imported in the main output via standard Python "from module import *".
Pass --inline to instead merge all imported code into a single self-contained
Python file.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		inline, _ := cmd.Flags().GetBool("inline")
		transpileWithOptions(args[0], inline, make(map[string]bool))
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
	transpileCmd.Flags().BoolP("inline", "i", false, "Inline all imported .abc files into a single Python output file")
	runCmd.Flags().StringP("vm", "V", "ivm", "VM to use for running .abc source files: 'ivm' (default) or 'ast'")
}

// RunFile executes an English source file using the instruction VM (ivm) by default.
// This is a convenience wrapper for RunFileIVM.
func RunFile(filename string) {
	RunFileIVM(filename)
}

// RunFileIVM parses and executes an English source file via the instruction-based VM.
// It is the default execution path for .abc source files.
func RunFileIVM(filename string) {
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
		stacktraces.Print(err)
		os.Exit(1)
	}

	typeErrs := vm.Check(program, stdlib.PredefinedNames()...)
	if len(typeErrs) > 0 {
		for _, e := range typeErrs {
			stacktraces.Print(e)
		}
		os.Exit(1)
	}

	chunk, compileErr := ivm.Compile(program)
	if compileErr != nil {
		fmt.Fprintf(os.Stderr, "Compile error: %v\n", compileErr)
		os.Exit(1)
	}

	_, execErr := ivm.Execute(chunk, stdlib.Eval, stdlib.PredefinedValues())
	if execErr != nil {
		stacktraces.Print(execErr)
		os.Exit(1)
	}
}

// RunFileAST parses and executes an English source file via the tree-walk evaluator.
// Use the --vm=ast flag on the run command to select this path.
func RunFileAST(filename string) {
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
		stacktraces.Print(err)
		os.Exit(1)
	}

	typeErrs := vm.Check(program, stdlib.PredefinedNames()...)
	if len(typeErrs) > 0 {
		for _, e := range typeErrs {
			stacktraces.Print(e)
		}
		os.Exit(1)
	}

	evaluator := vm.NewEvaluator(env, stdlib.Eval)
	_, err = evaluator.Eval(program)
	if err != nil {
		stacktraces.Print(err)
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
		stacktraces.Print(err)
		os.Exit(1)
	}

	typeErrs := vm.Check(program, stdlib.PredefinedNames()...)
	if len(typeErrs) > 0 {
		for _, e := range typeErrs {
			stacktraces.Print(e)
		}
		os.Exit(1)
	}

	chunk, compileErr := ivm.Compile(program)
	if compileErr != nil {
		fmt.Fprintf(os.Stderr, "Compile error: %v\n", compileErr)
		os.Exit(1)
	}

	data, encodeErr := ivm.EncodeFileWithSource(chunk, string(content))
	if encodeErr != nil {
		fmt.Fprintf(os.Stderr, "Encode error: %v\n", encodeErr)
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

// TranspileFile validates an English source or bytecode file and transpiles it
// to Python. This is a convenience wrapper that defaults to non-inline mode
// (each imported .abc file becomes a sibling .py file).
// The output file has the source filename with its extension replaced by ".py".
//
//	examples/fizzbuzz.abc  → examples/fizzbuzz.py
//	examples/fizzbuzz.101  → examples/fizzbuzz.py
func TranspileFile(filename string) {
	transpileWithOptions(filename, false, make(map[string]bool))
}

// transpileWithOptions is the recursive worker for TranspileFile.
// 'seen' prevents duplicate transpilation and infinite loops for circular imports.
func transpileWithOptions(filename string, inline bool, seen map[string]bool) {
	if seen[filename] {
		return
	}
	seen[filename] = true

	ext := strings.ToLower(filepath.Ext(filename))

	// Output filename: strip source extension, add ".py".
	// The same rule applies to both .abc and .101 inputs so that module names
	// are always valid Python identifiers (e.g. "fizzbuzz.abc" → "fizzbuzz.py").
	output := strings.TrimSuffix(filename, filepath.Ext(filename)) + ".py"

	var pySource string
	if ext == ".101" {
		data, err := os.ReadFile(filename)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
			os.Exit(1)
		}

		// Detect format: version 2 = ivm instruction format, version 1 = AST format.
		// v2 files compiled with `english compile` carry the original source code as
		// a trailing section; extract it and parse normally so the transpiler works
		// from a full AST (identical output to transpiling the .abc directly).
		if len(data) >= 5 && data[4] == ivm.InstructionFormatVersion {
			_, embeddedSrc, decodeErr := ivm.DecodeFileAll(data)
			if decodeErr != nil {
				fmt.Fprintf(os.Stderr, "Bytecode error: %v\n", decodeErr)
				os.Exit(1)
			}
			if embeddedSrc == "" {
				srcFile := strings.TrimSuffix(filename, filepath.Ext(filename)) + ".abc"
				fmt.Fprintf(os.Stderr,
					"Error: %q was compiled without embedded source and cannot be transpiled.\n"+
						"Recompile with 'english compile %s' and try again.\n",
					filename, srcFile)
				os.Exit(1)
			}
			lexer := parser.NewLexer(embeddedSrc)
			tokens := lexer.TokenizeAll()
			p := parser.NewParser(tokens)
			prog, parseErr := p.Parse()
			if parseErr != nil {
				stacktraces.Print(parseErr)
				os.Exit(1)
			}
			typeErrs := vm.Check(prog, stdlib.PredefinedNames()...)
			if len(typeErrs) > 0 {
				for _, e := range typeErrs {
					stacktraces.Print(e)
				}
				os.Exit(1)
			}
			if inline {
				pySource = transpiler.NewTranspilerInlined().Transpile(prog)
			} else {
				sourceDir := filepath.Dir(filename)
				pySource = transpiler.NewTranspiler().WithSourceDir(sourceDir).Transpile(prog)
			}
			if writeErr := os.WriteFile(output, []byte(pySource), 0644); writeErr != nil {
				fmt.Fprintf(os.Stderr, "Error writing Python file: %v\n", writeErr)
				os.Exit(1)
			}
			fmt.Printf("Transpiled %s -> %s\n", filename, output)
			return
		}

		decoder := bytecode.NewDecoder(data)
		prog, err := decoder.Decode()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Bytecode error: %v\n", err)
			os.Exit(1)
		}

		typeErrs := vm.Check(prog, stdlib.PredefinedNames()...)
		if len(typeErrs) > 0 {
			for _, e := range typeErrs {
				stacktraces.Print(e)
			}
			os.Exit(1)
		}

		pySource = transpiler.NewTranspilerStripped().Transpile(prog)
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
			stacktraces.Print(err)
			os.Exit(1)
		}

		typeErrs := vm.Check(prog, stdlib.PredefinedNames()...)
		if len(typeErrs) > 0 {
			for _, e := range typeErrs {
				stacktraces.Print(e)
			}
			os.Exit(1)
		}

		if inline {
			// --inline: resolve all imports by inlining their ASTs into a single file.
			pySource = transpiler.NewTranspilerInlined().Transpile(prog)
		} else {
			// Default: recursively transpile each imported .abc file to its own .py
			// file, then emit "from module import *" statements in the main output.
			// Only files with an explicit ".abc" extension are transpiled; other
			// import paths (e.g. bare "math") are left for Python to resolve.
			for _, stmt := range prog.Statements {
				imp, ok := stmt.(*ast.ImportStatement)
				if !ok {
					continue
				}
				if strings.ToLower(filepath.Ext(imp.Path)) == ".abc" {
					transpileWithOptions(imp.Path, false, seen)
				}
			}
			sourceDir := filepath.Dir(filename)
			pySource = transpiler.NewTranspiler().WithSourceDir(sourceDir).Transpile(prog)
		}
	}

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

	// Detect format version: version 2 = instruction-based ivm format
	if len(data) >= 5 && data[4] == ivm.InstructionFormatVersion {
		chunk, decodeErr := ivm.DecodeFile(data)
		if decodeErr != nil {
			fmt.Fprintf(os.Stderr, "Bytecode error: %v\n", decodeErr)
			os.Exit(1)
		}
		_, execErr := ivm.Execute(chunk, stdlib.Eval, stdlib.PredefinedValues())
		if execErr != nil {
			stacktraces.Print(execErr)
			os.Exit(1)
		}
		return
	}

	// Version 1: AST-based format
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
		stacktraces.Print(err)
		os.Exit(1)
	}
}
