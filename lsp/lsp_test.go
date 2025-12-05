package lsp

import (
	"testing"
)

func TestDocument(t *testing.T) {
	t.Run("NewDocument", func(t *testing.T) {
		doc := NewDocument("file:///test.abc", "english", 1, "Declare x to be 5.\nPrint x.")
		if doc.URI != "file:///test.abc" {
			t.Errorf("Expected URI file:///test.abc, got %s", doc.URI)
		}
		if doc.Version != 1 {
			t.Errorf("Expected version 1, got %d", doc.Version)
		}
		if len(doc.Lines) != 2 {
			t.Errorf("Expected 2 lines, got %d", len(doc.Lines))
		}
	})

	t.Run("GetLine", func(t *testing.T) {
		doc := NewDocument("test", "english", 1, "Line 1\nLine 2\nLine 3")
		if doc.GetLine(0) != "Line 1" {
			t.Errorf("Expected 'Line 1', got '%s'", doc.GetLine(0))
		}
		if doc.GetLine(1) != "Line 2" {
			t.Errorf("Expected 'Line 2', got '%s'", doc.GetLine(1))
		}
		if doc.GetLine(99) != "" {
			t.Errorf("Expected empty string for invalid line, got '%s'", doc.GetLine(99))
		}
	})

	t.Run("PositionToOffset", func(t *testing.T) {
		doc := NewDocument("test", "english", 1, "Hello\nWorld")
		// "Hello" is 5 chars, newline is 1, so "World" starts at offset 6
		offset := doc.PositionToOffset(Position{Line: 1, Character: 0})
		if offset != 6 {
			t.Errorf("Expected offset 6, got %d", offset)
		}

		offset = doc.PositionToOffset(Position{Line: 0, Character: 3})
		if offset != 3 {
			t.Errorf("Expected offset 3, got %d", offset)
		}
	})

	t.Run("OffsetToPosition", func(t *testing.T) {
		doc := NewDocument("test", "english", 1, "Hello\nWorld")
		pos := doc.OffsetToPosition(6)
		if pos.Line != 1 || pos.Character != 0 {
			t.Errorf("Expected (1, 0), got (%d, %d)", pos.Line, pos.Character)
		}

		pos = doc.OffsetToPosition(3)
		if pos.Line != 0 || pos.Character != 3 {
			t.Errorf("Expected (0, 3), got (%d, %d)", pos.Line, pos.Character)
		}
	})

	t.Run("GetWordAtPosition", func(t *testing.T) {
		doc := NewDocument("test", "english", 1, "Declare myVar to be 5.")
		word, _ := doc.GetWordAtPosition(Position{Line: 0, Character: 10})
		if word != "myVar" {
			t.Errorf("Expected 'myVar', got '%s'", word)
		}

		word, _ = doc.GetWordAtPosition(Position{Line: 0, Character: 0})
		if word != "Declare" {
			t.Errorf("Expected 'Declare', got '%s'", word)
		}
	})

	t.Run("ApplyContentChanges_Full", func(t *testing.T) {
		doc := NewDocument("test", "english", 1, "Original content")
		doc.ApplyContentChanges([]TextDocumentContentChangeEvent{
			{Text: "New content"},
		})
		if doc.Content != "New content" {
			t.Errorf("Expected 'New content', got '%s'", doc.Content)
		}
	})

	t.Run("ApplyContentChanges_Incremental", func(t *testing.T) {
		doc := NewDocument("test", "english", 1, "Hello World")
		doc.ApplyContentChanges([]TextDocumentContentChangeEvent{
			{
				Range: &Range{
					Start: Position{Line: 0, Character: 6},
					End:   Position{Line: 0, Character: 11},
				},
				Text: "Go",
			},
		})
		if doc.Content != "Hello Go" {
			t.Errorf("Expected 'Hello Go', got '%s'", doc.Content)
		}
	})
}

