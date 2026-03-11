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
	ip        int
	// expression stack – each entry is a Python expression string
	exprStack []string
	// output buffer for the current scope level
	buf    *strings.Builder
	indent int

	// tracking which Python modules / helpers are needed
	needsMath   bool
	needsRandom bool
	needsCopy   bool
	helpers     map[string]bool // helperDef keys from transpiler/helpers.go
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
