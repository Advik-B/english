package sema

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Advik-B/english/ast"
	"github.com/Advik-B/english/parser"
	"github.com/Advik-B/english/types"
)

// funcSig is a declared function or method.
type funcSig struct {
	Name       string
	Params     []ast.Param
	ReturnType *ast.TypeExpr
	Pos        ast.Position
	// Imported marks a function pulled in from another file, whose body this
	// analysis does not re-check.
	Imported bool
}

// structDef is a declared struct type.
type structDef struct {
	Name    string
	Fields  map[string]*ast.StructField
	Order   []string
	Methods map[string]*funcSig
	Pos     ast.Position
}

// Config controls an analysis run.
type Config struct {
	// Predefined names are the stdlib constants, which programs may not
	// shadow.
	Predefined []string
	// File is the name of the file being analysed, used in diagnostics for
	// imported files.
	File string
	// Dir is the directory imports are resolved against. Empty means the
	// process working directory.
	Dir string
}

// Analyzer walks a program, resolving names and checking types.
type Analyzer struct {
	cfg   Config
	diags []*Diagnostic
	scope *scope

	funcs      map[string]*funcSig
	structs    map[string]*structDef
	errorTypes map[string]bool

	// returnType is the declared result type of the function being checked,
	// and inFunction says whether there is one at all, so that a Return
	// outside any function can be reported.
	returnType *ast.TypeExpr
	inFunction bool
	// loopDepth counts enclosing loops so that break and continue outside a
	// loop can be reported.
	loopDepth int
	// discardingResult says the call being checked is a statement of its own,
	// which is the one place a function that gives back nothing may be called.
	discardingResult bool

	// seenImports guards against re-analysing a file, including cycles.
	seenImports map[string]bool
	// checked records the expressions already visited, so that a node cannot
	// be reported against twice.
	checked map[ast.Expression]bool
	// reportedTypes records the annotations already reported as unresolvable,
	// for the same reason.
	reportedTypes map[*ast.TypeExpr]bool
}

// Check analyses a program and returns every problem found, ordered by
// position.
//
// Unlike the checker it replaces, it reports every problem it finds rather
// than the first, and each carries a line and a column.
func Check(prog *ast.Program, cfg Config) []*Diagnostic {
	a := newAnalyzer(cfg)
	a.run(prog)
	sortDiagnostics(a.diags)
	return a.diags
}

func newAnalyzer(cfg Config) *Analyzer {
	a := &Analyzer{
		cfg:           cfg,
		scope:         newScope(nil),
		funcs:         make(map[string]*funcSig),
		structs:       make(map[string]*structDef),
		errorTypes:    make(map[string]bool),
		seenImports:   make(map[string]bool),
		checked:       make(map[ast.Expression]bool),
		reportedTypes: make(map[*ast.TypeExpr]bool),
	}
	for _, name := range cfg.Predefined {
		// Every stdlib constant is a number today; the table in stdlib is the
		// authority on the values, and Name is enough for diagnostics.
		a.declarePredefined(name, types.InfoFor(types.TypeF64))
	}
	return a
}

func (a *Analyzer) run(prog *ast.Program) {
	a.collectDeclarations(prog.Statements)
	a.checkStatements(prog.Statements)
}

// ─── Diagnostics ─────────────────────────────────────────────────────────────

func (a *Analyzer) errorAt(pos ast.Position, format string, args ...interface{}) {
	a.diags = append(a.diags, &Diagnostic{
		Pos:     pos,
		File:    a.cfg.File,
		Message: fmt.Sprintf(format, args...),
	})
}

func (a *Analyzer) errorWithHint(pos ast.Position, hint, format string, args ...interface{}) {
	a.diags = append(a.diags, &Diagnostic{
		Pos:     pos,
		File:    a.cfg.File,
		Message: fmt.Sprintf(format, args...),
		Hint:    hint,
	})
}

// ─── Declaration collection ──────────────────────────────────────────────────

