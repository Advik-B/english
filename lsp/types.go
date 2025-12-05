package lsp

// Completion Types

// CompletionParams contains parameters for completion requests
type CompletionParams struct {
	TextDocumentPositionParams
	Context *CompletionContext `json:"context,omitempty"`
}

// CompletionContext contains additional information about the context
type CompletionContext struct {
	TriggerKind      CompletionTriggerKind `json:"triggerKind"`
	TriggerCharacter string                `json:"triggerCharacter,omitempty"`
}

// CompletionTriggerKind defines how a completion was triggered
type CompletionTriggerKind int

const (
	CompletionTriggerKindInvoked                         CompletionTriggerKind = 1
	CompletionTriggerKindTriggerCharacter                CompletionTriggerKind = 2
	CompletionTriggerKindTriggerForIncompleteCompletions CompletionTriggerKind = 3
)

// CompletionList represents a collection of completion items
type CompletionList struct {
	IsIncomplete bool             `json:"isIncomplete"`
	Items        []CompletionItem `json:"items"`
}

// CompletionItem represents a completion item
type CompletionItem struct {
	Label               string              `json:"label"`
	Kind                CompletionItemKind  `json:"kind,omitempty"`
	Tags                []CompletionItemTag `json:"tags,omitempty"`
	Detail              string              `json:"detail,omitempty"`
	Documentation       interface{}         `json:"documentation,omitempty"`
	Deprecated          bool                `json:"deprecated,omitempty"`
	Preselect           bool                `json:"preselect,omitempty"`
	SortText            string              `json:"sortText,omitempty"`
	FilterText          string              `json:"filterText,omitempty"`
	InsertText          string              `json:"insertText,omitempty"`
	InsertTextFormat    InsertTextFormat    `json:"insertTextFormat,omitempty"`
	InsertTextMode      InsertTextMode      `json:"insertTextMode,omitempty"`
	TextEdit            *TextEdit           `json:"textEdit,omitempty"`
	AdditionalTextEdits []TextEdit          `json:"additionalTextEdits,omitempty"`
	CommitCharacters    []string            `json:"commitCharacters,omitempty"`
	Command             *Command            `json:"command,omitempty"`
	Data                interface{}         `json:"data,omitempty"`
}

// CompletionItemKind defines the kind of a completion entry
type CompletionItemKind int

const (
	CompletionItemKindText          CompletionItemKind = 1
	CompletionItemKindMethod        CompletionItemKind = 2
	CompletionItemKindFunction      CompletionItemKind = 3
	CompletionItemKindConstructor   CompletionItemKind = 4
	CompletionItemKindField         CompletionItemKind = 5
	CompletionItemKindVariable      CompletionItemKind = 6
	CompletionItemKindClass         CompletionItemKind = 7
	CompletionItemKindInterface     CompletionItemKind = 8
	CompletionItemKindModule        CompletionItemKind = 9
	CompletionItemKindProperty      CompletionItemKind = 10
	CompletionItemKindUnit          CompletionItemKind = 11
	CompletionItemKindValue         CompletionItemKind = 12
	CompletionItemKindEnum          CompletionItemKind = 13
	CompletionItemKindKeyword       CompletionItemKind = 14
	CompletionItemKindSnippet       CompletionItemKind = 15
	CompletionItemKindColor         CompletionItemKind = 16
	CompletionItemKindFile          CompletionItemKind = 17
	CompletionItemKindReference     CompletionItemKind = 18
	CompletionItemKindFolder        CompletionItemKind = 19
	CompletionItemKindEnumMember    CompletionItemKind = 20
	CompletionItemKindConstant      CompletionItemKind = 21
	CompletionItemKindStruct        CompletionItemKind = 22
	CompletionItemKindEvent         CompletionItemKind = 23
	CompletionItemKindOperator      CompletionItemKind = 24
	CompletionItemKindTypeParameter CompletionItemKind = 25
)

// CompletionItemTag defines extra annotations for completion items
type CompletionItemTag int

const (
	CompletionItemTagDeprecated CompletionItemTag = 1
)

// InsertTextFormat defines how insert text is interpreted
type InsertTextFormat int

const (
	InsertTextFormatPlainText InsertTextFormat = 1
	InsertTextFormatSnippet   InsertTextFormat = 2
)

// InsertTextMode defines how whitespace is handled during completion
type InsertTextMode int

const (
	InsertTextModeAsIs              InsertTextMode = 1
	InsertTextModeAdjustIndentation InsertTextMode = 2
)

