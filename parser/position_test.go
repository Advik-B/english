package parser

import (
	"reflect"
	"strings"
	"testing"

	"github.com/Advik-B/english/ast"
	"github.com/Advik-B/english/types"
)

// walkNodes visits every ast.Node reachable from v, reporting each one's path.
// It uses reflection so that a node type added later is covered automatically
// rather than needing a new case here.
func walkNodes(v reflect.Value, path string, visit func(n ast.Node, path string)) {
	if !v.IsValid() {
		return
	}
	switch v.Kind() {
	case reflect.Ptr, reflect.Interface:
		if v.IsNil() {
			return
		}
		if n, ok := v.Interface().(ast.Node); ok {
			visit(n, path)
		}
		walkNodes(v.Elem(), path, visit)
	case reflect.Struct:
		t := v.Type()
		for i := 0; i < v.NumField(); i++ {
			if t.Field(i).PkgPath != "" {
				continue // unexported
			}
			walkNodes(v.Field(i), path+"."+t.Field(i).Name, visit)
		}
	case reflect.Slice, reflect.Array:
		for i := 0; i < v.Len(); i++ {
			walkNodes(v.Index(i), path, visit)
		}
	case reflect.Map:
		for _, k := range v.MapKeys() {
			walkNodes(v.MapIndex(k), path, visit)
		}
	}
}

// programCoveringMostSyntax exercises as many node types as one program can, so
// the position sweep below has broad coverage.
const programCoveringMostSyntax = `# a comment
Import "helper.abc".

Declare AppError as an error type.
Declare SubError as a type of AppError.

Declare Point as a structure with the following fields:
    x is a number with 0 being the default.
    y is a number with 0 being the default.

    let magnitude be a function that gives back a number, and does the following:
        Return x * x + y * y.
    thats it.
thats it.

Declare count as number to be 0.
Declare name to be "Alice".
Declare flag to be true.
Declare nums to be [1, 2, 3].
Declare arr to be an array of number [1, 2].
Declare scores to be a lookup table.
Declare span to be [1 .. 5].
Declare nothing_here to be nothing.

Declare function add that takes a as number and b as number, and gives back a number, and does the following:
    Return a + b.
thats it.

Set count to be 1.
Set the item at position 0 in nums to be 9.
Set scores at "k" to be 5.
Toggle flag.
Swap count and count.

Print name, count, -count, not flag.
Write "no newline".
Print the length of nums.
Print the item at position 0 in nums.
Print scores has "k".
Print nums is equal to nums.
Print count is something.
Print count cast to text.
Print the type of count.
Print the location of count.
Print a copy of nums.
Print a reference to count.
Print the value of count.
Print uppercase of name.
Print name's uppercase.

If count is greater than 0, then
    Print "positive".
otherwise
    Print "not positive".
thats it.

Repeat the following while count is less than 3:
    Set count to be count + 1.
    If count is equal to 2, then
        Continue.
    thats it.
    break out of this loop.
thats it.

Repeat the following 2 times:
    Print "twice".
thats it.

For each n in nums, do the following:
    Print n.
thats it.

Try doing the following:
    Raise "bad" as AppError.
on AppError:
    Print "handled", error.
but finally:
    Print "cleanup".
thats it.

Sleep for 1ms.
`

// TestEveryNodeHasAPosition is the guard for the change that gave every AST
// node a source position. Expression nodes previously had no position field at
// all, so a diagnostic could never point at anything finer than the enclosing
// statement's line — and never at a column.
func TestEveryNodeHasAPosition(t *testing.T) {
	prog, err := parse(programCoveringMostSyntax)
	if err != nil {
		t.Fatalf("coverage program failed to parse: %v", err)
	}

	seen := map[string]bool{}
	var missing []string
	walkNodes(reflect.ValueOf(prog), "Program", func(n ast.Node, path string) {
		typeName := reflect.TypeOf(n).String()
		seen[typeName] = true
		if _, isProgram := n.(*ast.Program); isProgram {
			return // the root has no meaningful single position
		}
		if !n.Pos().IsKnown() {
			missing = append(missing, typeName+" at "+path)
		}
	})

	if len(missing) > 0 {
		t.Errorf("%d node(s) have no position:\n  %s", len(missing), strings.Join(missing, "\n  "))
	}
	if len(seen) < 30 {
		t.Errorf("coverage program only reached %d node types; expected at least 30", len(seen))
	}
	t.Logf("swept %d distinct node types", len(seen))
}

