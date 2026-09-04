// Package ast defines the Abstract Syntax Tree node types
// for the English programming language.
package ast

import "github.com/Advik-B/english/types"

// Position identifies a location in source text.
//
// Line and Col are 1-based; Offset is the byte offset of the first character.
// Col and Offset count bytes, not runes.
type Position struct {
	Line   int
	Col    int
	Offset int
}

// IsKnown reports whether the position was actually recorded.
func (p Position) IsKnown() bool { return p.Line > 0 }

// Base carries a node's source position. Every node embeds it — expressions
// included, which previously had no position at all, so an error inside an
// expression could never be pointed at more precisely than the enclosing
// statement's line, and never at a column.
//
// Position's fields are promoted through Base, so node.Line keeps working.
type Base struct {
	Position
	// inferred is the type the checker determined for this node. It is
	// meaningful only for expressions, and is unexported so that it can only
	// be reached through the Expression interface rather than set by accident
	// in a composite literal.
	inferred *types.TypeInfo
}

// Pos returns the node's source position.
func (b Base) Pos() Position { return b.Position }

// InferredType returns the type the checker determined for this expression,
// or nil if it has not been checked or the type could not be determined.
func (b *Base) InferredType() *types.TypeInfo { return b.inferred }

// SetInferredType records the type the checker determined for this expression.
func (b *Base) SetInferredType(t *types.TypeInfo) { b.inferred = t }

// At builds a Base for the given position.
func At(line, col, offset int) Base {
	return Base{Position: Position{Line: line, Col: col, Offset: offset}}
}

// TypeExpr is a type annotation as written in the source: the type in
// "Declare count as number", a struct field's type, a "cast to" target, or an
// array literal's element type.
//
// Those four positions used to hold a bare string, which meant the name was
// re-parsed with types.Parse every time it was needed — at run time, on every
// struct instantiation and every cast — and carried no position, so nothing
// could point at a bad annotation. Kind is resolved once, when the annotation
// is parsed.
type TypeExpr struct {
	Base
	// Name is the annotation as written, so that a struct name keeps the case
	// it was declared with.
	Name string
	// Kind is the built-in type Name resolves to, or types.TypeUnknown when it
	// names something only the type checker can resolve, such as a struct.
	Kind types.TypeKind
}

func (te *TypeExpr) node() {}

// String renders the annotation as written.
func (te *TypeExpr) String() string {
	if te == nil {
		return ""
	}
	return te.Name
}

// IsBuiltin reports whether the annotation names a built-in type.
func (te *TypeExpr) IsBuiltin() bool {
	return te != nil && te.Kind != types.TypeUnknown
}

// NewTypeExpr builds an annotation from a name alone, resolving its kind.
// Used by tools that recover a program from bytecode, where only the name of
// the annotation survives.
func NewTypeExpr(name string) *TypeExpr {
	if name == "" {
		return nil
	}
	return &TypeExpr{Name: name, Kind: types.Parse(name)}
}

// TypeName returns the annotation's name, or "" when there is no annotation.
func TypeName(te *TypeExpr) string {
	if te == nil {
		return ""
	}
	return te.Name
}

// Node is the base interface for all AST nodes
type Node interface {
	node()
	// Pos reports where the node starts in the source.
	Pos() Position
}

// Statement is the interface for all statement nodes
type Statement interface {
	Node
	statementNode()
}

// Expression is the interface for all expression nodes.
//
// Every expression carries a slot for the type the checker infers for it, so
// that the type of a subexpression is computed once and is available to the
// compiler, the transpiler and the editor rather than being re-derived by each.
type Expression interface {
	Node
	expressionNode()
	// InferredType is the type the checker determined, or nil if unchecked.
	InferredType() *types.TypeInfo
	// SetInferredType records the checker's result.
	SetInferredType(*types.TypeInfo)
}

