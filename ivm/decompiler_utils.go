package ivm

import (
	"fmt"
	"strings"
)

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
