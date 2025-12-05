package bytecode

import (
	"bytes"
	"english/ast"
	"testing"
)

func TestEncodeDecode(t *testing.T) {
	// Create a simple program
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.VariableDecl{
				Name:       "x",
				IsConstant: false,
				Value:      &ast.NumberLiteral{Value: 42},
			},
		},
	}

	// Encode
	encoder := NewEncoder()
	data, err := encoder.Encode(program)
	if err != nil {
		t.Fatalf("Encode error: %v", err)
	}

	// Verify magic bytes
	if !bytes.HasPrefix(data, MagicBytes) {
		t.Error("Encoded data should start with magic bytes")
	}

	// Decode
	decoder := NewDecoder(data)
	decoded, err := decoder.Decode()
	if err != nil {
		t.Fatalf("Decode error: %v", err)
	}

	// Verify
	if len(decoded.Statements) != 1 {
		t.Errorf("Expected 1 statement, got %d", len(decoded.Statements))
	}

	varDecl, ok := decoded.Statements[0].(*ast.VariableDecl)
	if !ok {
		t.Fatalf("Expected VariableDecl, got %T", decoded.Statements[0])
	}
	if varDecl.Name != "x" {
		t.Errorf("Expected name 'x', got %q", varDecl.Name)
	}
	if varDecl.IsConstant {
		t.Error("Expected IsConstant to be false")
	}
	numLit, ok := varDecl.Value.(*ast.NumberLiteral)
	if !ok {
		t.Fatalf("Expected NumberLiteral, got %T", varDecl.Value)
	}
	if numLit.Value != 42 {
		t.Errorf("Expected value 42, got %v", numLit.Value)
	}
}

func TestEncodeDecodeConstant(t *testing.T) {
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.VariableDecl{
				Name:       "PI",
				IsConstant: true,
				Value:      &ast.NumberLiteral{Value: 3.14159},
			},
		},
	}

	encoder := NewEncoder()
	data, err := encoder.Encode(program)
	if err != nil {
		t.Fatalf("Encode error: %v", err)
	}

	decoder := NewDecoder(data)
	decoded, err := decoder.Decode()
	if err != nil {
		t.Fatalf("Decode error: %v", err)
	}

	varDecl := decoded.Statements[0].(*ast.VariableDecl)
	if !varDecl.IsConstant {
		t.Error("Expected IsConstant to be true")
	}
	if varDecl.Value.(*ast.NumberLiteral).Value != 3.14159 {
		t.Errorf("Expected value 3.14159, got %v", varDecl.Value.(*ast.NumberLiteral).Value)
	}
}

func TestEncodeDecodeString(t *testing.T) {
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.VariableDecl{
				Name:       "greeting",
				IsConstant: false,
				Value:      &ast.StringLiteral{Value: "Hello, World!"},
			},
		},
	}

	encoder := NewEncoder()
	data, err := encoder.Encode(program)
	if err != nil {
		t.Fatalf("Encode error: %v", err)
	}

	decoder := NewDecoder(data)
	decoded, err := decoder.Decode()
	if err != nil {
		t.Fatalf("Decode error: %v", err)
	}

	varDecl := decoded.Statements[0].(*ast.VariableDecl)
	strLit := varDecl.Value.(*ast.StringLiteral)
	if strLit.Value != "Hello, World!" {
		t.Errorf("Expected 'Hello, World!', got %q", strLit.Value)
	}
}

func TestEncodeDecodeBoolean(t *testing.T) {
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.VariableDecl{
				Name:       "flag",
				IsConstant: false,
				Value:      &ast.BooleanLiteral{Value: true},
			},
		},
	}

	encoder := NewEncoder()
	data, err := encoder.Encode(program)
	if err != nil {
		t.Fatalf("Encode error: %v", err)
	}

	decoder := NewDecoder(data)
	decoded, err := decoder.Decode()
	if err != nil {
		t.Fatalf("Decode error: %v", err)
	}

	varDecl := decoded.Statements[0].(*ast.VariableDecl)
	boolLit := varDecl.Value.(*ast.BooleanLiteral)
	if !boolLit.Value {
		t.Error("Expected true, got false")
	}
}

