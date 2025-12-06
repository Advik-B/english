package vm

import "fmt"

// Environment represents a scope for variables and functions
type Environment struct {
	variables map[string]Value
	constants map[string]bool
	functions map[string]*FunctionValue
	structs   map[string]*StructDefinition
	parent    *Environment
}

// NewEnvironment creates a new environment
func NewEnvironment() *Environment {
	return &Environment{
		variables: make(map[string]Value),
		constants: make(map[string]bool),
		functions: make(map[string]*FunctionValue),
		structs:   make(map[string]*StructDefinition),
	}
}

// NewChild creates a child environment
func (e *Environment) NewChild() *Environment {
	return &Environment{
		variables: make(map[string]Value),
		constants: make(map[string]bool),
		functions: make(map[string]*FunctionValue),
		structs:   make(map[string]*StructDefinition),
		parent:    e,
	}
}

// Get retrieves a variable from the environment
func (e *Environment) Get(name string) (Value, bool) {
	if val, ok := e.variables[name]; ok {
		return val, ok
	}
	if e.parent != nil {
		return e.parent.Get(name)
	}
	return nil, false
}

// Set assigns a value to a variable
func (e *Environment) Set(name string, value Value) error {
	if e.constants[name] {
		return fmt.Errorf("cannot reassign constant '%s'\n  Hint: Constants are declared with 'to be always' or 'to always be'", name)
	}
	if _, exists := e.variables[name]; exists {
		e.variables[name] = value
		return nil
	}
	// Look up the scope chain for the variable and set it where it's found
	if e.parent != nil {
		return e.parent.Set(name, value)
	}
	e.variables[name] = value
	return nil
}

// Define declares a new variable
func (e *Environment) Define(name string, value Value, isConstant bool) error {
	if _, exists := e.variables[name]; exists {
		return fmt.Errorf("variable %s already defined", name)
	}
	e.variables[name] = value
	e.constants[name] = isConstant
	return nil
}

// GetFunction retrieves a function from the environment
func (e *Environment) GetFunction(name string) (*FunctionValue, bool) {
	if fn, ok := e.functions[name]; ok {
		return fn, ok
	}
	if e.parent != nil {
		return e.parent.GetFunction(name)
	}
	return nil, false
}

// DefineFunction declares a new function
func (e *Environment) DefineFunction(name string, fn *FunctionValue) {
	e.functions[name] = fn
}

// GetAllVariables returns a copy of all variables in the current scope
func (e *Environment) GetAllVariables() map[string]Value {
	result := make(map[string]Value)
	for k, v := range e.variables {
		result[k] = v
	}
	return result
}

// GetAllFunctions returns a copy of all functions in the current scope
func (e *Environment) GetAllFunctions() map[string]*FunctionValue {
	result := make(map[string]*FunctionValue)
	for k, v := range e.functions {
		result[k] = v
	}
	return result
}

// IsConstant returns whether a variable is a constant
func (e *Environment) IsConstant(name string) bool {
	return e.constants[name]
}

// GetStruct retrieves a struct definition from the environment
func (e *Environment) GetStruct(name string) (*StructDefinition, bool) {
	if s, ok := e.structs[name]; ok {
		return s, ok
	}
	if e.parent != nil {
		return e.parent.GetStruct(name)
	}
	return nil, false
}

// DefineStruct declares a new struct definition
func (e *Environment) DefineStruct(name string, def *StructDefinition) {
	e.structs[name] = def
}
