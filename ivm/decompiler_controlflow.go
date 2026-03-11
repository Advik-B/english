package ivm

import (
	"strings"
)

// ─── control-flow pattern recognition ────────────────────────────────────────

// isWhileLoopExit returns true if the jump at target-1 is a backward JUMP
// (the characteristic last instruction of a while-loop body).
func (d *decompiler) isWhileLoopExit(target int) bool {
	if target <= 0 || target > len(d.chunk.Code) {
		return false
	}
	prev := d.chunk.Code[target-1]
	return prev.Op == OP_JUMP && int(prev.Operand) < d.ip-1
}

// isLogicalAnd returns true when JUMP_IF_FALSE jumps to a LOAD_CONST false
// that is the false-branch of a short-circuit AND expression.
func (d *decompiler) isLogicalAnd(falseTarget int) bool {
	if falseTarget <= 0 || falseTarget >= len(d.chunk.Code) {
		return false
	}
	instr := d.chunk.Code[falseTarget]
	if instr.Op != OP_LOAD_CONST {
		return false
	}
	cv := d.chunk.Constants[instr.Operand]
	b, ok := cv.(bool)
	return ok && !b
}

// isLogicalOr returns true when JUMP_IF_TRUE jumps to a LOAD_CONST true
// that is the true-branch of a short-circuit OR expression.
func (d *decompiler) isLogicalOr(trueTarget int) bool {
	if trueTarget <= 0 || trueTarget >= len(d.chunk.Code) {
		return false
	}
	instr := d.chunk.Code[trueTarget]
	if instr.Op != OP_LOAD_CONST {
		return false
	}
	cv := d.chunk.Constants[instr.Operand]
	b, ok := cv.(bool)
	return ok && b
}

// ─── logical short-circuit decoders ──────────────────────────────────────────

// decodeLogicalAnd handles: [left]=cond, JUMP_IF_FALSE->falseTarget, [right], JUMP->end, falseTarget: LOAD_CONST false, end:
func (d *decompiler) decodeLogicalAnd(left string, falseTarget int) {
	// right expr: between current ip and the JUMP just before falseTarget
	rightEnd := falseTarget - 1 // JUMP instruction is at falseTarget-1
	d.decodeRange(rightEnd)
	right := d.pop()
	// skip JUMP -> end
	jumpInstr := d.chunk.Code[d.ip]
	endTarget := int(jumpInstr.Operand)
	d.ip++ // skip JUMP
	// skip LOAD_CONST false at falseTarget
	d.ip++ // skip LOAD_CONST false
	// end: now at endTarget
	d.ip = endTarget
	d.push("(" + left + " and " + right + ")")
}

// decodeLogicalOr handles: [left]=cond, JUMP_IF_TRUE->trueTarget, [right], JUMP->end, trueTarget: LOAD_CONST true, end:
func (d *decompiler) decodeLogicalOr(left string, trueTarget int) {
	rightEnd := trueTarget - 1
	d.decodeRange(rightEnd)
	right := d.pop()
	// skip JUMP -> end
	jumpInstr := d.chunk.Code[d.ip]
	endTarget := int(jumpInstr.Operand)
	d.ip++
	// skip LOAD_CONST true at trueTarget
	d.ip++
	d.ip = endTarget
	d.push("(" + left + " or " + right + ")")
}

// ─── if / elif / else ─────────────────────────────────────────────────────────

// decodeIf handles: cond already popped, ip is just past JUMP_IF_FALSE, target=falseTarget.
func (d *decompiler) decodeIf(cond string, falseTarget int) {
	code := d.chunk.Code
	d.emit("if " + stripParens(cond) + ":")
	d.indent++

	// Consume PUSH_SCOPE
	if d.ip < len(code) && code[d.ip].Op == OP_PUSH_SCOPE {
		d.ip++
	}

	// The then-body ends at the JUMP (if any) or POP_SCOPE just before falseTarget.
	// Check if there's a JUMP at falseTarget-1 that jumps forward (end-of-if jump).
	var endTarget int
	hasElse := false
	if falseTarget > 0 && falseTarget <= len(code) {
		prev := code[falseTarget-1]
		if prev.Op == OP_JUMP && int(prev.Operand) > falseTarget {
			hasElse = true
			endTarget = int(prev.Operand)
		}
	}

	// Then-body: up to but not including the POP_SCOPE (and optional JUMP)
	// The POP_SCOPE precedes the optional JUMP-to-end, which precedes falseTarget.
	// We need to find the POP_SCOPE position.
	thenBodyEnd := d.findMatchingPopScope(d.ip)
	thenStart := d.buf.Len()
	d.decodeRange(thenBodyEnd)

	// Check if empty body
	if d.bodyEmpty(thenStart) {
		d.emit("pass")
	}
	d.indent--

	// Consume POP_SCOPE
	if d.ip < len(code) && code[d.ip].Op == OP_POP_SCOPE {
		d.ip++
	}

	if !hasElse {
		// No else/elif – we're done
		return
	}

	// Consume JUMP -> end
	if d.ip < len(code) && code[d.ip].Op == OP_JUMP {
		d.ip++
	}

	// Now at falseTarget: could be another condition (elif) or else body (PUSH_SCOPE)
	d.decodeElseChain(endTarget)
}

