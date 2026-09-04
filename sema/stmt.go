package sema

import (
	"github.com/Advik-B/english/ast"
	"github.com/Advik-B/english/types"
)

func (a *Analyzer) checkStatements(stmts []ast.Statement) {
	for _, stmt := range stmts {
		a.checkStatement(stmt)
	}
}

func (a *Analyzer) checkStatement(stmt ast.Statement) {
	switch s := stmt.(type) {
	case *ast.CommentStatement, *ast.ImportStatement:
		// Nothing to check; imports are resolved during collection.

	case *ast.VariableDecl:
		t := a.checkExpr(s.Value)
		a.declare(s.Name, t, s.IsConstant, s.Pos(), true)

	case *ast.TypedVariableDecl:
		a.checkTypedDecl(s)

	case *ast.Assignment:
		a.checkAssignment(s)

	case *ast.IndexAssignment:
		a.checkIndexAssignment(s)

	case *ast.LookupKeyAssignment:
		a.checkLookupAssignment(s)

	case *ast.FieldAssignment:
		a.checkFieldAssignment(s)

	case *ast.OutputStatement:
		for _, v := range s.Values {
			a.checkExpr(v)
		}

	case *ast.CallStatement:
		// A statement is where a function that gives back nothing is meant to
		// be called, so its result is not wanted here.
		prev := a.discardingResult
		a.discardingResult = true
		switch {
		case s.MethodCall != nil:
			a.checkExpr(s.MethodCall)
		case s.FunctionCall != nil:
			a.checkExpr(s.FunctionCall)
		}
		a.discardingResult = prev

	case *ast.ReturnStatement:
		a.checkReturn(s)

	case *ast.FunctionDecl:
		a.checkFunctionBody(s, nil)

	case *ast.StructDecl:
		a.checkStructDecl(s)

	case *ast.ErrorTypeDecl:
		if s.ParentType != "" && !a.errorTypes[s.ParentType] {
			a.errorAt(s.Pos(), "'%s' is not a declared error type", s.ParentType)
		}

	case *ast.IfStatement:
		a.requireBool(s.Condition, "a condition")
		a.inScope(func() { a.checkStatements(s.Then) })
		for _, elif := range s.ElseIf {
			a.requireBool(elif.Condition, "a condition")
			a.inScope(func() { a.checkStatements(elif.Body) })
		}
		if s.Else != nil {
			a.inScope(func() { a.checkStatements(s.Else) })
		}

	case *ast.WhileLoop:
		a.requireBool(s.Condition, "a loop condition")
		a.inLoop(func() { a.checkStatements(s.Body) })

	case *ast.ForLoop:
		a.requireNumber(s.Count, "a repeat count")
		a.inLoop(func() { a.checkStatements(s.Body) })

	case *ast.ForEachLoop:
		a.checkForEach(s)

	case *ast.BreakStatement:
		if a.loopDepth == 0 {
			a.errorAt(s.Pos(), "'break' is only allowed inside a loop")
		}

	case *ast.ContinueStatement:
		if a.loopDepth == 0 {
			a.errorAt(s.Pos(), "'continue' is only allowed inside a loop")
		}

	case *ast.ToggleStatement:
		t := a.useName(s.Name, s.Pos())
		if t != nil && t.Kind != types.TypeUnknown && t.Kind != types.TypeBool {
			a.errorAt(s.Pos(), "cannot toggle '%s': it is %s, not a boolean", s.Name, describe(t))
		}

	case *ast.SwapStatement:
		a.checkSwap(s)

	case *ast.RaiseStatement:
		a.checkExpr(s.Message)
		a.useErrorType(s.ErrorType, s.Pos())

	case *ast.TryStatement:
		a.checkTry(s)
	}
}

// inScope runs fn in a fresh lexical scope.
func (a *Analyzer) inScope(fn func()) {
	a.pushScope()
	fn()
	a.popScope()
}

// inLoop runs fn in a fresh scope, with break and continue permitted.
func (a *Analyzer) inLoop(fn func()) {
	a.loopDepth++
	a.inScope(fn)
	a.loopDepth--
}

// ─── Declarations and assignment ─────────────────────────────────────────────

func (a *Analyzer) checkTypedDecl(s *ast.TypedVariableDecl) {
	declared := a.resolve(s.Type)
	if s.Value != nil {
		actual := a.checkExpr(s.Value)
		if declared != nil && !assignable(declared, actual) {
			a.errorWithHint(s.Value.Pos(),
				"Use 'cast to' if you meant to convert it.",
				"cannot initialize %s with %s", describe(declared), describe(actual))
		}
	}
	a.declare(s.Name, declared, s.IsConstant, s.Pos(), s.Value != nil)
}

