// Package bytecode provides binary serialization and deserialization
// of AST nodes for the English programming language.
// The binary format (.101 files) can be directly loaded and evaluated
// without parsing, similar to protobuf serialization.
package bytecode

import (
	"bytes"
	"encoding/binary"
	"english/ast"
	"fmt"
	"io"
	"math"
)

// Magic bytes to identify .101 files (binary identifier)
var MagicBytes = []byte{0x10, 0x1E, 0x4E, 0x47}

// Version of the bytecode format
const FormatVersion uint8 = 1

// Node type identifiers
const (
	NodeProgram byte = iota + 1
	NodeVariableDecl
	NodeAssignment
	NodeFunctionDecl
	NodeFunctionCall
	NodeCallStatement
	NodeIfStatement
	NodeElseIfPart
	NodeWhileLoop
	NodeForLoop
	NodeForEachLoop
	NodeNumberLiteral
	NodeStringLiteral
	NodeBooleanLiteral
	NodeListLiteral
	NodeIdentifier
	NodeBinaryExpression
	NodeUnaryExpression
	NodeIndexExpression
	NodeIndexAssignment
	NodeLengthExpression
	NodeReturnStatement
	NodeOutputStatement
	NodeToggleStatement
	NodeLocationExpression
)

// Encoder serializes AST to binary format
type Encoder struct {
	buf *bytes.Buffer
}

// NewEncoder creates a new bytecode encoder
func NewEncoder() *Encoder {
	return &Encoder{
		buf: new(bytes.Buffer),
	}
}

// Encode serializes a Program AST to binary bytecode
func (e *Encoder) Encode(program *ast.Program) ([]byte, error) {
	e.buf.Reset()

	// Write magic bytes
	e.buf.Write(MagicBytes)

	// Write version
	e.buf.WriteByte(FormatVersion)

	// Encode the program
	if err := e.encodeProgram(program); err != nil {
		return nil, err
	}

	return e.buf.Bytes(), nil
}

func (e *Encoder) encodeProgram(p *ast.Program) error {
	e.buf.WriteByte(NodeProgram)
	e.writeUint32(uint32(len(p.Statements)))

	for _, stmt := range p.Statements {
		if err := e.encodeStatement(stmt); err != nil {
			return err
		}
	}
	return nil
}

func (e *Encoder) writeString(s string) {
	data := []byte(s)
	e.writeUint32(uint32(len(data)))
	e.buf.Write(data)
}

func (e *Encoder) writeUint32(v uint32) {
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, v)
	e.buf.Write(b)
}

func (e *Encoder) writeFloat64(v float64) {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, math.Float64bits(v))
	e.buf.Write(b)
}

func (e *Encoder) writeBool(v bool) {
	if v {
		e.buf.WriteByte(1)
	} else {
		e.buf.WriteByte(0)
	}
}

