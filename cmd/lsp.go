package cmd

import (
	"english/lsp"
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
)

var lspCmd = &cobra.Command{
	Use:   "lsp",
	Short: "Start the Language Server Protocol server",
	Long: `Start the Language Server Protocol (LSP) server for the English programming language.

The LSP server provides IDE features like:
  - Auto-completion for keywords, variables, and functions
  - Hover information showing documentation and types
  - Go-to-definition for variables and functions
  - Find all references
  - Document symbols outline
  - Signature help for function calls
  - Code diagnostics (syntax errors)
  - Code actions and quick fixes
  - Document formatting
  - Folding ranges

The server communicates over stdin/stdout using the LSP protocol.`,
	Run: func(cmd *cobra.Command, args []string) {
		runLSPServer()
	},
}

func init() {
	rootCmd.AddCommand(lspCmd)
}

func runLSPServer() {
	// Create a logger that writes to stderr (stdout is for LSP communication)
	logger := log.New(os.Stderr, "[English LSP] ", log.LstdFlags)

	server := lsp.NewServer(
		lsp.WithLogger(logger),
		lsp.WithInitializeCallback(func(params *lsp.InitializeParams) error {
			if params.ClientInfo != nil {
				logger.Printf("Connected to %s %s", params.ClientInfo.Name, params.ClientInfo.Version)
			}
			return nil
		}),
		lsp.WithShutdownCallback(func() error {
			logger.Println("Server shutting down...")
			return nil
		}),
	)

	logger.Println("Starting English Language Server...")
	if err := server.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "LSP server error: %v\n", err)
		os.Exit(1)
	}
}
