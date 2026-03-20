package help

import (
	"sort"
	"strings"
)

// HelpEntry represents a single help topic with searchable metadata.
type HelpEntry struct {
	Name        string   // Primary name (e.g., "print", "if statement")
	Description string   // Short description
	Category    string   // "command", "keyword", "function", "operator", "concept"
	LongDesc    string   // Detailed explanation
	Examples    []string // Code examples
	Keywords    []string // Additional search terms
	Aliases     []string // Alternative names
	SeeAlso     []string // Related topics
}

// SearchResult represents a help entry with its relevance score.
type SearchResult struct {
	Entry *HelpEntry
	Score int // Higher is better
}

// Registry manages all help entries and provides search functionality.
type Registry struct {
	entries map[string]*HelpEntry
}

// NewRegistry creates a new help registry with default content.
func NewRegistry() *Registry {
	r := &Registry{
		entries: make(map[string]*HelpEntry),
	}
	r.loadDefaultEntries()
	return r
}

// Register adds a new help entry to the registry.
func (r *Registry) Register(entry *HelpEntry) {
	r.entries[strings.ToLower(entry.Name)] = entry
}

// Lookup retrieves a help entry by exact name (case-insensitive).
func (r *Registry) Lookup(name string) *HelpEntry {
	return r.entries[strings.ToLower(name)]
}

// Search performs fuzzy search across all help entries and returns ranked results.
func (r *Registry) Search(query string) []SearchResult {
	if query == "" {
		return nil
	}

	query = strings.ToLower(query)
	var results []SearchResult

	for _, entry := range r.entries {
		score := r.calculateRelevance(query, entry)
		if score > 0 {
			results = append(results, SearchResult{
				Entry: entry,
				Score: score,
			})
		}
	}

	// Sort by score (descending)
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	return results
}

// calculateRelevance computes a relevance score for a query against a help entry.
func (r *Registry) calculateRelevance(query string, entry *HelpEntry) int {
	name := strings.ToLower(entry.Name)
	desc := strings.ToLower(entry.Description)

	// Exact name match (highest priority)
	if name == query {
		return 1000
	}

	// Exact prefix match
	if strings.HasPrefix(name, query) {
		return 900
	}

	// Name contains query
	if strings.Contains(name, query) {
		return 800
	}

	// Alias exact match
	for _, alias := range entry.Aliases {
		if strings.ToLower(alias) == query {
			return 750
		}
	}

	// Alias prefix match
	for _, alias := range entry.Aliases {
		if strings.HasPrefix(strings.ToLower(alias), query) {
			return 700
		}
	}

	// Description contains query
	if strings.Contains(desc, query) {
		return 600
	}

	// Keyword exact match
	for _, keyword := range entry.Keywords {
		if strings.ToLower(keyword) == query {
			return 500
		}
	}

	// Keyword contains query
	for _, keyword := range entry.Keywords {
		if strings.Contains(strings.ToLower(keyword), query) {
			return 400
		}
	}

	// Fuzzy matching using Levenshtein distance
	distance := levenshteinDistance(query, name)
	maxLen := len(query)
	if len(name) > maxLen {
		maxLen = len(name)
	}

	// Only consider if distance is reasonable (within 40% of max length)
	if distance <= maxLen*4/10 {
		// Score based on similarity (closer = higher score)
		similarity := 1.0 - float64(distance)/float64(maxLen)
		return int(similarity * 300)
	}

	return 0
}

// AllCategories returns a list of all unique categories in the registry.
func (r *Registry) AllCategories() []string {
	categoryMap := make(map[string]bool)
	for _, entry := range r.entries {
		categoryMap[entry.Category] = true
	}

	var categories []string
	for cat := range categoryMap {
		categories = append(categories, cat)
	}
	sort.Strings(categories)
	return categories
}

// EntriesByCategory returns all help entries in a specific category.
func (r *Registry) EntriesByCategory(category string) []*HelpEntry {
	var entries []*HelpEntry
	for _, entry := range r.entries {
		if strings.EqualFold(entry.Category, category) {
			entries = append(entries, entry)
		}
	}

	// Sort by name
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name < entries[j].Name
	})

	return entries
}

// AllEntries returns all help entries sorted by name.
func (r *Registry) AllEntries() []*HelpEntry {
	var entries []*HelpEntry
	for _, entry := range r.entries {
		entries = append(entries, entry)
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name < entries[j].Name
	})

	return entries
}
