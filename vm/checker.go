package vm

import (
	"english/ast"
	"english/vm/types"
	"fmt"
)

// builtinArgTypes maps function name to the expected TypeKind of each positional argument.
// TypeUnknown means "any type accepted for this slot".
var builtinArgTypes = map[string][]types.TypeKind{
	// text-only functions
	"uppercase":         {types.TypeString},
	"lowercase":         {types.TypeString},
	"casefold":          {types.TypeString},
	"trim":              {types.TypeString},
	"trim_left":         {types.TypeString},
	"trim_right":        {types.TypeString},
	"split":             {types.TypeString},
	"replace":           {types.TypeString},
	"contains":          {types.TypeString},
	"starts_with":       {types.TypeString},
	"ends_with":         {types.TypeString},
	"index_of":          {types.TypeString},
	"substring":         {types.TypeString},
	"str_repeat":        {types.TypeString},
	"count_occurrences": {types.TypeString},
	"pad_left":          {types.TypeString},
	"pad_right":         {types.TypeString},
	"title":             {types.TypeString},
	"capitalize":        {types.TypeString},
	"swapcase":          {types.TypeString},
	"is_digit":          {types.TypeString},
	"is_alpha":          {types.TypeString},
	"is_alnum":          {types.TypeString},
	"is_space":          {types.TypeString},
	"is_upper":          {types.TypeString},
	"is_lower":          {types.TypeString},
	"center":            {types.TypeString},
	"zfill":             {types.TypeString},
	"to_number":         {types.TypeString},
	// number-only functions
	"is_integer": {types.TypeF64},
	"clamp":      {types.TypeF64},
	"sign":       {types.TypeF64},
	// list-only functions
	"average":    {types.TypeList},
	"min_value":  {types.TypeList},
	"max_value":  {types.TypeList},
	"any_true":   {types.TypeList},
	"all_true":   {types.TypeList},
	"product":    {types.TypeList},
	"sorted_desc": {types.TypeList},
	"zip_with":   {types.TypeList},
	"sort":       {types.TypeList},
	"reverse":    {types.TypeList},
	"sum":        {types.TypeList},
	"unique":     {types.TypeList},
	"first":      {types.TypeList},
	"last":       {types.TypeList},
	"flatten":    {types.TypeList},
	"slice":      {types.TypeList},
	// lookup-table-only functions
	"keys":           {types.TypeLookup},
	"values":         {types.TypeLookup},
	"table_remove":   {types.TypeLookup},
	"table_has":      {types.TypeLookup},
	"merge":          {types.TypeLookup},
	"get_or_default": {types.TypeLookup},
}

// TypeError represents a compile-time type error.
type TypeError struct {
	Line    int
	Message string
}

func (te *TypeError) Error() string {
	if te.Line > 0 {
		return fmt.Sprintf("TypeError at line %d: %s", te.Line, te.Message)
	}
	return fmt.Sprintf("TypeError: %s", te.Message)
}

// TypeChecker performs static type analysis on an AST before execution.
type TypeChecker struct {
	varTypes map[string]types.TypeKind
	errors   []*TypeError
}

// Check runs the type checker on a program and returns all type errors found.
func Check(program *ast.Program) []*TypeError {
	tc := &TypeChecker{varTypes: make(map[string]types.TypeKind)}
	tc.checkStatements(program.Statements)
	return tc.errors
}

func (tc *TypeChecker) error(line int, format string, args ...interface{}) {
	tc.errors = append(tc.errors, &TypeError{Line: line, Message: fmt.Sprintf(format, args...)})
}

// exprType infers the static TypeKind of an expression.
// Returns TypeUnknown when the type cannot be statically determined.
func (tc *TypeChecker) exprType(expr ast.Expression) types.TypeKind {
	switch e := expr.(type) {
	case *ast.NumberLiteral:
		return types.TypeF64
	case *ast.StringLiteral:
		return types.TypeString
	case *ast.BooleanLiteral:
		return types.TypeBool
	case *ast.ListLiteral:
		return types.TypeList
	case *ast.Identifier:
		if tk, ok := tc.varTypes[e.Name]; ok {
			return tk
		}
	}
	return types.TypeUnknown
}

func (tc *TypeChecker) checkStatements(stmts []ast.Statement) {
	for _, stmt := range stmts {
		tc.checkStatement(stmt)
	}
}

