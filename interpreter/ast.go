package interpreter

// Statement is the interface for all statement nodes
type Statement interface {
	statementNode()
}

// Expression is the interface for all expression nodes
type Expression interface {
	expressionNode()
}

// Program is the root node
type Program struct {
	Statements []Statement
}

// Variable Declaration/Assignment Statements

type VariableDecl struct {
	Name       string
	IsConstant bool
	Value      Expression
}

func (vd *VariableDecl) statementNode() {}

type Assignment struct {
	Name  string
	Value Expression
}

func (a *Assignment) statementNode() {}

// Function Declaration

type FunctionDecl struct {
	Name       string
	Parameters []string
	Body       []Statement
}

func (fd *FunctionDecl) statementNode() {}

// Function Call

type FunctionCall struct {
	Name      string
	Arguments []Expression
}

func (fc *FunctionCall) expressionNode() {}

type CallStatement struct {
	FunctionCall *FunctionCall
}

func (cs *CallStatement) statementNode() {}

// Control Flow

type IfStatement struct {
	Condition Expression
	Then      []Statement
	ElseIf    []*ElseIfPart
	Else      []Statement
}

func (is *IfStatement) statementNode() {}

type ElseIfPart struct {
	Condition Expression
	Body      []Statement
}

type WhileLoop struct {
	Condition Expression
	Body      []Statement
}

func (wl *WhileLoop) statementNode() {}

type ForLoop struct {
	Count Expression
	Body  []Statement
}

func (fl *ForLoop) statementNode() {}

type ForEachLoop struct {
	Item string
	List Expression
	Body []Statement
}

func (fel *ForEachLoop) statementNode() {}

// Expressions

type NumberLiteral struct {
	Value float64
}

func (nl *NumberLiteral) expressionNode() {}

type StringLiteral struct {
	Value string
}

func (sl *StringLiteral) expressionNode() {}

type ListLiteral struct {
	Elements []Expression
}

func (ll *ListLiteral) expressionNode() {}

type Identifier struct {
	Name string
}

func (i *Identifier) expressionNode() {}

type BinaryExpression struct {
	Left     Expression
	Operator string
	Right    Expression
}

func (be *BinaryExpression) expressionNode() {}

type UnaryExpression struct {
	Operator string
	Right    Expression
}

func (ue *UnaryExpression) expressionNode() {}

// IndexExpression represents array indexing like list[0] or "the item at position 0 in list"
type IndexExpression struct {
	List  Expression
	Index Expression
}

func (ie *IndexExpression) expressionNode() {}

// IndexAssignment represents assigning to an array index
type IndexAssignment struct {
	ListName string
	Index    Expression
	Value    Expression
}

func (ia *IndexAssignment) statementNode() {}

// LengthExpression represents getting the length of a list
type LengthExpression struct {
	List Expression
}

func (le *LengthExpression) expressionNode() {}

// Return Statement

type ReturnStatement struct {
	Value Expression
}

func (rs *ReturnStatement) statementNode() {}

// Output Statement

type OutputStatement struct {
	Value Expression
}

func (os *OutputStatement) statementNode() {}

// Toggle Statement - toggles a boolean variable
type ToggleStatement struct {
	Name string
}

func (ts *ToggleStatement) statementNode() {}

// Boolean Literal
type BooleanLiteral struct {
	Value bool
}

func (bl *BooleanLiteral) expressionNode() {}

// Location Expression - returns memory address/id of a variable
type LocationExpression struct {
	Name string
}

func (le *LocationExpression) expressionNode() {}
