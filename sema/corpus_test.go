package sema_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Advik-B/english/parser"
	"github.com/Advik-B/english/sema"
	"github.com/Advik-B/english/stdlib"
)

// analyseFile parses and analyses one source file.
func analyseFile(t *testing.T, path string) []*sema.Diagnostic {
	t.Helper()
	src, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("%s: %v", path, err)
	}
	lexer := parser.NewLexer(string(src))
	p := parser.NewParser(lexer.TokenizeAll())
	prog, parseErr := p.Parse()
	if parseErr != nil {
		t.Fatalf("%s: failed to parse: %v", path, parseErr)
	}
	return sema.Check(prog, sema.Config{
		Predefined: stdlib.PredefinedNames(),
		Dir:        filepath.Dir(path),
	})
}

// TestExampleCorpusIsClean analyses every shipped example. These programs are
// the language's own documentation of what valid code looks like, so anything
// the analyser reports here is either a real bug in the example or a false
// positive in the analyser.
func TestExampleCorpusIsClean(t *testing.T) {
	// Import paths in this language are written relative to the directory the
	// program is run from, and these examples are run from the repository
	// root, so analyse them from there.
	t.Chdir("..")

	paths, err := filepath.Glob(filepath.Join("examples", "*.abc"))
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) == 0 {
		t.Fatal("no examples found")
	}

	// This example exists to demonstrate error messages, so it is expected to
	// report problems rather than be clean.
	intentionallyBroken := map[string]bool{"test_errors.abc": true}

	total := 0
	for _, path := range paths {
		diags := analyseFile(t, path)
		if intentionallyBroken[filepath.Base(path)] {
			if len(diags) == 0 {
				t.Errorf("%s is meant to contain mistakes, but nothing was reported",
					filepath.Base(path))
			}
			continue
		}
		if len(diags) == 0 {
			continue
		}
		total += len(diags)
		var b strings.Builder
		for _, d := range diags {
			b.WriteString("\n    line ")
			b.WriteString(itoa(d.Pos.Line))
			b.WriteString(", col ")
			b.WriteString(itoa(d.Pos.Col))
			b.WriteString(": ")
			b.WriteString(d.Message)
		}
		t.Errorf("%s reports %d problem(s):%s", filepath.Base(path), len(diags), b.String())
	}
	if total > 0 {
		t.Logf("%d diagnostic(s) across the corpus", total)
	}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	if neg {
		return "-" + string(digits)
	}
	return string(digits)
}
