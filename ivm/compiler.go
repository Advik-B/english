package ivm

import (
	"english/ast"
	"fmt"
)

// Compiler walks an AST and emits instructions into a Chunk.
type Compiler struct {
	chunk            *Chunk
	loopStarts       []int   // jump-back targets for while loops (before condition test)
	loopContinues    [][]int // like loopEnds: positions of continue JUMPs to patch (for for/for-each)
	loopEnds         [][]int // positions of break JUMPs to patch to loop end
	loopScopeDepths  []int   // scope depth at the start of each loop's body
	scopeDepth       int     // current number of active scopes (each PUSH_SCOPE increments)
	funcName         string  // name of the function being compiled (for error messages)
	counter          int     // for generating unique hidden variable names
}

// Compile compiles an ast.Program to a Chunk.
func Compile(prog *ast.Program) (*Chunk, error) {
	c := &Compiler{chunk: NewChunk()}
	if err := c.compileStatements(prog.Statements); err != nil {
		return nil, err
	}
	return c.chunk, nil
}

func (c *Compiler) nextHidden() string {
	c.counter++
	return fmt.Sprintf("__hidden_%d", c.counter)
}

func (c *Compiler) compileStatements(stmts []ast.Statement) error {
	for _, stmt := range stmts {
		if err := c.compileStatement(stmt); err != nil {
			return err
		}
	}
	return nil
}

