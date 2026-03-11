package ivm

// Decompile converts an ivm Chunk directly to Python source without requiring
// the original .abc source. It reconstructs high-level control-flow structures
// (if/elif/else, while, for, for-each, try/except, functions, structs) by
// recognising the exact instruction patterns the Compiler emits.
//
// The output Python is functionally identical to what the AST-based transpiler
// would produce from the original source (modulo comment loss).

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// Decompile decompiles chunk and returns Python source code.
func Decompile(chunk *Chunk) string {
	d := newDecompiler(chunk)
	d.decode(0, len(chunk.Code))
	return d.finish()
}

// ─── decompiler struct ────────────────────────────────────────────────────────

type decompiler struct {
	// root chunk (used for struct-def and func lookups across sub-chunks)
	root *Chunk
	// current chunk being decoded
	chunk *Chunk
	// code pointer (index into chunk.Code)
	ip int
	// expression stack – each entry is a Python expression string
	exprStack []string
	// output buffer for the current scope level
	buf    *strings.Builder
	indent int

	// tracking which Python modules / helpers are needed
	needsMath   bool
	needsRandom bool
	needsCopy   bool
	helpers   map[string]bool // helperDef keys from transpiler/helpers.go
	// user-defined function names (to distinguish from stdlib)
	userFuncs map[string]bool
}

func newDecompiler(root *Chunk) *decompiler {
	var buf strings.Builder
	d := &decompiler{
		root:      root,
		chunk:     root,
		buf:       &buf,
		helpers:   make(map[string]bool),
		userFuncs: make(map[string]bool),
	}
	d.scanUserFuncs(root)
	return d
}

// scanUserFuncs pre-populates userFuncs so stdlib names don't collide.
func (d *decompiler) scanUserFuncs(chunk *Chunk) {
	for _, fc := range chunk.Funcs {
		d.userFuncs[fc.Name] = true
		d.scanUserFuncs(fc.Body)
	}
	for _, sd := range chunk.StructDefs {
		d.userFuncs[sd.Name] = true
		for _, m := range sd.Methods {
			d.userFuncs[m.Name] = true
		}
	}
}

// finish assembles the final Python file with imports and helpers.
func (d *decompiler) finish() string {
	var out strings.Builder
	out.WriteString("# Decompiled from ivm bytecode\n")

	if d.needsMath {
		out.WriteString("import math\n")
	}
	if d.needsRandom {
		out.WriteString("import random\n")
	}
	if d.needsCopy {
		out.WriteString("import copy\n")
	}
	hasMod := d.needsMath || d.needsRandom || d.needsCopy
	if hasMod && len(d.helpers) > 0 {
		out.WriteByte('\n')
	}

	// Emit helper function definitions (same set as AST transpiler uses).
	for h := range d.helpers {
		if def, ok := helperDefs[h]; ok {
			out.WriteString(def + "\n\n")
		}
	}

	body := d.buf.String()
	if len(body) > 0 {
		out.WriteString(body)
	}
	return out.String()
}

// ─── output helpers ───────────────────────────────────────────────────────────

func (d *decompiler) emit(s string) {
	if s == "" {
		return
	}
	d.buf.WriteString(strings.Repeat("    ", d.indent))
	d.buf.WriteString(s)
	d.buf.WriteByte('\n')
}

func (d *decompiler) emitLine(s string) { d.emit(s) }

// ─── expression stack ─────────────────────────────────────────────────────────

func (d *decompiler) push(s string) {
	d.exprStack = append(d.exprStack, s)
}

func (d *decompiler) pop() string {
	if len(d.exprStack) == 0 {
		return "None"
	}
	top := d.exprStack[len(d.exprStack)-1]
	d.exprStack = d.exprStack[:len(d.exprStack)-1]
	return top
}

// popN pops n items and returns them in push-order (oldest first).
func (d *decompiler) popN(n int) []string {
	if n <= 0 {
		return nil
	}
	start := len(d.exprStack) - n
	if start < 0 {
		start = 0
	}
	result := make([]string, len(d.exprStack)-start)
	copy(result, d.exprStack[start:])
	d.exprStack = d.exprStack[:start]
	return result
}

// ─── name helpers ─────────────────────────────────────────────────────────────

var mathConstantMap = map[string]string{
	"pi":       "math.pi",
	"e":        "math.e",
	"infinity": "math.inf",
}

func (d *decompiler) rawName(idx uint32) string {
	if int(idx) < len(d.chunk.Names) {
		return d.chunk.Names[idx]
	}
	return fmt.Sprintf("__n%d", idx)
}

func (d *decompiler) pyName(idx uint32) string {
	return sanitizeDecompIdent(d.rawName(idx))
}

// sanitizeDecompIdent mirrors transpiler.sanitizeIdent.
func sanitizeDecompIdent(name string) string {
	if pyKeywords[name] {
		return name + "_"
	}
	return name
}

