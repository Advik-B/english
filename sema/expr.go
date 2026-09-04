package sema

import (
	"fmt"

	"github.com/Advik-B/english/ast"
	"github.com/Advik-B/english/stdlib"
	"github.com/Advik-B/english/types"
)

// checkExpr infers the type of an expression, checks it, records the result on
// the node and returns it.
//
// A nil result means the type could not be determined; callers must treat that
// as "do not complain", so that the checker only ever rejects what it can
// prove wrong.
func (a *Analyzer) checkExpr(expr ast.Expression) *types.TypeInfo {
	if expr == nil {
		return nil
	}
	// Checking is idempotent: the possessive fallback re-visits the object it
	// has already typed, and a node must not be reported against twice.
	if a.checked[expr] {
		return expr.InferredType()
	}
	a.checked[expr] = true
	t := a.inferExpr(expr)
	expr.SetInferredType(t)
	return t
}

func (a *Analyzer) inferExpr(expr ast.Expression) *types.TypeInfo {
	switch e := expr.(type) {
	case *ast.NumberLiteral:
		return types.InfoFor(types.TypeF64)
	case *ast.StringLiteral:
		return types.InfoFor(types.TypeString)
	case *ast.BooleanLiteral:
		return types.InfoFor(types.TypeBool)
	case *ast.NothingLiteral:
		return types.InfoFor(types.TypeNull)
	case *ast.AskExpression:
		a.checkExpr(e.Prompt)
		return types.InfoFor(types.TypeString)
	case *ast.LocationExpression:
		a.useName(e.Name, e.Pos())
		return types.InfoFor(types.TypeString)
	case *ast.TypeExpression:
		a.checkExpr(e.Value)
		return types.InfoFor(types.TypeString)

	case *ast.Identifier:
		return a.useName(e.Name, e.Pos())

	case *ast.ListLiteral:
		for _, el := range e.Elements {
			a.checkExpr(el)
		}
		return types.InfoFor(types.TypeList)

	case *ast.ArrayLiteral:
		return a.arrayLiteralType(e)

	case *ast.LookupTableLiteral:
		return types.InfoFor(types.TypeLookup)

	case *ast.RangeLiteral:
		a.requireNumber(e.Start, "a range bound")
		a.requireNumber(e.End, "a range bound")
		if e.Step != nil {
			a.requireNumber(e.Step, "a range step")
		}
		// A range is its own runtime value and is not one of the named types.
		return nil

	case *ast.BinaryExpression:
		return a.binaryType(e)

	case *ast.UnaryExpression:
		return a.unaryType(e)

	case *ast.IndexExpression:
		return a.indexType(e)

	case *ast.LengthExpression:
		a.requireCollection(e.List, "the length of")
		return types.InfoFor(types.TypeF64)

	case *ast.CastExpression:
		return a.castType(e)

	case *ast.CopyExpression:
		return a.checkExpr(e.Value)

	case *ast.ReferenceExpression:
		a.useName(e.Name, e.Pos())
		return types.InfoFor(types.TypeRef)

	case *ast.NilCheckExpression:
		a.checkExpr(e.Value)
		return types.InfoFor(types.TypeBool)

	case *ast.ErrorTypeCheckExpression:
		// Only an error can be of an error type. "x is NetworkError" parses
		// for any x at all, and for anything but an error the answer is always
		// false — which is not a comparison anyone writes on purpose.
		a.requireKind(e.Value, types.TypeError, "an error type check")
		a.useErrorType(e.TypeName, e.Pos())
		return types.InfoFor(types.TypeBool)

	case *ast.HasExpression:
		a.requireKind(e.Table, types.TypeLookup, "'has'")
		a.checkExpr(e.Key)
		return types.InfoFor(types.TypeBool)

	case *ast.LookupKeyAccess:
		a.requireKind(e.Table, types.TypeLookup, "a key lookup")
		a.checkExpr(e.Key)
		// The value type of a lookup table is not tracked.
		return nil

	case *ast.FunctionCall:
		return a.callType(e.Name, e.Arguments, e.Pos())

	case *ast.MethodCall:
		return a.methodCallType(e)

	case *ast.FieldAccess:
		return a.fieldAccessType(e)

	case *ast.StructInstantiation:
		return a.structInstantiationType(e)
	}
	return nil
}