// collectDeclarations records functions, structs and error types before any
// body is checked, so that a function may be called before it is declared and
// two functions may call each other.
func (a *Analyzer) collectDeclarations(stmts []ast.Statement) {
	for _, stmt := range stmts {
		switch s := stmt.(type) {
		case *ast.FunctionDecl:
			if prev, ok := a.funcs[s.Name]; ok && !prev.Imported {
				a.errorAt(s.Pos(), "function '%s' is already declared at line %d", s.Name, prev.Pos.Line)
				continue
			}
			a.funcs[s.Name] = &funcSig{
				Name:       s.Name,
				Params:     s.Params,
				ReturnType: s.ReturnType,
				Pos:        s.Pos(),
			}
		case *ast.StructDecl:
			a.collectStruct(s)
		case *ast.ErrorTypeDecl:
			a.errorTypes[s.Name] = true
		case *ast.ImportStatement:
			a.collectImport(s)
		}
	}
}

func (a *Analyzer) collectStruct(s *ast.StructDecl) {
	if prev, ok := a.structs[s.Name]; ok {
		a.errorAt(s.Pos(), "struct '%s' is already declared at line %d", s.Name, prev.Pos.Line)
		return
	}
	def := &structDef{
		Name:    s.Name,
		Fields:  make(map[string]*ast.StructField, len(s.Fields)),
		Methods: make(map[string]*funcSig, len(s.Methods)),
		Pos:     s.Pos(),
	}
	for _, f := range s.Fields {
		if _, dup := def.Fields[f.Name]; dup {
			a.errorAt(f.Pos(), "struct '%s' already has a field named '%s'", s.Name, f.Name)
			continue
		}
		def.Fields[f.Name] = f
		def.Order = append(def.Order, f.Name)
	}
	for _, m := range s.Methods {
		if _, dup := def.Methods[m.Name]; dup {
			a.errorAt(m.Pos(), "struct '%s' already has a method named '%s'", s.Name, m.Name)
			continue
		}
		def.Methods[m.Name] = &funcSig{
			Name:       m.Name,
			Params:     m.Params,
			ReturnType: m.ReturnType,
			Pos:        m.Pos(),
		}
	}
	a.structs[s.Name] = def
}

// collectImport analyses an imported .abc file and brings its top-level
// declarations into scope.
//
// The previous checker analysed imported files but discarded what they
// declared, so every name an import provided looked undefined.
func (a *Analyzer) collectImport(s *ast.ImportStatement) {
	if !strings.EqualFold(filepath.Ext(s.Path), ".abc") {
		return // not an English source file; nothing to analyse
	}
	path, content, ok := a.readImport(s.Path)
	if !ok {
		return
	}

	lx := parser.NewLexer(string(content))
	p := parser.NewParser(lx.TokenizeAll())
	prog, parseErr := p.Parse()
	if parseErr != nil {
		a.diags = append(a.diags, &Diagnostic{
			Pos:     s.Pos(),
			File:    a.cfg.File,
			Message: fmt.Sprintf("cannot import '%s': %v", s.Path, parseErr),
		})
		return
	}

	// Analyse the imported file on its own, reporting its problems against its
	// own name, then adopt what it declares.
	sub := newAnalyzer(Config{
		Predefined: a.cfg.Predefined,
		File:       s.Path,
		Dir:        filepath.Dir(path),
	})
	sub.seenImports = a.seenImports
	sub.run(prog)
	a.diags = append(a.diags, sub.diags...)

	selected := make(map[string]bool, len(s.Items))
	for _, item := range s.Items {
		selected[item] = true
	}
	wanted := func(name string) bool {
		return len(selected) == 0 || selected[name]
	}

	for name, fn := range sub.funcs {
		if _, exists := a.funcs[name]; !exists && wanted(name) {
			imported := *fn
			imported.Imported = true
			a.funcs[name] = &imported
		}
	}
	for name, def := range sub.structs {
		if _, exists := a.structs[name]; !exists && wanted(name) {
			a.structs[name] = def
		}
	}
	for name := range sub.errorTypes {
		a.errorTypes[name] = true
	}
	// Top-level variables an imported file declares become visible too.
	for name, sym := range sub.globals() {
		if _, exists := a.scope.lookupLocal(name); !exists && wanted(name) {
			adopted := *sym
			a.scope.declare(&adopted)
		}
	}
}