// TestPositionsHaveColumns checks the position carries a column, not just a
// line — the AST had no column information anywhere before this change.
func TestPositionsHaveColumns(t *testing.T) {
	prog, err := parse("Declare x to be 1.\nDeclare y to be 2.")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}
	first := prog.Statements[0].Pos()
	second := prog.Statements[1].Pos()

	if first.Line != 1 || second.Line != 2 {
		t.Errorf("lines are %d and %d, want 1 and 2", first.Line, second.Line)
	}
	if first.Col < 1 || second.Col < 1 {
		t.Errorf("columns are %d and %d, want both >= 1", first.Col, second.Col)
	}
	if second.Offset <= first.Offset {
		t.Errorf("offsets are %d then %d, want increasing", first.Offset, second.Offset)
	}
}

// TestOperatorPositions checks a binary expression is positioned at its
// operator, so an error about "+" points at the "+".
func TestOperatorPositions(t *testing.T) {
	prog, err := parse(`Declare r to be 1 + 2.`)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}
	decl := prog.Statements[0].(*ast.VariableDecl)
	bin, ok := decl.Value.(*ast.BinaryExpression)
	if !ok {
		t.Fatalf("value is %T, want *ast.BinaryExpression", decl.Value)
	}
	// "Declare r to be 1 + 2." — the "+" is at column 19 (1-based).
	if bin.Pos().Col != 19 {
		t.Errorf("binary expression is at column %d, want 19 (the operator)", bin.Pos().Col)
	}
	if bin.Left.Pos().Col != 17 {
		t.Errorf("left operand is at column %d, want 17", bin.Left.Pos().Col)
	}
	if bin.Right.Pos().Col != 21 {
		t.Errorf("right operand is at column %d, want 21", bin.Right.Pos().Col)
	}
}

// TestTypeAnnotationsAreNodes covers the change from bare type-name strings to
// ast.TypeExpr. The four annotation positions used to hold a plain string, so
// the name was re-parsed with types.Parse every time it was needed — including
// on every cast and every struct instantiation at run time — and carried no
// position, so nothing could point at a bad annotation.
func TestTypeAnnotationsAreNodes(t *testing.T) {
	prog, err := parse(`Declare count as number to be 0.
Declare Point as a structure with the following fields:
    x is a number with 0 being the default.
thats it.
Print 1 cast to text.
Declare arr to be an array of number [1, 2].`)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	// Typed declaration: kind resolved, position recorded.
	decl := prog.Statements[0].(*ast.TypedVariableDecl)
	if decl.Type == nil {
		t.Fatal("typed declaration has no annotation")
	}
	if decl.Type.Name != "number" || decl.Type.Kind != types.TypeF64 {
		t.Errorf("annotation is %q/%v, want \"number\"/TypeF64", decl.Type.Name, decl.Type.Kind)
	}
	if !decl.Type.Pos().IsKnown() {
		t.Error("annotation has no position")
	}
	// "Declare count as number to be 0." — "number" starts at column 18.
	if got := decl.Type.Pos().Col; got != 18 {
		t.Errorf("annotation is at column %d, want 18", got)
	}

	// Struct field annotation.
	sd := prog.Statements[1].(*ast.StructDecl)
	if len(sd.Fields) != 1 {
		t.Fatalf("struct has %d field(s), want 1", len(sd.Fields))
	}
	if f := sd.Fields[0]; f.Type == nil || f.Type.Kind != types.TypeF64 {
		t.Errorf("field annotation is %v, want a number", f.Type)
	} else if !f.Type.Pos().IsKnown() {
		t.Error("field annotation has no position")
	}

	// Cast target.
	out := prog.Statements[2].(*ast.OutputStatement)
	cast, ok := out.Values[0].(*ast.CastExpression)
	if !ok {
		t.Fatalf("printed value is %T, want *ast.CastExpression", out.Values[0])
	}
	if cast.Type == nil || cast.Type.Kind != types.TypeString {
		t.Errorf("cast target is %v, want text", cast.Type)
	}
	if !cast.Type.Pos().IsKnown() {
		t.Error("cast target has no position")
	}

	// Array element type.
	arr := prog.Statements[3].(*ast.VariableDecl)
	lit, ok := arr.Value.(*ast.ArrayLiteral)
	if !ok {
		t.Fatalf("array value is %T, want *ast.ArrayLiteral", arr.Value)
	}
	if lit.ElemType == nil || lit.ElemType.Kind != types.TypeF64 {
		t.Errorf("element type is %v, want a number", lit.ElemType)
	}
}