func TestDocumentManager(t *testing.T) {
	t.Run("Open_Get_Close", func(t *testing.T) {
		dm := NewDocumentManager()

		doc := dm.Open("file:///test.abc", "english", 1, "content")
		if doc == nil {
			t.Fatal("Expected document, got nil")
		}

		retrieved, err := dm.Get("file:///test.abc")
		if err != nil {
			t.Fatalf("Error getting document: %v", err)
		}
		if retrieved.Content != "content" {
			t.Errorf("Expected 'content', got '%s'", retrieved.Content)
		}

		dm.Close("file:///test.abc")
		_, err = dm.Get("file:///test.abc")
		if err == nil {
			t.Error("Expected error after closing document")
		}
	})

	t.Run("Update", func(t *testing.T) {
		dm := NewDocumentManager()
		dm.Open("file:///test.abc", "english", 1, "original")

		err := dm.Update("file:///test.abc", 2, []TextDocumentContentChangeEvent{
			{Text: "updated"},
		})
		if err != nil {
			t.Fatalf("Error updating document: %v", err)
		}

		doc, _ := dm.Get("file:///test.abc")
		if doc.Content != "updated" {
			t.Errorf("Expected 'updated', got '%s'", doc.Content)
		}
		if doc.Version != 2 {
			t.Errorf("Expected version 2, got %d", doc.Version)
		}
	})
}

func TestAnalyzer(t *testing.T) {
	analyzer := NewAnalyzer()

	t.Run("Analyze_ValidProgram", func(t *testing.T) {
		doc := NewDocument("test", "english", 1, `Declare x to be 5.
Declare y to be 10.
Print x + y.`)

		result := analyzer.Analyze(doc)
		if result == nil {
			t.Fatal("Expected analysis result, got nil")
		}
		if len(result.Diagnostics) > 0 {
			t.Errorf("Expected no diagnostics, got %d: %v", len(result.Diagnostics), result.Diagnostics)
		}
		if len(result.Variables) != 2 {
			t.Errorf("Expected 2 variables, got %d", len(result.Variables))
		}
	})

	t.Run("Analyze_ParseError", func(t *testing.T) {
		doc := NewDocument("test", "english", 1, "Declare x to be")

		result := analyzer.Analyze(doc)
		if len(result.Diagnostics) == 0 {
			t.Error("Expected diagnostics for parse error")
		}
	})

	t.Run("Analyze_Function", func(t *testing.T) {
		doc := NewDocument("test", "english", 1, `Declare function greet that takes name and does the following:
    Print name.
Thats it.`)

		result := analyzer.Analyze(doc)
		if len(result.Functions) != 1 {
			t.Errorf("Expected 1 function, got %d", len(result.Functions))
		}
		if _, ok := result.Functions["greet"]; !ok {
			t.Error("Expected function 'greet' to be found")
		}
	})

	t.Run("GetCompletions", func(t *testing.T) {
		doc := NewDocument("test", "english", 1, `Declare myVariable to be 5.
Declare myConstant to always be 10.
`)
		result := analyzer.Analyze(doc)

		completions := analyzer.GetCompletions(doc, Position{Line: 2, Character: 2}, result)
		if len(completions) == 0 {
			t.Error("Expected completions")
		}

		// Should include keywords
		hasKeyword := false
		for _, c := range completions {
			if c.Kind == CompletionItemKindKeyword {
				hasKeyword = true
				break
			}
		}
		if !hasKeyword {
			t.Error("Expected keyword completions")
		}

		// Should include variables
		hasVariable := false
		for _, c := range completions {
			if c.Label == "myVariable" {
				hasVariable = true
				break
			}
		}
		if !hasVariable {
			t.Error("Expected variable 'myVariable' in completions")
		}
	})

	t.Run("GetHover_Variable", func(t *testing.T) {
		doc := NewDocument("test", "english", 1, "Declare x to be 5.")
		result := analyzer.Analyze(doc)

		hover := analyzer.GetHover(doc, Position{Line: 0, Character: 8}, result)
		if hover == nil {
			t.Fatal("Expected hover information")
		}
		if hover.Contents.Value == "" {
			t.Error("Expected hover content")
		}
	})

	t.Run("GetHover_Keyword", func(t *testing.T) {
		doc := NewDocument("test", "english", 1, "Declare x to be 5.")
		result := analyzer.Analyze(doc)

		hover := analyzer.GetHover(doc, Position{Line: 0, Character: 0}, result)
		if hover == nil {
			t.Fatal("Expected hover information for keyword")
		}
	})

	t.Run("GetDefinition", func(t *testing.T) {
		doc := NewDocument("test", "english", 1, `Declare myVar to be 5.
Print myVar.`)
		result := analyzer.Analyze(doc)

		def := analyzer.GetDefinition(doc, Position{Line: 1, Character: 7}, result)
		if def == nil {
			t.Fatal("Expected definition")
		}
		// Definition should be on line 0
		if def.Range.Start.Line != 0 {
			t.Errorf("Expected definition on line 0, got line %d", def.Range.Start.Line)
		}
	})

	t.Run("GetReferences", func(t *testing.T) {
		doc := NewDocument("test", "english", 1, `Declare x to be 5.
Set x to be 10.
Print x.`)
		result := analyzer.Analyze(doc)

		refs := analyzer.GetReferences(doc, Position{Line: 0, Character: 8}, result, true)
		if len(refs) < 2 {
			t.Errorf("Expected at least 2 references, got %d", len(refs))
		}
	})

	t.Run("GetDocumentSymbols", func(t *testing.T) {
		doc := NewDocument("test", "english", 1, `Declare x to be 5.
Declare function add that takes a and b and does the following:
    Return a + b.
Thats it.`)
		result := analyzer.Analyze(doc)

		symbols := analyzer.GetDocumentSymbols(result)
		if len(symbols) != 2 {
			t.Errorf("Expected 2 symbols, got %d", len(symbols))
		}

		hasFunction := false
		hasVariable := false
		for _, s := range symbols {
			if s.Kind == SymbolKindFunction {
				hasFunction = true
			}
			if s.Kind == SymbolKindVariable {
				hasVariable = true
			}
		}
		if !hasFunction {
			t.Error("Expected function symbol")
		}
		if !hasVariable {
			t.Error("Expected variable symbol")
		}
	})
}

