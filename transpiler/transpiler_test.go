package transpiler_test

import (
	"english/parser"
	"english/transpiler"
	"strings"
	"testing"
)

// ─── helpers ─────────────────────────────────────────────────────────────────

// transpile parses the given English source and transpiles it to Python.
// The returned string has leading/trailing whitespace stripped.
func transpile(t *testing.T, src string) string {
	t.Helper()
	lexer := parser.NewLexer(src)
	tokens := lexer.TokenizeAll()
	p := parser.NewParser(tokens)
	prog, err := p.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	result := transpiler.NewTranspiler().Transpile(prog)
	return strings.TrimSpace(result)
}

// containsLine returns true when the output contains a line that equals s
// (after stripping leading/trailing whitespace from each line).
func containsLine(output, s string) bool {
	for _, line := range strings.Split(output, "\n") {
		if strings.TrimSpace(line) == s {
			return true
		}
	}
	return false
}

// assertContains fails the test if needle is not found in haystack.
func assertContains(t *testing.T, haystack, needle string) {
	t.Helper()
	if !strings.Contains(haystack, needle) {
		t.Errorf("expected output to contain %q\ngot:\n%s", needle, haystack)
	}
}

// assertContainsLine fails if output does not contain an exact line equal to s.
func assertContainsLine(t *testing.T, output, s string) {
	t.Helper()
	if !containsLine(output, s) {
		t.Errorf("expected output to contain line %q\ngot:\n%s", s, output)
	}
}

// ─── Basic literals ───────────────────────────────────────────────────────────

func TestHelloWorld(t *testing.T) {
	out := transpile(t, `Print "Hello, World!".`)
	assertContainsLine(t, out, `print("Hello, World!")`)
}

func TestWriteNoNewline(t *testing.T) {
	out := transpile(t, `Write "hi".`)
	assertContainsLine(t, out, `print("hi", end="")`)
}

// ─── Variables ────────────────────────────────────────────────────────────────

func TestVariableDecl(t *testing.T) {
	out := transpile(t, `Declare x to be 42.`)
	assertContainsLine(t, out, `x = 42`)
}

func TestConstantDecl(t *testing.T) {
	out := transpile(t, `Declare PI to always be 3.14.`)
	assertContains(t, out, `from typing import Final`)
	assertContains(t, out, `PI: Final = 3.14`)
}

func TestTypedVariableDecl(t *testing.T) {
	out := transpile(t, `Declare n as number to be 5.`)
	assertContainsLine(t, out, `n: float = 5`)
}

func TestTypedConstant(t *testing.T) {
	out := transpile(t, `Declare n as number to always be 5.`)
	assertContains(t, out, `from typing import Final`)
	assertContains(t, out, `n: Final[float] = 5`)
}

func TestAssignment(t *testing.T) {
	out := transpile(t, `Declare x to be 0.
Set x to be 10.`)
	assertContainsLine(t, out, `x = 10`)
}

// ─── Arithmetic operators ─────────────────────────────────────────────────────

func TestAddition(t *testing.T) {
	out := transpile(t, `Declare x to be 3 + 4.`)
	assertContainsLine(t, out, `x = 3 + 4`)
}

func TestModulo(t *testing.T) {
	out := transpile(t, `Declare r to be the remainder of 10 divided by 3.`)
	assertContainsLine(t, out, `r = 10 % 3`)
}

// ─── Control flow ─────────────────────────────────────────────────────────────

func TestIfElse(t *testing.T) {
	out := transpile(t, `Declare x to be 5.
If x is greater than 3, then
    Print "big".
otherwise
    Print "small".
thats it.`)
	assertContains(t, out, "if x > 3:")
	assertContains(t, out, `print("big")`)
	assertContains(t, out, "else:")
	assertContains(t, out, `print("small")`)
}

func TestWhileLoop(t *testing.T) {
	out := transpile(t, `Declare i to be 0.
repeat the following while i is less than 5:
    Set i to be i + 1.
thats it.`)
	assertContains(t, out, "while i < 5:")
}

