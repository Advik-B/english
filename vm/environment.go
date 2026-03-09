package vm

import (
	"english/vm/types"
	"fmt"
	"strings"
)

// Environment represents a lexical scope: variables, constants, functions, and structs.
type Environment struct {
variables        map[string]Value
variableTypes    map[string]types.TypeKind // declared type; fixed at first Define
constants        map[string]bool
functions        map[string]*FunctionValue
structs          map[string]*StructDefinition
customErrorTypes map[string]bool // registered custom error type names
parent           *Environment
}

// NewEnvironment creates a new root environment.
func NewEnvironment() *Environment {
return &Environment{
variables:        make(map[string]Value),
variableTypes:    make(map[string]types.TypeKind),
constants:        make(map[string]bool),
functions:        make(map[string]*FunctionValue),
structs:          make(map[string]*StructDefinition),
customErrorTypes: make(map[string]bool),
}
}

// NewChild creates a child scope that inherits from this environment.
func (e *Environment) NewChild() *Environment {
return &Environment{
variables:        make(map[string]Value),
variableTypes:    make(map[string]types.TypeKind),
constants:        make(map[string]bool),
functions:        make(map[string]*FunctionValue),
structs:          make(map[string]*StructDefinition),
customErrorTypes: make(map[string]bool),
parent:           e,
}
}

// Get retrieves a variable value, searching up the scope chain.
func (e *Environment) Get(name string) (Value, bool) {
if val, ok := e.variables[name]; ok {
return val, true
}
if e.parent != nil {
return e.parent.Get(name)
}
return nil, false
}

// GetVarType returns the declared TypeKind of a variable in the scope chain.
func (e *Environment) GetVarType(name string) (types.TypeKind, bool) {
if tk, ok := e.variableTypes[name]; ok {
return tk, true
}
if e.parent != nil {
return e.parent.GetVarType(name)
}
return types.TypeUnknown, false
}

// Set assigns a new value to an existing variable, enforcing the declared type.
//
// Assigning nothing (nil) is always permitted — it acts as a typed null.
func (e *Environment) Set(name string, value Value) error {
if e.constants[name] {
return fmt.Errorf(
"TypeError: cannot reassign constant '%s'\n  Hint: constants are declared with 'to be always'",
name,
)
}
if _, exists := e.variables[name]; exists {
if value != nil {
declared := e.variableTypes[name]
actual := inferTypeKind(value)
if declared != types.TypeNull && declared != types.TypeUnknown &&
types.Canonical(actual) != types.Canonical(declared) {
return fmt.Errorf(
"TypeError: cannot assign %s to variable '%s' (declared as %s)\n  Hint: use 'cast to' for explicit conversion",
types.Name(actual), name, types.Name(declared),
)
}
}
e.variables[name] = value
return nil
}
if e.parent != nil {
return e.parent.Set(name, value)
}
// Variable unknown — create it dynamically (needed for stdlib internals)
e.variables[name] = value
return nil
}

// Define declares a new variable in the current scope, inferring its type from value.
func (e *Environment) Define(name string, value Value, isConstant bool) error {
if _, exists := e.variables[name]; exists {
return fmt.Errorf("variable '%s' is already defined in this scope", name)
}
e.variables[name] = value
e.constants[name] = isConstant
e.variableTypes[name] = inferTypeKind(value)
return nil
}

// DefineTyped declares a new variable with an explicit type annotation.
// The type annotation is enforced: the initial value (if any) must match the declared type,
// and all subsequent assignments must also match.
func (e *Environment) DefineTyped(name string, typeName string, value Value, isConstant bool) error {
if _, exists := e.variables[name]; exists {
return fmt.Errorf("variable '%s' is already defined in this scope", name)
}
targetType := types.Parse(typeName)
if targetType == types.TypeUnknown {
return fmt.Errorf("TypeError: unknown type '%s'\n  Hint: valid types are %s",
typeName, strings.Join(types.UserTypeNames(), ", "))
}
// Type-check the initial value if provided
if value != nil {
actual := inferTypeKind(value)
if types.Canonical(actual) != types.Canonical(targetType) {
return fmt.Errorf(
"TypeError: cannot initialize %s variable '%s' with %s value\n  Hint: use 'cast to' for explicit conversion",
types.Name(targetType), name, types.Name(actual),
)
}
}
e.variables[name] = value
e.constants[name] = isConstant
e.variableTypes[name] = targetType
return nil
}

// DefineErrorType registers a custom error type name in the root environment.
func (e *Environment) DefineErrorType(name string) {
root := e
for root.parent != nil {
root = root.parent
}
root.customErrorTypes[name] = true
}

// IsKnownErrorType reports whether name is a registered custom error type.
func (e *Environment) IsKnownErrorType(name string) bool {
root := e
for root.parent != nil {
root = root.parent
}
return root.customErrorTypes[name]
}

// GetFunction retrieves a function searching up the scope chain.
func (e *Environment) GetFunction(name string) (*FunctionValue, bool) {
if fn, ok := e.functions[name]; ok {
return fn, true
}
if e.parent != nil {
return e.parent.GetFunction(name)
}
return nil, false
}

// DefineFunction registers a function in the current scope.
func (e *Environment) DefineFunction(name string, fn *FunctionValue) {
e.functions[name] = fn
}

// GetAllVariables returns a shallow copy of variables in this scope only.
func (e *Environment) GetAllVariables() map[string]Value {
result := make(map[string]Value, len(e.variables))
for k, v := range e.variables {
result[k] = v
}
return result
}

// GetAllFunctions returns a shallow copy of functions in this scope only.
func (e *Environment) GetAllFunctions() map[string]*FunctionValue {
result := make(map[string]*FunctionValue, len(e.functions))
for k, v := range e.functions {
result[k] = v
}
return result
}

// IsConstant returns whether a variable is a constant (searches up the chain).
func (e *Environment) IsConstant(name string) bool {
if c, ok := e.constants[name]; ok {
return c
}
if e.parent != nil {
return e.parent.IsConstant(name)
}
return false
}

// GetStruct retrieves a struct definition searching up the scope chain.
func (e *Environment) GetStruct(name string) (*StructDefinition, bool) {
if s, ok := e.structs[name]; ok {
return s, true
}
if e.parent != nil {
return e.parent.GetStruct(name)
}
return nil, false
}

// DefineStruct registers a struct definition in the current scope.
func (e *Environment) DefineStruct(name string, def *StructDefinition) {
e.structs[name] = def
}
