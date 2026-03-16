package transpiler

import (
	"github.com/Advik-B/english/ast"
	"fmt"
	"path/filepath"
	"strings"
)

// ─── Statements ───────────────────────────────────────────────────────────────

// transpileStatement dispatches a single AST statement to the appropriate
// concrete translation method.
func (t *Transpiler) transpileStatement(stmt ast.Statement) {
	switch s := stmt.(type) {
	case *ast.ImportStatement:
		t.transpileImport(s)
	case *ast.VariableDecl:
		t.transpileVariableDecl(s)
	case *ast.TypedVariableDecl:
		t.transpileTypedVariableDecl(s)
	case *ast.Assignment:
		t.transpileAssignment(s)
	case *ast.IndexAssignment:
		t.transpileIndexAssignment(s)
	case *ast.FieldAssignment:
		t.transpileFieldAssignment(s)
	case *ast.LookupKeyAssignment:
		t.transpileLookupKeyAssignment(s)
	case *ast.FunctionDecl:
		t.transpileFunctionDecl(s)
	case *ast.CallStatement:
		t.transpileCallStatement(s)
	case *ast.OutputStatement:
		t.transpileOutput(s)
	case *ast.ReturnStatement:
		t.transpileReturn(s)
	case *ast.IfStatement:
		t.transpileIf(s)
	case *ast.WhileLoop:
		t.transpileWhile(s)
	case *ast.ForLoop:
		t.transpileForLoop(s)
	case *ast.ForEachLoop:
		t.transpileForEach(s)
	case *ast.ToggleStatement:
		n := sanitizeIdent(s.Name)
		t.writeLine(fmt.Sprintf("%s = not %s", n, n))
	case *ast.BreakStatement:
		t.writeLine("break")
	case *ast.ContinueStatement:
		t.writeLine("continue")
	case *ast.TryStatement:
		t.transpileTry(s)
	case *ast.RaiseStatement:
		t.transpileRaise(s)
	case *ast.ErrorTypeDecl:
		t.transpileErrorTypeDecl(s)
	case *ast.SwapStatement:
		n1, n2 := sanitizeIdent(s.Name1), sanitizeIdent(s.Name2)
		t.writeLine(fmt.Sprintf("%s, %s = %s, %s", n1, n2, n2, n1))
	case *ast.StructDecl:
		t.transpileStructDecl(s)
	case *ast.CommentStatement:
		t.transpileComment(s)
	default:
		t.writeLine(fmt.Sprintf("# unsupported statement: %T", stmt))
	}
}

// ─── Individual statement translators ────────────────────────────────────────

func (t *Transpiler) transpileImport(s *ast.ImportStatement) {
	if t.inlineMode {
		// In inline mode, ImportStatements are replaced before code generation.
		// This code path is only reached if the referenced file couldn't be resolved.
		if !t.keepComments {
			return
		}
		if len(s.Items) > 0 {
			t.writeLine(fmt.Sprintf("# from %q import %s", s.Path, strings.Join(s.Items, ", ")))
		} else {
			t.writeLine(fmt.Sprintf("# import %q", s.Path))
		}
		return
	}

	// Non-inline mode: the CLI has already transpiled the imported .abc file to a
	// sibling .py file with a clean name (e.g. "examples/math_library.py").
	// Emit standard Python import statements so the output stays simple and
	// readable for beginners.

	// Module name: base file name without any extensions (e.g. "math_library").
	base := filepath.Base(s.Path)
	module := strings.SplitN(base, ".", 2)[0]

	// If the library lives in a different directory from the main file, emit a
	// one-line sys.path insert so Python can find the module.
	importedDir := filepath.Dir(s.Path)
	mainDir := t.sourceDir
	if mainDir == "" {
		mainDir = "."
	}
	if rel, err := filepath.Rel(mainDir, importedDir); err == nil {
		rel = filepath.ToSlash(rel)
		if rel != "." {
			t.writeLine("import sys")
			t.writeLine("import os")
			t.writeLine(fmt.Sprintf("sys.path.insert(0, os.path.join(os.path.dirname(__file__), %q))", rel))
		}
	}

	// Emit the Python import.
	if len(s.Items) > 0 {
		// Selective: "from math_library import square, cube"
		sanitized := make([]string, len(s.Items))
		for i, item := range s.Items {
			sanitized[i] = sanitizeIdent(item)
		}
		t.writeLine(fmt.Sprintf("from %s import %s", module, strings.Join(sanitized, ", ")))
	} else {
		// ImportAll / IsSafe: "from math_library import *"
		t.writeLine(fmt.Sprintf("from %s import *", module))
	}
	if t.keepComments {
		t.write("\n")
	}
}