func TestEncodeDecodeList(t *testing.T) {
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.VariableDecl{
				Name:       "nums",
				IsConstant: false,
				Value: &ast.ListLiteral{
					Elements: []ast.Expression{
						&ast.NumberLiteral{Value: 1},
						&ast.NumberLiteral{Value: 2},
						&ast.NumberLiteral{Value: 3},
					},
				},
			},
		},
	}

	encoder := NewEncoder()
	data, err := encoder.Encode(program)
	if err != nil {
		t.Fatalf("Encode error: %v", err)
	}

	decoder := NewDecoder(data)
	decoded, err := decoder.Decode()
	if err != nil {
		t.Fatalf("Decode error: %v", err)
	}

	varDecl := decoded.Statements[0].(*ast.VariableDecl)
	listLit := varDecl.Value.(*ast.ListLiteral)
	if len(listLit.Elements) != 3 {
		t.Errorf("Expected 3 elements, got %d", len(listLit.Elements))
	}
}

func TestEncodeDecodeAssignment(t *testing.T) {
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.Assignment{
				Name:  "x",
				Value: &ast.NumberLiteral{Value: 10},
			},
		},
	}

	encoder := NewEncoder()
	data, err := encoder.Encode(program)
	if err != nil {
		t.Fatalf("Encode error: %v", err)
	}

	decoder := NewDecoder(data)
	decoded, err := decoder.Decode()
	if err != nil {
		t.Fatalf("Decode error: %v", err)
	}

	assign := decoded.Statements[0].(*ast.Assignment)
	if assign.Name != "x" {
		t.Errorf("Expected name 'x', got %q", assign.Name)
	}
}

func TestEncodeDecodeFunctionDecl(t *testing.T) {
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.FunctionDecl{
				Name:       "add",
				Parameters: []string{"a", "b"},
				Body: []ast.Statement{
					&ast.ReturnStatement{
						Value: &ast.BinaryExpression{
							Left:     &ast.Identifier{Name: "a"},
							Operator: "+",
							Right:    &ast.Identifier{Name: "b"},
						},
					},
				},
			},
		},
	}

	encoder := NewEncoder()
	data, err := encoder.Encode(program)
	if err != nil {
		t.Fatalf("Encode error: %v", err)
	}

	decoder := NewDecoder(data)
	decoded, err := decoder.Decode()
	if err != nil {
		t.Fatalf("Decode error: %v", err)
	}

	funcDecl := decoded.Statements[0].(*ast.FunctionDecl)
	if funcDecl.Name != "add" {
		t.Errorf("Expected name 'add', got %q", funcDecl.Name)
	}
	if len(funcDecl.Parameters) != 2 {
		t.Errorf("Expected 2 parameters, got %d", len(funcDecl.Parameters))
	}
	if len(funcDecl.Body) != 1 {
		t.Errorf("Expected 1 body statement, got %d", len(funcDecl.Body))
	}
}

func TestEncodeDecodeIfStatement(t *testing.T) {
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.IfStatement{
				Condition: &ast.BinaryExpression{
					Left:     &ast.Identifier{Name: "x"},
					Operator: "is equal to",
					Right:    &ast.NumberLiteral{Value: 5},
				},
				Then: []ast.Statement{
					&ast.OutputStatement{Value: &ast.StringLiteral{Value: "yes"}},
				},
				Else: []ast.Statement{
					&ast.OutputStatement{Value: &ast.StringLiteral{Value: "no"}},
				},
			},
		},
	}

	encoder := NewEncoder()
	data, err := encoder.Encode(program)
	if err != nil {
		t.Fatalf("Encode error: %v", err)
	}

	decoder := NewDecoder(data)
	decoded, err := decoder.Decode()
	if err != nil {
		t.Fatalf("Decode error: %v", err)
	}

	ifStmt := decoded.Statements[0].(*ast.IfStatement)
	if len(ifStmt.Then) != 1 {
		t.Errorf("Expected 1 then statement, got %d", len(ifStmt.Then))
	}
	if len(ifStmt.Else) != 1 {
		t.Errorf("Expected 1 else statement, got %d", len(ifStmt.Else))
	}
}

