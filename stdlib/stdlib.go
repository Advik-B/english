// Package stdlib provides the standard library for the English language runtime.
// It is shared by both the AST VM and the instruction VM.
package stdlib

import (
	"github.com/Advik-B/english/astvm"
	"math"
)

// Register registers all standard library functions into env.
func Register(env *vm.Environment) {
	registerMathConstants(env)
	registerMathFunctions(env)
	registerStringFunctions(env)
	registerListFunctions(env)
	registerIOFunctions(env)
	registerLookupTableFunctions(env)
	registerNumberFunctions(env)
	registerTimeFunctions(env)
}

// Eval evaluates a built-in function by name with the provided arguments.
func Eval(name string, args []vm.Value) (vm.Value, error) {
	switch name {
	// ── Math ──────────────────────────────────────────────────────────────────
	case "sqrt", "pow", "abs", "floor", "ceil", "round", "min", "max",
		"sin", "cos", "tan", "log", "log10", "log2", "exp",
		"random", "random_between", "is_nan", "is_infinite":
		return evalMath(name, args)

	// ── String ────────────────────────────────────────────────────────────────
	case "uppercase", "lowercase", "casefold", "split", "join", "trim",
		"replace", "contains", "starts_with", "ends_with", "index_of",
		"substring", "str_repeat", "count_occurrences", "pad_left", "pad_right",
		"to_number", "to_string", "is_empty",
		"title", "capitalize", "swapcase", "trim_left", "trim_right",
		"is_digit", "is_alpha", "is_alnum", "is_space", "is_upper", "is_lower",
		"center", "zfill":
		return evalString(name, args)

	// ── Number ────────────────────────────────────────────────────────────────
	case "is_integer", "clamp", "sign":
		return evalNumber(name, args)

	// ── List ──────────────────────────────────────────────────────────────────
	case "append", "remove", "insert", "sort", "reverse", "sum", "unique",
		"first", "last", "flatten", "count", "slice",
		"average", "min_value", "max_value", "any_true", "all_true",
		"product", "sorted_desc", "zip_with":
		return evalList(name, args)

	// ── I/O ───────────────────────────────────────────────────────────────────
	case "ask":
		return evalIO(name, args)

	// ── Lookup table ──────────────────────────────────────────────────────────
	case "keys", "values", "table_remove", "table_has", "merge", "get_or_default":
		return evalLookup(name, args)

	// ── Time ──────────────────────────────────────────────────────────────────
	case "current_time", "elapsed_time", "sleep":
		return evalTime(name, args)
	}

	return nil, vm.NewRuntimeError("unknown built-in function: " + name)
}

// PredefinedNames returns the names of all constants registered by the stdlib.
// Pass these to vm.Check so the compile-time checker can catch redeclarations.
func PredefinedNames() []string {
	return []string{"pi", "e", "infinity"}
}

// PredefinedValues returns all constants registered by the stdlib as a map.
// Used by ivm.Machine to initialize predefined constants.
func PredefinedValues() map[string]interface{} {
	return map[string]interface{}{
		"pi":       math.Pi,
		"e":        math.E,
		"infinity": math.Inf(1),
	}
}