func (e *Encoder) encodeStatement(stmt ast.Statement) error {
	switch s := stmt.(type) {
	case *ast.VariableDecl:
		e.buf.WriteByte(NodeVariableDecl)
		e.writeString(s.Name)
		e.writeBool(s.IsConstant)
		return e.encodeExpression(s.Value)

	case *ast.Assignment:
		e.buf.WriteByte(NodeAssignment)
		e.writeString(s.Name)
		return e.encodeExpression(s.Value)

	case *ast.FunctionDecl:
		e.buf.WriteByte(NodeFunctionDecl)
		e.writeString(s.Name)
		e.writeUint32(uint32(len(s.Parameters)))
		for _, param := range s.Parameters {
			e.writeString(param)
		}
		e.writeUint32(uint32(len(s.Body)))
		for _, bodyStmt := range s.Body {
			if err := e.encodeStatement(bodyStmt); err != nil {
				return err
			}
		}
		return nil

	case *ast.CallStatement:
		e.buf.WriteByte(NodeCallStatement)
		return e.encodeFunctionCall(s.FunctionCall)

	case *ast.IfStatement:
		e.buf.WriteByte(NodeIfStatement)
		if err := e.encodeExpression(s.Condition); err != nil {
			return err
		}
		e.writeUint32(uint32(len(s.Then)))
		for _, thenStmt := range s.Then {
			if err := e.encodeStatement(thenStmt); err != nil {
				return err
			}
		}
		e.writeUint32(uint32(len(s.ElseIf)))
		for _, elseIf := range s.ElseIf {
			if err := e.encodeElseIfPart(elseIf); err != nil {
				return err
			}
		}
		e.writeUint32(uint32(len(s.Else)))
		for _, elseStmt := range s.Else {
			if err := e.encodeStatement(elseStmt); err != nil {
				return err
			}
		}
		return nil

	case *ast.WhileLoop:
		e.buf.WriteByte(NodeWhileLoop)
		if err := e.encodeExpression(s.Condition); err != nil {
			return err
		}
		e.writeUint32(uint32(len(s.Body)))
		for _, bodyStmt := range s.Body {
			if err := e.encodeStatement(bodyStmt); err != nil {
				return err
			}
		}
		return nil

	case *ast.ForLoop:
		e.buf.WriteByte(NodeForLoop)
		if err := e.encodeExpression(s.Count); err != nil {
			return err
		}
		e.writeUint32(uint32(len(s.Body)))
		for _, bodyStmt := range s.Body {
			if err := e.encodeStatement(bodyStmt); err != nil {
				return err
			}
		}
		return nil

	case *ast.ForEachLoop:
		e.buf.WriteByte(NodeForEachLoop)
		e.writeString(s.Item)
		if err := e.encodeExpression(s.List); err != nil {
			return err
		}
		e.writeUint32(uint32(len(s.Body)))
		for _, bodyStmt := range s.Body {
			if err := e.encodeStatement(bodyStmt); err != nil {
				return err
			}
		}
		return nil

	case *ast.IndexAssignment:
		e.buf.WriteByte(NodeIndexAssignment)
		e.writeString(s.ListName)
		if err := e.encodeExpression(s.Index); err != nil {
			return err
		}
		return e.encodeExpression(s.Value)

	case *ast.ReturnStatement:
		e.buf.WriteByte(NodeReturnStatement)
		return e.encodeExpression(s.Value)

	case *ast.OutputStatement:
		e.buf.WriteByte(NodeOutputStatement)
		return e.encodeExpression(s.Value)

	case *ast.ToggleStatement:
		e.buf.WriteByte(NodeToggleStatement)
		e.writeString(s.Name)
		return nil

	default:
		return fmt.Errorf("unknown statement type: %T", stmt)
	}
}

func (e *Encoder) encodeElseIfPart(part *ast.ElseIfPart) error {
	e.buf.WriteByte(NodeElseIfPart)
	if err := e.encodeExpression(part.Condition); err != nil {
		return err
	}
	e.writeUint32(uint32(len(part.Body)))
	for _, stmt := range part.Body {
		if err := e.encodeStatement(stmt); err != nil {
			return err
		}
	}
	return nil
}

func (e *Encoder) encodeFunctionCall(fc *ast.FunctionCall) error {
	e.buf.WriteByte(NodeFunctionCall)
	e.writeString(fc.Name)
	e.writeUint32(uint32(len(fc.Arguments)))
	for _, arg := range fc.Arguments {
		if err := e.encodeExpression(arg); err != nil {
			return err
		}
	}
	return nil
}

func (e *Encoder) encodeExpression(expr ast.Expression) error {
	switch ex := expr.(type) {
	case *ast.NumberLiteral:
		e.buf.WriteByte(NodeNumberLiteral)
		e.writeFloat64(ex.Value)
		return nil

	case *ast.StringLiteral:
		e.buf.WriteByte(NodeStringLiteral)
		e.writeString(ex.Value)
		return nil

	case *ast.BooleanLiteral:
		e.buf.WriteByte(NodeBooleanLiteral)
		e.writeBool(ex.Value)
		return nil

	case *ast.ListLiteral:
		e.buf.WriteByte(NodeListLiteral)
		e.writeUint32(uint32(len(ex.Elements)))
		for _, elem := range ex.Elements {
			if err := e.encodeExpression(elem); err != nil {
				return err
			}
		}
		return nil

	case *ast.Identifier:
		e.buf.WriteByte(NodeIdentifier)
		e.writeString(ex.Name)
		return nil

	case *ast.BinaryExpression:
		e.buf.WriteByte(NodeBinaryExpression)
		e.writeString(ex.Operator)
		if err := e.encodeExpression(ex.Left); err != nil {
			return err
		}
		return e.encodeExpression(ex.Right)

	case *ast.UnaryExpression:
		e.buf.WriteByte(NodeUnaryExpression)
		e.writeString(ex.Operator)
		return e.encodeExpression(ex.Right)

	case *ast.FunctionCall:
		return e.encodeFunctionCall(ex)

	case *ast.IndexExpression:
		e.buf.WriteByte(NodeIndexExpression)
		if err := e.encodeExpression(ex.List); err != nil {
			return err
		}
		return e.encodeExpression(ex.Index)

	case *ast.LengthExpression:
		e.buf.WriteByte(NodeLengthExpression)
		return e.encodeExpression(ex.List)

	case *ast.LocationExpression:
		e.buf.WriteByte(NodeLocationExpression)
		e.writeString(ex.Name)
		return nil

	default:
		return fmt.Errorf("unknown expression type: %T", expr)
	}
}

