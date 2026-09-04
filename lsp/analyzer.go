package lsp

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/Advik-B/english/ast"
	"github.com/Advik-B/english/help"
	"github.com/Advik-B/english/parser"
	"github.com/Advik-B/english/sema"
	"github.com/Advik-B/english/stdlib"
	"github.com/Advik-B/english/token"
)

// SymbolType represents the type of a symbol
type SymbolType int

const (
	SymbolTypeVariable SymbolType = iota
	SymbolTypeConstant
	SymbolTypeFunction
	SymbolTypeParameter
)

// Symbol represents a symbol in the document
type Symbol struct {
	Name     string
	Type     SymbolType
	Range    Range
	DefRange Range // The range of just the name in the definition
	Detail   string
	Children []*Symbol
}

// Reference represents a reference to a symbol
type Reference struct {
	Name         string
	Range        Range
	IsDefinition bool
}

// AnalysisResult contains the result of analyzing a document
type AnalysisResult struct {
	Program     *ast.Program
	Tokens      []token.Token
	Symbols     []*Symbol
	References  []*Reference
	Diagnostics []Diagnostic
	Functions   map[string]*FunctionInfo
	Variables   map[string]*VariableInfo
}

// FunctionInfo contains information about a function
type FunctionInfo struct {
	Name          string
	Parameters    []string
	Range         Range
	DefRange      Range
	Body          []ast.Statement
	Documentation string
}

// VariableInfo contains information about a variable
type VariableInfo struct {
	Name       string
	IsConstant bool
	Range      Range
	DefRange   Range
	Value      string
}

// Analyzer analyzes English language documents
type Analyzer struct {
	// knowledge is the same help registry that backs "english help-topic",
	// so the editor and the command line describe the language identically
	// rather than from two hand-written lists.
	knowledge *help.Registry
}

// NewAnalyzer creates a new analyzer
func NewAnalyzer() *Analyzer {
	return &Analyzer{knowledge: help.NewRegistry()}
}

// Analyze analyzes a document and returns the analysis result
func (a *Analyzer) Analyze(doc *Document) *AnalysisResult {
	result := &AnalysisResult{
		Symbols:     make([]*Symbol, 0),
		References:  make([]*Reference, 0),
		Diagnostics: make([]Diagnostic, 0),
		Functions:   make(map[string]*FunctionInfo),
		Variables:   make(map[string]*VariableInfo),
	}

	// Tokenize
	lexer := parser.NewLexer(doc.Content)
	result.Tokens = a.tokenizeAll(lexer)

	// Parse
	p := parser.NewParser(result.Tokens)
	program, err := p.Parse()
	if err != nil {
		// Every syntax error the parse found, not just the first: the editor
		// used to show one, so a file with two typos needed two round trips.
		for _, syntaxErr := range parser.Errors(err) {
			result.Diagnostics = append(result.Diagnostics, a.parseErrorToDiagnostic(syntaxErr, doc))
		}
		if len(result.Diagnostics) == 0 {
			result.Diagnostics = append(result.Diagnostics, a.parseErrorToDiagnostic(err, doc))
		}
		return result
	}
	result.Program = program

	// Report the same problems the compiler would. The editor previously saw
	// only syntax errors, so a type error showed up for the first time when
	// the program was run.
	for _, d := range sema.Check(program, sema.Config{Predefined: stdlib.PredefinedNames()}) {
		result.Diagnostics = append(result.Diagnostics, semaDiagnostic(d, doc))
	}

	// Extract symbols and references
	a.extractSymbols(program, result, doc)

	return result
}

// semaDiagnostic converts a semantic-analysis problem into an editor
// diagnostic, using the position the analyser recorded rather than searching
// the document text for it.
func semaDiagnostic(d *sema.Diagnostic, doc *Document) Diagnostic {
	// Editor positions are zero-based; the analyser's are one-based.
	line := d.Pos.Line - 1
	if line < 0 {
		line = 0
	}
	col := d.Pos.Col - 1
	if col < 0 {
		col = 0
	}

	message := d.Message
	if d.Hint != "" {
		message += "\n" + d.Hint
	}
	return Diagnostic{
		Range:    Range{Start: Position{Line: line, Character: col}, End: wordEnd(doc, line, col)},
		Severity: DiagnosticSeverityError,
		Source:   "english",
		Message:  message,
	}
}