func TestServerCapabilities(t *testing.T) {
	t.Run("InitializeResult", func(t *testing.T) {
		// Verify that server capabilities are properly set
		result := &InitializeResult{
			Capabilities: ServerCapabilities{
				TextDocumentSync: &TextDocumentSyncOptions{
					OpenClose: true,
					Change:    TextDocumentSyncKindFull,
				},
				CompletionProvider: &CompletionOptions{
					TriggerCharacters: []string{".", " "},
				},
				HoverProvider:          true,
				DefinitionProvider:     true,
				ReferencesProvider:     true,
				DocumentSymbolProvider: true,
			},
			ServerInfo: &ServerInfo{
				Name:    "English Language Server",
				Version: "1.0.0",
			},
		}

		if result.Capabilities.TextDocumentSync == nil {
			t.Error("TextDocumentSync should not be nil")
		}
		if result.Capabilities.CompletionProvider == nil {
			t.Error("CompletionProvider should not be nil")
		}
		if result.Capabilities.HoverProvider != true {
			t.Error("HoverProvider should be true")
		}
	})
}

func TestDiagnosticSeverity(t *testing.T) {
	if DiagnosticSeverityError != 1 {
		t.Errorf("Expected DiagnosticSeverityError to be 1, got %d", DiagnosticSeverityError)
	}
	if DiagnosticSeverityWarning != 2 {
		t.Errorf("Expected DiagnosticSeverityWarning to be 2, got %d", DiagnosticSeverityWarning)
	}
}

func TestCompletionItemKind(t *testing.T) {
	if CompletionItemKindFunction != 3 {
		t.Errorf("Expected CompletionItemKindFunction to be 3, got %d", CompletionItemKindFunction)
	}
	if CompletionItemKindVariable != 6 {
		t.Errorf("Expected CompletionItemKindVariable to be 6, got %d", CompletionItemKindVariable)
	}
	if CompletionItemKindKeyword != 14 {
		t.Errorf("Expected CompletionItemKindKeyword to be 14, got %d", CompletionItemKindKeyword)
	}
}

func TestSymbolKind(t *testing.T) {
	if SymbolKindFunction != 12 {
		t.Errorf("Expected SymbolKindFunction to be 12, got %d", SymbolKindFunction)
	}
	if SymbolKindVariable != 13 {
		t.Errorf("Expected SymbolKindVariable to be 13, got %d", SymbolKindVariable)
	}
	if SymbolKindConstant != 14 {
		t.Errorf("Expected SymbolKindConstant to be 14, got %d", SymbolKindConstant)
	}
}

func TestTextDocumentSyncKind(t *testing.T) {
	if TextDocumentSyncKindNone != 0 {
		t.Errorf("Expected TextDocumentSyncKindNone to be 0, got %d", TextDocumentSyncKindNone)
	}
	if TextDocumentSyncKindFull != 1 {
		t.Errorf("Expected TextDocumentSyncKindFull to be 1, got %d", TextDocumentSyncKindFull)
	}
	if TextDocumentSyncKindIncremental != 2 {
		t.Errorf("Expected TextDocumentSyncKindIncremental to be 2, got %d", TextDocumentSyncKindIncremental)
	}
}