func (c *Compiler) compileStatement(stmt ast.Statement) error {
	switch s := stmt.(type) {
	case *ast.CommentStatement:
		// nothing

	case *ast.VariableDecl:
		if s.Value != nil {
			if err := c.compileExpression(s.Value); err != nil {
				return err
			}
		} else {
			c.chunk.Emit(OP_LOAD_NOTHING, 0)
		}
		nIdx := c.chunk.AddName(s.Name)
		if s.IsConstant {
			c.chunk.Emit(OP_DEFINE_CONST, nIdx)
		} else {
			c.chunk.Emit(OP_DEFINE_VAR, nIdx)
		}

	case *ast.TypedVariableDecl:
		// Stack: compile type name first, then value
		typeIdx := c.chunk.AddConst(s.TypeName)
		c.chunk.Emit(OP_LOAD_CONST, typeIdx)
		if s.Value != nil {
			if err := c.compileExpression(s.Value); err != nil {
				return err
			}
		} else {
			c.chunk.Emit(OP_LOAD_NOTHING, 0)
		}
		nIdx := c.chunk.AddName(s.Name)
		if s.IsConstant {
			c.chunk.Emit(OP_DEFINE_TYPED_CONST, nIdx)
		} else {
			c.chunk.Emit(OP_DEFINE_TYPED, nIdx)
		}

	case *ast.Assignment:
		if s.Line > 0 {
			c.chunk.Emit(OP_SET_LINE, uint32(s.Line))
		}
		if err := c.compileExpression(s.Value); err != nil {
			return err
		}
		nIdx := c.chunk.AddName(s.Name)
		c.chunk.Emit(OP_STORE_VAR, nIdx)

	case *ast.IndexAssignment:
		// compile index, compile value
		if err := c.compileExpression(s.Index); err != nil {
			return err
		}
		if err := c.compileExpression(s.Value); err != nil {
			return err
		}
		nIdx := c.chunk.AddName(s.ListName)
		c.chunk.Emit(OP_INDEX_SET, nIdx)

	case *ast.LookupKeyAssignment:
		// compile key, compile value
		if err := c.compileExpression(s.Key); err != nil {
			return err
		}
		if err := c.compileExpression(s.Value); err != nil {
			return err
		}
		nIdx := c.chunk.AddName(s.TableName)
		c.chunk.Emit(OP_LOOKUP_SET, nIdx)

	case *ast.FieldAssignment:
		// Load the struct instance, compile the value, then SET_FIELD
		// Stack: [struct_instance, new_value]
		objIdx := c.chunk.AddName(s.ObjectName)
		c.chunk.Emit(OP_LOAD_VAR, objIdx)
		if err := c.compileExpression(s.Value); err != nil {
			return err
		}
		fieldIdx := c.chunk.AddName(s.Field)
		c.chunk.Emit(OP_SET_FIELD, fieldIdx)

	case *ast.FunctionDecl:
		// Compile function body as a child FuncChunk
		bodyChunk, err := c.compileFuncBody(s.Name, s.Parameters, s.Body)
		if err != nil {
			return err
		}
		funcIdx := uint32(len(c.chunk.Funcs))
		c.chunk.Funcs = append(c.chunk.Funcs, bodyChunk)
		c.chunk.Emit(OP_DEFINE_FUNC, funcIdx)

	case *ast.CallStatement:
		if s.FunctionCall != nil {
			if err := c.compileExpression(s.FunctionCall); err != nil {
				return err
			}
		} else if s.MethodCall != nil {
			if err := c.compileExpression(s.MethodCall); err != nil {
				return err
			}
		}
		c.chunk.Emit(OP_POP, 0)

	case *ast.ReturnStatement:
		if s.Value != nil {
			if err := c.compileExpression(s.Value); err != nil {
				return err
			}
		} else {
			c.chunk.Emit(OP_LOAD_NOTHING, 0)
		}
		c.chunk.Emit(OP_RETURN, 0)

	case *ast.OutputStatement:
		count := uint32(len(s.Values))
		for _, v := range s.Values {
			if err := c.compileExpression(v); err != nil {
				return err
			}
		}
		newlineFlag := uint32(0)
		if s.Newline {
			newlineFlag = 1
		}
		c.chunk.Emit(OP_PRINT, count<<1|newlineFlag)

	case *ast.IfStatement:
		if err := c.compileIfStatement(s); err != nil {
			return err
		}

	case *ast.WhileLoop:
		if err := c.compileWhileLoop(s); err != nil {
			return err
		}

	case *ast.ForLoop:
		if err := c.compileForLoop(s); err != nil {
			return err
		}

	case *ast.ForEachLoop:
		if err := c.compileForEachLoop(s); err != nil {
			return err
		}

	case *ast.ToggleStatement:
		nIdx := c.chunk.AddName(s.Name)
		c.chunk.Emit(OP_TOGGLE_VAR, nIdx)

	case *ast.SwapStatement:
		n1 := c.chunk.AddName(s.Name1)
		n2 := c.chunk.AddName(s.Name2)
		c.chunk.Emit(OP_SWAP_VARS, n1<<16|n2)

	case *ast.BreakStatement:
		if len(c.loopEnds) == 0 {
			return fmt.Errorf("break outside loop")
		}
		// pop ALL scopes including the loop body scope (exit the loop entirely)
		loopBodyDepth := c.loopScopeDepths[len(c.loopScopeDepths)-1]
		for c.scopeDepth >= loopBodyDepth {
			c.chunk.Emit(OP_POP_SCOPE, 0)
			c.scopeDepth--
		}
		pos := c.chunk.CurrentPos()
		c.chunk.Emit(OP_JUMP, 0) // placeholder
		last := len(c.loopEnds) - 1
		c.loopEnds[last] = append(c.loopEnds[last], pos)

	case *ast.ContinueStatement:
		if len(c.loopStarts) == 0 {
			return fmt.Errorf("continue outside loop")
		}
		// pop ALL scopes including the loop body scope, then jump to the continue target.
		// The loop will re-push a fresh scope for the next iteration.
		loopBodyDepth := c.loopScopeDepths[len(c.loopScopeDepths)-1]
		for c.scopeDepth >= loopBodyDepth {
			c.chunk.Emit(OP_POP_SCOPE, 0)
			c.scopeDepth--
		}
		last := len(c.loopContinues) - 1
		if last >= 0 && c.loopContinues[last] != nil {
			// for/for-each loop: emit patchable JUMP (patched to increment/decrement pos after body)
			pos := c.chunk.CurrentPos()
			c.chunk.Emit(OP_JUMP, 0) // placeholder
			c.loopContinues[last] = append(c.loopContinues[last], pos)
		} else {
			// while loop: continue target is loopStart (known at compile time)
			start := c.loopStarts[len(c.loopStarts)-1]
			c.chunk.Emit(OP_JUMP, uint32(start))
		}

	case *ast.TryStatement:
		if err := c.compileTryStatement(s); err != nil {
			return err
		}

	case *ast.RaiseStatement:
		if err := c.compileExpression(s.Message); err != nil {
			return err
		}
		var typeIdx uint32
		if s.ErrorType != "" {
			typeIdx = c.chunk.AddName(s.ErrorType) + 1 // 0 reserved for generic
		}
		c.chunk.Emit(OP_RAISE, typeIdx)

	case *ast.ErrorTypeDecl:
		nIdx := c.chunk.AddName(s.Name)
		var pIdx uint32
		if s.ParentType != "" {
			pIdx = c.chunk.AddName(s.ParentType) + 1 // 0 = no parent
		}
		c.chunk.Emit(OP_DEFINE_ERROR_TYPE, nIdx<<16|pIdx)

	case *ast.StructDecl:
		if err := c.compileStructDecl(s); err != nil {
			return err
		}

	case *ast.ImportStatement:
		// Push the items count constant and path
		// operand = importAll<<2 | isSafe<<1 | hasItems
		flags := uint32(0)
		if s.ImportAll {
			flags |= 4
		}
		if s.IsSafe {
			flags |= 2
		}
		if len(s.Items) > 0 {
			flags |= 1
		}
		// Push path string
		pathIdx := c.chunk.AddConst(s.Path)
		c.chunk.Emit(OP_LOAD_CONST, pathIdx)
		// Push items as a list constant if needed
		if len(s.Items) > 0 {
			items := make([]interface{}, len(s.Items))
			for i, item := range s.Items {
				items[i] = item
			}
			itemsIdx := c.chunk.AddConst(items)
			c.chunk.Emit(OP_LOAD_CONST, itemsIdx)
		}
		c.chunk.Emit(OP_IMPORT, flags)

	default:
		return fmt.Errorf("ivm compiler: unsupported statement type %T", stmt)
	}
	return nil
}