// Decoder deserializes binary bytecode to AST
type Decoder struct {
	reader io.Reader
}

// NewDecoder creates a new bytecode decoder
func NewDecoder(data []byte) *Decoder {
	return &Decoder{
		reader: bytes.NewReader(data),
	}
}

// Decode deserializes binary bytecode to a Program AST
func (d *Decoder) Decode() (*ast.Program, error) {
	// Verify magic bytes
	magic := make([]byte, 4)
	if _, err := io.ReadFull(d.reader, magic); err != nil {
		return nil, fmt.Errorf("failed to read magic bytes: %w", err)
	}
	if !bytes.Equal(magic, MagicBytes) {
		return nil, fmt.Errorf("invalid bytecode file: wrong magic bytes")
	}

	// Read version
	version := make([]byte, 1)
	if _, err := io.ReadFull(d.reader, version); err != nil {
		return nil, fmt.Errorf("failed to read version: %w", err)
	}
	if version[0] != FormatVersion {
		return nil, fmt.Errorf("unsupported bytecode version: %d (expected %d)", version[0], FormatVersion)
	}

	return d.decodeProgram()
}

func (d *Decoder) readByte() (byte, error) {
	b := make([]byte, 1)
	if _, err := io.ReadFull(d.reader, b); err != nil {
		return 0, err
	}
	return b[0], nil
}

func (d *Decoder) readString() (string, error) {
	length, err := d.readUint32()
	if err != nil {
		return "", err
	}
	data := make([]byte, length)
	if _, err := io.ReadFull(d.reader, data); err != nil {
		return "", err
	}
	return string(data), nil
}

func (d *Decoder) readUint32() (uint32, error) {
	b := make([]byte, 4)
	if _, err := io.ReadFull(d.reader, b); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(b), nil
}

func (d *Decoder) readFloat64() (float64, error) {
	b := make([]byte, 8)
	if _, err := io.ReadFull(d.reader, b); err != nil {
		return 0, err
	}
	return math.Float64frombits(binary.LittleEndian.Uint64(b)), nil
}

func (d *Decoder) readBool() (bool, error) {
	b, err := d.readByte()
	if err != nil {
		return false, err
	}
	return b != 0, nil
}

func (d *Decoder) decodeProgram() (*ast.Program, error) {
	nodeType, err := d.readByte()
	if err != nil {
		return nil, err
	}
	if nodeType != NodeProgram {
		return nil, fmt.Errorf("expected Program node, got %d", nodeType)
	}

	count, err := d.readUint32()
	if err != nil {
		return nil, err
	}

	statements := make([]ast.Statement, count)
	for i := uint32(0); i < count; i++ {
		stmt, err := d.decodeStatement()
		if err != nil {
			return nil, err
		}
		statements[i] = stmt
	}

	return &ast.Program{Statements: statements}, nil
}

