# English VS Code Extension

This extension provides:
- Syntax highlighting for `.abc` files (runtime-colored to match `english cat` palette)
- Language server features (autocomplete, hover, go-to-definition, diagnostics)

## Requirements

- Bun (for extension development workflows)
- Go (required only when auto-building the compiler, e.g. when `english` is not already available)

If `english.languageServer.path` is unavailable, the extension downloads the latest source archive from GitHub, builds `english` locally with Go, and then starts the LSP from that compiled binary.

## Settings

- `english.languageServer.path` (default: `english`)
- `english.languageServer.args` (default: `["lsp"]`)

## Development

```bash
cd vscode-extension
bun install
bun run compile
```

## Test with real VS Code

```bash
cd vscode-extension
bun test
```

`bun test` compiles the extension, builds the local Go `english` binary, downloads/runs a real VS Code instance via `@vscode/test-electron`, and executes the integration suite.

## Build VSIX

```bash
cd vscode-extension
bun run package:vsix
```

This generates `english-language.vsix` in `vscode-extension/`.