var pyKeywords = map[string]bool{
	"False": true, "None": true, "True": true,
	"and": true, "as": true, "assert": true, "async": true, "await": true,
	"break": true, "class": true, "continue": true, "def": true, "del": true,
	"elif": true, "else": true, "except": true, "finally": true, "for": true,
	"from": true, "global": true, "if": true, "import": true, "in": true,
	"is": true, "lambda": true, "nonlocal": true, "not": true, "or": true,
	"pass": true, "raise": true, "return": true, "try": true, "type": true,
	"while": true, "with": true, "yield": true,
}

// ─── constant formatting ──────────────────────────────────────────────────────

func (d *decompiler) fmtConst(idx uint32) string {
	if int(idx) >= len(d.chunk.Constants) {
		return "None"
	}
	return fmtValue(d.chunk.Constants[idx])
}

func fmtValue(v interface{}) string {
	switch val := v.(type) {
	case float64:
		if math.IsInf(val, 1) {
			return "math.inf"
		}
		if math.IsInf(val, -1) {
			return "-math.inf"
		}
		if math.IsNaN(val) {
			return "float('nan')"
		}
		if val == math.Trunc(val) && math.Abs(val) < 1e15 {
			return strconv.FormatFloat(val, 'f', 0, 64)
		}
		return strconv.FormatFloat(val, 'g', -1, 64)
	case string:
		return fmt.Sprintf("%q", val)
	case bool:
		if val {
			return "True"
		}
		return "False"
	case nil:
		return "None"
	case []interface{}:
		parts := make([]string, len(val))
		for i, item := range val {
			parts[i] = fmtValue(item)
		}
		return "[" + strings.Join(parts, ", ") + "]"
	default:
		return fmt.Sprintf("None  # %T", v)
	}
}

// ─── main decode loop ─────────────────────────────────────────────────────────

// decode processes chunk.Code[start:end] (end is exclusive).
// d.chunk must be set before calling.
func (d *decompiler) decode(start, end int) {
	d.ip = start
	code := d.chunk.Code
	for d.ip < end {
		instr := code[d.ip]
		d.ip++ // advance; control-flow handlers may advance further
		d.processInstr(instr)
	}
}

// decodeRange processes d.chunk.Code[d.ip:end] updating d.ip.
func (d *decompiler) decodeRange(end int) {
	code := d.chunk.Code
	for d.ip < end {
		instr := code[d.ip]
		d.ip++
		d.processInstr(instr)
	}
}