// Program is the root node of the AST
type Program struct {
	Base
	Statements []Statement

	// PolitenessStats is populated by the parser when it processes statements.
	// PoliteCount is the number of top-level statements prefixed with a
	// politeness marker (please / kindly / could you / would you kindly).
	// TotalCount is the total number of countable top-level statements
	// (comments are excluded).  ImpoliteLines records the source line of
	// every impolite statement for use in error reporting.
	PoliteCount   int
	TotalCount    int
	ImpoliteLines []int
}

func (p *Program) node() {}

// ImportStatement represents an import statement
type ImportStatement struct {
	Base
	Path      string   // The file path to import
	Items     []string // Specific items to import (empty means import all)
	ImportAll bool     // True for "import everything/all"
	IsSafe    bool     // True for safe imports (don't run top-level code)
}

func (is *ImportStatement) node()          {}
func (is *ImportStatement) statementNode() {}

// VariableDecl represents a variable declaration
type VariableDecl struct {
	Base
	Name       string
	IsConstant bool
	Value      Expression
}

func (vd *VariableDecl) node()          {}
func (vd *VariableDecl) statementNode() {}

// Assignment represents a variable assignment
type Assignment struct {
	Base
	Name  string
	Value Expression
}

func (a *Assignment) node()          {}
func (a *Assignment) statementNode() {}

// Param is one function parameter.
//
// Type is nil for a parameter written without an annotation. Parameters were
// previously a plain []string, so a function had no signature at all: nothing
// could check an argument's type at a call site, or a return against what the
// function claims to give back.
type Param struct {
	Base
	Name string
	Type *TypeExpr
}

// FunctionDecl represents a function declaration
type FunctionDecl struct {
	Base
	Name   string
	Params []Param
	// ReturnType is what the function gives back, or nil when it is
	// unannotated or returns nothing.
	ReturnType *TypeExpr
	Body       []Statement
}

// ParamNames returns the parameter names in order. The runtime binds arguments
// by name and does not need the annotations.
func (fd *FunctionDecl) ParamNames() []string {
	if len(fd.Params) == 0 {
		return nil
	}
	names := make([]string, len(fd.Params))
	for i, p := range fd.Params {
		names[i] = p.Name
	}
	return names
}

// ParamsFromNames builds unannotated parameters from names alone, for tools
// that recover a program from bytecode where annotations do not survive.
func ParamsFromNames(names []string) []Param {
	if len(names) == 0 {
		return nil
	}
	params := make([]Param, len(names))
	for i, n := range names {
		params[i] = Param{Name: n}
	}
	return params
}

func (fd *FunctionDecl) node()          {}
func (fd *FunctionDecl) statementNode() {}

// FunctionCall represents a function call expression
type FunctionCall struct {
	Base
	Name      string
	Arguments []Expression
}

func (fc *FunctionCall) node()           {}
func (fc *FunctionCall) expressionNode() {}

// CallStatement represents a function call as a statement
type CallStatement struct {
	Base
	FunctionCall *FunctionCall
	MethodCall   *MethodCall
}

func (cs *CallStatement) node()          {}
func (cs *CallStatement) statementNode() {}

// IfStatement represents an if-then-else statement
type IfStatement struct {
	Base
	Condition Expression
	Then      []Statement
	ElseIf    []*ElseIfPart
	Else      []Statement
}

func (is *IfStatement) node()          {}
func (is *IfStatement) statementNode() {}

// ElseIfPart represents an else-if branch
type ElseIfPart struct {
	Base
	Condition Expression
	Body      []Statement
}

// WhileLoop represents a while loop
type WhileLoop struct {
	Base
	Condition Expression
	Body      []Statement
}

func (wl *WhileLoop) node()          {}
func (wl *WhileLoop) statementNode() {}

// ForLoop represents a counted for loop
type ForLoop struct {
	Base
	Count Expression
	Body  []Statement
}

func (fl *ForLoop) node()          {}
func (fl *ForLoop) statementNode() {}

