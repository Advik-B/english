package ivm

import "github.com/Advik-B/english/ast"

func (c *Compiler) compileStructDecl(s *ast.StructDecl) error {
	sd := &StructDef{Name: s.Name}

	// Compile fields
	for _, field := range s.Fields {
		fd := &FieldDef{
			Name:     field.Name,
			TypeName: ast.TypeName(field.Type),
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
		fc, err := c.compileFuncBody(method.Name, method.ParamNames(), method.Body)
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

// compileStructInstantiation emits a name and a value for each field written
// in the instantiation, so that OP_NEW_STRUCT can bind them by name.
//
// It used to emit only the values, in the order the instantiation wrote them,
// while the machine bound them in the order the *definition* declared them. So
// writing the fields in any order but the declared one silently put each value
// in the wrong field:
//
//	Declare Person as a structure with the following fields:
//	    name is a text with "?" being the default.
//	    age is a number with 0 being the default.
//	thats it.
//
//	a new instance of Person with the following fields:
//	    age is 30.
//	    name is "Alice".
//	thats it.
//
// gave name = 30 and age = "Alice", with nothing reported. Omitting a
// non-final field shifted every field after it.
func (c *Compiler) compileStructInstantiation(e *ast.StructInstantiation) error {
	// The struct name, so the machine can find the definition.
	snIdx := c.chunk.AddName(e.StructName)

	// A (name, value) pair per field written.
	for _, fieldName := range e.FieldOrder {
		nameIdx := c.chunk.AddConst(fieldName)
		c.chunk.Emit(OP_LOAD_CONST, nameIdx)

		val, ok := e.FieldValues[fieldName]
		if !ok {
			c.chunk.Emit(OP_LOAD_NOTHING, 0)
			continue
		}
		if err := c.compileExpression(val); err != nil {
			return err
		}
	}

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
