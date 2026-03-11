package cmd

// inspect_ivm.go – "english inspect-ivm [file]" command
//
// Displays a raw ivm opcode listing of a compiled .101 bytecode file or an
// English source file (.abc), without running or transpiling it.
//
// This complements "english inspect" (which shows the AST-level view) by
// working directly with the instruction representation used by the ivm VM.

import (
	"english/ivm"
	"english/parser"
	"english/stacktraces"
	"english/vm"
	"english/vm/stdlib"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var inspectIVMCmd = &cobra.Command{
	Use:     "inspect-ivm [file]",
	Aliases: []string{"disasm-ivm", "ivm-inspect"},
	Short:   "Show the raw ivm opcode listing of a bytecode (.101) or source (.abc) file",
	Long: `Compile and display the ivm instruction listing for an English source (.abc)
or bytecode (.101) file without running it.

The output shows every instruction in the chunk together with the constant pool,
name table, and all nested sub-chunks (functions, struct methods, default-value
expressions).  It is the ivm equivalent of Python's "dis" module.

For .abc source files the program is parsed, type-checked and compiled in memory;
no .101 file is written to disk.

For .101 bytecode files the binary is decoded directly.  If the file embeds the
original source (as produced by the current compiler) that source is not
displayed; the listing is always derived from the opcode stream.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		filename := args[0]
		ext := strings.ToLower(filepath.Ext(filename))
		useColor := stacktraces.HasColor()

		var chunk *ivm.Chunk

		if ext == ".101" {
			data, err := os.ReadFile(filename)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
				os.Exit(1)
			}
			decoded, _, decodeErr := ivm.DecodeFileAll(data)
			if decodeErr != nil {
				fmt.Fprintf(os.Stderr, "Error decoding bytecode: %v\n", decodeErr)
				os.Exit(1)
			}
			chunk = decoded
		} else {
			// .abc or any other extension – parse, check and compile in memory.
			content, err := os.ReadFile(filename)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
				os.Exit(1)
			}
			lexer := parser.NewLexer(string(content))
			tokens := lexer.TokenizeAll()
			p := parser.NewParser(tokens)
			program, parseErr := p.Parse()
			if parseErr != nil {
				stacktraces.Print(parseErr)
				os.Exit(1)
			}
			typeErrs := vm.Check(program, stdlib.PredefinedNames()...)
			if len(typeErrs) > 0 {
				for _, e := range typeErrs {
					stacktraces.Print(e)
				}
				os.Exit(1)
			}
			compiled, compileErr := ivm.Compile(program)
			if compileErr != nil {
				fmt.Fprintf(os.Stderr, "Compile error: %v\n", compileErr)
				os.Exit(1)
			}
			chunk = compiled
		}

		title := fmt.Sprintf("<module: %s>", filepath.Base(filename))
		fmt.Print(ivm.Listing(chunk, title, useColor))
	},
}

func init() {
	rootCmd.AddCommand(inspectIVMCmd)
}
