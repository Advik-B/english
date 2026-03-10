package cmd

import (
	"english/highlight"
	"english/stacktraces"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var catCmd = &cobra.Command{
	Use:     "cat [file]",
	Aliases: []string{"display", "highlight"},
	Short:   "Display a source file with syntax highlighting",
	Long: `Display the contents of an English source file (.abc) with full syntax
highlighting.

Keywords, string literals, numbers, comments and operators are each
coloured distinctly.  If the file contains a syntax error the highlighting
continues for as long as possible; any portion that cannot be recognised is
displayed in a dim grey so that the valid code is still clearly readable.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		filename := args[0]
		content, err := os.ReadFile(filename)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
			os.Exit(1)
		}
		useColor := stacktraces.HasColor()
		fmt.Print(highlight.Highlight(string(content), useColor))
	},
}

func init() {
	rootCmd.AddCommand(catCmd)
}