// useName resolves a name, reporting one that was never declared.
func (a *Analyzer) useName(name string, pos ast.Position) *types.TypeInfo {
	sym, ok := a.scope.lookup(name)
	if !ok {
		// A name that matches a function is a function value being referred to
		// without calling it, which the language has no syntax for; report it
		// as undefined rather than inventing a type.
		if hint := a.similarName(name); hint != "" {
			a.errorWithHint(pos, fmt.Sprintf("Perhaps you meant '%s'.", hint),
				"'%s' is not declared", name)
		} else {
			a.errorAt(pos, "'%s' is not declared", name)
		}
		return nil
	}
	if !sym.Assigned {
		a.errorWithHint(pos,
			"Give it a value before reading it, for example with 'Set "+name+" to be …'.",
			"'%s' is declared but has no value yet", name)
		// Report once; treat it as assigned from here on.
		sym.Assigned = true
	}
	return sym.Type
}

// similarName finds a declared name close to the one given, for a suggestion.
func (a *Analyzer) similarName(name string) string {
	best := ""
	bestDist := 3 // only suggest something genuinely close
	for cur := a.scope; cur != nil; cur = cur.parent {
		for candidate := range cur.names {
			if d := editDistance(name, candidate); d < bestDist ||
				(d == bestDist && candidate < best) {
				best, bestDist = candidate, d
			}
		}
	}
	return best
}

// builtinErrorTypes are the error types the runtime itself raises, which a
// program may catch without declaring them.
var builtinErrorTypes = map[string]bool{
	"error":              true, // the catch-all
	"RuntimeError":       true,
	"TypeError":          true,
	"KeyError":           true,
	"StackOverflowError": true,
}

func (a *Analyzer) useErrorType(name string, pos ast.Position) {
	if name == "" || a.errorTypes[name] || builtinErrorTypes[name] {
		return
	}
	a.errorWithHint(pos, "Declare it first: 'Declare "+name+" as an error type.'",
		"'%s' is not a declared error type", name)
}

// ─── Composite literals ──────────────────────────────────────────────────────

func (a *Analyzer) arrayLiteralType(e *ast.ArrayLiteral) *types.TypeInfo {
	declared := a.resolve(e.ElemType)
	var elem *types.TypeInfo
	if declared != nil {
		elem = declared
	}
	for _, el := range e.Elements {
		t := a.checkExpr(el)
		if t == nil || t.Kind == types.TypeNull {
			continue
		}
		if elem == nil {
			elem = t
			continue
		}
		if !assignable(elem, t) {
			a.errorAt(el.Pos(), "an array of %s cannot hold %s", describe(elem), describe(t))
		}
	}
	info := &types.TypeInfo{Kind: types.TypeArray, Name: "array"}
	if elem != nil {
		info.ElementType = elem
	}
	return info
}

// ─── Operators ───────────────────────────────────────────────────────────────

// binaryRule is one accepted operand pairing for an operator.
type binaryRule struct {
	left, right, result types.TypeKind
}

// binaryRules is the operator type table. It replaces the previous situation,
// where no operator was checked at all: "a" + 1 compiled cleanly and failed at
// run time.
var binaryRules = map[string][]binaryRule{
	"+": {
		{types.TypeF64, types.TypeF64, types.TypeF64},
		{types.TypeString, types.TypeString, types.TypeString},
		{types.TypeArray, types.TypeArray, types.TypeArray},
	},
	"-": {{types.TypeF64, types.TypeF64, types.TypeF64}},
	"*": {{types.TypeF64, types.TypeF64, types.TypeF64}},
	"/": {{types.TypeF64, types.TypeF64, types.TypeF64}},
	"%": {{types.TypeF64, types.TypeF64, types.TypeF64}},

	"is less than":                {{types.TypeF64, types.TypeF64, types.TypeBool}},
	"is greater than":             {{types.TypeF64, types.TypeF64, types.TypeBool}},
	"is less than or equal to":    {{types.TypeF64, types.TypeF64, types.TypeBool}},
	"is greater than or equal to": {{types.TypeF64, types.TypeF64, types.TypeBool}},

	"and": {{types.TypeBool, types.TypeBool, types.TypeBool}},
	"or":  {{types.TypeBool, types.TypeBool, types.TypeBool}},
}

