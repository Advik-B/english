package sema_test

import (
	"strings"
	"testing"
)

// ruleCase is one semantic rule: a program that must be rejected at a precise
// position, and a nearby program that must be accepted.
type ruleCase struct {
	name string
	// bad is a program that must produce a diagnostic.
	bad string
	// line and col are where the first diagnostic must point. A zero column
	// means the column is not asserted.
	line, col int
	// contains is a fragment the message must include.
	contains string
	// good is a program that must produce no diagnostics, guarding the rule
	// against over-reach.
	good string
}

// TestRules is the rule-by-rule table for the analyser. Each entry pins both
// halves of a check: that it fires where it should, and that it stays quiet on
// correct code.
func TestRules(t *testing.T) {
	cases := []ruleCase{
		{
			name: "assignment must match the declared type",
			bad:  "Declare total to be 0.\nSet total to be \"text\".",
			line: 2, col: 17,
			contains: "cannot assign text to 'total'",
			good:     "Declare total to be 0.\nSet total to be 5.",
		},
		{
			name: "assigning to an undeclared name",
			bad:  "Declare total to be 0.\nSet totl to be 5.",
			line: 2, col: 1,
			contains: "'totl' is not declared",
			good:     "Declare total to be 0.\nSet total to be 5.",
		},
		{
			name: "reading an undeclared name",
			bad:  "Print missing.",
			line: 1, col: 7,
			contains: "'missing' is not declared",
			good:     "Declare present to be 1.\nPrint present.",
		},
		{
			name: "reassigning a constant",
			bad:  "Declare limit to always be 10.\nSet limit to be 11.",
			line: 2, col: 1,
			contains: "it is a constant",
			good:     "Declare limit to be 10.\nSet limit to be 11.",
		},
		{
			name: "declaring the same name twice",
			bad:  "Declare x to be 1.\nDeclare x to be 2.",
			line: 2, col: 9,
			contains: "already declared",
			good:     "Declare x to be 1.\nDeclare y to be 2.",
		},
		{
			name: "shadowing a built-in constant",
			bad:  "Declare pi to be 3.",
			line: 1, col: 9,
			contains: "shadows a built-in constant",
			good:     "Declare my_pi to be 3.",
		},
		{
			name: "initialising a typed declaration with the wrong type",
			bad:  "Declare count as number to be \"ten\".",
			line: 1, col: 31,
			contains: "cannot initialize number with text",
			good:     "Declare count as number to be 10.",
		},
		{
			name: "an annotation that names nothing",
			bad:  "Declare x as banana to be 1.",
			line: 1, col: 14,
			contains: "unknown type 'banana'",
			good:     "Declare x as number to be 1.",
		},
		{
			name: "reading a declared but unassigned variable",
			bad:  "Declare score as number.\nPrint score.",
			line: 2, col: 7,
			contains: "has no value yet",
			good:     "Declare score as number.\nSet score to be 1.\nPrint score.",
		},
		{
			name: "adding mismatched types",
			bad:  "Print 1 + \"two\".",
			line: 1, col: 9,
			contains: "'+' cannot be used with number and text",
			good:     "Print 1 + 2.\nPrint \"one\" + \"two\".",
		},
		{
			name: "subtracting text",
			bad:  "Print \"a\" - 1.",
			line: 1, col: 11,
			contains: "'-' cannot be used with text and number",
			good:     "Print 2 - 1.",
		},
		{
			name: "a non-boolean condition",
			bad:  "Declare n to be 1.\nIf n, then\n    Print \"x\".\nthats it.",
			line: 2, col: 4,
			contains: "a condition must be a boolean",
			good:     "Declare n to be 1.\nIf n is greater than 0, then\n    Print \"x\".\nthats it.",
		},
		{
			name: "comparing values that can never be equal",
			bad:  "Print 1 is equal to \"one\".",
			line: 1, col: 7,
			contains: "they are never equal",
			good:     "Print 1 is equal to 2.",
		},
		{
			name: "negating text",
			bad:  "Print -\"x\".",
			line: 1, col: 7,
			contains: "cannot negate text",
			good:     "Print -1.",
		},
		{
			name: "calling a function with too few arguments",
			bad: `Declare function add that takes a and b and does the following:
    Return a + b.
thats it.
Print the result of calling add with 1.`,
			line: 4, col: 11,
			contains: "'add' takes 2 arguments, but 1 was given",
			good: `Declare function add that takes a and b and does the following:
    Return a + b.
thats it.
Print the result of calling add with 1 and 2.`,
		},
		{
			name: "passing the wrong type to an annotated parameter",
			bad: `Declare function twice that takes n as number and gives back a number, and does the following:
    Return n * 2.
thats it.
Print the result of calling twice with "x".`,
			line: 4, col: 40,
			contains: "expects number for 'n', but this is text",
			good: `Declare function twice that takes n as number and gives back a number, and does the following:
    Return n * 2.
thats it.
Print the result of calling twice with 4.`,
		},
		{
			name: "returning the wrong type",
			bad: `Declare function name_of that gives back a text, and does the following:
    Return 42.
thats it.`,
			line: 2, col: 5,
			contains: "gives back text, but this returns number",
			good: `Declare function name_of that gives back a text, and does the following:
    Return "x".
thats it.`,
		},
		{
			name: "a function that can finish without returning",
			bad: `Declare function pick that takes n as number and gives back a number, and does the following:
    If n is greater than 0, then
        Return 1.
    thats it.
thats it.`,
			line: 1, col: 9,
			contains: "can finish without returning a value",
			good: `Declare function pick that takes n as number and gives back a number, and does the following:
    If n is greater than 0, then
        Return 1.
    otherwise
        Return 0.
    thats it.
thats it.`,
		},
		{
			name: "calling something that is not a function",
			bad:  "Print the result of calling nope with 1.",
			line: 1, col: 11,
			contains: "'nope' is not a declared function",
			good:     "Print sqrt of 4.",
		},
		{
			name: "a built-in given the wrong type",
			bad:  "Print sqrt of \"x\".",
			line: 1, col: 15,
			contains: "'sqrt' expects number for 'x', but this is text",
			good:     "Print sqrt of 4.",
		},
		{
			name: "a built-in given too many arguments",
			bad:  "Print zfill of \"x\".",
			line: 1, col: 7,
			contains: "'zfill' takes 2 arguments, but 1 was given",
			good:     "Print uppercase of \"x\".",
		},
		{
			name: "break outside a loop",
			bad:  "break out of this loop.",
			line: 1, col: 1,
			contains: "only allowed inside a loop",
			good:     "Declare i to be 0.\nRepeat the following while i is less than 3:\n    break out of this loop.\nthats it.",
		},
		{
			name: "return outside a function",
			bad:  "Return 1.",
			line: 1, col: 1,
			contains: "only allowed inside a function",
			good:     "Declare function f that does the following:\n    Return 1.\nthats it.",
		},
		{
			name: "toggling something that is not a boolean",
			bad:  "Declare n to be 1.\nToggle n.",
			line: 2, col: 1,
			contains: "not a boolean",
			good:     "Declare flag to be true.\nToggle flag.",
		},
		{
			name: "a struct field given the wrong type",
			bad: `Declare Point as a structure with the following fields:
    x is a number with 0 being the default.
thats it.
Declare p to be a new instance of Point with the following fields:
    x is "left".
thats it.`,
			line: 5, col: 10,
			contains: "field 'x' of Point is number, but this is text",
			good: `Declare Point as a structure with the following fields:
    x is a number with 0 being the default.
thats it.
Declare p to be a new instance of Point with the following fields:
    x is 3.
thats it.`,
		},
		{
			name: "a struct field that does not exist",
			bad: `Declare Point as a structure with the following fields:
    x is a number with 0 being the default.
thats it.
Declare p to be a new instance of Point with the following fields:
    z is 1.
thats it.`,
			line: 5, col: 10,
			contains: "has no field named 'z'",
			good: `Declare Point as a structure with the following fields:
    x is a number with 0 being the default.
thats it.
Declare p to be a new instance of Point with the following fields:
    x is 1.
thats it.`,
		},
		{
			name: "catching an error type that was never declared",
			bad:  "Try doing the following:\n    Print 1.\non MysteryError:\n    Print 2.\nthats it.",
			line: 1, col: 1,
			contains: "'MysteryError' is not a declared error type",
			good:     "Declare MysteryError as an error type.\nTry doing the following:\n    Print 1.\non MysteryError:\n    Print 2.\nthats it.",
		},
		{
			name: "an array literal with mixed element types",
			bad:  "Declare a to be an array of number [1, \"two\"].",
			line: 1, col: 40,
			contains: "cannot hold text",
			good:     "Declare a to be an array of number [1, 2].",
		},
		{
			name: "taking the length of a number",
			bad:  "Declare n to be 1.\nPrint the length of n.",
			line: 2, col: 21,
			contains: "needs a list, an array, a lookup table or text",
			good:     "Declare l to be [1, 2].\nPrint the length of l.",
		},
		{
			name: "iterating something that is not a collection",
			bad:  "Declare n to be 1.\nFor each x in n, do the following:\n    Print x.\nthats it.",
			line: 2, col: 15,
			contains: "'for each' needs a list",
			good:     "Declare l to be [1, 2].\nFor each x in l, do the following:\n    Print x.\nthats it.",
		},
		{
			name: "casting to a type that cannot be produced",
			bad:  "Print 1 cast to list.",
			line: 1, col: 17,
			contains: "cannot cast to list",
			good:     "Print 1 cast to text.",
		},
		{
			// "x is NetworkError" parses for any x at all, and for anything
			// but an error the answer is always false — not a comparison
			// anyone writes on purpose.
			name: "an error type check on something that is not an error",
			bad: "Declare NetworkError as an error type.\nDeclare x to be 5.\n" +
				"If x is NetworkError, then\n    Print \"yes\".\nthats it.",
			line: 3, col: 4,
			contains: "an error type check needs error",
			good: "Declare NetworkError as an error type.\nTry doing the following:\n" +
				"    Raise \"down\" as NetworkError.\non error:\n" +
				"    If error is NetworkError, then\n        Print \"yes\".\n    thats it.\nthats it.",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			diags := check(c.bad)
			if len(diags) == 0 {
				t.Fatalf("expected a diagnostic for:\n%s", c.bad)
			}
			first := diags[0]
			if !strings.Contains(first.Message, c.contains) {
				t.Errorf("message is %q, want it to contain %q", first.Message, c.contains)
			}
			if first.Pos.Line != c.line {
				t.Errorf("reported on line %d, want %d (message: %s)", first.Pos.Line, c.line, first.Message)
			}
			if c.col != 0 && first.Pos.Col != c.col {
				t.Errorf("reported at column %d, want %d (message: %s)", first.Pos.Col, c.col, first.Message)
			}

			if c.good == "" {
				return
			}
			if ok := check(c.good); len(ok) != 0 {
				for _, d := range ok {
					t.Errorf("correct program reported: line %d, col %d: %s", d.Pos.Line, d.Pos.Col, d.Message)
				}
			}
		})
	}
}