func (t *Transpiler) transpileComment(s *ast.CommentStatement) {
	// Comments are only emitted when keepComments is true (i.e. .abc source files).
	// .101 bytecode files never contain CommentStatement nodes, but this guard
	// provides an explicit defense in depth.
	if !t.keepComments {
		return
	}
	if s.Text == "" {
		t.writeLine("#")
	} else {
		t.writeLine(fmt.Sprintf("# %s", s.Text))
	}
}

func (t *Transpiler) transpileVariableDecl(s *ast.VariableDecl) {
	name := sanitizeIdent(s.Name)
	val := t.transpileExpr(s.Value)
	if s.IsConstant {
		// Emit a typing.Final annotation so type-checkers treat this as constant.
		t.writeLine(fmt.Sprintf("%s: Final = %s", name, val))
	} else {
		t.writeLine(fmt.Sprintf("%s = %s", name, val))
	}
}

func (t *Transpiler) transpileTypedVariableDecl(s *ast.TypedVariableDecl) {
	name := sanitizeIdent(s.Name)
	val := t.transpileExpr(s.Value)
	typeName := mapTypeName(s.TypeName)
	if s.IsConstant {
		t.writeLine(fmt.Sprintf("%s: Final[%s] = %s", name, typeName, val))
	} else {
		t.writeLine(fmt.Sprintf("%s: %s = %s", name, typeName, val))
	}
}

func (t *Transpiler) transpileAssignment(s *ast.Assignment) {
	target := sanitizeIdent(s.Name)
	if t.methodFields[s.Name] {
		target = "self." + sanitizeIdent(s.Name)
	}
	t.writeLine(fmt.Sprintf("%s = %s", target, t.transpileExpr(s.Value)))
}

func (t *Transpiler) transpileIndexAssignment(s *ast.IndexAssignment) {
	target := sanitizeIdent(s.ListName)
	if t.methodFields[s.ListName] {
		target = "self." + sanitizeIdent(s.ListName)
	}
	idx := t.transpileExpr(s.Index)
	val := t.transpileExpr(s.Value)
	t.writeLine(fmt.Sprintf("%s[%s] = %s", target, maybeInt(idx), val))
}

func (t *Transpiler) transpileFieldAssignment(s *ast.FieldAssignment) {
	t.writeLine(fmt.Sprintf("%s.%s = %s", sanitizeIdent(s.ObjectName), s.Field, t.transpileExpr(s.Value)))
}

func (t *Transpiler) transpileLookupKeyAssignment(s *ast.LookupKeyAssignment) {
	key := t.transpileExpr(s.Key)
	val := t.transpileExpr(s.Value)
	t.writeLine(fmt.Sprintf("%s[%s] = %s", sanitizeIdent(s.TableName), key, val))
}

func (t *Transpiler) transpileFunctionDecl(s *ast.FunctionDecl) {
	params := make([]string, len(s.Parameters))
	for i, p := range s.Parameters {
		params[i] = sanitizeIdent(p)
	}
	t.writeLine(fmt.Sprintf("def %s(%s):", sanitizeIdent(s.Name), strings.Join(params, ", ")))
	t.indent++
	t.transpileBody(s.Body)
	t.indent--
	t.write("\n")
}

func (t *Transpiler) transpileCallStatement(s *ast.CallStatement) {
	if s.FunctionCall != nil {
		t.writeLine(t.transpileFuncCallExpr(s.FunctionCall))
	} else if s.MethodCall != nil {
		t.writeLine(t.transpileMethodCallExpr(s.MethodCall))
	}
}

