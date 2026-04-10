# English VS Code Extension

This extension provides:
- Syntax highlighting for `.abc` files
- Language server features (autocomplete, hover, go-to-definition, diagnostics)

## Requirements

The `english` binary must be available on `PATH` (or configured explicitly).

## Settings

- `english.languageServer.path` (default: `english`)
- `english.languageServer.args` (default: `["lsp"]`)

## Development

```bash
cd vscode-extension
npm install
npm run compile
```

## Test with real VS Code

```bash
cd vscode-extension
npm test
```

`npm test` compiles the extension, builds the local Go `english` binary, downloads/runs a real VS Code instance via `@vscode/test-electron`, and executes the integration suite.

## Build VSIX

```bash
cd vscode-extension
npm run package:vsix
```

This generates `english-language.vsix` in `vscode-extension/`.