func (c *Compiler) compileExpression(expr ast.Expression) error {
	switch e := expr.(type) {
	case *ast.NumberLiteral:
		idx := c.chunk.AddConst(e.Value)
		c.chunk.Emit(OP_LOAD_CONST, idx)

	case *ast.StringLiteral:
		idx := c.chunk.AddConst(e.Value)
		c.chunk.Emit(OP_LOAD_CONST, idx)

	case *ast.BooleanLiteral:
		idx := c.chunk.AddConst(e.Value)
		c.chunk.Emit(OP_LOAD_CONST, idx)

	case *ast.NothingLiteral:
		c.chunk.Emit(OP_LOAD_NOTHING, 0)

	case *ast.Identifier:
		nIdx := c.chunk.AddName(e.Name)
		c.chunk.Emit(OP_LOAD_VAR, nIdx)

	case *ast.BinaryExpression:
		if err := c.compileBinaryExpr(e); err != nil {
			return err
		}

	case *ast.UnaryExpression:
		if err := c.compileExpression(e.Right); err != nil {
			return err
		}
		switch e.Operator {
		case "-":
			c.chunk.Emit(OP_UNARY_OP, uint32(UnaryNeg))
		case "not":
			c.chunk.Emit(OP_UNARY_OP, uint32(UnaryNot))
		default:
			return fmt.Errorf("unknown unary operator: %s", e.Operator)
		}

	case *ast.ListLiteral:
		for _, elem := range e.Elements {
			if err := c.compileExpression(elem); err != nil {
				return err
			}
		}
		c.chunk.Emit(OP_BUILD_LIST, uint32(len(e.Elements)))

	case *ast.ArrayLiteral:
		for _, elem := range e.Elements {
			if err := c.compileExpression(elem); err != nil {
				return err
			}
		}
		typeIdx := c.chunk.AddConst(e.ElementType)
		c.chunk.Emit(OP_LOAD_CONST, typeIdx)
		c.chunk.Emit(OP_BUILD_ARRAY, uint32(len(e.Elements)))

	case *ast.LookupTableLiteral:
		c.chunk.Emit(OP_BUILD_LOOKUP, 0)

	case *ast.IndexExpression:
		if err := c.compileExpression(e.List); err != nil {
			return err
		}
		if err := c.compileExpression(e.Index); err != nil {
			return err
		}
		c.chunk.Emit(OP_INDEX_GET, 0)

	case *ast.LengthExpression:
		if err := c.compileExpression(e.List); err != nil {
			return err
		}
		c.chunk.Emit(OP_LENGTH, 0)

	case *ast.FunctionCall:
		for _, arg := range e.Arguments {
			if err := c.compileExpression(arg); err != nil {
				return err
			}
		}
		nIdx := c.chunk.AddName(e.Name)
		argc := uint32(len(e.Arguments))
		c.chunk.Emit(OP_CALL, argc<<16|nIdx)

	case *ast.MethodCall:
		if err := c.compileExpression(e.Object); err != nil {
			return err
		}
		for _, arg := range e.Arguments {
			if err := c.compileExpression(arg); err != nil {
				return err
			}
		}
		mIdx := c.chunk.AddName(e.MethodName)
		argc := uint32(len(e.Arguments))
		c.chunk.Emit(OP_CALL_METHOD, argc<<16|mIdx)

	case *ast.LocationExpression:
		nIdx := c.chunk.AddName(e.Name)
		c.chunk.Emit(OP_LOCATION, nIdx)

	case *ast.TypeExpression:
		if err := c.compileExpression(e.Value); err != nil {
			return err
		}
		c.chunk.Emit(OP_TYPEOF, 0)

	case *ast.CastExpression:
		if err := c.compileExpression(e.Value); err != nil {
			return err
		}
		tIdx := c.chunk.AddName(e.TypeName)
		c.chunk.Emit(OP_CAST, tIdx)

	case *ast.NilCheckExpression:
		if err := c.compileExpression(e.Value); err != nil {
			return err
		}
		flag := uint32(0)
		if e.IsSomethingCheck {
			flag = 1
		}
		c.chunk.Emit(OP_NIL_CHECK, flag)

	case *ast.ErrorTypeCheckExpression:
		if err := c.compileExpression(e.Value); err != nil {
			return err
		}
		tIdx := c.chunk.AddName(e.TypeName)
		c.chunk.Emit(OP_ERROR_TYPE_CHECK, tIdx)

	case *ast.ReferenceExpression:
		nIdx := c.chunk.AddName(e.Name)
		c.chunk.Emit(OP_MAKE_REFERENCE, nIdx)

	case *ast.CopyExpression:
		if err := c.compileExpression(e.Value); err != nil {
			return err
		}
		c.chunk.Emit(OP_MAKE_COPY, 0)

	case *ast.AskExpression:
		if e.Prompt != nil {
			if err := c.compileExpression(e.Prompt); err != nil {
				return err
			}
			c.chunk.Emit(OP_ASK, 1)
		} else {
			c.chunk.Emit(OP_ASK, 0)
		}

	case *ast.LookupKeyAccess:
		if err := c.compileExpression(e.Table); err != nil {
			return err
		}
		if err := c.compileExpression(e.Key); err != nil {
			return err
		}
		c.chunk.Emit(OP_LOOKUP_GET, 0)

	case *ast.HasExpression:
		if err := c.compileExpression(e.Table); err != nil {
			return err
		}
		if err := c.compileExpression(e.Key); err != nil {
			return err
		}
		c.chunk.Emit(OP_LOOKUP_HAS, 0)

	case *ast.FieldAccess:
		if err := c.compileExpression(e.Object); err != nil {
			return err
		}
		fIdx := c.chunk.AddName(e.Field)
		c.chunk.Emit(OP_GET_FIELD, fIdx)

	case *ast.StructInstantiation:
		if err := c.compileStructInstantiation(e); err != nil {
			return err
		}

	default:
		return fmt.Errorf("ivm compiler: unsupported expression type %T", expr)
	}
	return nil
}