func TestForLoopIntegerLiteral(t *testing.T) {
	// Integer literals should NOT be wrapped in int().
	out := transpile(t, `repeat the following 3 times:
    Print "hi".
thats it.`)
	assertContains(t, out, "for _ in range(3):")
	if strings.Contains(out, "int(3)") {
		t.Error("expected no int() wrapping for integer literal 3")
	}
}

func TestForEachLoop(t *testing.T) {
	out := transpile(t, `Declare nums to be [1, 2, 3].
for each n in nums, do the following:
    Print the value of n.
thats it.`)
	assertContains(t, out, "for n in nums:")
}

func TestBreakContinue(t *testing.T) {
	out := transpile(t, `Declare i to be 0.
repeat forever:
    Set i to be i + 1.
    If i is equal to 3, then
        break out of this loop.
    thats it.
thats it.`)
	assertContains(t, out, "while True:")
	assertContains(t, out, "break")
}

// ─── Functions ────────────────────────────────────────────────────────────────

func TestFunctionDecl(t *testing.T) {
	out := transpile(t, `Declare function add that takes a and b and does the following:
    Return a + b.
thats it.`)
	assertContains(t, out, "def add(a, b):")
	assertContains(t, out, "return a + b")
}

func TestFunctionCall(t *testing.T) {
	out := transpile(t, `Declare function greet that takes name and does the following:
    Print "Hello", the value of name.
thats it.
Call greet with "Alice".`)
	assertContains(t, out, `greet("Alice")`)
}

// ─── Booleans / nil ──────────────────────────────────────────────────────────

func TestBooleanLiterals(t *testing.T) {
	out := transpile(t, `Declare a to be true.
Declare b to be false.`)
	assertContainsLine(t, out, "a = True")
	assertContainsLine(t, out, "b = False")
}

func TestNothingLiteral(t *testing.T) {
	out := transpile(t, `Declare x to be nothing.`)
	assertContainsLine(t, out, "x = None")
}

func TestToggle(t *testing.T) {
	out := transpile(t, `Declare flag to be true.
Toggle flag.`)
	assertContainsLine(t, out, "flag = not flag")
}

func TestNilCheck(t *testing.T) {
	out := transpile(t, `Declare x to be nothing.
If x is nothing, then
    Print "null".
thats it.`)
	assertContains(t, out, "x is None")
}

// ─── Error handling ───────────────────────────────────────────────────────────

func TestTryCatch(t *testing.T) {
	out := transpile(t, `Try doing the following:
    Raise "oops".
on error:
    Print "caught", error.
thats it.`)
	assertContains(t, out, "try:")
	// Parser uses "RuntimeError" as the default error type when none is specified.
	assertContains(t, out, "raise RuntimeError(\"oops\")")
	assertContains(t, out, "except Exception as error:")
}

func TestCustomErrorType(t *testing.T) {
	out := transpile(t, `Declare MyError as an error type.`)
	assertContainsLine(t, out, "class MyError(Exception): pass")
}

func TestErrorSubtype(t *testing.T) {
	out := transpile(t, `Declare NetworkError as an error type.
Declare TimeoutError as a type of NetworkError.`)
	assertContainsLine(t, out, "class TimeoutError(NetworkError): pass")
}

// ─── Swap ─────────────────────────────────────────────────────────────────────

func TestSwap(t *testing.T) {
	out := transpile(t, `Declare a to be 1.
Declare b to be 2.
Swap a and b.`)
	assertContainsLine(t, out, "a, b = b, a")
}

// ─── Lists and arrays ─────────────────────────────────────────────────────────

func TestListLiteral(t *testing.T) {
	out := transpile(t, `Declare nums to be [1, 2, 3].`)
	assertContainsLine(t, out, "nums = [1, 2, 3]")
}

func TestIndexAccess(t *testing.T) {
	// Integer-literal indices must not be wrapped in int().
	out := transpile(t, `Declare nums to be [10, 20, 30].
Print the item at position 0 in nums.`)
	assertContains(t, out, "nums[0]")
	if strings.Contains(out, "int(0)") {
		t.Error("expected no int() wrapping for index literal 0")
	}
}