func (t *Transpiler) transpileOutput(s *ast.OutputStatement) {
	parts := make([]string, len(s.Values))
	for i, v := range s.Values {
		parts[i] = t.transpileExpr(v)
	}
	args := strings.Join(parts, ", ")
	if s.Newline {
		t.writeLine(fmt.Sprintf("print(%s)", args))
	} else {
		t.writeLine(fmt.Sprintf("print(%s, end=\"\")", args))
	}
}

func (t *Transpiler) transpileReturn(s *ast.ReturnStatement) {
	if s.Value == nil {
		t.writeLine("return")
	} else {
		t.writeLine(fmt.Sprintf("return %s", t.transpileExpr(s.Value)))
	}
}

func (t *Transpiler) transpileIf(s *ast.IfStatement) {
	t.writeLine(fmt.Sprintf("if %s:", t.transpileExpr(s.Condition)))
	t.indent++
	t.transpileBody(s.Then)
	t.indent--

	for _, elif := range s.ElseIf {
		t.writeLine(fmt.Sprintf("elif %s:", t.transpileExpr(elif.Condition)))
		t.indent++
		t.transpileBody(elif.Body)
		t.indent--
	}

	if len(s.Else) > 0 {
		t.writeLine("else:")
		t.indent++
		t.transpileBody(s.Else)
		t.indent--
	}
}

func (t *Transpiler) transpileWhile(s *ast.WhileLoop) {
	t.writeLine(fmt.Sprintf("while %s:", t.transpileExpr(s.Condition)))
	t.indent++
	t.transpileBody(s.Body)
	t.indent--
}

func (t *Transpiler) transpileForLoop(s *ast.ForLoop) {
	count := t.transpileExpr(s.Count)
	t.writeLine(fmt.Sprintf("for _ in range(%s):", maybeInt(count)))
	t.indent++
	t.transpileBody(s.Body)
	t.indent--
}

func (t *Transpiler) transpileForEach(s *ast.ForEachLoop) {
	t.writeLine(fmt.Sprintf("for %s in %s:", sanitizeIdent(s.Item), t.transpileExpr(s.List)))
	t.indent++
	t.transpileBody(s.Body)
	t.indent--
}

func (t *Transpiler) transpileTry(s *ast.TryStatement) {
	t.writeLine("try:")
	t.indent++
	t.transpileBody(s.TryBody)
	t.indent--

	if len(s.ErrorBody) > 0 {
		var excLine string
		if s.ErrorType != "" {
			if s.ErrorVar != "" {
				excLine = fmt.Sprintf("except %s as %s:", s.ErrorType, s.ErrorVar)
			} else {
				excLine = fmt.Sprintf("except %s:", s.ErrorType)
			}
		} else {
			if s.ErrorVar != "" {
				excLine = fmt.Sprintf("except Exception as %s:", s.ErrorVar)
			} else {
				excLine = "except Exception:"
			}
		}
		t.writeLine(excLine)
		t.indent++
		t.transpileBody(s.ErrorBody)
		t.indent--
	}

	if len(s.FinallyBody) > 0 {
		t.writeLine("finally:")
		t.indent++
		t.transpileBody(s.FinallyBody)
		t.indent--
	}
}

func (t *Transpiler) transpileRaise(s *ast.RaiseStatement) {
	msg := t.transpileExpr(s.Message)
	if s.ErrorType != "" {
		t.writeLine(fmt.Sprintf("raise %s(%s)", s.ErrorType, msg))
	} else {
		t.writeLine(fmt.Sprintf("raise Exception(%s)", msg))
	}
}

func (t *Transpiler) transpileErrorTypeDecl(s *ast.ErrorTypeDecl) {
	parent := "Exception"
	if s.ParentType != "" {
		parent = s.ParentType
	}
	t.writeLine(fmt.Sprintf("class %s(%s): pass", s.Name, parent))
	t.write("\n")
}

