package ivm

import (
	"fmt"
	"strings"
)

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
