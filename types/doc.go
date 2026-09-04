// Package types defines the English language's type system: the TypeKind enum,
// composite value types (array, lookup table), type metadata, key serialisation,
// explicit casting, and error helpers.
//
// It is a leaf package: it imports nothing else in this module, so the lexer,
// the parser, the AST, the standard library and both execution engines can all
// depend on one definition of the type system.
package types
