// Package stdlib provides the standard library for the English language runtime.
// It is shared by both the AST VM and the instruction VM.
package stdlib

import (
	"math"

	vm "github.com/Advik-B/english/astvm"
)

// predefinedConstants are the stdlib-provided constants visible to programs.
// PredefinedNames and PredefinedValues both derive from this map so the two can
// never disagree.
var predefinedConstants = map[string]interface{}{
	"pi":       math.Pi,
	"e":        math.E,
	"infinity": math.Inf(1),
}

// Register registers all standard library constants and functions into env.
// Function stubs are derived from signatureList, so every built-in is
// registered with exactly the arity its signature declares.
func Register(env *vm.Environment) {
	for name, value := range predefinedConstants {
		env.Define(name, value, true)
	}
	for _, s := range signatureList {
		env.DefineFunction(s.Name, &vm.FunctionValue{
			Name:       s.Name,
			Parameters: s.Params,
			Body:       nil, // nil body marks a built-in
			Closure:    env,
		})
	}
}

// Eval evaluates a built-in function by name with the provided arguments.
// The argument count is validated against the function's signature before
// dispatch, so implementations may index args positionally without risking a
// panic on a short call.
func Eval(name string, args []vm.Value) (vm.Value, error) {
	s, ok := Lookup(name)
	if !ok {
		return nil, vm.NewRuntimeError("unknown built-in function: " + name)
	}
	if err := checkArity(name, len(args)); err != nil {
		return nil, err
	}

	switch s.Module {
	case "math":
		return evalMath(name, args)
	case "string":
		return evalString(name, args)
	case "number":
		return evalNumber(name, args)
	case "list":
		return evalList(name, args)
	case "io":
		return evalIO(name, args)
	case "lookup":
		return evalLookup(name, args)
	case "time":
		return evalTime(name, args)
	}
	return nil, vm.NewRuntimeError("unknown built-in function: " + name)
}

// PredefinedNames returns the names of all constants registered by the stdlib.
// Pass these to the type checker so it can catch redeclarations.
func PredefinedNames() []string {
	out := make([]string, 0, len(predefinedConstants))
	for name := range predefinedConstants {
		out = append(out, name)
	}
	return out
}

// PredefinedValues returns all constants registered by the stdlib as a map.
// Used by ivm.Machine to initialize predefined constants.
func PredefinedValues() map[string]interface{} {
	out := make(map[string]interface{}, len(predefinedConstants))
	for name, value := range predefinedConstants {
		out[name] = value
	}
	return out
}
