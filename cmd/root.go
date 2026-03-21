package cmd

import (
	"github.com/Advik-B/english/ast"
	"github.com/Advik-B/english/bytecode"
	"github.com/Advik-B/english/help"
	"github.com/Advik-B/english/highlight"
	"github.com/Advik-B/english/ivm"
	"github.com/Advik-B/english/parser"
	"github.com/Advik-B/english/stacktraces"
	"github.com/Advik-B/english/transpiler"
	"github.com/Advik-B/english/astvm"
	"github.com/Advik-B/english/stdlib"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

const Version = "v1.2.1"

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
		minPoliteness, _ := cmd.Flags().GetFloat64("minimum-politeness")
		politeFlag, _ := cmd.Flags().GetBool("polite")
		// --polite is a convenience shorthand for --minimum-politeness 100.
		// --minimum-politeness takes precedence when both are provided.
		if politeFlag && !cmd.Flags().Changed("minimum-politeness") {
			minPoliteness = 100
		}
		ext := strings.ToLower(filepath.Ext(filename))
		if ext == ".101" {
			// Politeness only applies to .abc source files.
			RunBytecode(filename)
		} else {
			if strings.EqualFold(vmFlag, "ast") {
				RunFileAST(filename, minPoliteness)
			} else {
				RunFileIVM(filename, minPoliteness)
			}
		}
	},
}