func (a *Analyzer) checkAssignment(s *ast.Assignment) {
	value := a.checkExpr(s.Value)

	sym, ok := a.scope.lookup(s.Name)
	if !ok {
		// Assigning to a name that was never declared used to create it,
		// silently, in both engines: a typo in a Set statement introduced a
		// new variable rather than reporting a mistake.
		hint := "Declare it first: 'Declare " + s.Name + " to be …'."
		if similar := a.similarName(s.Name); similar != "" {
			hint = "Perhaps you meant '" + similar + "'."
		}
		a.errorWithHint(s.Pos(), hint, "'%s' is not declared", s.Name)
		return
	}
	if sym.IsConst {
		a.errorWithHint(s.Pos(), "Constants are declared with 'to always be'.",
			"cannot reassign '%s': it is a constant", s.Name)
		return
	}
	if !assignable(sym.Type, value) {
		a.errorWithHint(s.Value.Pos(),
			"Use 'cast to' if you meant to convert it.",
			"cannot assign %s to '%s', which is %s",
			describe(value), s.Name, describe(sym.Type))
		return
	}
	sym.Assigned = true
}

func (a *Analyzer) checkIndexAssignment(s *ast.IndexAssignment) {
	container := a.useName(s.ListName, s.Pos())
	a.requireNumber(s.Index, "an index")
	value := a.checkExpr(s.Value)

	if container == nil || container.Kind == types.TypeUnknown {
		return
	}
	switch container.Kind {
	case types.TypeList:
		// A list holds values of mixed type.
	case types.TypeArray:
		if container.ElementType != nil && !assignable(container.ElementType, value) {
			a.errorAt(s.Value.Pos(), "cannot put %s into an array of %s",
				describe(value), describe(container.ElementType))
		}
	default:
		a.errorAt(s.Pos(), "cannot assign into %s", describe(container))
	}
}

func (a *Analyzer) checkLookupAssignment(s *ast.LookupKeyAssignment) {
	table := a.useName(s.TableName, s.Pos())
	a.checkExpr(s.Key)
	a.checkExpr(s.Value)
	if table != nil && table.Kind != types.TypeUnknown && table.Kind != types.TypeLookup {
		a.errorAt(s.Pos(), "'%s' is %s, not a lookup table", s.TableName, describe(table))
	}
}

func (a *Analyzer) checkFieldAssignment(s *ast.FieldAssignment) {
	obj := a.useName(s.ObjectName, s.Pos())
	value := a.checkExpr(s.Value)
	if obj == nil || obj.Kind != types.TypeStruct {
		if obj != nil && obj.Kind != types.TypeUnknown {
			a.errorAt(s.Pos(), "'%s' is %s, not a struct", s.ObjectName, describe(obj))
		}
		return
	}
	def, ok := a.structs[obj.Name]
	if !ok {
		return
	}
	field, known := def.Fields[s.Field]
	if !known {
		a.errorWithHint(s.Pos(), "Fields of "+def.Name+" are: "+joinNames(def.Order)+".",
			"struct '%s' has no field named '%s'", def.Name, s.Field)
		return
	}
	if want := a.resolve(field.Type); want != nil && !assignable(want, value) {
		a.errorAt(s.Value.Pos(), "field '%s' of %s is %s, but this is %s",
			s.Field, def.Name, describe(want), describe(value))
	}
}

func (a *Analyzer) checkSwap(s *ast.SwapStatement) {
	first := a.useName(s.Name1, s.Pos())
	second := a.useName(s.Name2, s.Pos())
	if first == nil || second == nil {
		return
	}
	if !assignable(first, second) || !assignable(second, first) {
		a.errorAt(s.Pos(), "cannot swap '%s' (%s) with '%s' (%s): they hold different types",
			s.Name1, describe(first), s.Name2, describe(second))
	}
}

// ─── Loops ───────────────────────────────────────────────────────────────────

func (a *Analyzer) checkForEach(s *ast.ForEachLoop) {
	collection := a.checkExpr(s.List)
	if collection != nil && collection.Kind != types.TypeUnknown {
		switch collection.Kind {
		case types.TypeList, types.TypeArray, types.TypeLookup, types.TypeString:
		default:
			a.errorAt(s.List.Pos(), "'for each' needs a list, an array or a lookup table, but this is %s",
				describe(collection))
		}
	}

	// The loop variable's type follows the collection's elements where known.
	var item *types.TypeInfo
	if collection != nil && collection.Kind == types.TypeArray {
		item = collection.ElementType
	}

	a.loopDepth++
	a.pushScope()
	if s.Item != "" {
		a.declare(s.Item, item, false, s.Pos(), true)
	}
	a.checkStatements(s.Body)
	a.popScope()
	a.loopDepth--
}

// ─── Functions ───────────────────────────────────────────────────────────────

