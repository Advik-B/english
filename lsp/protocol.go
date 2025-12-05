// Package lsp provides a Language Server Protocol implementation for the English programming language.
// This package is designed to be modular, extensible, and batteries-included.
package lsp

import (
	"encoding/json"
)

// JSON-RPC Types

// Message represents a JSON-RPC message
type Message struct {
	JSONRPC string `json:"jsonrpc"`
}

// RequestMessage represents a JSON-RPC request
type RequestMessage struct {
	Message
	ID     interface{}     `json:"id"`
	Method string          `json:"method"`
	Params json.RawMessage `json:"params,omitempty"`
}

// ResponseMessage represents a JSON-RPC response
type ResponseMessage struct {
	Message
	ID     interface{} `json:"id"`
	Result interface{} `json:"result,omitempty"`
	Error  *Error      `json:"error,omitempty"`
}

// NotificationMessage represents a JSON-RPC notification (no ID)
type NotificationMessage struct {
	Message
	Method string          `json:"method"`
	Params json.RawMessage `json:"params,omitempty"`
}

// Error represents a JSON-RPC error
type Error struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Error codes defined by JSON-RPC and LSP
const (
	// JSON-RPC errors
	ParseError     = -32700
	InvalidRequest = -32600
	MethodNotFound = -32601
	InvalidParams  = -32602
	InternalError  = -32603

	// LSP errors
	ServerNotInitialized = -32002
	UnknownErrorCode     = -32001
	RequestCancelled     = -32800
	ContentModified      = -32801
)

// LSP Types

// Position represents a position in a text document
type Position struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

// Range represents a range in a text document
type Range struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}

// Location represents a location in a document
type Location struct {
	URI   string `json:"uri"`
	Range Range  `json:"range"`
}

// TextDocumentIdentifier identifies a text document
type TextDocumentIdentifier struct {
	URI string `json:"uri"`
}

// VersionedTextDocumentIdentifier identifies a specific version of a text document
type VersionedTextDocumentIdentifier struct {
	TextDocumentIdentifier
	Version int `json:"version"`
}

// TextDocumentItem represents a text document transfer object
type TextDocumentItem struct {
	URI        string `json:"uri"`
	LanguageID string `json:"languageId"`
	Version    int    `json:"version"`
	Text       string `json:"text"`
}

// TextDocumentPositionParams is a parameter literal used for requests
// that take a position inside a text document
type TextDocumentPositionParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     Position               `json:"position"`
}

// TextEdit represents a textual edit
type TextEdit struct {
	Range   Range  `json:"range"`
	NewText string `json:"newText"`
}

// Diagnostic represents a diagnostic, such as a compiler error or warning
type Diagnostic struct {
	Range              Range                          `json:"range"`
	Severity           DiagnosticSeverity             `json:"severity,omitempty"`
	Code               interface{}                    `json:"code,omitempty"`
	CodeDescription    *CodeDescription               `json:"codeDescription,omitempty"`
	Source             string                         `json:"source,omitempty"`
	Message            string                         `json:"message"`
	Tags               []DiagnosticTag                `json:"tags,omitempty"`
	RelatedInformation []DiagnosticRelatedInformation `json:"relatedInformation,omitempty"`
}

// DiagnosticSeverity represents the severity of a diagnostic
type DiagnosticSeverity int

const (
	DiagnosticSeverityError       DiagnosticSeverity = 1
	DiagnosticSeverityWarning     DiagnosticSeverity = 2
	DiagnosticSeverityInformation DiagnosticSeverity = 3
	DiagnosticSeverityHint        DiagnosticSeverity = 4
)

// DiagnosticTag represents a diagnostic tag
type DiagnosticTag int

const (
	DiagnosticTagUnnecessary DiagnosticTag = 1
	DiagnosticTagDeprecated  DiagnosticTag = 2
)

// CodeDescription represents a code description
type CodeDescription struct {
	Href string `json:"href"`
}

// DiagnosticRelatedInformation represents related information for a diagnostic
type DiagnosticRelatedInformation struct {
	Location Location `json:"location"`
	Message  string   `json:"message"`
}

// Initialization Types

// InitializeParams is sent from the client to initialize the server
type InitializeParams struct {
	ProcessID             *int               `json:"processId"`
	ClientInfo            *ClientInfo        `json:"clientInfo,omitempty"`
	Locale                string             `json:"locale,omitempty"`
	RootPath              *string            `json:"rootPath,omitempty"`
	RootURI               *string            `json:"rootUri"`
	InitializationOptions interface{}        `json:"initializationOptions,omitempty"`
	Capabilities          ClientCapabilities `json:"capabilities"`
	Trace                 string             `json:"trace,omitempty"`
	WorkspaceFolders      []WorkspaceFolder  `json:"workspaceFolders,omitempty"`
}