func TestEncodeDecodeIfElseIf(t *testing.T) {
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.IfStatement{
				Condition: &ast.BooleanLiteral{Value: false},
				Then:      []ast.Statement{},
				ElseIf: []*ast.ElseIfPart{
					{
						Condition: &ast.BooleanLiteral{Value: true},
						Body:      []ast.Statement{&ast.OutputStatement{Value: &ast.StringLiteral{Value: "else if"}}},
					},
				},
				Else: []ast.Statement{},
			},
		},
	}

	encoder := NewEncoder()
	data, err := encoder.Encode(program)
	if err != nil {
		t.Fatalf("Encode error: %v", err)
	}

	decoder := NewDecoder(data)
	decoded, err := decoder.Decode()
	if err != nil {
		t.Fatalf("Decode error: %v", err)
	}

	ifStmt := decoded.Statements[0].(*ast.IfStatement)
	if len(ifStmt.ElseIf) != 1 {
		t.Errorf("Expected 1 else-if part, got %d", len(ifStmt.ElseIf))
	}
}

func TestEncodeDecodeWhileLoop(t *testing.T) {
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.WhileLoop{
				Condition: &ast.BinaryExpression{
					Left:     &ast.Identifier{Name: "x"},
					Operator: "is less than",
					Right:    &ast.NumberLiteral{Value: 10},
				},
				Body: []ast.Statement{
					&ast.Assignment{Name: "x", Value: &ast.BinaryExpression{
						Left:     &ast.Identifier{Name: "x"},
						Operator: "+",
						Right:    &ast.NumberLiteral{Value: 1},
					}},
				},
			},
		},
	}

	encoder := NewEncoder()
	data, err := encoder.Encode(program)
	if err != nil {
		t.Fatalf("Encode error: %v", err)
	}

	decoder := NewDecoder(data)
	decoded, err := decoder.Decode()
	if err != nil {
		t.Fatalf("Decode error: %v", err)
	}

	whileLoop := decoded.Statements[0].(*ast.WhileLoop)
	if len(whileLoop.Body) != 1 {
		t.Errorf("Expected 1 body statement, got %d", len(whileLoop.Body))
	}
}

func TestEncodeDecodeForLoop(t *testing.T) {
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.ForLoop{
				Count: &ast.NumberLiteral{Value: 5},
				Body: []ast.Statement{
					&ast.OutputStatement{Value: &ast.StringLiteral{Value: "hello"}},
				},
			},
		},
	}

	encoder := NewEncoder()
	data, err := encoder.Encode(program)
	if err != nil {
		t.Fatalf("Encode error: %v", err)
	}

	decoder := NewDecoder(data)
	decoded, err := decoder.Decode()
	if err != nil {
		t.Fatalf("Decode error: %v", err)
	}

	forLoop := decoded.Statements[0].(*ast.ForLoop)
	if len(forLoop.Body) != 1 {
		t.Errorf("Expected 1 body statement, got %d", len(forLoop.Body))
	}
}