func (d *Decoder) decodeStatement() (ast.Statement, error) {
	nodeType, err := d.readByte()
	if err != nil {
		return nil, err
	}

	switch nodeType {
	case NodeVariableDecl:
		name, err := d.readString()
		if err != nil {
			return nil, err
		}
		isConstant, err := d.readBool()
		if err != nil {
			return nil, err
		}
		value, err := d.decodeExpression()
		if err != nil {
			return nil, err
		}
		return &ast.VariableDecl{Name: name, IsConstant: isConstant, Value: value}, nil

	case NodeAssignment:
		name, err := d.readString()
		if err != nil {
			return nil, err
		}
		value, err := d.decodeExpression()
		if err != nil {
			return nil, err
		}
		return &ast.Assignment{Name: name, Value: value}, nil

	case NodeFunctionDecl:
		name, err := d.readString()
		if err != nil {
			return nil, err
		}
		paramCount, err := d.readUint32()
		if err != nil {
			return nil, err
		}
		params := make([]string, paramCount)
		for i := uint32(0); i < paramCount; i++ {
			params[i], err = d.readString()
			if err != nil {
				return nil, err
			}
		}
		bodyCount, err := d.readUint32()
		if err != nil {
			return nil, err
		}
		body := make([]ast.Statement, bodyCount)
		for i := uint32(0); i < bodyCount; i++ {
			body[i], err = d.decodeStatement()
			if err != nil {
				return nil, err
			}
		}
		return &ast.FunctionDecl{Name: name, Parameters: params, Body: body}, nil

	case NodeCallStatement:
		fc, err := d.decodeFunctionCall()
		if err != nil {
			return nil, err
		}
		return &ast.CallStatement{FunctionCall: fc}, nil

	case NodeIfStatement:
		condition, err := d.decodeExpression()
		if err != nil {
			return nil, err
		}
		thenCount, err := d.readUint32()
		if err != nil {
			return nil, err
		}
		thenBody := make([]ast.Statement, thenCount)
		for i := uint32(0); i < thenCount; i++ {
			thenBody[i], err = d.decodeStatement()
			if err != nil {
				return nil, err
			}
		}
		elseIfCount, err := d.readUint32()
		if err != nil {
			return nil, err
		}
		elseIfParts := make([]*ast.ElseIfPart, elseIfCount)
		for i := uint32(0); i < elseIfCount; i++ {
			elseIfParts[i], err = d.decodeElseIfPart()
			if err != nil {
				return nil, err
			}
		}
		elseCount, err := d.readUint32()
		if err != nil {
			return nil, err
		}
		elseBody := make([]ast.Statement, elseCount)
		for i := uint32(0); i < elseCount; i++ {
			elseBody[i], err = d.decodeStatement()
			if err != nil {
				return nil, err
			}
		}
		return &ast.IfStatement{Condition: condition, Then: thenBody, ElseIf: elseIfParts, Else: elseBody}, nil

	case NodeWhileLoop:
		condition, err := d.decodeExpression()
		if err != nil {
			return nil, err
		}
		bodyCount, err := d.readUint32()
		if err != nil {
			return nil, err
		}
		body := make([]ast.Statement, bodyCount)
		for i := uint32(0); i < bodyCount; i++ {
			body[i], err = d.decodeStatement()
			if err != nil {
				return nil, err
			}
		}
		return &ast.WhileLoop{Condition: condition, Body: body}, nil

	case NodeForLoop:
		count, err := d.decodeExpression()
		if err != nil {
			return nil, err
		}
		bodyCount, err := d.readUint32()
		if err != nil {
			return nil, err
		}
		body := make([]ast.Statement, bodyCount)
		for i := uint32(0); i < bodyCount; i++ {
			body[i], err = d.decodeStatement()
			if err != nil {
				return nil, err
			}
		}
		return &ast.ForLoop{Count: count, Body: body}, nil

	case NodeForEachLoop:
		item, err := d.readString()
		if err != nil {
			return nil, err
		}
		list, err := d.decodeExpression()
		if err != nil {
			return nil, err
		}
		bodyCount, err := d.readUint32()
		if err != nil {
			return nil, err
		}
		body := make([]ast.Statement, bodyCount)
		for i := uint32(0); i < bodyCount; i++ {
			body[i], err = d.decodeStatement()
			if err != nil {
				return nil, err
			}
		}
		return &ast.ForEachLoop{Item: item, List: list, Body: body}, nil

	case NodeIndexAssignment:
		listName, err := d.readString()
		if err != nil {
			return nil, err
		}
		index, err := d.decodeExpression()
		if err != nil {
			return nil, err
		}
		value, err := d.decodeExpression()
		if err != nil {
			return nil, err
		}
		return &ast.IndexAssignment{ListName: listName, Index: index, Value: value}, nil

	case NodeReturnStatement:
		value, err := d.decodeExpression()
		if err != nil {
			return nil, err
		}
		return &ast.ReturnStatement{Value: value}, nil

	case NodeOutputStatement:
		value, err := d.decodeExpression()
		if err != nil {
			return nil, err
		}
		return &ast.OutputStatement{Value: value}, nil

	case NodeToggleStatement:
		name, err := d.readString()
		if err != nil {
			return nil, err
		}
		return &ast.ToggleStatement{Name: name}, nil

	default:
		return nil, fmt.Errorf("unknown statement node type: %d", nodeType)
	}
}