func (a *Analyzer) binaryType(e *ast.BinaryExpression) *types.TypeInfo {
	left := a.checkExpr(e.Left)
	right := a.checkExpr(e.Right)

	switch e.Operator {
	case "is equal to", "is not equal to":
		// Equality accepts any two values of the same type. Comparing values
		// of different types is always false, so it is a mistake rather than
		// a useful expression.
		if left != nil && right != nil &&
			left.Kind != types.TypeUnknown && right.Kind != types.TypeUnknown &&
			left.Kind != types.TypeNull && right.Kind != types.TypeNull &&
			!assignable(left, right) && !assignable(right, left) {
			a.errorWithHint(e.Pos(),
				"Use 'cast to' if you meant to convert one of them first.",
				"cannot compare %s with %s: they are never equal",
				describe(left), describe(right))
		}
		return types.InfoFor(types.TypeBool)
	}

	rules, known := binaryRules[e.Operator]
	if !known {
		return nil
	}
	// Unknown operand types mean the operator cannot be checked. The result is
	// only known when every rule agrees on it: "+" yields a number for two
	// numbers and text for two strings, so with an unknown operand the result
	// is unknown too. Guessing the first rule's result here reported
	// `"Hello, " + name + "!"` as adding a number to text.
	if left == nil || right == nil ||
		left.Kind == types.TypeUnknown || right.Kind == types.TypeUnknown {
		return sharedResult(rules)
	}

	for _, r := range rules {
		if types.Canonical(r.left) == types.Canonical(left.Kind) &&
			types.Canonical(r.right) == types.Canonical(right.Kind) {
			if r.result == types.TypeArray {
				// Concatenation preserves the element type.
				return left
			}
			return types.InfoFor(r.result)
		}
	}

	a.errorWithHint(e.Pos(), operatorHint(e.Operator, rules),
		"'%s' cannot be used with %s and %s",
		e.Operator, describe(left), describe(right))
	return sharedResult(rules)
}

// sharedResult returns the result type common to every rule for an operator,
// or nil when the rules disagree and the operands did not settle it.
func sharedResult(rules []binaryRule) *types.TypeInfo {
	result := rules[0].result
	for _, r := range rules[1:] {
		if r.result != result {
			return nil
		}
	}
	return types.InfoFor(result)
}

// operatorHint lists what an operator does accept.
func operatorHint(op string, rules []binaryRule) string {
	if len(rules) == 1 {
		return fmt.Sprintf("'%s' works on two %ss.", op, types.Name(rules[0].left))
	}
	out := fmt.Sprintf("'%s' works on ", op)
	for i, r := range rules {
		switch {
		case i == 0:
		case i == len(rules)-1:
			out += ", or "
		default:
			out += ", "
		}
		out += fmt.Sprintf("two %ss", types.Name(r.left))
	}
	return out + "."
}

func (a *Analyzer) unaryType(e *ast.UnaryExpression) *types.TypeInfo {
	operand := a.checkExpr(e.Right)
	switch e.Operator {
	case "-":
		if operand != nil && operand.Kind != types.TypeUnknown &&
			!types.IsNumeric(operand.Kind) {
			a.errorAt(e.Pos(), "cannot negate %s", describe(operand))
		}
		return types.InfoFor(types.TypeF64)
	case "not":
		if operand != nil && operand.Kind != types.TypeUnknown &&
			operand.Kind != types.TypeBool && operand.Kind != types.TypeNull {
			a.errorWithHint(e.Pos(), "'not' works on a boolean.",
				"cannot apply 'not' to %s", describe(operand))
		}
		return types.InfoFor(types.TypeBool)
	}
	return nil
}

// ─── Indexing and collections ────────────────────────────────────────────────

func (a *Analyzer) indexType(e *ast.IndexExpression) *types.TypeInfo {
	container := a.checkExpr(e.List)
	a.requireNumber(e.Index, "an index")

	if container == nil || container.Kind == types.TypeUnknown {
		return nil
	}
	switch container.Kind {
	case types.TypeArray:
		return container.ElementType
	case types.TypeList, types.TypeString:
		// A list is heterogeneous, and indexing text yields text.
		if container.Kind == types.TypeString {
			return types.InfoFor(types.TypeString)
		}
		return nil
	default:
		a.errorWithHint(e.Pos(), "Indexing works on a list, an array, a range or text.",
			"cannot take an item from %s", describe(container))
		return nil
	}
}

// requireNumber reports an expression that is provably not a number.
func (a *Analyzer) requireNumber(expr ast.Expression, what string) {
	t := a.checkExpr(expr)
	if t == nil || t.Kind == types.TypeUnknown {
		return
	}
	if !types.IsNumeric(t.Kind) {
		a.errorAt(expr.Pos(), "%s must be a number, but this is %s", what, describe(t))
	}
}

// requireBool reports a condition that is provably not a boolean.
func (a *Analyzer) requireBool(expr ast.Expression, what string) {
	t := a.checkExpr(expr)
	if t == nil || t.Kind == types.TypeUnknown || t.Kind == types.TypeNull {
		return
	}
	if t.Kind != types.TypeBool {
		a.errorWithHint(expr.Pos(),
			"Use a comparison such as 'x is greater than 0', or a boolean value.",
			"%s must be a boolean, but this is %s", what, describe(t))
	}
}

