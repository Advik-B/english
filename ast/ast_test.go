package ast

import "testing"

// TestProgramNode tests the Program node
func TestProgramNode(t *testing.T) {
	program := &Program{
		Statements: []Statement{},
	}
	program.node()

	if program.Statements == nil {
		t.Error("Program.Statements should not be nil")
	}
}

// TestVariableDecl tests VariableDecl node
func TestVariableDecl(t *testing.T) {
	vd := &VariableDecl{
		Name:       "x",
		IsConstant: false,
		Value:      &NumberLiteral{Value: 5},
	}
	vd.node()
	vd.statementNode()

	if vd.Name != "x" {
		t.Errorf("VariableDecl.Name = %q, want \"x\"", vd.Name)
	}
	if vd.IsConstant {
		t.Error("VariableDecl.IsConstant should be false")
	}
}

// TestConstantDecl tests constant declaration
func TestConstantDecl(t *testing.T) {
	vd := &VariableDecl{
		Name:       "PI",
		IsConstant: true,
		Value:      &NumberLiteral{Value: 3.14159},
	}

	if !vd.IsConstant {
		t.Error("VariableDecl.IsConstant should be true for constants")
	}
}

// TestAssignment tests Assignment node
func TestAssignment(t *testing.T) {
	a := &Assignment{
		Name:  "x",
		Value: &NumberLiteral{Value: 10},
	}
	a.node()
	a.statementNode()

	if a.Name != "x" {
		t.Errorf("Assignment.Name = %q, want \"x\"", a.Name)
	}
}

// TestFunctionDecl tests FunctionDecl node
func TestFunctionDecl(t *testing.T) {
	fd := &FunctionDecl{
		Name:       "add",
		Parameters: []string{"a", "b"},
		Body:       []Statement{},
	}
	fd.node()
	fd.statementNode()

	if fd.Name != "add" {
		t.Errorf("FunctionDecl.Name = %q, want \"add\"", fd.Name)
	}
	if len(fd.Parameters) != 2 {
		t.Errorf("len(FunctionDecl.Parameters) = %d, want 2", len(fd.Parameters))
	}
}

// TestFunctionCall tests FunctionCall node
func TestFunctionCall(t *testing.T) {
	fc := &FunctionCall{
		Name:      "greet",
		Arguments: []Expression{&StringLiteral{Value: "Alice"}},
	}
	fc.node()
	fc.expressionNode()

	if fc.Name != "greet" {
		t.Errorf("FunctionCall.Name = %q, want \"greet\"", fc.Name)
	}
	if len(fc.Arguments) != 1 {
		t.Errorf("len(FunctionCall.Arguments) = %d, want 1", len(fc.Arguments))
	}
}

// TestCallStatement tests CallStatement node
func TestCallStatement(t *testing.T) {
	cs := &CallStatement{
		FunctionCall: &FunctionCall{Name: "test", Arguments: []Expression{}},
	}
	cs.node()
	cs.statementNode()

	if cs.FunctionCall.Name != "test" {
		t.Errorf("CallStatement.FunctionCall.Name = %q, want \"test\"", cs.FunctionCall.Name)
	}
}

// TestIfStatement tests IfStatement node
func TestIfStatement(t *testing.T) {
	is := &IfStatement{
		Condition: &BooleanLiteral{Value: true},
		Then:      []Statement{},
		ElseIf:    []*ElseIfPart{},
		Else:      []Statement{},
	}
	is.node()
	is.statementNode()

	if is.Condition == nil {
		t.Error("IfStatement.Condition should not be nil")
	}
}

// TestWhileLoop tests WhileLoop node
func TestWhileLoop(t *testing.T) {
	wl := &WhileLoop{
		Condition: &BooleanLiteral{Value: true},
		Body:      []Statement{},
	}
	wl.node()
	wl.statementNode()

	if wl.Condition == nil {
		t.Error("WhileLoop.Condition should not be nil")
	}
}

// TestForLoop tests ForLoop node
func TestForLoop(t *testing.T) {
	fl := &ForLoop{
		Count: &NumberLiteral{Value: 5},
		Body:  []Statement{},
	}
	fl.node()
	fl.statementNode()

	if fl.Count == nil {
		t.Error("ForLoop.Count should not be nil")
	}
}

// TestForEachLoop tests ForEachLoop node
func TestForEachLoop(t *testing.T) {
	fel := &ForEachLoop{
		Item: "item",
		List: &Identifier{Name: "myList"},
		Body: []Statement{},
	}
	fel.node()
	fel.statementNode()

	if fel.Item != "item" {
		t.Errorf("ForEachLoop.Item = %q, want \"item\"", fel.Item)
	}
}

// TestNumberLiteral tests NumberLiteral node
func TestNumberLiteral(t *testing.T) {
	nl := &NumberLiteral{Value: 42.5}
	nl.node()
	nl.expressionNode()

	if nl.Value != 42.5 {
		t.Errorf("NumberLiteral.Value = %v, want 42.5", nl.Value)
	}
}

