package repl_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/Advik-B/english/repl"
)

// ── Helpers ──────────────────────────────────────────────────────────────────

// runLoop feeds input to a fresh REPL (no banner) and returns everything
// written to the output writer.
func runLoop(input string) string {
	in := strings.NewReader(input)
	var out bytes.Buffer
	r := repl.New(in, &out, false)
	r.Loop()
	return out.String()
}

// runWithBanner feeds input to a fresh REPL (with banner) and returns the
// full output including the banner.
func runWithBanner(input string) string {
	in := strings.NewReader(input)
	var out bytes.Buffer
	r := repl.New(in, &out, false)
	r.Run()
	return out.String()
}

// assertContains fails the test if output does not contain each of the
// expected substrings.
func assertContains(t *testing.T, output string, wants ...string) {
	t.Helper()
	for _, want := range wants {
		if !strings.Contains(output, want) {
			t.Errorf("expected %q in output\nfull output:\n%s", want, output)
		}
	}
}

// assertNotContains fails the test if output contains any of the unexpected
// substrings.
func assertNotContains(t *testing.T, output string, bads ...string) {
	t.Helper()
	for _, bad := range bads {
		if strings.Contains(output, bad) {
			t.Errorf("did not expect %q in output\nfull output:\n%s", bad, output)
		}
	}
}

// ── Prompt tests ─────────────────────────────────────────────────────────────

func TestPrimaryPromptShown(t *testing.T) {
	out := runLoop("Print \"hi\".\n")
	assertContains(t, out, repl.PrimaryPrompt)
}

func TestContinuationPromptShownInBlock(t *testing.T) {
	input := join(
		"For each n in [1], do the following:",
		"    Print the value of n.",
		"thats it.",
	)
	out := runLoop(input)
	assertContains(t, out, repl.ContinuationPrompt)
}

// ── Banner ───────────────────────────────────────────────────────────────────

func TestBannerContainsEnglish(t *testing.T) {
	out := runWithBanner("exit\n")
	assertContains(t, out, "English")
}

func TestBannerContainsVersion(t *testing.T) {
	out := runWithBanner("exit\n")
	assertContains(t, out, repl.Version)
}

// ── Exit / quit ───────────────────────────────────────────────────────────────

func TestExitTerminatesLoop(t *testing.T) {
	// If "exit" didn't terminate, we'd block waiting for more input.
	out := runLoop("exit\n")
	// After exit the REPL should not print another prompt.
	// Just verifying it does not hang is enough.
	_ = out
}

func TestQuitTerminatesLoop(t *testing.T) {
	out := runLoop("quit\n")
	_ = out
}

func TestExitWithPeriod(t *testing.T) {
	out := runLoop("exit.\n")
	_ = out
}

// ── Help ──────────────────────────────────────────────────────────────────────

func TestHelpCommandShowsHelp(t *testing.T) {
	out := runLoop("help\n")
	assertContains(t, out, "exit")
	assertContains(t, out, "help")
}

// ── Simple statements ────────────────────────────────────────────────────────

func TestPrintStringLiteral(t *testing.T) {
	out := runLoop("Print \"Hello, World!\".\n")
	assertContains(t, out, "Hello, World!")
}

func TestPrintNumber(t *testing.T) {
	out := runLoop("Print 42.\n")
	assertContains(t, out, "42")
}

func TestPrintBoolean(t *testing.T) {
	out := runLoop("Print true.\n")
	assertContains(t, out, "true")
}

// ── Variable declaration and persistence ─────────────────────────────────────

func TestVariableDeclarationAndPrint(t *testing.T) {
	out := runLoop(join(
		"Declare x to be 42.",
		"Print the value of x.",
	))
	assertContains(t, out, "42")
}

func TestVariablesPersistedAcrossStatements(t *testing.T) {
	out := runLoop(join(
		"Declare x to be 10.",
		"Declare y to be 20.",
		"Print x + y.",
	))
	assertContains(t, out, "30")
}

func TestConstantDeclaration(t *testing.T) {
	out := runLoop(join(
		"Declare pi to always be 3.",
		"Print the value of pi.",
	))
	assertContains(t, out, "3")
}

func TestAssignment(t *testing.T) {
	out := runLoop(join(
		"Declare x to be 1.",
		"Set x to be 99.",
		"Print the value of x.",
	))
	assertContains(t, out, "99")
}

// ── Arithmetic ───────────────────────────────────────────────────────────────

func TestArithmeticAddition(t *testing.T) {
	out := runLoop("Print 3 + 4.\n")
	assertContains(t, out, "7")
}

func TestArithmeticSubtraction(t *testing.T) {
	out := runLoop("Print 10 - 3.\n")
	assertContains(t, out, "7")
}

func TestArithmeticMultiplication(t *testing.T) {
	out := runLoop("Print 6 * 7.\n")
	assertContains(t, out, "42")
}

func TestArithmeticDivision(t *testing.T) {
	out := runLoop("Print 15 / 3.\n")
	assertContains(t, out, "5")
}

