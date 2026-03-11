package ivm

import "english/ast"

func (c *Compiler) compileIfStatement(s *ast.IfStatement) error {
	if err := c.compileExpression(s.Condition); err != nil {
		return err
	}
	skipThenPos := c.chunk.CurrentPos()
	c.chunk.Emit(OP_JUMP_IF_FALSE, 0)

	c.chunk.Emit(OP_PUSH_SCOPE, 0)
	c.scopeDepth++
	if err := c.compileStatements(s.Then); err != nil {
		return err
	}
	c.chunk.Emit(OP_POP_SCOPE, 0)
	c.scopeDepth--

	var endJumps []int

	if len(s.ElseIf) > 0 || len(s.Else) > 0 {
		endJumps = append(endJumps, c.chunk.CurrentPos())
		c.chunk.Emit(OP_JUMP, 0) // jump to end
	}

	c.chunk.PatchJump(skipThenPos, uint32(c.chunk.CurrentPos()))

	for _, elseif := range s.ElseIf {
		if err := c.compileExpression(elseif.Condition); err != nil {
			return err
		}
		skipElseIfPos := c.chunk.CurrentPos()
		c.chunk.Emit(OP_JUMP_IF_FALSE, 0)

		c.chunk.Emit(OP_PUSH_SCOPE, 0)
		c.scopeDepth++
		if err := c.compileStatements(elseif.Body); err != nil {
			return err
		}
		c.chunk.Emit(OP_POP_SCOPE, 0)
		c.scopeDepth--

		endJumps = append(endJumps, c.chunk.CurrentPos())
		c.chunk.Emit(OP_JUMP, 0) // jump to end

		c.chunk.PatchJump(skipElseIfPos, uint32(c.chunk.CurrentPos()))
	}

	if len(s.Else) > 0 {
		c.chunk.Emit(OP_PUSH_SCOPE, 0)
		c.scopeDepth++
		if err := c.compileStatements(s.Else); err != nil {
			return err
		}
		c.chunk.Emit(OP_POP_SCOPE, 0)
		c.scopeDepth--
	}

	endPos := uint32(c.chunk.CurrentPos())
	for _, pos := range endJumps {
		c.chunk.PatchJump(pos, endPos)
	}
	return nil
}

func (c *Compiler) compileWhileLoop(s *ast.WhileLoop) error {
	// Structure:
	// LOOP_START: compile condition; JUMP_IF_FALSE -> LOOP_END; PUSH_SCOPE; ...body...; POP_SCOPE; JUMP -> LOOP_START; LOOP_END:
	loopStart := c.chunk.CurrentPos()
	if err := c.compileExpression(s.Condition); err != nil {
		return err
	}
	exitJump := c.chunk.CurrentPos()
	c.chunk.Emit(OP_JUMP_IF_FALSE, 0)

	c.chunk.Emit(OP_PUSH_SCOPE, 0)
	c.scopeDepth++
	loopBodyDepth := c.scopeDepth

	c.loopStarts = append(c.loopStarts, loopStart)
	c.loopContinues = append(c.loopContinues, nil) // nil = while loop: continue uses loopStarts
	c.loopEnds = append(c.loopEnds, []int{})
	c.loopScopeDepths = append(c.loopScopeDepths, loopBodyDepth)

	if err := c.compileStatements(s.Body); err != nil {
		return err
	}

	// Restore scopeDepth in case breaks/continues modified it
	c.scopeDepth = loopBodyDepth
	c.chunk.Emit(OP_POP_SCOPE, 0)
	c.scopeDepth--

	c.chunk.Emit(OP_JUMP, uint32(loopStart))

	loopEnd := uint32(c.chunk.CurrentPos())
	c.chunk.PatchJump(exitJump, loopEnd)

	// Patch all break jumps
	breaks := c.loopEnds[len(c.loopEnds)-1]
	for _, pos := range breaks {
		c.chunk.PatchJump(pos, loopEnd)
	}
	c.loopStarts = c.loopStarts[:len(c.loopStarts)-1]
	c.loopContinues = c.loopContinues[:len(c.loopContinues)-1]
	c.loopEnds = c.loopEnds[:len(c.loopEnds)-1]
	c.loopScopeDepths = c.loopScopeDepths[:len(c.loopScopeDepths)-1]
	return nil
}

