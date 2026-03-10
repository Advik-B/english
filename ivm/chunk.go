package ivm

// Instruction is a single VM instruction: an opcode plus a 32-bit operand.
type Instruction struct {
	Op      Opcode
	Operand uint32
}

// Chunk is a compiled instruction stream together with its supporting data pools.
type Chunk struct {
	Constants  []interface{} // number (float64), string, bool, nil
	Names      []string      // variable/function names
	Code       []Instruction
	Funcs      []*FuncChunk // user-defined function sub-chunks
	StructDefs []*StructDef // struct type definitions
}

// FuncChunk is the compiled representation of a user-defined function.
type FuncChunk struct {
	Name   string
	Params []string
	Body   *Chunk
}

// StructDef is the compiled representation of a struct type declaration.
type StructDef struct {
	Name    string
	Fields  []*FieldDef
	Methods []*FuncChunk
}

// FieldDef describes a single struct field.
type FieldDef struct {
	Name             string
	TypeName         string
	DefaultExprChunk *Chunk // compiled default-value expression, or nil
}

// NewChunk allocates an empty Chunk.
func NewChunk() *Chunk {
	return &Chunk{
		Constants:  []interface{}{},
		Names:      []string{},
		Code:       []Instruction{},
		Funcs:      []*FuncChunk{},
		StructDefs: []*StructDef{},
	}
}

// AddConst appends a constant to the pool and returns its index.
// Constants are NOT deduplicated so every literal gets its own slot.
func (c *Chunk) AddConst(v interface{}) uint32 {
	c.Constants = append(c.Constants, v)
	return uint32(len(c.Constants) - 1)
}

// AddName appends a name to the pool, deduplicating by value, and returns the index.
func (c *Chunk) AddName(s string) uint32 {
	for i, n := range c.Names {
		if n == s {
			return uint32(i)
		}
	}
	c.Names = append(c.Names, s)
	return uint32(len(c.Names) - 1)
}

// Emit appends an instruction to the code stream.
func (c *Chunk) Emit(op Opcode, operand uint32) {
	c.Code = append(c.Code, Instruction{Op: op, Operand: operand})
}

// CurrentPos returns the index of the next instruction to be emitted.
// Used to record jump source positions for later patching.
func (c *Chunk) CurrentPos() int {
	return len(c.Code)
}

// PatchJump overwrites the operand of a previously emitted jump instruction.
func (c *Chunk) PatchJump(pos int, target uint32) {
	c.Code[pos].Operand = target
}