// ── Multiline: for loop ───────────────────────────────────────────────────────

func TestForEachLoop(t *testing.T) {
	out := runLoop(join(
		"For each n in [1, 2, 3], do the following:",
		"    Print the value of n.",
		"thats it.",
	))
	assertContains(t, out, "1", "2", "3")
}

func TestForEachLoopLowerCase(t *testing.T) {
	out := runLoop(join(
		"for each n in [4, 5], do the following:",
		"    Print the value of n.",
		"thats it.",
	))
	assertContains(t, out, "4", "5")
}

// ── Multiline: while loop ─────────────────────────────────────────────────────

func TestWhileLoop(t *testing.T) {
	out := runLoop(join(
		"Declare count to be 0.",
		"repeat the following while count is less than 3:",
		"    Set count to be count + 1.",
		"thats it.",
		"Print the value of count.",
	))
	assertContains(t, out, "3")
}

// ── Multiline: if / otherwise ────────────────────────────────────────────────

func TestIfThenBlock(t *testing.T) {
	out := runLoop(join(
		"Declare x to be 10.",
		"If x is greater than 5, then",
		"    Print \"big\".",
		"thats it.",
	))
	assertContains(t, out, "big")
}

func TestIfOtherwiseBlock(t *testing.T) {
	out := runLoop(join(
		"Declare x to be 2.",
		"If x is greater than 5, then",
		"    Print \"big\".",
		"otherwise",
		"    Print \"small\".",
		"thats it.",
	))
	assertContains(t, out, "small")
	assertNotContains(t, out, "big")
}

func TestIfOtherwiseIfBlock(t *testing.T) {
	out := runLoop(join(
		"Declare score to be 85.",
		"If score is greater than or equal to 90, then",
		"    Print \"A\".",
		"otherwise if score is greater than or equal to 80, then",
		"    Print \"B\".",
		"otherwise",
		"    Print \"C\".",
		"thats it.",
	))
	assertContains(t, out, "B")
	assertNotContains(t, out, "A")
	assertNotContains(t, out, "C")
}

// ── Multiline: function declaration ──────────────────────────────────────────

func TestFunctionDeclarationAndCall(t *testing.T) {
	out := runLoop(join(
		"Declare function double that takes n and does the following:",
		"    Return n * 2.",
		"thats it.",
		"Declare result to be 0.",
		"Set result to be the result of calling double with 5.",
		"Print the value of result.",
	))
	assertContains(t, out, "10")
}

func TestFunctionWithMultipleParams(t *testing.T) {
	out := runLoop(join(
		"Declare function add that takes a and b and does the following:",
		"    Return a + b.",
		"thats it.",
		"Declare s to be 0.",
		"Set s to be the result of calling add with 3 and 4.",
		"Print the value of s.",
	))
	assertContains(t, out, "7")
}

func TestFunctionPrintOutput(t *testing.T) {
	out := runLoop(join(
		"Declare function greet that takes name and does the following:",
		"    Print \"Hello,\", the value of name.",
		"thats it.",
		"Declare dummy to be 0.",
		"Set dummy to be the result of calling greet with \"World\".",
	))
	assertContains(t, out, "Hello,")
	assertContains(t, out, "World")
}

// ── Nested blocks ────────────────────────────────────────────────────────────

func TestNestedIfInsideLoop(t *testing.T) {
	out := runLoop(join(
		"Declare i to be 1.",
		"repeat the following while i is less than or equal to 3:",
		"    If i is equal to 2, then",
		"        Print \"two\".",
		"    otherwise",
		"        Print \"other\".",
		"    thats it.",
		"    Set i to be i + 1.",
		"thats it.",
	))
	assertContains(t, out, "two")
	assertContains(t, out, "other")
}

// ── Try/catch/finally ─────────────────────────────────────────────────────────

func TestTryCatch(t *testing.T) {
	out := runLoop(join(
		"Try doing the following:",
		"    Raise \"oops\".",
		"on error:",
		"    Print \"caught\".",
		"thats it.",
	))
	assertContains(t, out, "caught")
}

func TestTryCatchFinally(t *testing.T) {
	out := runLoop(join(
		"Try doing the following:",
		"    Raise \"oops\".",
		"on error:",
		"    Print \"caught\".",
		"but finally:",
		"    Print \"done\".",
		"thats it.",
	))
	assertContains(t, out, "caught")
	assertContains(t, out, "done")
}

// ── Error handling ────────────────────────────────────────────────────────────

func TestSyntaxErrorReported(t *testing.T) {
	out := runLoop("this is not valid !!!!\n")
	assertContains(t, out, "Error")
}

func TestRuntimeErrorReported(t *testing.T) {
	out := runLoop(join(
		"Declare x to be 0.",
		"Set x to be 10 / 0.",
	))
	// Division by zero should produce a runtime error.
	assertContains(t, out, "Error")
}