// ClientInfo contains information about the client
type ClientInfo struct {
	Name    string `json:"name"`
	Version string `json:"version,omitempty"`
}

// WorkspaceFolder represents a workspace folder
type WorkspaceFolder struct {
	URI  string `json:"uri"`
	Name string `json:"name"`
}

// ClientCapabilities defines the client capabilities
type ClientCapabilities struct {
	Workspace    *WorkspaceClientCapabilities    `json:"workspace,omitempty"`
	TextDocument *TextDocumentClientCapabilities `json:"textDocument,omitempty"`
	Window       *WindowClientCapabilities       `json:"window,omitempty"`
	General      *GeneralClientCapabilities      `json:"general,omitempty"`
	Experimental interface{}                     `json:"experimental,omitempty"`
}

// WorkspaceClientCapabilities defines workspace-specific client capabilities
type WorkspaceClientCapabilities struct {
	ApplyEdit              bool                            `json:"applyEdit,omitempty"`
	WorkspaceEdit          *WorkspaceEditClientCapabilities `json:"workspaceEdit,omitempty"`
	DidChangeConfiguration *struct {
		DynamicRegistration bool `json:"dynamicRegistration,omitempty"`
	} `json:"didChangeConfiguration,omitempty"`
	DidChangeWatchedFiles *struct {
		DynamicRegistration bool `json:"dynamicRegistration,omitempty"`
	} `json:"didChangeWatchedFiles,omitempty"`
	Symbol *struct {
		DynamicRegistration bool `json:"dynamicRegistration,omitempty"`
	} `json:"symbol,omitempty"`
	ExecuteCommand *struct {
		DynamicRegistration bool `json:"dynamicRegistration,omitempty"`
	} `json:"executeCommand,omitempty"`
	Configuration            bool `json:"configuration,omitempty"`
	WorkspaceFolders         bool `json:"workspaceFolders,omitempty"`
	SemanticTokens           *SemanticTokensWorkspaceClientCapabilities `json:"semanticTokens,omitempty"`
	CodeLens                 *struct {
		RefreshSupport bool `json:"refreshSupport,omitempty"`
	} `json:"codeLens,omitempty"`
	FileOperations *struct {
		DynamicRegistration bool `json:"dynamicRegistration,omitempty"`
		DidCreate           bool `json:"didCreate,omitempty"`
		WillCreate          bool `json:"willCreate,omitempty"`
		DidRename           bool `json:"didRename,omitempty"`
		WillRename          bool `json:"willRename,omitempty"`
		DidDelete           bool `json:"didDelete,omitempty"`
		WillDelete          bool `json:"willDelete,omitempty"`
	} `json:"fileOperations,omitempty"`
}

// WorkspaceEditClientCapabilities defines client capabilities for workspace edits
type WorkspaceEditClientCapabilities struct {
	DocumentChanges    bool     `json:"documentChanges,omitempty"`
	ResourceOperations []string `json:"resourceOperations,omitempty"`
	FailureHandling    string   `json:"failureHandling,omitempty"`
	NormalizesLineEndings bool  `json:"normalizesLineEndings,omitempty"`
}

// SemanticTokensWorkspaceClientCapabilities defines semantic tokens workspace capabilities
type SemanticTokensWorkspaceClientCapabilities struct {
	RefreshSupport bool `json:"refreshSupport,omitempty"`
}