// checkFunctionBody checks a function or method body with its parameters in
// scope. fields, when given, are a struct's fields, which a method body may
// refer to by bare name.
func (a *Analyzer) checkFunctionBody(fd *ast.FunctionDecl, fields *structDef) {
	prevReturn, prevIn, prevLoop := a.returnType, a.inFunction, a.loopDepth
	a.returnType, a.inFunction, a.loopDepth = fd.ReturnType, true, 0

	a.pushScope()
	if fields != nil {
		for _, name := range fields.Order {
			f := fields.Fields[name]
			a.declare(name, a.resolve(f.Type), false, f.Pos(), true)
		}
	}
	for _, p := range fd.Params {
		a.declare(p.Name, a.resolve(p.Type), false, p.Pos(), true)
	}
	a.checkStatements(fd.Body)
	a.popScope()

	// A function that promises a result must produce one on every path. One
	// that gives back nothing promises no such thing, and has no reason to end
	// in a Return at all.
	if givesAValue(fd.ReturnType) && !alwaysReturns(fd.Body) {
		a.errorWithHint(fd.Pos(),
			"Every path through the function must end in a 'Return'.",
			"'%s' promises to give back %s, but it can finish without returning a value",
			fd.Name, fd.ReturnType.Name)
	}

	a.returnType, a.inFunction, a.loopDepth = prevReturn, prevIn, prevLoop
}

func (a *Analyzer) checkReturn(s *ast.ReturnStatement) {
	value := a.checkExpr(s.Value)
	if !a.inFunction {
		a.errorAt(s.Pos(), "'Return' is only allowed inside a function")
		return
	}

	// A function that gives back nothing may still return, to finish early —
	// but it cannot return a value, because its callers were promised none.
	if !givesAValue(a.returnType) {
		if s.Value != nil {
			a.errorWithHint(s.Pos(),
				"Write 'Return.' on its own to finish early, or declare what the function gives back.",
				"this function gives back nothing, but this returns %s", describe(value))
		}
		return
	}

	declared := a.resolve(a.returnType)
	if declared == nil {
		return // the annotation itself was rejected; nothing to check against
	}
	if s.Value == nil {
		a.errorAt(s.Pos(), "this function gives back %s, but this returns no value",
			describe(declared))
		return
	}
	if !assignable(declared, value) {
		a.errorAt(s.Pos(), "this function gives back %s, but this returns %s",
			describe(declared), describe(value))
	}
}

// givesAValue reports whether a declared result type is one a caller can use.
// A function annotated "gives back nothing" produces none.
func givesAValue(te *ast.TypeExpr) bool {
	return te != nil && te.Kind != types.TypeNull
}

// alwaysReturns reports whether a block always ends by returning or raising.
//
// A loop with a literally true condition never falls through, and a try block
// counts only when both its body and its handler return.
func alwaysReturns(stmts []ast.Statement) bool {
	for i := len(stmts) - 1; i >= 0; i-- {
		switch s := stmts[i].(type) {
		case *ast.CommentStatement:
			continue
		case *ast.ReturnStatement, *ast.RaiseStatement:
			return true
		case *ast.IfStatement:
			if s.Else == nil || !alwaysReturns(s.Then) || !alwaysReturns(s.Else) {
				return false
			}
			for _, elif := range s.ElseIf {
				if !alwaysReturns(elif.Body) {
					return false
				}
			}
			return true
		case *ast.WhileLoop:
			if lit, ok := s.Condition.(*ast.BooleanLiteral); ok && lit.Value {
				return true // loops forever; control never reaches the end
			}
			return false
		case *ast.TryStatement:
			return alwaysReturns(s.TryBody) && alwaysReturns(s.ErrorBody)
		default:
			return false
		}
	}
	return false
}

// ─── Structs ─────────────────────────────────────────────────────────────────

func (a *Analyzer) checkStructDecl(s *ast.StructDecl) {
	def := a.structs[s.Name]
	if def == nil {
		return // a duplicate declaration, already reported
	}
	for _, f := range s.Fields {
		declared := a.resolve(f.Type)
		if f.DefaultValue == nil {
			continue
		}
		got := a.checkExpr(f.DefaultValue)
		if declared != nil && !assignable(declared, got) {
			a.errorAt(f.DefaultValue.Pos(),
				"field '%s' is %s, but its default is %s",
				f.Name, describe(declared), describe(got))
		}
	}
	for _, m := range s.Methods {
		a.checkFunctionBody(m, def)
	}
}

// ─── Error handling ──────────────────────────────────────────────────────────

func (a *Analyzer) checkTry(s *ast.TryStatement) {
	a.inScope(func() { a.checkStatements(s.TryBody) })

	a.useErrorType(s.ErrorType, s.Pos())
	a.pushScope()
	if s.ErrorVar != "" {
		a.declare(s.ErrorVar, types.InfoFor(types.TypeError), false, s.Pos(), true)
	}
	a.checkStatements(s.ErrorBody)
	a.popScope()

	a.inScope(func() { a.checkStatements(s.FinallyBody) })
}
