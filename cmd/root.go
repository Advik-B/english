package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Advik-B/english/ast"
	vm "github.com/Advik-B/english/astvm"
	"github.com/Advik-B/english/bytecode"
	"github.com/Advik-B/english/help"
	"github.com/Advik-B/english/highlight"
	"github.com/Advik-B/english/ivm"
	"github.com/Advik-B/english/parser"
	"github.com/Advik-B/english/sema"
	"github.com/Advik-B/english/stacktraces"
	"github.com/Advik-B/english/stdlib"
	"github.com/Advik-B/english/transpiler"
	"github.com/Advik-B/english/version"
	"github.com/spf13/cobra"
)

// Command-line flags, bound at registration.
//
// These used to be read back by name — cmd.Flags().GetString("vm") — with the
// error discarded at each of seven call sites. That error reports a misspelled
// flag name or the wrong accessor for its type, which is a mistake in this
// file rather than anything a user can cause, and discarding it turned such a
// mistake into a silent zero value. Binding leaves no name to misspell and no
// error to ignore.
var (
	runVM              string
	runMinPoliteness   float64
	runPolite          bool
	compileOutput      string
	compileStrip       bool
	compileMin         bool
	transpileInlineAll bool
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
		engine, err := engineNamed(runVM)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		minPoliteness := runMinPoliteness
		// --polite is a convenience shorthand for --minimum-politeness 100.
		// --minimum-politeness takes precedence when both are provided.
		if runPolite && !cmd.Flags().Changed("minimum-politeness") {
			minPoliteness = 100
		}

		if strings.ToLower(filepath.Ext(filename)) == ".101" {
			// Politeness only applies to .abc source files.
			RunBytecode(filename)
			return
		}
		if engine == engineAST {
			RunFileAST(filename, minPoliteness)
			return
		}
		RunFileIVM(filename, minPoliteness)
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
		CompileFileOptions(args[0], compileOutput, compileStrip || compileMin)
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
		if err := TranspileFileOptions(args[0], transpileInlineAll); err != nil {
			report(err)
			os.Exit(1)
		}
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number and check for updates",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(version.Version)
		CheckForUpdates()
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
	rootCmd.AddCommand(updatesCmd)

	compileCmd.Flags().StringVarP(&compileOutput, "output", "o", "", "Output file name (default: input file with .101 extension)")
	compileCmd.Flags().BoolVar(&compileStrip, "strip", false, "Omit the source trailer (smaller file; 'transpile' will use opcode decompiler)")
	compileCmd.Flags().BoolVar(&compileMin, "min", false, "Alias for --strip")
	// Expose only --strip in help; --min still works.
	mustHide(compileCmd, "min")
	transpileCmd.Flags().BoolVarP(&transpileInlineAll, "inline", "i", false, "Inline all imported .abc files into a single Python output file")
	runCmd.Flags().StringVarP(&runVM, "vm", "V", engineIVM, "VM to use for running .abc source files: 'ivm' (default) or 'ast'")
	runCmd.Flags().Float64Var(&runMinPoliteness, "minimum-politeness", -1,
		"Require at least this percentage (0–100) of statements to be polite "+
			"(prefixed with 'please', 'kindly', 'could you', or 'would you kindly'). "+
			"Only applies to .abc source files.")
	runCmd.Flags().BoolVar(&runPolite, "polite", false,
		"Require all statements to be polite (equivalent to --minimum-politeness 100). "+
			"Only applies to .abc source files.")
}

// The engines a program can be run with.
const (
	engineIVM = "ivm"
	engineAST = "ast"
)

// engineNamed resolves the --vm flag.
//
// An unrecognised value used to fall through to the instruction VM, so
// "--vm=asr" ran a different engine than the one asked for and said nothing.
// The two engines are meant to agree, but the flag exists precisely because
// which one ran is worth knowing.
func engineNamed(name string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case engineIVM, "":
		return engineIVM, nil
	case engineAST:
		return engineAST, nil
	}
	return "", fmt.Errorf("unknown VM %q\n  Hint: --vm takes '%s' (the default) or '%s'",
		name, engineIVM, engineAST)
}