// tokenizeAll returns all tokens with NEWLINE tokens stripped, matching the
// token stream tokeniser.TokenizeAll hands the parser.
func (a *Analyzer) tokenizeAll(lexer *parser.Lexer) []token.Token {
	var tokens []token.Token
	for {
		tok := lexer.NextToken()
		if tok.Type != token.NEWLINE {
			tokens = append(tokens, tok)
		}
		if tok.Type == token.EOF {
			break
		}
	}
	return tokens
}

// parseErrorToDiagnostic converts a parse error into an editor diagnostic.
//
// The parser reports a *parser.SyntaxError carrying the line, the column and a
// hint. This used to ignore all of that and scan the *rendered* message for the
// substring "at line ", re-parsing the digits by hand — so any rewording of the
// message silently moved every syntax error to line 1, column 1. It also ended
// the underline at a fixed column + 10, which ran past the end of short lines.
func (a *Analyzer) parseErrorToDiagnostic(err error, doc *Document) Diagnostic {
	line, col := 0, 0
	message := err.Error()

	var syntaxErr *parser.SyntaxError
	if errors.As(err, &syntaxErr) {
		line = syntaxErr.Line - 1
		col = syntaxErr.Col - 1
		message = syntaxErr.Msg
		if syntaxErr.Hint != "" {
			message += "\n" + syntaxErr.Hint
		}
	}
	if line < 0 {
		line = 0
	}
	if col < 0 {
		col = 0
	}

	return Diagnostic{
		Range:    Range{Start: Position{Line: line, Character: col}, End: wordEnd(doc, line, col)},
		Severity: DiagnosticSeverityError,
		Source:   "english",
		Message:  message,
	}
}

// wordEnd returns the position just past the word starting at col, clamped to
// the end of the line, so an underline never extends past the text.
func wordEnd(doc *Document, line, col int) Position {
	text := doc.GetLine(line)
	end := col
	for end < len(text) && isWordChar(text[end]) {
		end++
	}
	if end == col {
		// Nothing word-like starts here, so underline a single character.
		// An empty range renders as nothing at all in an editor.
		end = col + 1
	}
	return Position{Line: line, Character: end}
}

// extractSymbols extracts symbols from the AST
func (a *Analyzer) extractSymbols(program *ast.Program, result *AnalysisResult, doc *Document) {
	for _, stmt := range program.Statements {
		a.extractFromStatement(stmt, result, doc, nil)
	}
}

