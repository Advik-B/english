// Package transpiler converts an English language AST to human-readable Python.
//
// The package is organised across several files:
//
//   - transpiler.go  – Transpiler struct, Transpile(), scanning pass
//   - statements.go  – statement transpilation
//   - expressions.go – expression transpilation
//   - stdlib.go      – complete stdlib function name mapping
//   - helpers.go     – Python helper defs, formatNumber, mapOperator, mapTypeName
package transpiler

import (
	"github.com/Advik-B/english/ast"
	"strings"
)

// Transpiler converts an English AST to Python source code.
type Transpiler struct {
	indent int
	buf    strings.Builder

	// keepComments controls whether source comments are carried through to the
	// generated Python output.
	//
	// true  (default, for .abc source files)  – CommentStatement nodes become
	//       Python # comments, and the file banner is included.
	// false (for .101 bytecode files)          – all comments are suppressed;
	//       the generated Python contains no comment lines at all.
	keepComments bool

	// inlineMode controls how Import statements are handled.
	//
	// true  (--inline flag) – every ImportStatement is resolved by reading and
	//       inlining the referenced .abc file's AST directly. The output is a
	//       single self-contained Python file.
	// false (default)       – Import statements are left in the AST and
	//       transpileImport() emits Python importlib code that loads the
	//       corresponding .py file produced alongside the main file.
	inlineMode bool

	// sourceDir is the directory of the main source file being transpiled. It
	// is used (in non-inline mode) to compute the path from the main .py file
	// to each sibling library .py file in the generated importlib statements.
	sourceDir string

	// userFunctions is the set of user-defined function names collected during
	// the scan pass. Functions in this set take priority over any stdlib mapping
	// with the same name, so "Declare function average ..." is emitted as a plain
	// call rather than being mis-translated to the stdlib average expression.
	userFunctions map[string]bool

	// Python module imports required by the generated code.
	needsMath   bool
	needsCopy   bool
	needsRandom bool
	needsTyping bool // typing.Final for constants
	needsTime   bool

	// Python helper functions to inject at the top of the output.
	helpers map[string]bool

	// methodFields holds the field names of the struct currently being transpiled.
	// Bare identifier references to field names inside method bodies are
	// rewritten to self.<field>.
	methodFields map[string]bool
}

// NewTranspiler creates a Transpiler for .abc source files.
// Source comments are preserved in the output.
// Import statements produce sibling .py files and Python importlib calls (the
// default; use NewTranspilerInlined() or the --inline flag to instead inline
// all imported code into the single output file).
func NewTranspiler() *Transpiler {
	return &Transpiler{
		helpers:       make(map[string]bool),
		userFunctions: make(map[string]bool),
		keepComments:  true,
		inlineMode:    false,
	}
}

// NewTranspilerInlined creates a Transpiler that resolves all Import statements
// by reading and inlining the referenced .abc files into the single output file.
// This is the behaviour activated by the --inline flag.
func NewTranspilerInlined() *Transpiler {
	return &Transpiler{
		helpers:       make(map[string]bool),
		userFunctions: make(map[string]bool),
		keepComments:  true,
		inlineMode:    true,
	}
}

// NewTranspilerStripped creates a Transpiler that suppresses all comment lines
// in the generated Python output. Use this when transpiling .101 bytecode files,
// which contain no source comments.
func NewTranspilerStripped() *Transpiler {
	return &Transpiler{
		helpers:       make(map[string]bool),
		userFunctions: make(map[string]bool),
		keepComments:  false,
		inlineMode:    false,
	}
}

// WithSourceDir sets the directory of the main source file so that the
// transpiler can compute correct relative paths in Python importlib statements
// when resolving imports in non-inline mode.
func (t *Transpiler) WithSourceDir(dir string) *Transpiler {
	t.sourceDir = dir
	return t
}

