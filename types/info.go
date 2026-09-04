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
		if t.Name != "" {
			return t.Name
		}
		return "unknown"
	}
}

// TypedValue wraps a runtime value with explicit type information.
// Used by the struct field system.
type TypedValue struct {
	Value    interface{}
	TypeInfo *TypeInfo
}

// TypeNamer is implemented by runtime values whose type is defined outside this
// package — function values, struct instances and references, which live in the
// engine packages.  It lets Describe classify them without types importing the
// engines, so both the AST evaluator and the instruction VM share one
// implementation instead of maintaining divergent copies.
type TypeNamer interface {
	EnglishType() *TypeInfo
}

// Describe returns full type metadata for any runtime value.  It is the single
// source of truth for "what type is this value" across both engines, the stdlib
// and the type checker.
func Describe(v interface{}) *TypeInfo {
	// TypedValue carries its own annotation; prefer it.
	if tv, ok := v.(*TypedValue); ok {
		if tv.TypeInfo != nil {
			return tv.TypeInfo
		}
		return Describe(tv.Value)
	}
	// Engine-defined values describe themselves.
	if tn, ok := v.(TypeNamer); ok {
		return tn.EnglishType()
	}

	switch val := v.(type) {
	case float64:
		return &TypeInfo{Kind: TypeF64, Name: "f64"}
	case int32:
		return &TypeInfo{Kind: TypeI32, Name: "i32"}
	case int64:
		return &TypeInfo{Kind: TypeI64, Name: "i64"}
	case uint32:
		return &TypeInfo{Kind: TypeU32, Name: "u32"}
	case uint64:
		return &TypeInfo{Kind: TypeU64, Name: "u64"}
	case float32:
		return &TypeInfo{Kind: TypeF32, Name: "f32"}
	case string:
		return &TypeInfo{Kind: TypeString, Name: "text"}
	case bool:
		return &TypeInfo{Kind: TypeBool, Name: "boolean"}
	case []interface{}:
		return &TypeInfo{Kind: TypeList, Name: "list"}
	case *ArrayValue:
		return &TypeInfo{
			Kind:        TypeArray,
			Name:        "array",
			ElementType: &TypeInfo{Kind: val.ElementType, Name: Name(val.ElementType)},
		}
	case *RangeValue:
		return &TypeInfo{Kind: TypeUnknown, Name: "range"}
	case *LookupTableValue:
		return &TypeInfo{Kind: TypeLookup, Name: "lookup table"}
	case *ErrorValue:
		return &TypeInfo{Kind: TypeError, Name: "error"}
	case nil:
		return &TypeInfo{Kind: TypeNull, Name: "nothing"}
	default:
		return &TypeInfo{Kind: TypeUnknown, Name: "unknown"}
	}
}

// InfoFor returns type metadata for a kind alone. Describe answers "what type
// is this value"; InfoFor answers "what does this type look like", which is
// what the static side needs when there is no value to inspect.
func InfoFor(k TypeKind) *TypeInfo {
	return &TypeInfo{Kind: k, Name: Name(k)}
}

// Infer determines the TypeKind of a runtime value.
func Infer(v interface{}) TypeKind {
	return Describe(v).Kind
}

// NameOf returns the user-facing type name of a runtime value ("number",
// "text", …).  This is the renderer to use in error messages; TypeInfo.String
// is the finer-grained diagnostic renderer used by "the type of".
func NameOf(v interface{}) string {
	return Name(Infer(v))
}