func TestIndexAssignment(t *testing.T) {
	out := transpile(t, `Declare nums to be [10, 20, 30].
Set the item at position 1 in nums to be 99.`)
	assertContains(t, out, "nums[1] = 99")
}

func TestLength(t *testing.T) {
	out := transpile(t, `Declare nums to be [1, 2, 3].
Print the number of nums.`)
	assertContains(t, out, "len(nums)")
}

// ─── Lookup tables ────────────────────────────────────────────────────────────

func TestLookupTable(t *testing.T) {
	out := transpile(t, `Declare ages to be a lookup table.
Set ages at "Alice" to be 30.`)
	assertContainsLine(t, out, `ages = {}`)
	assertContainsLine(t, out, `ages["Alice"] = 30`)
}

func TestLookupHas(t *testing.T) {
	out := transpile(t, `Declare t to be a lookup table.
If t has "key", then
    Print "yes".
thats it.`)
	assertContains(t, out, `"key" in t`)
}

// ─── Structs ─────────────────────────────────────────────────────────────────

func TestStructDecl(t *testing.T) {
	out := transpile(t, `declare Point as a structure with the following fields:
    x is a number with 0 being the default.
    y is a number with 0 being the default.
thats it.`)
	assertContains(t, out, "class Point:")
	assertContains(t, out, "def __init__(self, x=0, y=0):")
	assertContains(t, out, "self.x = x")
	assertContains(t, out, "self.y = y")
}

func TestStructMethod(t *testing.T) {
	out := transpile(t, `declare Counter as a structure with the following fields:
    count is a number with 0 being the default.

    let increment be a function that does the following:
        Set count to be count + 1.
    thats it.
thats it.`)
	assertContains(t, out, "def increment(self):")
	// 'count' inside the method must become self.count
	assertContains(t, out, "self.count = self.count + 1")
}

// ─── Cast expressions ────────────────────────────────────────────────────────

func TestCastToNumber(t *testing.T) {
	out := transpile(t, `Declare x to be "42" cast to number.`)
	assertContains(t, out, `float("42")`)
}

func TestCastToText(t *testing.T) {
	out := transpile(t, `Declare x to be 42 cast to text.`)
	assertContains(t, out, `str(42)`)
}

func TestCastToBool(t *testing.T) {
	out := transpile(t, `Declare x to be 1 cast to boolean.`)
	assertContains(t, out, `bool(1)`)
}

// ─── stdlib – Math ────────────────────────────────────────────────────────────

func TestStdlibSqrt(t *testing.T) {
	out := transpile(t, `Print sqrt(9).`)
	assertContains(t, out, "import math")
	assertContains(t, out, "math.sqrt(9)")
}

func TestStdlibClamp(t *testing.T) {
	out := transpile(t, `Print clamp(10, 0, 5).`)
	assertContains(t, out, "max(0, min(5, 10))")
}

func TestStdlibSign(t *testing.T) {
	out := transpile(t, `Print sign(-3).`)
	assertContains(t, out, "_sign(-3)")
	assertContains(t, out, "def _sign")
}

func TestStdlibIsInteger(t *testing.T) {
	out := transpile(t, `Print is_integer(4.0).`)
	assertContains(t, out, "float(4).is_integer()")
}

func TestStdlibRandom(t *testing.T) {
	out := transpile(t, `Print random().`)
	assertContains(t, out, "import random")
	assertContains(t, out, "random.random()")
}

func TestStdlibRandomBetween(t *testing.T) {
	out := transpile(t, `Print random_between(1, 10).`)
	assertContains(t, out, "random.uniform(1, 10)")
}

// ─── stdlib – String ──────────────────────────────────────────────────────────

func TestStdlibUppercase(t *testing.T) {
	out := transpile(t, `Print uppercase("hello").`)
	assertContains(t, out, `"hello".upper()`)
}

func TestStdlibLowercase(t *testing.T) {
	out := transpile(t, `Print lowercase("HELLO").`)
	assertContains(t, out, `"HELLO".lower()`)
}

func TestStdlibTitle(t *testing.T) {
	out := transpile(t, `Print title("hello world").`)
	assertContains(t, out, `"hello world".title()`)
}