func (c *Compiler) compileForLoop(s *ast.ForLoop) error {
	// repeat N times: (counted loop using a hidden variable)
	// compile N, define __for_i = N
	// LOOP_START: LOAD __for_i; LOAD 0; GT; JUMP_IF_FALSE -> LOOP_END
	// PUSH_SCOPE; ...body...; POP_SCOPE
	// LOAD __for_i; LOAD 1; SUB; STORE __for_i; JUMP -> LOOP_START; LOOP_END:
	counterName := c.nextHidden()

	if err := c.compileExpression(s.Count); err != nil {
		return err
	}
	cnIdx := c.chunk.AddName(counterName)
	c.chunk.Emit(OP_DEFINE_VAR, cnIdx)

	loopStart := c.chunk.CurrentPos()

	c.chunk.Emit(OP_LOAD_VAR, cnIdx)
	zeroIdx := c.chunk.AddConst(float64(0))
	c.chunk.Emit(OP_LOAD_CONST, zeroIdx)
	c.chunk.Emit(OP_BINARY_OP, uint32(BinGt))

	exitJump := c.chunk.CurrentPos()
	c.chunk.Emit(OP_JUMP_IF_FALSE, 0)

	c.chunk.Emit(OP_PUSH_SCOPE, 0)
	c.scopeDepth++
	loopBodyDepth := c.scopeDepth

	c.loopStarts = append(c.loopStarts, loopStart)
	c.loopContinues = append(c.loopContinues, []int{}) // for loop: continue needs patchable JUMP
	c.loopEnds = append(c.loopEnds, []int{})
	c.loopScopeDepths = append(c.loopScopeDepths, loopBodyDepth)

	if err := c.compileStatements(s.Body); err != nil {
		return err
	}

	c.scopeDepth = loopBodyDepth
	c.chunk.Emit(OP_POP_SCOPE, 0)
	c.scopeDepth--

	// decrement counter — continue jumps land here
	decrementPos := uint32(c.chunk.CurrentPos())
	c.chunk.Emit(OP_LOAD_VAR, cnIdx)
	oneIdx := c.chunk.AddConst(float64(1))
	c.chunk.Emit(OP_LOAD_CONST, oneIdx)
	c.chunk.Emit(OP_BINARY_OP, uint32(BinSub))
	c.chunk.Emit(OP_STORE_VAR, cnIdx)

	c.chunk.Emit(OP_JUMP, uint32(loopStart))

	loopEnd := uint32(c.chunk.CurrentPos())
	c.chunk.PatchJump(exitJump, loopEnd)

	// Patch break and continue jumps
	breaks := c.loopEnds[len(c.loopEnds)-1]
	for _, pos := range breaks {
		c.chunk.PatchJump(pos, loopEnd)
	}
	conts := c.loopContinues[len(c.loopContinues)-1]
	for _, pos := range conts {
		c.chunk.PatchJump(pos, decrementPos)
	}
	c.loopStarts = c.loopStarts[:len(c.loopStarts)-1]
	c.loopContinues = c.loopContinues[:len(c.loopContinues)-1]
	c.loopEnds = c.loopEnds[:len(c.loopEnds)-1]
	c.loopScopeDepths = c.loopScopeDepths[:len(c.loopScopeDepths)-1]
	return nil
}