// requireKind reports an expression that is provably not of the given kind.
func (a *Analyzer) requireKind(expr ast.Expression, kind types.TypeKind, what string) {
	t := a.checkExpr(expr)
	if t == nil || t.Kind == types.TypeUnknown {
		return
	}
	if types.Canonical(t.Kind) != types.Canonical(kind) {
		a.errorAt(expr.Pos(), "%s needs %s, but this is %s", what, types.Name(kind), describe(t))
	}
}

// requireCollection reports an expression that has no length.
func (a *Analyzer) requireCollection(expr ast.Expression, what string) {
	t := a.checkExpr(expr)
	if t == nil || t.Kind == types.TypeUnknown {
		return
	}
	switch t.Kind {
	case types.TypeList, types.TypeArray, types.TypeLookup, types.TypeString:
		return
	}
	a.errorAt(expr.Pos(), "%s needs a list, an array, a lookup table or text, but this is %s",
		what, describe(t))
}

// ─── Casts ───────────────────────────────────────────────────────────────────

// castableTargets are the types an explicit cast can produce.
var castableTargets = map[types.TypeKind]bool{
	types.TypeF64:    true,
	types.TypeI32:    true,
	types.TypeI64:    true,
	types.TypeU32:    true,
	types.TypeU64:    true,
	types.TypeF32:    true,
	types.TypeString: true,
	types.TypeBool:   true,
}

func (a *Analyzer) castType(e *ast.CastExpression) *types.TypeInfo {
	a.checkExpr(e.Value)
	target := a.resolve(e.Type)
	if target == nil {
		return nil
	}
	if !castableTargets[target.Kind] {
		a.errorWithHint(e.Type.Pos(),
			"A cast can produce a number, text or a boolean.",
			"cannot cast to %s", describe(target))
		return nil
	}
	return target
}

// ─── Calls ───────────────────────────────────────────────────────────────────

// callType checks a call and returns its result type. A user-defined function
// shadows a built-in of the same name, matching how the engines resolve calls.
func (a *Analyzer) callType(name string, args []ast.Expression, pos ast.Position) *types.TypeInfo {
	if fn, ok := a.funcs[name]; ok {
		return a.checkUserCall(fn, args, pos)
	}
	if sig, ok := stdlib.Lookup(name); ok {
		return a.checkBuiltinCall(sig, args, pos)
	}
	for _, arg := range args {
		a.checkExpr(arg)
	}
	a.errorAt(pos, "'%s' is not a declared function", name)
	return nil
}

func (a *Analyzer) checkUserCall(fn *funcSig, args []ast.Expression, pos ast.Position) *types.TypeInfo {
	argTypes := make([]*types.TypeInfo, len(args))
	for i, arg := range args {
		argTypes[i] = a.checkExpr(arg)
	}

	if len(args) != len(fn.Params) {
		a.errorWithHint(pos, callHint(fn),
			"'%s' takes %s, but %s given",
			fn.Name, argCount(len(fn.Params)), argCountGiven(len(args)))
		return a.resolve(fn.ReturnType)
	}

	for i, p := range fn.Params {
		want := a.resolve(p.Type)
		if want == nil {
			continue // unannotated parameter; nothing to check against
		}
		if !assignable(want, argTypes[i]) {
			a.errorAt(args[i].Pos(),
				"'%s' expects %s for '%s', but this is %s",
				fn.Name, describe(want), p.Name, describe(argTypes[i]))
		}
	}

	// A function that gives back nothing has no value to use, so a call to one
	// only makes sense as a statement of its own.
	if !givesAValue(fn.ReturnType) && !a.discardingResult {
		a.errorWithHint(pos,
			fmt.Sprintf("Write 'Call %s' as a statement of its own.", fn.Name),
			"'%s' gives back nothing, so there is no value here", fn.Name)
		return nil
	}
	return a.resolve(fn.ReturnType)
}

// callHint renders a function's signature for a diagnostic.
func callHint(fn *funcSig) string {
	if len(fn.Params) == 0 {
		return fmt.Sprintf("'%s' takes no arguments.", fn.Name)
	}
	out := fmt.Sprintf("'%s' takes ", fn.Name)
	for i, p := range fn.Params {
		if i > 0 {
			out += ", "
		}
		out += p.Name
		if p.Type != nil {
			out += " (" + p.Type.Name + ")"
		}
	}
	return out + "."
}