// TestStringLiteral tests StringLiteral node
func TestStringLiteral(t *testing.T) {
	sl := &StringLiteral{Value: "hello"}
	sl.node()
	sl.expressionNode()

	if sl.Value != "hello" {
		t.Errorf("StringLiteral.Value = %q, want \"hello\"", sl.Value)
	}
}

// TestBooleanLiteral tests BooleanLiteral node
func TestBooleanLiteral(t *testing.T) {
	bl := &BooleanLiteral{Value: true}
	bl.node()
	bl.expressionNode()

	if !bl.Value {
		t.Error("BooleanLiteral.Value should be true")
	}
}

// TestListLiteral tests ListLiteral node
func TestListLiteral(t *testing.T) {
	ll := &ListLiteral{
		Elements: []Expression{
			&NumberLiteral{Value: 1},
			&NumberLiteral{Value: 2},
			&NumberLiteral{Value: 3},
		},
	}
	ll.node()
	ll.expressionNode()

	if len(ll.Elements) != 3 {
		t.Errorf("len(ListLiteral.Elements) = %d, want 3", len(ll.Elements))
	}
}

// TestIdentifier tests Identifier node
func TestIdentifier(t *testing.T) {
	id := &Identifier{Name: "myVar"}
	id.node()
	id.expressionNode()

	if id.Name != "myVar" {
		t.Errorf("Identifier.Name = %q, want \"myVar\"", id.Name)
	}
}

// TestBinaryExpression tests BinaryExpression node
func TestBinaryExpression(t *testing.T) {
	be := &BinaryExpression{
		Left:     &NumberLiteral{Value: 5},
		Operator: "+",
		Right:    &NumberLiteral{Value: 3},
	}
	be.node()
	be.expressionNode()

	if be.Operator != "+" {
		t.Errorf("BinaryExpression.Operator = %q, want \"+\"", be.Operator)
	}
}

// TestUnaryExpression tests UnaryExpression node
func TestUnaryExpression(t *testing.T) {
	ue := &UnaryExpression{
		Operator: "-",
		Right:    &NumberLiteral{Value: 5},
	}
	ue.node()
	ue.expressionNode()

	if ue.Operator != "-" {
		t.Errorf("UnaryExpression.Operator = %q, want \"-\"", ue.Operator)
	}
}

// TestIndexExpression tests IndexExpression node
func TestIndexExpression(t *testing.T) {
	ie := &IndexExpression{
		List:  &Identifier{Name: "myList"},
		Index: &NumberLiteral{Value: 0},
	}
	ie.node()
	ie.expressionNode()

	if ie.List == nil || ie.Index == nil {
		t.Error("IndexExpression fields should not be nil")
	}
}

// TestIndexAssignment tests IndexAssignment node
func TestIndexAssignment(t *testing.T) {
	ia := &IndexAssignment{
		ListName: "myList",
		Index:    &NumberLiteral{Value: 0},
		Value:    &NumberLiteral{Value: 42},
	}
	ia.node()
	ia.statementNode()

	if ia.ListName != "myList" {
		t.Errorf("IndexAssignment.ListName = %q, want \"myList\"", ia.ListName)
	}
}

// TestLengthExpression tests LengthExpression node
func TestLengthExpression(t *testing.T) {
	le := &LengthExpression{
		List: &Identifier{Name: "myList"},
	}
	le.node()
	le.expressionNode()

	if le.List == nil {
		t.Error("LengthExpression.List should not be nil")
	}
}

// TestReturnStatement tests ReturnStatement node
func TestReturnStatement(t *testing.T) {
	rs := &ReturnStatement{
		Value: &NumberLiteral{Value: 42},
	}
	rs.node()
	rs.statementNode()

	if rs.Value == nil {
		t.Error("ReturnStatement.Value should not be nil")
	}
}

// TestOutputStatement tests OutputStatement node
func TestOutputStatement(t *testing.T) {
	os := &OutputStatement{
		Value: &StringLiteral{Value: "Hello"},
	}
	os.node()
	os.statementNode()

	if os.Value == nil {
		t.Error("OutputStatement.Value should not be nil")
	}
}

// TestToggleStatement tests ToggleStatement node
func TestToggleStatement(t *testing.T) {
	ts := &ToggleStatement{
		Name: "isEnabled",
	}
	ts.node()
	ts.statementNode()

	if ts.Name != "isEnabled" {
		t.Errorf("ToggleStatement.Name = %q, want \"isEnabled\"", ts.Name)
	}
}

// TestLocationExpression tests LocationExpression node
func TestLocationExpression(t *testing.T) {
	le := &LocationExpression{
		Name: "x",
	}
	le.node()
	le.expressionNode()

	if le.Name != "x" {
		t.Errorf("LocationExpression.Name = %q, want \"x\"", le.Name)
	}
}

// TestElseIfPart tests ElseIfPart structure
func TestElseIfPart(t *testing.T) {
	eif := &ElseIfPart{
		Condition: &BooleanLiteral{Value: true},
		Body:      []Statement{},
	}

	if eif.Condition == nil {
		t.Error("ElseIfPart.Condition should not be nil")
	}
}