// TextDocumentClientCapabilities defines text document-specific client capabilities
type TextDocumentClientCapabilities struct {
	Synchronization *TextDocumentSyncClientCapabilities    `json:"synchronization,omitempty"`
	Completion      *CompletionClientCapabilities          `json:"completion,omitempty"`
	Hover           *HoverClientCapabilities               `json:"hover,omitempty"`
	SignatureHelp   *SignatureHelpClientCapabilities       `json:"signatureHelp,omitempty"`
	Declaration     *DeclarationClientCapabilities         `json:"declaration,omitempty"`
	Definition      *DefinitionClientCapabilities          `json:"definition,omitempty"`
	TypeDefinition  *TypeDefinitionClientCapabilities      `json:"typeDefinition,omitempty"`
	Implementation  *ImplementationClientCapabilities      `json:"implementation,omitempty"`
	References      *ReferencesClientCapabilities          `json:"references,omitempty"`
	DocumentHighlight *DocumentHighlightClientCapabilities `json:"documentHighlight,omitempty"`
	DocumentSymbol  *DocumentSymbolClientCapabilities      `json:"documentSymbol,omitempty"`
	CodeAction      *CodeActionClientCapabilities          `json:"codeAction,omitempty"`
	CodeLens        *CodeLensClientCapabilities            `json:"codeLens,omitempty"`
	DocumentLink    *DocumentLinkClientCapabilities        `json:"documentLink,omitempty"`
	ColorProvider   *DocumentColorClientCapabilities       `json:"colorProvider,omitempty"`
	Formatting      *DocumentFormattingClientCapabilities  `json:"formatting,omitempty"`
	RangeFormatting *DocumentRangeFormattingClientCapabilities `json:"rangeFormatting,omitempty"`
	OnTypeFormatting *DocumentOnTypeFormattingClientCapabilities `json:"onTypeFormatting,omitempty"`
	Rename          *RenameClientCapabilities              `json:"rename,omitempty"`
	PublishDiagnostics *PublishDiagnosticsClientCapabilities `json:"publishDiagnostics,omitempty"`
	FoldingRange    *FoldingRangeClientCapabilities        `json:"foldingRange,omitempty"`
	SelectionRange  *SelectionRangeClientCapabilities      `json:"selectionRange,omitempty"`
	LinkedEditingRange *LinkedEditingRangeClientCapabilities `json:"linkedEditingRange,omitempty"`
	CallHierarchy   *CallHierarchyClientCapabilities       `json:"callHierarchy,omitempty"`
	SemanticTokens  *SemanticTokensClientCapabilities      `json:"semanticTokens,omitempty"`
	Moniker         *MonikerClientCapabilities             `json:"moniker,omitempty"`
}

// Stub types for client capabilities - these allow for extensibility
type TextDocumentSyncClientCapabilities struct {
	DynamicRegistration bool `json:"dynamicRegistration,omitempty"`
	WillSave            bool `json:"willSave,omitempty"`
	WillSaveWaitUntil   bool `json:"willSaveWaitUntil,omitempty"`
	DidSave             bool `json:"didSave,omitempty"`
}

type CompletionClientCapabilities struct {
	DynamicRegistration bool `json:"dynamicRegistration,omitempty"`
	CompletionItem      *struct {
		SnippetSupport          bool     `json:"snippetSupport,omitempty"`
		CommitCharactersSupport bool     `json:"commitCharactersSupport,omitempty"`
		DocumentationFormat     []string `json:"documentationFormat,omitempty"`
		DeprecatedSupport       bool     `json:"deprecatedSupport,omitempty"`
		PreselectSupport        bool     `json:"preselectSupport,omitempty"`
	} `json:"completionItem,omitempty"`
	ContextSupport bool `json:"contextSupport,omitempty"`
}

type HoverClientCapabilities struct {
	DynamicRegistration bool     `json:"dynamicRegistration,omitempty"`
	ContentFormat       []string `json:"contentFormat,omitempty"`
}

type SignatureHelpClientCapabilities struct {
	DynamicRegistration bool `json:"dynamicRegistration,omitempty"`
	SignatureInformation *struct {
		DocumentationFormat  []string `json:"documentationFormat,omitempty"`
		ParameterInformation *struct {
			LabelOffsetSupport bool `json:"labelOffsetSupport,omitempty"`
		} `json:"parameterInformation,omitempty"`
	} `json:"signatureInformation,omitempty"`
	ContextSupport bool `json:"contextSupport,omitempty"`
}

type DeclarationClientCapabilities struct {
	DynamicRegistration bool `json:"dynamicRegistration,omitempty"`
	LinkSupport         bool `json:"linkSupport,omitempty"`
}

type DefinitionClientCapabilities struct {
	DynamicRegistration bool `json:"dynamicRegistration,omitempty"`
	LinkSupport         bool `json:"linkSupport,omitempty"`
}

type TypeDefinitionClientCapabilities struct {
	DynamicRegistration bool `json:"dynamicRegistration,omitempty"`
	LinkSupport         bool `json:"linkSupport,omitempty"`
}

type ImplementationClientCapabilities struct {
	DynamicRegistration bool `json:"dynamicRegistration,omitempty"`
	LinkSupport         bool `json:"linkSupport,omitempty"`
}

type ReferencesClientCapabilities struct {
	DynamicRegistration bool `json:"dynamicRegistration,omitempty"`
}

