package transpiler

import (
	"english/ast"
	"fmt"
	"strings"
)

// ─── Expressions ─────────────────────────────────────────────────────────────

// transpileExpr converts any AST expression node to a Python expression string.
func (t *Transpiler) transpileExpr(expr ast.Expression) string {
	if expr == nil {
		return "None"
	}
	switch e := expr.(type) {
	case *ast.NumberLiteral:
		return formatNumber(e.Value)
	case *ast.StringLiteral:
		return fmt.Sprintf("%q", e.Value)
	case *ast.BooleanLiteral:
		if e.Value {
			return "True"
		}
		return "False"
	case *ast.NothingLiteral:
		return "None"
	case *ast.Identifier:
		// Inside a struct method body, bare field names become self.<field>.
		if t.methodFields[e.Name] {
			return "self." + e.Name
		}
		// Well-known math constants are injected as env variables by the stdlib.
		// Map them to their Python equivalents.
		if pyConst, ok := mathConstantMap[e.Name]; ok {
			t.needsMath = true
			return pyConst
		}
		return e.Name
	case *ast.ListLiteral:
		return t.transpileListLit(e.Elements)
	case *ast.ArrayLiteral:
		return t.transpileListLit(e.Elements)
	case *ast.LookupTableLiteral:
		return "{}"
	case *ast.BinaryExpression:
		return t.transpileBinaryExpr(e)
	case *ast.UnaryExpression:
		return t.transpileUnaryExpr(e)
	case *ast.FunctionCall:
		return t.transpileFuncCallExpr(e)
	case *ast.MethodCall:
		return t.transpileMethodCallExpr(e)
	case *ast.IndexExpression:
		list := t.transpileExpr(e.List)
		idx := t.transpileExpr(e.Index)
		return fmt.Sprintf("%s[%s]", list, maybeInt(idx))
	case *ast.LengthExpression:
		return fmt.Sprintf("len(%s)", t.transpileExpr(e.List))
	case *ast.FieldAccess:
		return fmt.Sprintf("%s.%s", t.transpileExpr(e.Object), e.Field)
	case *ast.StructInstantiation:
		return t.transpileStructInst(e)
	case *ast.TypeExpression:
		return fmt.Sprintf("type(%s).__name__", t.transpileExpr(e.Value))
	case *ast.CastExpression:
		return t.transpileCast(e)
	case *ast.ReferenceExpression:
		// References are plain variable accesses in Python.
		return e.Name
	case *ast.CopyExpression:
		return fmt.Sprintf("copy.copy(%s)", t.transpileExpr(e.Value))
	case *ast.LocationExpression:
		return fmt.Sprintf("hex(id(%s))", e.Name)
	case *ast.AskExpression:
		if e.Prompt == nil {
			return "input()"
		}
		return fmt.Sprintf("input(%s)", t.transpileExpr(e.Prompt))
	case *ast.LookupKeyAccess:
		return fmt.Sprintf("%s[%s]", t.transpileExpr(e.Table), t.transpileExpr(e.Key))
	case *ast.HasExpression:
		return fmt.Sprintf("%s in %s", t.transpileExpr(e.Key), t.transpileExpr(e.Table))
	case *ast.NilCheckExpression:
		inner := t.transpileExpr(e.Value)
		if e.IsSomethingCheck {
			return fmt.Sprintf("%s is not None", inner)
		}
		return fmt.Sprintf("%s is None", inner)
	case *ast.ErrorTypeCheckExpression:
		return fmt.Sprintf("isinstance(%s, %s)", t.transpileExpr(e.Value), e.TypeName)
	default:
		return fmt.Sprintf("None  # unsupported expression: %T", expr)
	}
}

func (t *Transpiler) transpileListLit(elements []ast.Expression) string {
	parts := make([]string, len(elements))
	for i, el := range elements {
		parts[i] = t.transpileExpr(el)
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

func (t *Transpiler) transpileBinaryExpr(e *ast.BinaryExpression) string {
	left := t.transpileExpr(e.Left)
	right := t.transpileExpr(e.Right)
	op := mapOperator(e.Operator)

	// Wrap nested binary sub-expressions in parentheses to make precedence
	// unambiguous in the generated Python.
	if _, ok := e.Left.(*ast.BinaryExpression); ok {
		left = "(" + left + ")"
	}
	if _, ok := e.Right.(*ast.BinaryExpression); ok {
		right = "(" + right + ")"
	}
	return fmt.Sprintf("%s %s %s", left, op, right)
}

func (t *Transpiler) transpileUnaryExpr(e *ast.UnaryExpression) string {
	right := t.transpileExpr(e.Right)
	switch e.Operator {
	case "-":
		return fmt.Sprintf("-%s", right)
	case "not":
		return fmt.Sprintf("not %s", right)
	default:
		return fmt.Sprintf("%s%s", e.Operator, right)
	}
}

func (t *Transpiler) transpileMethodCallExpr(e *ast.MethodCall) string {
	obj := t.transpileExpr(e.Object)
	args := make([]string, len(e.Arguments))
	for i, a := range e.Arguments {
		args[i] = t.transpileExpr(a)
	}
	return fmt.Sprintf("%s.%s(%s)", obj, e.MethodName, strings.Join(args, ", "))
}

func (t *Transpiler) transpileCast(e *ast.CastExpression) string {
	inner := t.transpileExpr(e.Value)
	switch strings.ToLower(e.TypeName) {
	case "number", "float":
		return fmt.Sprintf("float(%s)", inner)
	case "integer", "int":
		return fmt.Sprintf("int(%s)", inner)
	case "text", "string", "str":
		return fmt.Sprintf("str(%s)", inner)
	case "boolean", "bool":
		return fmt.Sprintf("bool(%s)", inner)
	default:
		// Treat any other cast as a constructor / type call.
		return fmt.Sprintf("%s(%s)", e.TypeName, inner)
	}
}

func (t *Transpiler) transpileStructInst(e *ast.StructInstantiation) string {
	args := make([]string, 0, len(e.FieldOrder))
	for _, name := range e.FieldOrder {
		args = append(args, fmt.Sprintf("%s=%s", name, t.transpileExpr(e.FieldValues[name])))
	}
	return fmt.Sprintf("%s(%s)", e.StructName, strings.Join(args, ", "))
}