func (t *Transpiler) transpileStructDecl(s *ast.StructDecl) {
	t.writeLine(fmt.Sprintf("class %s:", s.Name))
	t.indent++

	if len(s.Fields) > 0 {
		// Build __init__ with all fields as parameters.
		// Fields with an explicit default use that value; fields without a default
		// fall back to the Python zero value for their declared type so that
		// instances can be created with no arguments (e.g. "a new instance of T").
		params := make([]string, 0, len(s.Fields)+1)
		params = append(params, "self")
		for _, field := range s.Fields {
			fname := sanitizeIdent(field.Name)
			if field.DefaultValue != nil {
				defVal := t.transpileExpr(field.DefaultValue)
				params = append(params, fmt.Sprintf("%s=%s", fname, defVal))
			} else {
				params = append(params, fmt.Sprintf("%s=%s", fname, typeZeroValue(field.TypeName)))
			}
		}
		t.writeLine(fmt.Sprintf("def __init__(%s):", strings.Join(params, ", ")))
		t.indent++
		for _, field := range s.Fields {
			fname := sanitizeIdent(field.Name)
			t.writeLine(fmt.Sprintf("self.%s = %s", fname, fname))
		}
		t.indent--
	} else {
		t.writeLine("pass")
	}

	// Activate self.<field> rewriting for method bodies.
	fields := make(map[string]bool, len(s.Fields))
	for _, field := range s.Fields {
		fields[field.Name] = true
	}
	savedFields := t.methodFields
	t.methodFields = fields
	defer func() { t.methodFields = savedFields }()

	for _, method := range s.Methods {
		t.write("\n")
		mparams := make([]string, 0, len(method.Parameters)+1)
		mparams = append(mparams, "self")
		for _, p := range method.Parameters {
			mparams = append(mparams, sanitizeIdent(p))
		}
		t.writeLine(fmt.Sprintf("def %s(%s):", sanitizeIdent(method.Name), strings.Join(mparams, ", ")))
		t.indent++
		t.transpileBody(method.Body)
		t.indent--
	}

	t.indent--
	t.write("\n")
}

// transpileBody writes a block of statements.
//
// When the body is empty it emits "pass" only when called from inside a
// function, method, if-branch, loop, or try-block body — i.e. when
// t.indent > 0. Every such caller increments t.indent before calling
// transpileBody, so this check is always correct. Top-level programs
// (t.indent == 0) do not need "pass" for an empty body.
//
// At the top level (indent == 0) it also enforces PEP 8 E302 — two blank
// lines before every function or class definition. The blank lines are placed
// before any run of indent-0 comments that immediately precedes the
// definition, so comments stay visually attached to the code they describe.
func (t *Transpiler) transpileBody(stmts []ast.Statement) {
	if len(stmts) == 0 {
		if t.indent > 0 {
			t.writeLine("pass")
		}
		return
	}
	for i, stmt := range stmts {
		if t.indent == 0 {
			switch stmt.(type) {
			case *ast.FunctionDecl, *ast.StructDecl, *ast.ErrorTypeDecl:
				// Add two blank lines before this definition unless it is
				// immediately preceded by a comment (in which case the blank
				// lines were already inserted before that comment block).
				if i == 0 || !isCommentStmt(stmts[i-1]) {
					t.ensureBlankLines(2)
				}
			case *ast.CommentStatement:
				// Add two blank lines before the first comment in a
				// contiguous run that leads directly into a definition.
				if isFirstCommentBeforeDef(stmts, i) {
					t.ensureBlankLines(2)
				}
			}
		}
		t.transpileStatement(stmt)
	}
}

// isCommentStmt reports whether s is a *ast.CommentStatement.
func isCommentStmt(s ast.Statement) bool {
	_, ok := s.(*ast.CommentStatement)
	return ok
}

// isFirstCommentBeforeDef reports whether stmts[i] is the first comment in a
// contiguous run of CommentStatements that is immediately followed by a
// top-level FunctionDecl, StructDecl, or ErrorTypeDecl.
func isFirstCommentBeforeDef(stmts []ast.Statement, i int) bool {
	// Must not be preceded by another comment.
	if i > 0 {
		if _, ok := stmts[i-1].(*ast.CommentStatement); ok {
			return false
		}
	}
	// Scan forward through any subsequent comments to find the next
	// non-comment statement.
	for j := i + 1; j < len(stmts); j++ {
		if _, ok := stmts[j].(*ast.CommentStatement); ok {
			continue
		}
		switch stmts[j].(type) {
		case *ast.FunctionDecl, *ast.StructDecl, *ast.ErrorTypeDecl:
			return true
		default:
			return false
		}
	}
	return false
}
