package vm

import "fmt"

func (ev *Evaluator) evalBuiltinFunction(name string, args []Value) (Value, error) {
	if ev.builtinFn == nil {
		return nil, fmt.Errorf("RuntimeError: no built-in evaluator registered for '%s'", name)
	}
	return ev.builtinFn(name, args)
}
