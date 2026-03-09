package vm

import "english/vm/types"

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
