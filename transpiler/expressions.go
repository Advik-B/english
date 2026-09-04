package transpiler

import (
	"fmt"
	"strings"

	"github.com/Advik-B/english/ast"
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
			return "self." + sanitizeIdent(e.Name)
		}
		// Well-known math constants are injected as env variables by the stdlib.
		// Map them to their Python equivalents.
		if pyConst, ok := mathConstantMap[e.Name]; ok {
			t.needsMath = true
			return pyConst
		}
		return sanitizeIdent(e.Name)
	case *ast.ListLiteral:
		return t.transpileListLit(e.Elements)
	case *ast.RangeLiteral:
		return t.transpileRangeLit(e)
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
		return fmt.Sprintf("%s.%s", t.transpileExpr(e.Object), sanitizeIdent(e.Field))
	case *ast.StructInstantiation:
		return t.transpileStructInst(e)
	case *ast.TypeExpression:
		return fmt.Sprintf("type(%s).__name__", t.transpileExpr(e.Value))
	case *ast.CastExpression:
		return t.transpileCast(e)
	case *ast.ReferenceExpression:
		// References are plain variable accesses in Python.
		return sanitizeIdent(e.Name)
	case *ast.CopyExpression:
		return fmt.Sprintf("copy.copy(%s)", t.transpileExpr(e.Value))
	case *ast.LocationExpression:
		return fmt.Sprintf("hex(id(%s))", sanitizeIdent(e.Name))
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
		return fmt.Sprintf("isinstance(%s, %s)", t.transpileExpr(e.Value), sanitizeIdent(e.TypeName))
	default:
		return fmt.Sprintf("None  # unsupported expression: %T", expr)
	}
}

// transpileShown renders an expression the way English writes it out.
//
// Most values need the renderer, since Python writes a whole number with a
// decimal point, a boolean in title case and nothing as None. A literal that
// Python already spells identically is passed through, which keeps the output
// readable.
func (t *Transpiler) transpileShown(expr ast.Expression) string {
	if rendersAsItself(expr) {
		return t.transpileExpr(expr)
	}
	return fmt.Sprintf("_show(%s)", t.transpileExpr(expr))
}

// rendersAsItself reports whether an expression's Python form already reads
// the way English would write it.
func rendersAsItself(expr ast.Expression) bool {
	switch expr.(type) {
	case *ast.StringLiteral:
		// Text is printed as itself in both languages.
		return true
	case *ast.NumberLiteral:
		// formatNumber already emits English's spelling of the number.
		return true
	}
	return false
}

func (t *Transpiler) transpileListLit(elements []ast.Expression) string {
	parts := make([]string, len(elements))
	for i, el := range elements {
		parts[i] = t.transpileExpr(el)
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

// transpileRangeLit translates a range, which includes its end and runs
// downwards when the end is below the start.
//
// A helper, rather than the conditional expression this used to emit: that
// named start, end and step up to three times each, so any of them that did
// something as well as producing a value did it repeatedly, and the line was
// unreadable.
func (t *Transpiler) transpileRangeLit(e *ast.RangeLiteral) string {
	start := t.transpileExpr(e.Start)
	end := t.transpileExpr(e.End)
	if e.Step != nil {
		return fmt.Sprintf("_range(%s, %s, %s)", start, end, t.transpileExpr(e.Step))
	}
	return fmt.Sprintf("_range(%s, %s)", start, end)
}

func (t *Transpiler) transpileBinaryExpr(e *ast.BinaryExpression) string {
	left := t.transpileExpr(e.Left)
	right := t.transpileExpr(e.Right)

	// The remainder is the one operator with no Python equivalent: English
	// truncates both operands and takes the sign of the dividend, while
	// Python's % is floored and takes the sign of the divisor, so -7 % 3 was
	// -1 in English and 2 here.
	if isRemainder(e.Operator) {
		return fmt.Sprintf("_remainder(%s, %s)", left, right)
	}

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
	return fmt.Sprintf("%s.%s(%s)", obj, sanitizeIdent(e.MethodName), strings.Join(args, ", "))
}

func (t *Transpiler) transpileCast(e *ast.CastExpression) string {
	inner := t.transpileExpr(e.Value)
	switch strings.ToLower(ast.TypeName(e.Type)) {
	case "number", "float":
		return fmt.Sprintf("float(%s)", inner)
	case "integer", "int":
		return fmt.Sprintf("int(%s)", inner)
	case "text", "string", "str":
		// English's renderer, not Python's str(): a whole number has no
		// decimal point, a boolean is lower case, nothing is "nothing".
		return fmt.Sprintf("_show(%s)", inner)
	case "boolean", "bool":
		// English reads the words a person would write and refuses anything
		// else; Python's bool() calls every non-empty string true, so casting
		// "no" gave False in English and True here.
		return fmt.Sprintf("_to_bool(%s)", inner)
	default:
		// Treat any other cast as a constructor / type call.
		return fmt.Sprintf("%s(%s)", sanitizeIdent(ast.TypeName(e.Type)), inner)
	}
}

func (t *Transpiler) transpileStructInst(e *ast.StructInstantiation) string {
	// The field names have to be escaped the same way the __init__ parameters
	// were, or the call does not match the definition: a field called "class"
	// was emitted as Point(class=1) against def __init__(self, class_=0).
	args := make([]string, 0, len(e.FieldOrder))
	for _, name := range e.FieldOrder {
		args = append(args, fmt.Sprintf("%s=%s", sanitizeIdent(name), t.transpileExpr(e.FieldValues[name])))
	}
	return fmt.Sprintf("%s(%s)", sanitizeIdent(e.StructName), strings.Join(args, ", "))
}
