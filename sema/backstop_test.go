package sema_test

import (
	"bytes"
	"strings"
	"testing"

	vm "github.com/Advik-B/english/astvm"
	"github.com/Advik-B/english/ivm"
	"github.com/Advik-B/english/parser"
	"github.com/Advik-B/english/stdlib"
)

// runUnchecked compiles and runs a program in both engines *without* analysis,
// which is what happens to a bytecode file produced elsewhere.
//
// The engines' own type checks are the backstop for exactly that case, so they
// have to hold on their own rather than relying on the checker having run.
func runUnchecked(t *testing.T, src string) (astErr, ivmErr error) {
	t.Helper()
	lexer := parser.NewLexer(src)
	p := parser.NewParser(lexer.TokenizeAll())
	prog, err := p.Parse()
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	env := vm.NewEnvironment()
	stdlib.Register(env)
	_, astErr = vm.NewEvaluator(env, stdlib.Eval).Eval(prog)

	chunk, compileErr := ivm.Compile(prog)
	if compileErr != nil {
		return astErr, compileErr
	}
	_, ivmErr = ivm.Execute(chunk, stdlib.Eval, stdlib.PredefinedValues())
	return astErr, ivmErr
}

// runUncheckedAST runs a program in the tree-walking engine without analysis
// and returns what it printed, for a case whose point is the output.
func runUncheckedAST(t *testing.T, src string) (string, error) {
	t.Helper()
	lexer := parser.NewLexer(src)
	p := parser.NewParser(lexer.TokenizeAll())
	prog, err := p.Parse()
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	var out bytes.Buffer
	env := vm.NewEnvironment()
	stdlib.Register(env)
	evaluator := vm.NewEvaluator(env, stdlib.Eval)
	evaluator.SetOutput(&out)
	_, err = evaluator.Eval(prog)
	return out.String(), err
}

// TestBackstopTypeLock covers the guarantee the language leads with. Only
// explicitly annotated names were locked in the instruction VM, so
// "Declare x to be 5." left x unlocked and a later assignment of text was
// accepted there while the other engine rejected it.
func TestBackstopTypeLock(t *testing.T) {
	astErr, ivmErr := runUnchecked(t, `Declare x to be 5.
Set x to be "hello".`)

	if astErr == nil {
		t.Error("the AST evaluator accepted a type-violating assignment")
	}
	if ivmErr == nil {
		t.Error("the instruction VM accepted a type-violating assignment")
	}
	for name, err := range map[string]error{"astvm": astErr, "ivm": ivmErr} {
		if err != nil && !strings.Contains(err.Error(), "cannot assign text") {
			t.Errorf("%s: unexpected message: %v", name, err)
		}
	}
}

// TestBackstopLogicalResultIsBoolean covers the short-circuit operators, which
// pushed the right operand unchanged in the instruction VM, so "true and 5"
// evaluated to 5 and "false or text" to the text.
func TestBackstopLogicalResultIsBoolean(t *testing.T) {
	for _, src := range []string{
		`Declare r to be true and 5.`,
		`Declare r to be false or "hi".`,
	} {
		astErr, ivmErr := runUnchecked(t, src)
		if astErr == nil {
			t.Errorf("the AST evaluator accepted a non-boolean operand: %s", src)
		}
		if ivmErr == nil {
			t.Errorf("the instruction VM accepted a non-boolean operand: %s", src)
		}
	}
}

// TestBackstopLogicalStillWorks guards the change above against rejecting
// correct code, including the short-circuit paths.
func TestBackstopLogicalStillWorks(t *testing.T) {
	astErr, ivmErr := runUnchecked(t, `Declare a to be true and false.
Declare b to be true or false.
Declare c to be false and true.
Declare d to be true or true.
Print a, b, c, d.`)
	if astErr != nil {
		t.Errorf("astvm rejected valid logic: %v", astErr)
	}
	if ivmErr != nil {
		t.Errorf("ivm rejected valid logic: %v", ivmErr)
	}
}

// TestBackstopStructFieldType covers a struct field given the wrong type,
// which the instruction VM did not check at run time at all.
func TestBackstopStructFieldType(t *testing.T) {
	astErr, ivmErr := runUnchecked(t, `Declare Point as a structure with the following fields:
    x is a number with 0 being the default.
thats it.

Declare label to be "left".
Declare p to be a new instance of Point with the following fields:
    x is label.
thats it.`)
	if astErr == nil {
		t.Error("the AST evaluator accepted a wrongly typed field")
	}
	if ivmErr == nil {
		t.Error("the instruction VM accepted a wrongly typed field")
	}
}

// TestBackstopStructFieldOrder covers the field-order corruption directly at
// the engine level, since the checker cannot see it: the program is valid, the
// values were simply bound to the wrong fields.
func TestBackstopStructFieldOrder(t *testing.T) {
	astErr, ivmErr := runUnchecked(t, `Declare Person as a structure with the following fields:
    name is a text with "?" being the default.
    age is a number with 0 being the default.
thats it.

Declare p to be a new instance of Person with the following fields:
    age is 30.
    name is "Alice".
thats it.`)
	if astErr != nil {
		t.Errorf("astvm rejected a valid instantiation: %v", astErr)
	}
	if ivmErr != nil {
		t.Errorf("ivm rejected a valid instantiation: %v", ivmErr)
	}
}

// TestBackstopMethodParameterShadowsAField covers a method whose parameter has
// the same name as a field of its own structure. The checker rejects that, so
// this is what the engines do with bytecode that was not checked: the
// parameter binding used to be silently dropped — the call that made it
// refused to redeclare the name and its error was discarded — so the method
// ran with the field's value in place of the argument it was called with.
func TestBackstopMethodParameterShadowsAField(t *testing.T) {
	src := `Declare Person as a structure with the following fields:
    name is a text with "?" being the default.

    let rename be a function that takes name as text, and gives back nothing, and does the following:
        Print name.
    thats it.
thats it.

Declare p to be a new instance of Person with the following fields:
    name is "Alice".
thats it.
call rename from p with "Bob".`

	out, err := runUncheckedAST(t, src)
	if err != nil {
		t.Fatalf("astvm: %v", err)
	}
	if !strings.Contains(out, "Bob") {
		t.Errorf("the method printed %q; it should see the argument it was called with",
			strings.TrimSpace(out))
	}
}
