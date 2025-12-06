// Package ast defines the Abstract Syntax Tree node types
// for the English programming language.
package ast

// Node is the base interface for all AST nodes
type Node interface {
	node()
}

// Statement is the interface for all statement nodes
type Statement interface {
	Node
	statementNode()
}

// Expression is the interface for all expression nodes
type Expression interface {
	Node
	expressionNode()
}

// Program is the root node of the AST
type Program struct {
	Statements []Statement
}

func (p *Program) node() {}

// VariableDecl represents a variable declaration
type VariableDecl struct {
	Name       string
	IsConstant bool
	Value      Expression
}

func (vd *VariableDecl) node()          {}
func (vd *VariableDecl) statementNode() {}

// Assignment represents a variable assignment
type Assignment struct {
	Name  string
	Value Expression
}

func (a *Assignment) node()          {}
func (a *Assignment) statementNode() {}

// FunctionDecl represents a function declaration
type FunctionDecl struct {
	Name       string
	Parameters []string
	Body       []Statement
}

func (fd *FunctionDecl) node()          {}
func (fd *FunctionDecl) statementNode() {}

// FunctionCall represents a function call expression
type FunctionCall struct {
	Name      string
	Arguments []Expression
}

func (fc *FunctionCall) node()           {}
func (fc *FunctionCall) expressionNode() {}

// CallStatement represents a function call as a statement
type CallStatement struct {
	FunctionCall *FunctionCall
}

func (cs *CallStatement) node()          {}
func (cs *CallStatement) statementNode() {}

// IfStatement represents an if-then-else statement
type IfStatement struct {
	Condition Expression
	Then      []Statement
	ElseIf    []*ElseIfPart
	Else      []Statement
}

func (is *IfStatement) node()          {}
func (is *IfStatement) statementNode() {}

// ElseIfPart represents an else-if branch
type ElseIfPart struct {
	Condition Expression
	Body      []Statement
}

// WhileLoop represents a while loop
type WhileLoop struct {
	Condition Expression
	Body      []Statement
}

func (wl *WhileLoop) node()          {}
func (wl *WhileLoop) statementNode() {}

// ForLoop represents a counted for loop
type ForLoop struct {
	Count Expression
	Body  []Statement
}

func (fl *ForLoop) node()          {}
func (fl *ForLoop) statementNode() {}

// ForEachLoop represents a for-each loop over a collection
type ForEachLoop struct {
	Item string
	List Expression
	Body []Statement
}

func (fel *ForEachLoop) node()          {}
func (fel *ForEachLoop) statementNode() {}

// NumberLiteral represents a numeric literal
type NumberLiteral struct {
	Value float64
}

func (nl *NumberLiteral) node()           {}
func (nl *NumberLiteral) expressionNode() {}

// StringLiteral represents a string literal
type StringLiteral struct {
	Value string
}

func (sl *StringLiteral) node()           {}
func (sl *StringLiteral) expressionNode() {}

// ListLiteral represents a list/array literal
type ListLiteral struct {
	Elements []Expression
}

func (ll *ListLiteral) node()           {}
func (ll *ListLiteral) expressionNode() {}

// Identifier represents a variable reference
type Identifier struct {
	Name string
}

func (i *Identifier) node()           {}
func (i *Identifier) expressionNode() {}

// BinaryExpression represents a binary operation (e.g., a + b)
type BinaryExpression struct {
	Left     Expression
	Operator string
	Right    Expression
}

func (be *BinaryExpression) node()           {}
func (be *BinaryExpression) expressionNode() {}

// UnaryExpression represents a unary operation (e.g., -x)
type UnaryExpression struct {
	Operator string
	Right    Expression
}

func (ue *UnaryExpression) node()           {}
func (ue *UnaryExpression) expressionNode() {}

// IndexExpression represents array indexing (e.g., list[0])
type IndexExpression struct {
	List  Expression
	Index Expression
}

func (ie *IndexExpression) node()           {}
func (ie *IndexExpression) expressionNode() {}

// IndexAssignment represents assigning to an array index
type IndexAssignment struct {
	ListName string
	Index    Expression
	Value    Expression
}

func (ia *IndexAssignment) node()          {}
func (ia *IndexAssignment) statementNode() {}

// LengthExpression represents getting the length of a list or string
type LengthExpression struct {
	List Expression
}

