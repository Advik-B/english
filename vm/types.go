package vm

import (
	"fmt"
)

// TypeKind represents the kind of type
type TypeKind int

const (
	TypeUnknown TypeKind = iota
	TypeI32              // 32-bit signed integer
	TypeI64              // 64-bit signed integer
	TypeU32              // 32-bit unsigned integer
	TypeU64              // 64-bit unsigned integer
	TypeF32              // 32-bit float
	TypeF64              // 64-bit float
	TypeString           // string
	TypeBool             // boolean
	TypeList             // list/array
	TypeStruct           // struct
	TypeFunction         // function
	TypeNull             // null/nil
	TypeError            // error type
	TypeReference        // reference to another value
)

// TypeInfo holds type metadata for a value
type TypeInfo struct {
	Kind         TypeKind
	Name         string // For structs, this is the struct name
	ElementType  *TypeInfo // For lists, this is the element type
	StructFields map[string]*TypeInfo // For structs, field name -> type
}

// String returns a string representation of the type
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
		return "string"
	case TypeBool:
		return "bool"
	case TypeList:
		if t.ElementType != nil {
			return fmt.Sprintf("list<%s>", t.ElementType.String())
		}
		return "list"
	case TypeStruct:
		return t.Name
	case TypeFunction:
		return "function"
	case TypeNull:
		return "null"
	case TypeError:
		return "error"
	case TypeReference:
		return "reference"
	default:
		return "unknown"
	}
}

// TypedValue wraps a value with type information
type TypedValue struct {
	Value    Value
	TypeInfo *TypeInfo
}

// StructDefinition represents a struct type definition
type StructDefinition struct {
	Name         string
	Fields       map[string]*FieldDefinition
	Methods      map[string]*FunctionValue
	FieldOrder   []string // To maintain field declaration order
}

// FieldDefinition represents a field in a struct
type FieldDefinition struct {
	Name         string
	TypeInfo     *TypeInfo
	DefaultValue Value
}

// StructInstance represents an instance of a struct
type StructInstance struct {
	Definition *StructDefinition
	Fields     map[string]Value
}

// ErrorValue represents an error that can be caught
type ErrorValue struct {
	Message   string
	ErrorType string // e.g., "ZeroDivisionError", "TypeError", "RuntimeError"
	CallStack []string
}

func (e *ErrorValue) Error() string {
	result := fmt.Sprintf("%s: %s\n", e.ErrorType, e.Message)
	if len(e.CallStack) > 0 {
		result += "\nCall Stack (most recent first):\n"
		for i, frame := range e.CallStack {
			result += fmt.Sprintf("  %d. %s\n", i+1, frame)
		}
	}
	return result
}

// ReferenceValue represents a reference to another value
type ReferenceValue struct {
	Name string // Variable name being referenced
	Env  *Environment
}

// GetType returns the TypeInfo for a value
func GetType(v Value) *TypeInfo {
	switch val := v.(type) {
	case *TypedValue:
		return val.TypeInfo
	case float64:
		// Default to f64 for untyped floats
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
		return &TypeInfo{Kind: TypeString, Name: "string"}
	case bool:
		return &TypeInfo{Kind: TypeBool, Name: "bool"}
	case []interface{}:
		return &TypeInfo{Kind: TypeList, Name: "list"}
	case *StructInstance:
		return &TypeInfo{Kind: TypeStruct, Name: val.Definition.Name}
	case *FunctionValue:
		return &TypeInfo{Kind: TypeFunction, Name: "function"}
	case *ErrorValue:
		return &TypeInfo{Kind: TypeError, Name: "error"}
	case *ReferenceValue:
		return &TypeInfo{Kind: TypeReference, Name: "reference"}
	case nil:
		return &TypeInfo{Kind: TypeNull, Name: "null"}
	default:
		return &TypeInfo{Kind: TypeUnknown, Name: "unknown"}
	}
}

// ParseTypeString converts a type string to TypeKind
func ParseTypeString(typeStr string) TypeKind {
	switch typeStr {
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
	case "f64", "double":
		return TypeF64
	case "string":
		return TypeString
	case "bool", "boolean":
		return TypeBool
	case "list":
		return TypeList
	default:
		return TypeUnknown
	}
}

// CastValue attempts to cast a value to a specific type
func CastValue(v Value, targetType TypeKind) (Value, error) {
	// Unwrap TypedValue if needed
	if tv, ok := v.(*TypedValue); ok {
		v = tv.Value
	}

	switch targetType {
	case TypeI32:
		switch val := v.(type) {
		case float64:
			return int32(val), nil
		case int64:
			return int32(val), nil
		case uint32:
			return int32(val), nil
		case uint64:
			return int32(val), nil
		case string:
			var i int32
			_, err := fmt.Sscanf(val, "%d", &i)
			if err != nil {
				return nil, fmt.Errorf("cannot cast string '%s' to i32", val)
			}
			return i, nil
		default:
			return nil, fmt.Errorf("cannot cast %T to i32", v)
		}
	case TypeI64:
		switch val := v.(type) {
		case float64:
			return int64(val), nil
		case int32:
			return int64(val), nil
		case uint32:
			return int64(val), nil
		case uint64:
			return int64(val), nil
		case string:
			var i int64
			_, err := fmt.Sscanf(val, "%d", &i)
			if err != nil {
				return nil, fmt.Errorf("cannot cast string '%s' to i64", val)
			}
			return i, nil
		default:
			return nil, fmt.Errorf("cannot cast %T to i64", v)
		}
	case TypeU32:
		switch val := v.(type) {
		case float64:
			if val < 0 {
				return nil, fmt.Errorf("cannot cast negative number to u32")
			}
			return uint32(val), nil
		case int32:
			if val < 0 {
				return nil, fmt.Errorf("cannot cast negative number to u32")
			}
			return uint32(val), nil
		case int64:
			if val < 0 {
				return nil, fmt.Errorf("cannot cast negative number to u32")
			}
			return uint32(val), nil
		case string:
			var i uint32
			_, err := fmt.Sscanf(val, "%d", &i)
			if err != nil {
				return nil, fmt.Errorf("cannot cast string '%s' to u32", val)
			}
			return i, nil
		default:
			return nil, fmt.Errorf("cannot cast %T to u32", v)
		}
	case TypeF64:
		switch val := v.(type) {
		case int32:
			return float64(val), nil
		case int64:
			return float64(val), nil
		case uint32:
			return float64(val), nil
		case uint64:
			return float64(val), nil
		case float32:
			return float64(val), nil
		case string:
			var f float64
			_, err := fmt.Sscanf(val, "%f", &f)
			if err != nil {
				return nil, fmt.Errorf("cannot cast string '%s' to f64", val)
			}
			return f, nil
		default:
			return nil, fmt.Errorf("cannot cast %T to f64", v)
		}
	case TypeString:
		return ToString(v), nil
	default:
		return nil, fmt.Errorf("unsupported cast to type %v", targetType)
	}
}