// extractFromStatement collects the symbols a statement defines and the names
// it mentions.
//
// Every statement kind is handled. It used to handle 10 of the 24, so anything
// written inside a try block, a struct declaration or method, a raise, a swap,
// a lookup-table assignment or a field assignment was invisible to the editor:
// no references, no rename, no go-to-definition, and no document symbol.
func (a *Analyzer) extractFromStatement(stmt ast.Statement, result *AnalysisResult, doc *Document, parent *Symbol) {
	switch s := stmt.(type) {
	case *ast.VariableDecl:
		detail := "variable"
		if s.IsConstant {
			detail = "constant"
		}
		a.declareVariable(s.Name, s.IsConstant, detail+": "+a.exprToString(s.Value),
			a.exprToString(s.Value), s.Pos(), result, doc, parent)
		a.extractReferencesFromExpr(s.Value, result, doc)

	case *ast.TypedVariableDecl:
		// An annotated declaration is a declaration: the editor showed nothing
		// for one, so "Declare total as a number." had no symbol at all.
		kind := "variable"
		if s.IsConstant {
			kind = "constant"
		}
		value := "?"
		if s.Value != nil {
			value = a.exprToString(s.Value)
		}
		a.declareVariable(s.Name, s.IsConstant,
			kind+": "+ast.TypeName(s.Type), value, s.Pos(), result, doc, parent)
		a.extractReferencesFromExpr(s.Value, result, doc)

	case *ast.FunctionDecl:
		a.extractFunction(s, result, doc, parent)

	case *ast.Assignment:
		a.reference(s.Name, s.Pos(), result, doc)
		a.extractReferencesFromExpr(s.Value, result, doc)

	case *ast.IfStatement:
		a.extractReferencesFromExpr(s.Condition, result, doc)
		a.extractFromBody(s.Then, result, doc, parent)
		for _, elseIf := range s.ElseIf {
			a.extractReferencesFromExpr(elseIf.Condition, result, doc)
			a.extractFromBody(elseIf.Body, result, doc, parent)
		}
		a.extractFromBody(s.Else, result, doc, parent)

	case *ast.WhileLoop:
		a.extractReferencesFromExpr(s.Condition, result, doc)
		a.extractFromBody(s.Body, result, doc, parent)

	case *ast.ForLoop:
		a.extractReferencesFromExpr(s.Count, result, doc)
		a.extractFromBody(s.Body, result, doc, parent)

	case *ast.ForEachLoop:
		// The loop variable is introduced here, so it is a definition; without
		// it, renaming the loop variable renamed only its uses.
		a.define(s.Item, s.Pos(), result, doc)
		a.extractReferencesFromExpr(s.List, result, doc)
		a.extractFromBody(s.Body, result, doc, parent)

	case *ast.OutputStatement:
		for _, value := range s.Values {
			a.extractReferencesFromExpr(value, result, doc)
		}

	case *ast.ReturnStatement:
		a.extractReferencesFromExpr(s.Value, result, doc)

	case *ast.CallStatement:
		if s.FunctionCall != nil {
			a.extractReferencesFromExpr(s.FunctionCall, result, doc)
		}
		if s.MethodCall != nil {
			a.extractReferencesFromExpr(s.MethodCall, result, doc)
		}

	case *ast.IndexAssignment:
		a.reference(s.ListName, s.Pos(), result, doc)
		a.extractReferencesFromExpr(s.Index, result, doc)
		a.extractReferencesFromExpr(s.Value, result, doc)

	case *ast.LookupKeyAssignment:
		a.reference(s.TableName, s.Pos(), result, doc)
		a.extractReferencesFromExpr(s.Key, result, doc)
		a.extractReferencesFromExpr(s.Value, result, doc)

	case *ast.FieldAssignment:
		a.reference(s.ObjectName, s.Pos(), result, doc)
		a.extractReferencesFromExpr(s.Value, result, doc)

	case *ast.ToggleStatement:
		a.reference(s.Name, s.Pos(), result, doc)

	case *ast.SwapStatement:
		a.reference(s.Name1, s.Pos(), result, doc)
		a.reference(s.Name2, s.Pos(), result, doc)

	case *ast.StructDecl:
		a.extractStruct(s, result, doc, parent)

	case *ast.TryStatement:
		a.extractFromBody(s.TryBody, result, doc, parent)
		if s.ErrorVar != "" {
			a.define(s.ErrorVar, s.Pos(), result, doc)
		}
		a.extractFromBody(s.ErrorBody, result, doc, parent)
		a.extractFromBody(s.FinallyBody, result, doc, parent)

	case *ast.RaiseStatement:
		a.extractReferencesFromExpr(s.Message, result, doc)

	case *ast.ImportStatement, *ast.ErrorTypeDecl, *ast.BreakStatement,
		*ast.ContinueStatement, *ast.CommentStatement:
		// Nothing here names a variable or a function.
	}
}

// extractFromBody walks a block of statements.
func (a *Analyzer) extractFromBody(body []ast.Statement, result *AnalysisResult, doc *Document, parent *Symbol) {
	for _, stmt := range body {
		a.extractFromStatement(stmt, result, doc, parent)
	}
}

// extractFunction records a function, its parameters and its body.
func (a *Analyzer) extractFunction(f *ast.FunctionDecl, result *AnalysisResult, doc *Document, parent *Symbol) {
	sym := a.createFunctionSymbol(f, doc)
	a.attach(sym, result, parent)

	result.Functions[f.Name] = &FunctionInfo{
		Name:          f.Name,
		Parameters:    f.ParamNames(),
		Range:         sym.Range,
		DefRange:      sym.DefRange,
		Body:          f.Body,
		Documentation: a.generateFunctionDoc(f),
	}
	result.References = append(result.References, &Reference{
		Name:         f.Name,
		Range:        sym.DefRange,
		IsDefinition: true,
	})

	// Parameters are declarations too, so hovering or renaming one works.
	for i := range f.Params {
		param := &f.Params[i]
		paramRange := a.nameRangeAt(param.Pos(), param.Name, doc)
		sym.Children = append(sym.Children, &Symbol{
			Name:     param.Name,
			Type:     SymbolTypeParameter,
			Range:    paramRange,
			DefRange: paramRange,
			Detail:   "parameter: " + ast.TypeName(param.Type),
		})
		result.References = append(result.References, &Reference{
			Name:         param.Name,
			Range:        paramRange,
			IsDefinition: true,
		})
	}

	a.extractFromBody(f.Body, result, doc, sym)
}

