package vm

import (
	"github.com/Advik-B/english/runtime"
	"github.com/Advik-B/english/types"
)

// StructDefinition describes a struct type declared in source code.
// It lives in vm/ (not vm/types/) because Methods references *FunctionValue.
type StructDefinition struct {
	Name       string
	Fields     map[string]*FieldDefinition
	Methods    map[string]*FunctionValue
	FieldOrder []string // preserves declaration order
}

// FieldDefinition describes a single field in a struct.
type FieldDefinition struct {
	Name         string
	TypeInfo     *types.TypeInfo
	DefaultValue Value
}

// StructInstance is a runtime instance of a StructDefinition.
type StructInstance struct {
	Definition *StructDefinition
	Fields     map[string]Value
}

// EnglishType implements types.TypeNamer, reporting the struct's declared name
// so error messages and "the type of" name the actual struct.
func (s *StructInstance) EnglishType() *types.TypeInfo {
	name := "struct"
	if s.Definition != nil {
		name = s.Definition.Name
	}
	return &types.TypeInfo{Kind: types.TypeStruct, Name: name}
}

// EnglishTypeName implements runtime.Fielded, reporting the declared struct
// name so that the shared runtime can compare and print a struct without
// knowing this engine's definition record.
func (s *StructInstance) EnglishTypeName() string {
	if s.Definition == nil {
		return "struct"
	}
	return s.Definition.Name
}

// EnglishFields implements runtime.Fielded.
func (s *StructInstance) EnglishFields() map[string]Value { return s.Fields }

// EnglishCopy implements runtime.Copier: only this engine can build a new
// instance around its own definition record.
func (s *StructInstance) EnglishCopy() Value {
	fields := make(map[string]Value, len(s.Fields))
	for name, value := range s.Fields {
		fields[name] = runtime.DeepCopy(value)
	}
	return &StructInstance{Definition: s.Definition, Fields: fields}
}
