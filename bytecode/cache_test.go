package bytecode

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Advik-B/english/ast"
)

func TestGetCachePath(t *testing.T) {
	tests := []struct {
		name       string
		sourcePath string
	}{
		{"simple file", "test.abc"},
		{"nested file", "examples/math_library.abc"},
		{"absolute path", "/home/user/program.abc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cachePath := GetCachePath(tt.sourcePath)

			// Check that cache path is in __engcache__ directory
			if filepath.Dir(cachePath) != CacheDir {
				t.Errorf("Cache path should be in %s directory, got: %s", CacheDir, cachePath)
			}

			// Check that cache path has .101 extension
			if filepath.Ext(cachePath) != ".101" {
				t.Errorf("Cache path should have .101 extension, got: %s", cachePath)
			}
		})
	}
}

// TestCacheValidityFollowsContent covers cache invalidation, which compared
// modification times: a cache file not older than its source was used.
//
// That is wrong in the dangerous direction. A checkout, a restore from a
// backup or an archive extraction can leave a source file older than a cache
// built from different text, and the stale cache was then used silently — the
// program you edited was not the program that ran.
func TestCacheValidityFollowsContent(t *testing.T) {
	tmpDir := t.TempDir()
	sourcePath := filepath.Join(tmpDir, "source.abc")
	cachePath := filepath.Join(tmpDir, CacheDir, "test.101")

	original := []byte(`Print "hello".`)
	if err := os.WriteFile(sourcePath, original, 0644); err != nil {
		t.Fatalf("cannot write the source: %v", err)
	}

	// Nothing cached yet.
	if IsCacheValid(sourcePath, cachePath) {
		t.Error("a cache that does not exist is reported as valid")
	}

	stamp, err := StampOfSource(sourcePath)
	if err != nil {
		t.Fatalf("cannot stamp the source: %v", err)
	}
	if err := WriteBytecodeCache(cachePath, stamp, []byte("payload")); err != nil {
		t.Fatalf("cannot write the cache: %v", err)
	}
	if !IsCacheValid(sourcePath, cachePath) {
		t.Error("a cache built from this exact source is reported as stale")
	}

	// The source changes, and its modification time is deliberately set
	// *older* than the cache file — the case timestamps get wrong.
	if err := os.WriteFile(sourcePath, []byte(`Print "goodbye".`), 0644); err != nil {
		t.Fatalf("cannot update the source: %v", err)
	}
	past := time.Now().Add(-time.Hour)
	if err := os.Chtimes(sourcePath, past, past); err != nil {
		t.Fatalf("cannot backdate the source: %v", err)
	}
	if IsCacheValid(sourcePath, cachePath) {
		t.Error("a cache built from different text is reported as valid because it is newer")
	}

	// Restoring the original text makes the same cache entry current again,
	// whatever the timestamps say.
	if err := os.WriteFile(sourcePath, original, 0644); err != nil {
		t.Fatalf("cannot restore the source: %v", err)
	}
	if !IsCacheValid(sourcePath, cachePath) {
		t.Error("a cache built from this exact source is reported as stale")
	}
}

// TestCacheRejectsAForeignStamp covers the other half: nothing recorded which
// build wrote a cache file, so an entry written by an older compiler was
// decoded by a newer one whose node numbering may have changed since.
func TestCacheRejectsAForeignStamp(t *testing.T) {
	tmpDir := t.TempDir()
	sourcePath := filepath.Join(tmpDir, "source.abc")
	cachePath := filepath.Join(tmpDir, CacheDir, "test.101")

	if err := os.WriteFile(sourcePath, []byte(`Print "hello".`), 0644); err != nil {
		t.Fatalf("cannot write the source: %v", err)
	}
	if err := WriteBytecodeCache(cachePath, CacheStamp([]byte("something else")), []byte("payload")); err != nil {
		t.Fatalf("cannot write the cache: %v", err)
	}
	if IsCacheValid(sourcePath, cachePath) {
		t.Error("a cache stamped for other source is reported as valid")
	}

	// A file that is not a cache envelope at all is rejected rather than
	// read as bytecode.
	if err := os.WriteFile(cachePath, []byte{0x10, 0x1E, 0x4E, 0x47, 0x01}, 0644); err != nil {
		t.Fatalf("cannot write the cache: %v", err)
	}
	if _, _, err := ReadBytecodeCache(cachePath); err == nil {
		t.Error("a compiled .101 file was accepted as a cache envelope")
	}
}

