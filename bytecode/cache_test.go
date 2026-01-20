package bytecode

import (
	"english/ast"
	"os"
	"path/filepath"
	"testing"
	"time"
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

func TestIsCacheValid(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "engcache_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a source file
	sourcePath := filepath.Join(tmpDir, "source.abc")
	if err := os.WriteFile(sourcePath, []byte("Print \"hello\"."), 0644); err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	// Create cache directory
	cacheDir := filepath.Join(tmpDir, CacheDir)
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		t.Fatalf("Failed to create cache dir: %v", err)
	}

	cachePath := filepath.Join(cacheDir, "test.101")

	// Test 1: Cache doesn't exist
	if IsCacheValid(sourcePath, cachePath) {
		t.Error("Cache should not be valid when it doesn't exist")
	}

	// Create cache file (newer than source)
	time.Sleep(10 * time.Millisecond)
	if err := os.WriteFile(cachePath, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create cache file: %v", err)
	}

	// Test 2: Cache is newer than source
	if !IsCacheValid(sourcePath, cachePath) {
		t.Error("Cache should be valid when it's newer than source")
	}

	// Update source file to be newer than cache
	time.Sleep(10 * time.Millisecond)
	if err := os.WriteFile(sourcePath, []byte("Print \"updated\"."), 0644); err != nil {
		t.Fatalf("Failed to update source file: %v", err)
	}

	// Test 3: Source is newer than cache
	if IsCacheValid(sourcePath, cachePath) {
		t.Error("Cache should not be valid when source is newer")
	}
}

func TestWriteAndReadBytecodeCache(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "engcache_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cachePath := filepath.Join(tmpDir, CacheDir, "test.101")
	testData := []byte{0x10, 0x1E, 0x4E, 0x47, 0x01, 0x02, 0x03}

	// Test writing cache
	if err := WriteBytecodeCache(cachePath, testData); err != nil {
		t.Fatalf("WriteBytecodeCache failed: %v", err)
	}

	// Verify cache directory was created
	if _, err := os.Stat(filepath.Join(tmpDir, CacheDir)); os.IsNotExist(err) {
		t.Error("Cache directory was not created")
	}

	// Test reading cache
	readData, err := ReadBytecodeCache(cachePath)
	if err != nil {
		t.Fatalf("ReadBytecodeCache failed: %v", err)
	}

	// Verify data matches
	if len(readData) != len(testData) {
		t.Errorf("Data length mismatch: got %d, want %d", len(readData), len(testData))
	}
	for i := range testData {
		if readData[i] != testData[i] {
			t.Errorf("Data mismatch at index %d: got %x, want %x", i, readData[i], testData[i])
		}
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

	// Update source file to invalidate cache
	time.Sleep(10 * time.Millisecond)
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