func (c *Compiler) compileForEachLoop(s *ast.ForEachLoop) error {
	// for each item in list:
	// compile list; define __each_list; define __each_idx = 0
	// LOOP_START: LOAD __each_idx; LOAD __each_list; LENGTH; LT; JUMP_IF_FALSE -> LOOP_END
	// PUSH_SCOPE; define item = __each_list[__each_idx]; ...body...; POP_SCOPE
	// LOAD __each_idx; LOAD 1; ADD; STORE __each_idx; JUMP -> LOOP_START; LOOP_END:
	listName := c.nextHidden()
	idxName := c.nextHidden()

	if err := c.compileExpression(s.List); err != nil {
		return err
	}
	listIdx := c.chunk.AddName(listName)
	c.chunk.Emit(OP_DEFINE_VAR, listIdx)

	startIdx := c.chunk.AddConst(float64(0)) // 0-based indexing
	c.chunk.Emit(OP_LOAD_CONST, startIdx)
	idxIdx := c.chunk.AddName(idxName)
	c.chunk.Emit(OP_DEFINE_VAR, idxIdx)

	loopStart := c.chunk.CurrentPos()

	c.chunk.Emit(OP_LOAD_VAR, idxIdx)
	c.chunk.Emit(OP_LOAD_VAR, listIdx)
	c.chunk.Emit(OP_LENGTH, 0)
	c.chunk.Emit(OP_BINARY_OP, uint32(BinLt)) // idx < len (0-based: stop when idx == len)

	exitJump := c.chunk.CurrentPos()
	c.chunk.Emit(OP_JUMP_IF_FALSE, 0)

	c.chunk.Emit(OP_PUSH_SCOPE, 0)
	c.scopeDepth++
	loopBodyDepth := c.scopeDepth

	c.loopStarts = append(c.loopStarts, loopStart)
	c.loopContinues = append(c.loopContinues, []int{}) // for-each: continue needs patchable JUMP
	c.loopEnds = append(c.loopEnds, []int{})
	c.loopScopeDepths = append(c.loopScopeDepths, loopBodyDepth)

	// define loop variable
	c.chunk.Emit(OP_LOAD_VAR, listIdx)
	c.chunk.Emit(OP_LOAD_VAR, idxIdx)
	c.chunk.Emit(OP_INDEX_GET, 0)
	itemIdx := c.chunk.AddName(s.Item)
	c.chunk.Emit(OP_DEFINE_VAR, itemIdx)

	if err := c.compileStatements(s.Body); err != nil {
		return err
	}

	c.scopeDepth = loopBodyDepth
	c.chunk.Emit(OP_POP_SCOPE, 0)
	c.scopeDepth--

	// increment index — continue jumps land here
	incrementPos := uint32(c.chunk.CurrentPos())
	c.chunk.Emit(OP_LOAD_VAR, idxIdx)
	oneIdx := c.chunk.AddConst(float64(1))
	c.chunk.Emit(OP_LOAD_CONST, oneIdx)
	c.chunk.Emit(OP_BINARY_OP, uint32(BinAdd))
	c.chunk.Emit(OP_STORE_VAR, idxIdx)

	c.chunk.Emit(OP_JUMP, uint32(loopStart))

	loopEnd := uint32(c.chunk.CurrentPos())
	c.chunk.PatchJump(exitJump, loopEnd)

	// Patch break and continue jumps
	breaks := c.loopEnds[len(c.loopEnds)-1]
	for _, pos := range breaks {
		c.chunk.PatchJump(pos, loopEnd)
	}
	conts := c.loopContinues[len(c.loopContinues)-1]
	for _, pos := range conts {
		c.chunk.PatchJump(pos, incrementPos)
	}
	c.loopStarts = c.loopStarts[:len(c.loopStarts)-1]
	c.loopContinues = c.loopContinues[:len(c.loopContinues)-1]
	c.loopEnds = c.loopEnds[:len(c.loopEnds)-1]
	c.loopScopeDepths = c.loopScopeDepths[:len(c.loopScopeDepths)-1]
	return nil
}