func (tc *TypeChecker) checkStatement(stmt ast.Statement) {
	switch s := stmt.(type) {
	case *ast.VariableDecl:
		if s.Value != nil {
			tk := tc.exprType(s.Value)
			if tk != types.TypeUnknown {
				tc.varTypes[s.Name] = tk
			}
			tc.checkExpression(s.Value)
		}
	case *ast.TypedVariableDecl:
		declaredKind := types.Parse(s.TypeName)
		if declaredKind != types.TypeUnknown {
			tc.varTypes[s.Name] = types.Canonical(declaredKind)
		}
		if s.Value != nil {
			actualKind := tc.exprType(s.Value)
			if actualKind != types.TypeUnknown && declaredKind != types.TypeUnknown {
				if types.Canonical(actualKind) != types.Canonical(declaredKind) {
					tc.error(0, "cannot initialize %s with %s", types.Name(declaredKind), types.Name(actualKind))
				}
			}
			tc.checkExpression(s.Value)
		}
	case *ast.IfStatement:
		tc.checkExpression(s.Condition)
		tc.checkStatements(s.Then)
		for _, elif := range s.ElseIf {
			tc.checkExpression(elif.Condition)
			tc.checkStatements(elif.Body)
		}
		if s.Else != nil {
			tc.checkStatements(s.Else)
		}
	case *ast.WhileLoop:
		tc.checkExpression(s.Condition)
		tc.checkStatements(s.Body)
	case *ast.ForLoop:
		tc.checkExpression(s.Count)
		tc.checkStatements(s.Body)
	case *ast.ForEachLoop:
		tc.checkStatements(s.Body)
	case *ast.TryStatement:
		tc.checkStatements(s.TryBody)
		tc.checkStatements(s.ErrorBody)
		tc.checkStatements(s.FinallyBody)
	case *ast.OutputStatement:
		for _, arg := range s.Values {
			tc.checkExpression(arg)
		}
	case *ast.Assignment:
		tc.checkExpression(s.Value)
	case *ast.CallStatement:
		if s.FunctionCall != nil {
			tc.checkFunctionCallArgs(s.FunctionCall.Name, s.FunctionCall.Arguments, 0)
			for _, arg := range s.FunctionCall.Arguments {
				tc.checkExpression(arg)
			}
		}
	case *ast.ReturnStatement:
		if s.Value != nil {
			tc.checkExpression(s.Value)
		}
	case *ast.FunctionDecl:
		tc.checkStatements(s.Body)
	}
}

func (tc *TypeChecker) checkExpression(expr ast.Expression) {
	switch e := expr.(type) {
	case *ast.FunctionCall:
		tc.checkFunctionCallArgs(e.Name, e.Arguments, 0)
		for _, arg := range e.Arguments {
			tc.checkExpression(arg)
		}
	case *ast.MethodCall:
		allArgs := append([]ast.Expression{e.Object}, e.Arguments...)
		tc.checkFunctionCallArgs(e.MethodName, allArgs, 0)
		tc.checkExpression(e.Object)
		for _, arg := range e.Arguments {
			tc.checkExpression(arg)
		}
	case *ast.BinaryExpression:
		tc.checkExpression(e.Left)
		tc.checkExpression(e.Right)
	case *ast.UnaryExpression:
		tc.checkExpression(e.Right)
	case *ast.IndexExpression:
		tc.checkExpression(e.List)
		tc.checkExpression(e.Index)
	case *ast.LengthExpression:
		tc.checkExpression(e.List)
	case *ast.NilCheckExpression:
		tc.checkExpression(e.Value)
	case *ast.CastExpression:
		tc.checkExpression(e.Value)
	case *ast.ErrorTypeCheckExpression:
		tc.checkExpression(e.Value)
	}
}

func (tc *TypeChecker) checkFunctionCallArgs(name string, args []ast.Expression, _ int) {
	expected, ok := builtinArgTypes[name]
	if !ok {
		return
	}
	for i, expectedKind := range expected {
		if i >= len(args) {
			break
		}
		if expectedKind == types.TypeUnknown {
			continue
		}
		actualKind := tc.exprType(args[i])
		if actualKind == types.TypeUnknown {
			continue
		}
		if types.Canonical(actualKind) != types.Canonical(expectedKind) {
			tc.error(0, "%s expects %s, got %s", name, types.Name(expectedKind), types.Name(actualKind))
		}
	}
}