// readImport locates and reads an imported file, returning its resolved path
// and contents.
//
// Import paths in this language are written relative to the working directory
// the program is run from, which is how the engines resolve them, but a path
// relative to the importing file is the more natural thing to write. Both are
// tried, and a file that cannot be read is left for the runtime to report,
// which has the better message for it.
func (a *Analyzer) readImport(importPath string) (string, []byte, bool) {
	candidates := []string{importPath}
	if a.cfg.Dir != "" && !filepath.IsAbs(importPath) {
		candidates = append(candidates, filepath.Join(a.cfg.Dir, importPath))
	}
	for _, path := range candidates {
		if a.seenImports[path] {
			return "", nil, false
		}
		content, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		a.seenImports[path] = true
		return path, content, true
	}
	return "", nil, false
}

// globals returns the names declared at the top level of this analysis.
func (a *Analyzer) globals() map[string]*symbol {
	root := a.scope
	for root.parent != nil {
		root = root.parent
	}
	return root.names
}

// ─── Type helpers ────────────────────────────────────────────────────────────

// resolve turns a written annotation into type metadata, reporting an
// annotation that names neither a built-in type nor a declared struct.
func (a *Analyzer) resolve(te *ast.TypeExpr) *types.TypeInfo {
	if te == nil {
		return nil
	}
	if te.Kind != types.TypeUnknown {
		return types.InfoFor(te.Kind)
	}
	if _, ok := a.structs[te.Name]; ok {
		return &types.TypeInfo{Kind: types.TypeStruct, Name: te.Name}
	}

	// An annotation is reached once per use — collected, checked against a
	// default, declared in a method body — so report it only the first time.
	if a.reportedTypes[te] {
		return nil
	}
	a.reportedTypes[te] = true

	if types.IsRetiredNumericName(te.Name) {
		a.errorWithHint(te.Pos(),
			"There is one number type, written 'number'. Use is_integer to ask whether a value is whole.",
			"'%s' is not a type", te.Name)
		return nil
	}
	a.errorWithHint(te.Pos(),
		fmt.Sprintf("Built-in types are: %s. A struct must be declared before it is used as a type.",
			strings.Join(types.UserTypeNames(), ", ")),
		"unknown type '%s'", te.Name)
	return nil
}

// assignable reports whether a value of type actual may be stored where target
// is expected. An unknown type on either side is accepted, because the checker
// only rejects what it can prove wrong.
func assignable(target, actual *types.TypeInfo) bool {
	if target == nil || actual == nil {
		return true
	}
	if target.Kind == types.TypeUnknown || actual.Kind == types.TypeUnknown {
		return true
	}
	// Assigning nothing is permitted for every type: the language has no
	// non-nullable types today. A variable whose declared type *is* nothing —
	// "Declare result to be nothing." — is likewise unconstrained, matching
	// the exemption the engines make when enforcing a declared type.
	if actual.Kind == types.TypeNull || target.Kind == types.TypeNull {
		return true
	}
	if types.Canonical(target.Kind) != types.Canonical(actual.Kind) {
		return false
	}
	if target.Kind == types.TypeStruct && target.Name != actual.Name {
		return false
	}
	return true
}

// describe renders a type for a diagnostic.
func describe(t *types.TypeInfo) string {
	if t == nil {
		return "an unknown type"
	}
	if t.Kind == types.TypeStruct {
		return t.Name
	}
	return types.Name(t.Kind)
}

// ─── Incremental analysis ────────────────────────────────────────────────────

// NewIncremental returns an analyser that keeps what it has learned between
// calls to CheckNext.
//
// A REPL feeds one line at a time, each parsed as its own program, but they
// share a single session: a name declared on one line is in scope on the next.
// Check starts fresh every time and so cannot serve that.
func NewIncremental(cfg Config) *Analyzer {
	return newAnalyzer(cfg)
}

// CheckNext analyses another program in the same session, returning only the
// problems found in it. Declarations it makes stay in scope for later calls.
func (a *Analyzer) CheckNext(prog *ast.Program) []*Diagnostic {
	a.diags = nil
	a.run(prog)
	sortDiagnostics(a.diags)
	return a.diags
}