type DocumentHighlightClientCapabilities struct {
	DynamicRegistration bool `json:"dynamicRegistration,omitempty"`
}

type DocumentSymbolClientCapabilities struct {
	DynamicRegistration               bool `json:"dynamicRegistration,omitempty"`
	SymbolKind                        *struct {
		ValueSet []SymbolKind `json:"valueSet,omitempty"`
	} `json:"symbolKind,omitempty"`
	HierarchicalDocumentSymbolSupport bool `json:"hierarchicalDocumentSymbolSupport,omitempty"`
}

type CodeActionClientCapabilities struct {
	DynamicRegistration bool `json:"dynamicRegistration,omitempty"`
	CodeActionLiteralSupport *struct {
		CodeActionKind struct {
			ValueSet []string `json:"valueSet,omitempty"`
		} `json:"codeActionKind,omitempty"`
	} `json:"codeActionLiteralSupport,omitempty"`
	IsPreferredSupport bool `json:"isPreferredSupport,omitempty"`
}

type CodeLensClientCapabilities struct {
	DynamicRegistration bool `json:"dynamicRegistration,omitempty"`
}

type DocumentLinkClientCapabilities struct {
	DynamicRegistration bool `json:"dynamicRegistration,omitempty"`
	TooltipSupport      bool `json:"tooltipSupport,omitempty"`
}

type DocumentColorClientCapabilities struct {
	DynamicRegistration bool `json:"dynamicRegistration,omitempty"`
}

type DocumentFormattingClientCapabilities struct {
	DynamicRegistration bool `json:"dynamicRegistration,omitempty"`
}

type DocumentRangeFormattingClientCapabilities struct {
	DynamicRegistration bool `json:"dynamicRegistration,omitempty"`
}

type DocumentOnTypeFormattingClientCapabilities struct {
	DynamicRegistration bool `json:"dynamicRegistration,omitempty"`
}

type RenameClientCapabilities struct {
	DynamicRegistration bool `json:"dynamicRegistration,omitempty"`
	PrepareSupport      bool `json:"prepareSupport,omitempty"`
}

type PublishDiagnosticsClientCapabilities struct {
	RelatedInformation     bool `json:"relatedInformation,omitempty"`
	TagSupport             *struct {
		ValueSet []DiagnosticTag `json:"valueSet,omitempty"`
	} `json:"tagSupport,omitempty"`
	VersionSupport         bool `json:"versionSupport,omitempty"`
	CodeDescriptionSupport bool `json:"codeDescriptionSupport,omitempty"`
	DataSupport            bool `json:"dataSupport,omitempty"`
}

type FoldingRangeClientCapabilities struct {
	DynamicRegistration bool `json:"dynamicRegistration,omitempty"`
	RangeLimit          int  `json:"rangeLimit,omitempty"`
	LineFoldingOnly     bool `json:"lineFoldingOnly,omitempty"`
}

type SelectionRangeClientCapabilities struct {
	DynamicRegistration bool `json:"dynamicRegistration,omitempty"`
}

type LinkedEditingRangeClientCapabilities struct {
	DynamicRegistration bool `json:"dynamicRegistration,omitempty"`
}

type CallHierarchyClientCapabilities struct {
	DynamicRegistration bool `json:"dynamicRegistration,omitempty"`
}

type SemanticTokensClientCapabilities struct {
	DynamicRegistration     bool `json:"dynamicRegistration,omitempty"`
	Requests                *struct {
		Range interface{} `json:"range,omitempty"`
		Full  interface{} `json:"full,omitempty"`
	} `json:"requests,omitempty"`
	TokenTypes     []string `json:"tokenTypes,omitempty"`
	TokenModifiers []string `json:"tokenModifiers,omitempty"`
	Formats        []string `json:"formats,omitempty"`
	OverlappingTokenSupport bool `json:"overlappingTokenSupport,omitempty"`
	MultilineTokenSupport   bool `json:"multilineTokenSupport,omitempty"`
}

type MonikerClientCapabilities struct {
	DynamicRegistration bool `json:"dynamicRegistration,omitempty"`
}

// WindowClientCapabilities defines window-specific client capabilities
type WindowClientCapabilities struct {
	WorkDoneProgress bool `json:"workDoneProgress,omitempty"`
	ShowMessage      *struct {
		MessageActionItem *struct {
			AdditionalPropertiesSupport bool `json:"additionalPropertiesSupport,omitempty"`
		} `json:"messageActionItem,omitempty"`
	} `json:"showMessage,omitempty"`
	ShowDocument *struct {
		Support bool `json:"support,omitempty"`
	} `json:"showDocument,omitempty"`
}