// decodeElseChain decodes zero or more elif/else branches.
// d.ip is at the start of the else/elif block; endTarget is end of entire if.
func (d *decompiler) decodeElseChain(endTarget int) {
	code := d.chunk.Code
	if d.ip >= endTarget {
		return
	}

	// Is the next instruction the start of an elif condition (not PUSH_SCOPE)?
	// The elif pattern: [cond instrs]; JUMP_IF_FALSE; PUSH_SCOPE; body; POP_SCOPE; JUMP->end
	// The else pattern: PUSH_SCOPE; body; POP_SCOPE
	if code[d.ip].Op == OP_PUSH_SCOPE {
		// else branch
		d.emit("else:")
		d.indent++
		d.ip++ // consume PUSH_SCOPE
		elseBodyEnd := d.findMatchingPopScope(d.ip)
		elseStart := d.buf.Len()
		d.decodeRange(elseBodyEnd)
		if d.bodyEmpty(elseStart) {
			d.emit("pass")
		}
		d.indent--
		if d.ip < len(code) && code[d.ip].Op == OP_POP_SCOPE {
			d.ip++
		}
		return
	}

	// elif: decode the condition expression(s) to get to JUMP_IF_FALSE
	elifCondEnd := d.findNextOp(d.ip, OP_JUMP_IF_FALSE)
	if elifCondEnd < 0 || elifCondEnd >= endTarget {
		return
	}
	d.decodeRange(elifCondEnd) // decode condition exprs
	elifCond := d.pop()
	elifFalseTarget := int(code[d.ip].Operand)
	d.ip++ // consume JUMP_IF_FALSE

	// Check for another JUMP at elifFalseTarget-1
	var elifEndTarget int
	hasMore := false
	if elifFalseTarget > 0 && elifFalseTarget <= len(code) {
		prev := code[elifFalseTarget-1]
		if prev.Op == OP_JUMP && int(prev.Operand) == endTarget {
			hasMore = true
			elifEndTarget = endTarget
		}
	}

	d.emit("elif " + stripParens(elifCond) + ":")
	d.indent++
	if d.ip < len(code) && code[d.ip].Op == OP_PUSH_SCOPE {
		d.ip++
	}
	elifBodyEnd := d.findMatchingPopScope(d.ip)
	elifStart := d.buf.Len()
	d.decodeRange(elifBodyEnd)
	if d.bodyEmpty(elifStart) {
		d.emit("pass")
	}
	d.indent--
	if d.ip < len(code) && code[d.ip].Op == OP_POP_SCOPE {
		d.ip++
	}
	if hasMore {
		// consume the JUMP -> end
		if d.ip < len(code) && code[d.ip].Op == OP_JUMP {
			d.ip++
		}
		d.decodeElseChain(elifEndTarget)
	}
}

// ─── while loop ───────────────────────────────────────────────────────────────

// decodeWhileBody handles the body of a while loop.
// cond is the condition expression; exitTarget is the first instruction after
// the loop (JUMP_IF_FALSE operand).
func (d *decompiler) decodeWhileBody(cond string, exitTarget int) {
	code := d.chunk.Code
	d.emit("while " + stripParens(cond) + ":")
	d.indent++

	// Consume PUSH_SCOPE
	if d.ip < len(code) && code[d.ip].Op == OP_PUSH_SCOPE {
		d.ip++
	}

	bodyEnd := d.findMatchingPopScope(d.ip)
	whileStart := d.buf.Len()
	d.decodeRange(bodyEnd)
	if d.bodyEmpty(whileStart) {
		d.emit("pass")
	}
	d.indent--

	// Consume POP_SCOPE
	if d.ip < len(code) && code[d.ip].Op == OP_POP_SCOPE {
		d.ip++
	}
	// Consume JUMP -> loopStart
	if d.ip < len(code) && code[d.ip].Op == OP_JUMP {
		d.ip++
	}
	// d.ip should now equal exitTarget
	d.ip = exitTarget
}

// ─── for / for-each loop detection ───────────────────────────────────────────

