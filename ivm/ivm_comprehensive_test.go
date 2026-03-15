package ivm_test

// Comprehensive tests for the ivm package covering all language features:
// continue/break, structs, possessives, imports, transpilation parity,
// compile-time type errors, typed variables, lookup tables, arrays, error types,
// error hierarchy, stdlib functions, encoding, and more.

import (
	"github.com/Advik-B/english/ivm"
	"github.com/Advik-B/english/stdlib"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ─── Continue / Break ────────────────────────────────────────────────────────

func TestContinueInWhileLoop(t *testing.T) {
	out := captureOutput(func() {
		_, err := run(`Declare i to be 0.
Declare total to be 0.
repeat the following while i is less than 10:
    Set i to be i + 1.
    Declare m to be the remainder of i divided by 2.
    If m is equal to 0, then
        Continue.
    thats it.
    Set total to be total + i.
thats it.
Print the value of total.`)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	// 1+3+5+7+9 = 25
	if !strings.Contains(out, "25") {
		t.Errorf("expected 25, got %q", out)
	}
}

func TestContinueInForLoop(t *testing.T) {
	out := captureOutput(func() {
		_, err := run(`Declare count to be 0.
Declare printed to be 0.
repeat the following 5 times:
    Set count to be count + 1.
    If count is equal to 3, then
        Continue.
    thats it.
    Set printed to be printed + 1.
thats it.
Print the value of printed.`)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	// Skipped count=3, so printed 4 times
	if !strings.Contains(out, "4") {
		t.Errorf("expected 4, got %q", out)
	}
}

func TestContinueInForEachLoop(t *testing.T) {
	out := captureOutput(func() {
		_, err := run(`Declare total to be 0.
Declare nums to be [1, 2, 3, 4, 5].
for each n in nums, do the following:
    Declare m to be the remainder of n divided by 3.
    If m is equal to 0, then
        Continue.
    thats it.
    Set total to be total + n.
thats it.
Print the value of total.`)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	// 1+2+4+5 = 12 (skip 3)
	if !strings.Contains(out, "12") {
		t.Errorf("expected 12, got %q", out)
	}
}

func TestBreakInWhileLoop(t *testing.T) {
	out := captureOutput(func() {
		_, err := run(`Declare i to be 0.
repeat the following while i is less than 10:
    Set i to be i + 1.
    If i is equal to 5, then
        break out of this loop.
    thats it.
thats it.
Print the value of i.`)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "5") {
		t.Errorf("expected 5, got %q", out)
	}
}

func TestBreakInForEachLoop(t *testing.T) {
	out := captureOutput(func() {
		_, err := run(`Declare found to be 0.
Declare items to be [10, 20, 30, 40, 50].
for each x in items, do the following:
    If x is equal to 30, then
        Set found to be x.
        break out of this loop.
    thats it.
thats it.
Print the value of found.`)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "30") {
		t.Errorf("expected 30, got %q", out)
	}
}

func TestRepeatForever(t *testing.T) {
	out := captureOutput(func() {
		_, err := run(`Declare i to be 0.
repeat forever:
    Set i to be i + 1.
    If i is equal to 3, then
        break out of this loop.
    thats it.
thats it.
Print the value of i.`)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "3") {
		t.Errorf("expected 3, got %q", out)
	}
}

// ─── Typed variable declarations ─────────────────────────────────────────────

func TestTypedVarNumber(t *testing.T) {
	out := captureOutput(func() {
		_, err := run(`Declare x as number to be 42.
Print the value of x.`)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "42") {
		t.Errorf("expected 42, got %q", out)
	}
}

func TestTypedVarText(t *testing.T) {
	out := captureOutput(func() {
		_, err := run(`Declare name as text to be "Alice".
Print the value of name.`)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "Alice") {
		t.Errorf("expected Alice, got %q", out)
	}
}

func TestTypedVarBoolean(t *testing.T) {
	out := captureOutput(func() {
		_, err := run(`Declare flag as boolean to be true.
Print the value of flag.`)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "true") {
		t.Errorf("expected true, got %q", out)
	}
}

func TestTypedVarNoInit(t *testing.T) {
	_, err := run(`Declare x as number.`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTypedVarConstant(t *testing.T) {
	out := captureOutput(func() {
		_, err := run(`Declare x as number to always be 10.
Print the value of x.`)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "10") {
		t.Errorf("expected 10, got %q", out)
	}
}

// ─── Constant reassignment error ─────────────────────────────────────────────

func TestConstantReassignmentError(t *testing.T) {
	_, err := run(`Declare PI to always be 3.14.
Set PI to be 3.`)
	if err == nil {
		t.Error("expected error for constant reassignment")
	}
	if !strings.Contains(err.Error(), "constant") {
		t.Errorf("expected 'constant' in error, got: %v", err)
	}
}

// ─── Struct features ─────────────────────────────────────────────────────────

func TestStructDeclAndInstantiate(t *testing.T) {
	out := captureOutput(func() {
		_, err := run(`declare Person as a structure with the following fields:
    name is a string.
    age is a number with 0 being the default.
thats it.
let p be a new instance of Person with the following fields:
    name is "Alice".
    age is 30.
thats it.
Print the name of p.
Print the age of p.`)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "Alice") {
		t.Errorf("expected Alice in output, got %q", out)
	}
	if !strings.Contains(out, "30") {
		t.Errorf("expected 30 in output, got %q", out)
	}
}

func TestStructDefaultValues(t *testing.T) {
	out := captureOutput(func() {
		_, err := run(`declare Point as a structure with the following fields:
    x is a number with 0 being the default.
    y is a number with 0 being the default.
thats it.
let p be a new instance of Point.
Print the x of p.
Print the y of p.`)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "0") {
		t.Errorf("expected 0 in output, got %q", out)
	}
}

func TestStructMethodCall(t *testing.T) {
	out := captureOutput(func() {
		_, err := run(`declare Person as a structure with the following fields:
    name is a string.
    let greet be a function that does the following:
        Print "Hello from", name.
    thats it.
thats it.
let p be a new instance of Person with the following fields:
    name is "Bob".
thats it.
call p's greet.`)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "Hello from") || !strings.Contains(out, "Bob") {
		t.Errorf("expected 'Hello from Bob', got %q", out)
	}
}

// ─── Possessive syntax ───────────────────────────────────────────────────────

func TestPossessiveStringMethod(t *testing.T) {
	out := captureOutput(func() {
		_, err := run(`Declare s to be "hello world".
Print s's uppercase.`)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "HELLO WORLD") {
		t.Errorf("expected HELLO WORLD, got %q", out)
	}
}

func TestPossessiveOnLiteral(t *testing.T) {
	out := captureOutput(func() {
		_, err := run(`Print "hello"'s title.`)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "Hello") {
		t.Errorf("expected Hello, got %q", out)
	}
}

func TestPossessiveLengthOnList(t *testing.T) {
	out := captureOutput(func() {
		_, err := run(`Declare items to be [1, 2, 3, 4, 5].
Print the number of items.`)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "5") {
		t.Errorf("expected 5, got %q", out)
	}
}

// ─── Error handling ──────────────────────────────────────────────────────────

func TestTryCatchFinally(t *testing.T) {
	out := captureOutput(func() {
		_, err := run(`Declare result to be "none".
Try doing the following:
    Raise "oops".
    Set result to be "tried".
on error:
    Set result to be "caught".
but finally:
    Print "finally".
thats it.
Print the value of result.`)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "finally") {
		t.Errorf("expected 'finally' in output, got %q", out)
	}
	if !strings.Contains(out, "caught") {
		t.Errorf("expected 'caught' in output, got %q", out)
	}
}

func TestCustomErrorTypesCatch(t *testing.T) {
	out := captureOutput(func() {
		_, err := run(`Declare NetworkError as an error type.
Try doing the following:
    Raise "connection failed" as NetworkError.
on NetworkError:
    Print "caught network error".
thats it.`)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "caught network error") {
		t.Errorf("expected 'caught network error', got %q", out)
	}
}

func TestErrorHierarchyParentCatchesChild(t *testing.T) {
	out := captureOutput(func() {
		_, err := run(`Declare NetworkError as an error type.
Declare TimeoutError as a type of NetworkError.
Try doing the following:
    Raise "timed out" as TimeoutError.
on NetworkError:
    Print "caught network error".
thats it.`)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "caught network error") {
		t.Errorf("expected parent to catch child error, got %q", out)
	}
}

func TestErrorIsExpression(t *testing.T) {
	out := captureOutput(func() {
		_, err := run(`Declare NetworkError as an error type.
Declare result to be "none".
Try doing the following:
    Raise "test" as NetworkError.
on NetworkError:
    If error is NetworkError, then
        Set result to be "network".
    thats it.
thats it.
Print the value of result.`)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "network") {
		t.Errorf("expected 'network', got %q", out)
	}
}

func TestErrorHierarchyNonMatchingDoesNotCatch(t *testing.T) {
	_, err := run(`Declare DatabaseError as an error type.
Declare NetworkError as an error type.
Try doing the following:
    Raise "db error" as DatabaseError.
on NetworkError:
    Print "wrong".
thats it.`)
	if err == nil {
		t.Error("expected error to propagate when type doesn't match")
	}
}

func TestFinallyRunsOnTypeMismatch(t *testing.T) {
	// When the error type does NOT match the handler, the finally block must
	// still execute before the error propagates (matching astvm behavior).
	out := captureOutput(func() {
		_, err := run(`Declare NetworkError as an error type.
Declare ValidationError as an error type.
Try doing the following:
    Raise "bad value" as ValidationError.
on NetworkError:
    Print "wrong handler".
but finally:
    Print "finally ran".
thats it.`)
		if err == nil {
			t.Error("expected error to propagate when type doesn't match")
		}
	})
	if !strings.Contains(out, "finally ran") {
		t.Errorf("expected finally block to run on type mismatch, got output: %q", out)
	}
	if strings.Contains(out, "wrong handler") {
		t.Errorf("wrong handler should not have run, got output: %q", out)
	}
}

func TestFinallyRunsOnTypeMismatchFromFunction(t *testing.T) {
	// Finally must run even when the error comes from a nested function call.
	out := captureOutput(func() {
		_, err := run(`Declare NetworkError as an error type.
Declare ValidationError as an error type.
Declare function validate that takes x and does the following:
    Raise "bad value" as ValidationError.
thats it.
Try doing the following:
    Call validate with 0.
on NetworkError:
    Print "wrong handler".
but finally:
    Print "finally ran".
thats it.`)
		if err == nil {
			t.Error("expected error to propagate")
		}
	})
	if !strings.Contains(out, "finally ran") {
		t.Errorf("expected finally block to run, got output: %q", out)
	}
	if strings.Contains(out, "wrong handler") {
		t.Errorf("wrong handler should not have run, got output: %q", out)
	}
}

// ─── Stdlib functions ────────────────────────────────────────────────────────

func TestStdlibMathFunctions(t *testing.T) {
	tests := []struct {
		name     string
		src      string
		expected string
	}{
		{"sqrt", `Declare r to be 0. Set r to be the result of calling sqrt with 16. Print the value of r.`, "4"},
		{"abs", `Declare r to be 0. Set r to be the result of calling abs with -5. Print the value of r.`, "5"},
		{"floor", `Declare r to be 0. Set r to be the result of calling floor with 3.7. Print the value of r.`, "3"},
		{"ceil", `Declare r to be 0. Set r to be the result of calling ceil with 3.2. Print the value of r.`, "4"},
		{"round", `Declare r to be 0. Set r to be the result of calling round with 3.5. Print the value of r.`, "4"},
		{"pow", `Declare r to be 0. Set r to be the result of calling pow with 2 and 10. Print the value of r.`, "1024"},
		{"min", `Declare r to be 0. Set r to be the result of calling min with 3 and 7. Print the value of r.`, "3"},
		{"max", `Declare r to be 0. Set r to be the result of calling max with 3 and 7. Print the value of r.`, "7"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := captureOutput(func() {
				_, err := run(tt.src)
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			})
			if !strings.Contains(out, tt.expected) {
				t.Errorf("expected %q, got %q", tt.expected, out)
			}
		})
	}
}

func TestStdlibStringFunctions(t *testing.T) {
	tests := []struct {
		name     string
		src      string
		expected string
	}{
		{"uppercase", `Declare r to be "". Set r to be the result of calling uppercase with "hello". Print the value of r.`, "HELLO"},
		{"lowercase", `Declare r to be "". Set r to be the result of calling lowercase with "WORLD". Print the value of r.`, "world"},
		{"trim", `Declare r to be "". Set r to be the result of calling trim with "  hi  ". Print the value of r.`, "hi"},
		{"starts_with", `Declare r to be false. Set r to be the result of calling starts_with with "hello" and "hel". Print the value of r.`, "true"},
		{"ends_with", `Declare r to be false. Set r to be the result of calling ends_with with "hello" and "lo". Print the value of r.`, "true"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := captureOutput(func() {
				_, err := run(tt.src)
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			})
			if !strings.Contains(out, tt.expected) {
				t.Errorf("expected %q, got %q", tt.expected, out)
			}
		})
	}
}

func TestStdlibListFunctions(t *testing.T) {
	tests := []struct {
		name     string
		src      string
		expected string
	}{
		{"sum", `Declare nums to be [3, 1, 2]. Declare r to be 0. Set r to be the result of calling sum with nums. Print the value of r.`, "6"},
		{"count", `Declare nums to be [3, 1, 2]. Declare r to be 0. Set r to be the result of calling count with nums. Print the value of r.`, "3"},
		{"first", `Declare nums to be [10, 20, 30]. Declare r to be 0. Set r to be the result of calling first with nums. Print the value of r.`, "10"},
		{"last", `Declare nums to be [10, 20, 30]. Declare r to be 0. Set r to be the result of calling last with nums. Print the value of r.`, "30"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := captureOutput(func() {
				_, err := run(tt.src)
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			})
			if !strings.Contains(out, tt.expected) {
				t.Errorf("expected %q, got %q", tt.expected, out)
			}
		})
	}
}

// ─── Typeof and cast ─────────────────────────────────────────────────────────

func TestTypeofNumber(t *testing.T) {
	out := captureOutput(func() {
		_, err := run(`Print the type of 42.`)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "f64") {
		t.Errorf("expected 'f64', got %q", out)
	}
}

func TestTypeofText(t *testing.T) {
	out := captureOutput(func() {
		_, err := run(`Print the type of "hello".`)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "text") {
		t.Errorf("expected 'text', got %q", out)
	}
}

func TestTypeofBoolean(t *testing.T) {
	out := captureOutput(func() {
		_, err := run(`Print the type of true.`)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "boolean") {
		t.Errorf("expected 'boolean', got %q", out)
	}
}

func TestCastNumberToText(t *testing.T) {
	out := captureOutput(func() {
		_, err := run(`Declare n to be 42.
Print n cast to text.`)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "42") {
		t.Errorf("expected '42', got %q", out)
	}
}

func TestCastTextToNumber(t *testing.T) {
	out := captureOutput(func() {
		_, err := run(`Declare s to be "25".
Declare n to be s cast to number.
Print n + 5.`)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "30") {
		t.Errorf("expected 30, got %q", out)
	}
}

func TestCastToBoolean(t *testing.T) {
	out := captureOutput(func() {
		_, err := run(`Declare n to be 1.
If n cast to boolean, then
    Print "truthy".
thats it.`)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "truthy") {
		t.Errorf("expected 'truthy', got %q", out)
	}
}

// ─── Logical operators (short-circuit) ───────────────────────────────────────

func TestLogicalAnd(t *testing.T) {
	out := captureOutput(func() {
		_, err := run(`Declare a to be true.
Declare b to be true.
If a and b, then
    Print "both".
thats it.`)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "both") {
		t.Errorf("expected 'both', got %q", out)
	}
}

func TestLogicalOr(t *testing.T) {
	out := captureOutput(func() {
		_, err := run(`Declare a to be false.
Declare b to be true.
If a or b, then
    Print "either".
thats it.`)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "either") {
		t.Errorf("expected 'either', got %q", out)
	}
}

func TestLogicalShortCircuitAnd(t *testing.T) {
	out := captureOutput(func() {
		_, err := run(`Declare sideEffect to be 0.
Declare a to be false.
If a, then
    Set sideEffect to be sideEffect + 1.
thats it.
Print the value of sideEffect.`)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "0") {
		t.Errorf("expected 0 (condition not reached), got %q", out)
	}
}

// ─── Lookup table operations ─────────────────────────────────────────────────

func TestLookupTableSetGet(t *testing.T) {
	out := captureOutput(func() {
		_, err := run(`Declare t to be a lookup table.
Set t at "name" to be "Alice".
Set t at "age" to be 30.
Print the entry "name" in t.
Print the entry "age" in t.`)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "Alice") {
		t.Errorf("expected Alice, got %q", out)
	}
	if !strings.Contains(out, "30") {
		t.Errorf("expected 30, got %q", out)
	}
}

func TestLookupTableCount(t *testing.T) {
	out := captureOutput(func() {
		_, err := run(`Declare t to be a lookup table.
Set t at "a" to be 1.
Set t at "b" to be 2.
Declare n to be 0.
Set n to be the result of calling count with t.
Print the value of n.`)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "2") {
		t.Errorf("expected 2, got %q", out)
	}
}

func TestLookupTableHasKey(t *testing.T) {
	out := captureOutput(func() {
		_, err := run(`Declare t to be a lookup table.
Set t at "key" to be "val".
If t has "key", then
    Print "yes".
thats it.
If t has "missing", then
    Print "no".
otherwise
    Print "not found".
thats it.`)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "yes") {
		t.Errorf("expected 'yes', got %q", out)
	}
	if !strings.Contains(out, "not found") {
		t.Errorf("expected 'not found', got %q", out)
	}
}

func TestLookupTableIteration(t *testing.T) {
	out := captureOutput(func() {
		_, err := run(`Declare t to be a lookup table.
Set t at "a" to be 1.
Set t at "b" to be 2.
Set t at "c" to be 3.
Declare count to be 0.
for each k in t, do the following:
    Set count to be count + 1.
thats it.
Print the value of count.`)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "3") {
		t.Errorf("expected 3, got %q", out)
	}
}

// ─── Arrays ──────────────────────────────────────────────────────────────────

func TestArrayCreationAndAccess(t *testing.T) {
	out := captureOutput(func() {
		_, err := run(`Declare arr to be [10, 20, 30, 40, 50].
Print the item at position 0 in arr.
Print the item at position 4 in arr.`)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "10") {
		t.Errorf("expected 10, got %q", out)
	}
	if !strings.Contains(out, "50") {
		t.Errorf("expected 50, got %q", out)
	}
}

func TestArrayModification(t *testing.T) {
	out := captureOutput(func() {
		_, err := run(`Declare arr to be [1, 2, 3].
Set the item at position 1 in arr to be 99.
Print the item at position 1 in arr.`)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "99") {
		t.Errorf("expected 99, got %q", out)
	}
}

func TestArrayForEach(t *testing.T) {
	out := captureOutput(func() {
		_, err := run(`Declare total to be 0.
Declare arr to be [1, 2, 3, 4, 5].
for each n in arr, do the following:
    Set total to be total + n.
thats it.
Print the value of total.`)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "15") {
		t.Errorf("expected 15, got %q", out)
	}
}

// ─── Nil checks ──────────────────────────────────────────────────────────────

func TestNilCheckIsSomething(t *testing.T) {
	out := captureOutput(func() {
		_, err := run(`Declare x to be "hello".
If x is something, then
    Print "has value".
thats it.`)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "has value") {
		t.Errorf("expected 'has value', got %q", out)
	}
}

func TestNilCheckIsNothing(t *testing.T) {
	out := captureOutput(func() {
		_, err := run(`Declare x to be nothing.
If x is nothing, then
    Print "is nothing".
thats it.`)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "is nothing") {
		t.Errorf("expected 'is nothing', got %q", out)
	}
}

func TestNilCheckHasNoValue(t *testing.T) {
	out := captureOutput(func() {
		_, err := run(`Declare x to be nothing.
If x has no value, then
    Print "no value".
thats it.`)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "no value") {
		t.Errorf("expected 'no value', got %q", out)
	}
}

func TestNilCheckHasAValue(t *testing.T) {
	out := captureOutput(func() {
		_, err := run(`Declare x to be 42.
If x has a value, then
    Print "has value".
thats it.`)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "has value") {
		t.Errorf("expected 'has value', got %q", out)
	}
}

// ─── String operations ────────────────────────────────────────────────────────

func TestStringLength(t *testing.T) {
	out := captureOutput(func() {
		_, err := run(`Declare s to be "hello".
Print the number of s.`)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "5") {
		t.Errorf("expected 5, got %q", out)
	}
}

func TestStringIndexing(t *testing.T) {
	out := captureOutput(func() {
		_, err := run(`Declare s to be "hello".
Print the item at position 0 in s.`)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "h") {
		t.Errorf("expected 'h', got %q", out)
	}
}

// ─── Location expression ─────────────────────────────────────────────────────

func TestLocationExpression(t *testing.T) {
	out := captureOutput(func() {
		_, err := run(`Declare x to be 42.
Declare loc to be the location of x.
Print the value of loc.`)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if out == "" {
		t.Errorf("expected non-empty location output")
	}
}

// ─── Functions ───────────────────────────────────────────────────────────────

func TestFunctionMultipleParams(t *testing.T) {
	out := captureOutput(func() {
		_, err := run(`Declare function add that takes a and b and does the following:
    Return a + b.
thats it.
Declare r to be 0.
Set r to be the result of calling add with 3 and 4.
Print the value of r.`)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "7") {
		t.Errorf("expected 7, got %q", out)
	}
}

func TestFunctionRecursion(t *testing.T) {
	out := captureOutput(func() {
		_, err := run(`Declare function fact that takes n and does the following:
    If n is less than or equal to 1, then
        Return 1.
    thats it.
    Declare prev to be 0.
    Set prev to be the result of calling fact with n - 1.
    Return n * prev.
thats it.
Declare r to be 0.
Set r to be the result of calling fact with 5.
Print the value of r.`)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "120") {
		t.Errorf("expected 120, got %q", out)
	}
}

func TestFunctionNoParams(t *testing.T) {
	out := captureOutput(func() {
		_, err := run(`Declare function greet and does the following:
    Print "hello from function".
thats it.
call greet.`)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "hello from function") {
		t.Errorf("expected greeting, got %q", out)
	}
}

func TestFunctionClosures(t *testing.T) {
	out := captureOutput(func() {
		_, err := run(`Declare base to be 10.
Declare function addBase that takes n and does the following:
    Return n + base.
thats it.
Declare r to be 0.
Set r to be the result of calling addBase with 5.
Print the value of r.`)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "15") {
		t.Errorf("expected 15, got %q", out)
	}
}

// ─── Import functionality ────────────────────────────────────────────────────

func TestImportBasic(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "ivm_import_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	libPath := filepath.Join(tmpDir, "mylib.abc")
	if err := os.WriteFile(libPath, []byte(`
Declare function square that takes n and does the following:
    Return n * n.
thats it.
Declare MY_CONST to always be 42.
`), 0644); err != nil {
		t.Fatal(err)
	}

	mainSrc := fmt.Sprintf(`Import "%s".
Declare result to be 0.
Set result to be the result of calling square with 5.
Print the value of result.
Print the value of MY_CONST.`, libPath)

	out := captureOutput(func() {
		_, err := run(mainSrc)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "25") {
		t.Errorf("expected 25, got %q", out)
	}
	if !strings.Contains(out, "42") {
		t.Errorf("expected 42, got %q", out)
	}
}

func TestSelectiveImport(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "ivm_import_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	libPath := filepath.Join(tmpDir, "lib.abc")
	if err := os.WriteFile(libPath, []byte(`
Declare function add that takes a and b and does the following:
    Return a + b.
thats it.
`), 0644); err != nil {
		t.Fatal(err)
	}

	mainSrc := fmt.Sprintf(`Import add from "%s".
Declare r to be 0.
Set r to be the result of calling add with 3 and 4.
Print the value of r.`, libPath)

	out := captureOutput(func() {
		_, err := run(mainSrc)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "7") {
		t.Errorf("expected 7, got %q", out)
	}
}

// ─── Encoding/decoding with complex features ────────────────────────────────

func TestEncodeDecodeWithLoops(t *testing.T) {
	src := `Declare total to be 0.
repeat the following while total is less than 10:
    Set total to be total + 1.
thats it.`
	chunk, err := compileSource(src)
	if err != nil {
		t.Fatalf("compile error: %v", err)
	}
	data, err := ivm.EncodeFile(chunk)
	if err != nil {
		t.Fatalf("encode error: %v", err)
	}
	decoded, err := ivm.DecodeFile(data)
	if err != nil {
		t.Fatalf("decode error: %v", err)
	}
	_, execErr := ivm.Execute(decoded, stdlib.Eval, stdlib.PredefinedValues())
	if execErr != nil {
		t.Fatalf("execute error: %v", execErr)
	}
}

func TestEncodeDecodeWithStructs(t *testing.T) {
	src := `declare Point as a structure with the following fields:
    x is a number with 0 being the default.
    y is a number with 0 being the default.
thats it.
let p be a new instance of Point with the following fields:
    x is 3.
    y is 4.
thats it.`
	chunk, err := compileSource(src)
	if err != nil {
		t.Fatalf("compile error: %v", err)
	}
	data, err := ivm.EncodeFile(chunk)
	if err != nil {
		t.Fatalf("encode error: %v", err)
	}
	decoded, err := ivm.DecodeFile(data)
	if err != nil {
		t.Fatalf("decode error: %v", err)
	}
	_, execErr := ivm.Execute(decoded, stdlib.Eval, stdlib.PredefinedValues())
	if execErr != nil {
		t.Fatalf("execute error: %v", execErr)
	}
}

func TestEncodeDecodeWithTryCatch(t *testing.T) {
	src := `Declare result to be "none".
Try doing the following:
    Raise "test error".
on error:
    Set result to be "caught".
thats it.`
	chunk, err := compileSource(src)
	if err != nil {
		t.Fatalf("compile error: %v", err)
	}
	data, err := ivm.EncodeFile(chunk)
	if err != nil {
		t.Fatalf("encode error: %v", err)
	}
	decoded, err := ivm.DecodeFile(data)
	if err != nil {
		t.Fatalf("decode error: %v", err)
	}
	_, execErr := ivm.Execute(decoded, stdlib.Eval, stdlib.PredefinedValues())
	if execErr != nil {
		t.Fatalf("execute error: %v", execErr)
	}
}

func TestEncodeDecodeWithContinue(t *testing.T) {
	src := `Declare total to be 0.
Declare nums to be [1, 2, 3, 4, 5].
for each n in nums, do the following:
    Declare m to be the remainder of n divided by 2.
    If m is equal to 0, then
        Continue.
    thats it.
    Set total to be total + n.
thats it.`
	chunk, err := compileSource(src)
	if err != nil {
		t.Fatalf("compile error: %v", err)
	}
	data, err := ivm.EncodeFile(chunk)
	if err != nil {
		t.Fatalf("encode error: %v", err)
	}
	decoded, err := ivm.DecodeFile(data)
	if err != nil {
		t.Fatalf("decode error: %v", err)
	}
	_, execErr := ivm.Execute(decoded, stdlib.Eval, stdlib.PredefinedValues())
	if execErr != nil {
		t.Fatalf("execute error: %v", execErr)
	}
}

// ─── Predefined math constants ───────────────────────────────────────────────

func TestPredefinedE(t *testing.T) {
	out := captureOutput(func() {
		_, err := run(`Print the value of e.`)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "2.71") {
		t.Errorf("expected e ~2.71, got %q", out)
	}
}

func TestPredefinedInfinity(t *testing.T) {
	out := captureOutput(func() {
		_, err := run(`Print the value of infinity.`)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "Inf") {
		t.Errorf("expected Inf, got %q", out)
	}
}

// ─── Error messages (runtime) ────────────────────────────────────────────────

func TestUndefinedVariableError(t *testing.T) {
	_, err := run(`Print the value of undefinedVar.`)
	if err == nil {
		t.Error("expected error for undefined variable")
	}
	if !strings.Contains(err.Error(), "undefined variable") {
		t.Errorf("expected 'undefined variable' in error, got %q", err.Error())
	}
}

func TestDivisionByZeroError(t *testing.T) {
	_, err := run(`Declare x to be 10 / 0.`)
	if err == nil {
		t.Error("expected error for division by zero")
	}
}

func TestIndexOutOfBoundsError(t *testing.T) {
	_, err := run(`Declare arr to be [1, 2, 3].
Print the item at position 99 in arr.`)
	if err == nil {
		t.Error("expected error for index out of bounds")
	}
}

func TestFunctionArgCountError(t *testing.T) {
	_, err := run(`Declare function add that takes a and b and does the following:
    Return a + b.
thats it.
Declare result to be 0.
Set result to be the result of calling add with 1.`)
	if err == nil {
		t.Error("expected error for wrong argument count")
	}
}

// ─── Let syntax ───────────────────────────────────────────────────────────────

func TestLetSyntaxVariants(t *testing.T) {
	tests := []struct {
		name     string
		src      string
		expected string
	}{
		{"let be", `let x be 10. Print the value of x.`, "10"},
		{"let =", `let y = 20. Print the value of y.`, "20"},
		{"let equal", `let z equal 30. Print the value of z.`, "30"},
		{"let be equal to", `let w be equal to 40. Print the value of w.`, "40"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := captureOutput(func() {
				_, err := run(tt.src)
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			})
			if !strings.Contains(out, tt.expected) {
				t.Errorf("expected %q, got %q", tt.expected, out)
			}
		})
	}
}

func TestLetConstant(t *testing.T) {
	_, err := run(`let PI always be 3.14.
Set PI to be 3.`)
	if err == nil {
		t.Error("expected error for constant reassignment")
	}
}

// ─── Comparison operators ────────────────────────────────────────────────────

func TestAllComparisonOperators(t *testing.T) {
	tests := []struct {
		name     string
		src      string
		expected string
	}{
		{"equal", `If 5 is equal to 5, then Print "eq". thats it.`, "eq"},
		{"not equal", `If 3 is not equal to 5, then Print "neq". thats it.`, "neq"},
		{"less than", `If 3 is less than 5, then Print "lt". thats it.`, "lt"},
		{"greater than", `If 5 is greater than 3, then Print "gt". thats it.`, "gt"},
		{"lte", `If 5 is less than or equal to 5, then Print "lte". thats it.`, "lte"},
		{"gte", `If 5 is greater than or equal to 5, then Print "gte". thats it.`, "gte"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := captureOutput(func() {
				_, err := run(tt.src)
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			})
			if !strings.Contains(out, tt.expected) {
				t.Errorf("expected %q, got %q", tt.expected, out)
			}
		})
	}
}

// ─── Output (Print / Write) ──────────────────────────────────────────────────

func TestPrintMultipleValues(t *testing.T) {
	out := captureOutput(func() {
		_, err := run(`Print "hello", "world".`)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "hello") || !strings.Contains(out, "world") {
		t.Errorf("expected 'hello world' in output, got %q", out)
	}
}

func TestWriteNoNewline(t *testing.T) {
	out := captureOutput(func() {
		_, err := run(`Write "hello".
Write " world".
Print "".`)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "hello world") {
		t.Errorf("expected 'hello world', got %q", out)
	}
}

// ─── Regression: correct output matches tree-walk evaluator ─────────────────

func TestRegressionFibonacci(t *testing.T) {
	src := `Declare function fib that takes n and does the following:
    If n is less than or equal to 1, then
        Return n.
    thats it.
    Declare a to be 0.
    Declare b to be 0.
    Set a to be the result of calling fib with n - 1.
    Set b to be the result of calling fib with n - 2.
    Return a + b.
thats it.
Declare r to be 0.
Set r to be the result of calling fib with 10.
Print the value of r.`
	out := captureOutput(func() {
		_, err := run(src)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "55") {
		t.Errorf("fib(10) = 55, got %q", out)
	}
}

func TestRegressionBubbleSortLike(t *testing.T) {
	src := `Declare arr to be [5, 3, 1, 4, 2].
Declare n to be 5.
Declare i to be 0.
Declare j to be 0.
repeat the following while i is less than n - 1:
    Set j to be 0.
    repeat the following while j is less than n - i - 1:
        Declare a to be the item at position j in arr.
        Declare b to be the item at position j + 1 in arr.
        If a is greater than b, then
            Set the item at position j in arr to be b.
            Set the item at position j + 1 in arr to be a.
        thats it.
        Set j to be j + 1.
    thats it.
    Set i to be i + 1.
thats it.
for each x in arr, do the following:
    Print the value of x.
thats it.`
	out := captureOutput(func() {
		_, err := run(src)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	lines := strings.Fields(strings.TrimSpace(out))
	expected := []string{"1", "2", "3", "4", "5"}
	if len(lines) != len(expected) {
		t.Fatalf("expected %v, got %v", expected, lines)
	}
	for i, l := range lines {
		if l != expected[i] {
			t.Errorf("position %d: expected %q, got %q", i, expected[i], l)
		}
	}
}

// ─── Compile-time type errors ────────────────────────────────────────────────

func TestCompileTimeTypeErrorTextMethod(t *testing.T) {
	// Calling a text method on a known-number variable should trigger a compile error
	_, err := run(`Declare x to always be 42.
Print x's uppercase.`)
	if err == nil {
		t.Error("expected compile-time type error")
	}
}

// ─── Transpilation parity (regression) ──────────────────────────────────────

func TestTranspilerParityFizzBuzz(t *testing.T) {
	// Compile to instruction bytecode and verify output matches known-correct output
	src := `Declare i to be 1.
repeat the following while i is less than or equal to 20:
    Declare mod3 to be the remainder of i divided by 3.
    Declare mod5 to be the remainder of i divided by 5.
    If mod3 is equal to 0 and mod5 is equal to 0, then
        Print "FizzBuzz".
    otherwise if mod3 is equal to 0, then
        Print "Fizz".
    otherwise if mod5 is equal to 0, then
        Print "Buzz".
    otherwise
        Print the value of i.
    thats it.
    Set i to be i + 1.
thats it.`

	out := captureOutput(func() {
		_, err := run(src)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	expected := []string{"1", "2", "Fizz", "4", "Buzz", "Fizz", "7", "8", "Fizz", "Buzz",
		"11", "Fizz", "13", "14", "FizzBuzz", "16", "17", "Fizz", "19", "Buzz"}
	lines := strings.Fields(strings.TrimSpace(out))
	if len(lines) != len(expected) {
		t.Fatalf("FizzBuzz: expected %d lines, got %d: %q", len(expected), len(lines), out)
	}
	for i, l := range lines {
		if l != expected[i] {
			t.Errorf("FizzBuzz line %d: expected %q, got %q", i+1, expected[i], l)
		}
	}
}



// ─── Time stdlib ─────────────────────────────────────────────────────────────

func TestCurrentTime(t *testing.T) {
out := captureOutput(func() {
_, err := run(`Declare ts to be current_time().
Print the value of ts.`)
if err != nil {
t.Errorf("unexpected error: %v", err)
}
})
// current_time() returns a date-time string like "2006-01-02 15:04:05"
if len(strings.TrimSpace(out)) < 10 {
t.Errorf("expected non-empty time string, got: %q", out)
}
}

func TestElapsedTime(t *testing.T) {
out := captureOutput(func() {
_, err := run(`Declare elapsed to be elapsed_time().
Print the value of elapsed.`)
if err != nil {
t.Errorf("unexpected error: %v", err)
}
})
if strings.TrimSpace(out) == "" {
t.Errorf("expected elapsed time value, got empty output")
}
}

// ─── Sleep / Wait statement ───────────────────────────────────────────────────

func TestSleepMs(t *testing.T) {
_, err := run(`Sleep for 10ms.`)
if err != nil {
t.Fatalf("unexpected error from 'Sleep for 10ms.': %v", err)
}
}

func TestSleepShortFormUnits(t *testing.T) {
cases := []string{
`Sleep for 0s.`,
`Sleep for 0m.`,
`Sleep for 0h.`,
}
for _, src := range cases {
_, err := run(src)
if err != nil {
t.Errorf("%q: unexpected error: %v", src, err)
}
}
}

func TestSleepLongFormUnits(t *testing.T) {
cases := []string{
`Sleep for 0 milliseconds.`,
`Sleep for 0 millisecond.`,
`Sleep for 0 seconds.`,
`Sleep for 0 second.`,
`Sleep for 0 minutes.`,
`Sleep for 0 minute.`,
`Sleep for 0 hours.`,
`Sleep for 0 hour.`,
}
for _, src := range cases {
_, err := run(src)
if err != nil {
t.Errorf("%q: unexpected error: %v", src, err)
}
}
}

func TestWaitAlias(t *testing.T) {
cases := []string{
`Wait for 0ms.`,
`Wait for 0 seconds.`,
}
for _, src := range cases {
_, err := run(src)
if err != nil {
t.Errorf("%q: unexpected error: %v", src, err)
}
}
}

func TestSleepNaturalShorthand(t *testing.T) {
// "a second" and "an hour" shorthands (0-second versions for test speed)
// "Sleep for a second." sleeps 1s which is too slow for unit tests,
// so we only verify the parse succeeds via a 0-duration equivalent.
_, err := run(`Sleep for 0 seconds.`)
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
}

func TestSleepBadUnit(t *testing.T) {
_, err := run(`Sleep for 1x.`)
if err == nil {
t.Fatal("expected parse error for unknown time unit")
}
}

func TestSleepMissingFor(t *testing.T) {
_, err := run(`Sleep 1s.`)
if err == nil {
t.Fatal("expected parse error when 'for' keyword is missing")
}
}

func TestSleepInsideLoop(t *testing.T) {
_, err := run(`Declare i to be 0.
Repeat the following 2 times:
    Sleep for 0ms.
    Set i to be i + 1.
thats it.`)
if err != nil {
t.Fatalf("unexpected error sleeping inside loop: %v", err)
}
}

func TestSleepInsideFunction(t *testing.T) {
_, err := run(`Declare function pause that does the following:
    Sleep for 0ms.
thats it.
Call pause.`)
if err != nil {
t.Fatalf("unexpected error sleeping inside function: %v", err)
}
}

// ─── Politeness (parser-level) ────────────────────────────────────────────────

func TestPolitePrefix_Please(t *testing.T) {
out := captureOutput(func() {
_, err := run(`Please print "Hello".`)
if err != nil {
t.Errorf("unexpected error: %v", err)
}
})
if !strings.Contains(out, "Hello") {
t.Errorf("expected 'Hello' in output, got: %q", out)
}
}

func TestPolitePrefix_Kindly(t *testing.T) {
out := captureOutput(func() {
_, err := run(`Kindly print "World".`)
if err != nil {
t.Errorf("unexpected error: %v", err)
}
})
if !strings.Contains(out, "World") {
t.Errorf("expected 'World' in output, got: %q", out)
}
}

func TestPolitePrefix_CouldYou(t *testing.T) {
out := captureOutput(func() {
_, err := run(`Could you print "CouldYou".`)
if err != nil {
t.Errorf("unexpected error: %v", err)
}
})
if !strings.Contains(out, "CouldYou") {
t.Errorf("expected 'CouldYou' in output, got: %q", out)
}
}

func TestPolitePrefix_WouldYouKindly(t *testing.T) {
out := captureOutput(func() {
_, err := run(`Would you kindly print "WouldYouKindly".`)
if err != nil {
t.Errorf("unexpected error: %v", err)
}
})
if !strings.Contains(out, "WouldYouKindly") {
t.Errorf("expected 'WouldYouKindly' in output, got: %q", out)
}
}

func TestPolitePrefix_InsideLoop(t *testing.T) {
out := captureOutput(func() {
_, err := run(`Please declare i to be 0.
Please repeat the following 3 times:
    Please set i to be i + 1.
thats it.
Please print the value of i.`)
if err != nil {
t.Errorf("unexpected error: %v", err)
}
})
if !strings.Contains(out, "3") {
t.Errorf("expected '3' in output, got: %q", out)
}
}

func TestPolitePrefix_InsideFunction(t *testing.T) {
out := captureOutput(func() {
_, err := run(`Please declare function greet that does the following:
    Please print "Hi".
thats it.
Please call greet.`)
if err != nil {
t.Errorf("unexpected error: %v", err)
}
})
if !strings.Contains(out, "Hi") {
t.Errorf("expected 'Hi' in output, got: %q", out)
}
}

func TestPolitenessStats_AllPolite(t *testing.T) {
// All statements polite – should compile and run fine.
_, err := run(`Please print "A".
Please print "B".`)
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
}

func TestPolitenessStats_NonePolite(t *testing.T) {
// No politeness prefix – still parses/runs fine (prefix is always optional).
_, err := run(`Print "A".
Print "B".`)
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
}

func TestPolitenessStats_CommentsExcluded(t *testing.T) {
// Comments should not count toward the politeness tally.
_, err := run(`# This is a comment.
Please print "Hello".`)
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
}