func TestReplContinuesAfterError(t *testing.T) {
	// The REPL should keep running after an error.
	out := runLoop(join(
		"this is not valid !!!!", // bad line – should produce an error
		"Print \"still running\".",
	))
	assertContains(t, out, "Error")
	assertContains(t, out, "still running")
}

// ── Full programs ─────────────────────────────────────────────────────────────

func TestFizzBuzz(t *testing.T) {
	out := runLoop(join(
		"Declare i to be 1.",
		"Declare mod3 to be 0.",
		"Declare mod5 to be 0.",
		"repeat the following while i is less than or equal to 15:",
		"    Set mod3 to be the remainder of i divided by 3.",
		"    Set mod5 to be the remainder of i divided by 5.",
		"    If mod3 is equal to 0, then",
		"        If mod5 is equal to 0, then",
		"            Print \"FizzBuzz\".",
		"        otherwise",
		"            Print \"Fizz\".",
		"        thats it.",
		"    otherwise if mod5 is equal to 0, then",
		"        Print \"Buzz\".",
		"    otherwise",
		"        Print the value of i.",
		"    thats it.",
		"    Set i to be i + 1.",
		"thats it.",
	))
	assertContains(t, out, "Fizz")
	assertContains(t, out, "Buzz")
	assertContains(t, out, "FizzBuzz")
}

func TestFibonacci(t *testing.T) {
	out := runLoop(join(
		"Declare a to be 0.",
		"Declare b to be 1.",
		"Declare temp to be 0.",
		"Declare count to be 0.",
		"repeat the following while count is less than 7:",
		"    Print the value of a.",
		"    Set temp to be a + b.",
		"    Set a to be b.",
		"    Set b to be temp.",
		"    Set count to be count + 1.",
		"thats it.",
	))
	// First 7 Fibonacci numbers: 0 1 1 2 3 5 8
	for _, want := range []string{"0", "1", "2", "3", "5", "8"} {
		assertContains(t, out, want)
	}
}

func TestFactorial(t *testing.T) {
	out := runLoop(join(
		"Declare function factorial that takes n and does the following:",
		"    If n is less than or equal to 1, then",
		"        Return 1.",
		"    thats it.",
		"    Declare smaller to be 0.",
		"    Set smaller to be the result of calling factorial with n - 1.",
		"    Return n * smaller.",
		"thats it.",
		"Declare result to be 0.",
		"Set result to be the result of calling factorial with 5.",
		"Print the value of result.",
	))
	assertContains(t, out, "120")
}

func TestArrayOperations(t *testing.T) {
	out := runLoop(join(
		"Declare numbers to be [10, 20, 30].",
		"Print the item at position 0 of numbers.",
		"Print count(numbers).",
	))
	assertContains(t, out, "10")
	assertContains(t, out, "3")
}

func TestStringConcatenation(t *testing.T) {
	out := runLoop(join(
		"Declare greeting to be \"Hello\".",
		"Declare name to be \"World\".",
		"Print greeting + \", \" + name + \"!\".",
	))
	assertContains(t, out, "Hello, World!")
}

func TestBooleanLogic(t *testing.T) {
	out := runLoop(join(
		"Declare a to be true.",
		"Declare b to be false.",
		"If a and b, then",
		"    Print \"and is true\".",
		"otherwise",
		"    Print \"and is false\".",
		"thats it.",
		"If a or b, then",
		"    Print \"or is true\".",
		"otherwise",
		"    Print \"or is false\".",
		"thats it.",
	))
	assertContains(t, out, "and is false")
	assertContains(t, out, "or is true")
}

// ── Blank lines inside a block ────────────────────────────────────────────────

func TestBlankLinesInsideBlock(t *testing.T) {
	// Blank lines in a block should be tolerated.
	out := runLoop(join(
		"For each n in [7], do the following:",
		"",
		"    Print the value of n.",
		"",
		"thats it.",
	))
	assertContains(t, out, "7")
}

// ── Multiline function with nested if ────────────────────────────────────────

func TestFunctionWithNestedIf(t *testing.T) {
	out := runLoop(join(
		"Declare function classify that takes n and does the following:",
		"    If n is greater than 0, then",
		"        Print \"positive\".",
		"    otherwise if n is less than 0, then",
		"        Print \"negative\".",
		"    otherwise",
		"        Print \"zero\".",
		"    thats it.",
		"thats it.",
		"Declare dummy to be 0.",
		"Set dummy to be the result of calling classify with 5.",
		"Set dummy to be the result of calling classify with 0.",
		"Set dummy to be the result of calling classify with -3.",
	))
	assertContains(t, out, "positive")
	assertContains(t, out, "zero")
	assertContains(t, out, "negative")
}

// ── join helper ───────────────────────────────────────────────────────────────

// join concatenates the given lines with newlines appended, producing an input
// string that simulates the user pressing Enter after each line.
func join(lines ...string) string {
	return strings.Join(lines, "\n") + "\n"
}