// Command represents a reference to a command
type Command struct {
	Title     string        `json:"title"`
	Command   string        `json:"command"`
	Arguments []interface{} `json:"arguments,omitempty"`
}

// Hover Types

// HoverParams contains parameters for hover requests
type HoverParams struct {
	TextDocumentPositionParams
}

// Hover represents hover information
type Hover struct {
	Contents MarkupContent `json:"contents"`
	Range    *Range        `json:"range,omitempty"`
}

// MarkupContent represents a string value with content type
type MarkupContent struct {
	Kind  MarkupKind `json:"kind"`
	Value string     `json:"value"`
}

// MarkupKind describes the content type
type MarkupKind string

const (
	MarkupKindPlainText MarkupKind = "plaintext"
	MarkupKindMarkdown  MarkupKind = "markdown"
)

// Signature Help Types

// SignatureHelpParams contains parameters for signature help requests
type SignatureHelpParams struct {
	TextDocumentPositionParams
	Context *SignatureHelpContext `json:"context,omitempty"`
}

// SignatureHelpContext contains additional information about the context
type SignatureHelpContext struct {
	TriggerKind         SignatureHelpTriggerKind `json:"triggerKind"`
	TriggerCharacter    string                   `json:"triggerCharacter,omitempty"`
	IsRetrigger         bool                     `json:"isRetrigger"`
	ActiveSignatureHelp *SignatureHelp           `json:"activeSignatureHelp,omitempty"`
}

// SignatureHelpTriggerKind defines how signature help was triggered
type SignatureHelpTriggerKind int

const (
	SignatureHelpTriggerKindInvoked         SignatureHelpTriggerKind = 1
	SignatureHelpTriggerKindTriggerCharacter SignatureHelpTriggerKind = 2
	SignatureHelpTriggerKindContentChange   SignatureHelpTriggerKind = 3
)

// SignatureHelp represents the signature of something callable
type SignatureHelp struct {
	Signatures      []SignatureInformation `json:"signatures"`
	ActiveSignature *int                   `json:"activeSignature,omitempty"`
	ActiveParameter *int                   `json:"activeParameter,omitempty"`
}

// SignatureInformation represents the signature of something callable
type SignatureInformation struct {
	Label           string                 `json:"label"`
	Documentation   interface{}            `json:"documentation,omitempty"`
	Parameters      []ParameterInformation `json:"parameters,omitempty"`
	ActiveParameter *int                   `json:"activeParameter,omitempty"`
}

// ParameterInformation represents a parameter of a callable
type ParameterInformation struct {
	Label         interface{} `json:"label"`
	Documentation interface{} `json:"documentation,omitempty"`
}

// Document Symbol Types

// DocumentSymbolParams contains parameters for document symbol requests
type DocumentSymbolParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
}

// DocumentSymbol represents a symbol found in a document
type DocumentSymbol struct {
	Name           string           `json:"name"`
	Detail         string           `json:"detail,omitempty"`
	Kind           SymbolKind       `json:"kind"`
	Tags           []SymbolTag      `json:"tags,omitempty"`
	Deprecated     bool             `json:"deprecated,omitempty"`
	Range          Range            `json:"range"`
	SelectionRange Range            `json:"selectionRange"`
	Children       []DocumentSymbol `json:"children,omitempty"`
}

// SymbolInformation represents information about a symbol
type SymbolInformation struct {
	Name          string     `json:"name"`
	Kind          SymbolKind `json:"kind"`
	Tags          []SymbolTag `json:"tags,omitempty"`
	Deprecated    bool       `json:"deprecated,omitempty"`
	Location      Location   `json:"location"`
	ContainerName string     `json:"containerName,omitempty"`
}

// SymbolKind represents the kind of a symbol
type SymbolKind int

const (
	SymbolKindFile          SymbolKind = 1
	SymbolKindModule        SymbolKind = 2
	SymbolKindNamespace     SymbolKind = 3
	SymbolKindPackage       SymbolKind = 4
	SymbolKindClass         SymbolKind = 5
	SymbolKindMethod        SymbolKind = 6
	SymbolKindProperty      SymbolKind = 7
	SymbolKindField         SymbolKind = 8
	SymbolKindConstructor   SymbolKind = 9
	SymbolKindEnum          SymbolKind = 10
	SymbolKindInterface     SymbolKind = 11
	SymbolKindFunction      SymbolKind = 12
	SymbolKindVariable      SymbolKind = 13
	SymbolKindConstant      SymbolKind = 14
	SymbolKindString        SymbolKind = 15
	SymbolKindNumber        SymbolKind = 16
	SymbolKindBoolean       SymbolKind = 17
	SymbolKindArray         SymbolKind = 18
	SymbolKindObject        SymbolKind = 19
	SymbolKindKey           SymbolKind = 20
	SymbolKindNull          SymbolKind = 21
	SymbolKindEnumMember    SymbolKind = 22
	SymbolKindStruct        SymbolKind = 23
	SymbolKindEvent         SymbolKind = 24
	SymbolKindOperator      SymbolKind = 25
	SymbolKindTypeParameter SymbolKind = 26
)

