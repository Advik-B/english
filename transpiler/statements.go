package transpiler

import (
	"english/ast"
	"fmt"
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
		t.writeLine(fmt.Sprintf("%s = not %s", s.Name, s.Name))
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
		t.writeLine(fmt.Sprintf("%s, %s = %s, %s", s.Name1, s.Name2, s.Name2, s.Name1))
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
	if len(s.Items) > 0 {
		t.writeLine(fmt.Sprintf("# from %q import %s", s.Path, strings.Join(s.Items, ", ")))
	} else {
		t.writeLine(fmt.Sprintf("# import %q", s.Path))
	}
}

func (t *Transpiler) transpileVariableDecl(s *ast.VariableDecl) {
	val := t.transpileExpr(s.Value)
	if s.IsConstant {
		// Emit a typing.Final annotation so type-checkers treat this as constant.
		t.writeLine(fmt.Sprintf("%s: Final = %s", s.Name, val))
	} else {
		t.writeLine(fmt.Sprintf("%s = %s", s.Name, val))
	}
}

func (t *Transpiler) transpileTypedVariableDecl(s *ast.TypedVariableDecl) {
	val := t.transpileExpr(s.Value)
	typeName := mapTypeName(s.TypeName)
	if s.IsConstant {
		t.writeLine(fmt.Sprintf("%s: Final[%s] = %s", s.Name, typeName, val))
	} else {
		t.writeLine(fmt.Sprintf("%s: %s = %s", s.Name, typeName, val))
	}
}

func (t *Transpiler) transpileAssignment(s *ast.Assignment) {
	target := s.Name
	if t.methodFields[target] {
		target = "self." + target
	}
	t.writeLine(fmt.Sprintf("%s = %s", target, t.transpileExpr(s.Value)))
}

func (t *Transpiler) transpileIndexAssignment(s *ast.IndexAssignment) {
	target := s.ListName
	if t.methodFields[target] {
		target = "self." + target
	}
	idx := t.transpileExpr(s.Index)
	val := t.transpileExpr(s.Value)
	t.writeLine(fmt.Sprintf("%s[%s] = %s", target, maybeInt(idx), val))
}

func (t *Transpiler) transpileFieldAssignment(s *ast.FieldAssignment) {
	t.writeLine(fmt.Sprintf("%s.%s = %s", s.ObjectName, s.Field, t.transpileExpr(s.Value)))
}

func (t *Transpiler) transpileLookupKeyAssignment(s *ast.LookupKeyAssignment) {
	key := t.transpileExpr(s.Key)
	val := t.transpileExpr(s.Value)
	t.writeLine(fmt.Sprintf("%s[%s] = %s", s.TableName, key, val))
}

func (t *Transpiler) transpileFunctionDecl(s *ast.FunctionDecl) {
	params := strings.Join(s.Parameters, ", ")
	t.writeLine(fmt.Sprintf("def %s(%s):", s.Name, params))
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
	t.writeLine(fmt.Sprintf("for %s in %s:", s.Item, t.transpileExpr(s.List)))
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
		// Build __init__ with all fields as parameters (with defaults if present).
		params := make([]string, 0, len(s.Fields)+1)
		params = append(params, "self")
		for _, field := range s.Fields {
			if field.DefaultValue != nil {
				defVal := t.transpileExpr(field.DefaultValue)
				params = append(params, fmt.Sprintf("%s=%s", field.Name, defVal))
			} else {
				params = append(params, field.Name)
			}
		}
		t.writeLine(fmt.Sprintf("def __init__(%s):", strings.Join(params, ", ")))
		t.indent++
		for _, field := range s.Fields {
			t.writeLine(fmt.Sprintf("self.%s = %s", field.Name, field.Name))
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
		params := make([]string, 0, len(method.Parameters)+1)
		params = append(params, "self")
		params = append(params, method.Parameters...)
		t.writeLine(fmt.Sprintf("def %s(%s):", method.Name, strings.Join(params, ", ")))
		t.indent++
		t.transpileBody(method.Body)
		t.indent--
	}

	t.indent--
	t.write("\n")
}

// transpileBody writes a block of statements.
// Emits a single "pass" when the body is empty (required by Python syntax).
func (t *Transpiler) transpileBody(stmts []ast.Statement) {
	if len(stmts) == 0 {
		t.writeLine("pass")
		return
	}
	for _, stmt := range stmts {
		t.transpileStatement(stmt)
	}
}
