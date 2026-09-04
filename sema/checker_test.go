package sema_test

import (
	"os"
	"strings"
	"testing"

	"github.com/Advik-B/english/ast"
	"github.com/Advik-B/english/parser"
	"github.com/Advik-B/english/sema"
	"github.com/Advik-B/english/stdlib"
	"github.com/Advik-B/english/types"
)

// check parses the given source and analyses it with the stdlib constants
// predeclared, mirroring what every entry point does before execution.
func check(input string) []*sema.Diagnostic {
	lexer := parser.NewLexer(input)
	p := parser.NewParser(lexer.TokenizeAll())
	program, err := p.Parse()
	if err != nil {
		return nil
	}
	return sema.Check(program, sema.Config{Predefined: stdlib.PredefinedNames()})
}

func TestChecker_DuplicateVarTopLevel(t *testing.T) {
	errs := check(`Declare x to be 1.
Declare x to be 2.`)
	if len(errs) == 0 {
		t.Fatal("expected a duplicate-variable error, got none")
	}
	msg := errs[0].Error()
	if !strings.Contains(msg, "x") {
		t.Errorf("error should mention variable name 'x', got: %s", msg)
	}
	if errs[0].Pos.Line != 2 {
		t.Errorf("error should be on line 2, got line %d", errs[0].Pos.Line)
	}
}

func TestChecker_DuplicateShadowsStdlibConstant(t *testing.T) {
	errs := check(`Declare pi to be 3.`)
	if len(errs) == 0 {
		t.Fatal("expected error for redeclaring stdlib constant 'pi', got none")
	}
	msg := errs[0].Error()
	if !strings.Contains(msg, "pi") {
		t.Errorf("error should mention 'pi', got: %s", msg)
	}
}

func TestChecker_DuplicateLetSyntax(t *testing.T) {
	errs := check(`let x be 1.
let x be 2.`)
	if len(errs) == 0 {
		t.Fatal("expected a duplicate-variable error, got none")
	}
	if errs[0].Pos.Line != 2 {
		t.Errorf("error should be on line 2, got line %d", errs[0].Pos.Line)
	}
}

func TestChecker_NoDuplicateInDifferentScopes(t *testing.T) {
	// Same name in an inner scope (if body) must NOT be flagged.
	errs := check(`Declare x to be 1.
If yes, then
    Declare x to be 2.
thats it.`)
	for _, e := range errs {
		if strings.Contains(e.Error(), "x") {
			t.Errorf("unexpected error for shadowing in inner scope: %s", e.Error())
		}
	}
}

func TestChecker_NoDuplicateForUniqueNames(t *testing.T) {
	errs := check(`Declare a to be 1.
Declare b to be 2.
Declare c to be 3.`)
	if len(errs) != 0 {
		t.Errorf("expected no errors for distinct names, got: %v", errs)
	}
}

// TestChecker_FollowsImports verifies that the checker reads and validates
// imported .abc files, catching stdlib-constant shadowing at compile time so
// that `english run` never partially executes a program with a compile error.
func TestChecker_FollowsImports(t *testing.T) {
	// Write a temporary library that redefines the stdlib constant "pi".
	libFile, err := os.CreateTemp(t.TempDir(), "lib_*.abc")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := libFile.WriteString("Declare pi to always be 3.14159.\n"); err != nil {
		t.Fatal(err)
	}
	libFile.Close()

	// The main program just imports that library.
	errs := check(`Import "` + libFile.Name() + `".`)
	if len(errs) == 0 {
		t.Fatal("expected a compile error for importing a file that shadows 'pi', got none")
	}
	e := errs[0]
	msg := e.Error()
	if !strings.Contains(msg, "pi") {
		t.Errorf("error should mention 'pi', got: %s", msg)
	}
	if !strings.Contains(msg, "shadows") {
		t.Errorf("error should say 'shadows', got: %s", msg)
	}
	// The error must identify the file it came from.
	if e.File == "" {
		t.Errorf("error should carry the imported file path, but File is empty")
	}
	if e.File != libFile.Name() {
		t.Errorf("expected error File to be %q, got %q", libFile.Name(), e.File)
	}
}

// TestCheckerRecordsInferredTypes covers the expression type slot added to the
// AST. Expression nodes had nowhere to record a type, so every stage that
// needed one re-derived it; the checker now writes its result onto the node.
func TestCheckerRecordsInferredTypes(t *testing.T) {
	src := `Declare n to be 1.
Declare s to be "x".
Declare b to be true.
Declare c to be "7" cast to number.`

	lexer := parser.NewLexer(src)
	p := parser.NewParser(lexer.TokenizeAll())
	prog, err := p.Parse()
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}
	if errs := sema.Check(prog, sema.Config{}); len(errs) != 0 {
		t.Fatalf("unexpected check errors: %v", errs)
	}

	want := []types.TypeKind{
		types.TypeF64,    // 1
		types.TypeString, // "x"
		types.TypeBool,   // true
		types.TypeF64,    // "7" cast to number
	}
	for i, w := range want {
		decl, ok := prog.Statements[i].(*ast.VariableDecl)
		if !ok {
			t.Fatalf("statement %d is %T, want *ast.VariableDecl", i, prog.Statements[i])
		}
		got := decl.Value.InferredType()
		if got == nil {
			t.Errorf("statement %d: no inferred type recorded", i)
			continue
		}
		if got.Kind != w {
			t.Errorf("statement %d: inferred %v, want %v", i, got.Kind, w)
		}
	}
}

// TestCastResultIsTypeChecked follows from recording the cast target's kind:
// a cast's result type is now known statically, so a mismatched initialiser is
// caught at compile time instead of at run time.
func TestCastResultIsTypeChecked(t *testing.T) {
	lexer := parser.NewLexer(`Declare x as text to be 5 cast to number.`)
	p := parser.NewParser(lexer.TokenizeAll())
	prog, err := p.Parse()
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}
	errs := sema.Check(prog, sema.Config{})
	if len(errs) == 0 {
		t.Fatal("expected a compile error for initialising text with a number cast")
	}
	if !strings.Contains(errs[0].Message, "cannot initialize text with number") {
		t.Errorf("unexpected error: %v", errs[0].Message)
	}
}
