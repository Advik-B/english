package cmd

import (
	"english/bytecode"
	"english/highlight"
	"english/stacktraces"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var catCmd = &cobra.Command{
	Use:     "cat [file]",
	Aliases: []string{"display", "highlight"},
	Short:   "Display a source file with syntax highlighting, or disassemble a .101 bytecode file",
	Long: `Display the contents of an English source file (.abc) with full syntax
highlighting, or – when given a compiled bytecode file (.101) – print a
colourised, instruction-level disassembly without running or transpiling it.

Source files (.abc):
  Keywords, string literals, numbers, comments and operators are each
  coloured distinctly.  If the file contains a syntax error the highlighting
  continues for as long as possible; any portion that cannot be recognised is
  displayed in a dim grey so that the valid code is still clearly readable.

Bytecode files (.101):
  Decoded instructions are printed in the style of Python's dis module:
  each statement becomes one or more labelled opcodes with colourised
  operands.  The output never contains English prose – only raw instruction
  names and their arguments.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		filename := args[0]
		ext := strings.ToLower(filepath.Ext(filename))

		if ext == ".101" {
			catBytecode(filename)
			return
		}

		content, err := os.ReadFile(filename)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
			os.Exit(1)
		}
		useColor := stacktraces.HasColor()
		fmt.Print(highlight.Highlight(string(content), useColor))
	},
}

// catBytecode decodes a .101 file and prints a colourised disassembly.
func catBytecode(filename string) {
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

	useColor := stacktraces.HasColor()
	fmt.Print(bytecode.Disassemble(program, filename, useColor))
}

func init() {
	rootCmd.AddCommand(catCmd)
}