// tryDecodeForLoop is called when we define a __hidden_ variable.
// It looks ahead to determine whether this is the start of a for or for-each
// loop and, if so, decodes the full loop and returns true.
func (d *decompiler) tryDecodeForLoop(hiddenName string, initExpr string) bool {
	code := d.chunk.Code

	// FOR-EACH pattern:
	//   DEFINE_VAR __hidden_list = listExpr
	//   LOAD_CONST 0
	//   DEFINE_VAR __hidden_idx
	//   loopStart: LOAD_VAR __hidden_idx; LOAD_VAR __hidden_list; LENGTH; BINARY_OP BinLt; JUMP_IF_FALSE
	//   PUSH_SCOPE
	//   LOAD_VAR __hidden_list; LOAD_VAR __hidden_idx; INDEX_GET; DEFINE_VAR itemName
	//   body
	//   POP_SCOPE
	//   INCREMENT; JUMP -> loopStart; loopEnd:
	if d.ip+2 < len(code) &&
		code[d.ip].Op == OP_LOAD_CONST &&
		fmtValue(d.chunk.Constants[code[d.ip].Operand]) == "0" &&
		code[d.ip+1].Op == OP_DEFINE_VAR {
		idxHidden := d.rawName(code[d.ip+1].Operand)
		if strings.HasPrefix(idxHidden, "__hidden_") {
			// Also verify the loop-start pattern: LOAD_VAR idxHidden
			if d.ip+2 < len(code) && code[d.ip+2].Op == OP_LOAD_VAR &&
				d.rawName(code[d.ip+2].Operand) == idxHidden {
				return d.decodeForEach(hiddenName, initExpr, idxHidden, d.ip+2)
			}
		}
	}

	// FOR-LOOP (repeat N times) pattern:
	//   DEFINE_VAR __hidden_counter = countExpr
	//   loopStart: LOAD_VAR __hidden_counter; LOAD_CONST 0; BINARY_OP BinGt; JUMP_IF_FALSE
	//   PUSH_SCOPE; body; POP_SCOPE
	//   LOAD_VAR __hidden_counter; LOAD_CONST 1; BINARY_OP BinSub; STORE_VAR __hidden_counter
	//   JUMP -> loopStart; loopEnd:
	if d.ip < len(code) && code[d.ip].Op == OP_LOAD_VAR &&
		d.rawName(code[d.ip].Operand) == hiddenName {
		return d.decodeRepeatN(hiddenName, initExpr, d.ip)
	}

	return false
}

// decodeForEach decodes a for-each loop.
// listHidden: name of hidden list variable; listExpr: its Python expression.
// idxHidden: name of hidden index variable; loopStart: ip of loop-start instruction.
func (d *decompiler) decodeForEach(listHidden, listExpr, idxHidden string, loopStart int) bool {
	code := d.chunk.Code
	// Consume LOAD_CONST 0 and DEFINE_VAR __hidden_idx
	d.ip += 2

	// Consume loop-start check: LOAD_VAR idx; LOAD_VAR list; LENGTH; BINARY_OP BinLt; JUMP_IF_FALSE
	// That's 5 instructions. Advance to JUMP_IF_FALSE.
	condEnd := d.ip + 4 // d.ip is at LOAD_VAR idx (index 0), +4 = JUMP_IF_FALSE
	if condEnd >= len(code) {
		return false
	}
	if code[condEnd].Op != OP_JUMP_IF_FALSE {
		return false
	}
	exitTarget := int(code[condEnd].Operand)
	d.ip = condEnd + 1 // past JUMP_IF_FALSE

	// Consume PUSH_SCOPE
	if d.ip < len(code) && code[d.ip].Op == OP_PUSH_SCOPE {
		d.ip++
	}

	// The next 4 instructions define the loop variable: LOAD_VAR list; LOAD_VAR idx; INDEX_GET; DEFINE_VAR itemName
	if d.ip+3 >= len(code) {
		return false
	}
	itemNameIdx := code[d.ip+3].Operand
	itemName := sanitizeDecompIdent(d.rawName(itemNameIdx))
	d.ip += 4 // consume those 4 instructions

	// Find the POP_SCOPE for the body
	bodyEnd := d.findMatchingPopScope(d.ip)

	d.emit("for " + itemName + " in " + listExpr + ":")
	d.indent++
	forEachStart := d.buf.Len()
	d.decodeRange(bodyEnd)
	if d.bodyEmpty(forEachStart) {
		d.emit("pass")
	}
	d.indent--

	// Consume POP_SCOPE
	if d.ip < len(code) && code[d.ip].Op == OP_POP_SCOPE {
		d.ip++
	}
	// Consume increment: LOAD_VAR idx; LOAD_CONST 1; BINARY_OP BinAdd; STORE_VAR idx
	d.ip += 4
	// Consume JUMP -> loopStart
	if d.ip < len(code) && code[d.ip].Op == OP_JUMP {
		d.ip++
	}
	d.ip = exitTarget
	return true
}