// Transpile converts a parsed English program to a Python source string.
//
// It runs two passes:
//  1. Scan the AST to discover which Python imports and helper functions are
//     needed.
//  2. Generate the body of the Python file.
//
// The final output is assembled as:
//
// banner + imports + helpers + body
func (t *Transpiler) Transpile(program *ast.Program) string {
	// Pass 0 (inline mode only) – resolve imports by inlining the referenced
	// files so that the generated Python is self-contained. In non-inline mode
	// ImportStatements are kept in the AST and transpileImport() emits Python
	// importlib code that loads the sibling .py files produced by the CLI.
	if t.inlineMode {
		program = inlineImports(program, make(map[string]bool))
	}

	// Pass 1 – collect import and helper requirements.
	t.scanProgram(program)

	// Pass 2 – generate the body.
	var bodyBuf strings.Builder
	savedBuf := t.buf
	t.buf = bodyBuf
	t.transpileBody(program.Statements) // blank-line formatting applied here
	body := t.buf.String()
	t.buf = savedBuf

	// Assemble the final file.
	var out strings.Builder

	// The banner and any generated comment lines are only emitted when comments
	// are enabled (i.e. for .abc source files). .101 bytecode files produce
	// comment-free Python.
	if t.keepComments {
		out.WriteString("# Transpiled from English language source\n")
	}

	if t.needsMath {
		out.WriteString("import math\n")
	}
	if t.needsCopy {
		out.WriteString("import copy\n")
	}
	if t.needsRandom {
		out.WriteString("import random\n")
	}
	if t.needsTime {
		out.WriteString("import time\n")
	}
	if t.needsTyping {
		out.WriteString("from typing import Final\n")
	}
	if t.needsMath || t.needsCopy || t.needsRandom || t.needsTime || t.needsTyping {
		out.WriteString("\n")
	}

	// Emit helper functions in a deterministic order.
	// Each helper is followed by two newlines (\n\n) so the body that follows
	// is separated by one blank line.
	for _, name := range helperOrder {
		if t.helpers[name] {
			out.WriteString(helperDefs[name])
			out.WriteString("\n\n")
		}
	}

	out.WriteString(body)
	return out.String()
}

// ─── Scanning pass ────────────────────────────────────────────────────────────
//
// The scan pass walks the entire AST before code generation to determine which
// Python imports and helper functions the generated code needs.

func (t *Transpiler) scanProgram(program *ast.Program) {
	for _, stmt := range program.Statements {
		t.scanStmt(stmt)
	}
}

func (t *Transpiler) scanStmt(stmt ast.Statement) {
	switch s := stmt.(type) {
	case *ast.FunctionDecl:
		// Register this as a user-defined function so that transpileFuncCallExpr
		// prefers it over any stdlib mapping with the same name.
		t.userFunctions[s.Name] = true
		for _, c := range s.Body {
			t.scanStmt(c)
		}
	case *ast.IfStatement:
		for _, c := range s.Then {
			t.scanStmt(c)
		}
		for _, elif := range s.ElseIf {
			for _, c := range elif.Body {
				t.scanStmt(c)
			}
		}
		for _, c := range s.Else {
			t.scanStmt(c)
		}
	case *ast.WhileLoop:
		for _, c := range s.Body {
			t.scanStmt(c)
		}
	case *ast.ForLoop:
		for _, c := range s.Body {
			t.scanStmt(c)
		}
	case *ast.ForEachLoop:
		for _, c := range s.Body {
			t.scanStmt(c)
		}
	case *ast.TryStatement:
		for _, c := range s.TryBody {
			t.scanStmt(c)
		}
		for _, c := range s.ErrorBody {
			t.scanStmt(c)
		}
		for _, c := range s.FinallyBody {
			t.scanStmt(c)
		}
	case *ast.StructDecl:
		for _, m := range s.Methods {
			for _, c := range m.Body {
				t.scanStmt(c)
			}
		}
	case *ast.VariableDecl:
		t.scanExpr(s.Value)
		if s.IsConstant {
			t.needsTyping = true
		}
	case *ast.TypedVariableDecl:
		t.scanExpr(s.Value)
		if s.IsConstant {
			t.needsTyping = true
		}
	case *ast.Assignment:
		t.scanExpr(s.Value)
	case *ast.IndexAssignment:
		t.scanExpr(s.Index)
		t.scanExpr(s.Value)
	case *ast.FieldAssignment:
		t.scanExpr(s.Value)
	case *ast.LookupKeyAssignment:
		t.scanExpr(s.Key)
		t.scanExpr(s.Value)
	case *ast.OutputStatement:
		for _, v := range s.Values {
			t.scanExpr(v)
		}
	case *ast.ReturnStatement:
		t.scanExpr(s.Value)
	case *ast.CallStatement:
		if s.FunctionCall != nil {
			t.scanFuncCall(s.FunctionCall.Name)
			for _, a := range s.FunctionCall.Arguments {
				t.scanExpr(a)
			}
		}
		if s.MethodCall != nil {
			t.scanExpr(s.MethodCall.Object)
		}
	case *ast.RaiseStatement:
		t.scanExpr(s.Message)
	}
}