func (c *Compiler) compileBinaryExpr(e *ast.BinaryExpression) error {
	switch e.Operator {
	case "and":
		// Short-circuit AND:
		// compile left; JUMP_IF_FALSE -> false_label (pops); compile right; JUMP -> end; false_label: LOAD false; end:
		if err := c.compileExpression(e.Left); err != nil {
			return err
		}
		falseJump := c.chunk.CurrentPos()
		c.chunk.Emit(OP_JUMP_IF_FALSE, 0) // placeholder
		if err := c.compileExpression(e.Right); err != nil {
			return err
		}
		endJump := c.chunk.CurrentPos()
		c.chunk.Emit(OP_JUMP, 0)
		// false_label:
		c.chunk.PatchJump(falseJump, uint32(c.chunk.CurrentPos()))
		falseLitIdx := c.chunk.AddConst(false)
		c.chunk.Emit(OP_LOAD_CONST, falseLitIdx)
		// end:
		c.chunk.PatchJump(endJump, uint32(c.chunk.CurrentPos()))

	case "or":
		// Short-circuit OR:
		// compile left; JUMP_IF_TRUE -> true_label (pops); compile right; JUMP -> end; true_label: LOAD true; end:
		if err := c.compileExpression(e.Left); err != nil {
			return err
		}
		trueJump := c.chunk.CurrentPos()
		c.chunk.Emit(OP_JUMP_IF_TRUE, 0) // placeholder
		if err := c.compileExpression(e.Right); err != nil {
			return err
		}
		endJump := c.chunk.CurrentPos()
		c.chunk.Emit(OP_JUMP, 0)
		// true_label:
		c.chunk.PatchJump(trueJump, uint32(c.chunk.CurrentPos()))
		trueLitIdx := c.chunk.AddConst(true)
		c.chunk.Emit(OP_LOAD_CONST, trueLitIdx)
		// end:
		c.chunk.PatchJump(endJump, uint32(c.chunk.CurrentPos()))

	default:
		if err := c.compileExpression(e.Left); err != nil {
			return err
		}
		if err := c.compileExpression(e.Right); err != nil {
			return err
		}
		binOp, err := parseBinOp(e.Operator)
		if err != nil {
			return err
		}
		c.chunk.Emit(OP_BINARY_OP, uint32(binOp))
	}
	return nil
}

func parseBinOp(op string) (BinOp, error) {
	switch op {
	case "+":
		return BinAdd, nil
	case "-":
		return BinSub, nil
	case "*":
		return BinMul, nil
	case "/":
		return BinDiv, nil
	case "%":
		return BinMod, nil
	case "is equal to":
		return BinEq, nil
	case "is not equal to":
		return BinNeq, nil
	case "is less than":
		return BinLt, nil
	case "is less than or equal to":
		return BinLte, nil
	case "is greater than":
		return BinGt, nil
	case "is greater than or equal to":
		return BinGte, nil
	default:
		return 0, fmt.Errorf("unknown binary operator: %s", op)
	}
}