func TestStdlibCapitalize(t *testing.T) {
	out := transpile(t, `Print capitalize("hello").`)
	assertContains(t, out, `"hello".capitalize()`)
}

func TestStdlibSwapcase(t *testing.T) {
	out := transpile(t, `Print swapcase("Hello").`)
	assertContains(t, out, `"Hello".swapcase()`)
}

func TestStdlibTrim(t *testing.T) {
	out := transpile(t, `Print trim("  hi  ").`)
	assertContains(t, out, `"  hi  ".strip()`)
}

func TestStdlibTrimLeft(t *testing.T) {
	out := transpile(t, `Print trim_left("  hi").`)
	assertContains(t, out, `"  hi".lstrip()`)
}

func TestStdlibTrimRight(t *testing.T) {
	out := transpile(t, `Print trim_right("hi  ").`)
	assertContains(t, out, `"hi  ".rstrip()`)
}

func TestStdlibSplit(t *testing.T) {
	out := transpile(t, `Print split("a,b,c", ",").`)
	assertContains(t, out, `"a,b,c".split(",")`)
}

func TestStdlibJoin(t *testing.T) {
	out := transpile(t, `Declare parts to be ["a", "b", "c"].
Print join(parts, "-").`)
	assertContains(t, out, `"-".join(parts)`)
}

func TestStdlibContains(t *testing.T) {
	out := transpile(t, `Print contains("hello world", "world").`)
	assertContains(t, out, `("world" in "hello world")`)
}

func TestStdlibStartsWith(t *testing.T) {
	out := transpile(t, `Print starts_with("hello", "he").`)
	assertContains(t, out, `"hello".startswith("he")`)
}

func TestStdlibEndsWith(t *testing.T) {
	out := transpile(t, `Print ends_with("hello", "lo").`)
	assertContains(t, out, `"hello".endswith("lo")`)
}

func TestStdlibSubstring(t *testing.T) {
	out := transpile(t, `Print substring("hello world", 6, 5).`)
	// No int() wrapping on integer literals.
	assertContains(t, out, `"hello world"[6:6+5]`)
}

func TestStdlibStrRepeat(t *testing.T) {
	out := transpile(t, `Print str_repeat("ha", 3).`)
	assertContains(t, out, `"ha" * 3`)
}

func TestStdlibReplace(t *testing.T) {
	out := transpile(t, `Print replace("hello", "l", "r").`)
	assertContains(t, out, `"hello".replace("l", "r")`)
}

func TestStdlibCenter(t *testing.T) {
	out := transpile(t, `Print center("hi", 10, "*").`)
	assertContains(t, out, `"hi".center(10, "*")`)
}

func TestStdlibZfill(t *testing.T) {
	out := transpile(t, `Print zfill("42", 6).`)
	assertContains(t, out, `"42".zfill(6)`)
}

func TestStdlibIsDigit(t *testing.T) {
	out := transpile(t, `Print is_digit("123").`)
	assertContains(t, out, `"123".isdigit()`)
}

func TestStdlibIsAlpha(t *testing.T) {
	out := transpile(t, `Print is_alpha("abc").`)
	assertContains(t, out, `"abc".isalpha()`)
}

func TestStdlibIsUpper(t *testing.T) {
	out := transpile(t, `Print is_upper("ABC").`)
	assertContains(t, out, `"ABC".isupper()`)
}

func TestStdlibIsLower(t *testing.T) {
	out := transpile(t, `Print is_lower("abc").`)
	assertContains(t, out, `"abc".islower()`)
}

func TestStdlibPadLeft(t *testing.T) {
	out := transpile(t, `Print pad_left("42", 8, "0").`)
	assertContains(t, out, `"42".rjust(8, "0")`)
}

func TestStdlibPadRight(t *testing.T) {
	out := transpile(t, `Print pad_right("hi", 10, ".").`)
	assertContains(t, out, `"hi".ljust(10, ".")`)
}

func TestStdlibToNumber(t *testing.T) {
	out := transpile(t, `Print to_number("3.14").`)
	assertContains(t, out, `float("3.14")`)
}

