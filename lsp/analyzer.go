package lsp

import (
	"english/ast"
	"english/parser"
	"english/token"
	"fmt"
	"strings"
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
	Name  string
	Range Range
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
	Name       string
	Parameters []string
	Range      Range
	DefRange   Range
	Body       []ast.Statement
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
type Analyzer struct{}

// NewAnalyzer creates a new analyzer
func NewAnalyzer() *Analyzer {
	return &Analyzer{}
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
		// Add parse error as diagnostic
		diag := a.parseErrorToDiagnostic(err.Error(), doc)
		result.Diagnostics = append(result.Diagnostics, diag)
		return result
	}
	result.Program = program

	// Extract symbols and references
	a.extractSymbols(program, result, doc)

	return result
}

// tokenizeAll returns all tokens including newlines
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

// parseErrorToDiagnostic converts a parse error to a diagnostic
func (a *Analyzer) parseErrorToDiagnostic(errMsg string, doc *Document) Diagnostic {
	// Try to extract line and column from error message
	line := 0
	col := 0

	// Look for "at line X, column Y" pattern
	if idx := strings.Index(errMsg, "at line "); idx != -1 {
		// Parse line number
		remaining := errMsg[idx+8:]
		for i, c := range remaining {
			if c >= '0' && c <= '9' {
				line = line*10 + int(c-'0')
			} else if c == ',' {
				// Found comma, look for column
				colStr := remaining[i+1:]
				if colIdx := strings.Index(colStr, "column "); colIdx != -1 {
					colPart := colStr[colIdx+7:]
					for _, c := range colPart {
						if c >= '0' && c <= '9' {
							col = col*10 + int(c-'0')
						} else {
							break
						}
					}
				}
				break
			}
		}
	}

	// Convert to 0-indexed
	if line > 0 {
		line--
	}
	if col > 0 {
		col--
	}

	return Diagnostic{
		Range: Range{
			Start: Position{Line: line, Character: col},
			End:   Position{Line: line, Character: col + 10},
		},
		Severity: DiagnosticSeverityError,
		Source:   "english",
		Message:  errMsg,
	}
}

// extractSymbols extracts symbols from the AST
func (a *Analyzer) extractSymbols(program *ast.Program, result *AnalysisResult, doc *Document) {
	for _, stmt := range program.Statements {
		a.extractFromStatement(stmt, result, doc, nil)
	}
}

