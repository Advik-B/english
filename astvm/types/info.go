package types

import "fmt"

// TypeInfo holds type metadata used by the struct / typed-variable system.
type TypeInfo struct {
	Kind         TypeKind
	Name         string
	ElementType  *TypeInfo            // for array: element type
	StructFields map[string]*TypeInfo // for struct: field name → type
}

func (t *TypeInfo) String() string {
	switch t.Kind {
	case TypeI32:
		return "i32"
	case TypeI64:
		return "i64"
	case TypeU32:
		return "u32"
	case TypeU64:
		return "u64"
	case TypeF32:
		return "f32"
	case TypeF64:
		return "f64"
	case TypeString:
		return "text"
	case TypeBool:
		return "boolean"
	case TypeList:
		return "list"
	case TypeArray:
		if t.ElementType != nil {
			return fmt.Sprintf("array of %s", t.ElementType.String())
		}
		return "array"
	case TypeLookup:
		return "lookup table"
	case TypeStruct:
		return t.Name
	case TypeFunction:
		return "function"
	case TypeNull:
		return "nothing"
	case TypeError:
		return "error"
	case TypeRef:
		return "reference"
	default:
		return "unknown"
	}
}

// TypedValue wraps a runtime value with explicit type information.
// Used by the struct field system.
type TypedValue struct {
	Value    interface{}
	TypeInfo *TypeInfo
}
