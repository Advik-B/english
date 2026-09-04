package transpiler_test

import (
	"os"
	"strings"
	"testing"

	ast_pkg "github.com/Advik-B/english/ast"
	"github.com/Advik-B/english/parser"
	"github.com/Advik-B/english/transpiler"
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

// transpileInlined is like transpile but uses NewTranspilerInlined, which
// reads and inlines all imported .abc files into the single output (the
// behaviour activated by the --inline CLI flag).
func transpileInlined(t *testing.T, src string) string {
	t.Helper()
	lexer := parser.NewLexer(src)
	tokens := lexer.TokenizeAll()
	p := parser.NewParser(tokens)
	prog, err := p.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	result := transpiler.NewTranspilerInlined().Transpile(prog)
	return strings.TrimSpace(result)
}

// parse parses English source and returns the AST without transpiling.
func parse(t *testing.T, src string) *ast_pkg.Program {
	t.Helper()
	lexer := parser.NewLexer(src)
	tokens := lexer.TokenizeAll()
	p := parser.NewParser(tokens)
	prog, err := p.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	return prog
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

// assertContains fails the test if any needle is not found in haystack.
func assertContains(t *testing.T, haystack string, needles ...string) {
	t.Helper()
	for _, needle := range needles {
		if !strings.Contains(haystack, needle) {
			t.Errorf("expected output to contain %q\ngot:\n%s", needle, haystack)
		}
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

// TestModulo covers the remainder, which is not Python's %. English truncates
// both operands and takes the sign of the dividend; Python's % is floored and
// takes the sign of the divisor, so -7 % 3 was -1 in English and 2 in the
// generated Python.
func TestModulo(t *testing.T) {
	out := transpile(t, `Declare r to be the remainder of 10 divided by 3.`)
	assertContainsLine(t, out, `r = _remainder(10, 3)`)
	assertContains(t, out, "def _remainder", "import math")
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

// The range tests below used to expect a conditional expression that named
// start, end and step up to three times each, so any of them that did
// something as well as producing a value did it repeatedly.

func TestRangeLiteralProgrammerStyle(t *testing.T) {
	out := transpile(t, `Declare nums to be [1 .. 5].`)
	assertContains(t, out, "_range(1, 5)", "def _range")
}

func TestRangeLiteralNaturalEnglish(t *testing.T) {
	out := transpile(t, `Let myRange be a range from 10 to 15.`)
	assertContains(t, out, "_range(10, 15)")
}

func TestRangeLiteralDescending(t *testing.T) {
	out := transpile(t, `Declare countdown to be [10 .. 5].`)
	assertContains(t, out, "_range(10, 5)")
}

func TestRangeLiteralInLoop(t *testing.T) {
	out := transpile(t, `For each n in [1 .. 3], do the following:
    Print n.
thats it.`)
	assertContains(t, out, "for n in _range(1, 3):")
	assertContains(t, out, "print(_show(n))")
}

// TestRangeArgumentsAreEvaluatedOnce is the point of the helper: an argument
// that does something as well as producing a value must appear once.
func TestRangeArgumentsAreEvaluatedOnce(t *testing.T) {
	out := transpile(t, `Declare function next_bound that does the following:
    Return 5.
thats it.

Declare nums to be a range from 1 to the result of calling next_bound.`)
	var rangeLine string
	for _, line := range strings.Split(out, "\n") {
		if strings.Contains(line, "_range(") && !strings.HasPrefix(strings.TrimSpace(line), "def ") {
			rangeLine = line
		}
	}
	if rangeLine == "" {
		t.Fatalf("no range expression in the output:\n%s", out)
	}
	if got := strings.Count(rangeLine, "next_bound()"); got != 1 {
		t.Errorf("next_bound is called %d time(s) in %q, want 1", got, strings.TrimSpace(rangeLine))
	}
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

// TestCastToText covers "cast to text", which goes through English's renderer:
// Python's str() writes a whole number as 42.0, a boolean as True and the
// absence of a value as None.
func TestCastToText(t *testing.T) {
	out := transpile(t, `Declare x to be 42 cast to text.`)
	assertContains(t, out, `_show(42)`, "def _show")
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

// TestStdlibSubstring covers substring, whose inline form named the start
// twice — "s[start:start+length]" — and so evaluated it twice.
func TestStdlibSubstring(t *testing.T) {
	out := transpile(t, `Print substring("hello world", 6, 5).`)
	assertContains(t, out, `_substring("hello world", 6, 5)`, "def _substring")
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
	assertContains(t, out, `_show(42)`)
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

// TestStdlibAverage covers average, whose inline form named its argument
// twice — "(sum(x) / len(x))" — and so evaluated it twice.
func TestStdlibAverage(t *testing.T) {
	out := transpile(t, `Declare nums to be [1, 2, 3].
Print average(nums).`)
	assertContains(t, out, "_average(nums)", "def _average")
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
	assertContains(t, out, "_remove_at(nums, 1)", "def _remove_at")
}

// TestStdlibInsert covers insert, whose inline form named the list and the
// index twice each — "xs[:i] + [v] + xs[i:]" — and so evaluated them twice.
func TestStdlibInsert(t *testing.T) {
	out := transpile(t, `Declare nums to be [1, 3].
Set nums to be insert(nums, 1, 2).`)
	assertContains(t, out, "_insert(nums, 1, 2)", "def _insert")
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
	// In non-inline mode (default) an import emits a standard Python
	// "from module import *" statement, not a comment.
	out := transpile(t, `Import "math".
Print "x".`)
	assertContains(t, out, "from math import *")
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
	// Stripped mode must not emit any '#' comment lines.
	if strings.Contains(out, "#") {
		t.Errorf("stripped mode should produce no '#' lines at all, got:\n%s", out)
	}
	// The import should still be emitted as Python code.
	assertContains(t, out, "from math import *")
}

func TestStrippedModeCodeStillCorrect(t *testing.T) {
	// Stripping comments must not affect the generated code itself.
	out := transpileStripped(t, `# compute something
Declare x to be 5 + 3.
Print the value of x.`)
	assertContains(t, out, "x = 5 + 3")
	assertContains(t, out, "print(_show(x))")
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

// ─── Import inlining (--inline mode) ─────────────────────────────────────────

func TestImportInlining(t *testing.T) {
	// Write a small library file to a temp dir.
	dir := t.TempDir()
	libPath := dir + "/mylib.abc"
	libSrc := `Declare function double that takes n and does the following:
    Return n * 2.
thats it.
`
	if err := os.WriteFile(libPath, []byte(libSrc), 0644); err != nil {
		t.Fatalf("write lib: %v", err)
	}

	// Main file imports the library and calls the function.
	mainSrc := `Import "` + libPath + `".
Declare result to be 0.
Set result to the result of calling double with 5.
Print the value of result.`

	// Use inlined mode — the library code is merged into the single output.
	out := transpileInlined(t, mainSrc)
	assertContains(t, out, "def double(n)")
	assertContains(t, out, "result = double(5)")
	assertContains(t, out, "print(_show(result))")
}

func TestSelectiveImportInlining(t *testing.T) {
	dir := t.TempDir()
	libPath := dir + "/mathlib.abc"
	libSrc := `Declare function square that takes x and does the following:
    Return x * x.
thats it.

Declare function cube that takes x and does the following:
    Return x * x * x.
thats it.
`
	if err := os.WriteFile(libPath, []byte(libSrc), 0644); err != nil {
		t.Fatalf("write lib: %v", err)
	}

	mainSrc := `Import square from "` + libPath + `".
Print square(3).`

	// Use inlined mode — only square is merged in.
	out := transpileInlined(t, mainSrc)
	assertContains(t, out, "def square(x)")
	// cube should NOT be included since we only imported square.
	if strings.Contains(out, "def cube(") {
		t.Errorf("cube should not be inlined when only square is imported")
	}
}

// ─── Non-inline import (default mode) ────────────────────────────────────────

func TestNonInlineImportAll(t *testing.T) {
	// In non-inline mode, "Import" emits "from module import *".
	out := transpile(t, `Import "examples/math_library.abc".
Print "hello".`)
	assertContains(t, out, "from math_library import *")
	// The function definition must NOT be inlined.
	if strings.Contains(out, "def ") {
		t.Errorf("non-inline mode should not inline function definitions")
	}
}

func TestNonInlineSelectiveImport(t *testing.T) {
	// Selective import emits "from module import X, Y".
	out := transpile(t, `Import square and cube from "examples/math_library.abc".
Print square(3).`)
	assertContains(t, out, "from math_library import square, cube")
}

func TestNonInlineCrossDirectoryImport(t *testing.T) {
	// When the library is in a subdirectory relative to the main file, a
	// sys.path.insert line is emitted before the from-import.
	prog := parse(t, `Import "subdir/utils.abc".
Print "hello".`)
	result := strings.TrimSpace(
		transpiler.NewTranspiler().WithSourceDir(".").Transpile(prog))
	assertContains(t, result, `sys.path.insert(0, os.path.join(os.path.dirname(__file__), "subdir"))`)
	assertContains(t, result, "from utils import *")
}

func TestNonInlineSameDirectoryImport(t *testing.T) {
	// When library and main file are in the same directory, no sys.path line needed.
	prog := parse(t, `Import "examples/lib.abc".
Print "hi".`)
	result := strings.TrimSpace(
		transpiler.NewTranspiler().WithSourceDir("examples").Transpile(prog))
	assertContains(t, result, "from lib import *")
	if strings.Contains(result, "sys.path") {
		t.Errorf("same-directory import should not emit sys.path manipulation")
	}
}

func TestUserDefinedFunctionOverridesStdlib(t *testing.T) {
	// A user-defined function named "average" taking numbers should not be
	// mis-translated to the stdlib average(list) expression.
	out := transpile(t, `Declare function average that takes x and y and z and does the following:
    Return (x + y + z) / 3.
thats it.

Declare result to be 0.
Set result to the result of calling average with 10 and 20 and 30.
Print the value of result.`)
	assertContains(t, out, "def average(x, y, z)")
	assertContains(t, out, "result = average(10, 20, 30)")
}

func TestPythonKeywordEscaping(t *testing.T) {
	out := transpile(t, `Declare class to be "A".
Print the value of class.`)
	assertContains(t, out, `class_ = "A"`)
	assertContains(t, out, "print(_show(class_))")
}

func TestStructZeroValueDefaults(t *testing.T) {
	out := transpile(t, `declare Person as a structure with the following fields:
    name is a string.
    age is an integer.
thats it.

let p be a new instance of Person.
Print the name of p.`)
	// Both fields should have zero-value defaults so Person() works.
	assertContains(t, out, `name=""`)
	assertContains(t, out, "age=0")
}

func TestIndexOfReturnsNegOne(t *testing.T) {
	out := transpile(t, `Declare s to be "hello".
Print index_of(s, "xyz").`)
	// Should use .find() which returns -1, not .index() which raises ValueError.
	assertContains(t, out, ".find(")
	if strings.Contains(out, ".index(") {
		t.Errorf("index_of should emit .find(), not .index()")
	}
}

// ─── PEP 8 formatting (during code generation) ───────────────────────────────

func TestTwoBlankLinesBeforeDef(t *testing.T) {
	// A top-level function that follows regular code must be separated by
	// exactly two blank lines.
	out := transpile(t, `Print "hello".
Declare function foo that takes x and does the following:
    Return x.
thats it.`)
	// Two blank lines = three consecutive newlines between the print and def.
	assertContains(t, out, "print(\"hello\")\n\n\ndef foo(x)")
}

func TestTwoBlankLinesBetweenDefs(t *testing.T) {
	// Two top-level functions must be separated by exactly two blank lines.
	out := transpile(t, `Declare function foo that takes x and does the following:
    Return x.
thats it.

Declare function bar that takes y and does the following:
    Return y.
thats it.`)
	assertContains(t, out, "return x\n\n\ndef bar(y)")
}

func TestCommentAttachedToDef(t *testing.T) {
	// A comment immediately before a top-level function should stay adjacent
	// to the def — the two blank lines go before the comment, not between it
	// and the def.
	out := transpile(t, `Print "hi".
# My function
Declare function foo that takes x and does the following:
    Return x.
thats it.`)
	// Blank lines must appear before the comment, not between comment and def.
	assertContains(t, out, "print(\"hi\")\n\n\n# My function\ndef foo(x)")
}

// TestIndexIsWholeInPython covers indices, which English computes as numbers —
// it has one number type and it is a float — where Python demands an int.
//
// These tests used to assert the opposite, locking in a maybeInt that returned
// its argument unchanged behind a doc comment claiming no wrapping was needed.
// Python needs it: xs[1.0] is a TypeError, so any index that came from
// division, an average or a square root failed the moment it was transpiled.
func TestIndexIsWholeInPython(t *testing.T) {
	out := transpile(t, `Declare arr to be [1, 2, 3].
Declare half to be 4 / 2.
Print the item at position half in arr.`)
	assertContains(t, out, "arr[int(half)]")
}

func TestSliceBoundsAreWholeInPython(t *testing.T) {
	out := transpile(t, `Declare s to be "hello world".
Declare start to be 6 / 2.
Print substring(s, start, 5).`)
	assertContains(t, out, "int(start)")
}

// TestWholeNumbersAreNotWrappedTwice keeps the output readable: a literal, a
// length and an int() are already whole, so wrapping them again is noise.
func TestWholeNumbersAreNotWrappedTwice(t *testing.T) {
	out := transpile(t, `Declare arr to be [1, 2, 3].
Print the item at position 0 in arr.
Repeat the following the length of arr times:
    Print "x".
thats it.`)
	assertContains(t, out, "arr[0]")
	assertContains(t, out, "range(len(arr))")
	if strings.Contains(out, "int(0)") || strings.Contains(out, "int(len(") {
		t.Errorf("a value already known to be whole was wrapped again:\n%s", out)
	}
}

func TestNoBlankLinesAtStartOfFile(t *testing.T) {
	// A def at the very start of the file must not be preceded by blank lines.
	// Use raw (non-trimmed) output so that leading newlines are visible.
	prog := parse(t, `Declare function foo that does the following:
    Print "hi".
thats it.`)
	raw := transpiler.NewTranspiler().Transpile(prog)
	// Skip any banner/comment lines, then expect no blank line before the def.
	for _, line := range strings.Split(raw, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "#") {
			continue // banner and source comments are fine
		}
		if line == "" {
			t.Errorf("unexpected blank line before first def:\n%s", raw)
		}
		break // first non-comment line reached
	}
}

// ─── Time functions ───────────────────────────────────────────────────────────

func TestSleepTranslation(t *testing.T) {
	out := transpile(t, `Sleep for 1 second.`)
	assertContains(t, out, "import time")
	assertContains(t, out, "time.sleep(1)")
}

func TestCurrentTimeTranslation(t *testing.T) {
	out := transpile(t, `Declare current to be current_time().`)
	assertContains(t, out, "import time")
	assertContains(t, out, `time.strftime("%Y-%m-%d %H:%M:%S")`)
}

func TestElapsedTimeTranslation(t *testing.T) {
	out := transpile(t, `Declare elapsed to be elapsed_time().`)
	assertContains(t, out, "import time")
	assertContains(t, out, "_program_start = time.time()")
	assertContains(t, out, "(time.time() - _program_start)")
}

func TestSleepNoBareFunctionCall(t *testing.T) {
	// sleep() must translate to time.sleep(), not a bare sleep() call.
	out := transpile(t, `Sleep for 2 seconds.`)
	if strings.Contains(out, "sleep(2)") && !strings.Contains(out, "time.sleep(2)") {
		t.Errorf("sleep should emit time.sleep(), not bare sleep(); got:\n%s", out)
	}
}

// ─── Identifier escaping ─────────────────────────────────────────────────────

// TestBuiltinNamesAreEscaped covers a name that collides with a Python
// built-in the generated code itself calls. "Declare sum to be 5." emitted
// "sum = 5", and every later sum(...) — the translation of English's own
// "sum of" — then failed with "int object is not callable".
func TestBuiltinNamesAreEscaped(t *testing.T) {
	out := transpile(t, `Declare sum to be 5.
Declare nums to be [1, 2, 3].
Print sum.
Print sum of nums.`)
	assertContains(t, out, "sum_ = 5", "print(_show(sum_))", "sum(nums)")
}

// TestEscapingIsConsistentAcrossEmissionSites covers the sites the escaping
// was skipped at. A struct field called "class" was emitted as Point(class=1)
// against a def __init__(self, class_=0), so the call did not match the
// definition it was generated beside.
func TestEscapingIsConsistentAcrossEmissionSites(t *testing.T) {
	out := transpile(t, `Declare Shape as a structure with the following fields:
    class is a text with "round" being the default.
thats it.

Declare s to be a new instance of Shape with the following fields:
    class is "square".
thats it.
Print the class of s.`)
	assertContains(t, out,
		"class Shape:",              // the struct itself
		"def __init__(self, class_", // the field as a parameter
		"self.class_ = class_",      // the field as an attribute
		`Shape(class_="square")`,    // the field at the call site
		"s.class_",                  // the field being read
	)
}

// TestStructDefaultsAreNotSharedBetweenInstances covers a list-valued default,
// which Python evaluates once when the def is executed, so every instance
// shared the same list and adding an item to one added it to all of them.
func TestStructDefaultsAreNotSharedBetweenInstances(t *testing.T) {
	out := transpile(t, `Declare Basket as a structure with the following fields:
    items is a list with [] being the default.
thats it.`)
	assertContains(t, out, "items=None", "self.items = [] if items is None else items")
	if strings.Contains(out, "items=[]") {
		t.Errorf("a list default is shared between instances:\n%s", out)
	}
}

// TestCollectionFieldsStartEmpty covers a field with no default at all, which
// the interpreter starts as an empty collection and this rendered as None, so
// the first thing done to it failed.
func TestCollectionFieldsStartEmpty(t *testing.T) {
	out := transpile(t, `Declare Basket as a structure with the following fields:
    items is a list.
    prices is a lookup table.
thats it.`)
	assertContains(t, out, "self.items = [] if items is None else items")
	assertContains(t, out, "self.prices = {} if prices is None else prices")
}

// ─── Rendering ───────────────────────────────────────────────────────────────

// TestPrintUsesEnglishRendering covers what a program writes out. English has
// one number type and prints a whole one without a decimal point, writes
// true/false in lower case and "nothing" for the absence of a value; Python
// writes 5.0, True and None, so every such program said something different
// once transpiled.
func TestPrintUsesEnglishRendering(t *testing.T) {
	out := transpile(t, `Declare ratio to be 10 / 2.
Declare flag to be true.
Declare missing to be nothing.
Print ratio, flag, missing.`)
	assertContains(t, out, "print(_show(ratio), _show(flag), _show(missing))", "def _show")
}

// TestPrintOfALiteralStaysPlain keeps the output readable: a literal Python
// already spells the same way needs no renderer around it.
func TestPrintOfALiteralStaysPlain(t *testing.T) {
	out := transpile(t, `Print "hello", 42.`)
	assertContainsLine(t, out, `print("hello", 42)`)
	if strings.Contains(out, "def _show") {
		t.Errorf("the renderer was emitted for a program that does not need it:\n%s", out)
	}
}

// TestRoundIsHalfAwayFromZero covers rounding: English rounds half away from
// zero, Python's round() rounds half to even, so round(2.5) was 3 in English
// and 2 in the generated Python.
func TestRoundIsHalfAwayFromZero(t *testing.T) {
	out := transpile(t, `Print round(2.5).`)
	assertContains(t, out, "_round(2.5)", "def _round", "import math")
	if strings.Contains(out, "print(_show(round(") {
		t.Errorf("Python's round() was emitted:\n%s", out)
	}
}

// TestCastToBooleanReadsTheWords covers "cast to boolean", which in English
// accepts the words a person would write and refuses anything else. Python's
// bool() calls every non-empty string true, so casting "no" gave False in
// English and True here.
func TestCastToBooleanReadsTheWords(t *testing.T) {
	out := transpile(t, `Declare answer to be "no".
Declare yes to be answer cast to boolean.`)
	assertContains(t, out, "_to_bool(answer)", "def _to_bool")
}

// ─── Module names ────────────────────────────────────────────────────────────

// TestImportedModuleNameMatchesItsFile covers the name an imported English
// file is known by. The .py file was named by stripping the last extension
// while the import took everything before the *first* dot, so "my.lib.abc"
// was written as "my.lib.py" and imported as "my" — a module that does not
// exist, and not a legal module name either way.
func TestImportedModuleNameMatchesItsFile(t *testing.T) {
	dir := t.TempDir()
	libPath := dir + "/my.helper.lib.abc"
	if err := os.WriteFile(libPath, []byte("Declare answer to be 42.\n"), 0644); err != nil {
		t.Fatalf("write lib: %v", err)
	}
	out := transpile(t, `Import "`+libPath+`".
Print answer.`)
	assertContains(t, out, "from my_helper_lib import *")
}

// TestImportOfAPythonModuleIsPassedThrough guards the change above: a path
// with no ".abc" extension names a Python module, which is Python's to
// resolve and not ours to rename.
func TestImportOfAPythonModuleIsPassedThrough(t *testing.T) {
	out := transpile(t, `Import "os.path".
Print "x".`)
	assertContains(t, out, "from os.path import *")
}

// TestRangeStopsBeforeItsEnd covers the range's bounds, which the generated
// Python got wrong in two ways: it added one to the end, so every transpiled
// range had one element the interpreter left out, and it chose the step's sign
// from the direction, so "a range from 10 to 5" counted down where the
// interpreter produces nothing at all.
func TestRangeStopsBeforeItsEnd(t *testing.T) {
	out := transpile(t, `Declare xs to be a range from 1 to 5.`)
	assertContains(t, out, "_range(1, 5)")
	if strings.Contains(out, "+ 1") {
		t.Errorf("the range end was adjusted:\n%s", out)
	}
	if !strings.Contains(out, "return range(int(start), int(end), step)") {
		t.Errorf("the range helper does not mirror the interpreter:\n%s", out)
	}
}
