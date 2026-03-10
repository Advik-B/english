package cmd

import (
	"english/bytecode"
	"english/bytecode/disasm"
	"english/parser"
	"english/stacktraces"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var inspectFriendly bool

var inspectCmd = &cobra.Command{
	Use:     "inspect [file]",
	Aliases: []string{"disassemble", "decompile"},
	Short:   "Disassemble a bytecode (.101) or source (.abc) file to an instruction-level listing",
	Long: `Decode and print a colourised, instruction-level disassembly of a compiled
bytecode file (.101) or an English source file (.abc), without running or
transpiling it.

For .abc source files the program is parsed first and the AST is rendered as
if it had been compiled; no bytecode file is written to disk.

For .101 bytecode files the binary is decoded directly.

The output mirrors the style of Python's dis module: each statement becomes one
or more labelled opcodes with colourised operands.  Line numbers always appear
on the left-hand side; indentation shifts the opcode column to the right for
nested constructs.

By default comparison and logical operators are shown as conventional symbols
(==, !=, <, <=, >, >=, &&, ||, !).  Pass --friendly to display them as the
original English prose stored in the bytecode instead
(e.g. "is less than or equal to").  Arithmetic operators (+, -, *, /, %)
are always shown symbolically regardless of this flag.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		filename := args[0]
		ext := strings.ToLower(filepath.Ext(filename))
		useColor := stacktraces.HasColor()

		if ext == ".101" {
			data, err := os.ReadFile(filename)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
				os.Exit(1)
			}
			dec := bytecode.NewDecoder(data)
			program, err := dec.Decode()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error decoding bytecode: %v\n", err)
				os.Exit(1)
			}
			fmt.Print(disasm.Disassemble(program, filename, useColor, inspectFriendly))
			return
		}

		// .abc or any other extension – parse the source and disassemble the AST
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
		fmt.Print(disasm.Disassemble(program, filename, useColor, inspectFriendly))
	},
}

func init() {
	inspectCmd.Flags().BoolVar(&inspectFriendly, "friendly", false,
		"Show operators as English prose instead of symbols")
	rootCmd.AddCommand(inspectCmd)
}
