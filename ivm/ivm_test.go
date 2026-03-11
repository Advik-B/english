package ivm_test

import (
"bytes"
"english/ivm"
"english/parser"
"english/vm/stdlib"
"fmt"
"io"
"os"
"strings"
"testing"
)

// compileSource parses English source code and compiles it to a Chunk.
func compileSource(src string) (*ivm.Chunk, error) {
lexer := parser.NewLexer(src)
tokens := lexer.TokenizeAll()
p := parser.NewParser(tokens)
prog, err := p.Parse()
if err != nil {
return nil, err
}
return ivm.Compile(prog)
}

// run compiles and executes English source code, returning the last value.
func run(src string) (interface{}, error) {
chunk, err := compileSource(src)
if err != nil {
return nil, err
}
return ivm.Execute(chunk, stdlib.Eval, stdlib.PredefinedValues())
}

// captureOutput captures stdout during execution.
func captureOutput(fn func()) string {
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

// ─── Arithmetic ───────────────────────────────────────────────────────────────

func TestArithmeticAdd(t *testing.T) {
_, err := run(`Declare x to be 3 + 4.`)
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
}

func TestArithmeticAll(t *testing.T) {
tests := []string{
`Declare x to be 10 + 5.`,
`Declare x to be 10 - 3.`,
`Declare x to be 4 * 5.`,
`Declare x to be 10 / 2.`,
`Declare x to be the remainder of 10 divided by 3.`,
}
for _, src := range tests {
_, err := run(src)
if err != nil {
t.Fatalf("%q: error: %v", src, err)
}
}
}

// ─── Variables ────────────────────────────────────────────────────────────────

func TestVariableDecl(t *testing.T) {
out := captureOutput(func() {
_, err := run(`Declare x to be 42.
Print the value of x.`)
if err != nil {
t.Errorf("unexpected error: %v", err)
}
})
if !strings.Contains(out, "42") {
t.Errorf("expected output to contain 42, got: %q", out)
}
}

func TestConstantDecl(t *testing.T) {
_, err := run(`Declare x to always be 10.`)
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
}

func TestAssignment(t *testing.T) {
out := captureOutput(func() {
_, err := run(`Declare x to be 5.
Set x to be 10.
Print the value of x.`)
if err != nil {
t.Errorf("unexpected error: %v", err)
}
})
if !strings.Contains(out, "10") {
t.Errorf("expected 10, got %q", out)
}
}

// ─── Strings ──────────────────────────────────────────────────────────────────

func TestStringLiteral(t *testing.T) {
out := captureOutput(func() {
_, err := run(`Print "hello world".`)
if err != nil {
t.Errorf("unexpected error: %v", err)
}
})
if !strings.Contains(out, "hello world") {
t.Errorf("expected 'hello world', got %q", out)
}
}

// ─── Booleans ─────────────────────────────────────────────────────────────────

func TestBooleanLogic(t *testing.T) {
out := captureOutput(func() {
_, err := run(`Declare a to be true.
Declare b to be false.
If a and not b, then
    Print "yes".
thats it.`)
if err != nil {
t.Errorf("unexpected error: %v", err)
}
})
if !strings.Contains(out, "yes") {
t.Errorf("expected 'yes', got %q", out)
}
}

// ─── If/Else ──────────────────────────────────────────────────────────────────

func TestIfElse(t *testing.T) {
out := captureOutput(func() {
_, err := run(`Declare x to be 10.
If x is greater than 5, then
    Print "big".
otherwise
    Print "small".
thats it.`)
if err != nil {
t.Errorf("unexpected error: %v", err)
}
})
if !strings.Contains(out, "big") {
t.Errorf("expected 'big', got %q", out)
}
}

func TestIfElseIf(t *testing.T) {
out := captureOutput(func() {
_, err := run(`Declare x to be 5.
If x is greater than 10, then
    Print "big".
otherwise if x is greater than 3, then
    Print "medium".
otherwise
    Print "small".
thats it.`)
if err != nil {
t.Errorf("unexpected error: %v", err)
}
})
if !strings.Contains(out, "medium") {
t.Errorf("expected 'medium', got %q", out)
}
}

// ─── While loop ───────────────────────────────────────────────────────────────

func TestWhileLoop(t *testing.T) {
out := captureOutput(func() {
_, err := run(`Declare i to be 0.
repeat the following while i is less than 3:
    Set i to be i + 1.
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

// ─── For loop ─────────────────────────────────────────────────────────────────

func TestForLoop(t *testing.T) {
out := captureOutput(func() {
_, err := run(`Declare total to be 0.
repeat the following 5 times:
    Set total to be total + 1.
thats it.
Print the value of total.`)
if err != nil {
t.Errorf("unexpected error: %v", err)
}
})
if !strings.Contains(out, "5") {
t.Errorf("expected 5, got %q", out)
}
}

// ─── For-each loop ────────────────────────────────────────────────────────────

func TestForEachLoop(t *testing.T) {
out := captureOutput(func() {
_, err := run(`Declare nums to be [1, 2, 3].
Declare total to be 0.
for each n in nums, do the following:
    Set total to be total + n.
thats it.
Print the value of total.`)
if err != nil {
t.Errorf("unexpected error: %v", err)
}
})
if !strings.Contains(out, "6") {
t.Errorf("expected 6, got %q", out)
}
}

// ─── Functions ────────────────────────────────────────────────────────────────

func TestFunctionCall(t *testing.T) {
out := captureOutput(func() {
_, err := run(`Declare function add that takes a and b and does the following:
    Return a + b.
thats it.
Declare result to be 0.
Set result to be the result of calling add with 3 and 4.
Print the value of result.`)
if err != nil {
t.Errorf("unexpected error: %v", err)
}
})
if !strings.Contains(out, "7") {
t.Errorf("expected 7, got %q", out)
}
}

func TestRecursion(t *testing.T) {
out := captureOutput(func() {
_, err := run(`Declare function factorial that takes n and does the following:
    If n is less than or equal to 1, then
        Return 1.
    thats it.
    Declare prev to be 0.
    Set prev to be the result of calling factorial with n - 1.
    Return n * prev.
thats it.
Declare result to be 0.
Set result to be the result of calling factorial with 5.
Print the value of result.`)
if err != nil {
t.Errorf("unexpected error: %v", err)
}
})
if !strings.Contains(out, "120") {
t.Errorf("expected 120, got %q", out)
}
}

// ─── Lists ────────────────────────────────────────────────────────────────────

func TestListLiteral(t *testing.T) {
out := captureOutput(func() {
_, err := run(`Declare items to be [10, 20, 30].
Print the item at position 2 of items.`)
if err != nil {
t.Errorf("unexpected error: %v", err)
}
})
if !strings.Contains(out, "30") {
t.Errorf("expected 30, got %q", out)
}
}

func TestListLength(t *testing.T) {
out := captureOutput(func() {
_, err := run(`Declare items to be [1, 2, 3, 4].
Print the number of items.`)
if err != nil {
t.Errorf("unexpected error: %v", err)
}
})
if !strings.Contains(out, "4") {
t.Errorf("expected 4, got %q", out)
}
}

// ─── Nothing ─────────────────────────────────────────────────────────────────

func TestNothingLiteral(t *testing.T) {
out := captureOutput(func() {
_, err := run(`Declare x to be nothing.
Print the value of x.`)
if err != nil {
t.Errorf("unexpected error: %v", err)
}
})
if !strings.Contains(out, "nothing") {
t.Errorf("expected 'nothing', got %q", out)
}
}

// ─── Nil check ────────────────────────────────────────────────────────────────

func TestNilCheck(t *testing.T) {
out := captureOutput(func() {
_, err := run(`Declare x to be nothing.
If x is nothing, then
    Print "nil".
thats it.`)
if err != nil {
t.Errorf("unexpected error: %v", err)
}
})
if !strings.Contains(out, "nil") {
t.Errorf("expected 'nil', got %q", out)
}
}

// ─── Toggle ───────────────────────────────────────────────────────────────────

func TestToggle(t *testing.T) {
out := captureOutput(func() {
_, err := run(`Declare flag to be true.
Toggle flag.
Print the value of flag.`)
if err != nil {
t.Errorf("unexpected error: %v", err)
}
})
if !strings.Contains(out, "false") {
t.Errorf("expected false, got %q", out)
}
}

// ─── Break ────────────────────────────────────────────────────────────────────

func TestBreak(t *testing.T) {
out := captureOutput(func() {
_, err := run(`Declare i to be 0.
repeat forever:
    If i is equal to 3, then
        break out of this loop.
    thats it.
    Set i to be i + 1.
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

// ─── Stdlib ───────────────────────────────────────────────────────────────────

func TestStdlibSqrt(t *testing.T) {
out := captureOutput(func() {
_, err := run(`Declare result to be 0.
Set result to be the result of calling sqrt with 9.
Print the value of result.`)
if err != nil {
t.Errorf("unexpected error: %v", err)
}
})
if !strings.Contains(out, "3") {
t.Errorf("expected 3, got %q", out)
}
}

func TestPredefinedPi(t *testing.T) {
out := captureOutput(func() {
_, err := run(`Print the value of pi.`)
if err != nil {
t.Errorf("unexpected error: %v", err)
}
})
if !strings.Contains(out, "3.14") {
t.Errorf("expected pi ~3.14, got %q", out)
}
}

// ─── Type operations ──────────────────────────────────────────────────────────

func TestTypeof(t *testing.T) {
out := captureOutput(func() {
_, err := run(`Declare x to be 42.
Print the type of x.`)
if err != nil {
t.Errorf("unexpected error: %v", err)
}
})
if out == "" {
t.Errorf("expected non-empty output for typeof")
}
}

func TestCast(t *testing.T) {
out := captureOutput(func() {
_, err := run(`Declare n to be 42.
Declare s to be n cast to text.
Print the value of s.`)
if err != nil {
t.Errorf("unexpected error: %v", err)
}
})
if !strings.Contains(out, "42") {
t.Errorf("expected '42', got %q", out)
}
}

// ─── Encode/Decode roundtrip ──────────────────────────────────────────────────

func TestEncodeDecodeRoundtrip(t *testing.T) {
src := `Declare x to be 42.
Declare y to be x + 8.`
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

if len(decoded.Code) != len(chunk.Code) {
t.Errorf("code length mismatch: got %d, want %d", len(decoded.Code), len(chunk.Code))
}
if len(decoded.Constants) != len(chunk.Constants) {
t.Errorf("constants length mismatch: got %d, want %d", len(decoded.Constants), len(chunk.Constants))
}

_, execErr := ivm.Execute(decoded, stdlib.Eval, stdlib.PredefinedValues())
if execErr != nil {
t.Fatalf("execute decoded chunk error: %v", execErr)
}
}

func TestEncodeDecodeWithFunctions(t *testing.T) {
src := `Declare function double that takes x and does the following:
    Return x * 2.
thats it.
Declare result to be 0.
Set result to be the result of calling double with 21.`
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
t.Fatalf("execute decoded chunk error: %v", execErr)
}
}

// ─── Try/catch ────────────────────────────────────────────────────────────────

func TestTryCatch(t *testing.T) {
out := captureOutput(func() {
_, err := run(`Try doing the following:
    Raise "oops".
on error:
    Print "caught".
thats it.`)
if err != nil {
t.Errorf("unexpected error: %v", err)
}
})
if !strings.Contains(out, "caught") {
t.Errorf("expected 'caught', got %q", out)
}
}

// ─── Swap ─────────────────────────────────────────────────────────────────────

func TestSwap(t *testing.T) {
out := captureOutput(func() {
_, err := run(`Declare a to be 1.
Declare b to be 2.
Swap a and b.
Print the value of a.
Print the value of b.`)
if err != nil {
t.Errorf("unexpected error: %v", err)
}
})
if !strings.Contains(out, "2") || !strings.Contains(out, "1") {
t.Errorf("expected swap result, got %q", out)
}
}

// ─── Lookup table ─────────────────────────────────────────────────────────────

func TestLookupTable(t *testing.T) {
out := captureOutput(func() {
_, err := run(`Declare t to be a lookup table.
Set t at "key" to be "value".
Print the entry "key" in t.`)
if err != nil {
t.Errorf("unexpected error: %v", err)
}
})
if !strings.Contains(out, "value") {
t.Errorf("expected 'value', got %q", out)
}
}

// ─── Error type declaration ───────────────────────────────────────────────────

func TestErrorTypeDecl(t *testing.T) {
_, err := run(`Declare NetworkError as an error type.`)
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
}

// ─── String concatenation ─────────────────────────────────────────────────────

func TestStringConcat(t *testing.T) {
out := captureOutput(func() {
_, err := run(`Declare s to be "Hello" + " World".
Print the value of s.`)
if err != nil {
t.Errorf("unexpected error: %v", err)
}
})
if !strings.Contains(out, "Hello World") {
t.Errorf("expected 'Hello World', got %q", out)
}
}

// ─── Error variable not leaking across try blocks ─────────────────────────────

func TestErrorVarNotLeakingAcrossTryBlocks(t *testing.T) {
out := captureOutput(func() {
_, err := run(`Declare NetworkError as an error type.
Try doing the following:
    Raise "first" as NetworkError.
on NetworkError:
    Print "caught:", error.
thats it.
Try doing the following:
    Raise "second".
on error:
    Print "caught:", error.
thats it.`)
if err != nil {
t.Errorf("unexpected error: %v", err)
}
})
lines := strings.Split(strings.TrimSpace(out), "\n")
if len(lines) != 2 {
t.Fatalf("expected 2 output lines, got %d: %q", len(lines), out)
}
if !strings.Contains(lines[0], "first") {
t.Errorf("line 0: expected 'first', got %q", lines[0])
}
if !strings.Contains(lines[1], "second") {
t.Errorf("line 1: expected 'second', got %q (error leaked from previous block)", lines[1])
}
}

// ─── Safe import runs only declarations ──────────────────────────────────────

func TestSafeImportRunsOnlyDeclarations(t *testing.T) {
tmpDir, err := os.MkdirTemp("", "ivm_safe_import")
if err != nil {
t.Fatal(err)
}
defer os.RemoveAll(tmpDir)

libPath := tmpDir + "/lib.abc"
if err := os.WriteFile(libPath, []byte(`
Print "SHOULD NOT PRINT".
Declare version to always be "1.0".
Declare function greet and does the following:
    Print "Hello from lib".
thats it.
Print "ALSO SHOULD NOT PRINT".
`), 0644); err != nil {
t.Fatal(err)
}

src := `Import all from "` + libPath + `" safely.
Declare r to be "".
Set r to be version.
Print r.`

out := captureOutput(func() {
_, err := run(src)
if err != nil {
t.Errorf("unexpected error: %v", err)
}
})
if strings.Contains(out, "SHOULD NOT PRINT") {
t.Errorf("safe import ran side-effectful code; got %q", out)
}
if !strings.Contains(out, "1.0") {
t.Errorf("expected version '1.0' to be imported; got %q", out)
}
}

// ─── Source embedding ─────────────────────────────────────────────────────────

func TestEncodeDecodeFileWithSource(t *testing.T) {
src := `Declare x to be 42.
Print the value of x.`
chunk, err := compileSource(src)
if err != nil {
t.Fatalf("compile error: %v", err)
}
data, err := ivm.EncodeFileWithSource(chunk, src)
if err != nil {
t.Fatalf("encode error: %v", err)
}
_, got, err := ivm.DecodeFileAll(data)
if err != nil {
t.Fatalf("decode error: %v", err)
}
if got != src {
t.Errorf("embedded source mismatch:\nwant %q\ngot  %q", src, got)
}
}

func TestDecodeFileAllNoSource(t *testing.T) {
src := `Declare x to be 7.`
chunk, err := compileSource(src)
if err != nil {
t.Fatalf("compile error: %v", err)
}
data, err := ivm.EncodeFile(chunk) // no source embedded
if err != nil {
t.Fatalf("encode error: %v", err)
}
_, embeddedSrc, err := ivm.DecodeFileAll(data)
if err != nil {
t.Fatalf("decode error: %v", err)
}
if embeddedSrc != "" {
t.Errorf("expected empty embedded source, got %q", embeddedSrc)
}
}

func TestDecodeFileBackwardsCompat(t *testing.T) {
// DecodeFile must still work on files produced without a source trailer
// (i.e. compiled with EncodeFile which omits the trailer).
src := `Declare y to be 99.`
chunk, err := compileSource(src)
if err != nil {
t.Fatalf("compile error: %v", err)
}
data, err := ivm.EncodeFile(chunk) // no source trailer
if err != nil {
t.Fatalf("encode error: %v", err)
}
decoded, err := ivm.DecodeFile(data)
if err != nil {
t.Fatalf("DecodeFile error: %v", err)
}
if decoded == nil {
t.Fatal("decoded chunk is nil")
}
// DecodeFileAll should return empty source for trailer-less files.
_, embeddedSrc, err := ivm.DecodeFileAll(data)
if err != nil {
t.Fatalf("DecodeFileAll error: %v", err)
}
if embeddedSrc != "" {
t.Errorf("expected empty embedded source for trailer-less file, got %q", embeddedSrc)
}
}

// ─── Decompiler ───────────────────────────────────────────────────────────────

// decompileSource is a test helper: compile a source snippet, decompile the
// resulting chunk to Python, then exec it and return the combined stdout.
func decompileSource(src string) (string, error) {
chunk, err := compileSource(src)
if err != nil {
return "", fmt.Errorf("compile: %w", err)
}
py := ivm.Decompile(chunk)
return py, nil
}

func TestDecompileHelloWorld(t *testing.T) {
py, err := decompileSource(`Print "Hello, World!".`)
if err != nil {
t.Fatal(err)
}
if !strings.Contains(py, `print("Hello, World!")`) {
t.Errorf("expected print statement, got:\n%s", py)
}
}

func TestDecompileVariables(t *testing.T) {
py, err := decompileSource(`
Declare x to be 42.
Declare name to be "Alice".
Print x, name.
`)
if err != nil {
t.Fatal(err)
}
if !strings.Contains(py, "x = 42") || !strings.Contains(py, `name = "Alice"`) {
t.Errorf("unexpected output:\n%s", py)
}
}

func TestDecompileIfElse(t *testing.T) {
py, err := decompileSource(`Declare x to be 10.
If x is greater than 5, then
    Print "big".
otherwise if x is equal to 5, then
    Print "five".
otherwise
    Print "small".
thats it.`)
if err != nil {
t.Fatal(err)
}
if !strings.Contains(py, "if") || !strings.Contains(py, "elif") || !strings.Contains(py, "else:") {
t.Errorf("missing if/elif/else in:\n%s", py)
}
}

func TestDecompileWhileLoop(t *testing.T) {
py, err := decompileSource(`Declare i to be 0.
repeat the following while i is less than 3:
    Print i.
    Set i to be i + 1.
thats it.`)
if err != nil {
t.Fatal(err)
}
if !strings.Contains(py, "while") {
t.Errorf("missing while loop in:\n%s", py)
}
if !strings.Contains(py, "i = 0") {
t.Errorf("missing variable init in:\n%s", py)
}
}

func TestDecompileForEachLoop(t *testing.T) {
py, err := decompileSource(`Declare items to be [1, 2, 3].
For each item in items, do the following:
    Print item.
thats it.`)
if err != nil {
t.Fatal(err)
}
if !strings.Contains(py, "for item in") {
t.Errorf("missing for-each in:\n%s", py)
}
}

func TestDecompileRepeatLoop(t *testing.T) {
py, err := decompileSource(`Repeat the following 3 times:
    Print "hi".
thats it.`)
if err != nil {
t.Fatal(err)
}
if !strings.Contains(py, "for _ in range(") {
t.Errorf("missing repeat loop in:\n%s", py)
}
}

func TestDecompileFunction(t *testing.T) {
py, err := decompileSource(`Declare function double that takes x and does the following:
    Return x * 2.
thats it.
Declare result to be 0.
Set result to be the result of calling double with 5.
Print result.`)
if err != nil {
t.Fatal(err)
}
if !strings.Contains(py, "def double(x):") {
t.Errorf("missing function def in:\n%s", py)
}
}

func TestDecompileTryCatch(t *testing.T) {
py, err := decompileSource(`Try doing the following:
    Raise "oops".
on error:
    Print "caught:", error.
thats it.`)
if err != nil {
t.Fatal(err)
}
if !strings.Contains(py, "try:") || !strings.Contains(py, "except") {
t.Errorf("missing try/except in:\n%s", py)
}
}

func TestDecompileNestedTry(t *testing.T) {
py, err := decompileSource(`Declare NetworkError as an error type.
Declare DatabaseError as an error type.
Try doing the following:
    Try doing the following:
        Raise "inner" as DatabaseError.
    on NetworkError:
        Print "wrong handler".
    thats it.
on DatabaseError:
    Print "outer".
thats it.`)
if err != nil {
t.Fatal(err)
}
// Should have two try blocks and no duplicate print statements
count := strings.Count(py, "try:")
if count != 2 {
t.Errorf("expected 2 try blocks, got %d:\n%s", count, py)
}
// The "outer" print should appear exactly once
if strings.Count(py, `"outer"`) != 1 {
t.Errorf("duplicate print statement detected:\n%s", py)
}
}

func TestDecompileLogicalOperators(t *testing.T) {
py, err := decompileSource(`Declare x to be 5.
Declare y to be 10.
If (x is greater than 0 and y is greater than 0), then
    Print "both positive".
thats it.`)
if err != nil {
t.Fatal(err)
}
if !strings.Contains(py, " and ") {
t.Errorf("missing and operator in:\n%s", py)
}
}

func TestDecompileCustomErrorType(t *testing.T) {
py, err := decompileSource(`Declare MyError as an error type.
Declare ChildError as a type of MyError.`)
if err != nil {
t.Fatal(err)
}
if !strings.Contains(py, "class MyError(Exception): pass") {
t.Errorf("missing error class def in:\n%s", py)
}
if !strings.Contains(py, "class ChildError(MyError): pass") {
t.Errorf("missing child error class def in:\n%s", py)
}
}

func TestDecompileMathFunctions(t *testing.T) {
py, err := decompileSource(`Declare result to be sqrt(16).
Print result.`)
if err != nil {
t.Fatal(err)
}
if !strings.Contains(py, "import math") || !strings.Contains(py, "math.sqrt") {
t.Errorf("missing math import or sqrt call:\n%s", py)
}
}

func TestDecompileLookupTable(t *testing.T) {
py, err := decompileSource(`Declare t to be a lookup table.
Set t at "greeting" to be "hello".
Print the entry "greeting" in t.`)
if err != nil {
t.Fatal(err)
}
if !strings.Contains(py, "t = {}") || !strings.Contains(py, `"greeting"`) {
t.Errorf("unexpected lookup table output:\n%s", py)
}
}

func TestDecompileStructDefinition(t *testing.T) {
py, err := decompileSource(`declare Point as a structure with the following fields:
    x is a number with 0 being the default.
    y is a number with 0 being the default.
thats it.`)
if err != nil {
t.Fatal(err)
}
if !strings.Contains(py, "class Point:") {
t.Errorf("missing struct class in:\n%s", py)
}
if !strings.Contains(py, "def __init__(self,") {
t.Errorf("missing __init__ in:\n%s", py)
}
}
