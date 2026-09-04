package ivm

import "fmt"

// Validate checks that every instruction in the chunk refers to entries that
// actually exist in its constant pools, recursing into nested function and
// method chunks.
//
// The machine indexes chunk.Names, chunk.Constants, chunk.Funcs and
// chunk.StructDefs directly with an instruction's operand. A truncated,
// hand-edited or version-mismatched .101 file could therefore drive the VM into
// an out-of-range index and panic mid-execution, with no recover anywhere to
// contain it. Validating once at load time turns that into an ordinary decode
// error, before any user code runs.
func (c *Chunk) Validate() error {
	return c.validate(make(map[*Chunk]bool))
}

func (c *Chunk) validate(seen map[*Chunk]bool) error {
	if c == nil || seen[c] {
		return nil
	}
	seen[c] = true

	nNames := uint32(len(c.Names))
	nConsts := uint32(len(c.Constants))
	nFuncs := uint32(len(c.Funcs))
	nStructs := uint32(len(c.StructDefs))

	// checkIdx validates one pool index.
	checkIdx := func(pc int, op Opcode, idx, size uint32, pool string) error {
		if idx >= size {
			return fmt.Errorf(
				"ivm: corrupt bytecode: instruction %d (%s) refers to %s entry %d, but only %d exist",
				pc, OpName(op), pool, idx, size)
		}
		return nil
	}

	for pc, ins := range c.Code {
		op, operand := ins.Op, ins.Operand

		switch op {
		// ── Single name index ──────────────────────────────────────────────
		case OP_LOAD_VAR, OP_STORE_VAR, OP_DEFINE_VAR, OP_DEFINE_CONST,
			OP_DEFINE_TYPED, OP_DEFINE_TYPED_CONST, OP_TOGGLE_VAR,
			OP_GET_FIELD, OP_SET_FIELD, OP_INDEX_SET, OP_LOOKUP_SET,
			OP_MAKE_REFERENCE, OP_LOCATION, OP_CAST, OP_RAISE,
			OP_ERROR_TYPE_CHECK, OP_CATCH:
			if err := checkIdx(pc, op, operand, nNames, "name"); err != nil {
				return err
			}

		// ── Name index, offset by one so that 0 means "absent" ─────────────
		case OP_TRY_SET_ERRORTYPE:
			if operand > 0 {
				if err := checkIdx(pc, op, operand-1, nNames, "name"); err != nil {
					return err
				}
			}

		// ── Two packed indices: high 16 bits and low 16 bits ───────────────
		case OP_DEFINE_ERROR_TYPE, OP_SWAP_VARS:
			if err := checkIdx(pc, op, operand>>16, nNames, "name"); err != nil {
				return err
			}
			if err := checkIdx(pc, op, operand&0xFFFF, nNames, "name"); err != nil {
				return err
			}

		// ── Count in the high 16 bits, name index in the low 16 ────────────
		case OP_CALL, OP_CALL_METHOD, OP_NEW_STRUCT:
			if err := checkIdx(pc, op, operand&0xFFFF, nNames, "name"); err != nil {
				return err
			}

		// ── Other pools ────────────────────────────────────────────────────
		case OP_LOAD_CONST:
			if err := checkIdx(pc, op, operand, nConsts, "constant"); err != nil {
				return err
			}
		case OP_DEFINE_FUNC:
			if err := checkIdx(pc, op, operand, nFuncs, "function"); err != nil {
				return err
			}
		case OP_DEFINE_STRUCT:
			if err := checkIdx(pc, op, operand, nStructs, "struct"); err != nil {
				return err
			}
		}
	}

	// Recurse into nested chunks.
	for _, fn := range c.Funcs {
		if fn == nil {
			return fmt.Errorf("ivm: corrupt bytecode: nil entry in function pool")
		}
		if err := fn.Body.validate(seen); err != nil {
			return err
		}
	}
	for _, sd := range c.StructDefs {
		if sd == nil {
			return fmt.Errorf("ivm: corrupt bytecode: nil entry in struct pool")
		}
		for _, fd := range sd.Fields {
			if fd != nil && fd.DefaultExprChunk != nil {
				if err := fd.DefaultExprChunk.validate(seen); err != nil {
					return err
				}
			}
		}
		for _, m := range sd.Methods {
			if m != nil {
				if err := m.Body.validate(seen); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// maxPackedIndex is the largest value that fits in the 16-bit half of a packed
// operand. OP_CALL, OP_CALL_METHOD, OP_NEW_STRUCT, OP_SWAP_VARS and
// OP_DEFINE_ERROR_TYPE pack two values into one 32-bit operand, so a name pool
// larger than this would silently wrap and resolve the wrong name.
const maxPackedIndex = 0xFFFF

// CheckLimits reports whether the chunk stays within the encoding's structural
// limits, recursing into nested function and method chunks. It is checked after
// compilation so that an over-large program fails loudly instead of emitting
// bytecode that quietly refers to the wrong names.
func (c *Chunk) CheckLimits() error {
	return c.checkLimits(make(map[*Chunk]bool))
}

func (c *Chunk) checkLimits(seen map[*Chunk]bool) error {
	if c == nil || seen[c] {
		return nil
	}
	seen[c] = true

	if len(c.Names) > maxPackedIndex {
		return fmt.Errorf(
			"ivm: program uses %d distinct names, but the bytecode format allows at most %d",
			len(c.Names), maxPackedIndex)
	}

	for pc, ins := range c.Code {
		switch ins.Op {
		case OP_CALL, OP_CALL_METHOD, OP_NEW_STRUCT:
			if n := ins.Operand >> 16; n > maxPackedIndex {
				return fmt.Errorf(
					"ivm: instruction %d (%s) has %d operands, but at most %d are supported",
					pc, OpName(ins.Op), n, maxPackedIndex)
			}
		}
	}

	for _, fn := range c.Funcs {
		if fn != nil {
			if err := fn.Body.checkLimits(seen); err != nil {
				return err
			}
		}
	}
	for _, sd := range c.StructDefs {
		if sd == nil {
			continue
		}
		for _, fd := range sd.Fields {
			if fd != nil && fd.DefaultExprChunk != nil {
				if err := fd.DefaultExprChunk.checkLimits(seen); err != nil {
					return err
				}
			}
		}
		for _, m := range sd.Methods {
			if m != nil {
				if err := m.Body.checkLimits(seen); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