func TestEncodeDecodeForEachLoop(t *testing.T) {
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.ForEachLoop{
				Item: "item",
				List: &ast.Identifier{Name: "myList"},
				Body: []ast.Statement{
					&ast.OutputStatement{Value: &ast.Identifier{Name: "item"}},
				},
			},
		},
	}

	encoder := NewEncoder()
	data, err := encoder.Encode(program)
	if err != nil {
		t.Fatalf("Encode error: %v", err)
	}

	decoder := NewDecoder(data)
	decoded, err := decoder.Decode()
	if err != nil {
		t.Fatalf("Decode error: %v", err)
	}

	forEachLoop := decoded.Statements[0].(*ast.ForEachLoop)
	if forEachLoop.Item != "item" {
		t.Errorf("Expected item 'item', got %q", forEachLoop.Item)
	}
}

func TestEncodeDecodeCallStatement(t *testing.T) {
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.CallStatement{
				FunctionCall: &ast.FunctionCall{
					Name:      "greet",
					Arguments: []ast.Expression{&ast.StringLiteral{Value: "World"}},
				},
			},
		},
	}

	encoder := NewEncoder()
	data, err := encoder.Encode(program)
	if err != nil {
		t.Fatalf("Encode error: %v", err)
	}

	decoder := NewDecoder(data)
	decoded, err := decoder.Decode()
	if err != nil {
		t.Fatalf("Decode error: %v", err)
	}

	callStmt := decoded.Statements[0].(*ast.CallStatement)
	if callStmt.FunctionCall.Name != "greet" {
		t.Errorf("Expected name 'greet', got %q", callStmt.FunctionCall.Name)
	}
}

func TestEncodeDecodeToggle(t *testing.T) {
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.ToggleStatement{Name: "flag"},
		},
	}

	encoder := NewEncoder()
	data, err := encoder.Encode(program)
	if err != nil {
		t.Fatalf("Encode error: %v", err)
	}

	decoder := NewDecoder(data)
	decoded, err := decoder.Decode()
	if err != nil {
		t.Fatalf("Decode error: %v", err)
	}

	toggle := decoded.Statements[0].(*ast.ToggleStatement)
	if toggle.Name != "flag" {
		t.Errorf("Expected name 'flag', got %q", toggle.Name)
	}
}

func TestEncodeDecodeUnaryExpression(t *testing.T) {
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.VariableDecl{
				Name:       "negative",
				IsConstant: false,
				Value: &ast.UnaryExpression{
					Operator: "-",
					Right:    &ast.NumberLiteral{Value: 5},
				},
			},
		},
	}

	encoder := NewEncoder()
	data, err := encoder.Encode(program)
	if err != nil {
		t.Fatalf("Encode error: %v", err)
	}

	decoder := NewDecoder(data)
	decoded, err := decoder.Decode()
	if err != nil {
		t.Fatalf("Decode error: %v", err)
	}

	varDecl := decoded.Statements[0].(*ast.VariableDecl)
	unary := varDecl.Value.(*ast.UnaryExpression)
	if unary.Operator != "-" {
		t.Errorf("Expected operator '-', got %q", unary.Operator)
	}
}

func TestEncodeDecodeIndexExpression(t *testing.T) {
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.OutputStatement{
				Value: &ast.IndexExpression{
					List:  &ast.Identifier{Name: "myList"},
					Index: &ast.NumberLiteral{Value: 0},
				},
			},
		},
	}

	encoder := NewEncoder()
	data, err := encoder.Encode(program)
	if err != nil {
		t.Fatalf("Encode error: %v", err)
	}

	decoder := NewDecoder(data)
	decoded, err := decoder.Decode()
	if err != nil {
		t.Fatalf("Decode error: %v", err)
	}

	output := decoded.Statements[0].(*ast.OutputStatement)
	indexExpr := output.Value.(*ast.IndexExpression)
	ident := indexExpr.List.(*ast.Identifier)
	if ident.Name != "myList" {
		t.Errorf("Expected list name 'myList', got %q", ident.Name)
	}
}

