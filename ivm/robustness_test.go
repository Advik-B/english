package ivm_test

import (
	"math/rand"
	"testing"

	"github.com/Advik-B/english/ivm"
	"github.com/Advik-B/english/parser"
	"github.com/Advik-B/english/stdlib"
)

// sampleBytecode compiles a program exercising most of the encoding: functions,
// structs, error types, loops, lookup tables and casts.
func sampleBytecode(t *testing.T) []byte {
	t.Helper()
	src := `Declare AppError as an error type.

Declare Point as a structure with the following fields:
    x is a number with 0 being the default.
    y is a number with 0 being the default.
thats it.

Declare function add that takes a and b and does the following:
    Return a + b.
thats it.

Declare total to be 0.
Declare i to be 0.
Repeat the following while i is less than 3:
    Set total to be the result of calling add with total and 2.
    Set i to be i + 1.
thats it.

Declare scores to be a lookup table.
Set scores at "k" to be total.
Print scores at "k".
Print total cast to text.`

	lexer := parser.NewLexer(src)
	p := parser.NewParser(lexer.TokenizeAll())
	prog, err := p.Parse()
	if err != nil {
		t.Fatalf("sample program failed to parse: %v", err)
	}
	chunk, err := ivm.Compile(prog)
	if err != nil {
		t.Fatalf("sample program failed to compile: %v", err)
	}
	data, err := ivm.EncodeFile(chunk)
	if err != nil {
		t.Fatalf("sample program failed to encode: %v", err)
	}
	return data
}

// decodeSafely reports a panic instead of letting it take the test binary down.
//
// It decodes only. A corrupted chunk that still passes validation is a
// well-formed program with different instructions, and running it could loop
// forever by design; what is being asserted here is that decoding never panics
// and never allocates unboundedly. Execution of a *valid* file is covered by
// TestValidRoundTripStillWorks.
func decodeSafely(data []byte) (panicked interface{}) {
	defer func() {
		if r := recover(); r != nil {
			panicked = r
		}
	}()
	_, _ = ivm.DecodeFile(data)
	return nil
}

// TestDecodeTruncatedFileDoesNotPanic covers the class of bug where a length
// prefix read from the file drove an unchecked make(), or an instruction
// operand indexed a constant pool out of range. Every prefix of a valid file
// must be rejected with an error rather than a panic or a huge allocation.
func TestDecodeTruncatedFileDoesNotPanic(t *testing.T) {
	data := sampleBytecode(t)
	for n := 0; n < len(data); n++ {
		if p := decodeSafely(data[:n]); p != nil {
			t.Fatalf("truncating to %d byte(s) panicked: %v", n, p)
		}
	}
}

// TestDecodeCorruptedFileDoesNotPanic flips bytes throughout a valid file. The
// decoder must always either succeed or return an error.
func TestDecodeCorruptedFileDoesNotPanic(t *testing.T) {
	data := sampleBytecode(t)
	rng := rand.New(rand.NewSource(1))

	for i := 0; i < len(data); i++ {
		for _, v := range []byte{0x00, 0x01, 0x7F, 0xFF} {
			corrupt := make([]byte, len(data))
			copy(corrupt, data)
			corrupt[i] = v
			if p := decodeSafely(corrupt); p != nil {
				t.Fatalf("setting byte %d to %#x panicked: %v", i, v, p)
			}
		}
	}

	// A few hundred random multi-byte corruptions for good measure.
	for iter := 0; iter < 300; iter++ {
		corrupt := make([]byte, len(data))
		copy(corrupt, data)
		for k := 0; k < 4; k++ {
			corrupt[rng.Intn(len(corrupt))] = byte(rng.Intn(256))
		}
		if p := decodeSafely(corrupt); p != nil {
			t.Fatalf("random corruption %d panicked: %v", iter, p)
		}
	}
}

// TestValidRoundTripStillWorks guards the hardening above against over-reach:
// a file the encoder produced must still decode and run.
func TestValidRoundTripStillWorks(t *testing.T) {
	data := sampleBytecode(t)
	chunk, err := ivm.DecodeFile(data)
	if err != nil {
		t.Fatalf("valid bytecode failed to decode: %v", err)
	}
	if _, err := ivm.Execute(chunk, stdlib.Eval, stdlib.PredefinedValues()); err != nil {
		t.Fatalf("valid bytecode failed to execute: %v", err)
	}
}