// GeneralClientCapabilities defines general client capabilities
type GeneralClientCapabilities struct {
	StaleRequestSupport *struct {
		Cancel                   bool     `json:"cancel,omitempty"`
		RetryOnContentModified   []string `json:"retryOnContentModified,omitempty"`
	} `json:"staleRequestSupport,omitempty"`
	RegularExpressions *struct {
		Engine  string `json:"engine,omitempty"`
		Version string `json:"version,omitempty"`
	} `json:"regularExpressions,omitempty"`
	Markdown *struct {
		Parser  string   `json:"parser,omitempty"`
		Version string   `json:"version,omitempty"`
		AllowedTags []string `json:"allowedTags,omitempty"`
	} `json:"markdown,omitempty"`
}

// InitializeResult is the result of the initialize request
type InitializeResult struct {
	Capabilities ServerCapabilities `json:"capabilities"`
	ServerInfo   *ServerInfo        `json:"serverInfo,omitempty"`
}

// ServerInfo contains information about the server
type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version,omitempty"`
}

// ServerCapabilities defines the server capabilities
type ServerCapabilities struct {
	TextDocumentSync           *TextDocumentSyncOptions       `json:"textDocumentSync,omitempty"`
	CompletionProvider         *CompletionOptions             `json:"completionProvider,omitempty"`
	HoverProvider              interface{}                    `json:"hoverProvider,omitempty"`
	SignatureHelpProvider      *SignatureHelpOptions          `json:"signatureHelpProvider,omitempty"`
	DeclarationProvider        interface{}                    `json:"declarationProvider,omitempty"`
	DefinitionProvider         interface{}                    `json:"definitionProvider,omitempty"`
	TypeDefinitionProvider     interface{}                    `json:"typeDefinitionProvider,omitempty"`
	ImplementationProvider     interface{}                    `json:"implementationProvider,omitempty"`
	ReferencesProvider         interface{}                    `json:"referencesProvider,omitempty"`
	DocumentHighlightProvider  interface{}                    `json:"documentHighlightProvider,omitempty"`
	DocumentSymbolProvider     interface{}                    `json:"documentSymbolProvider,omitempty"`
	CodeActionProvider         interface{}                    `json:"codeActionProvider,omitempty"`
	CodeLensProvider           *CodeLensOptions               `json:"codeLensProvider,omitempty"`
	DocumentLinkProvider       *DocumentLinkOptions           `json:"documentLinkProvider,omitempty"`
	ColorProvider              interface{}                    `json:"colorProvider,omitempty"`
	DocumentFormattingProvider interface{}                    `json:"documentFormattingProvider,omitempty"`
	DocumentRangeFormattingProvider interface{}               `json:"documentRangeFormattingProvider,omitempty"`
	DocumentOnTypeFormattingProvider *DocumentOnTypeFormattingOptions `json:"documentOnTypeFormattingProvider,omitempty"`
	RenameProvider             interface{}                    `json:"renameProvider,omitempty"`
	FoldingRangeProvider       interface{}                    `json:"foldingRangeProvider,omitempty"`
	ExecuteCommandProvider     *ExecuteCommandOptions         `json:"executeCommandProvider,omitempty"`
	SelectionRangeProvider     interface{}                    `json:"selectionRangeProvider,omitempty"`
	LinkedEditingRangeProvider interface{}                    `json:"linkedEditingRangeProvider,omitempty"`
	CallHierarchyProvider      interface{}                    `json:"callHierarchyProvider,omitempty"`
	SemanticTokensProvider     interface{}                    `json:"semanticTokensProvider,omitempty"`
	MonikerProvider            interface{}                    `json:"monikerProvider,omitempty"`
	WorkspaceSymbolProvider    interface{}                    `json:"workspaceSymbolProvider,omitempty"`
	Workspace                  *ServerWorkspaceCapabilities   `json:"workspace,omitempty"`
	Experimental               interface{}                    `json:"experimental,omitempty"`
}

// TextDocumentSyncOptions defines options for text document sync
type TextDocumentSyncOptions struct {
	OpenClose bool                 `json:"openClose,omitempty"`
	Change    TextDocumentSyncKind `json:"change,omitempty"`
	WillSave  bool                 `json:"willSave,omitempty"`
	WillSaveWaitUntil bool         `json:"willSaveWaitUntil,omitempty"`
	Save      *SaveOptions         `json:"save,omitempty"`
}

