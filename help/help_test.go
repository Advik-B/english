package help

import (
	"testing"
)

func TestLevenshteinDistance(t *testing.T) {
	tests := []struct {
		s1       string
		s2       string
		expected int
	}{
		{"", "", 0},
		{"hello", "", 5},
		{"", "world", 5},
		{"hello", "hello", 0},
		{"hello", "hallo", 1},
		{"kitten", "sitting", 3},
		{"saturday", "sunday", 3},
		{"print", "printf", 1},
		{"if", "for", 3},
	}

	for _, tt := range tests {
		result := levenshteinDistance(tt.s1, tt.s2)
		if result != tt.expected {
			t.Errorf("levenshteinDistance(%q, %q) = %d; want %d",
				tt.s1, tt.s2, result, tt.expected)
		}
	}
}

func TestRegistryLookup(t *testing.T) {
	r := NewRegistry()

	// Test exact lookup
	entry := r.Lookup("print")
	if entry == nil {
		t.Fatal("Expected to find 'print' entry")
	}
	if entry.Name != "print" {
		t.Errorf("Expected Name to be 'print', got %q", entry.Name)
	}
	if entry.Category != "function" {
		t.Errorf("Expected Category to be 'function', got %q", entry.Category)
	}

	// Test case-insensitive lookup
	entry = r.Lookup("PRINT")
	if entry == nil {
		t.Fatal("Expected case-insensitive lookup to find 'PRINT'")
	}

	// Test non-existent entry
	entry = r.Lookup("nonexistent")
	if entry != nil {
		t.Error("Expected nil for non-existent entry")
	}
}

func TestRegistrySearch(t *testing.T) {
	r := NewRegistry()

	tests := []struct {
		query          string
		expectResults  bool
		expectFirstName string
	}{
		{"print", true, "print"},
		{"PRINT", true, "print"},
		{"loop", true, ""},  // Should find multiple loop-related topics
		{"if", true, "if"},
		{"prnt", true, "print"}, // Fuzzy match
		{"nonexistent123", false, ""},
	}

	for _, tt := range tests {
		results := r.Search(tt.query)
		if tt.expectResults && len(results) == 0 {
			t.Errorf("Search(%q) returned no results, expected some", tt.query)
			continue
		}
		if !tt.expectResults && len(results) > 0 {
			t.Errorf("Search(%q) returned %d results, expected none", tt.query, len(results))
			continue
		}
		if tt.expectResults && tt.expectFirstName != "" {
			if results[0].Entry.Name != tt.expectFirstName {
				t.Errorf("Search(%q) first result is %q, expected %q",
					tt.query, results[0].Entry.Name, tt.expectFirstName)
			}
		}
	}
}

func TestSearchScoring(t *testing.T) {
	r := NewRegistry()

	// Search for "print" - should get exact match with high score
	results := r.Search("print")
	if len(results) == 0 {
		t.Fatal("Expected results for 'print'")
	}
	if results[0].Entry.Name != "print" {
		t.Errorf("Expected first result to be 'print', got %q", results[0].Entry.Name)
	}
	if results[0].Score < 900 {
		t.Errorf("Expected high score for exact match, got %d", results[0].Score)
	}

	// Search with fuzzy match
	results = r.Search("prnt")
	if len(results) == 0 {
		t.Fatal("Expected fuzzy match results for 'prnt'")
	}
}

func TestAllCategories(t *testing.T) {
	r := NewRegistry()
	categories := r.AllCategories()

	if len(categories) == 0 {
		t.Error("Expected at least one category")
	}

	// Check that categories are sorted
	for i := 1; i < len(categories); i++ {
		if categories[i-1] > categories[i] {
			t.Errorf("Categories not sorted: %q > %q", categories[i-1], categories[i])
		}
	}

	// Check for expected categories
	expectedCategories := []string{"command", "keyword", "function", "type", "operator", "concept"}
	for _, expected := range expectedCategories {
		found := false
		for _, cat := range categories {
			if cat == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected category %q not found", expected)
		}
	}
}

func TestEntriesByCategory(t *testing.T) {
	r := NewRegistry()

	// Test function category
	entries := r.EntriesByCategory("function")
	if len(entries) == 0 {
		t.Error("Expected at least one function entry")
	}

	// Check that all entries have the correct category
	for _, entry := range entries {
		if entry.Category != "function" {
			t.Errorf("Entry %q has category %q, expected 'function'", entry.Name, entry.Category)
		}
	}

	// Check that entries are sorted by name
	for i := 1; i < len(entries); i++ {
		if entries[i-1].Name > entries[i].Name {
			t.Errorf("Entries not sorted: %q > %q", entries[i-1].Name, entries[i].Name)
		}
	}
}

func TestAllEntries(t *testing.T) {
	r := NewRegistry()
	entries := r.AllEntries()

	if len(entries) == 0 {
		t.Error("Expected at least one entry")
	}

	// Check that entries are sorted by name
	for i := 1; i < len(entries); i++ {
		if entries[i-1].Name > entries[i].Name {
			t.Errorf("Entries not sorted: %q > %q", entries[i-1].Name, entries[i].Name)
		}
	}
}

func TestSearchWithAliases(t *testing.T) {
	r := NewRegistry()

	// "quit" is an alias for "exit"
	results := r.Search("quit")
	if len(results) == 0 {
		t.Fatal("Expected results for alias 'quit'")
	}

	// Should find the exit entry
	found := false
	for _, result := range results {
		if result.Entry.Name == "exit" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected to find 'exit' entry when searching for alias 'quit'")
	}
}

func TestSearchWithKeywords(t *testing.T) {
	r := NewRegistry()

	// Search for a keyword that should match
	results := r.Search("variable")
	if len(results) == 0 {
		t.Fatal("Expected results for keyword 'variable'")
	}

	// Should find the declare entry
	found := false
	for _, result := range results {
		if result.Entry.Name == "declare" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected to find 'declare' entry when searching for keyword 'variable'")
	}
}

func TestHelpEntryStructure(t *testing.T) {
	r := NewRegistry()

	// Check that "print" entry has all expected fields
	entry := r.Lookup("print")
	if entry == nil {
		t.Fatal("Expected to find 'print' entry")
	}

	if entry.Name == "" {
		t.Error("Entry Name should not be empty")
	}
	if entry.Description == "" {
		t.Error("Entry Description should not be empty")
	}
	if entry.Category == "" {
		t.Error("Entry Category should not be empty")
	}
	if entry.LongDesc == "" {
		t.Error("Entry LongDesc should not be empty for 'print'")
	}
	if len(entry.Examples) == 0 {
		t.Error("Entry should have at least one example")
	}
}

func TestPrefixMatching(t *testing.T) {
	r := NewRegistry()

	// Test prefix matching
	results := r.Search("pri")
	if len(results) == 0 {
		t.Fatal("Expected results for prefix 'pri'")
	}

	// "print" should be the first result since it starts with "pri"
	if results[0].Entry.Name != "print" {
		t.Errorf("Expected 'print' as first result, got %q", results[0].Entry.Name)
	}
}

func TestEmptySearch(t *testing.T) {
	r := NewRegistry()

	results := r.Search("")
	if results != nil {
		t.Error("Expected nil results for empty search")
	}
}