func (t *Transpiler) scanExpr(expr ast.Expression) {
	if expr == nil {
		return
	}
	switch e := expr.(type) {
	case *ast.FunctionCall:
		t.scanFuncCall(e.Name)
		for _, a := range e.Arguments {
			t.scanExpr(a)
		}
	case *ast.MethodCall:
		t.scanExpr(e.Object)
		for _, a := range e.Arguments {
			t.scanExpr(a)
		}
	case *ast.BinaryExpression:
		t.scanExpr(e.Left)
		t.scanExpr(e.Right)
	case *ast.UnaryExpression:
		t.scanExpr(e.Right)
	case *ast.CopyExpression:
		t.needsCopy = true
		t.scanExpr(e.Value)
	case *ast.IndexExpression:
		t.scanExpr(e.List)
		t.scanExpr(e.Index)
	case *ast.LengthExpression:
		t.scanExpr(e.List)
	case *ast.FieldAccess:
		t.scanExpr(e.Object)
	case *ast.LookupKeyAccess:
		t.scanExpr(e.Table)
		t.scanExpr(e.Key)
	case *ast.HasExpression:
		t.scanExpr(e.Table)
		t.scanExpr(e.Key)
	case *ast.NilCheckExpression:
		t.scanExpr(e.Value)
	case *ast.CastExpression:
		t.scanExpr(e.Value)
	case *ast.TypeExpression:
		t.scanExpr(e.Value)
	case *ast.AskExpression:
		t.scanExpr(e.Prompt)
	case *ast.StructInstantiation:
		for _, v := range e.FieldValues {
			t.scanExpr(v)
		}
	case *ast.ErrorTypeCheckExpression:
		t.scanExpr(e.Value)
	case *ast.ListLiteral:
		for _, el := range e.Elements {
			t.scanExpr(el)
		}
	case *ast.RangeLiteral:
		t.scanExpr(e.Start)
		t.scanExpr(e.End)
	case *ast.ArrayLiteral:
		for _, el := range e.Elements {
			t.scanExpr(el)
		}
	}
}

// scanFuncCall marks the Python modules and helper functions required by the
// given English stdlib function name.
func (t *Transpiler) scanFuncCall(name string) {
	switch name {
	case "sqrt", "pow", "floor", "ceil", "sin", "cos", "tan", "log", "log10", "log2", "exp":
		t.needsMath = true
	case "is_nan":
		t.needsMath = true
		t.helpers["_is_nan"] = true
	case "is_infinite":
		t.needsMath = true
		t.helpers["_is_infinite"] = true
	case "random", "random_between":
		t.needsRandom = true
	case "table_remove":
		t.helpers["_table_remove"] = true
	case "flatten":
		t.helpers["_flatten"] = true
	case "product":
		t.helpers["_product"] = true
	case "unique":
		t.helpers["_unique"] = true
	case "zip_with":
		t.helpers["_zip_with"] = true
	case "sign":
		t.helpers["_sign"] = true
	case "read_file":
		t.helpers["_read_file"] = true
	case "write_file":
		t.helpers["_write_file"] = true
	case "sleep", "current_time", "elapsed_time":
		t.needsTime = true
		if name == "elapsed_time" {
			t.helpers["_program_start"] = true
		}
	}
}

// ensureBlankLines writes newlines to the current buffer until there are at
// least n blank lines (i.e. n+1 trailing newline characters) before the next
// write position. It is a no-op when the buffer is empty (start of file).
func (t *Transpiler) ensureBlankLines(n int) {
	s := t.buf.String()
	if s == "" {
		return
	}
	// Count how many '\n' characters trail the current buffer content.
	trailing := 0
	for i := len(s) - 1; i >= 0 && s[i] == '\n'; i-- {
		trailing++
	}
	// n blank lines require n+1 total trailing newlines.
	for need := (n + 1) - trailing; need > 0; need-- {
		t.buf.WriteByte('\n')
	}
}

func (t *Transpiler) write(s string) {
	t.buf.WriteString(s)
}

func (t *Transpiler) writeLine(s string) {
	t.buf.WriteString(strings.Repeat("    ", t.indent))
	t.buf.WriteString(s)
	t.buf.WriteString("\n")
}