func TestEncodeDecodeIndexAssignment(t *testing.T) {
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.IndexAssignment{
				ListName: "myList",
				Index:    &ast.NumberLiteral{Value: 0},
				Value:    &ast.NumberLiteral{Value: 42},
			},
		},
	}

	encoder := NewEncoder()
	data, err := encoder.Encode(program)
	if err != nil {
		t.Fatalf("Encode error: %v", err)
	}

	decoder := NewDecoder(data)
	decoded, err := decoder.Decode()
	if err != nil {
		t.Fatalf("Decode error: %v", err)
	}

	indexAssign := decoded.Statements[0].(*ast.IndexAssignment)
	if indexAssign.ListName != "myList" {
		t.Errorf("Expected list name 'myList', got %q", indexAssign.ListName)
	}
}

func TestEncodeDecodeLengthExpression(t *testing.T) {
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.OutputStatement{
				Value: &ast.LengthExpression{
					List: &ast.Identifier{Name: "myList"},
				},
			},
		},
	}

	encoder := NewEncoder()
	data, err := encoder.Encode(program)
	if err != nil {
		t.Fatalf("Encode error: %v", err)
	}

	decoder := NewDecoder(data)
	decoded, err := decoder.Decode()
	if err != nil {
		t.Fatalf("Decode error: %v", err)
	}

	output := decoded.Statements[0].(*ast.OutputStatement)
	lengthExpr := output.Value.(*ast.LengthExpression)
	ident := lengthExpr.List.(*ast.Identifier)
	if ident.Name != "myList" {
		t.Errorf("Expected list name 'myList', got %q", ident.Name)
	}
}

func TestEncodeDecodeLocationExpression(t *testing.T) {
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.OutputStatement{
				Value: &ast.LocationExpression{Name: "x"},
			},
		},
	}

	encoder := NewEncoder()
	data, err := encoder.Encode(program)
	if err != nil {
		t.Fatalf("Encode error: %v", err)
	}

	decoder := NewDecoder(data)
	decoded, err := decoder.Decode()
	if err != nil {
		t.Fatalf("Decode error: %v", err)
	}

	output := decoded.Statements[0].(*ast.OutputStatement)
	locExpr := output.Value.(*ast.LocationExpression)
	if locExpr.Name != "x" {
		t.Errorf("Expected name 'x', got %q", locExpr.Name)
	}
}

func TestInvalidMagicBytes(t *testing.T) {
	invalidData := []byte{0xFF, 0xFF, 0xFF, 0xFF, 0x01}
	decoder := NewDecoder(invalidData)
	_, err := decoder.Decode()
	if err == nil {
		t.Error("Expected error for invalid magic bytes")
	}
}

func TestInvalidVersion(t *testing.T) {
	invalidData := append(MagicBytes, 0xFF) // Invalid version
	decoder := NewDecoder(invalidData)
	_, err := decoder.Decode()
	if err == nil {
		t.Error("Expected error for invalid version")
	}
}

func TestFunctionCallExpression(t *testing.T) {
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.VariableDecl{
				Name:       "result",
				IsConstant: false,
				Value: &ast.FunctionCall{
					Name:      "add",
					Arguments: []ast.Expression{&ast.NumberLiteral{Value: 1}, &ast.NumberLiteral{Value: 2}},
				},
			},
		},
	}

	encoder := NewEncoder()
	data, err := encoder.Encode(program)
	if err != nil {
		t.Fatalf("Encode error: %v", err)
	}

	decoder := NewDecoder(data)
	decoded, err := decoder.Decode()
	if err != nil {
		t.Fatalf("Decode error: %v", err)
	}

	varDecl := decoded.Statements[0].(*ast.VariableDecl)
	funcCall := varDecl.Value.(*ast.FunctionCall)
	if funcCall.Name != "add" {
		t.Errorf("Expected function name 'add', got %q", funcCall.Name)
	}
	if len(funcCall.Arguments) != 2 {
		t.Errorf("Expected 2 arguments, got %d", len(funcCall.Arguments))
	}
}