func TestStdlibToString(t *testing.T) {
	out := transpile(t, `Print to_string(42).`)
	assertContains(t, out, `str(42)`)
}

func TestStdlibIsEmpty(t *testing.T) {
	out := transpile(t, `Print is_empty("").`)
	assertContains(t, out, `(len("") == 0)`)
}

// ─── stdlib – List ────────────────────────────────────────────────────────────

func TestStdlibCount(t *testing.T) {
	out := transpile(t, `Declare nums to be [1, 2, 3].
Print count(nums).`)
	assertContains(t, out, "len(nums)")
}

func TestStdlibSum(t *testing.T) {
	out := transpile(t, `Declare nums to be [1, 2, 3].
Print sum(nums).`)
	assertContains(t, out, "sum(nums)")
}

func TestStdlibProduct(t *testing.T) {
	out := transpile(t, `Declare nums to be [1, 2, 3].
Print product(nums).`)
	assertContains(t, out, "_product(nums)")
	assertContains(t, out, "def _product")
}

func TestStdlibAverage(t *testing.T) {
	out := transpile(t, `Declare nums to be [1, 2, 3].
Print average(nums).`)
	assertContains(t, out, "(sum(nums) / len(nums))")
}

func TestStdlibMinValue(t *testing.T) {
	out := transpile(t, `Declare nums to be [3, 1, 2].
Print min_value(nums).`)
	assertContains(t, out, "min(nums)")
}

func TestStdlibMaxValue(t *testing.T) {
	out := transpile(t, `Declare nums to be [3, 1, 2].
Print max_value(nums).`)
	assertContains(t, out, "max(nums)")
}

func TestStdlibSort(t *testing.T) {
	out := transpile(t, `Declare nums to be [3, 1, 2].
Print sort(nums).`)
	assertContains(t, out, "sorted(nums)")
}

func TestStdlibSortedDesc(t *testing.T) {
	out := transpile(t, `Declare nums to be [3, 1, 2].
Print sorted_desc(nums).`)
	assertContains(t, out, "sorted(nums, reverse=True)")
}

func TestStdlibReverse(t *testing.T) {
	out := transpile(t, `Declare nums to be [1, 2, 3].
Print reverse(nums).`)
	assertContains(t, out, "list(reversed(nums))")
}

func TestStdlibFirst(t *testing.T) {
	out := transpile(t, `Declare nums to be [1, 2, 3].
Print first(nums).`)
	assertContains(t, out, "nums[0]")
}

func TestStdlibLast(t *testing.T) {
	out := transpile(t, `Declare nums to be [1, 2, 3].
Print last(nums).`)
	assertContains(t, out, "nums[-1]")
}

func TestStdlibAppend(t *testing.T) {
	out := transpile(t, `Declare nums to be [1, 2].
Set nums to be append(nums, 3).`)
	assertContains(t, out, "nums + [3]")
}

func TestStdlibSlice(t *testing.T) {
	out := transpile(t, `Declare nums to be [1, 2, 3, 4, 5].
Print slice(nums, 1, 4).`)
	assertContains(t, out, "nums[1:4]")
}

func TestStdlibUnique(t *testing.T) {
	out := transpile(t, `Declare nums to be [1, 2, 2, 3].
Print unique(nums).`)
	assertContains(t, out, "_unique(nums)")
	assertContains(t, out, "def _unique")
}

func TestStdlibFlatten(t *testing.T) {
	out := transpile(t, `Declare matrix to be [[1, 2], [3, 4]].
Print flatten(matrix).`)
	assertContains(t, out, "_flatten(matrix)")
	assertContains(t, out, "def _flatten")
}

func TestStdlibAnyTrue(t *testing.T) {
	out := transpile(t, `Declare flags to be [true, false].
Print any_true(flags).`)
	assertContains(t, out, "any(flags)")
}

func TestStdlibAllTrue(t *testing.T) {
	out := transpile(t, `Declare flags to be [true, true].
Print all_true(flags).`)
	assertContains(t, out, "all(flags)")
}

func TestStdlibRemove(t *testing.T) {
	out := transpile(t, `Declare nums to be [1, 2, 3].
Set nums to be remove(nums, 1).`)
	assertContains(t, out, "for i, v in enumerate(nums)")
}