// TextDocumentSyncKind defines the kind of text document sync
type TextDocumentSyncKind int

const (
	TextDocumentSyncKindNone        TextDocumentSyncKind = 0
	TextDocumentSyncKindFull        TextDocumentSyncKind = 1
	TextDocumentSyncKindIncremental TextDocumentSyncKind = 2
)

// SaveOptions defines options for save events
type SaveOptions struct {
	IncludeText bool `json:"includeText,omitempty"`
}

// CompletionOptions defines options for completion
type CompletionOptions struct {
	TriggerCharacters   []string `json:"triggerCharacters,omitempty"`
	AllCommitCharacters []string `json:"allCommitCharacters,omitempty"`
	ResolveProvider     bool     `json:"resolveProvider,omitempty"`
	WorkDoneProgress    bool     `json:"workDoneProgress,omitempty"`
}

// SignatureHelpOptions defines options for signature help
type SignatureHelpOptions struct {
	TriggerCharacters   []string `json:"triggerCharacters,omitempty"`
	RetriggerCharacters []string `json:"retriggerCharacters,omitempty"`
	WorkDoneProgress    bool     `json:"workDoneProgress,omitempty"`
}

// CodeLensOptions defines options for code lens
type CodeLensOptions struct {
	ResolveProvider  bool `json:"resolveProvider,omitempty"`
	WorkDoneProgress bool `json:"workDoneProgress,omitempty"`
}

// DocumentLinkOptions defines options for document links
type DocumentLinkOptions struct {
	ResolveProvider  bool `json:"resolveProvider,omitempty"`
	WorkDoneProgress bool `json:"workDoneProgress,omitempty"`
}

// DocumentOnTypeFormattingOptions defines options for on-type formatting
type DocumentOnTypeFormattingOptions struct {
	FirstTriggerCharacter string   `json:"firstTriggerCharacter"`
	MoreTriggerCharacter  []string `json:"moreTriggerCharacter,omitempty"`
}

// ExecuteCommandOptions defines options for execute command
type ExecuteCommandOptions struct {
	Commands         []string `json:"commands,omitempty"`
	WorkDoneProgress bool     `json:"workDoneProgress,omitempty"`
}

// ServerWorkspaceCapabilities defines server workspace capabilities
type ServerWorkspaceCapabilities struct {
	WorkspaceFolders *WorkspaceFoldersServerCapabilities `json:"workspaceFolders,omitempty"`
	FileOperations   *FileOperationOptions               `json:"fileOperations,omitempty"`
}

// WorkspaceFoldersServerCapabilities defines workspace folders capabilities
type WorkspaceFoldersServerCapabilities struct {
	Supported           bool        `json:"supported,omitempty"`
	ChangeNotifications interface{} `json:"changeNotifications,omitempty"`
}

// FileOperationOptions defines file operation options
type FileOperationOptions struct {
	DidCreate  *FileOperationRegistrationOptions `json:"didCreate,omitempty"`
	WillCreate *FileOperationRegistrationOptions `json:"willCreate,omitempty"`
	DidRename  *FileOperationRegistrationOptions `json:"didRename,omitempty"`
	WillRename *FileOperationRegistrationOptions `json:"willRename,omitempty"`
	DidDelete  *FileOperationRegistrationOptions `json:"didDelete,omitempty"`
	WillDelete *FileOperationRegistrationOptions `json:"willDelete,omitempty"`
}

// FileOperationRegistrationOptions defines file operation registration options
type FileOperationRegistrationOptions struct {
	Filters []FileOperationFilter `json:"filters,omitempty"`
}

// FileOperationFilter defines a file operation filter
type FileOperationFilter struct {
	Scheme  string               `json:"scheme,omitempty"`
	Pattern FileOperationPattern `json:"pattern"`
}

// FileOperationPattern defines a file operation pattern
type FileOperationPattern struct {
	Glob    string                       `json:"glob"`
	Matches FileOperationPatternKind     `json:"matches,omitempty"`
	Options *FileOperationPatternOptions `json:"options,omitempty"`
}

// FileOperationPatternKind defines the kind of file operation pattern
type FileOperationPatternKind string

const (
	FileOperationPatternKindFile   FileOperationPatternKind = "file"
	FileOperationPatternKindFolder FileOperationPatternKind = "folder"
)

// FileOperationPatternOptions defines file operation pattern options
type FileOperationPatternOptions struct {
	IgnoreCase bool `json:"ignoreCase,omitempty"`
}