// SymbolTag represents extra annotations for symbols
type SymbolTag int

const (
	SymbolTagDeprecated SymbolTag = 1
)

// Definition Types

// DefinitionParams contains parameters for definition requests
type DefinitionParams struct {
	TextDocumentPositionParams
}

// Reference Types

// ReferenceParams contains parameters for reference requests
type ReferenceParams struct {
	TextDocumentPositionParams
	Context ReferenceContext `json:"context"`
}

// ReferenceContext contains additional information for reference requests
type ReferenceContext struct {
	IncludeDeclaration bool `json:"includeDeclaration"`
}

// Text Document Sync Types

// DidOpenTextDocumentParams contains parameters for textDocument/didOpen
type DidOpenTextDocumentParams struct {
	TextDocument TextDocumentItem `json:"textDocument"`
}

// DidCloseTextDocumentParams contains parameters for textDocument/didClose
type DidCloseTextDocumentParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
}

// DidChangeTextDocumentParams contains parameters for textDocument/didChange
type DidChangeTextDocumentParams struct {
	TextDocument   VersionedTextDocumentIdentifier  `json:"textDocument"`
	ContentChanges []TextDocumentContentChangeEvent `json:"contentChanges"`
}

// TextDocumentContentChangeEvent describes textual changes
type TextDocumentContentChangeEvent struct {
	Range       *Range `json:"range,omitempty"`
	RangeLength int    `json:"rangeLength,omitempty"`
	Text        string `json:"text"`
}

// DidSaveTextDocumentParams contains parameters for textDocument/didSave
type DidSaveTextDocumentParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Text         string                 `json:"text,omitempty"`
}

// Diagnostic Types

// PublishDiagnosticsParams contains parameters for textDocument/publishDiagnostics
type PublishDiagnosticsParams struct {
	URI         string       `json:"uri"`
	Version     *int         `json:"version,omitempty"`
	Diagnostics []Diagnostic `json:"diagnostics"`
}

// Code Action Types

// CodeActionParams contains parameters for code action requests
type CodeActionParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Range        Range                  `json:"range"`
	Context      CodeActionContext      `json:"context"`
}

// CodeActionContext contains additional context information
type CodeActionContext struct {
	Diagnostics []Diagnostic `json:"diagnostics"`
	Only        []string     `json:"only,omitempty"`
}

// CodeAction represents a code action
type CodeAction struct {
	Title       string         `json:"title"`
	Kind        string         `json:"kind,omitempty"`
	Diagnostics []Diagnostic   `json:"diagnostics,omitempty"`
	IsPreferred bool           `json:"isPreferred,omitempty"`
	Disabled    *struct {
		Reason string `json:"reason"`
	} `json:"disabled,omitempty"`
	Edit    *WorkspaceEdit `json:"edit,omitempty"`
	Command *Command       `json:"command,omitempty"`
	Data    interface{}    `json:"data,omitempty"`
}

// CodeActionKind defines well-known code action kinds
const (
	CodeActionKindQuickFix              = "quickfix"
	CodeActionKindRefactor              = "refactor"
	CodeActionKindRefactorExtract       = "refactor.extract"
	CodeActionKindRefactorInline        = "refactor.inline"
	CodeActionKindRefactorRewrite       = "refactor.rewrite"
	CodeActionKindSource                = "source"
	CodeActionKindSourceOrganizeImports = "source.organizeImports"
)

// WorkspaceEdit represents changes to many resources managed in the workspace
type WorkspaceEdit struct {
	Changes         map[string][]TextEdit `json:"changes,omitempty"`
	DocumentChanges []interface{}         `json:"documentChanges,omitempty"`
}

// Rename Types

// RenameParams contains parameters for rename requests
type RenameParams struct {
	TextDocumentPositionParams
	NewName string `json:"newName"`
}

// Formatting Types

