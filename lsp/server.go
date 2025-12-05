package lsp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
)

// Language-specific formatting constants
const (
	blockEndKeyword     = "thats it"
	elseKeyword         = "otherwise"
	blockStartSuffix    = ":"
	thenSuffix          = ", then"
)

// Server represents the LSP server
type Server struct {
	reader   *bufio.Reader
	writer   io.Writer
	logger   *log.Logger
	mu       sync.Mutex

	initialized bool
	shutdown    bool

	documents *DocumentManager
	analyzer  *Analyzer
	analyses  map[string]*AnalysisResult
	analysisMu sync.RWMutex

	// Callbacks for extensibility
	onInitialize  func(*InitializeParams) error
	onShutdown    func() error
	customMethods map[string]MethodHandler
}

// MethodHandler is a function that handles a custom method
type MethodHandler func(params json.RawMessage) (interface{}, error)

// ServerOption is a function that configures the server
type ServerOption func(*Server)

// WithLogger sets a custom logger
func WithLogger(logger *log.Logger) ServerOption {
	return func(s *Server) {
		s.logger = logger
	}
}

// WithReader sets a custom reader
func WithReader(reader io.Reader) ServerOption {
	return func(s *Server) {
		s.reader = bufio.NewReader(reader)
	}
}

// WithWriter sets a custom writer
func WithWriter(writer io.Writer) ServerOption {
	return func(s *Server) {
		s.writer = writer
	}
}

// WithInitializeCallback sets a callback for initialization
func WithInitializeCallback(callback func(*InitializeParams) error) ServerOption {
	return func(s *Server) {
		s.onInitialize = callback
	}
}

// WithShutdownCallback sets a callback for shutdown
func WithShutdownCallback(callback func() error) ServerOption {
	return func(s *Server) {
		s.onShutdown = callback
	}
}