// extractFromStatement extracts symbols from a statement
func (a *Analyzer) extractFromStatement(stmt ast.Statement, result *AnalysisResult, doc *Document, parent *Symbol) {
	switch s := stmt.(type) {
	case *ast.VariableDecl:
		sym := a.createVariableSymbol(s, doc)
		if parent != nil {
			parent.Children = append(parent.Children, sym)
		} else {
			result.Symbols = append(result.Symbols, sym)
		}

		// Add to variables map
		result.Variables[s.Name] = &VariableInfo{
			Name:       s.Name,
			IsConstant: s.IsConstant,
			Range:      sym.Range,
			DefRange:   sym.DefRange,
			Value:      a.exprToString(s.Value),
		}

		// Add reference for the definition
		result.References = append(result.References, &Reference{
			Name:         s.Name,
			Range:        sym.DefRange,
			IsDefinition: true,
		})

		// Extract references from value expression
		a.extractReferencesFromExpr(s.Value, result, doc)

	case *ast.FunctionDecl:
		sym := a.createFunctionSymbol(s, doc)
		if parent != nil {
			parent.Children = append(parent.Children, sym)
		} else {
			result.Symbols = append(result.Symbols, sym)
		}

		// Add to functions map
		result.Functions[s.Name] = &FunctionInfo{
			Name:       s.Name,
			Parameters: s.Parameters,
			Range:      sym.Range,
			DefRange:   sym.DefRange,
			Body:       s.Body,
			Documentation: a.generateFunctionDoc(s),
		}

		// Add reference for the definition
		result.References = append(result.References, &Reference{
			Name:         s.Name,
			Range:        sym.DefRange,
			IsDefinition: true,
		})

		// Extract symbols from function body
		for _, bodyStmt := range s.Body {
			a.extractFromStatement(bodyStmt, result, doc, sym)
		}

	case *ast.Assignment:
		// Add reference for the variable being assigned
		varRange := a.findIdentifierRange(s.Name, doc)
		result.References = append(result.References, &Reference{
			Name:  s.Name,
			Range: varRange,
		})
		// Extract references from value
		a.extractReferencesFromExpr(s.Value, result, doc)

	case *ast.IfStatement:
		a.extractReferencesFromExpr(s.Condition, result, doc)
		for _, thenStmt := range s.Then {
			a.extractFromStatement(thenStmt, result, doc, parent)
		}
		for _, elseIf := range s.ElseIf {
			a.extractReferencesFromExpr(elseIf.Condition, result, doc)
			for _, stmt := range elseIf.Body {
				a.extractFromStatement(stmt, result, doc, parent)
			}
		}
		for _, elseStmt := range s.Else {
			a.extractFromStatement(elseStmt, result, doc, parent)
		}

	case *ast.WhileLoop:
		a.extractReferencesFromExpr(s.Condition, result, doc)
		for _, bodyStmt := range s.Body {
			a.extractFromStatement(bodyStmt, result, doc, parent)
		}

	case *ast.ForLoop:
		a.extractReferencesFromExpr(s.Count, result, doc)
		for _, bodyStmt := range s.Body {
			a.extractFromStatement(bodyStmt, result, doc, parent)
		}

	case *ast.ForEachLoop:
		a.extractReferencesFromExpr(s.List, result, doc)
		for _, bodyStmt := range s.Body {
			a.extractFromStatement(bodyStmt, result, doc, parent)
		}

	case *ast.OutputStatement:
		a.extractReferencesFromExpr(s.Value, result, doc)

	case *ast.ReturnStatement:
		a.extractReferencesFromExpr(s.Value, result, doc)

	case *ast.CallStatement:
		if s.FunctionCall != nil {
			a.extractReferencesFromExpr(s.FunctionCall, result, doc)
		}

	case *ast.IndexAssignment:
		varRange := a.findIdentifierRange(s.ListName, doc)
		result.References = append(result.References, &Reference{
			Name:  s.ListName,
			Range: varRange,
		})
		a.extractReferencesFromExpr(s.Index, result, doc)
		a.extractReferencesFromExpr(s.Value, result, doc)

	case *ast.ToggleStatement:
		varRange := a.findIdentifierRange(s.Name, doc)
		result.References = append(result.References, &Reference{
			Name:  s.Name,
			Range: varRange,
		})
	}
}

// extractReferencesFromExpr extracts variable/function references from an expression
func (a *Analyzer) extractReferencesFromExpr(expr ast.Expression, result *AnalysisResult, doc *Document) {
	if expr == nil {
		return
	}

	switch e := expr.(type) {
	case *ast.Identifier:
		varRange := a.findIdentifierRange(e.Name, doc)
		result.References = append(result.References, &Reference{
			Name:  e.Name,
			Range: varRange,
		})

	case *ast.FunctionCall:
		// Add reference to function
		funcRange := a.findIdentifierRange(e.Name, doc)
		result.References = append(result.References, &Reference{
			Name:  e.Name,
			Range: funcRange,
		})
		// Extract references from arguments
		for _, arg := range e.Arguments {
			a.extractReferencesFromExpr(arg, result, doc)
		}

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
		for _, elem := range e.Elements {
			a.extractReferencesFromExpr(elem, result, doc)
		}

	case *ast.LocationExpression:
		varRange := a.findIdentifierRange(e.Name, doc)
		result.References = append(result.References, &Reference{
			Name:  e.Name,
			Range: varRange,
		})
	}
}