// extractStruct records a struct, its fields and its methods.
func (a *Analyzer) extractStruct(sd *ast.StructDecl, result *AnalysisResult, doc *Document, parent *Symbol) {
	structRange := a.nameRangeAt(sd.Pos(), sd.Name, doc)
	sym := &Symbol{
		Name:     sd.Name,
		Type:     SymbolTypeVariable,
		Range:    structRange,
		DefRange: structRange,
		Detail:   "structure",
		Children: make([]*Symbol, 0, len(sd.Fields)),
	}
	a.attach(sym, result, parent)
	result.References = append(result.References, &Reference{
		Name:         sd.Name,
		Range:        structRange,
		IsDefinition: true,
	})

	for _, field := range sd.Fields {
		fieldRange := a.nameRangeAt(field.Pos(), field.Name, doc)
		sym.Children = append(sym.Children, &Symbol{
			Name:     field.Name,
			Type:     SymbolTypeVariable,
			Range:    fieldRange,
			DefRange: fieldRange,
			Detail:   "field: " + ast.TypeName(field.Type),
		})
		a.extractReferencesFromExpr(field.DefaultValue, result, doc)
	}

	for _, method := range sd.Methods {
		a.extractFunction(method, result, doc, sym)
	}
}

// declareVariable records a variable declaration as a symbol, a reference and
// an entry in the variables map.
func (a *Analyzer) declareVariable(name string, isConstant bool, detail, value string,
	pos ast.Position, result *AnalysisResult, doc *Document, parent *Symbol) {
	symType := SymbolTypeVariable
	if isConstant {
		symType = SymbolTypeConstant
	}
	nameRange := a.nameRangeAt(pos, name, doc)
	sym := &Symbol{
		Name:     name,
		Type:     symType,
		Range:    nameRange,
		DefRange: nameRange,
		Detail:   detail,
	}
	a.attach(sym, result, parent)

	result.Variables[name] = &VariableInfo{
		Name:       name,
		IsConstant: isConstant,
		Range:      nameRange,
		DefRange:   nameRange,
		Value:      value,
	}
	result.References = append(result.References, &Reference{
		Name:         name,
		Range:        nameRange,
		IsDefinition: true,
	})
}

// attach files a symbol under its enclosing symbol, or at the top level.
func (a *Analyzer) attach(sym *Symbol, result *AnalysisResult, parent *Symbol) {
	if parent != nil {
		parent.Children = append(parent.Children, sym)
		return
	}
	result.Symbols = append(result.Symbols, sym)
}

// reference records a use of a name at a node's position.
func (a *Analyzer) reference(name string, pos ast.Position, result *AnalysisResult, doc *Document) {
	result.References = append(result.References, &Reference{
		Name:  name,
		Range: a.nameRangeAt(pos, name, doc),
	})
}

// define records a name introduced by a binding construct — a loop variable or
// a caught error — which is a definition rather than a use.
func (a *Analyzer) define(name string, pos ast.Position, result *AnalysisResult, doc *Document) {
	result.References = append(result.References, &Reference{
		Name:         name,
		Range:        a.nameRangeAt(pos, name, doc),
		IsDefinition: true,
	})
}

