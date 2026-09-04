package sema

import (
	"github.com/Advik-B/english/ast"
	"github.com/Advik-B/english/types"
)

// symbol is a name declared in a scope.
type symbol struct {
	Name string
	// Type is the declared or inferred type. A nil type means the type could
	// not be determined, and nothing about the name is checked.
	Type    *types.TypeInfo
	IsConst bool
	Pos     ast.Position
	// Predefined marks a stdlib constant, which has no source position.
	Predefined bool
	// Assigned reports whether the name has been given a value. A typed
	// declaration with no initialiser starts unassigned.
	Assigned bool
}

// scope is one lexical level of the symbol table.
type scope struct {
	names  map[string]*symbol
	parent *scope
}

func newScope(parent *scope) *scope {
	return &scope{names: make(map[string]*symbol), parent: parent}
}

// lookup finds a name in this scope or any enclosing one.
func (s *scope) lookup(name string) (*symbol, bool) {
	for cur := s; cur != nil; cur = cur.parent {
		if sym, ok := cur.names[name]; ok {
			return sym, true
		}
	}
	return nil, false
}

// lookupLocal finds a name declared in this scope only, which is what
// duplicate-declaration detection needs: shadowing an outer name is allowed.
func (s *scope) lookupLocal(name string) (*symbol, bool) {
	sym, ok := s.names[name]
	return sym, ok
}

// declare adds a name to this scope.
func (s *scope) declare(sym *symbol) {
	s.names[sym.Name] = sym
}

// ─── Scope stack ─────────────────────────────────────────────────────────────

// pushScope opens a new lexical scope.
//
// Types live in the scope that declared them. The previous checker kept a flat
// map of variable types beside a separate stack of declared names, so a type
// recorded inside a block stayed visible after the block ended and could be
// used to check code it had nothing to do with.
func (a *Analyzer) pushScope() {
	a.scope = newScope(a.scope)
}

// popScope closes the innermost lexical scope, discarding its names and their
// types together.
func (a *Analyzer) popScope() {
	if a.scope != nil && a.scope.parent != nil {
		a.scope = a.scope.parent
	}
}

// declare records a declaration, reporting a duplicate in the same scope.
func (a *Analyzer) declare(name string, t *types.TypeInfo, isConst bool, pos ast.Position, assigned bool) *symbol {
	if prev, ok := a.scope.lookupLocal(name); ok {
		if prev.Predefined {
			a.errorAt(pos, "'%s' shadows a built-in constant", name)
		} else {
			a.errorAt(pos, "'%s' is already declared at line %d", name, prev.Pos.Line)
		}
		return prev
	}
	// Shadowing a predefined constant from an inner scope is equally confusing.
	if prev, ok := a.scope.lookup(name); ok && prev.Predefined {
		a.errorAt(pos, "'%s' shadows a built-in constant", name)
		return prev
	}
	sym := &symbol{Name: name, Type: t, IsConst: isConst, Pos: pos, Assigned: assigned}
	a.scope.declare(sym)
	return sym
}

// declarePredefined records a stdlib constant in the global scope.
func (a *Analyzer) declarePredefined(name string, t *types.TypeInfo) {
	a.scope.declare(&symbol{Name: name, Type: t, IsConst: true, Predefined: true, Assigned: true})
}
