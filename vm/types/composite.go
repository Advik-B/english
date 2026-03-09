package types

// ArrayValue is a homogeneous array: every element must share the same TypeKind.
// The zero value has ElementType == TypeUnknown.
type ArrayValue struct {
	ElementType TypeKind
	Elements    []interface{}
}

// LookupTableValue is an ordered dictionary that maps hashable keys (number,
// text, boolean) to values of any type.
type LookupTableValue struct {
	Entries  map[string]interface{} // serialised key (SerializeKey) → value
	KeyOrder []string               // insertion-order serialised keys
}

// NewLookupTable returns an initialised, empty LookupTableValue.
func NewLookupTable() *LookupTableValue {
	return &LookupTableValue{
		Entries:  make(map[string]interface{}),
		KeyOrder: []string{},
	}
}

// Set inserts or updates an entry.  key must already be serialised via SerializeKey.
func (lt *LookupTableValue) Set(serialKey string, value interface{}) {
	if _, exists := lt.Entries[serialKey]; !exists {
		lt.KeyOrder = append(lt.KeyOrder, serialKey)
	}
	lt.Entries[serialKey] = value
}

// Delete removes an entry.  Returns true if the key was present.
func (lt *LookupTableValue) Delete(serialKey string) bool {
	if _, exists := lt.Entries[serialKey]; !exists {
		return false
	}
	delete(lt.Entries, serialKey)
	newOrder := make([]string, 0, len(lt.KeyOrder))
	for _, k := range lt.KeyOrder {
		if k != serialKey {
			newOrder = append(newOrder, k)
		}
	}
	lt.KeyOrder = newOrder
	return true
}