func (d *decompiler) processInstr(instr Instruction) { //nolint:gocyclo
	op := instr.Op
	operand := instr.Operand
	code := d.chunk.Code

	switch op {

	// ── Meta ──────────────────────────────────────────────────────────────────
	case OP_SET_LINE:
		// no output

	// ── Constants / values ───────────────────────────────────────────────────
	case OP_LOAD_CONST:
		cv := d.fmtConst(operand)
		// Infinity needs math import
		if strings.Contains(cv, "math.inf") || strings.Contains(cv, "math.nan") {
			d.needsMath = true
		}
		d.push(cv)

	case OP_LOAD_NOTHING:
		d.push("None")

	case OP_LOAD_VAR:
		raw := d.rawName(operand)
		if py, ok := mathConstantMap[raw]; ok {
			d.needsMath = true
			d.push(py)
		} else {
			d.push(sanitizeDecompIdent(raw))
		}

	// ── Variable definition ───────────────────────────────────────────────────
	case OP_DEFINE_VAR, OP_DEFINE_CONST:
		val := d.pop()
		raw := d.rawName(operand)
		if strings.HasPrefix(raw, "__hidden_") {
			// May be the start of a for/for-each loop.
			if handled := d.tryDecodeForLoop(raw, val); handled {
				return
			}
		}
		d.emit(sanitizeDecompIdent(raw) + " = " + val)

	case OP_DEFINE_TYPED, OP_DEFINE_TYPED_CONST:
		val := d.pop()
		_ = d.pop() // type name string – Python is dynamically typed
		raw := d.rawName(operand)
		d.emit(sanitizeDecompIdent(raw) + " = " + val)

	// ── Assignment ────────────────────────────────────────────────────────────
	case OP_STORE_VAR:
		val := d.pop()
		d.emit(d.pyName(operand) + " = " + val)

	case OP_TOGGLE_VAR:
		name := d.pyName(operand)
		d.emit(name + " = not " + name)

	case OP_SWAP_VARS:
		n1 := d.pyName(operand >> 16)
		n2 := d.pyName(operand & 0xFFFF)
		d.emit(n1 + ", " + n2 + " = " + n2 + ", " + n1)

	// ── Arithmetic / comparison ───────────────────────────────────────────────
	case OP_BINARY_OP:
		right := d.pop()
		left := d.pop()
		d.push(d.fmtBinOp(left, BinOp(operand), right))

	case OP_UNARY_OP:
		operand2 := d.pop()
		switch UnaryOp(operand) {
		case UnaryNeg:
			d.push("-" + operand2)
		case UnaryNot:
			d.push("not " + operand2)
		}

	// ── Control flow ──────────────────────────────────────────────────────────
	case OP_JUMP_IF_FALSE:
		cond := d.pop()
		target := int(operand)
		// Classify: while-loop, logical AND, or if statement
		if d.isWhileLoopExit(target) {
			d.decodeWhileBody(cond, target)
		} else if d.isLogicalAnd(target) {
			// Short-circuit AND: consume right side and end
			d.decodeLogicalAnd(cond, target)
		} else {
			d.decodeIf(cond, target)
		}

	case OP_JUMP_IF_TRUE:
		cond := d.pop()
		target := int(operand)
		if d.isLogicalOr(target) {
			d.decodeLogicalOr(cond, target)
		}
		// else: unusual pattern – discard

	case OP_JUMP:
		// Standalone JUMP in the middle of a range = break/continue.
		// Emit Python break/continue based on direction.
		// (Jumps that are part of if/while/for structure are consumed by the
		// structure handlers and never seen here.)
		target := int(operand)
		if target < d.ip-1 {
			d.emit("continue")
		} else {
			d.emit("break")
		}

	// Scope delimiters are consumed by structural handlers; if we see one
	// unexpectedly, just skip it.
	case OP_PUSH_SCOPE, OP_POP_SCOPE:
		// handled structurally

	// ── Functions ─────────────────────────────────────────────────────────────
	case OP_DEFINE_FUNC:
		if int(operand) < len(d.chunk.Funcs) {
			d.decodeFunc(d.chunk.Funcs[operand])
		}

	case OP_CALL:
		argc := operand >> 16
		nameIdx := operand & 0xFFFF
		args := d.popN(int(argc))
		funcName := d.rawName(nameIdx)
		d.push(d.fmtFuncCall(funcName, args))

	case OP_CALL_METHOD:
		argc := operand >> 16
		methIdx := operand & 0xFFFF
		args := d.popN(int(argc))
		obj := d.pop()
		meth := d.rawName(methIdx)
		argStr := strings.Join(args, ", ")
		d.push(fmt.Sprintf("%s.%s(%s)", obj, meth, argStr))

	case OP_RETURN:
		val := d.pop()
		if val == "None" {
			d.emit("return")
		} else {
			d.emit("return " + val)
		}

	// ── Output ────────────────────────────────────────────────────────────────
	case OP_PRINT:
		count := operand >> 1
		newline := (operand & 1) == 1
		args := d.popN(int(count))
		if newline {
			if len(args) == 0 {
				d.emit("print()")
			} else {
				d.emit("print(" + strings.Join(args, ", ") + ")")
			}
		} else {
			if len(args) == 0 {
				d.emit("print(end='')")
			} else {
				d.emit("print(" + strings.Join(args, ", ") + ", end='')")
			}
		}

	// ── Stack management ──────────────────────────────────────────────────────
	case OP_POP:
		if len(d.exprStack) > 0 {
			expr := d.pop()
			// Only emit expression-statements that have observable side-effects.
			if strings.Contains(expr, "(") {
				d.emit(expr)
			}
		}

	// ── Collections ───────────────────────────────────────────────────────────
	case OP_BUILD_LIST:
		elems := d.popN(int(operand))
		d.push("[" + strings.Join(elems, ", ") + "]")

	case OP_BUILD_ARRAY:
		count := operand
		_ = d.pop() // type name
		elems := d.popN(int(count))
		d.push("[" + strings.Join(elems, ", ") + "]")

	case OP_BUILD_LOOKUP:
		d.push("{}")

	case OP_INDEX_GET:
		idx := d.pop()
		list := d.pop()
		// ivm uses 0-based indexing internally
		d.push(fmt.Sprintf("%s[%s]", list, idx))

	case OP_INDEX_SET:
		val := d.pop()
		idx := d.pop()
		listName := d.pyName(operand)
		d.emit(fmt.Sprintf("%s[%s] = %s", listName, idx, val))

	case OP_LENGTH:
		val := d.pop()
		d.push(fmt.Sprintf("len(%s)", val))

	// ── Lookup table ──────────────────────────────────────────────────────────
	case OP_LOOKUP_GET:
		key := d.pop()
		table := d.pop()
		d.push(fmt.Sprintf("%s[%s]", table, key))

	case OP_LOOKUP_SET:
		val := d.pop()
		key := d.pop()
		tableName := d.pyName(operand)
		d.emit(fmt.Sprintf("%s[%s] = %s", tableName, key, val))

	case OP_LOOKUP_HAS:
		key := d.pop()
		table := d.pop()
		d.push(fmt.Sprintf("(%s in %s)", key, table))

	// ── Type operations ───────────────────────────────────────────────────────
	case OP_TYPEOF:
		val := d.pop()
		d.push(fmt.Sprintf("type(%s).__name__", val))

	case OP_CAST:
		val := d.pop()
		typeName := d.rawName(operand)
		d.push(d.fmtCast(typeName, val))

	case OP_NIL_CHECK:
		val := d.pop()
		if operand == 1 {
			d.push(fmt.Sprintf("(%s is not None)", val))
		} else {
			d.push(fmt.Sprintf("(%s is None)", val))
		}

	case OP_ERROR_TYPE_CHECK:
		val := d.pop()
		typeName := d.pyName(operand)
		d.push(fmt.Sprintf("isinstance(%s, %s)", val, typeName))

	// ── Input ─────────────────────────────────────────────────────────────────
	case OP_ASK:
		if operand == 1 {
			prompt := d.pop()
			d.push(fmt.Sprintf("input(%s)", prompt))
		} else {
			d.push("input()")
		}

	// ── Location ──────────────────────────────────────────────────────────────
	case OP_LOCATION:
		name := d.pyName(operand)
		d.push(fmt.Sprintf("id(%s)", name))

	// ── Structs ───────────────────────────────────────────────────────────────
	case OP_DEFINE_STRUCT:
		if int(operand) < len(d.chunk.StructDefs) {
			d.decodeStruct(d.chunk.StructDefs[operand])
		}

	case OP_NEW_STRUCT:
		fieldCount := operand >> 16
		snIdx := operand & 0xFFFF
		structName := d.rawName(snIdx)
		fieldVals := d.popN(int(fieldCount))
		// Look up field names from the struct def
		sd := d.findStructDef(structName)
		var parts []string
		if sd != nil && len(sd.Fields) == len(fieldVals) {
			for i, fv := range fieldVals {
				parts = append(parts, sd.Fields[i].Name+"="+fv)
			}
		} else {
			parts = fieldVals
		}
		d.push(fmt.Sprintf("%s(%s)", structName, strings.Join(parts, ", ")))

	case OP_GET_FIELD:
		obj := d.pop()
		field := d.pyName(operand)
		d.push(fmt.Sprintf("%s.%s", obj, field))

	case OP_SET_FIELD:
		val := d.pop()
		obj := d.pop()
		field := d.pyName(operand)
		d.emit(fmt.Sprintf("%s.%s = %s", obj, field, val))

	// ── Error handling ────────────────────────────────────────────────────────
	case OP_RAISE:
		msg := d.pop()
		if operand == 0 {
			d.emit(fmt.Sprintf("raise Exception(%s)", msg))
		} else {
			typeName := d.pyName(operand - 1)
			d.emit(fmt.Sprintf("raise %s(%s)", typeName, msg))
		}

	case OP_TRY_BEGIN:
		d.decodeTry(int(operand))

	case OP_TRY_END, OP_CATCH:
		// consumed by decodeTry

	case OP_DEFINE_ERROR_TYPE:
		nameIdx := operand >> 16
		parentIdx := operand & 0xFFFF
		typeName := d.rawName(nameIdx)
		var parent string
		if parentIdx == 0 {
			parent = "Exception"
		} else {
			parent = d.rawName(parentIdx - 1)
		}
		d.emit(fmt.Sprintf("class %s(%s): pass", typeName, parent))

	// ── Reference / copy ──────────────────────────────────────────────────────
	case OP_MAKE_REFERENCE:
		name := d.pyName(operand)
		d.push(name) // Python names already are references

	case OP_MAKE_COPY:
		val := d.pop()
		d.needsCopy = true
		d.push(fmt.Sprintf("copy.deepcopy(%s)", val))

	// ── Import ────────────────────────────────────────────────────────────────
	case OP_IMPORT:
		flags := operand
		hasItems := (flags & 1) != 0
		isSafe := (flags & 2) != 0
		_ = isSafe
		importAll := (flags & 4) != 0

		var items []string
		if hasItems {
			raw := d.pop()
			// raw is a Python list literal like ["a", "b"]
			items = extractListLiteral(raw)
		}
		pathExpr := d.pop()
		// Strip surrounding quotes
		path := strings.Trim(pathExpr, "\"")
		moduleName := pathToModuleName(path)

		if importAll || len(items) == 0 {
			d.emit(fmt.Sprintf("from %s import *", moduleName))
		} else {
			d.emit(fmt.Sprintf("from %s import %s", moduleName, strings.Join(items, ", ")))
		}

	// ── Normally-never-seen-standalone ────────────────────────────────────────
	default:
		// emit a comment so the output is still valid Python
		d.emit(fmt.Sprintf("pass  # unhandled opcode %s", OpName(op)))
	}

	_ = code // suppress unused warning (code is used via d.chunk.Code in helpers)
}

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
	d.push(fmt.Sprintf("(%s and %s)", left, right))
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
	d.push(fmt.Sprintf("(%s or %s)", left, right))
}