// ForEachLoop represents a for-each loop over a collection
type ForEachLoop struct {
	Base
	Item string
	List Expression
	Body []Statement
}

func (fel *ForEachLoop) node()          {}
func (fel *ForEachLoop) statementNode() {}

// NumberLiteral represents a numeric literal
type NumberLiteral struct {
	Base
	Value float64
}

func (nl *NumberLiteral) node()           {}
func (nl *NumberLiteral) expressionNode() {}

// StringLiteral represents a string literal
type StringLiteral struct {
	Base
	Value string
}

func (sl *StringLiteral) node()           {}
func (sl *StringLiteral) expressionNode() {}

// ListLiteral represents a list/array literal
type ListLiteral struct {
	Base
	Elements []Expression
}

func (ll *ListLiteral) node()           {}
func (ll *ListLiteral) expressionNode() {}

// RangeLiteral represents a range expression like [1 .. 30] or a range from 1 to 30
// Optionally supports custom step: [1 .. 10 by 2] or a range from 1 to 10 by 2
type RangeLiteral struct {
	Base
	Start Expression
	End   Expression
	Step  Expression // optional, nil for default step of 1 or -1
}

func (rl *RangeLiteral) node()           {}
func (rl *RangeLiteral) expressionNode() {}

// Identifier represents a variable reference
type Identifier struct {
	Base
	Name string
}

func (i *Identifier) node()           {}
func (i *Identifier) expressionNode() {}

// BinaryExpression represents a binary operation (e.g., a + b)
type BinaryExpression struct {
	Base
	Left     Expression
	Operator string
	Right    Expression
}

func (be *BinaryExpression) node()           {}
func (be *BinaryExpression) expressionNode() {}

// UnaryExpression represents a unary operation (e.g., -x)
type UnaryExpression struct {
	Base
	Operator string
	Right    Expression
}

func (ue *UnaryExpression) node()           {}
func (ue *UnaryExpression) expressionNode() {}

// IndexExpression represents array indexing (e.g., list[0])
type IndexExpression struct {
	Base
	List  Expression
	Index Expression
}

func (ie *IndexExpression) node()           {}
func (ie *IndexExpression) expressionNode() {}

// IndexAssignment represents assigning to an array index
type IndexAssignment struct {
	Base
	ListName string
	Index    Expression
	Value    Expression
}

func (ia *IndexAssignment) node()          {}
func (ia *IndexAssignment) statementNode() {}

// LengthExpression represents getting the length of a list or string
type LengthExpression struct {
	Base
	List Expression
}

func (le *LengthExpression) node()           {}
func (le *LengthExpression) expressionNode() {}

// ReturnStatement represents a return statement
type ReturnStatement struct {
	Base
	Value Expression
}

func (rs *ReturnStatement) node()          {}
func (rs *ReturnStatement) statementNode() {}

// OutputStatement represents a print statement
type OutputStatement struct {
	Base
	Values  []Expression
	Newline bool // true for Print, false for Write
}

func (os *OutputStatement) node()          {}
func (os *OutputStatement) statementNode() {}

// ToggleStatement toggles a boolean variable
type ToggleStatement struct {
	Base
	Name string
}

func (ts *ToggleStatement) node()          {}
func (ts *ToggleStatement) statementNode() {}

// BreakStatement breaks out of a loop
type BreakStatement struct{ Base }

func (bs *BreakStatement) node()          {}
func (bs *BreakStatement) statementNode() {}

// BooleanLiteral represents a boolean literal (true/false)
type BooleanLiteral struct {
	Base
	Value bool
}

func (bl *BooleanLiteral) node()           {}
func (bl *BooleanLiteral) expressionNode() {}

// LocationExpression returns the memory address of a variable
type LocationExpression struct {
	Base
	Name string
}

func (le *LocationExpression) node()           {}
func (le *LocationExpression) expressionNode() {}