func TestStdlibInsert(t *testing.T) {
	out := transpile(t, `Declare nums to be [1, 3].
Set nums to be insert(nums, 1, 2).`)
	assertContains(t, out, "nums[:1] + [2] + nums[1:]")
}

func TestStdlibZipWith(t *testing.T) {
	out := transpile(t, `Declare a to be [1, 2].
Declare b to be [3, 4].
Print zip_with(a, b).`)
	assertContains(t, out, "_zip_with(a, b)")
	assertContains(t, out, "def _zip_with")
}

// ─── stdlib – Lookup table ───────────────────────────────────────────────────

func TestStdlibKeys(t *testing.T) {
	out := transpile(t, `Declare t to be a lookup table.
Print keys(t).`)
	assertContains(t, out, "list(t.keys())")
}

func TestStdlibValues(t *testing.T) {
	out := transpile(t, `Declare t to be a lookup table.
Print values(t).`)
	assertContains(t, out, "list(t.values())")
}

func TestStdlibTableRemove(t *testing.T) {
	out := transpile(t, `Declare t to be a lookup table.
Set t at "k" to be 1.
Set t to be table_remove(t, "k").`)
	assertContains(t, out, `_table_remove(t, "k")`)
	assertContains(t, out, "def _table_remove")
}

func TestStdlibTableHas(t *testing.T) {
	out := transpile(t, `Declare t to be a lookup table.
Print table_has(t, "k").`)
	assertContains(t, out, `("k" in t)`)
}

func TestStdlibMerge(t *testing.T) {
	out := transpile(t, `Declare a to be a lookup table.
Declare b to be a lookup table.
Set a at "x" to be 1.
Set b at "y" to be 2.
Declare c to be merge(a, b).`)
	assertContains(t, out, "{**a, **b}")
}

func TestStdlibGetOrDefault(t *testing.T) {
	out := transpile(t, `Declare t to be a lookup table.
Set t at "x" to be 1.
Print get_or_default(t, "missing", 0).`)
	assertContains(t, out, `t.get("missing", 0)`)
}

func TestMathConstants(t *testing.T) {
	out := transpile(t, `Print the value of pi.
Print the value of e.
Print the value of infinity.`)
	assertContains(t, out, "import math")
	assertContains(t, out, "math.pi")
	assertContains(t, out, "math.e")
	assertContains(t, out, "math.inf")
}

// ─── Python import gating ─────────────────────────────────────────────────────

func TestNoMathImportWhenNotNeeded(t *testing.T) {
	out := transpile(t, `Print "hello".`)
	if strings.Contains(out, "import math") {
		t.Error("should not emit 'import math' when no math functions are used")
	}
}

func TestNoCopyImportWhenNotNeeded(t *testing.T) {
	out := transpile(t, `Declare x to be 5.`)
	if strings.Contains(out, "import copy") {
		t.Error("should not emit 'import copy' when copy is not used")
	}
}

func TestNoTypingImportWhenNotNeeded(t *testing.T) {
	out := transpile(t, `Declare x to be 5.`)
	if strings.Contains(out, "typing") {
		t.Error("should not emit typing import for non-constant")
	}
}

// ─── Precedence / parenthesisation ───────────────────────────────────────────

func TestNestedBinaryParens(t *testing.T) {
	out := transpile(t, `Declare x to be 1 + 2 + 3.`)
	// The outer + should parenthesise its nested binary left operand.
	assertContains(t, out, "(1 + 2) + 3")
}

// ─── int() avoidance ─────────────────────────────────────────────────────────

