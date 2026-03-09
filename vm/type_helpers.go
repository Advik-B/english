package vm

import (
	"english/vm/types"
	"fmt"
)

// inferTypeKind determines the TypeKind of a runtime value.
// It extends types.Infer() with vm-only types (FunctionValue, StructInstance, ReferenceValue).
func inferTypeKind(v Value) types.TypeKind {
	switch v.(type) {
	case *FunctionValue:
		return types.TypeFunction
	case *StructInstance:
		return types.TypeStruct
	case *ReferenceValue:
		return types.TypeRef
	default:
		return types.Infer(v)
	}
}

// typeKindName returns the user-facing name for a TypeKind.
// Delegates to types.Name — provided as a vm-local shorthand.
func typeKindName(tk types.TypeKind) string {
	return types.Name(tk)
}

// GetType returns a *types.TypeInfo describing a runtime value.
func GetType(v Value) *types.TypeInfo {
	switch val := v.(type) {
	case *types.TypedValue:
		return val.TypeInfo
	case float64:
		return &types.TypeInfo{Kind: types.TypeF64, Name: "f64"}
	case int32:
		return &types.TypeInfo{Kind: types.TypeI32, Name: "i32"}
	case int64:
		return &types.TypeInfo{Kind: types.TypeI64, Name: "i64"}
	case uint32:
		return &types.TypeInfo{Kind: types.TypeU32, Name: "u32"}
	case uint64:
		return &types.TypeInfo{Kind: types.TypeU64, Name: "u64"}
	case float32:
		return &types.TypeInfo{Kind: types.TypeF32, Name: "f32"}
	case string:
		return &types.TypeInfo{Kind: types.TypeString, Name: "text"}
	case bool:
		return &types.TypeInfo{Kind: types.TypeBool, Name: "boolean"}
	case []interface{}:
		return &types.TypeInfo{Kind: types.TypeList, Name: "list"}
	case *ArrayValue:
		return &types.TypeInfo{
			Kind: types.TypeArray,
			Name: fmt.Sprintf("array of %s", types.Name(val.ElementType)),
		}
	case *LookupTableValue:
		return &types.TypeInfo{Kind: types.TypeLookup, Name: "lookup table"}
	case *FunctionValue:
		return &types.TypeInfo{Kind: types.TypeFunction, Name: "function"}
	case *StructInstance:
		return &types.TypeInfo{Kind: types.TypeStruct, Name: val.Definition.Name}
	case *types.ErrorValue:
		return &types.TypeInfo{Kind: types.TypeError, Name: "error"}
	case *ReferenceValue:
		return &types.TypeInfo{Kind: types.TypeRef, Name: "reference"}
	case nil:
		return &types.TypeInfo{Kind: types.TypeNull, Name: "nothing"}
	default:
		return &types.TypeInfo{Kind: types.TypeUnknown, Name: "unknown"}
	}
}

// CastValue performs an explicit "cast to" conversion.
// For complex types that need the full ToString (arrays, lookup tables),
// it delegates to the vm-level ToString before calling types.Cast.
func CastValue(v Value, target types.TypeKind) (Value, error) {
	if target == types.TypeString {
		return ToString(v), nil
	}
	return types.Cast(v, target)
}
