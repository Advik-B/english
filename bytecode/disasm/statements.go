// Package disasm – statements.go
//
// stmt() dispatches over every AST statement node type and emits the
// corresponding opcode line(s).  Nested bodies (function bodies, loop bodies,
// if branches, try/catch blocks, struct fields) are indented by incrementing
// d.depth around the recursive calls.
package disasm

import (
	"fmt"
	"strings"

	"english/ast"
)

// stmt emits one or more disassembly lines for a single AST statement.
func (d *disassembler) stmt(node ast.Statement) {
	switch s := node.(type) {

	case *ast.VariableDecl:
		name := d.s(styleIdent, s.Name)
		constTag := ""
		if s.IsConstant {
			constTag = " " + d.s(styleConst, "[const]")
		}
		arrow := d.s(styleArrow, "←")
		d.emit(styleOpcodeDecl, "DECLARE_VAR",
			name+constTag+"  "+arrow+"  "+d.expr(s.Value))

	case *ast.TypedVariableDecl:
		name := d.s(styleIdent, s.Name)
		typeTag := d.s(styleType, ":"+s.TypeName)
		constTag := ""
		if s.IsConstant {
			constTag = " " + d.s(styleConst, "[const]")
		}
		arrow := d.s(styleArrow, "←")
		d.emit(styleOpcodeDecl, "DECLARE_VAR",
			name+typeTag+constTag+"  "+arrow+"  "+d.expr(s.Value))

	case *ast.ErrorTypeDecl:
		name := d.s(styleIdent, s.Name)
		parentPart := ""
		if s.ParentType != "" {
			parentPart = "  " + d.s(styleOp, "extends") + "  " + d.s(styleIdent, s.ParentType)
		}
		d.emit(styleOpcodeDecl, "DECL_ERROR_TYPE", name+parentPart)

	case *ast.Assignment:
		name := d.s(styleIdent, s.Name)
		arrow := d.s(styleArrow, "←")
		d.emit(styleOpcodeAssign, "ASSIGN", name+"  "+arrow+"  "+d.expr(s.Value))

	case *ast.FunctionDecl:
		params := make([]string, len(s.Parameters))
		for i, p := range s.Parameters {
			params[i] = d.s(styleIdent, p)
		}
		paramStr := d.s(stylePunct, "(") +
			strings.Join(params, d.s(stylePunct, ", ")) +
			d.s(stylePunct, ")")
		d.emit(styleOpcodeDecl, "FUNC_DECL",
			d.s(styleLabel, s.Name)+"  "+paramStr)
		d.depth++
		for _, child := range s.Body {
			d.stmt(child)
		}
		d.depth--
		d.emitLabel(styleOpcodeEnd,
			fmt.Sprintf("%-18s", "END_FUNC"),
			d.s(styleMeta, s.Name))

	case *ast.CallStatement:
		if s.FunctionCall != nil {
			d.emit(styleOpcodeCall, "CALL",
				d.s(styleLabel, s.FunctionCall.Name)+d.argList(s.FunctionCall.Arguments))
		}

	case *ast.IfStatement:
		d.emit(styleOpcodeControl, "IF", d.expr(s.Condition))
		d.depth++
		for _, child := range s.Then {
			d.stmt(child)
		}
		d.depth--
		for _, ei := range s.ElseIf {
			d.emitLabel(styleOpcodeControl, fmt.Sprintf("%-18s", "ELSE_IF"), d.expr(ei.Condition))
			d.depth++
			for _, child := range ei.Body {
				d.stmt(child)
			}
			d.depth--
		}
		if len(s.Else) > 0 {
			d.emitLabel(styleOpcodeControl, fmt.Sprintf("%-18s", "ELSE"), "")
			d.depth++
			for _, child := range s.Else {
				d.stmt(child)
			}
			d.depth--
		}
		d.emitLabel(styleOpcodeEnd, fmt.Sprintf("%-18s", "END_IF"), "")

	case *ast.WhileLoop:
		d.emit(styleOpcodeControl, "WHILE", d.expr(s.Condition))
		d.depth++
		for _, child := range s.Body {
			d.stmt(child)
		}
		d.depth--
		d.emitLabel(styleOpcodeEnd, fmt.Sprintf("%-18s", "END_WHILE"), "")

	case *ast.ForLoop:
		d.emit(styleOpcodeControl, "FOR_LOOP", d.expr(s.Count)+"  "+d.s(styleMeta, "times"))
		d.depth++
		for _, child := range s.Body {
			d.stmt(child)
		}
		d.depth--
		d.emitLabel(styleOpcodeEnd, fmt.Sprintf("%-18s", "END_FOR_LOOP"), "")

	case *ast.ForEachLoop:
		item := d.s(styleIdent, s.Item)
		list := d.expr(s.List)
		d.emit(styleOpcodeControl, "FOR_EACH",
			item+"  "+d.s(styleOp, "in")+"  "+list)
		d.depth++
		for _, child := range s.Body {
			d.stmt(child)
		}
		d.depth--
		d.emitLabel(styleOpcodeEnd, fmt.Sprintf("%-18s", "END_FOR_EACH"), "")

	case *ast.IndexAssignment:
		listName := d.s(styleIdent, s.ListName)
		idxPart := d.s(stylePunct, "[") + d.expr(s.Index) + d.s(stylePunct, "]")
		arrow := d.s(styleArrow, "←")
		d.emit(styleOpcodeAssign, "INDEX_ASSIGN",
			listName+idxPart+"  "+arrow+"  "+d.expr(s.Value))

	case *ast.ReturnStatement:
		d.emit(styleOpcodeControl, "RETURN", d.expr(s.Value))

	case *ast.OutputStatement:
		opcode := "OUTPUT_PRINT"
		if !s.Newline {
			opcode = "OUTPUT_WRITE"
		}
		vals := make([]string, len(s.Values))
		for i, v := range s.Values {
			vals[i] = d.expr(v)
		}
		d.emit(styleOpcodeIO, opcode,
			strings.Join(vals, d.s(stylePunct, ", ")))

	case *ast.ToggleStatement:
		d.emit(styleOpcodeAssign, "TOGGLE", d.s(styleIdent, s.Name))

	case *ast.BreakStatement:
		d.emit(styleOpcodeControl, "BREAK", "")

	case *ast.ContinueStatement:
		d.emit(styleOpcodeControl, "CONTINUE", "")

	case *ast.SwapStatement:
		d.emit(styleOpcodeAssign, "SWAP",
			d.s(styleIdent, s.Name1)+"  "+d.s(styleOp, "↔")+"  "+d.s(styleIdent, s.Name2))

	case *ast.RaiseStatement:
		errType := ""
		if s.ErrorType != "" {
			errType = "  " + d.s(styleOp, "as") + "  " + d.s(styleIdent, s.ErrorType)
		}
		d.emit(styleOpcodeControl, "RAISE", d.expr(s.Message)+errType)

	case *ast.TryStatement:
		d.emitLabel(styleOpcodeControl, fmt.Sprintf("%-18s", "TRY"), "")
		d.depth++
		for _, child := range s.TryBody {
			d.stmt(child)
		}
		d.depth--
		catchLabel := "ON_ERROR"
		catchExtra := ""
		if s.ErrorType != "" {
			catchExtra = d.s(styleIdent, s.ErrorType)
		}
		if s.ErrorVar != "" {
			varPart := d.s(styleMeta, "→") + "  " + d.s(styleIdent, s.ErrorVar)
			if catchExtra != "" {
				catchExtra += "  " + varPart
			} else {
				catchExtra = varPart
			}
		}
		d.emitLabel(styleOpcodeControl, fmt.Sprintf("%-18s", catchLabel), catchExtra)
		d.depth++
		for _, child := range s.ErrorBody {
			d.stmt(child)
		}
		d.depth--
		if len(s.FinallyBody) > 0 {
			d.emitLabel(styleOpcodeControl, fmt.Sprintf("%-18s", "FINALLY"), "")
			d.depth++
			for _, child := range s.FinallyBody {
				d.stmt(child)
			}
			d.depth--
		}
		d.emitLabel(styleOpcodeEnd, fmt.Sprintf("%-18s", "END_TRY"), "")

	case *ast.StructDecl:
		d.emit(styleOpcodeDecl, "STRUCT_DECL", d.s(styleLabel, s.Name))
		d.depth++
		for _, f := range s.Fields {
			typeTag := d.s(styleType, ":"+f.TypeName)
			defPart := ""
			if f.DefaultValue != nil {
				defPart = "  " + d.s(styleArrow, "←") + "  " + d.expr(f.DefaultValue)
			}
			d.emitLabel(styleOpcodeDecl,
				fmt.Sprintf("%-18s", "FIELD"),
				d.s(styleIdent, f.Name)+typeTag+defPart)
		}
		for _, m := range s.Methods {
			params := make([]string, len(m.Parameters))
			for i, p := range m.Parameters {
				params[i] = d.s(styleIdent, p)
			}
			paramStr := d.s(stylePunct, "(") +
				strings.Join(params, d.s(stylePunct, ", ")) +
				d.s(stylePunct, ")")
			d.emitLabel(styleOpcodeDecl,
				fmt.Sprintf("%-18s", "METHOD"),
				d.s(styleLabel, m.Name)+paramStr)
		}
		d.depth--
		d.emitLabel(styleOpcodeEnd, fmt.Sprintf("%-18s", "END_STRUCT"), d.s(styleMeta, s.Name))

	case *ast.FieldAssignment:
		arrow := d.s(styleArrow, "←")
		d.emit(styleOpcodeAssign, "FIELD_ASSIGN",
			d.s(styleIdent, s.ObjectName)+d.s(stylePunct, ".")+d.s(styleIdent, s.Field)+
				"  "+arrow+"  "+d.expr(s.Value))

	case *ast.LookupKeyAssignment:
		arrow := d.s(styleArrow, "←")
		d.emit(styleOpcodeAssign, "LOOKUP_ASSIGN",
			d.s(styleIdent, s.TableName)+
				d.s(stylePunct, "[")+d.expr(s.Key)+d.s(stylePunct, "]")+
				"  "+arrow+"  "+d.expr(s.Value))

	case *ast.ImportStatement:
		path := d.s(styleStr, `"`+s.Path+`"`)
		detail := ""
		switch {
		case s.ImportAll:
			detail = "  " + d.s(styleMeta, "(import all)")
		case len(s.Items) > 0:
			items := make([]string, len(s.Items))
			for i, it := range s.Items {
				items[i] = d.s(styleIdent, it)
			}
			detail = "  " + d.s(stylePunct, "(") +
				strings.Join(items, d.s(stylePunct, ", ")) +
				d.s(stylePunct, ")")
		}
		if s.IsSafe {
			detail += "  " + d.s(styleConst, "[safe]")
		}
		d.emit(styleOpcodeDecl, "IMPORT", path+detail)
	}
}