// extractReferencesFromExpr collects the names an expression mentions.
//
// Every expression kind is handled. It used to handle 8 of about 27, so a name
// used in a cast, a copy, a range, an array or lookup-table literal, a struct
// instantiation, a field access, a method call, an "ask", a "has" or a nothing
// check was invisible.
func (a *Analyzer) extractReferencesFromExpr(expr ast.Expression, result *AnalysisResult, doc *Document) {
	if expr == nil {
		return
	}

	switch e := expr.(type) {
	case *ast.Identifier:
		a.reference(e.Name, e.Pos(), result, doc)

	case *ast.LocationExpression:
		a.reference(e.Name, e.Pos(), result, doc)

	case *ast.ReferenceExpression:
		a.reference(e.Name, e.Pos(), result, doc)

	case *ast.FunctionCall:
		a.reference(e.Name, e.Pos(), result, doc)
		a.extractReferencesFromExprs(e.Arguments, result, doc)

	case *ast.MethodCall:
		a.extractReferencesFromExpr(e.Object, result, doc)
		a.extractReferencesFromExprs(e.Arguments, result, doc)

	case *ast.BinaryExpression:
		a.extractReferencesFromExpr(e.Left, result, doc)
		a.extractReferencesFromExpr(e.Right, result, doc)

	case *ast.UnaryExpression:
		a.extractReferencesFromExpr(e.Right, result, doc)

	case *ast.IndexExpression:
		a.extractReferencesFromExpr(e.List, result, doc)
		a.extractReferencesFromExpr(e.Index, result, doc)

	case *ast.LengthExpression:
		a.extractReferencesFromExpr(e.List, result, doc)

	case *ast.ListLiteral:
		a.extractReferencesFromExprs(e.Elements, result, doc)

	case *ast.ArrayLiteral:
		a.extractReferencesFromExprs(e.Elements, result, doc)

	case *ast.RangeLiteral:
		a.extractReferencesFromExpr(e.Start, result, doc)
		a.extractReferencesFromExpr(e.End, result, doc)
		a.extractReferencesFromExpr(e.Step, result, doc)

	case *ast.StructInstantiation:
		a.reference(e.StructName, e.Pos(), result, doc)
		// FieldOrder, not the map, so the references come out in source order
		// rather than in Go's randomised map order.
		for _, name := range e.FieldOrder {
			a.extractReferencesFromExpr(e.FieldValues[name], result, doc)
		}

	case *ast.FieldAccess:
		a.extractReferencesFromExpr(e.Object, result, doc)

	case *ast.CastExpression:
		a.extractReferencesFromExpr(e.Value, result, doc)

	case *ast.CopyExpression:
		a.extractReferencesFromExpr(e.Value, result, doc)

	case *ast.TypeExpression:
		a.extractReferencesFromExpr(e.Value, result, doc)

	case *ast.AskExpression:
		a.extractReferencesFromExpr(e.Prompt, result, doc)

	case *ast.LookupKeyAccess:
		a.extractReferencesFromExpr(e.Table, result, doc)
		a.extractReferencesFromExpr(e.Key, result, doc)

	case *ast.HasExpression:
		a.extractReferencesFromExpr(e.Table, result, doc)
		a.extractReferencesFromExpr(e.Key, result, doc)

	case *ast.NilCheckExpression:
		a.extractReferencesFromExpr(e.Value, result, doc)

	case *ast.ErrorTypeCheckExpression:
		a.extractReferencesFromExpr(e.Value, result, doc)

	case *ast.NumberLiteral, *ast.StringLiteral, *ast.BooleanLiteral,
		*ast.NothingLiteral, *ast.LookupTableLiteral:
		// A literal names nothing.
	}
}

// extractReferencesFromExprs walks a list of expressions.
func (a *Analyzer) extractReferencesFromExprs(exprs []ast.Expression, result *AnalysisResult, doc *Document) {
	for _, expr := range exprs {
		a.extractReferencesFromExpr(expr, result, doc)
	}
}

// createFunctionSymbol creates a symbol for a function declaration
func (a *Analyzer) createFunctionSymbol(f *ast.FunctionDecl, doc *Document) *Symbol {
	nameRange := a.nameRangeAt(f.Pos(), f.Name, doc)

	params := strings.Join(f.ParamNames(), ", ")
	detail := "function"
	if len(f.ParamNames()) > 0 {
		detail = "function(" + params + ")"
	}

	return &Symbol{
		Name:     f.Name,
		Type:     SymbolTypeFunction,
		Range:    nameRange,
		DefRange: nameRange,
		Detail:   detail,
		Children: make([]*Symbol, 0),
	}
}

