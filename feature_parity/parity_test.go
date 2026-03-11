// Package feature_parity verifies that the AST-walk VM (english/astvm) and the
// instruction VM (english/ivm) produce byte-for-byte identical stdout output
// for every English program.  Every test in this file runs the same source
// program through both VMs and asserts that:
//
//  1. Both VMs succeed (no error returned), OR
//  2. Both VMs fail (returns a non-nil error, runtime or compile-time).
//
// The captured stdout must be identical in both cases (point 1).
//
// Convention: whenever a new language feature is added, it MUST be implemented
// in both english/astvm AND english/ivm, and a parity test MUST be added here
// to confirm that both implementations produce the same output.
package feature_parity_test

import (
	"bytes"
	"english/astvm"
	"english/astvm/stdlib"
	"english/ivm"
	"english/parser"
	"io"
	"os"
	"strings"
	"testing"
)

// ─── Helpers ──────────────────────────────────────────────────────────────────

// runAST executes src through the AST-walk VM and returns captured stdout.
func runAST(src string) (string, error) {
	var out string
	var runErr error
	out = captureStdout(func() {
		lexer := parser.NewLexer(src)
		tokens := lexer.TokenizeAll()
		p := parser.NewParser(tokens)
		program, err := p.Parse()
		if err != nil {
			runErr = err
			return
		}
		env := vm.NewEnvironment()
		stdlib.Register(env)
		evaluator := vm.NewEvaluator(env, stdlib.Eval)
		_, runErr = evaluator.Eval(program)
	})
	return out, runErr
}

// runIVM executes src through the instruction VM and returns captured stdout.
func runIVM(src string) (string, error) {
	var out string
	var runErr error
	out = captureStdout(func() {
		lexer := parser.NewLexer(src)
		tokens := lexer.TokenizeAll()
		p := parser.NewParser(tokens)
		prog, err := p.Parse()
		if err != nil {
			runErr = err
			return
		}
		chunk, err := ivm.Compile(prog)
		if err != nil {
			runErr = err
			return
		}
		_, runErr = ivm.Execute(chunk, stdlib.Eval, stdlib.PredefinedValues())
	})
	return out, runErr
}