func (a *Analyzer) checkBuiltinCall(sig stdlib.Signature, args []ast.Expression, pos ast.Position) *types.TypeInfo {
	argTypes := make([]*types.TypeInfo, len(args))
	for i, arg := range args {
		argTypes[i] = a.checkExpr(arg)
	}

	if len(args) < sig.MinArgs() || len(args) > sig.MaxArgs() {
		a.errorWithHint(pos, "Usage: "+stdlib.Usage(sig),
			"'%s' takes %s, but %s given",
			sig.Name, builtinArgCount(sig), argCountGiven(len(args)))
		return returnTypeOf(sig)
	}

	for i, p := range sig.Params {
		if i >= len(argTypes) {
			break
		}
		t := argTypes[i]
		if t == nil {
			continue
		}
		if !p.Accepts(t.Kind) {
			a.errorWithHint(args[i].Pos(), "Usage: "+stdlib.Usage(sig),
				"'%s' expects %s for '%s', but this is %s",
				sig.Name, p.Describe(), p.Name, describe(t))
		}
	}
	return returnTypeOf(sig)
}

func returnTypeOf(sig stdlib.Signature) *types.TypeInfo {
	if sig.Returns == types.TypeUnknown {
		return nil
	}
	return types.InfoFor(sig.Returns)
}

func builtinArgCount(sig stdlib.Signature) string {
	if sig.MinArgs() == sig.MaxArgs() {
		return argCount(sig.MinArgs())
	}
	return fmt.Sprintf("%d to %s", sig.MinArgs(), argCount(sig.MaxArgs()))
}

func argCount(n int) string {
	switch n {
	case 0:
		return "no arguments"
	case 1:
		return "1 argument"
	default:
		return fmt.Sprintf("%d arguments", n)
	}
}

func argCountGiven(n int) string {
	switch n {
	case 0:
		return "none was"
	case 1:
		return "1 was"
	default:
		return fmt.Sprintf("%d were", n)
	}
}

// ─── Structs ─────────────────────────────────────────────────────────────────

func (a *Analyzer) structInstantiationType(e *ast.StructInstantiation) *types.TypeInfo {
	def, ok := a.structs[e.StructName]
	if !ok {
		for _, name := range e.FieldOrder {
			a.checkExpr(e.FieldValues[name])
		}
		a.errorAt(e.Pos(), "'%s' is not a declared struct", e.StructName)
		return nil
	}

	for _, name := range e.FieldOrder {
		value := e.FieldValues[name]
		got := a.checkExpr(value)
		field, known := def.Fields[name]
		if !known {
			pos := e.Pos()
			if value != nil {
				pos = value.Pos()
			}
			a.errorWithHint(pos, "Fields of "+def.Name+" are: "+joinNames(def.Order)+".",
				"struct '%s' has no field named '%s'", def.Name, name)
			continue
		}
		want := a.resolve(field.Type)
		if want != nil && !assignable(want, got) && value != nil {
			a.errorAt(value.Pos(), "field '%s' of %s is %s, but this is %s",
				name, def.Name, describe(want), describe(got))
		}
	}
	return &types.TypeInfo{Kind: types.TypeStruct, Name: def.Name}
}

func (a *Analyzer) fieldAccessType(e *ast.FieldAccess) *types.TypeInfo {
	obj := a.checkExpr(e.Object)
	if obj == nil || obj.Kind != types.TypeStruct {
		return nil
	}
	def, ok := a.structs[obj.Name]
	if !ok {
		return nil
	}
	field, known := def.Fields[e.Field]
	if !known {
		a.errorWithHint(e.Pos(), "Fields of "+def.Name+" are: "+joinNames(def.Order)+".",
			"struct '%s' has no field named '%s'", def.Name, e.Field)
		return nil
	}
	return a.resolve(field.Type)
}

func (a *Analyzer) methodCallType(e *ast.MethodCall) *types.TypeInfo {
	obj := a.checkExpr(e.Object)

	if obj != nil && obj.Kind == types.TypeStruct {
		if def, ok := a.structs[obj.Name]; ok {
			if m, ok := def.Methods[e.MethodName]; ok {
				return a.checkUserCall(m, e.Arguments, e.Pos())
			}
			a.errorAt(e.Pos(), "struct '%s' has no method named '%s'", def.Name, e.MethodName)
			for _, arg := range e.Arguments {
				a.checkExpr(arg)
			}
			return nil
		}
	}

	// Possessive syntax on a non-struct reads as a call with the object as the
	// first argument: text's casefold means casefold(text).
	args := append([]ast.Expression{e.Object}, e.Arguments...)
	return a.callType(e.MethodName, args, e.Pos())
}

func joinNames(names []string) string {
	out := ""
	for i, n := range names {
		if i > 0 {
			out += ", "
		}
		out += n
	}
	if out == "" {
		return "(none)"
	}
	return out
}