func (c *Compiler) compileTryStatement(s *ast.TryStatement) error {
	// Layout (no finally):
	//   TRY_BEGIN(catch_offset)
	//   [TRY_SET_ERRORTYPE(nameIdx+1)]   ← only when ErrorType is set
	//   ...try body...
	//   TRY_END(end_offset)              ← jumps past catch section
	//   catch_offset:
	//     PUSH_SCOPE
	//     CATCH(error_var_idx)
	//     ...catch body...
	//     POP_SCOPE
	//   end_offset:
	//
	// Layout (with finally):
	//   TRY_BEGIN(catch_offset)
	//   [TRY_SET_ERRORTYPE(nameIdx+1)]   ← only when ErrorType is set
	//   TRY_SET_FINALLY(0)               ← placeholder; patched to finally_offset
	//   ...try body...
	//   TRY_END(end_offset)              ← jumps to end_offset = finally start
	//   catch_offset:
	//     PUSH_SCOPE
	//     CATCH(error_var_idx)
	//     ...catch body...
	//     POP_SCOPE
	//   end_offset (= finally_offset):
	//   ...finally body...
	//   RERAISE_PENDING                  ← re-raises error if type mismatched
	//
	// When handleError detects a type mismatch AND finallyOffset is set, it
	// stores the error as frame.pendingError and jumps directly to finallyOffset,
	// skipping the entire catch section.  RERAISE_PENDING then re-propagates the
	// error after the finally body executes.

	tryBeginPos := c.chunk.CurrentPos()
	c.chunk.Emit(OP_TRY_BEGIN, 0) // placeholder for catch_offset

	// If there's a type filter, record it in the try frame at runtime.
	if s.ErrorType != "" {
		nameIdx := c.chunk.AddName(s.ErrorType)
		c.chunk.Emit(OP_TRY_SET_ERRORTYPE, nameIdx+1) // +1 so 0 means "no filter"
	}

	// If there's a finally block, reserve a placeholder for the finally offset.
	tryFinallyPos := -1
	if len(s.FinallyBody) > 0 {
		tryFinallyPos = c.chunk.CurrentPos()
		c.chunk.Emit(OP_TRY_SET_FINALLY, 0) // placeholder; patched below
	}

	if err := c.compileStatements(s.TryBody); err != nil {
		return err
	}

	tryEndPos := c.chunk.CurrentPos()
	c.chunk.Emit(OP_TRY_END, 0) // placeholder for end_offset (past catch body, before finally)

	// catch_offset:
	catchOffset := uint32(c.chunk.CurrentPos())
	c.chunk.PatchJump(tryBeginPos, catchOffset)

	// CATCH instruction — operand is the error variable name index only;
	// the type check has been moved to handleError.
	var errVarIdx uint32
	if s.ErrorVar != "" {
		errVarIdx = c.chunk.AddName(s.ErrorVar)
	}
	// catch section: always push a fresh scope so the error variable is scoped
	// to this handler and does not leak into subsequent try blocks (even if the
	// catch body is empty, the scope is needed to contain the error variable).
	c.chunk.Emit(OP_PUSH_SCOPE, 0)
	c.scopeDepth++
	c.chunk.Emit(OP_CATCH, errVarIdx)

	// catch body (inside the same scope as the error variable)
	if len(s.ErrorBody) > 0 {
		if err := c.compileStatements(s.ErrorBody); err != nil {
			return err
		}
	}

	// Always pop the catch scope (paired with the PUSH_SCOPE above).
	c.chunk.Emit(OP_POP_SCOPE, 0)
	c.scopeDepth--

	// end_offset (past catch body, before finally):
	endOffset := uint32(c.chunk.CurrentPos())
	c.chunk.PatchJump(tryEndPos, endOffset)

	// Patch the TRY_SET_FINALLY placeholder with the actual finally offset.
	if tryFinallyPos >= 0 {
		c.chunk.PatchJump(tryFinallyPos, endOffset)
	}

	// finally body (always runs)
	if len(s.FinallyBody) > 0 {
		if err := c.compileStatements(s.FinallyBody); err != nil {
			return err
		}
		// After the finally body, re-raise any pending error from a type-mismatch path.
		c.chunk.Emit(OP_RERAISE_PENDING, 0)
	}

	return nil
}