var compileCmd = &cobra.Command{
	Use:   "compile [file]",
	Short: "Compile an English source file (.abc) to bytecode (.101)",
	Long: `Compile an English source file to binary bytecode format.
The output file will have the same name with .101 extension.
Bytecode files can be executed directly without parsing.

By default the original source code is embedded as a trailing section in the
.101 file so that "english transpile" can later reconstruct idiomatic Python
without a separate .abc file.  Pass --strip (or --min) to omit the source
trailer and produce a smaller, standalone bytecode file.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		output, _ := cmd.Flags().GetString("output")
		strip, _ := cmd.Flags().GetBool("strip")
		min, _ := cmd.Flags().GetBool("min")
		CompileFileOptions(args[0], output, strip || min)
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
		fmt.Println(Version)
	},
}

var helpTopicCmd = &cobra.Command{
	Use:   "help-topic [topic]",
	Short: "Search and display help for English language features",
	Long: `Search for help topics using fuzzy matching. Provides detailed information
about language keywords, functions, operators, and concepts.

Examples:
  english help-topic print
  english help-topic loop
  english help-topic if`,
	Run: func(cmd *cobra.Command, args []string) {
		registry := help.NewRegistry()

		if len(args) == 0 {
			// Show all categories
			fmt.Println("English Language Help")
			fmt.Println(strings.Repeat("─", 50))
			fmt.Println("\nAvailable categories:")
			categories := registry.AllCategories()
			for _, cat := range categories {
				entries := registry.EntriesByCategory(cat)
				fmt.Printf("  %s (%d topics)\n", cat, len(entries))
			}
			fmt.Println("\nUsage: english help-topic <search-term>")
			fmt.Println("Examples: english help-topic print")
			fmt.Println("          english help-topic loop")
			return
		}

		// Search for the topic
		query := strings.Join(args, " ")
		results := registry.Search(query)

		if len(results) == 0 {
			fmt.Printf("No help topics found for '%s'.\n", query)
			return
		}

		// Show the best match in detail if it's a very good match
		if results[0].Score >= 700 {
			printDetailedHelp(results[0].Entry)
			// Show other related topics if available
			if len(results) > 1 && len(results) <= 5 {
				fmt.Println("\nRelated topics:")
				for i := 1; i < len(results) && i < 5; i++ {
					fmt.Printf("  - %s: %s\n",
						results[i].Entry.Name,
						results[i].Entry.Description)
				}
			}
		} else {
			// Show multiple search results
			fmt.Printf("Search results for '%s':\n", query)
			limit := 10
			if len(results) < limit {
				limit = len(results)
			}
			for i := 0; i < limit; i++ {
				entry := results[i].Entry
				fmt.Printf("  %s [%s]\n", entry.Name, entry.Category)
				fmt.Printf("    %s\n", entry.Description)
			}
			fmt.Println("\nUse 'english help-topic <topic>' for detailed information.")
		}
	},
}

func printDetailedHelp(entry *help.HelpEntry) {
	fmt.Printf("%s [%s]\n", entry.Name, entry.Category)
	fmt.Println(strings.Repeat("─", 50))
	fmt.Println(entry.Description)

	if entry.LongDesc != "" {
		fmt.Println()
		fmt.Println(entry.LongDesc)
	}

	if len(entry.Examples) > 0 {
		fmt.Println()
		fmt.Println("Examples:")
		useColor := stacktraces.HasColor()
		for _, example := range entry.Examples {
			// Apply syntax highlighting to examples
			highlighted := highlight.Highlight(example, useColor)
			fmt.Printf("  %s\n", highlighted)
		}
	}

	if len(entry.Aliases) > 0 {
		fmt.Println()
		fmt.Printf("Aliases: %s\n", strings.Join(entry.Aliases, ", "))
	}

	if len(entry.SeeAlso) > 0 {
		fmt.Println()
		fmt.Printf("See also: %s\n", strings.Join(entry.SeeAlso, ", "))
	}
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
	rootCmd.AddCommand(helpTopicCmd)

	compileCmd.Flags().StringP("output", "o", "", "Output file name (default: input file with .101 extension)")
	compileCmd.Flags().Bool("strip", false, "Omit the source trailer (smaller file; 'transpile' will use opcode decompiler)")
	compileCmd.Flags().Bool("min", false, "Alias for --strip")
	_ = compileCmd.Flags().MarkHidden("min") // expose only --strip in help; --min still works
	transpileCmd.Flags().BoolP("inline", "i", false, "Inline all imported .abc files into a single Python output file")
	runCmd.Flags().StringP("vm", "V", "ivm", "VM to use for running .abc source files: 'ivm' (default) or 'ast'")
	runCmd.Flags().Float64("minimum-politeness", -1,
		"Require at least this percentage (0–100) of statements to be polite "+
			"(prefixed with 'please', 'kindly', 'could you', or 'would you kindly'). "+
			"Only applies to .abc source files.")
	runCmd.Flags().Bool("polite", false,
		"Require all statements to be polite (equivalent to --minimum-politeness 100). "+
			"Only applies to .abc source files.")
}

// RunFile executes an English source file using the instruction VM (ivm) by default.
// This is a convenience wrapper for RunFileIVM.
func RunFile(filename string) {
	RunFileIVM(filename, -1)
}

// RunFileIVM parses and executes an English source file via the instruction-based VM.
// It is the default execution path for .abc source files.
// minPoliteness is the minimum required politeness percentage (0–100); pass a
// negative value to disable the check.
func RunFileIVM(filename string, minPoliteness float64) {
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

	if minPoliteness >= 0 {
		if errs := checkPoliteness(program, minPoliteness); len(errs) > 0 {
			for _, e := range errs {
				stacktraces.Print(e)
			}
			os.Exit(1)
		}
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
// minPoliteness is the minimum required politeness percentage (0–100); pass a
// negative value to disable the check.
func RunFileAST(filename string, minPoliteness float64) {
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

	if minPoliteness >= 0 {
		if errs := checkPoliteness(program, minPoliteness); len(errs) > 0 {
			for _, e := range errs {
				stacktraces.Print(e)
			}
			os.Exit(1)
		}
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

// CompileFile compiles an English source file to bytecode, embedding the
// original source as a trailing section (for later transpilation).
// This is a convenience wrapper around CompileFileOptions.
func CompileFile(filename string, output string) {
	CompileFileOptions(filename, output, false)
}

// CompileFileOptions compiles an English source file to bytecode.
// When stripSource is true the source trailer is omitted, producing a smaller
// file; "english transpile" will then fall back to opcode decompilation.
func CompileFileOptions(filename string, output string, stripSource bool) {
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

	var data []byte
	var encodeErr error
	if stripSource {
		data, encodeErr = ivm.EncodeFile(chunk)
	} else {
		data, encodeErr = ivm.EncodeFileWithSource(chunk, string(content))
	}
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
		// When no embedded source is present, fall back to direct opcode decompilation.
		if len(data) >= 5 && data[4] == ivm.InstructionFormatVersion {
			chunk, embeddedSrc, decodeErr := ivm.DecodeFileAll(data)
			if decodeErr != nil {
				fmt.Fprintf(os.Stderr, "Bytecode error: %v\n", decodeErr)
				os.Exit(1)
			}
			if embeddedSrc != "" {
				// Preferred path: re-parse the embedded source and use the full
				// AST transpiler for idiomatic, comment-preserving output.
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
			} else {
				// Fallback: decompile from pure opcode stream.
				// Comments are lost but all logic is preserved.
				pySource = ivm.Decompile(chunk)
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

// checkPoliteness verifies that the program meets the minimum politeness
// percentage.  It returns one *parser.SyntaxError per impolite statement when
// the threshold is not met, or nil when the program is sufficiently polite.
// This check only applies to .abc source files – bytecode execution paths
// never call this function.
func checkPoliteness(program *ast.Program, minPercent float64) []error {
if program.TotalCount == 0 {
return nil
}

actual := float64(program.PoliteCount) / float64(program.TotalCount) * 100
if actual >= minPercent {
return nil
}

var errs []error
for _, line := range program.ImpoliteLines {
errs = append(errs, &parser.SyntaxError{
Msg:  "Statement is not polite.",
Line: line,
Hint: "Prefix the statement with 'Please', 'Kindly', 'Could you', or 'Would you kindly'.",
})
}
return errs
}