func TestIntWrappingAvoidedForLiterals(t *testing.T) {
	// All integer literals in positions that used to unconditionally emit int()
	// should be left bare.
	cases := []struct {
		src  string
		want string
	}{
		{`repeat the following 5 times:
    Print "x".
thats it.`, "range(5)"},
		{`Declare nums to be [1, 2, 3].
Print slice(nums, 0, 2).`, "nums[0:2]"},
		{`Declare nums to be [1, 2, 3].
Set the item at position 0 in nums to be 9.`, "nums[0] = 9"},
	}
	for _, tc := range cases {
		out := transpile(t, tc.src)
		if !strings.Contains(out, tc.want) {
			t.Errorf("expected %q in output, got:\n%s", tc.want, out)
		}
		if strings.Contains(out, "int(0)") || strings.Contains(out, "int(2)") || strings.Contains(out, "int(5)") {
			t.Errorf("unexpected int() wrapping of literal in:\n%s", out)
		}
	}
}

// ─── Comment carry-over (.abc behaviour) ─────────────────────────────────────

// transpileStripped parses English source and transpiles using NewTranspilerStripped
// (the mode used for .101 bytecode files — no comments in output).
func transpileStripped(t *testing.T, src string) string {
t.Helper()
lexer := parser.NewLexer(src)
tokens := lexer.TokenizeAll()
p := parser.NewParser(tokens)
prog, err := p.Parse()
if err != nil {
t.Fatalf("parse error: %v", err)
}
result := transpiler.NewTranspilerStripped().Transpile(prog)
return strings.TrimSpace(result)
}

func TestCommentCarriedOver(t *testing.T) {
out := transpile(t, `# This is a comment
Print "hello".`)
// The banner should be present.
assertContains(t, out, "# Transpiled from English language source")
// The source comment should appear.
assertContains(t, out, "# This is a comment")
}

func TestEmptyCommentCarriedOver(t *testing.T) {
// A bare '#' with no text should produce a Python '#' line.
out := transpile(t, `#
Print "hello".`)
assertContains(t, out, "# Transpiled from English language source")
assertContainsLine(t, out, "#")
}

func TestCommentInsideFunction(t *testing.T) {
out := transpile(t, `Declare function greet that takes name and does the following:
    # say hello
    Print "Hello", the value of name.
thats it.`)
assertContains(t, out, "# say hello")
}

func TestImportCommentCarriedOver(t *testing.T) {
out := transpile(t, `Import "math".
Print "x".`)
// Import produces a Python comment when keepComments is true.
assertContains(t, out, `# import "math"`)
}

func TestMultipleComments(t *testing.T) {
out := transpile(t, `# First comment
# Second comment
Print "hi".`)
assertContains(t, out, "# First comment")
assertContains(t, out, "# Second comment")
}

// ─── Comment suppression (.101 / stripped mode) ───────────────────────────────

func TestStrippedModeNoBanner(t *testing.T) {
out := transpileStripped(t, `Print "hello".`)
if strings.Contains(out, "#") {
t.Errorf("stripped mode should produce no '#' lines, got:\n%s", out)
}
}

func TestStrippedModeNoSourceComments(t *testing.T) {
out := transpileStripped(t, `# This is a comment
Print "hello".`)
if strings.Contains(out, "#") {
t.Errorf("stripped mode should produce no '#' lines, got:\n%s", out)
}
// The actual code should still be emitted.
assertContains(t, out, `print("hello")`)
}

func TestStrippedModeNoImportComments(t *testing.T) {
out := transpileStripped(t, `Import "math".
Print "hi".`)
if strings.Contains(out, "#") {
t.Errorf("stripped mode should produce no '#' lines at all, got:\n%s", out)
}
}

func TestStrippedModeCodeStillCorrect(t *testing.T) {
// Stripping comments must not affect the generated code itself.
out := transpileStripped(t, `# compute something
Declare x to be 5 + 3.
Print the value of x.`)
assertContains(t, out, "x = 5 + 3")
assertContains(t, out, "print(x)")
if strings.Contains(out, "#") {
t.Errorf("stripped mode should produce no '#' lines, got:\n%s", out)
}
}

func TestStrippedModeNoConstantComment(t *testing.T) {
// Constants use typing.Final; the Final annotation itself is not a comment.
// There should be no '#' lines in stripped output.
out := transpileStripped(t, `Declare PI to always be 3.14.`)
assertContains(t, out, "PI: Final = 3.14")
if strings.Contains(out, "#") {
t.Errorf("stripped mode should produce no '#' lines, got:\n%s", out)
}
}
