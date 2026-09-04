package ivm

import (
	"fmt"

	"github.com/Advik-B/english/runtime"
	"github.com/Advik-B/english/types"
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
	value interface{}
	// declared is the type the name is locked to, or TypeUnknown when it is
	// not locked at all. It is set for every declaration, whether the type was
	// written or inferred from the initial value.
	declared types.TypeKind
	// typeName is the annotation as written, for diagnostics. It is empty when
	// the type was inferred.
	typeName string
	isConst  bool
}

// declaredName renders an entry's type for a diagnostic.
func (e *envEntry) declaredName() string {
	if e.typeName != "" {
		return e.typeName
	}
	return types.Name(e.declared)
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
		// Enforce the type the name was declared with, whether it was written
		// or inferred. Only explicitly annotated names used to be checked, so
		// "Declare x to be 5." left x unlocked and "Set x to be \"hello\"."
		// was accepted here while the other engine rejected it — the guarantee
		// the language leads with, absent from the engine that runs by default.
		if value != nil && en.declared != types.TypeNull && en.declared != types.TypeUnknown {
			if got := types.Infer(value); types.Canonical(got) != types.Canonical(en.declared) {
				return fmt.Errorf(
					"TypeError: cannot assign %s to variable '%s' (declared as %s)\n  Hint: use 'cast to' for explicit conversion",
					types.Name(got), name, en.declaredName())
			}
		}
		en.value = value
		return nil
	}

	if e.parent != nil {
		return e.parent.setVar(name, value)
	}
	// A name that was never declared used to be created here, silently, so a
	// typo in a Set statement introduced a variable instead of reporting a
	// mistake. Analysis rejects that before either engine runs; reaching it
	// here means the program was not analysed, so create the name rather than
	// failing, but lock it to the value's type as a declaration would.
	e.vars[name] = &envEntry{value: value, declared: types.Canonical(types.Infer(value))}
	return nil
}

func (e *ivmEnv) defineVar(name string, value interface{}, isConst bool) error {
	if _, ok := e.vars[name]; ok {
		return fmt.Errorf("variable '%s' is already defined in this scope", name)
	}
	// Record the inferred type, so the name is locked to it from here on.
	e.vars[name] = &envEntry{
		value:    value,
		declared: types.Canonical(types.Infer(value)),
		isConst:  isConst,
	}
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
	e.vars[name] = &envEntry{
		value:    value,
		declared: types.Canonical(target),
		typeName: typeName,
		isConst:  isConst,
	}
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

// EnglishType implements types.TypeNamer, reporting the struct's declared name
// so that ivm and astvm produce identical type names.
func (s *StructInstance) EnglishType() *types.TypeInfo {
	name := s.DefName
	if name == "" {
		name = "struct"
	}
	return &types.TypeInfo{Kind: types.TypeStruct, Name: name}
}

// EnglishType implements types.TypeNamer for references.
func (r *ReferenceValue) EnglishType() *types.TypeInfo {
	return &types.TypeInfo{Kind: types.TypeRef, Name: "reference"}
}

// EnglishTypeName implements runtime.Fielded, so that the shared runtime can
// compare and print a struct without knowing this engine's definition record.
func (s *StructInstance) EnglishTypeName() string {
	if s.DefName == "" {
		return "struct"
	}
	return s.DefName
}

// EnglishFields implements runtime.Fielded.
func (s *StructInstance) EnglishFields() map[string]interface{} { return s.Fields }

// EnglishCopy implements runtime.Copier: only this engine can build a new
// instance around its own definition record.
func (s *StructInstance) EnglishCopy() interface{} {
	fields := make(map[string]interface{}, len(s.Fields))
	for name, value := range s.Fields {
		fields[name] = runtime.DeepCopy(value)
	}
	return &StructInstance{DefName: s.DefName, DefRef: s.DefRef, Fields: fields}
}

// EnglishString implements runtime.Displayer for references.
func (r *ReferenceValue) EnglishString() string { return "<ref: " + r.Name + ">" }

// field returns the definition of a named field, or nil when the struct does
// not declare one.
func (s *StructInstance) field(name string) *FieldDef {
	if s.DefRef == nil {
		return nil
	}
	for _, fd := range s.DefRef.Fields {
		if fd != nil && fd.Name == name {
			return fd
		}
	}
	return nil
}
