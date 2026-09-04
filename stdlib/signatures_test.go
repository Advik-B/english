package stdlib

import (
	"testing"

	vm "github.com/Advik-B/english/astvm"
)

// callSafely invokes Eval and converts a panic into an error so that a crashing
// built-in is reported as a test failure rather than taking the suite down.
func callSafely(name string, args []vm.Value) (err error, panicked interface{}) {
	defer func() {
		if r := recover(); r != nil {
			panicked = r
		}
	}()
	_, err = Eval(name, args)
	return err, nil
}

// TestBuiltinsRejectWrongArity is the regression test for the crash class where
// a built-in indexed args positionally with no arity guard, turning a simple
// user mistake into an uncaught Go panic that killed the process and the REPL.
// Every built-in must reject every out-of-range argument count with an error.
func TestBuiltinsRejectWrongArity(t *testing.T) {
	for _, s := range signatureList {
		for n := 0; n <= s.MaxArgs()+2; n++ {
			if n >= s.MinArgs() && n <= s.MaxArgs() {
				continue // valid arity; covered by TestBuiltinsSurviveNilArguments
			}
			args := make([]vm.Value, n)
			err, panicked := callSafely(s.Name, args)
			if panicked != nil {
				t.Errorf("%s with %d arg(s): panicked: %v", s.Name, n, panicked)
				continue
			}
			if err == nil {
				t.Errorf("%s with %d arg(s): expected an arity error, got nil", s.Name, n)
			}
		}
	}
}

// TestBuiltinsSurviveNilArguments calls every built-in at each of its legal
// argument counts with nothing (nil) for every argument. A type error is the
// expected outcome for most; a panic never is.
func TestBuiltinsSurviveNilArguments(t *testing.T) {
	// ask reads stdin and sleep blocks, so neither is useful to invoke here.
	skip := map[string]bool{"ask": true, "sleep": true}
	for _, s := range signatureList {
		if skip[s.Name] {
			continue
		}
		for n := s.MinArgs(); n <= s.MaxArgs(); n++ {
			args := make([]vm.Value, n)
			_, panicked := callSafely(s.Name, args)
			if panicked != nil {
				t.Errorf("%s with %d nil arg(s): panicked: %v", s.Name, n, panicked)
			}
		}
	}
}

// TestEverySignatureDispatches guards against a signature being declared with a
// Module that no dispatcher handles, which would make the built-in unreachable.
func TestEverySignatureDispatches(t *testing.T) {
	for _, s := range signatureList {
		args := make([]vm.Value, s.MinArgs())
		if s.Name == "ask" || s.Name == "sleep" {
			continue
		}
		_, err := Eval(s.Name, args)
		if err != nil && err.Error() == "Runtime Error: unknown built-in function: "+s.Name+"\n" {
			t.Errorf("%s: declared in signatureList but not dispatched by Eval", s.Name)
		}
	}
}

// TestRegisterMatchesSignatures verifies that what Register installs into an
// environment is exactly what signatureList declares — the two used to be
// maintained by hand in separate places and could drift.
func TestRegisterMatchesSignatures(t *testing.T) {
	env := vm.NewEnvironment()
	Register(env)
	fns := env.GetAllFunctions()

	for _, s := range signatureList {
		fn, ok := fns[s.Name]
		if !ok {
			t.Errorf("%s: declared in signatureList but not registered", s.Name)
			continue
		}
		if len(fn.Parameters) != len(s.Params) {
			t.Errorf("%s: registered with %d parameter(s), signature declares %d",
				s.Name, len(fn.Parameters), len(s.Params))
		}
		if fn.Body != nil {
			t.Errorf("%s: registered with a non-nil body; built-ins must have a nil body", s.Name)
		}
	}
	if len(fns) != len(signatureList) {
		t.Errorf("Register installed %d functions, signatureList declares %d",
			len(fns), len(signatureList))
	}
}

// TestPredefinedNamesMatchValues keeps the two views of the constant table in
// step; they were previously two hand-written lists.
func TestPredefinedNamesMatchValues(t *testing.T) {
	names := PredefinedNames()
	values := PredefinedValues()
	if len(names) != len(values) {
		t.Fatalf("PredefinedNames has %d entries, PredefinedValues has %d", len(names), len(values))
	}
	for _, n := range names {
		if _, ok := values[n]; !ok {
			t.Errorf("%q in PredefinedNames but not PredefinedValues", n)
		}
	}
}

// TestUsageRendering documents the shape of the usage hint shown on an arity error.
func TestUsageRendering(t *testing.T) {
	cases := map[string]string{
		"sqrt":     "sqrt(x)",
		"replace":  "replace(text, old, new)",
		"pad_left": "pad_left(text, width[, char])",
		"random":   "random()",
		"ask":      "ask([prompt])",
	}
	for name, want := range cases {
		s, ok := Lookup(name)
		if !ok {
			t.Fatalf("%s: not found in signature table", name)
		}
		if got := Usage(s); got != want {
			t.Errorf("usage(%s) = %q, want %q", name, got, want)
		}
	}
}
