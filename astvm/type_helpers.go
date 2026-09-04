package vm

import (
	"github.com/Advik-B/english/astvm/types"
)

// inferTypeKind determines the TypeKind of a runtime value.
// Engine-defined values (FunctionValue, StructInstance, ReferenceValue) are
// recognised via their EnglishType method, so this is a straight delegation.
func inferTypeKind(v Value) types.TypeKind {
	return types.Infer(v)
}

// typeKindName returns the user-facing name for a TypeKind.
// Delegates to types.Name — provided as a vm-local shorthand.
func typeKindName(tk types.TypeKind) string {
	return types.Name(tk)
}

// GetType returns a *types.TypeInfo describing a runtime value.
func GetType(v Value) *types.TypeInfo {
	return types.Describe(v)
}

// CastValue performs an explicit "cast to" conversion.
// Casting to text uses the vm-level ToString so that composite values (arrays,
// lookup tables, struct instances) render fully rather than via the limited
// renderer inside the types package.
func CastValue(v Value, target types.TypeKind) (Value, error) {
	if target == types.TypeString {
		return ToString(v), nil
	}
	return types.Cast(v, target)
}
