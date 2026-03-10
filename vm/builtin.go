package vm

func (ev *Evaluator) evalBuiltinFunction(name string, args []Value) (Value, error) {
	if ev.builtinFn == nil {
		return nil, ev.runtimeError("no built-in evaluator registered for '" + name + "'")
	}
	return ev.builtinFn(name, args)
}