// captureStdout redirects stdout during fn and returns the captured text.
func captureStdout(fn func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	fn()
	w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

// assertParity is the core assertion: both VMs must produce the same stdout.
// If expectErr is true, both VMs must also return a non-nil error.
func assertParity(t *testing.T, src string) {
	t.Helper()
	astOut, astErr := runAST(src)
	ivmOut, ivmErr := runIVM(src)

	// Error agreement
	if (astErr != nil) != (ivmErr != nil) {
		t.Errorf("error parity mismatch:\n  astvm err: %v\n  ivm   err: %v", astErr, ivmErr)
	}

	// Stdout agreement
	if astOut != ivmOut {
		t.Errorf("output parity mismatch for:\n%s\n\nastvm output:\n%q\n\nivm output:\n%q",
			src, astOut, ivmOut)
	}
}

// assertParityError is like assertParity but also asserts that BOTH VMs fail.
func assertParityError(t *testing.T, src string) {
	t.Helper()
	_, astErr := runAST(src)
	_, ivmErr := runIVM(src)
	if astErr == nil {
		t.Error("expected astvm to return an error, but it succeeded")
	}
	if ivmErr == nil {
		t.Error("expected ivm to return an error, but it succeeded")
	}
}

// ─── Basic Output ─────────────────────────────────────────────────────────────

func TestParityHelloWorld(t *testing.T) {
	assertParity(t, `Print "Hello, World!".`)
}

func TestParityMultiplePrints(t *testing.T) {
	assertParity(t, `Print "line1".
Print "line2".
Print "line3".`)
}

func TestParityPrintNumber(t *testing.T) {
	assertParity(t, `Print 42.`)
}

func TestParityPrintBoolean(t *testing.T) {
	assertParity(t, `Print true.
Print false.`)
}

// ─── Variables ────────────────────────────────────────────────────────────────

func TestParityVariableDeclaration(t *testing.T) {
	assertParity(t, `Declare x to be 5.
Print x.`)
}

func TestParityVariableReassignment(t *testing.T) {
	assertParity(t, `Declare x to be 1.
Set x to be 99.
Print x.`)
}

func TestParityStringVariable(t *testing.T) {
	assertParity(t, `Declare msg to be "hello".
Print msg.`)
}

func TestParityTypedVariableNumber(t *testing.T) {
	assertParity(t, `Declare x as number to be 3.14.
Print x.`)
}

func TestParityTypedVariableText(t *testing.T) {
	assertParity(t, `Declare s as text to be "world".
Print s.`)
}

func TestParityTypedVariableBoolean(t *testing.T) {
	assertParity(t, `Declare b as boolean to be true.
Print b.`)
}

func TestParityConstant(t *testing.T) {
	assertParity(t, `Declare PI as a constant to be 3.14159.
Print PI.`)
}

// ─── Arithmetic ───────────────────────────────────────────────────────────────

func TestParityAddition(t *testing.T) {
	assertParity(t, `Print 3 + 4.`)
}

func TestParitySubtraction(t *testing.T) {
	assertParity(t, `Print 10 - 3.`)
}

func TestParityMultiplication(t *testing.T) {
	assertParity(t, `Print 6 * 7.`)
}

func TestParityDivision(t *testing.T) {
	assertParity(t, `Print 10 / 4.`)
}

func TestParityModulo(t *testing.T) {
	assertParity(t, `Print the remainder of 17 divided by 5.`)
}

func TestParityCompoundArithmetic(t *testing.T) {
	assertParity(t, `Declare x to be (3 + 4) * 2.
Print x.`)
}

// ─── Comparisons ─────────────────────────────────────────────────────────────

func TestParityEquality(t *testing.T) {
	assertParity(t, `Print (5 is equal to 5).
Print (5 is equal to 6).`)
}

func TestParityInequality(t *testing.T) {
	assertParity(t, `Print (5 is not equal to 4).`)
}

func TestParityLessThan(t *testing.T) {
	assertParity(t, `Print (3 is less than 5).`)
}

func TestParityGreaterThan(t *testing.T) {
	assertParity(t, `Print (7 is greater than 4).`)
}

// ─── Logical Operators ────────────────────────────────────────────────────────

func TestParityLogicalAnd(t *testing.T) {
	assertParity(t, `Declare a to be true.
Declare b to be false.
Print (a and b).`)
}

func TestParityLogicalOr(t *testing.T) {
	assertParity(t, `Declare a to be true.
Declare b to be false.
Print (a or b).`)
}

func TestParityLogicalNot(t *testing.T) {
	assertParity(t, `Declare a to be true.
Print (not a).`)
}

func TestParityShortCircuitAnd(t *testing.T) {
	// If short-circuit works, the second operand is never evaluated
	assertParity(t, `Declare x to be 0.
If (false and (x is equal to 1)), then
    Print "inside".
thats it.
Print "done".`)
}

func TestParityShortCircuitOr(t *testing.T) {
	assertParity(t, `Declare x to be 0.
If (true or (x is equal to 1)), then
    Print "inside".
thats it.
Print "done".`)
}

// ─── Conditionals ────────────────────────────────────────────────────────────

func TestParityIfTrue(t *testing.T) {
	assertParity(t, `If true, then
    Print "yes".
thats it.`)
}

func TestParityIfFalse(t *testing.T) {
	assertParity(t, `If false, then
    Print "yes".
thats it.
Print "no".`)
}

func TestParityIfElse(t *testing.T) {
	assertParity(t, `Declare x to be 10.
If x is greater than 5, then
    Print "big".
Otherwise:
    Print "small".
thats it.`)
}

func TestParityIfElseIf(t *testing.T) {
	assertParity(t, `Declare n to be 5.
If n is less than 0, then
    Print "negative".
Otherwise, if n is equal to 0, then
    Print "zero".
Otherwise:
    Print "positive".
thats it.`)
}

func TestParityNestedIf(t *testing.T) {
	assertParity(t, `Declare x to be 3.
Declare y to be 7.
If x is less than 5, then
    If y is greater than 5, then
        Print "both".
    thats it.
thats it.`)
}

// ─── While Loops ─────────────────────────────────────────────────────────────

func TestParityWhileLoop(t *testing.T) {
	assertParity(t, `Declare i to be 1.
repeat the following while i is less than or equal to 5:
    Print i.
    Set i to be i + 1.
thats it.`)
}

func TestParityWhileBreak(t *testing.T) {
	assertParity(t, `Declare i to be 0.
repeat the following while true:
    Set i to be i + 1.
    If i is equal to 3, then
        Break.
    thats it.
thats it.
Print i.`)
}

func TestParityWhileContinue(t *testing.T) {
	assertParity(t, `Declare i to be 0.
Declare total to be 0.
repeat the following while i is less than 10:
    Set i to be i + 1.
    Declare m to be the remainder of i divided by 2.
    If m is equal to 0, then
        Continue.
    thats it.
    Set total to be total + i.
thats it.
Print total.`)
}

// ─── Repeat-N Loops ──────────────────────────────────────────────────────────

func TestParityRepeatN(t *testing.T) {
	assertParity(t, `Declare count to be 0.
repeat the following 5 times:
    Set count to be count + 1.
thats it.
Print count.`)
}

func TestParityRepeatNContinue(t *testing.T) {
	assertParity(t, `Declare count to be 0.
Declare printed to be 0.
repeat the following 5 times:
    Set count to be count + 1.
    If count is equal to 3, then
        Continue.
    thats it.
    Set printed to be printed + 1.
thats it.
Print printed.`)
}

// ─── For-Each Loops ──────────────────────────────────────────────────────────

func TestParityForEach(t *testing.T) {
	assertParity(t, `Declare nums to be [1, 2, 3].
for each n in nums, do the following:
    Print n.
thats it.`)
}

func TestParityForEachBreak(t *testing.T) {
	assertParity(t, `Declare nums to be [1, 2, 3, 4, 5].
for each n in nums, do the following:
    If n is equal to 3, then
        Break.
    thats it.
    Print n.
thats it.`)
}

func TestParityForEachContinue(t *testing.T) {
	assertParity(t, `Declare total to be 0.
Declare nums to be [1, 2, 3, 4, 5].
for each n in nums, do the following:
    Declare m to be the remainder of n divided by 3.
    If m is equal to 0, then
        Continue.
    thats it.
    Set total to be total + n.
thats it.
Print total.`)
}

// ─── Functions ───────────────────────────────────────────────────────────────

func TestParityFunctionDeclaration(t *testing.T) {
	assertParity(t, `To greet (name):
    Print "Hello, ".
    Print name.
done.
greet("World").`)
}

func TestParityFunctionReturn(t *testing.T) {
	assertParity(t, `To add (a, b):
    Return a + b.
done.
Declare result to be add(3, 4).
Print result.`)
}

func TestParityRecursiveFunction(t *testing.T) {
	assertParity(t, `To factorial (n):
    If n is less than or equal to 1, then
        Return 1.
    thats it.
    Return n * factorial(n - 1).
done.
Print factorial(5).`)
}

func TestParityFibonacci(t *testing.T) {
	assertParity(t, `To fib (n):
    If n is less than or equal to 1, then
        Return n.
    thats it.
    Return fib(n - 1) + fib(n - 2).
done.
Print fib(10).`)
}

// ─── Lists (Arrays) ──────────────────────────────────────────────────────────

func TestParityListAccess(t *testing.T) {
	assertParity(t, `Declare nums to be [10, 20, 30].
Print the item at position 0 in nums.`)
}

func TestParityListModification(t *testing.T) {
	assertParity(t, `Declare nums to be [1, 2, 3].
Set the item at position 1 in nums to be 99.
Print the item at position 1 in nums.`)
}

func TestParityListLength(t *testing.T) {
	assertParity(t, `Declare nums to be [1, 2, 3, 4, 5].
Print the length of nums.`)
}

// ─── Lookup Tables ───────────────────────────────────────────────────────────

func TestParityLookupTable(t *testing.T) {
	assertParity(t, `Declare t to be a lookup table.
Set t at "key" to be "value".
Print the entry "key" in t.`)
}

func TestParityLookupHas(t *testing.T) {
	assertParity(t, `Declare t to be a lookup table.
Set t at "a" to be 1.
Print (t has the key "a").
Print (t has the key "b").`)
}

func TestParityLookupDelete(t *testing.T) {
	assertParity(t, `Declare t to be a lookup table.
Set t at "x" to be 42.
Remove "x" from t.
Print (t has the key "x").`)
}

// ─── Nil / Nothing ───────────────────────────────────────────────────────────

func TestParityNilCheck(t *testing.T) {
	assertParity(t, `Declare x to be nothing.
If x has no value, then
    Print "nothing".
thats it.`)
}

func TestParityNilCheckSomething(t *testing.T) {
	assertParity(t, `Declare x to be 42.
If x has a value, then
    Print "something".
thats it.`)
}

// ─── Structs ─────────────────────────────────────────────────────────────────

func TestParityStructDeclaration(t *testing.T) {
	assertParity(t, `declare Point as a structure with the following fields:
    x is a number with 0 being the default.
    y is a number with 0 being the default.
thats it.
Declare p to be a new Point.
Print p's x.
Print p's y.`)
}

func TestParityStructFieldSet(t *testing.T) {
	assertParity(t, `declare Point as a structure with the following fields:
    x is a number with 0 being the default.
    y is a number with 0 being the default.
thats it.
Declare p to be a new Point.
Set p's x to be 5.
Set p's y to be 10.
Print p's x.
Print p's y.`)
}

func TestParityStructMethod(t *testing.T) {
	assertParity(t, `declare Counter as a structure with the following fields:
    value is a number with 0 being the default.
thats it.
To increment (c):
    Set c's value to be c's value + 1.
done.
Declare c to be a new Counter.
increment(c).
Print c's value.`)
}

// ─── Try / Catch ─────────────────────────────────────────────────────────────

func TestParityTryCatch(t *testing.T) {
	assertParity(t, `Try:
    Raise an error with the message "oops".
on error:
    Print "caught".
thats it.`)
}

func TestParityTryCatchError(t *testing.T) {
	assertParity(t, `Try:
    Raise an error with the message "bad".
on error:
    Print error.
thats it.`)
}

func TestParityTryFinally(t *testing.T) {
	assertParity(t, `Try:
    Print "try".
finally:
    Print "finally".
thats it.`)
}

func TestParityTryCatchFinally(t *testing.T) {
	assertParity(t, `Try:
    Raise an error with the message "err".
on error:
    Print "caught".
finally:
    Print "done".
thats it.`)
}

func TestParityErrorNotLeakingAcrossTryBlocks(t *testing.T) {
	assertParity(t, `Try:
    Raise an error with the message "first".
on error:
    Print error.
thats it.
Try:
    Raise an error with the message "second".
on error:
    Print error.
thats it.`)
}

// ─── Custom Error Types ───────────────────────────────────────────────────────

func TestParityCustomErrorType(t *testing.T) {
	assertParity(t, `Declare NetworkError as an error type.
Try:
    Raise a NetworkError with the message "connection refused".
on NetworkError:
    Print "network error".
thats it.`)
}

func TestParityErrorHierarchy(t *testing.T) {
	assertParity(t, `Declare AppError as an error type.
Declare DatabaseError as a type of AppError.
Try:
    Raise a DatabaseError with the message "query failed".
on AppError:
    Print "app error caught".
thats it.`)
}

// ─── Cast ────────────────────────────────────────────────────────────────────

func TestParityCastNumberToText(t *testing.T) {
	assertParity(t, `Declare n to be 42.
Declare s to be cast n to text.
Print s.`)
}

func TestParityCastTextToNumber(t *testing.T) {
	assertParity(t, `Declare s to be "3.14".
Declare n to be cast s to number.
Print n.`)
}

func TestParityCastNumberToBoolean(t *testing.T) {
	assertParity(t, `Declare n to be 1.
Declare b to be cast n to boolean.
Print b.`)
}

// ─── Standard Library ────────────────────────────────────────────────────────

func TestParityStdlibSqrt(t *testing.T) {
	assertParity(t, `Print sqrt(16).`)
}

func TestParityStdlibAbs(t *testing.T) {
	assertParity(t, `Print abs(-5).`)
}

func TestParityStdlibFloor(t *testing.T) {
	assertParity(t, `Print floor(3.7).`)
}

func TestParityStdlibCeil(t *testing.T) {
	assertParity(t, `Print ceil(3.2).`)
}

func TestParityStdlibRound(t *testing.T) {
	assertParity(t, `Print round(3.5).`)
}

func TestParityStdlibPow(t *testing.T) {
	assertParity(t, `Print pow(2, 10).`)
}

func TestParityStdlibMinMax(t *testing.T) {
	assertParity(t, `Print min(3, 7).
Print max(3, 7).`)
}

func TestParityStdlibUpperLower(t *testing.T) {
	assertParity(t, `Declare s to be "Hello World".
Print uppercase(s).
Print lowercase(s).`)
}

func TestParityStdlibTrim(t *testing.T) {
	assertParity(t, `Print trim("  hello  ").`)
}

func TestParityStdlibStartsEnds(t *testing.T) {
	assertParity(t, `Declare s to be "hello world".
Print starts_with(s, "hello").
Print ends_with(s, "world").`)
}

func TestParityStdlibLength(t *testing.T) {
	assertParity(t, `Print length("hello").`)
}

func TestParityStdlibSumCount(t *testing.T) {
	assertParity(t, `Declare nums to be [1, 2, 3, 4, 5].
Print sum(nums).
Print count(nums).`)
}

func TestParityStdlibFirstLast(t *testing.T) {
	assertParity(t, `Declare nums to be [10, 20, 30].
Print first(nums).
Print last(nums).`)
}

// ─── Predefined Constants ────────────────────────────────────────────────────

func TestParityPiConstant(t *testing.T) {
	assertParity(t, `Print pi.`)
}

func TestParityEConstant(t *testing.T) {
	assertParity(t, `Print e.`)
}

// ─── Swap ────────────────────────────────────────────────────────────────────

func TestParitySwap(t *testing.T) {
	assertParity(t, `Declare a to be 1.
Declare b to be 2.
swap a and b.
Print a.
Print b.`)
}

// ─── Regression Programs ─────────────────────────────────────────────────────

func TestParityFizzBuzz(t *testing.T) {
	assertParity(t, `Declare i to be 1.
repeat the following while i is less than or equal to 20:
    Declare m3 to be the remainder of i divided by 3.
    Declare m5 to be the remainder of i divided by 5.
    If (m3 is equal to 0 and m5 is equal to 0), then
        Print "FizzBuzz".
    Otherwise, if m3 is equal to 0, then
        Print "Fizz".
    Otherwise, if m5 is equal to 0, then
        Print "Buzz".
    Otherwise:
        Print i.
    thats it.
    Set i to be i + 1.
thats it.`)
}

func TestParityBubbleSort(t *testing.T) {
	assertParity(t, `Declare arr to be [64, 34, 25, 12, 22, 11, 90].
Declare n to be the length of arr.
Declare i to be 0.
repeat the following while i is less than n:
    Declare j to be 0.
    repeat the following while j is less than n - i - 1:
        Declare a to be the item at position j in arr.
        Declare b to be the item at position (j + 1) in arr.
        If a is greater than b, then
            Set the item at position j in arr to be b.
            Set the item at position (j + 1) in arr to be a.
        thats it.
        Set j to be j + 1.
    thats it.
    Set i to be i + 1.
thats it.
Declare k to be 0.
repeat the following while k is less than n:
    Print the item at position k in arr.
    Set k to be k + 1.
thats it.`)
}

func TestParityNestedFunctions(t *testing.T) {
	assertParity(t, `To double (x):
    Return x * 2.
done.
To quadruple (x):
    Return double(double(x)).
done.
Print quadruple(3).`)
}

// ─── Error Cases (both VMs should fail identically) ──────────────────────────

func TestParityRuntimeErrorDivByZero(t *testing.T) {
	// Both VMs should produce an error (exact message may differ but both fail)
	assertParityError(t, `Declare x to be 1 / 0.`)
}

func TestParityRuntimeErrorUndefinedVar(t *testing.T) {
	assertParityError(t, `Print undefined_variable.`)
}

func TestParityConstantReassignmentError(t *testing.T) {
	assertParityError(t, `Declare X as a constant to be 5.
Set X to be 10.`)
}

// ─── Print with write (no newline) ────────────────────────────────────────────

func TestParityWrite(t *testing.T) {
	assertParity(t, `Write "Hello".
Write " World".
Print ".".`)
}

// ─── Multiple Return Values (none in this language, but test complex return) ──

func TestParityEarlyReturn(t *testing.T) {
	assertParity(t, `To checkPositive (n):
    If n is greater than 0, then
        Return "positive".
    thats it.
    Return "non-positive".
done.
Print checkPositive(5).
Print checkPositive(-3).`)
}

// ─── Nested Loops ─────────────────────────────────────────────────────────────

func TestParityNestedLoops(t *testing.T) {
	assertParity(t, `Declare result to be 0.
Declare i to be 0.
repeat the following while i is less than 3:
    Declare j to be 0.
    repeat the following while j is less than 3:
        Set result to be result + 1.
        Set j to be j + 1.
    thats it.
    Set i to be i + 1.
thats it.
Print result.`)
}

// ─── String Operations ────────────────────────────────────────────────────────

func TestParityStringConcat(t *testing.T) {
	assertParity(t, `Declare a to be "Hello".
Declare b to be " World".
Declare c to be a + b.
Print c.`)
}

func TestParityStringLength(t *testing.T) {
	assertParity(t, `Declare s to be "abcde".
Print the length of s.`)
}

// ─── Boolean Expressions ─────────────────────────────────────────────────────

func TestParityComplexBooleans(t *testing.T) {
	assertParity(t, `Declare a to be true.
Declare b to be true.
Declare c to be false.
Print ((a and b) or c).
Print ((a or c) and (b or c)).`)
}

// ─── Nested Try/Catch ─────────────────────────────────────────────────────────

func TestParityNestedTryCatch(t *testing.T) {
	assertParity(t, `Try:
    Try:
        Raise an error with the message "inner".
    on error:
        Print "inner caught".
    thats it.
    Print "outer try".
on error:
    Print "outer caught".
thats it.`)
}

// assertOutputContains is a convenience for when exact output is too strict.
// It still compares astvm vs ivm stdout for equality.
func assertOutputContains(t *testing.T, src string, want string) {
	t.Helper()
	astOut, astErr := runAST(src)
	ivmOut, ivmErr := runIVM(src)

	if astErr != nil {
		t.Errorf("astvm error: %v", astErr)
	}
	if ivmErr != nil {
		t.Errorf("ivm error: %v", ivmErr)
	}
	if astOut != ivmOut {
		t.Errorf("output parity mismatch:\n  astvm: %q\n  ivm:   %q", astOut, ivmOut)
	}
	if !strings.Contains(astOut, want) {
		t.Errorf("output %q does not contain %q", astOut, want)
	}
}

func TestParityPredefinedConstants(t *testing.T) {
	// Just assert parity — exact float formatting must match
	assertParity(t, `Print pi.
Print e.`)
}