// createVariableSymbol creates a symbol for a variable declaration
func (a *Analyzer) createVariableSymbol(v *ast.VariableDecl, doc *Document) *Symbol {
	symType := SymbolTypeVariable
	detail := "variable"
	if v.IsConstant {
		symType = SymbolTypeConstant
		detail = "constant"
	}

	// Find the range of the declaration in the document
	nameRange := a.findIdentifierRange(v.Name, doc)

	return &Symbol{
		Name:     v.Name,
		Type:     symType,
		Range:    nameRange, // For simple cases, use the name range
		DefRange: nameRange,
		Detail:   detail + ": " + a.exprToString(v.Value),
	}
}

// createFunctionSymbol creates a symbol for a function declaration
func (a *Analyzer) createFunctionSymbol(f *ast.FunctionDecl, doc *Document) *Symbol {
	nameRange := a.findIdentifierRange(f.Name, doc)

	params := strings.Join(f.Parameters, ", ")
	detail := "function"
	if len(f.Parameters) > 0 {
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

// findIdentifierRange finds the range of an identifier in the document
func (a *Analyzer) findIdentifierRange(name string, doc *Document) Range {
	// Simple search - find the identifier in the document
	for lineNum, line := range doc.Lines {
		idx := strings.Index(line, name)
		if idx != -1 {
			// Make sure it's a whole word match
			before := idx == 0 || !isWordChar(line[idx-1])
			after := idx+len(name) >= len(line) || !isWordChar(line[idx+len(name)])
			if before && after {
				return Range{
					Start: Position{Line: lineNum, Character: idx},
					End:   Position{Line: lineNum, Character: idx + len(name)},
				}
			}
		}
	}
	return Range{}
}

// exprToString converts an expression to a string representation
func (a *Analyzer) exprToString(expr ast.Expression) string {
	if expr == nil {
		return "?"
	}

	switch e := expr.(type) {
	case *ast.NumberLiteral:
		if e.Value == float64(int64(e.Value)) {
			return strings.TrimSuffix(strings.TrimSuffix(
				strings.TrimSuffix(string(rune(int64(e.Value)+'0')), "0"),
				"."), "0")
		}
		return "number"
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

	if len(f.Parameters) > 0 {
		doc.WriteString("Parameters:\n")
		for _, param := range f.Parameters {
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

	// Get the word being typed
	word, _ := doc.GetWordAtPosition(pos)
	wordLower := strings.ToLower(word)

	// Add keyword completions
	items = append(items, a.getKeywordCompletions(wordLower)...)

	// Add variable completions
	for name, info := range result.Variables {
		if wordLower == "" || strings.HasPrefix(strings.ToLower(name), wordLower) {
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
		if wordLower == "" || strings.HasPrefix(strings.ToLower(name), wordLower) {
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

	return items
}

// getKeywordCompletions returns keyword completions
func (a *Analyzer) getKeywordCompletions(prefix string) []CompletionItem {
	keywords := []struct {
		label  string
		detail string
		snippet string
	}{
		{"Declare", "Declare a variable", "Declare ${1:name} to be ${2:value}."},
		{"Set", "Assign a value", "Set ${1:name} to be ${2:value}."},
		{"Print", "Print a value", "Print ${1:value}."},
		{"If", "Conditional statement", "If ${1:condition}, then\n\t${2:statements}\nThats it."},
		{"Otherwise", "Else clause", "Otherwise\n\t${1:statements}"},
		{"Repeat", "Loop statement", "Repeat the following ${1:count} times:\n\t${2:statements}\nThats it."},
		{"For", "For-each loop", "For each ${1:item} in ${2:list}, do the following:\n\t${3:statements}\nThats it."},
		{"Call", "Call a function", "Call ${1:function}."},
		{"Return", "Return from function", "Return ${1:value}."},
		{"Break", "Break out of loop", "Break out of the loop."},
		{"Toggle", "Toggle boolean", "Toggle ${1:variable}."},
		{"Declare function", "Declare a function", "Declare function ${1:name} that does the following:\n\t${2:statements}\nThats it."},
		{"true", "Boolean true", "true"},
		{"false", "Boolean false", "false"},
		{"the item at position", "Access list element", "the item at position ${1:index} in ${2:list}"},
		{"the length of", "Get length", "the length of ${1:list}"},
		{"the remainder of", "Modulo operation", "the remainder of ${1:a} divided by ${2:b}"},
		{"is equal to", "Equality comparison", "is equal to"},
		{"is not equal to", "Inequality comparison", "is not equal to"},
		{"is less than", "Less than comparison", "is less than"},
		{"is greater than", "Greater than comparison", "is greater than"},
		{"is less than or equal to", "Less than or equal comparison", "is less than or equal to"},
		{"is greater than or equal to", "Greater than or equal comparison", "is greater than or equal to"},
	}

	items := make([]CompletionItem, 0)
	for _, kw := range keywords {
		if prefix == "" || strings.HasPrefix(strings.ToLower(kw.label), prefix) {
			item := CompletionItem{
				Label:            kw.label,
				Kind:             CompletionItemKindKeyword,
				Detail:           kw.detail,
				InsertText:       kw.snippet,
				InsertTextFormat: InsertTextFormatSnippet,
			}
			items = append(items, item)
		}
	}

	return items
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

// getKeywordDocumentation returns documentation for a keyword
func (a *Analyzer) getKeywordDocumentation(word string) string {
	wordLower := strings.ToLower(word)
	docs := map[string]string{
		"declare":   "**Declare**\n\nDeclares a new variable or function.\n\nExample:\n```\nDeclare x to be 5.\nDeclare function greet does the following:\n    Print \"Hello\".\nThats it.\n```",
		"set":       "**Set**\n\nAssigns a value to an existing variable.\n\nExample:\n```\nSet x to be 10.\n```",
		"print":     "**Print**\n\nOutputs a value to the console.\n\nExample:\n```\nPrint \"Hello, World!\".\nPrint x.\n```",
		"if":        "**If**\n\nConditional statement.\n\nExample:\n```\nIf x is equal to 5, then\n    Print \"x is five\".\nOtherwise\n    Print \"x is not five\".\nThats it.\n```",
		"otherwise": "**Otherwise**\n\nElse clause for if statements.\n\nExample:\n```\nIf condition, then\n    statements\nOtherwise\n    other statements\nThats it.\n```",
		"repeat":    "**Repeat**\n\nLoop statement.\n\nExample:\n```\nRepeat the following 5 times:\n    Print \"Hello\".\nThats it.\n\nRepeat the following while x is less than 10:\n    Set x to be x + 1.\nThats it.\n```",
		"for":       "**For**\n\nFor-each loop.\n\nExample:\n```\nFor each item in list, do the following:\n    Print item.\nThats it.\n```",
		"call":      "**Call**\n\nCalls a function.\n\nExample:\n```\nCall greet.\n```",
		"return":    "**Return**\n\nReturns a value from a function.\n\nExample:\n```\nReturn x + y.\n```",
		"break":     "**Break**\n\nExits the current loop.\n\nExample:\n```\nBreak out of the loop.\n```",
		"toggle":    "**Toggle**\n\nToggles a boolean variable.\n\nExample:\n```\nToggle isActive.\n```",
		"true":      "**true**\n\nBoolean literal representing true.",
		"false":     "**false**\n\nBoolean literal representing false.",
		"always":    "**always**\n\nMakes a variable constant (immutable).\n\nExample:\n```\nDeclare PI to always be 3.14159.\n```",
	}

	if doc, ok := docs[wordLower]; ok {
		return doc
	}
	return ""
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
	funcName := ""
	for _, c := range afterCalling {
		if isWordChar(byte(c)) {
			funcName += string(c)
		} else {
			break
		}
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
		Label:      funcName + "(" + strings.Join(funcInfo.Parameters, ", ") + ")",
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
