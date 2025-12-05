package lsp

import (
	"fmt"
	"strings"
	"sync"
)

// Document represents an open text document
type Document struct {
	URI        string
	LanguageID string
	Version    int
	Content    string
	Lines      []string
}

// NewDocument creates a new document
func NewDocument(uri, languageID string, version int, content string) *Document {
	doc := &Document{
		URI:        uri,
		LanguageID: languageID,
		Version:    version,
		Content:    content,
	}
	doc.updateLines()
	return doc
}

// updateLines splits the content into lines
func (d *Document) updateLines() {
	d.Lines = strings.Split(d.Content, "\n")
}

// GetLine returns the content of a specific line (0-indexed)
func (d *Document) GetLine(line int) string {
	if line < 0 || line >= len(d.Lines) {
		return ""
	}
	return d.Lines[line]
}

// GetText returns text in a range
func (d *Document) GetText(r Range) string {
	if r.Start.Line == r.End.Line {
		line := d.GetLine(r.Start.Line)
		if r.Start.Character >= len(line) {
			return ""
		}
		end := r.End.Character
		if end > len(line) {
			end = len(line)
		}
		return line[r.Start.Character:end]
	}

	var result strings.Builder
	for lineNum := r.Start.Line; lineNum <= r.End.Line; lineNum++ {
		line := d.GetLine(lineNum)
		if lineNum == r.Start.Line {
			if r.Start.Character < len(line) {
				result.WriteString(line[r.Start.Character:])
			}
			result.WriteString("\n")
		} else if lineNum == r.End.Line {
			end := r.End.Character
			if end > len(line) {
				end = len(line)
			}
			result.WriteString(line[:end])
		} else {
			result.WriteString(line)
			result.WriteString("\n")
		}
	}
	return result.String()
}

// PositionToOffset converts a position to a byte offset
func (d *Document) PositionToOffset(pos Position) int {
	offset := 0
	for i := 0; i < pos.Line && i < len(d.Lines); i++ {
		offset += len(d.Lines[i]) + 1 // +1 for newline
	}
	if pos.Line < len(d.Lines) {
		line := d.Lines[pos.Line]
		if pos.Character <= len(line) {
			offset += pos.Character
		} else {
			offset += len(line)
		}
	}
	return offset
}

// OffsetToPosition converts a byte offset to a position
func (d *Document) OffsetToPosition(offset int) Position {
	currentOffset := 0
	for lineNum, line := range d.Lines {
		lineLen := len(line) + 1 // +1 for newline
		if currentOffset+lineLen > offset {
			return Position{
				Line:      lineNum,
				Character: offset - currentOffset,
			}
		}
		currentOffset += lineLen
	}
	// Return end of document
	if len(d.Lines) == 0 {
		return Position{Line: 0, Character: 0}
	}
	lastLine := len(d.Lines) - 1
	return Position{
		Line:      lastLine,
		Character: len(d.Lines[lastLine]),
	}
}

// ApplyContentChanges applies content changes to the document
func (d *Document) ApplyContentChanges(changes []TextDocumentContentChangeEvent) {
	for _, change := range changes {
		if change.Range == nil {
			// Full document update
			d.Content = change.Text
		} else {
			// Incremental update
			startOffset := d.PositionToOffset(change.Range.Start)
			endOffset := d.PositionToOffset(change.Range.End)
			d.Content = d.Content[:startOffset] + change.Text + d.Content[endOffset:]
		}
		d.updateLines()
	}
}

// GetWordAtPosition returns the word at the given position
func (d *Document) GetWordAtPosition(pos Position) (string, Range) {
	line := d.GetLine(pos.Line)
	if pos.Character > len(line) {
		return "", Range{}
	}

	// Find word boundaries
	start := pos.Character
	end := pos.Character

	// Move start backwards to find the beginning of the word
	for start > 0 && isWordChar(line[start-1]) {
		start--
	}

	// Move end forwards to find the end of the word
	for end < len(line) && isWordChar(line[end]) {
		end++
	}

	if start == end {
		return "", Range{}
	}

	return line[start:end], Range{
		Start: Position{Line: pos.Line, Character: start},
		End:   Position{Line: pos.Line, Character: end},
	}
}

// isWordChar returns true if c is a valid word character
func isWordChar(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_'
}

// DocumentManager manages open documents
type DocumentManager struct {
	mu        sync.RWMutex
	documents map[string]*Document
}

// NewDocumentManager creates a new document manager
func NewDocumentManager() *DocumentManager {
	return &DocumentManager{
		documents: make(map[string]*Document),
	}
}

// Open opens a new document
func (dm *DocumentManager) Open(uri, languageID string, version int, content string) *Document {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	doc := NewDocument(uri, languageID, version, content)
	dm.documents[uri] = doc
	return doc
}

// Close closes a document
func (dm *DocumentManager) Close(uri string) {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	delete(dm.documents, uri)
}

// Get returns a document by URI
func (dm *DocumentManager) Get(uri string) (*Document, error) {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	doc, ok := dm.documents[uri]
	if !ok {
		return nil, fmt.Errorf("document not found: %s", uri)
	}
	return doc, nil
}

// Update updates a document with changes
func (dm *DocumentManager) Update(uri string, version int, changes []TextDocumentContentChangeEvent) error {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	doc, ok := dm.documents[uri]
	if !ok {
		return fmt.Errorf("document not found: %s", uri)
	}

	doc.Version = version
	doc.ApplyContentChanges(changes)
	return nil
}

// All returns all open documents
func (dm *DocumentManager) All() []*Document {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	docs := make([]*Document, 0, len(dm.documents))
	for _, doc := range dm.documents {
		docs = append(docs, doc)
	}
	return docs
}