// nameRangeAt locates a name in the document, starting from the position of
// the node that mentions it.
//
// This used to scan the whole document and return the *first* whole-word match
// of the name, ignoring which node was being asked about. So go-to-definition
// on the fifth use of a variable jumped to the first occurrence anywhere in
// the file — including inside a string or a comment, since nothing filtered
// those — and find-all-references returned one identical range repeated once
// per reference, which made rename unusable.
//
// Anchoring the search at the node's own position means the answer can only be
// at or after where that node begins, so it cannot land on an unrelated
// occurrence. A node records where it starts, which for a statement is its
// keyword rather than the name, so the name is found forward from there.
func (a *Analyzer) nameRangeAt(pos ast.Position, name string, doc *Document) Range {
	// Positions are 1-based in the AST and 0-based in the protocol.
	startLine := pos.Line - 1
	startCol := pos.Col - 1
	if startLine < 0 {
		startLine, startCol = 0, 0
	}
	if startCol < 0 {
		startCol = 0
	}

	// A statement rarely spans more than a handful of lines; searching a few
	// past the node keeps a multi-line construct working without reopening
	// the whole-document scan.
	const lookahead = 8
	for line := startLine; line < len(doc.Lines) && line <= startLine+lookahead; line++ {
		from := 0
		if line == startLine {
			from = startCol
		}
		if idx := wholeWordIndex(doc.Lines[line], name, from); idx >= 0 {
			return Range{
				Start: Position{Line: line, Character: idx},
				End:   Position{Line: line, Character: idx + len(name)},
			}
		}
	}

	// Nothing found: point at the node itself rather than at the top of the
	// file, which is where an empty range would send the editor.
	return Range{
		Start: Position{Line: startLine, Character: startCol},
		End:   Position{Line: startLine, Character: startCol + len(name)},
	}
}

// wholeWordIndex returns the index of the first whole-word occurrence of name
// in line at or after from, or -1.
func wholeWordIndex(line, name string, from int) int {
	if name == "" || from > len(line) {
		return -1
	}
	for at := from; ; {
		idx := strings.Index(line[at:], name)
		if idx < 0 {
			return -1
		}
		idx += at
		before := idx == 0 || !isWordChar(line[idx-1])
		after := idx+len(name) >= len(line) || !isWordChar(line[idx+len(name)])
		if before && after {
			return idx
		}
		at = idx + 1
	}
}

// exprToString converts an expression to a string representation
func (a *Analyzer) exprToString(expr ast.Expression) string {
	if expr == nil {
		return "?"
	}

	switch e := expr.(type) {
	case *ast.NumberLiteral:
		if e.Value == float64(int64(e.Value)) {
			return fmt.Sprintf("%d", int64(e.Value))
		}
		return fmt.Sprintf("%g", e.Value)
	case *ast.StringLiteral:
		return `"` + e.Value + `"`
	case *ast.BooleanLiteral:
		if e.Value {
			return "true"
		}
		return "false"
	case *ast.Identifier:
		return e.Name
	case *ast.ListLiteral:
		return "list"
	case *ast.FunctionCall:
		return e.Name + "(...)"
	case *ast.BinaryExpression:
		return a.exprToString(e.Left) + " " + e.Operator + " " + a.exprToString(e.Right)
	default:
		return "expression"
	}
}

// generateFunctionDoc generates documentation for a function
func (a *Analyzer) generateFunctionDoc(f *ast.FunctionDecl) string {
	var doc strings.Builder
	doc.WriteString("**")
	doc.WriteString(f.Name)
	doc.WriteString("**\n\n")

	if len(f.ParamNames()) > 0 {
		doc.WriteString("Parameters:\n")
		for _, param := range f.ParamNames() {
			doc.WriteString("- `")
			doc.WriteString(param)
			doc.WriteString("`\n")
		}
	} else {
		doc.WriteString("Takes no parameters.\n")
	}

	return doc.String()
}

