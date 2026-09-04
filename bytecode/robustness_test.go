package bytecode_test

import (
	"math/rand"
	"testing"

	"github.com/Advik-B/english/bytecode"
	"github.com/Advik-B/english/parser"
)

// sampleV1Bytecode encodes a program in the version-1 (AST) bytecode format.
func sampleV1Bytecode(t *testing.T) []byte {
	t.Helper()
	src := `Declare greeting to be "hello".
Declare count to be 0.
Declare items to be [1, 2, 3].

Declare function twice that takes n as number, and gives back a number, and does the following:
    Return n * 2.
thats it.

Repeat the following while count is less than 3:
    Set count to be count + 1.
    Print greeting, count.
thats it.

If count is greater than 2, then
    Print "done".
otherwise
    Print "not done".
thats it.`

	lexer := parser.NewLexer(src)
	p := parser.NewParser(lexer.TokenizeAll())
	prog, err := p.Parse()
	if err != nil {
		t.Fatalf("sample program failed to parse: %v", err)
	}
	data, err := bytecode.NewEncoder().Encode(prog)
	if err != nil {
		t.Fatalf("sample program failed to encode: %v", err)
	}
	return data
}

// decodeV1Safely reports a panic instead of letting it take the test binary down.
func decodeV1Safely(data []byte) (panicked interface{}) {
	defer func() {
		if r := recover(); r != nil {
			panicked = r
		}
	}()
	_, _ = bytecode.NewDecoder(data).Decode()
	return nil
}

// TestV1DecodeTruncatedDoesNotPanic covers the class of bug where a uint32
// length prefix read straight from the file drove an unchecked make(), so a
// truncated or hostile file could request a multi-gigabyte allocation before
// the truncation was ever noticed.
func TestV1DecodeTruncatedDoesNotPanic(t *testing.T) {
	data := sampleV1Bytecode(t)
	for n := 0; n < len(data); n++ {
		if p := decodeV1Safely(data[:n]); p != nil {
			t.Fatalf("truncating to %d byte(s) panicked: %v", n, p)
		}
	}
}

// TestV1DecodeCorruptedDoesNotPanic flips bytes throughout a valid file.
func TestV1DecodeCorruptedDoesNotPanic(t *testing.T) {
	data := sampleV1Bytecode(t)
	rng := rand.New(rand.NewSource(7))

	for i := 0; i < len(data); i++ {
		for _, v := range []byte{0x00, 0x01, 0x7F, 0xFF} {
			corrupt := make([]byte, len(data))
			copy(corrupt, data)
			corrupt[i] = v
			if p := decodeV1Safely(corrupt); p != nil {
				t.Fatalf("setting byte %d to %#x panicked: %v", i, v, p)
			}
		}
	}

	for iter := 0; iter < 300; iter++ {
		corrupt := make([]byte, len(data))
		copy(corrupt, data)
		for k := 0; k < 4; k++ {
			corrupt[rng.Intn(len(corrupt))] = byte(rng.Intn(256))
		}
		if p := decodeV1Safely(corrupt); p != nil {
			t.Fatalf("random corruption %d panicked: %v", iter, p)
		}
	}
}

// TestV1ValidRoundTrip guards the hardening against over-reach.
func TestV1ValidRoundTrip(t *testing.T) {
	data := sampleV1Bytecode(t)
	prog, err := bytecode.NewDecoder(data).Decode()
	if err != nil {
		t.Fatalf("valid bytecode failed to decode: %v", err)
	}
	if len(prog.Statements) == 0 {
		t.Fatal("decoded program has no statements")
	}
}

// TestV1EncodeMethodCallStatement covers the nil dereference where a
// CallStatement carrying a MethodCall (rather than a FunctionCall) was encoded
// by dereferencing the nil FunctionCall field.
func TestV1EncodeMethodCallStatement(t *testing.T) {
	src := `Declare s to be "hi".
Call s's uppercase.`
	lexer := parser.NewLexer(src)
	p := parser.NewParser(lexer.TokenizeAll())
	prog, err := p.Parse()
	if err != nil {
		t.Skipf("method-call statement did not parse: %v", err)
	}

	done := make(chan struct{})
	var encErr error
	var panicked interface{}
	go func() {
		defer func() {
			if r := recover(); r != nil {
				panicked = r
			}
			close(done)
		}()
		_, encErr = bytecode.NewEncoder().Encode(prog)
	}()
	<-done

	if panicked != nil {
		t.Fatalf("encoding a method-call statement panicked: %v", panicked)
	}
	if encErr == nil {
		t.Log("method calls are now representable in the v1 format")
	}
}
