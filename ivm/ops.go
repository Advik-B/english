package ivm

import (
	"fmt"

	"github.com/Advik-B/english/runtime"
	"github.com/Advik-B/english/types"
)

// The value operations live in the runtime package, which both engines share.
// This file maps opcodes onto them.
//
// It used to hold a hand-copied second implementation of all of it, and the
// copies had drifted. This engine measured string length in runes while the
// other counted bytes; allowed text to be indexed while the other refused;
// returned nothing for a missing lookup key while the other raised an error;
// deep-copied arrays while the other aliased them; named types "f64" in
// messages where the other said "number"; and reported that "+" requires
// matching types without naming the offending one.

func doBinaryOp(op BinOp, left, right interface{}) (interface{}, error) {
	switch op {
	case BinAdd:
		return runtime.Add(left, right)
	case BinSub:
		return runtime.Subtract(left, right)
	case BinMul:
		return runtime.Multiply(left, right)
	case BinDiv:
		return runtime.Divide(left, right)
	case BinMod:
		return runtime.Modulo(left, right)
	case BinEq:
		return runtime.Compare("is equal to", left, right)
	case BinNeq:
		return runtime.Compare("is not equal to", left, right)
	case BinLt:
		return runtime.Compare("is less than", left, right)
	case BinLte:
		return runtime.Compare("is less than or equal to", left, right)
	case BinGt:
		return runtime.Compare("is greater than", left, right)
	case BinGte:
		return runtime.Compare("is greater than or equal to", left, right)
	}
	return nil, fmt.Errorf("unknown binary op: %d", op)
}

func doUnaryOp(op UnaryOp, val interface{}) (interface{}, error) {
	switch op {
	case UnaryNeg:
		return runtime.Negate(val)
	case UnaryNot:
		return runtime.Not(val)
	}
	return nil, fmt.Errorf("unknown unary op: %d", op)
}

// ivmToBool converts a value for use as a condition.
func ivmToBool(v interface{}) (bool, error) { return runtime.ToBool(v) }

// ivmToString renders a value as text.
func ivmToString(v interface{}) string { return runtime.ToString(v) }

// inferKindName is the user-facing name of a value's type.
func inferKindName(v interface{}) string { return runtime.NameOf(v) }

// doIndexGet reads the item at a position.
//
// Indexing a lookup table by position used to be accepted here, returning the
// key at that position. The other engine rejected it, nothing documented it,
// and keys(table) is the way to ask for that, so it is no longer special.
func doIndexGet(container, index interface{}) (interface{}, error) {
	return runtime.Index(container, index)
}

// doIndexSet writes the item at a position.
func doIndexSet(container, index, value interface{}) error {
	return runtime.SetIndex(container, index, value)
}

// doLength returns the number of items in a collection, or of characters in
// text.
func doLength(val interface{}) (float64, error) { return runtime.Length(val) }

// doLookupGet reads a value from a lookup table.
func doLookupGet(table, key interface{}) (interface{}, error) {
	return runtime.LookupGet(table, key)
}

// deepCopyValue returns an independent copy of a value.
func deepCopyValue(val interface{}) interface{} { return runtime.DeepCopy(val) }

// typeDefault is the zero value for a declared type, used for a struct field
// with no default expression.
//
// It takes the name written in the source, because that is what survives into
// the bytecode. Every numeric kind used to collapse to float64 here, so an i32
// field held a float64 under this engine and an int32 under the other.
func typeDefault(typeName string) interface{} {
	return runtime.TypeDefault(types.Parse(typeName))
}

// checkFieldType verifies a struct field value against its declared type.
//
// Neither the field values written at instantiation nor those assigned later
// were checked by this engine at all, so a text value could be stored in a
// number field and only fail much later, somewhere else.
func checkFieldType(structName string, fd *FieldDef, value interface{}) error {
	if value == nil || fd.TypeName == "" {
		return nil
	}
	declared := types.Parse(fd.TypeName)
	if declared == types.TypeUnknown {
		return nil // a struct-typed field; only the checker can resolve it
	}
	if got := types.Infer(value); types.Canonical(got) != types.Canonical(declared) {
		return runtime.TypeErrorf("TypeError: field '%s' of %s is %s, but this is %s",
			fd.Name, structName, types.Name(declared), types.Name(got))
	}
	return nil
}