func (le *LengthExpression) node()           {}
func (le *LengthExpression) expressionNode() {}

// ReturnStatement represents a return statement
type ReturnStatement struct {
	Value Expression
}

func (rs *ReturnStatement) node()          {}
func (rs *ReturnStatement) statementNode() {}

// OutputStatement represents a print statement
type OutputStatement struct {
	Values  []Expression
	Newline bool // true for Print, false for Write
}

func (os *OutputStatement) node()          {}
func (os *OutputStatement) statementNode() {}

// ToggleStatement toggles a boolean variable
type ToggleStatement struct {
	Name string
}

func (ts *ToggleStatement) node()          {}
func (ts *ToggleStatement) statementNode() {}

// BreakStatement breaks out of a loop
type BreakStatement struct{}

func (bs *BreakStatement) node()          {}
func (bs *BreakStatement) statementNode() {}

// BooleanLiteral represents a boolean literal (true/false)
type BooleanLiteral struct {
	Value bool
}

func (bl *BooleanLiteral) node()           {}
func (bl *BooleanLiteral) expressionNode() {}

// LocationExpression returns the memory address of a variable
type LocationExpression struct {
	Name string
}

func (le *LocationExpression) node()           {}
func (le *LocationExpression) expressionNode() {}

// StructDecl represents a struct type declaration
type StructDecl struct {
	Name         string
	Fields       []*StructField
	Methods      []*FunctionDecl
}

func (sd *StructDecl) node()          {}
func (sd *StructDecl) statementNode() {}

// StructField represents a field in a struct definition
type StructField struct {
	Name         string
	TypeName     string
	DefaultValue Expression
	IsUnsigned   bool
}

// StructInstantiation creates a new instance of a struct
type StructInstantiation struct {
	StructName   string
	FieldValues  map[string]Expression
	FieldOrder   []string // Maintain field order
}

func (si *StructInstantiation) node()           {}
func (si *StructInstantiation) expressionNode() {}

// FieldAccess accesses a field of a struct
type FieldAccess struct {
	Object Expression
	Field  string
}

func (fa *FieldAccess) node()           {}
func (fa *FieldAccess) expressionNode() {}

// FieldAssignment assigns a value to a struct field
type FieldAssignment struct {
	ObjectName string
	Field      string
	Value      Expression
}

func (fa *FieldAssignment) node()          {}
func (fa *FieldAssignment) statementNode() {}

// TryStatement represents try/error/finally block
type TryStatement struct {
	TryBody     []Statement
	ErrorVar    string // Variable name to bind the error to
	ErrorBody   []Statement
	FinallyBody []Statement
}

func (ts *TryStatement) node()          {}
func (ts *TryStatement) statementNode() {}

// RaiseStatement raises an error
type RaiseStatement struct {
	Message   Expression
	ErrorType string // Optional error type
}

func (rs *RaiseStatement) node()          {}
func (rs *RaiseStatement) statementNode() {}

// TypeExpression gets the type of a value
type TypeExpression struct {
	Value Expression
}

func (te *TypeExpression) node()           {}
func (te *TypeExpression) expressionNode() {}

// CastExpression casts a value to a type
type CastExpression struct {
	Value    Expression
	TypeName string
}

func (ce *CastExpression) node()           {}
func (ce *CastExpression) expressionNode() {}

// ReferenceExpression creates a reference to a variable
type ReferenceExpression struct {
	Name string
}

func (re *ReferenceExpression) node()           {}
func (re *ReferenceExpression) expressionNode() {}

// CopyExpression creates a copy of a value
type CopyExpression struct {
	Value Expression
}

func (ce *CopyExpression) node()           {}
func (ce *CopyExpression) expressionNode() {}

// SwapStatement swaps two variables
type SwapStatement struct {
	Name1 string
	Name2 string
}

func (ss *SwapStatement) node()          {}
func (ss *SwapStatement) statementNode() {}

// TypedVariableDecl represents a variable declaration with explicit type
type TypedVariableDecl struct {
	Name       string
	TypeName   string
	IsConstant bool
	Value      Expression
}

func (tvd *TypedVariableDecl) node()          {}
func (tvd *TypedVariableDecl) statementNode() {}

// MethodCall represents calling a method on an object
type MethodCall struct {
	Object     Expression
	MethodName string
	Arguments  []Expression
}

func (mc *MethodCall) node()           {}
func (mc *MethodCall) expressionNode() {}
