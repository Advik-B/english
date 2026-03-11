package ivm

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
)

// Magic bytes and version for .101 instruction-format files.
var MagicBytes = []byte{0x10, 0x1E, 0x4E, 0x47}

// InstructionFormatVersion is the bytecode format version for instruction-based .101 files.
const InstructionFormatVersion uint8 = 3

// EncodeFile serialises chunk with magic header + version byte.
func EncodeFile(chunk *Chunk) ([]byte, error) {
	return EncodeFileWithSource(chunk, "")
}

// EncodeFileWithSource serialises chunk with magic header + version byte and
// appends the original source code as a trailing section so the file can be
// transpiled to Python without the original .abc file.
// The trailer format is: [uint32-LE source_len][source UTF-8 bytes].
// If source is empty the trailer is omitted and the output is identical to
// EncodeFile.
func EncodeFileWithSource(chunk *Chunk, source string) ([]byte, error) {
	body, err := EncodeChunk(chunk)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	buf.Write(MagicBytes)
	buf.WriteByte(InstructionFormatVersion)
	buf.Write(body)
	if source != "" {
		srcBytes := []byte(source)
		var lenBuf [4]byte
		binary.LittleEndian.PutUint32(lenBuf[:], uint32(len(srcBytes)))
		buf.Write(lenBuf[:])
		buf.Write(srcBytes)
	}
	return buf.Bytes(), nil
}

// DecodeFile verifies magic + version and deserialises the chunk.
// Any embedded source trailer is silently ignored; use DecodeFileAll to
// retrieve it.
func DecodeFile(data []byte) (*Chunk, error) {
	chunk, _, err := DecodeFileAll(data)
	return chunk, err
}

// DecodeFileAll verifies magic + version, deserialises the chunk, and returns
// any embedded source code. The returned source is empty when the file was
// produced without a source trailer (e.g. compiled with an older version of
// the tool).
func DecodeFileAll(data []byte) (*Chunk, string, error) {
	if len(data) < 5 {
		return nil, "", fmt.Errorf("ivm: file too short")
	}
	if !bytes.Equal(data[:4], MagicBytes) {
		return nil, "", fmt.Errorf("ivm: bad magic bytes")
	}
	if data[4] != InstructionFormatVersion {
		return nil, "", fmt.Errorf("ivm: unsupported format version %d (expected %d)", data[4], InstructionFormatVersion)
	}
	d := &decoder{data: data[5:], pos: 0}
	chunk, err := d.readChunk()
	if err != nil {
		return nil, "", err
	}
	// Check for optional source trailer: [uint32-LE len][source bytes]
	remaining := d.data[d.pos:]
	var source string
	if len(remaining) >= 4 {
		srcLen := binary.LittleEndian.Uint32(remaining[:4])
		// Guard against malformed files: srcLen must fit in a signed int and not
		// exceed the available bytes.
		if uint64(srcLen) <= uint64(len(remaining)-4) {
			source = string(remaining[4 : 4+srcLen])
		}
	}
	return chunk, source, nil
}

