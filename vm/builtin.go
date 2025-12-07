package vm

// evalBuiltinFunctionWrapper is a wrapper to call the package-level evalBuiltinFunction
func (ev *Evaluator) evalBuiltinFunction(name string, args []Value) (Value, error) {
	return evalBuiltinFunction(name, args)
}
