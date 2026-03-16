package ivm

import (
	"github.com/Advik-B/english/astvm/types"
	"fmt"
)

// BuiltinFunc is the stdlib function dispatcher.
type BuiltinFunc func(name string, args []interface{}) (interface{}, error)

// ─── Value types ──────────────────────────────────────────────────────────────

// StructInstance is the ivm runtime representation of a struct instance.
// Field values are mutable through the pointer.
type StructInstance struct {
	DefName string
	DefRef  *StructDef
	Fields  map[string]interface{}
}

// ReferenceValue holds a reference to a named variable in a specific scope.
type ReferenceValue struct {
	Name string
	Env  *ivmEnv
}

// ─── Environment ──────────────────────────────────────────────────────────────

type envEntry struct {
	value    interface{}
	typeName string // declared type name ("" = inferred)
	isConst  bool
}

type ivmEnv struct {
	vars       map[string]*envEntry
	funcs      map[string]*FuncChunk
	errorTypes map[string]string // name -> parent type name ("" = root)
	structDefs map[string]*StructDef
	parent     *ivmEnv
}

func newIvmEnv() *ivmEnv {
	return &ivmEnv{
		vars:       make(map[string]*envEntry),
		funcs:      make(map[string]*FuncChunk),
		errorTypes: make(map[string]string),
		structDefs: make(map[string]*StructDef),
	}
}

func (e *ivmEnv) newChild() *ivmEnv {
	return &ivmEnv{
		vars:       make(map[string]*envEntry),
		funcs:      make(map[string]*FuncChunk),
		errorTypes: make(map[string]string),
		structDefs: make(map[string]*StructDef),
		parent:     e,
	}
}

func (e *ivmEnv) getVar(name string) (interface{}, bool) {
	if en, ok := e.vars[name]; ok {
		return en.value, true
	}
	if e.parent != nil {
		return e.parent.getVar(name)
	}
	return nil, false
}

func (e *ivmEnv) setVar(name string, value interface{}) error {
	if en, ok := e.vars[name]; ok {
		if en.isConst {
			return fmt.Errorf("TypeError: cannot reassign constant '%s'", name)
		}
		if value != nil && en.typeName != "" {
			actual := inferKindName(value)
			declared := types.Parse(en.typeName)
			actualKind := types.Infer(value)
			if declared != types.TypeNull && declared != types.TypeUnknown &&
				types.Canonical(actualKind) != types.Canonical(declared) {
				return fmt.Errorf("TypeError: cannot assign %s to variable '%s' (declared as %s)\n  Hint: use 'cast to' for explicit conversion", actual, name, en.typeName)
			}
		}
		en.value = value
		return nil
	}
	if e.parent != nil {
		return e.parent.setVar(name, value)
	}
	// Auto-create (needed for internal variables)
	e.vars[name] = &envEntry{value: value}
	return nil
}

func (e *ivmEnv) defineVar(name string, value interface{}, isConst bool) error {
	if _, ok := e.vars[name]; ok {
		return fmt.Errorf("variable '%s' is already defined in this scope", name)
	}
	e.vars[name] = &envEntry{value: value, isConst: isConst}
	return nil
}

func (e *ivmEnv) defineTypedVar(name string, typeName string, value interface{}, isConst bool) error {
	if _, ok := e.vars[name]; ok {
		return fmt.Errorf("variable '%s' is already defined in this scope", name)
	}
	target := types.Parse(typeName)
	if target == types.TypeUnknown {
		return fmt.Errorf("TypeError: unknown type '%s'", typeName)
	}
	if value != nil {
		actual := types.Infer(value)
		if types.Canonical(actual) != types.Canonical(target) {
			return fmt.Errorf("TypeError: cannot initialize %s variable '%s' with %s value\n  Hint: use 'cast to' for explicit conversion", typeName, name, types.Name(actual))
		}
	}
	e.vars[name] = &envEntry{value: value, typeName: typeName, isConst: isConst}
	return nil
}

func (e *ivmEnv) getFunc(name string) (*FuncChunk, bool) {
	if fn, ok := e.funcs[name]; ok {
		return fn, true
	}
	if e.parent != nil {
		return e.parent.getFunc(name)
	}
	return nil, false
}

func (e *ivmEnv) defineFunc(name string, fn *FuncChunk) {
	e.funcs[name] = fn
}

func (e *ivmEnv) getStructDef(name string) (*StructDef, bool) {
	if s, ok := e.structDefs[name]; ok {
		return s, true
	}
	if e.parent != nil {
		return e.parent.getStructDef(name)
	}
	return nil, false
}

func (e *ivmEnv) defineStructDef(name string, sd *StructDef) {
	e.structDefs[name] = sd
}

// Root walks up to the root environment.
func (e *ivmEnv) root() *ivmEnv {
	r := e
	for r.parent != nil {
		r = r.parent
	}
	return r
}

func (e *ivmEnv) defineErrorType(name, parent string) {
	e.root().errorTypes[name] = parent
}

func (e *ivmEnv) isKnownErrorType(name string) bool {
	_, ok := e.root().errorTypes[name]
	return ok
}

func (e *ivmEnv) isSubtypeOf(child, parent string) bool {
	r := e.root()
	current := child
	for current != "" {
		if current == parent {
			return true
		}
		p, ok := r.errorTypes[current]
		if !ok {
			break
		}
		current = p
	}
	return false
}