func (d *Decoder) decodeElseIfPart() (*ast.ElseIfPart, error) {
	nodeType, err := d.readByte()
	if err != nil {
		return nil, err
	}
	if nodeType != NodeElseIfPart {
		return nil, fmt.Errorf("expected ElseIfPart node, got %d", nodeType)
	}

	condition, err := d.decodeExpression()
	if err != nil {
		return nil, err
	}
	bodyCount, err := d.readUint32()
	if err != nil {
		return nil, err
	}
	body := make([]ast.Statement, bodyCount)
	for i := uint32(0); i < bodyCount; i++ {
		body[i], err = d.decodeStatement()
		if err != nil {
			return nil, err
		}
	}
	return &ast.ElseIfPart{Condition: condition, Body: body}, nil
}

func (d *Decoder) decodeFunctionCall() (*ast.FunctionCall, error) {
	nodeType, err := d.readByte()
	if err != nil {
		return nil, err
	}
	if nodeType != NodeFunctionCall {
		return nil, fmt.Errorf("expected FunctionCall node, got %d", nodeType)
	}

	name, err := d.readString()
	if err != nil {
		return nil, err
	}
	argCount, err := d.readUint32()
	if err != nil {
		return nil, err
	}
	args := make([]ast.Expression, argCount)
	for i := uint32(0); i < argCount; i++ {
		args[i], err = d.decodeExpression()
		if err != nil {
			return nil, err
		}
	}
	return &ast.FunctionCall{Name: name, Arguments: args}, nil
}

func (d *Decoder) decodeExpression() (ast.Expression, error) {
	nodeType, err := d.readByte()
	if err != nil {
		return nil, err
	}

	switch nodeType {
	case NodeNumberLiteral:
		value, err := d.readFloat64()
		if err != nil {
			return nil, err
		}
		return &ast.NumberLiteral{Value: value}, nil

	case NodeStringLiteral:
		value, err := d.readString()
		if err != nil {
			return nil, err
		}
		return &ast.StringLiteral{Value: value}, nil

	case NodeBooleanLiteral:
		value, err := d.readBool()
		if err != nil {
			return nil, err
		}
		return &ast.BooleanLiteral{Value: value}, nil

	case NodeListLiteral:
		count, err := d.readUint32()
		if err != nil {
			return nil, err
		}
		elements := make([]ast.Expression, count)
		for i := uint32(0); i < count; i++ {
			elements[i], err = d.decodeExpression()
			if err != nil {
				return nil, err
			}
		}
		return &ast.ListLiteral{Elements: elements}, nil

	case NodeIdentifier:
		name, err := d.readString()
		if err != nil {
			return nil, err
		}
		return &ast.Identifier{Name: name}, nil

	case NodeBinaryExpression:
		operator, err := d.readString()
		if err != nil {
			return nil, err
		}
		left, err := d.decodeExpression()
		if err != nil {
			return nil, err
		}
		right, err := d.decodeExpression()
		if err != nil {
			return nil, err
		}
		return &ast.BinaryExpression{Left: left, Operator: operator, Right: right}, nil

	case NodeUnaryExpression:
		operator, err := d.readString()
		if err != nil {
			return nil, err
		}
		right, err := d.decodeExpression()
		if err != nil {
			return nil, err
		}
		return &ast.UnaryExpression{Operator: operator, Right: right}, nil

	case NodeFunctionCall:
		// Re-read because we already consumed the type byte
		name, err := d.readString()
		if err != nil {
			return nil, err
		}
		argCount, err := d.readUint32()
		if err != nil {
			return nil, err
		}
		args := make([]ast.Expression, argCount)
		for i := uint32(0); i < argCount; i++ {
			args[i], err = d.decodeExpression()
			if err != nil {
				return nil, err
			}
		}
		return &ast.FunctionCall{Name: name, Arguments: args}, nil

	case NodeIndexExpression:
		list, err := d.decodeExpression()
		if err != nil {
			return nil, err
		}
		index, err := d.decodeExpression()
		if err != nil {
			return nil, err
		}
		return &ast.IndexExpression{List: list, Index: index}, nil

	case NodeLengthExpression:
		list, err := d.decodeExpression()
		if err != nil {
			return nil, err
		}
		return &ast.LengthExpression{List: list}, nil

	case NodeLocationExpression:
		name, err := d.readString()
		if err != nil {
			return nil, err
		}
		return &ast.LocationExpression{Name: name}, nil

	default:
		return nil, fmt.Errorf("unknown expression node type: %d", nodeType)
	}
}
