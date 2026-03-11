package ivm

import "english/ast"

func (c *Compiler) compileStructDecl(s *ast.StructDecl) error {
	sd := &StructDef{Name: s.Name}

	// Compile fields
	for _, field := range s.Fields {
		fd := &FieldDef{
			Name:     field.Name,
			TypeName: field.TypeName,
		}
		if field.DefaultValue != nil {
			// Compile default value expression as a mini-chunk
			subComp := &Compiler{chunk: NewChunk()}
			if err := subComp.compileExpression(field.DefaultValue); err != nil {
				return err
			}
			subComp.chunk.Emit(OP_RETURN, 0)
			fd.DefaultExprChunk = subComp.chunk
		}
		sd.Fields = append(sd.Fields, fd)
	}

	// Compile methods
	for _, method := range s.Methods {
		fc, err := c.compileFuncBody(method.Name, method.Parameters, method.Body)
		if err != nil {
			return err
		}
		sd.Methods = append(sd.Methods, fc)
	}

	structIdx := uint32(len(c.chunk.StructDefs))
	c.chunk.StructDefs = append(c.chunk.StructDefs, sd)
	c.chunk.Emit(OP_DEFINE_STRUCT, structIdx)
	return nil
}

func (c *Compiler) compileStructInstantiation(e *ast.StructInstantiation) error {
	// We need the StructDef to know field order and defaults.
	// Since we don't have access to the struct definition at compile time,
	// we compile the provided fields in the order they appear in FieldOrder,
	// and let the machine look up defaults for missing fields.
	// Stack: [field_value1, field_value2, ...] for each field in FieldOrder
	// Then the machine uses NEW_STRUCT to construct the instance.

	// Push the struct name so the machine can look up the def
	snIdx := c.chunk.AddName(e.StructName)

	// Push each specified field in order
	for _, fieldName := range e.FieldOrder {
		val, ok := e.FieldValues[fieldName]
		if !ok {
			c.chunk.Emit(OP_LOAD_NOTHING, 0)
		} else {
			if err := c.compileExpression(val); err != nil {
				return err
			}
		}
	}

	// Encode field_count<<16 | struct_name_idx
	fieldCount := uint32(len(e.FieldOrder))
	c.chunk.Emit(OP_NEW_STRUCT, fieldCount<<16|snIdx)
	return nil
}

func (c *Compiler) compileFuncBody(name string, params []string, body []ast.Statement) (*FuncChunk, error) {
	subComp := &Compiler{
		chunk:    NewChunk(),
		funcName: name,
	}
	if err := subComp.compileStatements(body); err != nil {
		return nil, err
	}
	// Implicit nil return at end of function
	subComp.chunk.Emit(OP_LOAD_NOTHING, 0)
	subComp.chunk.Emit(OP_RETURN, 0)

	return &FuncChunk{
		Name:   name,
		Params: params,
		Body:   subComp.chunk,
	}, nil
}