// DocumentFormattingParams contains parameters for formatting requests
type DocumentFormattingParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Options      FormattingOptions      `json:"options"`
}

// FormattingOptions defines value-object describing formatting options
type FormattingOptions struct {
	TabSize                int  `json:"tabSize"`
	InsertSpaces           bool `json:"insertSpaces"`
	TrimTrailingWhitespace bool `json:"trimTrailingWhitespace,omitempty"`
	InsertFinalNewline     bool `json:"insertFinalNewline,omitempty"`
	TrimFinalNewlines      bool `json:"trimFinalNewlines,omitempty"`
}

// Folding Range Types

// FoldingRangeParams contains parameters for folding range requests
type FoldingRangeParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
}

// FoldingRange represents a folding range
type FoldingRange struct {
	StartLine      int    `json:"startLine"`
	StartCharacter *int   `json:"startCharacter,omitempty"`
	EndLine        int    `json:"endLine"`
	EndCharacter   *int   `json:"endCharacter,omitempty"`
	Kind           string `json:"kind,omitempty"`
}

// FoldingRangeKind defines well-known folding range kinds
const (
	FoldingRangeKindComment = "comment"
	FoldingRangeKindImports = "imports"
	FoldingRangeKindRegion  = "region"
)

// Document Highlight Types

// DocumentHighlightParams contains parameters for document highlight requests
type DocumentHighlightParams struct {
	TextDocumentPositionParams
}

// DocumentHighlight represents a document highlight
type DocumentHighlight struct {
	Range Range                 `json:"range"`
	Kind  DocumentHighlightKind `json:"kind,omitempty"`
}

// DocumentHighlightKind represents the kind of a document highlight
type DocumentHighlightKind int

const (
	DocumentHighlightKindText  DocumentHighlightKind = 1
	DocumentHighlightKindRead  DocumentHighlightKind = 2
	DocumentHighlightKindWrite DocumentHighlightKind = 3
)

// Semantic Tokens Types

// SemanticTokensParams contains parameters for semantic tokens requests
type SemanticTokensParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
}

// SemanticTokens represents semantic tokens
type SemanticTokens struct {
	ResultID string `json:"resultId,omitempty"`
	Data     []int  `json:"data"`
}

// SemanticTokensLegend defines the legend for semantic tokens
type SemanticTokensLegend struct {
	TokenTypes     []string `json:"tokenTypes"`
	TokenModifiers []string `json:"tokenModifiers"`
}

// SemanticTokensOptions defines options for semantic tokens
type SemanticTokensOptions struct {
	Legend SemanticTokensLegend `json:"legend"`
	Range  bool                 `json:"range,omitempty"`
	Full   interface{}          `json:"full,omitempty"`
}

// Well-known semantic token types
const (
	SemanticTokenTypeNamespace        = "namespace"
	SemanticTokenTypeType             = "type"
	SemanticTokenTypeClass            = "class"
	SemanticTokenTypeEnum             = "enum"
	SemanticTokenTypeInterface        = "interface"
	SemanticTokenTypeStruct           = "struct"
	SemanticTokenTypeTypeParameter    = "typeParameter"
	SemanticTokenTypeParameter        = "parameter"
	SemanticTokenTypeVariable         = "variable"
	SemanticTokenTypeProperty         = "property"
	SemanticTokenTypeEnumMember       = "enumMember"
	SemanticTokenTypeEvent            = "event"
	SemanticTokenTypeFunction         = "function"
	SemanticTokenTypeMethod           = "method"
	SemanticTokenTypeMacro            = "macro"
	SemanticTokenTypeKeyword          = "keyword"
	SemanticTokenTypeModifier         = "modifier"
	SemanticTokenTypeComment          = "comment"
	SemanticTokenTypeString           = "string"
	SemanticTokenTypeNumber           = "number"
	SemanticTokenTypeRegexp           = "regexp"
	SemanticTokenTypeOperator         = "operator"
)

// Well-known semantic token modifiers
const (
	SemanticTokenModifierDeclaration    = "declaration"
	SemanticTokenModifierDefinition     = "definition"
	SemanticTokenModifierReadonly       = "readonly"
	SemanticTokenModifierStatic         = "static"
	SemanticTokenModifierDeprecated     = "deprecated"
	SemanticTokenModifierAbstract       = "abstract"
	SemanticTokenModifierAsync          = "async"
	SemanticTokenModifierModification   = "modification"
	SemanticTokenModifierDocumentation  = "documentation"
	SemanticTokenModifierDefaultLibrary = "defaultLibrary"
)