func TestWriteAndReadBytecodeCache(t *testing.T) {
	tmpDir := t.TempDir()
	cachePath := filepath.Join(tmpDir, CacheDir, "test.101")
	payload := []byte{0x10, 0x1E, 0x4E, 0x47, 0x01, 0x02, 0x03}
	stamp := CacheStamp([]byte(`Print "hello".`))

	if err := WriteBytecodeCache(cachePath, stamp, payload); err != nil {
		t.Fatalf("WriteBytecodeCache failed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(tmpDir, CacheDir)); os.IsNotExist(err) {
		t.Error("the cache directory was not created")
	}

	readPayload, readStamp, err := ReadBytecodeCache(cachePath)
	if err != nil {
		t.Fatalf("ReadBytecodeCache failed: %v", err)
	}
	if readStamp != stamp {
		t.Errorf("the stamp came back as %016x, want %016x", readStamp, stamp)
	}
	if !bytes.Equal(readPayload, payload) {
		t.Errorf("the payload came back as %x, want %x", readPayload, payload)
	}
}

func TestLoadCachedOrParse(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "engcache_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Change to temp directory for testing
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	sourcePath := "test.abc"
	if err := os.WriteFile(sourcePath, []byte("Print \"test\"."), 0644); err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	parseCallCount := 0
	parseFunc := func(path string) (*ast.Program, error) {
		parseCallCount++
		return &ast.Program{
			Statements: []ast.Statement{
				&ast.OutputStatement{
					Values:  []ast.Expression{&ast.StringLiteral{Value: "test"}},
					Newline: true,
				},
			},
		}, nil
	}

	// First call - should parse and cache
	program1, fromCache1, err := LoadCachedOrParse(sourcePath, parseFunc)
	if err != nil {
		t.Fatalf("LoadCachedOrParse failed: %v", err)
	}
	if fromCache1 {
		t.Error("First load should not be from cache")
	}
	if program1 == nil {
		t.Fatal("Program should not be nil")
	}
	if parseCallCount != 1 {
		t.Errorf("Parse function should be called once, got %d", parseCallCount)
	}

	// Second call - should use cache
	program2, fromCache2, err := LoadCachedOrParse(sourcePath, parseFunc)
	if err != nil {
		t.Fatalf("LoadCachedOrParse failed: %v", err)
	}
	if !fromCache2 {
		t.Error("Second load should be from cache")
	}
	if program2 == nil {
		t.Fatal("Program should not be nil")
	}
	if parseCallCount != 1 {
		t.Errorf("Parse function should still be called once, got %d", parseCallCount)
	}

	// The source changes, so the cache entry no longer describes it.
	if err := os.WriteFile(sourcePath, []byte("Print \"updated\"."), 0644); err != nil {
		t.Fatalf("Failed to update source file: %v", err)
	}

	// Third call - should re-parse due to stale cache
	program3, fromCache3, err := LoadCachedOrParse(sourcePath, parseFunc)
	if err != nil {
		t.Fatalf("LoadCachedOrParse failed: %v", err)
	}
	if fromCache3 {
		t.Error("Third load should not be from cache (stale)")
	}
	if program3 == nil {
		t.Fatal("Program should not be nil")
	}
	if parseCallCount != 2 {
		t.Errorf("Parse function should be called twice, got %d", parseCallCount)
	}
}