// mustHide hides a flag from the help output.
//
// A failure means the name does not match the flag registered just above it,
// which is a mistake here rather than anything a user can cause. The two call
// sites disagreed about that: one panicked and the other discarded the error.
func mustHide(cmd *cobra.Command, name string) {
	if err := cmd.Flags().MarkHidden(name); err != nil {
		panic(fmt.Sprintf("cannot hide the --%s flag of %q: %v", name, cmd.Name(), err))
	}
}

// report prints an error the way the command line should: a set of semantic
// diagnostics one at a time, and anything else through the renderer.
func report(err error) {
	var diags diagnostics
	if errors.As(err, &diags) {
		for _, d := range diags {
			stacktraces.Print(d)
		}
		return
	}
	stacktraces.Print(err)
}

// diagnostics is everything wrong with a program, reported together.
type diagnostics []*sema.Diagnostic

func (d diagnostics) Error() string {
	switch len(d) {
	case 0:
		return "no problems"
	case 1:
		return d[0].Error()
	default:
		return fmt.Sprintf("%s (and %d more problem(s))", d[0].Error(), len(d)-1)
	}
}

// check type-checks a program and returns everything wrong with it.
//
// Every entry point goes through here, so that running a program, compiling
// it, disassembling it and transpiling it all apply the same rules. The
// version-1 bytecode path used to skip checking entirely.
func check(prog *ast.Program, filename string) error {
	diags := sema.Check(prog, sema.Config{
		Predefined: stdlib.PredefinedNames(),
		Dir:        filepath.Dir(filename),
	})
	if len(diags) == 0 {
		return nil
	}
	return diagnostics(diags)
}

// analyse type-checks a program and exits with a report if anything is wrong.
// It is for the entry points that own the process; anything that recurses
// calls check and returns the error instead.
func analyse(prog *ast.Program, filename string) {
	if err := check(prog, filename); err != nil {
		report(err)
		os.Exit(1)
	}
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

	analyse(program, filename)

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

	analyse(program, filename)

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

	analyse(program, filename)

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

// TranspileFileOptions validates an English source or bytecode file and
// transpiles it to Python, along with everything it imports.
//
// The output file has the source filename with its extension replaced by
// ".py", and so does each imported .abc file unless inline is set:
//
//	examples/fizzbuzz.abc  → examples/fizzbuzz.py
//	examples/fizzbuzz.101  → examples/fizzbuzz.py
//
// Nothing is written until the whole tree has been translated. The recursion
// used to write each file as it finished and call os.Exit from twelve places
// along the way, so a syntax error in the third imported file left a partly
// written tree of .py files behind — some new, some left over from a previous
// run, with no way to tell which — and the error was reported by a function
// several frames deep that could not be recovered from or tested.
func TranspileFileOptions(filename string, inline bool) error {
	var outputs []pythonOutput
	if err := transpileTree(filename, inline, make(map[string]bool), &outputs); err != nil {
		return err
	}
	for _, out := range outputs {
		if err := os.WriteFile(out.path, []byte(out.source), 0644); err != nil {
			return fmt.Errorf("cannot write the Python file %s: %w", out.path, err)
		}
		fmt.Printf("Transpiled %s -> %s\n", out.from, out.path)
	}
	return nil
}

// pythonOutput is a translated file waiting to be written.
type pythonOutput struct {
	from   string // the file it was translated from
	path   string // where it goes
	source string
}

// transpileTree translates a file and, unless inlining, everything it imports.
// The seen set prevents duplicate work and infinite recursion on a circular
// import.
func transpileTree(filename string, inline bool, seen map[string]bool, outputs *[]pythonOutput) error {
	if seen[filename] {
		return nil
	}
	seen[filename] = true

	// Output filename: strip source extension, add ".py".
	// The same rule applies to both .abc and .101 inputs so that module names
	// are always valid Python identifiers (e.g. "fizzbuzz.abc" → "fizzbuzz.py").
	output := strings.TrimSuffix(filename, filepath.Ext(filename)) + ".py"

	var pySource string
	var err error
	if strings.EqualFold(filepath.Ext(filename), ".101") {
		pySource, err = transpileBytecode(filename, inline)
	} else {
		pySource, err = transpileSource(filename, inline, seen, outputs)
	}
	if err != nil {
		return err
	}

	*outputs = append(*outputs, pythonOutput{from: filename, path: output, source: pySource})
	return nil
}

// transpileBytecode translates a .101 file.
//
// A file compiled with "english compile" carries the original source as a
// trailing section; that is re-parsed so the output is identical to
// transpiling the .abc directly, comments included. With no embedded source,
// the opcode stream is decompiled instead: comments are lost, logic is not.
func transpileBytecode(filename string, inline bool) (string, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return "", fmt.Errorf("cannot read %s: %w", filename, err)
	}

	// Two formats share the same magic bytes and are told apart by the version
	// byte: the instruction format this engine writes, and the older AST one.
	if len(data) >= 5 && data[4] == ivm.InstructionFormatVersion {
		chunk, embeddedSrc, err := ivm.DecodeFileAll(data)
		if err != nil {
			return "", fmt.Errorf("bytecode error in %s: %w", filename, err)
		}
		if embeddedSrc == "" {
			return ivm.Decompile(chunk), nil
		}
		prog, err := parseText(embeddedSrc)
		if err != nil {
			return "", err
		}
		if err := check(prog, filename); err != nil {
			return "", err
		}
		if inline {
			return transpiler.NewTranspilerInlined().Transpile(prog), nil
		}
		return transpiler.NewTranspiler().WithSourceDir(filepath.Dir(filename)).Transpile(prog), nil
	}

	prog, err := bytecode.NewDecoder(data).Decode()
	if err != nil {
		return "", fmt.Errorf("bytecode error in %s: %w", filename, err)
	}
	if err := check(prog, filename); err != nil {
		return "", err
	}
	return transpiler.NewTranspilerStripped().Transpile(prog), nil
}