// TestReportsEveryProblem covers the analyser reporting all the problems in a
// program rather than stopping at the first, which the previous checker did.
func TestReportsEveryProblem(t *testing.T) {
	diags := check(`Declare total to be 0.
Set total to be "text".
Print 1 + "two".
Print missing.`)

	if len(diags) < 3 {
		t.Fatalf("expected at least 3 diagnostics, got %d", len(diags))
	}
	// They must arrive in source order so the first thing reported is the
	// first thing wrong.
	for i := 1; i < len(diags); i++ {
		prev, cur := diags[i-1], diags[i]
		if cur.Pos.Line < prev.Pos.Line {
			t.Errorf("diagnostics are out of order: line %d before line %d", prev.Pos.Line, cur.Pos.Line)
		}
	}
}

// TestScopedTypesDoNotLeak covers the flat type map the previous checker used
// beside its scope stack, which let a type recorded inside a block stay
// visible after the block ended.
func TestScopedTypesDoNotLeak(t *testing.T) {
	// inner is a number inside the loop; the outer inner is text. Assigning
	// text to the outer one must be accepted.
	diags := check(`Declare inner to be "text".
Declare i to be 0.
Repeat the following while i is less than 1:
    Declare shadowed to be 1.
    Set i to be 1.
thats it.
Set inner to be "still text".`)
	for _, d := range diags {
		t.Errorf("unexpected diagnostic: line %d: %s", d.Pos.Line, d.Message)
	}
}

// TestBlockScopedNamesAreNotVisibleOutside is the other half: a name declared
// in a block must not be usable after it.
func TestBlockScopedNamesAreNotVisibleOutside(t *testing.T) {
	diags := check(`Declare i to be 0.
Repeat the following while i is less than 1:
    Declare inside to be 1.
    Set i to be 1.
thats it.
Print inside.`)
	if len(diags) == 0 {
		t.Fatal("expected a diagnostic for using a block-scoped name outside its block")
	}
	if !strings.Contains(diags[0].Message, "'inside' is not declared") {
		t.Errorf("unexpected message: %s", diags[0].Message)
	}
}