// EncodeChunk serialises a Chunk to binary (without file header).
func EncodeChunk(chunk *Chunk) ([]byte, error) {
	var buf bytes.Buffer
	e := &encoder{buf: &buf}
	if err := e.writeChunk(chunk); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// DecodeChunk deserialises a Chunk from raw binary (without file header).
func DecodeChunk(data []byte) (*Chunk, error) {
	d := &decoder{data: data, pos: 0}
	return d.readChunk()
}

// ─── encoder ──────────────────────────────────────────────────────────────────

type encoder struct {
	buf *bytes.Buffer
}

func (e *encoder) writeUint32(v uint32) {
	b := [4]byte{}
	binary.LittleEndian.PutUint32(b[:], v)
	e.buf.Write(b[:])
}

func (e *encoder) writeByte(b byte) {
	e.buf.WriteByte(b)
}

func (e *encoder) writeString(s string) {
	e.writeUint32(uint32(len(s)))
	e.buf.WriteString(s)
}

func (e *encoder) writeFloat64(f float64) {
	b := [8]byte{}
	binary.LittleEndian.PutUint64(b[:], math.Float64bits(f))
	e.buf.Write(b[:])
}

func (e *encoder) writeChunk(c *Chunk) error {
	// Constants
	e.writeUint32(uint32(len(c.Constants)))
	for _, cv := range c.Constants {
		if err := e.writeConstant(cv); err != nil {
			return err
		}
	}
	// Names
	e.writeUint32(uint32(len(c.Names)))
	for _, name := range c.Names {
		e.writeString(name)
	}
	// Instructions
	e.writeUint32(uint32(len(c.Code)))
	for _, instr := range c.Code {
		e.writeByte(byte(instr.Op))
		e.writeUint32(instr.Operand)
	}
	// Funcs
	e.writeUint32(uint32(len(c.Funcs)))
	for _, fc := range c.Funcs {
		if err := e.writeFuncChunk(fc); err != nil {
			return err
		}
	}
	// StructDefs
	e.writeUint32(uint32(len(c.StructDefs)))
	for _, sd := range c.StructDefs {
		if err := e.writeStructDef(sd); err != nil {
			return err
		}
	}
	return nil
}

func (e *encoder) writeConstant(v interface{}) error {
	switch val := v.(type) {
	case float64:
		e.writeByte(0)
		e.writeFloat64(val)
	case string:
		e.writeByte(1)
		e.writeString(val)
	case bool:
		e.writeByte(2)
		if val {
			e.writeByte(1)
		} else {
			e.writeByte(0)
		}
	case nil:
		e.writeByte(3)
	case []interface{}:
		// String-slice of items (used for import items)
		e.writeByte(4)
		e.writeUint32(uint32(len(val)))
		for _, item := range val {
			s, _ := item.(string)
			e.writeString(s)
		}
	default:
		return fmt.Errorf("ivm encoding: unsupported constant type %T", v)
	}
	return nil
}

func (e *encoder) writeFuncChunk(fc *FuncChunk) error {
	e.writeString(fc.Name)
	e.writeUint32(uint32(len(fc.Params)))
	for _, p := range fc.Params {
		e.writeString(p)
	}
	return e.writeChunk(fc.Body)
}

func (e *encoder) writeStructDef(sd *StructDef) error {
	e.writeString(sd.Name)
	// Fields
	e.writeUint32(uint32(len(sd.Fields)))
	for _, fd := range sd.Fields {
		e.writeString(fd.Name)
		e.writeString(fd.TypeName)
		if fd.DefaultExprChunk != nil {
			e.writeByte(1)
			if err := e.writeChunk(fd.DefaultExprChunk); err != nil {
				return err
			}
		} else {
			e.writeByte(0)
		}
	}
	// Methods
	e.writeUint32(uint32(len(sd.Methods)))
	for _, m := range sd.Methods {
		if err := e.writeFuncChunk(m); err != nil {
			return err
		}
	}
	return nil
}

// ─── decoder ──────────────────────────────────────────────────────────────────

type decoder struct {
	data []byte
	pos  int
}

func (d *decoder) readUint32() (uint32, error) {
	if d.pos+4 > len(d.data) {
		return 0, fmt.Errorf("ivm decode: unexpected end of data reading uint32")
	}
	v := binary.LittleEndian.Uint32(d.data[d.pos:])
	d.pos += 4
	return v, nil
}

func (d *decoder) readByte() (byte, error) {
	if d.pos >= len(d.data) {
		return 0, fmt.Errorf("ivm decode: unexpected end of data reading byte")
	}
	b := d.data[d.pos]
	d.pos++
	return b, nil
}

func (d *decoder) readString() (string, error) {
	n, err := d.readUint32()
	if err != nil {
		return "", err
	}
	if d.pos+int(n) > len(d.data) {
		return "", fmt.Errorf("ivm decode: unexpected end of data reading string")
	}
	s := string(d.data[d.pos : d.pos+int(n)])
	d.pos += int(n)
	return s, nil
}

func (d *decoder) readFloat64() (float64, error) {
	if d.pos+8 > len(d.data) {
		return 0, fmt.Errorf("ivm decode: unexpected end of data reading float64")
	}
	bits := binary.LittleEndian.Uint64(d.data[d.pos:])
	d.pos += 8
	return math.Float64frombits(bits), nil
}

func (d *decoder) readChunk() (*Chunk, error) {
	c := NewChunk()

	// Constants
	cCount, err := d.readUint32()
	if err != nil {
		return nil, err
	}
	c.Constants = make([]interface{}, cCount)
	for i := uint32(0); i < cCount; i++ {
		cv, err := d.readConstant()
		if err != nil {
			return nil, err
		}
		c.Constants[i] = cv
	}

	// Names
	nCount, err := d.readUint32()
	if err != nil {
		return nil, err
	}
	c.Names = make([]string, nCount)
	for i := uint32(0); i < nCount; i++ {
		s, err := d.readString()
		if err != nil {
			return nil, err
		}
		c.Names[i] = s
	}

	// Instructions
	iCount, err := d.readUint32()
	if err != nil {
		return nil, err
	}
	c.Code = make([]Instruction, iCount)
	for i := uint32(0); i < iCount; i++ {
		opByte, err := d.readByte()
		if err != nil {
			return nil, err
		}
		operand, err := d.readUint32()
		if err != nil {
			return nil, err
		}
		c.Code[i] = Instruction{Op: Opcode(opByte), Operand: operand}
	}

	// Funcs
	fCount, err := d.readUint32()
	if err != nil {
		return nil, err
	}
	c.Funcs = make([]*FuncChunk, fCount)
	for i := uint32(0); i < fCount; i++ {
		fc, err := d.readFuncChunk()
		if err != nil {
			return nil, err
		}
		c.Funcs[i] = fc
	}

	// StructDefs
	sCount, err := d.readUint32()
	if err != nil {
		return nil, err
	}
	c.StructDefs = make([]*StructDef, sCount)
	for i := uint32(0); i < sCount; i++ {
		sd, err := d.readStructDef()
		if err != nil {
			return nil, err
		}
		c.StructDefs[i] = sd
	}

	return c, nil
}

func (d *decoder) readConstant() (interface{}, error) {
	tag, err := d.readByte()
	if err != nil {
		return nil, err
	}
	switch tag {
	case 0: // float64
		return d.readFloat64()
	case 1: // string
		return d.readString()
	case 2: // bool
		b, err := d.readByte()
		if err != nil {
			return nil, err
		}
		return b != 0, nil
	case 3: // nil
		return nil, nil
	case 4: // []interface{} of strings
		n, err := d.readUint32()
		if err != nil {
			return nil, err
		}
		items := make([]interface{}, n)
		for i := uint32(0); i < n; i++ {
			s, err := d.readString()
			if err != nil {
				return nil, err
			}
			items[i] = s
		}
		return items, nil
	default:
		return nil, fmt.Errorf("ivm decode: unknown constant tag %d", tag)
	}
}

func (d *decoder) readFuncChunk() (*FuncChunk, error) {
	name, err := d.readString()
	if err != nil {
		return nil, err
	}
	pCount, err := d.readUint32()
	if err != nil {
		return nil, err
	}
	params := make([]string, pCount)
	for i := uint32(0); i < pCount; i++ {
		p, err := d.readString()
		if err != nil {
			return nil, err
		}
		params[i] = p
	}
	body, err := d.readChunk()
	if err != nil {
		return nil, err
	}
	return &FuncChunk{Name: name, Params: params, Body: body}, nil
}

func (d *decoder) readStructDef() (*StructDef, error) {
	name, err := d.readString()
	if err != nil {
		return nil, err
	}
	sd := &StructDef{Name: name}

	// Fields
	fCount, err := d.readUint32()
	if err != nil {
		return nil, err
	}
	sd.Fields = make([]*FieldDef, fCount)
	for i := uint32(0); i < fCount; i++ {
		fname, err := d.readString()
		if err != nil {
			return nil, err
		}
		ftypeName, err := d.readString()
		if err != nil {
			return nil, err
		}
		hasDefault, err := d.readByte()
		if err != nil {
			return nil, err
		}
		fd := &FieldDef{Name: fname, TypeName: ftypeName}
		if hasDefault == 1 {
			defChunk, err := d.readChunk()
			if err != nil {
				return nil, err
			}
			fd.DefaultExprChunk = defChunk
		}
		sd.Fields[i] = fd
	}

	// Methods
	mCount, err := d.readUint32()
	if err != nil {
		return nil, err
	}
	sd.Methods = make([]*FuncChunk, mCount)
	for i := uint32(0); i < mCount; i++ {
		mc, err := d.readFuncChunk()
		if err != nil {
			return nil, err
		}
		sd.Methods[i] = mc
	}

	return sd, nil
}