// GetCompletions returns completion items at the given position
func (a *Analyzer) GetCompletions(doc *Document, pos Position, result *AnalysisResult) []CompletionItem {
	items := make([]CompletionItem, 0)

	prefix := strings.ToLower(completionPrefixAtPosition(doc, pos))

	// Add keyword completions
	items = append(items, a.getKeywordCompletions(prefix)...)

	// Add the standard library, which the editor previously knew nothing about.
	items = append(items, a.getBuiltinCompletions(prefix)...)

	// Add variable completions
	for name, info := range result.Variables {
		if prefix == "" || strings.HasPrefix(strings.ToLower(name), prefix) {
			kind := CompletionItemKindVariable
			if info.IsConstant {
				kind = CompletionItemKindConstant
			}
			items = append(items, CompletionItem{
				Label:  name,
				Kind:   kind,
				Detail: info.Value,
				Documentation: MarkupContent{
					Kind:  MarkupKindMarkdown,
					Value: fmt.Sprintf("Variable `%s`", name),
				},
			})
		}
	}

	// Add function completions
	for name, info := range result.Functions {
		if prefix == "" || strings.HasPrefix(strings.ToLower(name), prefix) {
			items = append(items, CompletionItem{
				Label:  name,
				Kind:   CompletionItemKindFunction,
				Detail: "function(" + strings.Join(info.Parameters, ", ") + ")",
				Documentation: MarkupContent{
					Kind:  MarkupKindMarkdown,
					Value: info.Documentation,
				},
			})
		}
	}

	return normalizeCompletionItems(items)
}

func completionPrefixAtPosition(doc *Document, pos Position) string {
	line := doc.GetLine(pos.Line)
	if line == "" {
		return ""
	}
	if pos.Character < 0 {
		return ""
	}
	if pos.Character > len(line) {
		pos.Character = len(line)
	}
	start := pos.Character
	for start > 0 && isWordChar(line[start-1]) {
		start--
	}
	return line[start:pos.Character]
}

func normalizeCompletionItems(items []CompletionItem) []CompletionItem {
	seen := make(map[string]struct{}, len(items))
	out := make([]CompletionItem, 0, len(items))
	for _, item := range items {
		key := strings.ToLower(item.Label) + ":" + strconv.Itoa(int(item.Kind))
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		if item.FilterText == "" {
			item.FilterText = item.Label
		}
		if item.InsertText == "" {
			item.InsertText = item.Label
		}
		out = append(out, item)
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Kind != out[j].Kind {
			return out[i].Kind < out[j].Kind
		}
		return strings.ToLower(out[i].Label) < strings.ToLower(out[j].Label)
	})
	return out
}

// GetHover returns hover information at the given position
func (a *Analyzer) GetHover(doc *Document, pos Position, result *AnalysisResult) *Hover {
	word, wordRange := doc.GetWordAtPosition(pos)
	if word == "" {
		return nil
	}

	// Check if it's a variable
	if info, ok := result.Variables[word]; ok {
		kind := "variable"
		if info.IsConstant {
			kind = "constant"
		}
		return &Hover{
			Contents: MarkupContent{
				Kind:  MarkupKindMarkdown,
				Value: fmt.Sprintf("**%s** `%s`\n\nValue: `%s`", kind, word, info.Value),
			},
			Range: &wordRange,
		}
	}

	// Check if it's a function
	if info, ok := result.Functions[word]; ok {
		return &Hover{
			Contents: MarkupContent{
				Kind:  MarkupKindMarkdown,
				Value: info.Documentation,
			},
			Range: &wordRange,
		}
	}

	// Check if it's a keyword
	if doc := a.getKeywordDocumentation(word); doc != "" {
		return &Hover{
			Contents: MarkupContent{
				Kind:  MarkupKindMarkdown,
				Value: doc,
			},
			Range: &wordRange,
		}
	}

	return nil
}

// GetDefinition returns the definition location for a symbol at the given position
func (a *Analyzer) GetDefinition(doc *Document, pos Position, result *AnalysisResult) *Location {
	word, _ := doc.GetWordAtPosition(pos)
	if word == "" {
		return nil
	}

	// Check variables
	if info, ok := result.Variables[word]; ok {
		return &Location{
			URI:   doc.URI,
			Range: info.DefRange,
		}
	}

	// Check functions
	if info, ok := result.Functions[word]; ok {
		return &Location{
			URI:   doc.URI,
			Range: info.DefRange,
		}
	}

	return nil
}

