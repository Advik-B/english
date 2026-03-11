// Package vm provides the virtual machine (evaluator) for the English programming language.
//
// The vm package is split into multiple files for better organization:
//   - values.go: Value types (Value, FunctionValue, ReturnValue, BreakValue, RuntimeError)
//   - environment.go: Environment struct and methods for scope management
//   - evaluator.go: Evaluator struct and all eval* methods
//   - operations.go: Arithmetic and comparison operations
//   - conversions.go: Type conversion functions
//   - helpers.go: Helper utilities
package vm
