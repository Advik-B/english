package stdlib_test

import (
	"os"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/Advik-B/english/stdlib"
)

// caseNames extracts the built-in names a Go source file switches on, which is
// how the two Python back-ends and the help registry each enumerate the
// standard library.
func caseNames(t *testing.T, path string, pattern string) map[string]bool {
	t.Helper()
	src, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("%s: %v", path, err)
	}
	re := regexp.MustCompile(pattern)
	found := make(map[string]bool)
	for _, m := range re.FindAllStringSubmatch(string(src), -1) {
		for _, name := range strings.Split(m[1], `", "`) {
			name = strings.Trim(name, `"`)
			if name != "" {
				found[name] = true
			}
		}
	}
	return found
}

func sorted(set map[string]bool) []string {
	out := make([]string, 0, len(set))
	for name := range set {
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}

// diff reports the names in a that are missing from b.
func diff(a, b map[string]bool) []string {
	var out []string
	for _, name := range sorted(a) {
		if !b[name] {
			out = append(out, name)
		}
	}
	return out
}

// TestSignatureTableIsAuthoritative pins the signature table as the one place
// the standard library is declared.
//
// The library used to be enumerated independently in six places, and they had
// drifted: the type checker's copy listed 54 of 84 built-ins, the two Python
// back-ends each mapped three functions that do not exist, and the decompiler
// was missing the whole time module. Nothing kept them in step.
func TestSignatureTableIsAuthoritative(t *testing.T) {
	declared := make(map[string]bool)
	for _, name := range stdlib.Names() {
		declared[name] = true
	}
	if len(declared) < 80 {
		t.Fatalf("the signature table declares only %d built-ins; the table is probably not being read", len(declared))
	}

	// Every registered function is a declared one, and vice versa. Register
	// and Eval both derive from the table, so this is a guard against the
	// table itself disagreeing with what it produces.
	for _, name := range stdlib.Names() {
		if _, ok := stdlib.Lookup(name); !ok {
			t.Errorf("%s is listed by Names but has no signature", name)
		}
	}

	casePattern := `(?m)^\s*case ((?:"[a-z_0-9]+"(?:, )?)+):`

	for _, target := range []struct {
		name string
		path string
	}{
		{"the AST transpiler", "../transpiler/stdlib.go"},
		{"the bytecode decompiler", "../ivm/decompiler_stdlib.go"},
	} {
		mapped := caseNames(t, target.path, casePattern)

		if extra := diff(mapped, declared); len(extra) > 0 {
			t.Errorf("%s translates built-ins that do not exist: %s\n"+
				"  Calling one of these in English reports an undefined function, so the translation is unreachable.",
				target.name, strings.Join(extra, ", "))
		}
		if missing := diff(declared, mapped); len(missing) > 0 {
			t.Errorf("%s has no translation for: %s\n"+
				"  The generated Python would call a name that does not exist there.",
				target.name, strings.Join(missing, ", "))
		}
	}
}

// TestHelpCoversEveryBuiltin keeps the help registry in step with the library.
// A function nobody can look up may as well not be documented.
func TestHelpCoversEveryBuiltin(t *testing.T) {
	src, err := os.ReadFile("../help/registry.go")
	if err != nil {
		t.Fatalf("cannot read the help registry: %v", err)
	}
	text := string(src)

	var undocumented []string
	for _, name := range stdlib.Names() {
		if !strings.Contains(text, `Name:        "`+name+`"`) {
			undocumented = append(undocumented, name)
		}
	}
	if len(undocumented) > 0 {
		t.Errorf("%d built-in(s) have no help entry: %s",
			len(undocumented), strings.Join(undocumented, ", "))
	}
}

// TestStringFunctionsCountCharacters covers the string functions, which were
// byte-oriented: substring could cut a multi-byte character in half, and the
// padding functions both measured width in bytes and sliced one byte off the
// pad character, so padding with a multi-byte character produced invalid text.
func TestStringFunctionsCountCharacters(t *testing.T) {
	cases := []struct {
		name string
		args []interface{}
		want string
	}{
		{"substring", []interface{}{"héllo", 1.0, 3.0}, "éll"},
		{"substring", []interface{}{"日本語", 1.0, 2.0}, "本語"},
		{"pad_left", []interface{}{"5", 4.0, "·"}, "···5"},
		{"pad_right", []interface{}{"5", 3.0, "·"}, "5··"},
		{"center", []interface{}{"x", 5.0, "·"}, "··x··"},
		{"zfill", []interface{}{"é", 3.0}, "00é"},
	}
	for _, c := range cases {
		got, err := stdlib.Eval(c.name, c.args)
		if err != nil {
			t.Errorf("%s%v: %v", c.name, c.args, err)
			continue
		}
		if got != c.want {
			t.Errorf("%s%v = %q, want %q", c.name, c.args, got, c.want)
		}
	}
}

// TestUniqueUsesLanguageEquality covers dedup, which keyed on the rendered
// form of a value, so the number 1 and the text "1" were the same key.
func TestUniqueUsesLanguageEquality(t *testing.T) {
	got, err := stdlib.Eval("unique", []interface{}{[]interface{}{1.0, "1", 1.0, "1"}})
	if err != nil {
		t.Fatalf("unique: %v", err)
	}
	list, ok := got.([]interface{})
	if !ok {
		t.Fatalf("unique returned %T, want a list", got)
	}
	if len(list) != 2 {
		t.Errorf("unique([1, \"1\", 1, \"1\"]) kept %d value(s), want 2", len(list))
	}
}

// TestSortIsConsistent covers the ordering predicate, which decided per pair —
// numeric when both happened to be numbers and textual otherwise — which is
// not transitive on a mixed list, so the result was undefined.
func TestSortIsConsistent(t *testing.T) {
	mixed := []interface{}{"b", 2.0, nil, true, "a", 1.0, false}

	first, err := stdlib.Eval("sort", []interface{}{mixed})
	if err != nil {
		t.Fatalf("sort: %v", err)
	}
	// Sorting an already sorted list must not change it, which a
	// non-transitive comparison cannot promise.
	second, err := stdlib.Eval("sort", []interface{}{first})
	if err != nil {
		t.Fatalf("sort: %v", err)
	}
	a := first.([]interface{})
	b := second.([]interface{})
	for i := range a {
		if !sameValue(a[i], b[i]) {
			t.Errorf("sorting twice changed the order at %d: %v then %v", i, a[i], b[i])
		}
	}
}

func sameValue(a, b interface{}) bool { return a == b }

// TestWrongTypeIsReported covers the functions that answered rather than
// reporting: is_nan said true for text, is_infinite said false, and is_empty
// said false for a number.
func TestWrongTypeIsReported(t *testing.T) {
	for _, name := range []string{"is_nan", "is_infinite"} {
		if _, err := stdlib.Eval(name, []interface{}{"hello"}); err == nil {
			t.Errorf("%s accepted text", name)
		}
	}
	if _, err := stdlib.Eval("is_empty", []interface{}{1.0}); err == nil {
		t.Error("is_empty accepted a number")
	}
	if _, err := stdlib.Eval("to_number", []interface{}{"12abc"}); err == nil {
		t.Error("to_number accepted trailing text")
	}
}