// GetReferences returns all references to a symbol at the given position
func (a *Analyzer) GetReferences(doc *Document, pos Position, result *AnalysisResult, includeDeclaration bool) []Location {
	word, _ := doc.GetWordAtPosition(pos)
	if word == "" {
		return nil
	}

	locations := make([]Location, 0)
	for _, ref := range result.References {
		if ref.Name == word {
			if !includeDeclaration && ref.IsDefinition {
				continue
			}
			locations = append(locations, Location{
				URI:   doc.URI,
				Range: ref.Range,
			})
		}
	}

	return locations
}

// GetDocumentSymbols returns all symbols in the document
func (a *Analyzer) GetDocumentSymbols(result *AnalysisResult) []DocumentSymbol {
	symbols := make([]DocumentSymbol, 0, len(result.Symbols))

	for _, sym := range result.Symbols {
		kind := SymbolKindVariable
		switch sym.Type {
		case SymbolTypeConstant:
			kind = SymbolKindConstant
		case SymbolTypeFunction:
			kind = SymbolKindFunction
		case SymbolTypeParameter:
			kind = SymbolKindVariable
		}

		docSym := DocumentSymbol{
			Name:           sym.Name,
			Detail:         sym.Detail,
			Kind:           kind,
			Range:          sym.Range,
			SelectionRange: sym.DefRange,
		}

		// Add children
		if len(sym.Children) > 0 {
			docSym.Children = make([]DocumentSymbol, 0, len(sym.Children))
			for _, child := range sym.Children {
				childKind := SymbolKindVariable
				if child.Type == SymbolTypeConstant {
					childKind = SymbolKindConstant
				}
				docSym.Children = append(docSym.Children, DocumentSymbol{
					Name:           child.Name,
					Detail:         child.Detail,
					Kind:           childKind,
					Range:          child.Range,
					SelectionRange: child.DefRange,
				})
			}
		}

		symbols = append(symbols, docSym)
	}

	return symbols
}

// GetSignatureHelp returns signature help for a function call at the given position
func (a *Analyzer) GetSignatureHelp(doc *Document, pos Position, result *AnalysisResult) *SignatureHelp {
	// Look backwards for a function name
	line := doc.GetLine(pos.Line)
	if pos.Character > len(line) {
		return nil
	}

	// Find the opening parenthesis or function call context
	// In English, function calls look like: "the result of calling FuncName with arg1 and arg2"
	lineBeforeCursor := line[:pos.Character]

	// Look for "calling " pattern
	callingIdx := strings.LastIndex(strings.ToLower(lineBeforeCursor), "calling ")
	if callingIdx == -1 {
		return nil
	}

	// Extract function name after "calling "
	afterCalling := lineBeforeCursor[callingIdx+8:]
	// Scanning bytes, not runes: isWordChar takes a byte, and truncating a
	// multi-byte rune to its low byte can land on a letter, which would splice
	// half a character into the name.
	funcName := ""
	for i := 0; i < len(afterCalling); i++ {
		if !isWordChar(afterCalling[i]) {
			break
		}
		funcName += string(afterCalling[i])
	}

	if funcName == "" {
		return nil
	}

	// Look up the function
	funcInfo, ok := result.Functions[funcName]
	if !ok {
		return nil
	}

	// Count "and" to determine which parameter we're on
	withIdx := strings.Index(strings.ToLower(afterCalling), " with ")
	activeParam := 0
	if withIdx != -1 {
		// Count "and" occurrences after "with"
		afterWith := afterCalling[withIdx+6:]
		activeParam = strings.Count(strings.ToLower(afterWith), " and ")
	}

	// Build signature
	paramLabels := make([]ParameterInformation, 0, len(funcInfo.Parameters))
	for _, param := range funcInfo.Parameters {
		paramLabels = append(paramLabels, ParameterInformation{
			Label: param,
		})
	}

	sig := SignatureInformation{
		Label: funcName + "(" + strings.Join(funcInfo.Parameters, ", ") + ")",
		Documentation: MarkupContent{
			Kind:  MarkupKindMarkdown,
			Value: funcInfo.Documentation,
		},
		Parameters: paramLabels,
	}

	if activeParam < len(funcInfo.Parameters) {
		sig.ActiveParameter = &activeParam
	}

	return &SignatureHelp{
		Signatures:      []SignatureInformation{sig},
		ActiveSignature: intPtr(0),
		ActiveParameter: &activeParam,
	}
}

// intPtr returns a pointer to an int
func intPtr(i int) *int {
	return &i
}