// decodeRepeatN decodes a "repeat N times" for loop.
// counterHidden: name of counter; countExpr: the Python count expression.
// loopStart: ip of the first instruction of the loop (LOAD_VAR counter).
func (d *decompiler) decodeRepeatN(counterHidden, countExpr string, loopStart int) bool {
	code := d.chunk.Code
	// Loop-start check: LOAD_VAR counter; LOAD_CONST 0; BINARY_OP BinGt; JUMP_IF_FALSE
	// That's 4 instructions. Advance to JUMP_IF_FALSE.
	jifPos := loopStart + 3
	if jifPos >= len(code) || code[jifPos].Op != OP_JUMP_IF_FALSE {
		return false
	}
	exitTarget := int(code[jifPos].Operand)
	d.ip = jifPos + 1 // past JUMP_IF_FALSE

	// Consume PUSH_SCOPE
	if d.ip < len(code) && code[d.ip].Op == OP_PUSH_SCOPE {
		d.ip++
	}

	bodyEnd := d.findMatchingPopScope(d.ip)

	d.emit("for _ in range(int(" + countExpr + ")):")
	d.indent++
	repeatStart := d.buf.Len()
	d.decodeRange(bodyEnd)
	if d.bodyEmpty(repeatStart) {
		d.emit("pass")
	}
	d.indent--

	// Consume POP_SCOPE
	if d.ip < len(code) && code[d.ip].Op == OP_POP_SCOPE {
		d.ip++
	}
	// Consume decrement: LOAD_VAR counter; LOAD_CONST 1; BINARY_OP BinSub; STORE_VAR counter
	d.ip += 4
	// Consume JUMP -> loopStart
	if d.ip < len(code) && code[d.ip].Op == OP_JUMP {
		d.ip++
	}
	d.ip = exitTarget
	return true
}

// ─── try / except ─────────────────────────────────────────────────────────────

// decodeTry decodes a try/except/finally block.
// d.ip is just past the TRY_BEGIN instruction; catchOffset is its operand.
func (d *decompiler) decodeTry(catchOffset int) {
	code := d.chunk.Code

	// TRY_END is at the end of the try body; its operand = end offset (past catch, before finally).
	// Use a depth-aware search so nested try blocks don't confuse the outer match.
	tryEndPos := d.findMatchingTryEnd(d.ip)
	if tryEndPos < 0 {
		return
	}
	endOffset := int(code[tryEndPos].Operand)

	d.emit("try:")
	d.indent++
	tryStart := d.buf.Len()
	d.decodeRange(tryEndPos) // decode try body
	if d.bodyEmpty(tryStart) {
		d.emit("pass")
	}
	d.indent--
	d.ip++ // consume TRY_END

	// We're now at catchOffset.
	// Expect: PUSH_SCOPE; CATCH(errVar<<16 | errType); [catch body]; POP_SCOPE
	if d.ip < len(code) && code[d.ip].Op == OP_PUSH_SCOPE {
		d.ip++
	}
	if d.ip < len(code) && code[d.ip].Op == OP_CATCH {
		catchInstr := code[d.ip]
		d.ip++
		errVarIdx := catchInstr.Operand >> 16
		errTypeIdx := catchInstr.Operand & 0xFFFF
		var clause string
		if errTypeIdx == 0 {
			// catch any error
			if errVarIdx > 0 {
				errVar := d.rawName(errVarIdx)
				clause = "except Exception as " + errVar + ":"
			} else {
				clause = "except Exception:"
			}
		} else {
			errType := d.rawName(errTypeIdx - 1)
			if errVarIdx > 0 {
				errVar := d.rawName(errVarIdx)
				clause = "except " + errType + " as " + errVar + ":"
			} else {
				clause = "except " + errType + ":"
			}
		}
		d.emit(clause)
		d.indent++
		// Catch body ends at POP_SCOPE before endOffset
		catchBodyEnd := d.findMatchingPopScope(d.ip)
		catchStart := d.buf.Len()
		d.decodeRange(catchBodyEnd)
		if d.bodyEmpty(catchStart) {
			d.emit("pass")
		}
		d.indent--
		if d.ip < len(code) && code[d.ip].Op == OP_POP_SCOPE {
			d.ip++
		}
	}

	// d.ip = endOffset now; any remaining instructions are finally body
	d.ip = endOffset
	if d.ip < len(code) {
		// Peek: is there a finally body?
		// There's no explicit OP_FINALLY, so just decode remaining statements
		// in the enclosing range (handled by the outer decode loop).
	}
}
