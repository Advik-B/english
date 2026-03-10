// Package disasm – expressions.go
//
// expr() renders any AST expression node as a compact, operator-notation
// string suitable for appearing as an operand on a single disassembly line.
// argList() formats a function-call argument list.
// symOp() maps English-prose operator names to conventional symbols.
package disasm

import (
	"fmt"
	"strings"

	"english/ast"
)

// ─── Expression renderer ──────────────────────────────────────────────────────

// expr renders an expression as a compact, operator-notation string.
// It is intentionally brief because expressions appear as operands on a single
// disassembly line; complex sub-expressions are shown in full but stay inline.
func (d *disassembler) expr(node ast.Expression) string {
	if node == nil {
		return d.s(styleNull, "null")
	}
	switch ex := node.(type) {

	case *ast.NumberLiteral:
		return d.s(styleNum, fmt.Sprintf("%g", ex.Value))

	case *ast.StringLiteral:
		return d.s(styleStr, `"`+ex.Value+`"`)

	case *ast.BooleanLiteral:
		if ex.Value {
			return d.s(styleBool, "true")
		}
		return d.s(styleBool, "false")

	case *ast.NothingLiteral:
		return d.s(styleNull, "null")

	case *ast.ListLiteral:
		elems := make([]string, len(ex.Elements))
		for i, e := range ex.Elements {
			elems[i] = d.expr(e)
		}
		return d.s(stylePunct, "[") +
			strings.Join(elems, d.s(stylePunct, ", ")) +
			d.s(stylePunct, "]")

	case *ast.Identifier:
		return d.s(styleIdent, ex.Name)

	case *ast.BinaryExpression:
		left := d.expr(ex.Left)
		right := d.expr(ex.Right)
		op := d.s(styleOp, d.opStr(ex.Operator))
		return d.s(stylePunct, "(") + left + " " + op + " " + right + d.s(stylePunct, ")")

	case *ast.UnaryExpression:
		return d.s(styleOp, d.opStr(ex.Operator)) + d.expr(ex.Right)

	case *ast.FunctionCall:
		return d.s(styleLabel, ex.Name) + d.argList(ex.Arguments)

	case *ast.IndexExpression:
		return d.expr(ex.List) +
			d.s(stylePunct, "[") + d.expr(ex.Index) + d.s(stylePunct, "]")

	case *ast.LengthExpression:
		return d.s(styleOpcodeCall, "len") +
			d.s(stylePunct, "(") + d.expr(ex.List) + d.s(stylePunct, ")")

	case *ast.LocationExpression:
		return d.s(styleOp, "&") + d.s(styleIdent, ex.Name)

	case *ast.ErrorTypeCheckExpression:
		return d.expr(ex.Value) + " " + d.s(styleOp, "is") + " " + d.s(styleIdent, ex.TypeName)

	case *ast.CastExpression:
		return d.s(stylePunct, "(") +
			d.expr(ex.Value) +
			d.s(stylePunct, ")") +
			d.s(styleMeta, " as ") +
			d.s(styleType, ex.TypeName)

	case *ast.AskExpression:
		if ex.Prompt != nil {
			return d.s(styleOpcodeIO, "ask") +
				d.s(stylePunct, "(") + d.expr(ex.Prompt) + d.s(stylePunct, ")")
		}
		return d.s(styleOpcodeIO, "ask") + d.s(stylePunct, "()")

	case *ast.ArrayLiteral:
		elems := make([]string, len(ex.Elements))
		for i, e := range ex.Elements {
			elems[i] = d.expr(e)
		}
		typeTag := ""
		if ex.ElementType != "" {
			typeTag = d.s(styleType, ex.ElementType+"[]")
		}
		return typeTag + d.s(stylePunct, "[") +
			strings.Join(elems, d.s(stylePunct, ", ")) +
			d.s(stylePunct, "]")

	case *ast.LookupTableLiteral:
		return d.s(stylePunct, "{}")

	case *ast.LookupKeyAccess:
		return d.expr(ex.Table) +
			d.s(stylePunct, "[") + d.expr(ex.Key) + d.s(stylePunct, "]")

	case *ast.HasExpression:
		return d.expr(ex.Table) + " " + d.s(styleOp, "has") + " " + d.expr(ex.Key)

	case *ast.NilCheckExpression:
		if ex.IsSomethingCheck {
			return d.expr(ex.Value) + " " + d.s(styleOp, "is something")
		}
		return d.expr(ex.Value) + " " + d.s(styleOp, "is nothing")

	case *ast.MethodCall:
		return d.expr(ex.Object) +
			d.s(stylePunct, ".") +
			d.s(styleLabel, ex.MethodName) +
			d.argList(ex.Arguments)

	case *ast.FieldAccess:
		return d.expr(ex.Object) +
			d.s(stylePunct, ".") +
			d.s(styleIdent, ex.Field)

	case *ast.StructInstantiation:
		if len(ex.FieldOrder) == 0 {
			return d.s(styleOpcodeDecl, "new") + d.s(stylePunct, "(") +
				d.s(styleLabel, ex.StructName) + d.s(stylePunct, ")")
		}
		parts := make([]string, 0, len(ex.FieldOrder))
		for _, k := range ex.FieldOrder {
			v := ex.FieldValues[k]
			parts = append(parts,
				d.s(styleIdent, k)+d.s(styleArrow, "←")+d.expr(v))
		}
		return d.s(styleOpcodeDecl, "new") +
			d.s(stylePunct, "(") +
			d.s(styleLabel, ex.StructName) +
			d.s(stylePunct, " {") +
			strings.Join(parts, d.s(stylePunct, ", ")) +
			d.s(stylePunct, "})")

	default:
		return d.s(styleNull, "?")
	}
}

// argList renders a function-call argument list as "(a, b, c)".
func (d *disassembler) argList(args []ast.Expression) string {
	if len(args) == 0 {
		return d.s(stylePunct, "()")
	}
	parts := make([]string, len(args))
	for i, a := range args {
		parts[i] = d.expr(a)
	}
	return d.s(stylePunct, "(") +
		strings.Join(parts, d.s(stylePunct, ", ")) +
		d.s(stylePunct, ")")
}

// ─── Operator symbol mapping ──────────────────────────────────────────────────

// symOp converts English-phrased operator names (as stored in the AST) to their
// conventional symbolic equivalents so that the disassembly does not contain
// English prose.
func symOp(op string) string {
	switch op {
	case "is equal to":
		return "=="
	case "is not equal to":
		return "!="
	case "is less than":
		return "<"
	case "is greater than":
		return ">"
	case "is less than or equal to":
		return "<="
	case "is greater than or equal to":
		return ">="
	case "and":
		return "&&"
	case "or":
		return "||"
	case "not":
		return "!"
	default:
		// +, -, *, /, % and any future operators pass through unchanged.
		return op
	}
}
