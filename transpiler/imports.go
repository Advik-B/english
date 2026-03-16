package transpiler

import (
	"github.com/Advik-B/english/ast"
	"github.com/Advik-B/english/parser"
	"os"
)

// inlineImports walks the program's statement list and replaces every
// ImportStatement with the actual statements from the referenced file,
// making the transpiled Python output self-contained.
//
// Import semantics mirror the English evaluator:
//   - Default (ImportAll): inline all statements from the imported file.
//   - IsSafe: inline only function/variable declarations (skip output/call
//     statements), so top-level side-effects in the library are not executed.
//   - Selective (Items non-empty): inline only the named declarations.
//
// 'seen' tracks which file paths have already been inlined; a second import
// of the same file is silently skipped to avoid duplicate definitions.
//
// If the referenced file cannot be read or parsed, the ImportStatement is kept
// in place so that transpileImport() can emit it as an informational comment.
func inlineImports(program *ast.Program, seen map[string]bool) *ast.Program {
	var newStmts []ast.Statement
	for _, stmt := range program.Statements {
		imp, ok := stmt.(*ast.ImportStatement)
		if !ok {
			newStmts = append(newStmts, stmt)
			continue
		}

		// Second import of the same file → skip (no duplicate definitions).
		if seen[imp.Path] {
			continue
		}

		// Try to read and parse the referenced file.
		content, err := os.ReadFile(imp.Path)
		if err != nil {
			// File not found or unreadable; keep the ImportStatement so that
			// transpileImport() can emit it as a comment.
			newStmts = append(newStmts, stmt)
			continue
		}

		seen[imp.Path] = true

		lx := parser.NewLexer(string(content))
		tokens := lx.TokenizeAll()
		p := parser.NewParser(tokens)
		importedProg, err := p.Parse()
		if err != nil {
			// Parse error in the imported file; keep ImportStatement as comment.
			newStmts = append(newStmts, stmt)
			continue
		}

		// Recursively resolve imports in the imported file.
		importedProg = inlineImports(importedProg, seen)

		// Select which statements to inline based on import mode.
		var toInline []ast.Statement
		switch {
		case len(imp.Items) > 0:
			toInline = selectNamedDecls(importedProg.Statements, imp.Items)
		case imp.IsSafe:
			toInline = filterDecls(importedProg.Statements)
		default:
			toInline = importedProg.Statements
		}

		newStmts = append(newStmts, toInline...)
	}
	return &ast.Program{Statements: newStmts}
}

// filterDecls retains only function/variable/struct declarations, discarding
// top-level statements with side effects (Print, Call, etc.).
// Used for safe imports ("Import from").
func filterDecls(stmts []ast.Statement) []ast.Statement {
	var result []ast.Statement
	for _, s := range stmts {
		switch s.(type) {
		case *ast.FunctionDecl, *ast.VariableDecl, *ast.TypedVariableDecl,
			*ast.StructDecl, *ast.ErrorTypeDecl, *ast.CommentStatement:
			result = append(result, s)
		}
	}
	return result
}

// selectNamedDecls retains only the declarations whose names appear in 'items'.
// Used for selective imports ("Import X and Y from").
func selectNamedDecls(stmts []ast.Statement, items []string) []ast.Statement {
	want := make(map[string]bool, len(items))
	for _, item := range items {
		want[item] = true
	}
	var result []ast.Statement
	for _, s := range stmts {
		switch decl := s.(type) {
		case *ast.FunctionDecl:
			if want[decl.Name] {
				result = append(result, s)
			}
		case *ast.VariableDecl:
			if want[decl.Name] {
				result = append(result, s)
			}
		case *ast.TypedVariableDecl:
			if want[decl.Name] {
				result = append(result, s)
			}
		case *ast.StructDecl:
			if want[decl.Name] {
				result = append(result, s)
			}
		}
	}
	return result
}
