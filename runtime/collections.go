package runtime

import (
	"github.com/Advik-B/english/types"
)

// Length returns the number of items in a collection, or the number of
// characters in text.
//
// Text is measured in characters, not bytes. One engine counted bytes and the
// other counted runes, so the length of any string with a non-ASCII character
// depended on which engine you ran.
func Length(v Value) (float64, error) {
	switch val := v.(type) {
	case []any:
		return float64(len(val)), nil
	case *types.ArrayValue:
		return float64(len(val.Elements)), nil
	case *types.LookupTableValue:
		return float64(len(val.Entries)), nil
	case *types.RangeValue:
		return float64(val.Length()), nil
	case string:
		return float64(len([]rune(val))), nil
	case *types.TypedValue:
		return Length(val.Value)
	}
	return 0, TypeErrorf("TypeError: cannot get the length of %s", NameOf(v))
}

// Index reads the item at a position.
//
// Indexing text yields a one-character string. One engine supported that and
// the other rejected it, so the same program worked or failed depending on how
// it was run.
func Index(container, index Value) (Value, error) {
	i, err := ToNumber(index)
	if err != nil {
		return nil, TypeErrorf("TypeError: an index must be a number, got %s", NameOf(index))
	}
	at := int(i)

	switch c := container.(type) {
	case []any:
		if at < 0 || at >= len(c) {
			return nil, outOfRange(at, len(c), "list")
		}
		return c[at], nil
	case *types.ArrayValue:
		if at < 0 || at >= len(c.Elements) {
			return nil, outOfRange(at, len(c.Elements), "array")
		}
		return c.Elements[at], nil
	case *types.RangeValue:
		v, ok := c.Get(at)
		if !ok {
			return nil, outOfRange(at, c.Length(), "range")
		}
		return v, nil
	case string:
		runes := []rune(c)
		if at < 0 || at >= len(runes) {
			return nil, outOfRange(at, len(runes), "text")
		}
		return string(runes[at]), nil
	case *types.TypedValue:
		return Index(c.Value, index)
	}
	return nil, TypeErrorf("TypeError: cannot take an item from %s", NameOf(container))
}

// SetIndex writes the item at a position.
//
// An array checks the new value against its element type; a list holds mixed
// types by design. One engine performed that check and the other did not.
func SetIndex(container, index, value Value) error {
	i, err := ToNumber(index)
	if err != nil {
		return TypeErrorf("TypeError: an index must be a number, got %s", NameOf(index))
	}
	at := int(i)

	switch c := container.(type) {
	case []any:
		if at < 0 || at >= len(c) {
			return outOfRange(at, len(c), "list")
		}
		c[at] = value
		return nil
	case *types.ArrayValue:
		if at < 0 || at >= len(c.Elements) {
			return outOfRange(at, len(c.Elements), "array")
		}
		if value != nil && c.ElementType != types.TypeUnknown {
			if got := types.Infer(value); types.Canonical(got) != types.Canonical(c.ElementType) {
				return TypeErrorf("TypeError: cannot put %s into an array of %s",
					types.Name(got), types.Name(c.ElementType))
			}
		}
		c.Elements[at] = value
		return nil
	case *types.RangeValue:
		return TypeErrorf("TypeError: cannot modify a range")
	}
	return TypeErrorf("TypeError: cannot assign into %s", NameOf(container))
}

func outOfRange(at, length int, what string) error {
	return TypeErrorf("index %d is out of range for a %s of length %d", at, what, length)
}

// LookupGet reads a value from a lookup table.
//
// A key that is not present is an error. One engine reported it and the other
// returned nothing, so a typo in a key produced a confusing failure much later
// in one of them.
func LookupGet(table, key Value) (Value, error) {
	lt, ok := table.(*types.LookupTableValue)
	if !ok {
		return nil, TypeErrorf("TypeError: cannot look up a key in %s; expected a lookup table", NameOf(table))
	}
	serial, err := types.SerializeKey(key)
	if err != nil {
		return nil, err
	}
	value, present := lt.Entries[serial]
	if !present {
		return nil, TypeErrorf("KeyError: the key %s is not in the lookup table", ToString(key))
	}
	return value, nil
}

// LookupSet writes a value into a lookup table.
func LookupSet(table, key, value Value) error {
	lt, ok := table.(*types.LookupTableValue)
	if !ok {
		return TypeErrorf("TypeError: cannot set a key in %s; expected a lookup table", NameOf(table))
	}
	serial, err := types.SerializeKey(key)
	if err != nil {
		return err
	}
	lt.Set(serial, value)
	return nil
}

// LookupHas reports whether a key is present.
func LookupHas(table, key Value) (bool, error) {
	lt, ok := table.(*types.LookupTableValue)
	if !ok {
		return false, TypeErrorf("TypeError: 'has' requires a lookup table, got %s", NameOf(table))
	}
	serial, err := types.SerializeKey(key)
	if err != nil {
		return false, err
	}
	_, present := lt.Entries[serial]
	return present, nil
}

// NewArray builds an array literal, checking that the elements share a type.
//
// When no element type is declared it is taken from the first non-nothing
// element. One engine enforced this and the other accepted anything, so an
// array of mixed types was rejected or silently built depending on the engine.
func NewArray(declared types.TypeKind, elements []Value) (Value, error) {
	elemType := declared
	for _, el := range elements {
		if el == nil {
			continue
		}
		got := types.Canonical(types.Infer(el))
		if elemType == types.TypeUnknown {
			elemType = got
			continue
		}
		if got != types.Canonical(elemType) {
			return nil, TypeErrorf("TypeError: an array of %s cannot hold %s",
				types.Name(elemType), types.Name(got))
		}
	}
	return &types.ArrayValue{ElementType: elemType, Elements: elements}, nil
}

// TypeDefault is the zero value for a declared type, used for a struct field
// with no default expression.
func TypeDefault(kind types.TypeKind) Value {
	switch kind {
	case types.TypeI32:
		return int32(0)
	case types.TypeI64:
		return int64(0)
	case types.TypeU32:
		return uint32(0)
	case types.TypeU64:
		return uint64(0)
	case types.TypeF32:
		return float32(0)
	case types.TypeF64:
		return float64(0)
	case types.TypeString:
		return ""
	case types.TypeBool:
		return false
	case types.TypeList:
		return []any{}
	case types.TypeArray:
		return &types.ArrayValue{}
	case types.TypeLookup:
		return types.NewLookupTable()
	}
	return nil
}

// Element reads the nth item of a collection for iteration.
//
// It differs from Index in one place: iterating a lookup table yields its
// keys, in insertion order, which is not what indexing one would mean. Keeping
// the two apart lets Index refuse a lookup table, matching what the checker
// allows, without breaking "for each key in table".
func Element(collection, index Value) (Value, error) {
	if lt, ok := collection.(*types.LookupTableValue); ok {
		i, err := ToNumber(index)
		if err != nil {
			return nil, TypeErrorf("TypeError: an index must be a number, got %s", NameOf(index))
		}
		at := int(i)
		if at < 0 || at >= len(lt.KeyOrder) {
			return nil, outOfRange(at, len(lt.KeyOrder), "lookup table")
		}
		key, _, ok := types.DeserializeKey(lt.KeyOrder[at])
		if !ok {
			return lt.KeyOrder[at], nil
		}
		return key, nil
	}
	return Index(collection, index)
}