// ─── if / elif / else ─────────────────────────────────────────────────────────

// decodeIf handles: cond already popped, ip is just past JUMP_IF_FALSE, target=falseTarget.
func (d *decompiler) decodeIf(cond string, falseTarget int) {
	code := d.chunk.Code
	d.emit("if " + cond + ":")
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
	if d.indent > 0 && d.bodyEmpty(thenStart) {
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

	d.emit("elif " + elifCond + ":")
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
	d.emit("while " + cond + ":")
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

	d.emit(fmt.Sprintf("for %s in %s:", itemName, listExpr))
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

	d.emit(fmt.Sprintf("for _ in range(int(%s)):", countExpr))
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
				clause = fmt.Sprintf("except Exception as %s:", errVar)
			} else {
				clause = "except Exception:"
			}
		} else {
			errType := d.rawName(errTypeIdx - 1)
			if errVarIdx > 0 {
				errVar := d.rawName(errVarIdx)
				clause = fmt.Sprintf("except %s as %s:", errType, errVar)
			} else {
				clause = fmt.Sprintf("except %s:", errType)
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

// ─── functions ────────────────────────────────────────────────────────────────

func (d *decompiler) decodeFunc(fc *FuncChunk) {
	params := make([]string, len(fc.Params))
	for i, p := range fc.Params {
		params[i] = sanitizeDecompIdent(p)
	}
	d.emit(fmt.Sprintf("def %s(%s):", sanitizeDecompIdent(fc.Name), strings.Join(params, ", ")))
	d.indent++

	saved := d.chunk
	savedIP := d.ip
	savedStack := d.exprStack
	d.chunk = fc.Body
	d.exprStack = nil

	// Function body ends just before the implicit return (last 2 instructions:
	// OP_LOAD_NOTHING; OP_RETURN) emitted by compileFuncBody.
	bodyLen := len(fc.Body.Code)
	bodyEnd := bodyLen
	if bodyLen >= 2 &&
		fc.Body.Code[bodyLen-2].Op == OP_LOAD_NOTHING &&
		fc.Body.Code[bodyLen-1].Op == OP_RETURN {
		bodyEnd = bodyLen - 2
	}

	funcBodyStart := d.buf.Len()
	d.decode(0, bodyEnd)

	// If no code emitted (only empty scope lines), emit pass
	if d.bodyEmpty(funcBodyStart) {
		d.emit("pass")
	}
	d.indent--

	d.chunk = saved
	d.ip = savedIP
	d.exprStack = savedStack
}

// ─── structs ──────────────────────────────────────────────────────────────────

func (d *decompiler) decodeStruct(sd *StructDef) {
	d.emit(fmt.Sprintf("class %s:", sd.Name))
	d.indent++
	d.emit("def __init__(self, " + d.structInitParams(sd) + "):")
	d.indent++
	if len(sd.Fields) == 0 {
		d.emit("pass")
	} else {
		for _, f := range sd.Fields {
			// self.field = field  (parameter carries the default from structInitParams)
			d.emit(fmt.Sprintf("self.%s = %s", f.Name, sanitizeDecompIdent(f.Name)))
		}
	}
	d.indent--

	// Methods
	for _, m := range sd.Methods {
		d.decodeMethod(m)
	}
	d.indent--
}

func (d *decompiler) structInitParams(sd *StructDef) string {
	parts := make([]string, len(sd.Fields))
	for i, f := range sd.Fields {
		var defExpr string
		if f.DefaultExprChunk != nil {
			defExpr = d.evalDefaultExpr(f.DefaultExprChunk)
			parts[i] = fmt.Sprintf("%s=%s", sanitizeDecompIdent(f.Name), defExpr)
		} else {
			parts[i] = sanitizeDecompIdent(f.Name)
		}
	}
	return strings.Join(parts, ", ")
}

func (d *decompiler) decodeMethod(fc *FuncChunk) {
	params := make([]string, len(fc.Params)+1)
	params[0] = "self"
	for i, p := range fc.Params {
		params[i+1] = sanitizeDecompIdent(p)
	}
	d.emit(fmt.Sprintf("def %s(%s):", sanitizeDecompIdent(fc.Name), strings.Join(params, ", ")))
	d.indent++

	saved := d.chunk
	savedIP := d.ip
	savedStack := d.exprStack
	d.chunk = fc.Body
	d.exprStack = nil

	bodyLen := len(fc.Body.Code)
	bodyEnd := bodyLen
	if bodyLen >= 2 &&
		fc.Body.Code[bodyLen-2].Op == OP_LOAD_NOTHING &&
		fc.Body.Code[bodyLen-1].Op == OP_RETURN {
		bodyEnd = bodyLen - 2
	}

	methBodyStart := d.buf.Len()
	d.decode(0, bodyEnd)
	if d.bodyEmpty(methBodyStart) {
		d.emit("pass")
	}
	d.indent--

	d.chunk = saved
	d.ip = savedIP
	d.exprStack = savedStack
}

// evalDefaultExpr evaluates a simple default-value chunk to a Python literal.
func (d *decompiler) evalDefaultExpr(chunk *Chunk) string {
	if len(chunk.Code) == 0 {
		return "None"
	}
	saved := d.chunk
	savedIP := d.ip // must save/restore d.ip: d.decode resets it to 0
	savedStack := d.exprStack
	d.chunk = chunk
	d.exprStack = nil
	// Decode up to the RETURN at the end
	end := len(chunk.Code)
	if end > 0 && chunk.Code[end-1].Op == OP_RETURN {
		end--
	}
	d.decode(0, end)
	var result string
	if len(d.exprStack) > 0 {
		result = d.pop()
	} else {
		result = "None"
	}
	d.chunk = saved
	d.ip = savedIP
	d.exprStack = savedStack
	return result
}

// ─── helper utilities ─────────────────────────────────────────────────────────

// findMatchingPopScope scans forward from start and returns the position of the
// first OP_POP_SCOPE that closes the scope opened by the PUSH_SCOPE that
// precedes start (i.e. the direct matching POP_SCOPE, accounting for nesting).
func (d *decompiler) findMatchingPopScope(start int) int {
	code := d.chunk.Code
	depth := 0
	for i := start; i < len(code); i++ {
		switch code[i].Op {
		case OP_PUSH_SCOPE:
			depth++
		case OP_POP_SCOPE:
			if depth == 0 {
				return i
			}
			depth--
		}
	}
	return len(code)
}

// findMatchingTryEnd returns the position of the TRY_END that matches the
// TRY_BEGIN that preceded position start. It accounts for nesting so that
// inner try blocks (which also contain TRY_END) don't confuse the search.
func (d *decompiler) findMatchingTryEnd(start int) int {
	code := d.chunk.Code
	depth := 0
	for i := start; i < len(code); i++ {
		switch code[i].Op {
		case OP_TRY_BEGIN:
			depth++
		case OP_TRY_END:
			if depth == 0 {
				return i
			}
			depth--
		}
	}
	return -1
}

// findNextOp returns the position of the next instruction with the given opcode
// starting from pos, or -1 if not found.
func (d *decompiler) findNextOp(pos int, op Opcode) int {
	code := d.chunk.Code
	for i := pos; i < len(code); i++ {
		if code[i].Op == op {
			return i
		}
	}
	return -1
}

// findStructDef searches root and all sub-chunks for a struct def with the given name.
func (d *decompiler) findStructDef(name string) *StructDef {
	return findStructInChunk(d.root, name)
}

func findStructInChunk(chunk *Chunk, name string) *StructDef {
	for _, sd := range chunk.StructDefs {
		if sd.Name == name {
			return sd
		}
	}
	for _, fc := range chunk.Funcs {
		if sd := findStructInChunk(fc.Body, name); sd != nil {
			return sd
		}
	}
	return nil
}

// bodyEmpty returns true if nothing was written to the output buffer since
// bodyStart (the value of d.buf.Len() captured before entering the body).
// This is an O(1) check — no string scanning required.
func (d *decompiler) bodyEmpty(bodyStart int) bool {
	return d.buf.Len() == bodyStart
}

// ─── formatting helpers ───────────────────────────────────────────────────────

func (d *decompiler) fmtBinOp(left string, op BinOp, right string) string {
	switch op {
	case BinAdd:
		return fmt.Sprintf("(%s + %s)", left, right)
	case BinSub:
		return fmt.Sprintf("(%s - %s)", left, right)
	case BinMul:
		return fmt.Sprintf("(%s * %s)", left, right)
	case BinDiv:
		return fmt.Sprintf("(%s / %s)", left, right)
	case BinMod:
		return fmt.Sprintf("(%s %% %s)", left, right)
	case BinEq:
		return fmt.Sprintf("(%s == %s)", left, right)
	case BinNeq:
		return fmt.Sprintf("(%s != %s)", left, right)
	case BinLt:
		return fmt.Sprintf("(%s < %s)", left, right)
	case BinLte:
		return fmt.Sprintf("(%s <= %s)", left, right)
	case BinGt:
		return fmt.Sprintf("(%s > %s)", left, right)
	case BinGte:
		return fmt.Sprintf("(%s >= %s)", left, right)
	default:
		return fmt.Sprintf("(%s ? %s)", left, right)
	}
}

func (d *decompiler) fmtCast(typeName, val string) string {
	switch strings.ToLower(typeName) {
	case "number", "float":
		return fmt.Sprintf("float(%s)", val)
	case "integer", "int":
		return fmt.Sprintf("int(%s)", val)
	case "text", "string", "str":
		return fmt.Sprintf("str(%s)", val)
	case "boolean", "bool":
		return fmt.Sprintf("bool(%s)", val)
	default:
		return fmt.Sprintf("%s(%s)", typeName, val)
	}
}

// fmtFuncCall maps English stdlib function names to Python equivalents,
// mirroring the AST transpiler's stdlib.go mapping.
func (d *decompiler) fmtFuncCall(name string, args []string) string {
	a := func(i int) string {
		if i < len(args) {
			return args[i]
		}
		return "None"
	}
	joined := strings.Join(args, ", ")

	if d.userFuncs[name] {
		return fmt.Sprintf("%s(%s)", sanitizeDecompIdent(name), joined)
	}

	switch name {
	// Math
	case "sqrt":
		d.needsMath = true
		return fmt.Sprintf("math.sqrt(%s)", a(0))
	case "pow":
		d.needsMath = true
		return fmt.Sprintf("math.pow(%s, %s)", a(0), a(1))
	case "abs":
		return fmt.Sprintf("abs(%s)", a(0))
	case "floor":
		d.needsMath = true
		return fmt.Sprintf("math.floor(%s)", a(0))
	case "ceil":
		d.needsMath = true
		return fmt.Sprintf("math.ceil(%s)", a(0))
	case "round":
		return fmt.Sprintf("round(%s)", a(0))
	case "min":
		return fmt.Sprintf("min(%s, %s)", a(0), a(1))
	case "max":
		return fmt.Sprintf("max(%s, %s)", a(0), a(1))
	case "sin":
		d.needsMath = true
		return fmt.Sprintf("math.sin(%s)", a(0))
	case "cos":
		d.needsMath = true
		return fmt.Sprintf("math.cos(%s)", a(0))
	case "tan":
		d.needsMath = true
		return fmt.Sprintf("math.tan(%s)", a(0))
	case "log":
		d.needsMath = true
		return fmt.Sprintf("math.log(%s)", a(0))
	case "log10":
		d.needsMath = true
		return fmt.Sprintf("math.log10(%s)", a(0))
	case "log2":
		d.needsMath = true
		return fmt.Sprintf("math.log2(%s)", a(0))
	case "exp":
		d.needsMath = true
		return fmt.Sprintf("math.exp(%s)", a(0))
	case "random":
		d.needsRandom = true
		return "random.random()"
	case "random_between":
		d.needsRandom = true
		return fmt.Sprintf("random.uniform(%s, %s)", a(0), a(1))
	case "is_nan":
		d.helpers["_is_nan"] = true
		d.needsMath = true
		return fmt.Sprintf("_is_nan(%s)", a(0))
	case "is_infinite":
		d.helpers["_is_infinite"] = true
		d.needsMath = true
		return fmt.Sprintf("_is_infinite(%s)", a(0))
	case "is_integer":
		return fmt.Sprintf("float(%s).is_integer()", a(0))
	case "clamp":
		return fmt.Sprintf("max(%s, min(%s, %s))", a(1), a(2), a(0))
	case "sign":
		d.helpers["_sign"] = true
		return fmt.Sprintf("_sign(%s)", a(0))
	// String
	case "uppercase":
		return fmt.Sprintf("%s.upper()", a(0))
	case "lowercase":
		return fmt.Sprintf("%s.lower()", a(0))
	case "casefold":
		return fmt.Sprintf("%s.casefold()", a(0))
	case "title":
		return fmt.Sprintf("%s.title()", a(0))
	case "capitalize":
		return fmt.Sprintf("%s.capitalize()", a(0))
	case "swapcase":
		return fmt.Sprintf("%s.swapcase()", a(0))
	case "split":
		return fmt.Sprintf("%s.split(%s)", a(0), a(1))
	case "join":
		return fmt.Sprintf("%s.join(%s)", a(1), a(0))
	case "trim":
		return fmt.Sprintf("%s.strip()", a(0))
	case "trim_left":
		return fmt.Sprintf("%s.lstrip()", a(0))
	case "trim_right":
		return fmt.Sprintf("%s.rstrip()", a(0))
	case "replace":
		return fmt.Sprintf("%s.replace(%s, %s)", a(0), a(1), a(2))
	case "contains":
		return fmt.Sprintf("(%s in %s)", a(1), a(0))
	case "starts_with":
		return fmt.Sprintf("%s.startswith(%s)", a(0), a(1))
	case "ends_with":
		return fmt.Sprintf("%s.endswith(%s)", a(0), a(1))
	case "index_of":
		return fmt.Sprintf("%s.find(%s)", a(0), a(1))
	case "substring":
		start := a(1)
		length := a(2)
		return fmt.Sprintf("%s[int(%s):int(%s)+int(%s)]", a(0), start, start, length)
	case "str_repeat":
		return fmt.Sprintf("%s * int(%s)", a(0), a(1))
	case "count_occurrences":
		return fmt.Sprintf("%s.count(%s)", a(0), a(1))
	case "pad_left":
		if len(args) > 2 {
			return fmt.Sprintf("%s.rjust(int(%s), %s)", a(0), a(1), a(2))
		}
		return fmt.Sprintf("%s.rjust(int(%s))", a(0), a(1))
	case "pad_right":
		if len(args) > 2 {
			return fmt.Sprintf("%s.ljust(int(%s), %s)", a(0), a(1), a(2))
		}
		return fmt.Sprintf("%s.ljust(int(%s))", a(0), a(1))
	case "center":
		if len(args) > 2 {
			return fmt.Sprintf("%s.center(int(%s), %s)", a(0), a(1), a(2))
		}
		return fmt.Sprintf("%s.center(int(%s))", a(0), a(1))
	case "zfill":
		return fmt.Sprintf("%s.zfill(int(%s))", a(0), a(1))
	case "to_number":
		return fmt.Sprintf("float(%s)", a(0))
	case "to_string":
		return fmt.Sprintf("str(%s)", a(0))
	case "is_empty":
		return fmt.Sprintf("(len(%s) == 0)", a(0))
	case "is_digit":
		return fmt.Sprintf("%s.isdigit()", a(0))
	case "is_alpha":
		return fmt.Sprintf("%s.isalpha()", a(0))
	case "is_alnum":
		return fmt.Sprintf("%s.isalnum()", a(0))
	case "is_space":
		return fmt.Sprintf("%s.isspace()", a(0))
	case "is_upper":
		return fmt.Sprintf("%s.isupper()", a(0))
	case "is_lower":
		return fmt.Sprintf("%s.islower()", a(0))
	// List
	case "count":
		return fmt.Sprintf("len(%s)", a(0))
	case "first":
		return fmt.Sprintf("%s[0]", a(0))
	case "last":
		return fmt.Sprintf("%s[-1]", a(0))
	case "sort":
		return fmt.Sprintf("sorted(%s)", a(0))
	case "sorted_desc":
		return fmt.Sprintf("sorted(%s, reverse=True)", a(0))
	case "reverse":
		return fmt.Sprintf("list(reversed(%s))", a(0))
	case "append":
		return fmt.Sprintf("%s + [%s]", a(0), a(1))
	case "pop":
		return fmt.Sprintf("%s[:-1]", a(0))
	case "remove":
		return fmt.Sprintf("[v for i, v in enumerate(%s) if i != int(%s)]", a(0), a(1))
	case "insert":
		return fmt.Sprintf("(%s[:int(%s)] + [%s] + %s[int(%s):])", a(0), a(1), a(2), a(0), a(1))
	case "sum":
		return fmt.Sprintf("sum(%s)", a(0))
	case "product":
		d.helpers["_product"] = true
		return fmt.Sprintf("_product(%s)", a(0))
	case "average":
		return fmt.Sprintf("(sum(%s) / len(%s))", a(0), a(0))
	case "min_value":
		return fmt.Sprintf("min(%s)", a(0))
	case "max_value":
		return fmt.Sprintf("max(%s)", a(0))
	case "any_true":
		return fmt.Sprintf("any(%s)", a(0))
	case "all_true":
		return fmt.Sprintf("all(%s)", a(0))
	case "unique":
		d.helpers["_unique"] = true
		return fmt.Sprintf("_unique(%s)", a(0))
	case "flatten":
		d.helpers["_flatten"] = true
		return fmt.Sprintf("_flatten(%s)", a(0))
	case "slice":
		return fmt.Sprintf("%s[int(%s):int(%s)]", a(0), a(1), a(2))
	case "zip_with":
		d.helpers["_zip_with"] = true
		return fmt.Sprintf("_zip_with(%s, %s)", a(0), a(1))
	// Lookup table
	case "keys":
		return fmt.Sprintf("list(%s.keys())", a(0))
	case "values":
		return fmt.Sprintf("list(%s.values())", a(0))
	case "table_remove":
		d.helpers["_table_remove"] = true
		return fmt.Sprintf("_table_remove(%s, %s)", a(0), a(1))
	case "table_has":
		return fmt.Sprintf("(%s in %s)", a(1), a(0))
	case "merge":
		return fmt.Sprintf("{**%s, **%s}", a(0), a(1))
	case "get_or_default":
		return fmt.Sprintf("%s.get(%s, %s)", a(0), a(1), a(2))
	// I/O
	case "ask":
		return fmt.Sprintf("input(%s)", a(0))
	case "read_file":
		d.helpers["_read_file"] = true
		return fmt.Sprintf("_read_file(%s)", a(0))
	case "write_file":
		d.helpers["_write_file"] = true
		return fmt.Sprintf("_write_file(%s, %s)", a(0), a(1))
	}
	return fmt.Sprintf("%s(%s)", sanitizeDecompIdent(name), joined)
}

// ─── import path utilities ────────────────────────────────────────────────────

func pathToModuleName(path string) string {
	// "math_utils.abc" → "math_utils"
	// "math_utils" → "math_utils"
	base := path
	if idx := strings.LastIndex(base, "/"); idx >= 0 {
		base = base[idx+1:]
	}
	if dot := strings.LastIndex(base, "."); dot >= 0 {
		base = base[:dot]
	}
	return base
}

// extractListLiteral parses a Python list-literal string like ["a", "b"] and
// returns the items.
func extractListLiteral(s string) []string {
	s = strings.TrimSpace(s)
	// Need at least "[" and "]"
	if len(s) < 2 || s[0] != '[' || s[len(s)-1] != ']' {
		return nil
	}
	inner := s[1 : len(s)-1]
	if strings.TrimSpace(inner) == "" {
		return nil
	}
	var items []string
	for _, part := range strings.Split(inner, ", ") {
		part = strings.TrimSpace(part)
		part = strings.Trim(part, "\"")
		if part != "" {
			items = append(items, part)
		}
	}
	return items
}

// helperDefs mirrors the definitions in transpiler/helpers.go so the decompiler
// can inject the same helper functions when needed.
var helperDefs = map[string]string{
	"_table_remove": `def _table_remove(d, k):
    result = dict(d)
    result.pop(k, None)
    return result`,

	"_flatten": `def _flatten(lst):
    return [item for sublist in lst for item in sublist]`,

	"_read_file": `def _read_file(path):
    with open(path, "r") as f:
        return f.read()`,

	"_write_file": `def _write_file(path, content):
    with open(path, "w") as f:
        f.write(str(content))`,

	"_is_nan": `def _is_nan(x):
    try:
        return math.isnan(float(x))
    except (TypeError, ValueError):
        return True`,

	"_is_infinite": `def _is_infinite(x):
    try:
        return math.isinf(float(x))
    except (TypeError, ValueError):
        return False`,

	"_sign": `def _sign(x):
    if x > 0:
        return 1
    elif x < 0:
        return -1
    return 0`,

	"_unique": `def _unique(lst):
    seen = []
    for item in lst:
        if item not in seen:
            seen.append(item)
    return seen`,

	"_product": `def _product(lst):
    result = 1
    for item in lst:
        result *= item
    return result`,

	"_zip_with": `def _zip_with(a, b):
    return [[x, y] for x, y in zip(a, b)]`,
}
