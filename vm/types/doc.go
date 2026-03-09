// Package types defines the English language's type system: the TypeKind enum,
// composite value types (array, lookup table), type metadata, key serialisation,
// explicit casting, and error helpers.
//
// This package has no dependency on the vm package so that vm can import it
// without creating a circular import.
package types
