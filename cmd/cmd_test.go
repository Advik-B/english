package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestEngineNamed covers the --vm flag, which fell through to the instruction
// VM for any value it did not recognise. So "--vm=asr" ran a different engine
// than the one asked for and said nothing about it.
func TestEngineNamed(t *testing.T) {
	for _, good := range []struct{ flag, want string }{
		{"ivm", engineIVM},
		{"ast", engineAST},
		{"AST", engineAST},
		{"  ivm  ", engineIVM},
		{"", engineIVM},
	} {
		got, err := engineNamed(good.flag)
		if err != nil {
			t.Errorf("--vm=%q was rejected: %v", good.flag, err)
			continue
		}
		if got != good.want {
			t.Errorf("--vm=%q selected %q, want %q", good.flag, got, good.want)
		}
	}

	for _, bad := range []string{"asr", "vm", "tree", "ivm2", "a s t"} {
		if _, err := engineNamed(bad); err == nil {
			t.Errorf("--vm=%q was accepted", bad)
		}
	}
}

// TestTranspileWritesNothingWhenAnImportFails covers the transpiler's output
// tree. The recursion used to write each .py file as it finished and exit the
// process from twelve places along the way, so a bad file part-way through the
// imports left a half-written tree behind: some files new, some left over from
// an earlier run, and no way to tell which were which.
func TestTranspileWritesNothingWhenAnImportFails(t *testing.T) {
	dir := t.TempDir()

	broken := filepath.Join(dir, "broken.abc")
	if err := os.WriteFile(broken, []byte("Declare to be 3.\n"), 0644); err != nil {
		t.Fatalf("cannot write the broken import: %v", err)
	}
	good := filepath.Join(dir, "good.abc")
	if err := os.WriteFile(good, []byte("Declare answer to be 42.\n"), 0644); err != nil {
		t.Fatalf("cannot write the good import: %v", err)
	}
	main := filepath.Join(dir, "main.abc")
	source := "Import \"" + filepath.ToSlash(good) + "\".\n" +
		"Import \"" + filepath.ToSlash(broken) + "\".\n" +
		"Print answer.\n"
	if err := os.WriteFile(main, []byte(source), 0644); err != nil {
		t.Fatalf("cannot write the main file: %v", err)
	}

	err := TranspileFileOptions(main, false)
	if err == nil {
		t.Fatal("transpiling a tree with a broken import reported success")
	}

	// The good import parses, so it would have been written first.
	for _, name := range []string{"main.py", "good.py", "broken.py"} {
		if _, statErr := os.Stat(filepath.Join(dir, name)); statErr == nil {
			t.Errorf("%s was written even though the tree failed", name)
		}
	}
}

// TestTranspileWritesTheWholeTree guards the change above against writing
// nothing at all.
func TestTranspileWritesTheWholeTree(t *testing.T) {
	dir := t.TempDir()

	lib := filepath.Join(dir, "lib.abc")
	if err := os.WriteFile(lib, []byte("Declare answer to be 42.\n"), 0644); err != nil {
		t.Fatalf("cannot write the import: %v", err)
	}
	main := filepath.Join(dir, "main.abc")
	source := "Import \"" + filepath.ToSlash(lib) + "\".\nPrint answer.\n"
	if err := os.WriteFile(main, []byte(source), 0644); err != nil {
		t.Fatalf("cannot write the main file: %v", err)
	}

	if err := TranspileFileOptions(main, false); err != nil {
		t.Fatalf("transpiling a valid tree failed: %v", err)
	}
	for _, name := range []string{"main.py", "lib.py"} {
		content, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			t.Errorf("%s was not written: %v", name, err)
			continue
		}
		if !strings.Contains(string(content), "answer") {
			t.Errorf("%s does not mention the name it should define or use", name)
		}
	}
}