// TestStructNameAnnotationKeepsCase checks a non-built-in annotation is left
// alone so the type checker can match it against a declared struct, while a
// built-in name is normalised.
func TestStructNameAnnotationKeepsCase(t *testing.T) {
	prog, err := parse(`Declare p as Point.
Declare n as NUMBER to be 1.`)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}
	structAnnot := prog.Statements[0].(*ast.TypedVariableDecl).Type
	if structAnnot.Name != "Point" {
		t.Errorf("struct annotation is %q, want %q (case preserved)", structAnnot.Name, "Point")
	}
	if structAnnot.Kind != types.TypeUnknown {
		t.Errorf("struct annotation kind is %v, want TypeUnknown (only sema can resolve it)", structAnnot.Kind)
	}
	builtin := prog.Statements[1].(*ast.TypedVariableDecl).Type
	if builtin.Name != "number" {
		t.Errorf("built-in annotation is %q, want %q (normalised)", builtin.Name, "number")
	}
}

// TestFunctionParametersArePositioned covers the change from a bare []string
// of parameter names to []ast.Param. A function had no signature at all
// before: nothing could check an argument's type at a call site, or a return
// against what the function claims to give back.
func TestFunctionParametersArePositioned(t *testing.T) {
	prog, err := parse(`Declare function add that takes a as number and b as number, and gives back a number, and does the following:
    Return a + b.
thats it.`)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}
	fd := prog.Statements[0].(*ast.FunctionDecl)

	if len(fd.Params) != 2 {
		t.Fatalf("got %d parameter(s), want 2", len(fd.Params))
	}
	if fd.Params[0].Name != "a" || fd.Params[1].Name != "b" {
		t.Errorf("parameter names are %q and %q, want \"a\" and \"b\"",
			fd.Params[0].Name, fd.Params[1].Name)
	}
	for i, p := range fd.Params {
		if !p.Pos().IsKnown() {
			t.Errorf("parameter %d (%s) has no position", i, p.Name)
		}
		if ast.TypeName(p.Type) != "number" {
			t.Errorf("parameter %d (%s) is annotated %q, want number",
				i, p.Name, ast.TypeName(p.Type))
		}
		if p.Type != nil && !p.Type.Pos().IsKnown() {
			t.Errorf("parameter %d (%s) has an annotation with no position", i, p.Name)
		}
	}
	if fd.Params[1].Pos().Col <= fd.Params[0].Pos().Col {
		t.Error("parameter positions are not in source order")
	}
	if ast.TypeName(fd.ReturnType) != "number" {
		t.Errorf("return type is %q, want number", ast.TypeName(fd.ReturnType))
	}

	// ParamNames keeps the runtime's binding path simple.
	names := fd.ParamNames()
	if len(names) != 2 || names[0] != "a" || names[1] != "b" {
		t.Errorf("ParamNames() = %v, want [a b]", names)
	}
}