// NewServer creates a new LSP server
func NewServer(opts ...ServerOption) *Server {
	s := &Server{
		reader:        bufio.NewReader(os.Stdin),
		writer:        os.Stdout,
		logger:        log.New(os.Stderr, "[LSP] ", log.LstdFlags),
		documents:     NewDocumentManager(),
		analyzer:      NewAnalyzer(),
		analyses:      make(map[string]*AnalysisResult),
		customMethods: make(map[string]MethodHandler),
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

// RegisterMethod registers a custom method handler
func (s *Server) RegisterMethod(method string, handler MethodHandler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.customMethods[method] = handler
}

// Run starts the LSP server main loop
func (s *Server) Run() error {
	s.logger.Println("Server starting...")

	for {
		msg, err := s.readMessage()
		if err != nil {
			if err == io.EOF {
				s.logger.Println("Connection closed")
				return nil
			}
			s.logger.Printf("Error reading message: %v", err)
			continue
		}

		s.handleMessage(msg)

		if s.shutdown {
			s.logger.Println("Server shutting down")
			return nil
		}
	}
}

// readMessage reads a single LSP message from the input
func (s *Server) readMessage() (json.RawMessage, error) {
	// Read headers
	var contentLength int
	for {
		line, err := s.reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		line = strings.TrimSpace(line)

		if line == "" {
			break // End of headers
		}

		if strings.HasPrefix(line, "Content-Length: ") {
			lengthStr := strings.TrimPrefix(line, "Content-Length: ")
			contentLength, err = strconv.Atoi(lengthStr)
			if err != nil {
				return nil, fmt.Errorf("invalid Content-Length: %v", err)
			}
		}
	}

	if contentLength == 0 {
		return nil, fmt.Errorf("missing Content-Length header")
	}

	// Read content
	content := make([]byte, contentLength)
	_, err := io.ReadFull(s.reader, content)
	if err != nil {
		return nil, err
	}

	return json.RawMessage(content), nil
}

// writeMessage writes an LSP message to the output
func (s *Server) writeMessage(msg interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	content, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	header := fmt.Sprintf("Content-Length: %d\r\n\r\n", len(content))
	_, err = s.writer.Write([]byte(header))
	if err != nil {
		return err
	}

	_, err = s.writer.Write(content)
	return err
}

// handleMessage dispatches a message to the appropriate handler
func (s *Server) handleMessage(msg json.RawMessage) {
	// Try to parse as a request (has id and method)
	var request RequestMessage
	if err := json.Unmarshal(msg, &request); err == nil && request.Method != "" {
		if request.ID != nil {
			// It's a request
			s.handleRequest(request)
		} else {
			// It's a notification (no ID)
			var notification NotificationMessage
			if err := json.Unmarshal(msg, &notification); err != nil {
				s.logger.Printf("Error parsing notification: %v", err)
				return
			}
			s.handleNotification(notification)
		}
		return
	}

	// Could be a response - log it
	s.logger.Printf("Received unknown message type")
}

// handleRequest handles LSP requests
func (s *Server) handleRequest(req RequestMessage) {
	s.logger.Printf("Request: %s", req.Method)

	var result interface{}
	var err error

	// Check for server state
	if !s.initialized && req.Method != "initialize" {
		s.sendError(req.ID, ServerNotInitialized, "Server not initialized")
		return
	}

	switch req.Method {
	case "initialize":
		result, err = s.handleInitialize(req.Params)
	case "shutdown":
		result, err = s.handleShutdown()
	case "textDocument/completion":
		result, err = s.handleCompletion(req.Params)
	case "textDocument/hover":
		result, err = s.handleHover(req.Params)
	case "textDocument/definition":
		result, err = s.handleDefinition(req.Params)
	case "textDocument/references":
		result, err = s.handleReferences(req.Params)
	case "textDocument/documentSymbol":
		result, err = s.handleDocumentSymbol(req.Params)
	case "textDocument/signatureHelp":
		result, err = s.handleSignatureHelp(req.Params)
	case "textDocument/formatting":
		result, err = s.handleFormatting(req.Params)
	case "textDocument/codeAction":
		result, err = s.handleCodeAction(req.Params)
	case "textDocument/documentHighlight":
		result, err = s.handleDocumentHighlight(req.Params)
	case "textDocument/foldingRange":
		result, err = s.handleFoldingRange(req.Params)
	default:
		// Check custom methods
		s.mu.Lock()
		handler, ok := s.customMethods[req.Method]
		s.mu.Unlock()
		if ok {
			result, err = handler(req.Params)
		} else {
			s.sendError(req.ID, MethodNotFound, fmt.Sprintf("Method not found: %s", req.Method))
			return
		}
	}

	if err != nil {
		s.sendError(req.ID, InternalError, err.Error())
		return
	}

	s.sendResult(req.ID, result)
}

// handleNotification handles LSP notifications
func (s *Server) handleNotification(notif NotificationMessage) {
	s.logger.Printf("Notification: %s", notif.Method)

	switch notif.Method {
	case "initialized":
		s.logger.Println("Client initialized")
	case "exit":
		s.shutdown = true
	case "textDocument/didOpen":
		s.handleDidOpen(notif.Params)
	case "textDocument/didClose":
		s.handleDidClose(notif.Params)
	case "textDocument/didChange":
		s.handleDidChange(notif.Params)
	case "textDocument/didSave":
		s.handleDidSave(notif.Params)
	}
}

// sendResult sends a successful response
func (s *Server) sendResult(id interface{}, result interface{}) {
	response := ResponseMessage{
		Message: Message{JSONRPC: "2.0"},
		ID:      id,
		Result:  result,
	}
	if err := s.writeMessage(response); err != nil {
		s.logger.Printf("Error sending response: %v", err)
	}
}

// sendError sends an error response
func (s *Server) sendError(id interface{}, code int, message string) {
	response := ResponseMessage{
		Message: Message{JSONRPC: "2.0"},
		ID:      id,
		Error: &Error{
			Code:    code,
			Message: message,
		},
	}
	if err := s.writeMessage(response); err != nil {
		s.logger.Printf("Error sending error response: %v", err)
	}
}

// sendNotification sends a notification to the client
func (s *Server) sendNotification(method string, params interface{}) {
	paramsJSON, err := json.Marshal(params)
	if err != nil {
		s.logger.Printf("Error marshaling notification params: %v", err)
		return
	}
	notif := NotificationMessage{
		Message: Message{JSONRPC: "2.0"},
		Method:  method,
		Params:  paramsJSON,
	}
	if err := s.writeMessage(notif); err != nil {
		s.logger.Printf("Error sending notification: %v", err)
	}
}

// publishDiagnostics publishes diagnostics for a document
func (s *Server) publishDiagnostics(uri string, diagnostics []Diagnostic) {
	s.sendNotification("textDocument/publishDiagnostics", PublishDiagnosticsParams{
		URI:         uri,
		Diagnostics: diagnostics,
	})
}

// Handler implementations

func (s *Server) handleInitialize(params json.RawMessage) (*InitializeResult, error) {
	var initParams InitializeParams
	if err := json.Unmarshal(params, &initParams); err != nil {
		return nil, fmt.Errorf("invalid initialize params: %v", err)
	}

	// Call user callback if set
	if s.onInitialize != nil {
		if err := s.onInitialize(&initParams); err != nil {
			return nil, err
		}
	}

	s.initialized = true

	return &InitializeResult{
		Capabilities: ServerCapabilities{
			TextDocumentSync: &TextDocumentSyncOptions{
				OpenClose: true,
				Change:    TextDocumentSyncKindFull,
				Save: &SaveOptions{
					IncludeText: true,
				},
			},
			CompletionProvider: &CompletionOptions{
				TriggerCharacters: []string{".", " "},
				ResolveProvider:   false,
			},
			HoverProvider: true,
			SignatureHelpProvider: &SignatureHelpOptions{
				TriggerCharacters:   []string{"(", ",", " "},
				RetriggerCharacters: []string{",", " "},
			},
			DefinitionProvider:        true,
			ReferencesProvider:        true,
			DocumentHighlightProvider: true,
			DocumentSymbolProvider:    true,
			CodeActionProvider:        true,
			DocumentFormattingProvider: true,
			FoldingRangeProvider:      true,
		},
		ServerInfo: &ServerInfo{
			Name:    "English Language Server",
			Version: "1.0.0",
		},
	}, nil
}

func (s *Server) handleShutdown() (interface{}, error) {
	if s.onShutdown != nil {
		if err := s.onShutdown(); err != nil {
			return nil, err
		}
	}
	s.shutdown = true
	return nil, nil
}

func (s *Server) handleDidOpen(params json.RawMessage) {
	var openParams DidOpenTextDocumentParams
	if err := json.Unmarshal(params, &openParams); err != nil {
		s.logger.Printf("Error parsing didOpen: %v", err)
		return
	}

	doc := s.documents.Open(
		openParams.TextDocument.URI,
		openParams.TextDocument.LanguageID,
		openParams.TextDocument.Version,
		openParams.TextDocument.Text,
	)

	// Analyze and publish diagnostics
	s.analyzeAndPublish(doc)
}

func (s *Server) handleDidClose(params json.RawMessage) {
	var closeParams DidCloseTextDocumentParams
	if err := json.Unmarshal(params, &closeParams); err != nil {
		s.logger.Printf("Error parsing didClose: %v", err)
		return
	}

	s.documents.Close(closeParams.TextDocument.URI)

	// Remove analysis
	s.analysisMu.Lock()
	delete(s.analyses, closeParams.TextDocument.URI)
	s.analysisMu.Unlock()

	// Clear diagnostics
	s.publishDiagnostics(closeParams.TextDocument.URI, []Diagnostic{})
}

func (s *Server) handleDidChange(params json.RawMessage) {
	var changeParams DidChangeTextDocumentParams
	if err := json.Unmarshal(params, &changeParams); err != nil {
		s.logger.Printf("Error parsing didChange: %v", err)
		return
	}

	err := s.documents.Update(
		changeParams.TextDocument.URI,
		changeParams.TextDocument.Version,
		changeParams.ContentChanges,
	)
	if err != nil {
		s.logger.Printf("Error updating document: %v", err)
		return
	}

	doc, err := s.documents.Get(changeParams.TextDocument.URI)
	if err != nil {
		return
	}

	// Re-analyze and publish diagnostics
	s.analyzeAndPublish(doc)
}

func (s *Server) handleDidSave(params json.RawMessage) {
	var saveParams DidSaveTextDocumentParams
	if err := json.Unmarshal(params, &saveParams); err != nil {
		s.logger.Printf("Error parsing didSave: %v", err)
		return
	}

	// If text is included, update the document
	if saveParams.Text != "" {
		doc, err := s.documents.Get(saveParams.TextDocument.URI)
		if err != nil {
			return
		}
		doc.Content = saveParams.Text
		doc.updateLines()
		s.analyzeAndPublish(doc)
	}
}

func (s *Server) analyzeAndPublish(doc *Document) {
	result := s.analyzer.Analyze(doc)

	s.analysisMu.Lock()
	s.analyses[doc.URI] = result
	s.analysisMu.Unlock()

	s.publishDiagnostics(doc.URI, result.Diagnostics)
}

func (s *Server) getAnalysis(uri string) *AnalysisResult {
	s.analysisMu.RLock()
	defer s.analysisMu.RUnlock()
	return s.analyses[uri]
}

func (s *Server) handleCompletion(params json.RawMessage) (interface{}, error) {
	var compParams CompletionParams
	if err := json.Unmarshal(params, &compParams); err != nil {
		return nil, err
	}

	doc, err := s.documents.Get(compParams.TextDocument.URI)
	if err != nil {
		return nil, err
	}

	result := s.getAnalysis(compParams.TextDocument.URI)
	if result == nil {
		result = s.analyzer.Analyze(doc)
	}

	items := s.analyzer.GetCompletions(doc, compParams.Position, result)

	return CompletionList{
		IsIncomplete: false,
		Items:        items,
	}, nil
}

func (s *Server) handleHover(params json.RawMessage) (interface{}, error) {
	var hoverParams HoverParams
	if err := json.Unmarshal(params, &hoverParams); err != nil {
		return nil, err
	}

	doc, err := s.documents.Get(hoverParams.TextDocument.URI)
	if err != nil {
		return nil, err
	}

	result := s.getAnalysis(hoverParams.TextDocument.URI)
	if result == nil {
		result = s.analyzer.Analyze(doc)
	}

	hover := s.analyzer.GetHover(doc, hoverParams.Position, result)
	return hover, nil
}

func (s *Server) handleDefinition(params json.RawMessage) (interface{}, error) {
	var defParams DefinitionParams
	if err := json.Unmarshal(params, &defParams); err != nil {
		return nil, err
	}

	doc, err := s.documents.Get(defParams.TextDocument.URI)
	if err != nil {
		return nil, err
	}

	result := s.getAnalysis(defParams.TextDocument.URI)
	if result == nil {
		result = s.analyzer.Analyze(doc)
	}

	location := s.analyzer.GetDefinition(doc, defParams.Position, result)
	return location, nil
}

func (s *Server) handleReferences(params json.RawMessage) (interface{}, error) {
	var refParams ReferenceParams
	if err := json.Unmarshal(params, &refParams); err != nil {
		return nil, err
	}

	doc, err := s.documents.Get(refParams.TextDocument.URI)
	if err != nil {
		return nil, err
	}

	result := s.getAnalysis(refParams.TextDocument.URI)
	if result == nil {
		result = s.analyzer.Analyze(doc)
	}

	locations := s.analyzer.GetReferences(doc, refParams.Position, result, refParams.Context.IncludeDeclaration)
	return locations, nil
}

func (s *Server) handleDocumentSymbol(params json.RawMessage) (interface{}, error) {
	var symbolParams DocumentSymbolParams
	if err := json.Unmarshal(params, &symbolParams); err != nil {
		return nil, err
	}

	doc, err := s.documents.Get(symbolParams.TextDocument.URI)
	if err != nil {
		return nil, err
	}

	result := s.getAnalysis(symbolParams.TextDocument.URI)
	if result == nil {
		result = s.analyzer.Analyze(doc)
	}

	symbols := s.analyzer.GetDocumentSymbols(result)
	return symbols, nil
}

func (s *Server) handleSignatureHelp(params json.RawMessage) (interface{}, error) {
	var sigParams SignatureHelpParams
	if err := json.Unmarshal(params, &sigParams); err != nil {
		return nil, err
	}

	doc, err := s.documents.Get(sigParams.TextDocument.URI)
	if err != nil {
		return nil, err
	}

	result := s.getAnalysis(sigParams.TextDocument.URI)
	if result == nil {
		result = s.analyzer.Analyze(doc)
	}

	sigHelp := s.analyzer.GetSignatureHelp(doc, sigParams.Position, result)
	return sigHelp, nil
}

func (s *Server) handleFormatting(params json.RawMessage) (interface{}, error) {
	var fmtParams DocumentFormattingParams
	if err := json.Unmarshal(params, &fmtParams); err != nil {
		return nil, err
	}

	doc, err := s.documents.Get(fmtParams.TextDocument.URI)
	if err != nil {
		return nil, err
	}

	// Simple formatting: ensure proper spacing and indentation
	formatted := s.formatDocument(doc, fmtParams.Options)
	if formatted == doc.Content {
		return []TextEdit{}, nil
	}

	// Return a single edit that replaces the entire document
	return []TextEdit{
		{
			Range: Range{
				Start: Position{Line: 0, Character: 0},
				End:   doc.OffsetToPosition(len(doc.Content)),
			},
			NewText: formatted,
		},
	}, nil
}

func (s *Server) formatDocument(doc *Document, options FormattingOptions) string {
	var result strings.Builder
	lines := doc.Lines
	indentLevel := 0
	indent := "\t"
	if options.InsertSpaces {
		indent = strings.Repeat(" ", options.TabSize)
	}

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		lowerTrimmed := strings.ToLower(trimmed)

		// Decrease indent for closing statements
		if strings.HasPrefix(lowerTrimmed, blockEndKeyword) ||
			strings.HasPrefix(lowerTrimmed, elseKeyword) {
			if indentLevel > 0 {
				indentLevel--
			}
		}

		// Write indented line
		if trimmed != "" {
			result.WriteString(strings.Repeat(indent, indentLevel))
			result.WriteString(trimmed)
		}

		if i < len(lines)-1 {
			result.WriteString("\n")
		}

		// Increase indent for block-opening statements
		if strings.HasSuffix(lowerTrimmed, blockStartSuffix) ||
			strings.HasSuffix(lowerTrimmed, thenSuffix) ||
			(strings.HasPrefix(lowerTrimmed, elseKeyword) && !strings.HasSuffix(lowerTrimmed, ".")) {
			indentLevel++
		}
	}

	if options.InsertFinalNewline && !strings.HasSuffix(result.String(), "\n") {
		result.WriteString("\n")
	}

	if options.TrimFinalNewlines {
		return strings.TrimRight(result.String(), "\n") + "\n"
	}

	return result.String()
}

func (s *Server) handleCodeAction(params json.RawMessage) (interface{}, error) {
	var caParams CodeActionParams
	if err := json.Unmarshal(params, &caParams); err != nil {
		return nil, err
	}

	actions := make([]CodeAction, 0)

	// Provide quick fixes for diagnostics
	for _, diag := range caParams.Context.Diagnostics {
		// Extract suggestions from diagnostic message
		if strings.Contains(diag.Message, "Perhaps you meant") {
			// Extract the suggestion
			idx := strings.Index(diag.Message, "Perhaps you meant")
			if idx != -1 {
				suggestion := strings.TrimSpace(diag.Message[idx+17:])
				if colonIdx := strings.Index(suggestion, ":"); colonIdx != -1 {
					suggestion = strings.TrimSpace(suggestion[colonIdx+1:])
				}
				suggestion = strings.Trim(suggestion, "'\"")

				if suggestion != "" {
					actions = append(actions, CodeAction{
						Title: "Did you mean: " + suggestion,
						Kind:  CodeActionKindQuickFix,
						Diagnostics: []Diagnostic{diag},
						Edit: &WorkspaceEdit{
							Changes: map[string][]TextEdit{
								caParams.TextDocument.URI: {
									{
										Range:   diag.Range,
										NewText: suggestion,
									},
								},
							},
						},
					})
				}
			}
		}

		// Suggest adding a period if missing
		if strings.Contains(diag.Message, "period") {
			actions = append(actions, CodeAction{
				Title: "Add period at end of statement",
				Kind:  CodeActionKindQuickFix,
				Diagnostics: []Diagnostic{diag},
				Edit: &WorkspaceEdit{
					Changes: map[string][]TextEdit{
						caParams.TextDocument.URI: {
							{
								Range:   Range{Start: diag.Range.End, End: diag.Range.End},
								NewText: ".",
							},
						},
					},
				},
			})
		}
	}

	return actions, nil
}

func (s *Server) handleDocumentHighlight(params json.RawMessage) (interface{}, error) {
	var hlParams DocumentHighlightParams
	if err := json.Unmarshal(params, &hlParams); err != nil {
		return nil, err
	}

	doc, err := s.documents.Get(hlParams.TextDocument.URI)
	if err != nil {
		return nil, err
	}

	result := s.getAnalysis(hlParams.TextDocument.URI)
	if result == nil {
		result = s.analyzer.Analyze(doc)
	}

	word, _ := doc.GetWordAtPosition(hlParams.Position)
	if word == "" {
		return []DocumentHighlight{}, nil
	}

	highlights := make([]DocumentHighlight, 0)
	for _, ref := range result.References {
		if ref.Name == word {
			kind := DocumentHighlightKindRead
			if ref.IsDefinition {
				kind = DocumentHighlightKindWrite
			}
			highlights = append(highlights, DocumentHighlight{
				Range: ref.Range,
				Kind:  kind,
			})
		}
	}

	return highlights, nil
}

func (s *Server) handleFoldingRange(params json.RawMessage) (interface{}, error) {
	var frParams FoldingRangeParams
	if err := json.Unmarshal(params, &frParams); err != nil {
		return nil, err
	}

	doc, err := s.documents.Get(frParams.TextDocument.URI)
	if err != nil {
		return nil, err
	}

	ranges := make([]FoldingRange, 0)

	// Find foldable regions (blocks ending with "Thats it.")
	var blockStarts []int
	for i, line := range doc.Lines {
		trimmed := strings.TrimSpace(strings.ToLower(line))

		// Check for block starts
		if strings.HasSuffix(trimmed, ":") ||
			strings.HasSuffix(trimmed, ", then") {
			blockStarts = append(blockStarts, i)
		}

		// Check for block ends
		if strings.HasPrefix(trimmed, "thats it") && len(blockStarts) > 0 {
			startLine := blockStarts[len(blockStarts)-1]
			blockStarts = blockStarts[:len(blockStarts)-1]
			ranges = append(ranges, FoldingRange{
				StartLine: startLine,
				EndLine:   i,
				Kind:      FoldingRangeKindRegion,
			})
		}
	}

	// Find comment blocks
	commentStart := -1
	for i, line := range doc.Lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") {
			if commentStart == -1 {
				commentStart = i
			}
		} else {
			if commentStart != -1 && i-commentStart > 1 {
				ranges = append(ranges, FoldingRange{
					StartLine: commentStart,
					EndLine:   i - 1,
					Kind:      FoldingRangeKindComment,
				})
			}
			commentStart = -1
		}
	}

	return ranges, nil
}
