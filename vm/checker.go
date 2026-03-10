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
	"average":     {types.TypeList},
	"min_value":   {types.TypeList},
	"max_value":   {types.TypeList},
	"any_true":    {types.TypeList},
	"all_true":    {types.TypeList},
	"product":     {types.TypeList},
	"sorted_desc": {types.TypeList},
	"zip_with":    {types.TypeList},
	"sort":        {types.TypeList},
	"reverse":     {types.TypeList},
	"sum":         {types.TypeList},
	"unique":      {types.TypeList},
	"first":       {types.TypeList},
	"last":        {types.TypeList},
	"flatten":     {types.TypeList},
	"slice":       {types.TypeList},
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
	varTypes      map[string]types.TypeKind
	userFunctions map[string]bool // names of user-defined functions (skip stdlib type check)
	errors        []*TypeError
	// scopeStack is a stack of per-scope declared variable sets.
	// Each entry maps variable name → line of declaration (0 = predefined by stdlib).
	// Duplicate detection is limited to the innermost matching scope.
	scopeStack []map[string]int
}

// Check runs the type checker on a program and returns all type errors found.
// Provide the names of any stdlib-predefined variables via predefines so the
// checker can report redeclarations as compile-time errors.
func Check(program *ast.Program, predefines ...string) []*TypeError {
	globalScope := make(map[string]int)
	for _, name := range predefines {
		globalScope[name] = 0 // 0 = predefined (no source line)
	}
	tc := &TypeChecker{
		varTypes:      make(map[string]types.TypeKind),
		userFunctions: make(map[string]bool),
		scopeStack:    []map[string]int{globalScope},
	}
	// Pre-scan top-level function declarations so that user-defined functions
	// sharing a name with a stdlib function are not falsely type-checked.
	for _, stmt := range program.Statements {
		if fn, ok := stmt.(*ast.FunctionDecl); ok {
			tc.userFunctions[fn.Name] = true
		}
	}
	tc.checkStatements(program.Statements)
	return tc.errors
}

func (tc *TypeChecker) error(line int, format string, args ...interface{}) {
	tc.errors = append(tc.errors, &TypeError{Line: line, Message: fmt.Sprintf(format, args...)})
}

// pushScope opens a new lexical scope for duplicate-variable detection.
func (tc *TypeChecker) pushScope() {
	tc.scopeStack = append(tc.scopeStack, make(map[string]int))
}

// popScope closes the innermost lexical scope.
func (tc *TypeChecker) popScope() {
	if len(tc.scopeStack) > 0 {
		tc.scopeStack = tc.scopeStack[:len(tc.scopeStack)-1]
	}
}

// declareVar records a variable declaration in the current innermost scope.
// If the name is already declared in that same scope (or is a predefined
// stdlib constant in the global scope) an error is emitted and false is returned.
func (tc *TypeChecker) declareVar(name string, line int) {
	if len(tc.scopeStack) == 0 {
		return
	}
	current := tc.scopeStack[len(tc.scopeStack)-1]
	if prevLine, exists := current[name]; exists {
		if prevLine == 0 {
			tc.error(line, "variable '%s' shadows a predefined constant", name)
		} else {
			tc.error(line, "variable '%s' is already declared at line %d", name, prevLine)
		}
		return
	}
	current[name] = line
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
		tc.declareVar(s.Name, s.Line)
		if s.Value != nil {
			tk := tc.exprType(s.Value)
			if tk != types.TypeUnknown {
				tc.varTypes[s.Name] = tk
			}
			tc.checkExpression(s.Value)
		}
	case *ast.TypedVariableDecl:
		tc.declareVar(s.Name, s.Line)
		declaredKind := types.Parse(s.TypeName)
		if declaredKind != types.TypeUnknown {
			tc.varTypes[s.Name] = types.Canonical(declaredKind)
		}
		if s.Value != nil {
			actualKind := tc.exprType(s.Value)
			if actualKind != types.TypeUnknown && declaredKind != types.TypeUnknown {
				if types.Canonical(actualKind) != types.Canonical(declaredKind) {
					tc.error(s.Line, "cannot initialize %s with %s", types.Name(declaredKind), types.Name(actualKind))
				}
			}
			tc.checkExpression(s.Value)
		}
	case *ast.IfStatement:
		tc.checkExpression(s.Condition)
		tc.pushScope()
		tc.checkStatements(s.Then)
		tc.popScope()
		for _, elif := range s.ElseIf {
			tc.checkExpression(elif.Condition)
			tc.pushScope()
			tc.checkStatements(elif.Body)
			tc.popScope()
		}
		if s.Else != nil {
			tc.pushScope()
			tc.checkStatements(s.Else)
			tc.popScope()
		}
	case *ast.WhileLoop:
		tc.checkExpression(s.Condition)
		tc.pushScope()
		tc.checkStatements(s.Body)
		tc.popScope()
	case *ast.ForLoop:
		tc.checkExpression(s.Count)
		tc.pushScope()
		tc.checkStatements(s.Body)
		tc.popScope()
	case *ast.ForEachLoop:
		tc.pushScope()
		if s.Item != "" {
			tc.declareVar(s.Item, s.Line)
		}
		tc.checkStatements(s.Body)
		tc.popScope()
	case *ast.TryStatement:
		tc.pushScope()
		tc.checkStatements(s.TryBody)
		tc.popScope()
		tc.pushScope()
		tc.checkStatements(s.ErrorBody)
		tc.popScope()
		tc.pushScope()
		tc.checkStatements(s.FinallyBody)
		tc.popScope()
	case *ast.OutputStatement:
		for _, arg := range s.Values {
			tc.checkExpression(arg)
		}
	case *ast.Assignment:
		tc.checkExpression(s.Value)
	case *ast.CallStatement:
		if s.FunctionCall != nil {
			tc.checkFunctionCallArgs(s.FunctionCall.Name, s.FunctionCall.Arguments)
			for _, arg := range s.FunctionCall.Arguments {
				tc.checkExpression(arg)
			}
		}
	case *ast.ReturnStatement:
		if s.Value != nil {
			tc.checkExpression(s.Value)
		}
	case *ast.FunctionDecl:
		tc.pushScope()
		tc.checkStatements(s.Body)
		tc.popScope()
	}
}

func (tc *TypeChecker) checkExpression(expr ast.Expression) {
	switch e := expr.(type) {
	case *ast.FunctionCall:
		tc.checkFunctionCallArgs(e.Name, e.Arguments)
		for _, arg := range e.Arguments {
			tc.checkExpression(arg)
		}
	case *ast.MethodCall:
		allArgs := append([]ast.Expression{e.Object}, e.Arguments...)
		tc.checkFunctionCallArgs(e.MethodName, allArgs)
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

func (tc *TypeChecker) checkFunctionCallArgs(name string, args []ast.Expression) {
	// User-defined functions shadow stdlib functions; skip the stdlib type check.
	if tc.userFunctions[name] {
		return
	}
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
