package types

import "strings"

// TypeKind enumerates every distinct type in the English language.
type TypeKind int

const (
	TypeUnknown TypeKind = iota

	// Fine-grained numeric types (used by struct field declarations and explicit casts)
	TypeI32 // 32-bit signed integer
	TypeI64 // 64-bit signed integer
	TypeU32 // 32-bit unsigned integer
	TypeU64 // 64-bit unsigned integer
	TypeF32 // 32-bit float
	TypeF64 // 64-bit float — the canonical "number" type for all user variables

	// Primary language types
	TypeString   // text
	TypeBool     // boolean
	TypeList     // heterogeneous list  ([]interface{})
	TypeStruct   // struct instance
	TypeFunction // function value
	TypeNull     // nothing / nil
	TypeError    // catchable error value
	TypeRef      // reference to a variable

	// Composite types introduced by the static type system
	TypeArray  // homogeneous array   (*ArrayValue)
	TypeLookup // lookup table / dict (*LookupTableValue)
)

// Name returns the user-facing type name for a TypeKind.
func Name(tk TypeKind) string {
	switch tk {
	case TypeI32, TypeI64, TypeU32, TypeU64, TypeF32, TypeF64:
		return "number"
	case TypeString:
		return "text"
	case TypeBool:
		return "boolean"
	case TypeList:
		return "list"
	case TypeArray:
		return "array"
	case TypeLookup:
		return "lookup table"
	case TypeNull:
		return "nothing"
	case TypeFunction:
		return "function"
	case TypeStruct:
		return "struct"
	case TypeError:
		return "error"
	case TypeRef:
		return "reference"
	default:
		return "unknown"
	}
}

// IsNumeric returns true for all numeric TypeKinds.
func IsNumeric(tk TypeKind) bool {
	switch tk {
	case TypeI32, TypeI64, TypeU32, TypeU64, TypeF32, TypeF64:
		return true
	}
	return false
}

// Canonical maps every numeric TypeKind to TypeF64 (the single "number" type
// visible to users), leaving all other kinds unchanged.  Used for assignment
// compatibility checks.
func Canonical(tk TypeKind) TypeKind {
	if IsNumeric(tk) {
		return TypeF64
	}
	return tk
}

// Parse converts a user-supplied type name string into a TypeKind.
func Parse(s string) TypeKind {
	switch strings.ToLower(s) {
	case "i32", "integer":
		return TypeI32
	case "i64":
		return TypeI64
	case "u32", "unsigned integer":
		return TypeU32
	case "u64":
		return TypeU64
	case "f32", "float":
		return TypeF32
	case "f64", "double", "number", "num":
		return TypeF64
	case "string", "text", "str":
		return TypeString
	case "bool", "boolean":
		return TypeBool
	case "list":
		return TypeList
	case "array":
		return TypeArray
	case "lookup", "table", "lookup table":
		return TypeLookup
	default:
		return TypeUnknown
	}
}

// UserTypeNames returns the canonical user-facing type names that are valid
// for explicit type annotations.  Used in error messages.
func UserTypeNames() []string {
	return []string{"number", "text", "boolean", "list", "array", "lookup table"}
}