// transpileSource translates an .abc file, first translating what it imports
// unless everything is being inlined into one output.
func transpileSource(filename string, inline bool, seen map[string]bool, outputs *[]pythonOutput) (string, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return "", fmt.Errorf("cannot read %s: %w", filename, err)
	}
	prog, err := parseText(string(content))
	if err != nil {
		return "", err
	}
	if err := check(prog, filename); err != nil {
		return "", err
	}

	if inline {
		// --inline: resolve all imports by inlining their ASTs into one file.
		return transpiler.NewTranspilerInlined().Transpile(prog), nil
	}

	// Default: recursively transpile each imported .abc file to its own .py
	// file, then emit "from module import *" statements in the main output.
	// Only files with an explicit ".abc" extension are transpiled; other
	// import paths (e.g. bare "math") are left for Python to resolve.
	for _, stmt := range prog.Statements {
		imp, ok := stmt.(*ast.ImportStatement)
		if !ok {
			continue
		}
		if !strings.EqualFold(filepath.Ext(imp.Path), ".abc") {
			continue
		}
		if err := transpileTree(imp.Path, false, seen, outputs); err != nil {
			return "", err
		}
	}
	return transpiler.NewTranspiler().WithSourceDir(filepath.Dir(filename)).Transpile(prog), nil
}

// parseText lexes and parses source text.
func parseText(source string) (*ast.Program, error) {
	lexer := parser.NewLexer(source)
	return parser.NewParser(lexer.TokenizeAll()).Parse()
}

func RunBytecode(filename string) {
	data, err := os.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
		os.Exit(1)
	}

	// Told apart by the version byte; see the note in transpileWithOptions.
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

	// The older AST-based format.
	decoder := bytecode.NewDecoder(data)
	program, err := decoder.Decode()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Bytecode error: %v\n", err)
		os.Exit(1)
	}

	analyse(program, filename)

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
