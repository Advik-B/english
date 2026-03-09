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
	"english/ast"
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

	// Python module imports required by the generated code.
	needsMath   bool
	needsCopy   bool
	needsRandom bool
	needsTyping bool // typing.Final for constants

	// Python helper functions to inject at the top of the output.
	helpers map[string]bool

	// methodFields holds the field names of the struct currently being transpiled.
	// Bare identifier references to field names inside method bodies are
	// rewritten to self.<field>.
	methodFields map[string]bool
}

// NewTranspiler creates a Transpiler that preserves source comments in the
// generated Python output. Use this when transpiling .abc source files.
func NewTranspiler() *Transpiler {
	return &Transpiler{
		helpers:      make(map[string]bool),
		keepComments: true,
	}
}

// NewTranspilerStripped creates a Transpiler that suppresses all comment lines
// in the generated Python output. Use this when transpiling .101 bytecode files,
// which contain no source comments.
func NewTranspilerStripped() *Transpiler {
	return &Transpiler{
		helpers:      make(map[string]bool),
		keepComments: false,
	}
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
	// Pass 1 – collect import and helper requirements.
	t.scanProgram(program)

	// Pass 2 – generate the body.
	var bodyBuf strings.Builder
	savedBuf := t.buf
	t.buf = bodyBuf
	for _, stmt := range program.Statements {
		t.transpileStatement(stmt)
	}
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
	if t.needsTyping {
		out.WriteString("from typing import Final\n")
	}
	if t.needsMath || t.needsCopy || t.needsRandom || t.needsTyping {
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
	}
}

// ─── Low-level output helpers ─────────────────────────────────────────────────

func (t *Transpiler) write(s string) {
	t.buf.WriteString(s)
}

func (t *Transpiler) writeLine(s string) {
	t.buf.WriteString(strings.Repeat("    ", t.indent))
	t.buf.WriteString(s)
	t.buf.WriteString("\n")
}