// StructDecl represents a struct type declaration
type StructDecl struct {
	Base
	Name    string
	Fields  []*StructField
	Methods []*FunctionDecl
}

func (sd *StructDecl) node()          {}
func (sd *StructDecl) statementNode() {}

// StructField represents a field in a struct definition
type StructField struct {
	Base
	Name         string
	Type         *TypeExpr
	DefaultValue Expression
}

// StructInstantiation creates a new instance of a struct
type StructInstantiation struct {
	Base
	StructName  string
	FieldValues map[string]Expression
	FieldOrder  []string // Maintain field order
}

func (si *StructInstantiation) node()           {}
func (si *StructInstantiation) expressionNode() {}

// FieldAccess accesses a field of a struct
type FieldAccess struct {
	Base
	Object Expression
	Field  string
}

func (fa *FieldAccess) node()           {}
func (fa *FieldAccess) expressionNode() {}

// FieldAssignment assigns a value to a struct field
type FieldAssignment struct {
	Base
	ObjectName string
	Field      string
	Value      Expression
}

func (fa *FieldAssignment) node()          {}
func (fa *FieldAssignment) statementNode() {}

// TryStatement represents try/error/finally block
type TryStatement struct {
	Base
	TryBody     []Statement
	ErrorVar    string // Variable name to bind the error to
	ErrorType   string // If non-empty, only catch errors of this type
	ErrorBody   []Statement
	FinallyBody []Statement
}

func (ts *TryStatement) node()          {}
func (ts *TryStatement) statementNode() {}

// RaiseStatement raises an error
type RaiseStatement struct {
	Base
	Message   Expression
	ErrorType string // Optional error type
}

func (rs *RaiseStatement) node()          {}
func (rs *RaiseStatement) statementNode() {}

// ErrorTypeDecl declares a custom error type.
// Syntax:
//
//	Declare NetworkError as an error type.
//	Declare CustomErr1 as a type of NetworkError.
type ErrorTypeDecl struct {
	Base
	Name       string
	ParentType string // empty for root error types; parent name for subtypes
}

func (etd *ErrorTypeDecl) node()          {}
func (etd *ErrorTypeDecl) statementNode() {}

// TypeExpression gets the type of a value
type TypeExpression struct {
	Base
	Value Expression
}

func (te *TypeExpression) node()           {}
func (te *TypeExpression) expressionNode() {}

// CastExpression casts a value to a type
type CastExpression struct {
	Base
	Value Expression
	Type  *TypeExpr
}

func (ce *CastExpression) node()           {}
func (ce *CastExpression) expressionNode() {}

// ReferenceExpression creates a reference to a variable
type ReferenceExpression struct {
	Base
	Name string
}

func (re *ReferenceExpression) node()           {}
func (re *ReferenceExpression) expressionNode() {}

// CopyExpression creates a copy of a value
type CopyExpression struct {
	Base
	Value Expression
}

func (ce *CopyExpression) node()           {}
func (ce *CopyExpression) expressionNode() {}

// SwapStatement swaps two variables
type SwapStatement struct {
	Base
	Name1 string
	Name2 string
}

func (ss *SwapStatement) node()          {}
func (ss *SwapStatement) statementNode() {}

// ContinueStatement skips the rest of the current loop iteration
type ContinueStatement struct{ Base }

func (cs *ContinueStatement) node()          {}
func (cs *ContinueStatement) statementNode() {}

// NothingLiteral represents a null/nil value (nothing, none, null)
type NothingLiteral struct{ Base }

func (nl *NothingLiteral) node()           {}
func (nl *NothingLiteral) expressionNode() {}

// AskExpression reads a line of user input after displaying an optional prompt
type AskExpression struct {
	Base
	Prompt Expression // optional prompt to display
}

func (ae *AskExpression) node()           {}
func (ae *AskExpression) expressionNode() {}

// ArrayLiteral is a typed homogeneous array literal: "an array of number [1, 2, 3]"
type ArrayLiteral struct {
	Base
	// ElemType is the declared element type, or nil to infer it from the
	// elements.
	ElemType *TypeExpr
	Elements []Expression
}

func (al *ArrayLiteral) node()           {}
func (al *ArrayLiteral) expressionNode() {}

// LookupTableLiteral creates an empty lookup table: "a lookup table"
type LookupTableLiteral struct{ Base }

func (lt *LookupTableLiteral) node()           {}
func (lt *LookupTableLiteral) expressionNode() {}

// LookupKeyAccess reads a value from a lookup table: "TABLE at KEY" or "the entry KEY in TABLE"
type LookupKeyAccess struct {
	Base
	Table Expression
	Key   Expression
}

func (la *LookupKeyAccess) node()           {}
func (la *LookupKeyAccess) expressionNode() {}

// LookupKeyAssignment sets a value in a lookup table: "Set TABLE at KEY to be VALUE." or "Set the entry KEY in TABLE to be VALUE."
type LookupKeyAssignment struct {
	Base
	TableName string
	Key       Expression
	Value     Expression
}

func (la *LookupKeyAssignment) node()          {}
func (la *LookupKeyAssignment) statementNode() {}

// HasExpression checks whether a lookup table contains a key: "TABLE has KEY"
type HasExpression struct {
	Base
	Table Expression
	Key   Expression
}

func (he *HasExpression) node()           {}
func (he *HasExpression) expressionNode() {}

// NilCheckExpression checks whether a value is something (not nil) or nothing (nil).
//   - IsSomethingCheck == true  → "x is something" / "x has a value"  (returns true when x != nil)
//   - IsSomethingCheck == false → "x is nothing"   / "x has no value" (returns true when x == nil)
type NilCheckExpression struct {
	Base
	Value            Expression
	IsSomethingCheck bool // true = "is something"; false = "is nothing"
}

func (nc *NilCheckExpression) node()           {}
func (nc *NilCheckExpression) expressionNode() {}

// TypedVariableDecl represents a variable declaration with explicit type
type TypedVariableDecl struct {
	Base
	Name       string
	Type       *TypeExpr
	IsConstant bool
	Value      Expression
}

func (tvd *TypedVariableDecl) node()          {}
func (tvd *TypedVariableDecl) statementNode() {}

// MethodCall represents calling a method on an object
type MethodCall struct {
	Base
	Object     Expression
	MethodName string
	Arguments  []Expression
}

func (mc *MethodCall) node()           {}
func (mc *MethodCall) expressionNode() {}

// ErrorTypeCheckExpression checks whether an error value's type matches a named
// error type (including inherited types).
// Syntax: error is NetworkError
type ErrorTypeCheckExpression struct {
	Base
	Value    Expression
	TypeName string
}

func (etc *ErrorTypeCheckExpression) node()           {}
func (etc *ErrorTypeCheckExpression) expressionNode() {}

// CommentStatement carries a source comment through to the output.
// Text holds the comment body (everything after the leading '#', trimmed).
type CommentStatement struct {
	Base
	Text string
}

func (cs *CommentStatement) node()          {}
func (cs *CommentStatement) statementNode() {}

// setPos records a position on a node unless it already carries one.
// Promoted through Base, so every node has it.
func (b *Base) setPos(p Position) {
	if !b.Position.IsKnown() {
		b.Position = p
	}
}

// SetPosIfUnknown records pos on n unless n already has a position.
//
// The parser calls this once per precedence layer rather than threading a
// position through all ~66 node constructions, so no expression can be built
// without one. Expression nodes previously had no position field at all, which
// is why the type checker reported every argument-type error at "line 0".
func SetPosIfUnknown(n Node, pos Position) {
	if n == nil {
		return
	}
	if s, ok := n.(interface{ setPos(Position) }); ok {
		s.setPos(pos)
	}
}
