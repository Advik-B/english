package vm

import "github.com/Advik-B/english/types"

func (ev *Evaluator) evalBuiltinFunction(name string, args []Value) (Value, error) {
	if ev.builtinFn == nil {
		return nil, ev.runtimeError("no built-in evaluator registered for '" + name + "'")
	}
	value, err := ev.builtinFn(name, args)
	if err != nil {
		// A failure inside the standard library belongs to the call site. The
		// library's own error carried a call stack of just "<stdlib>", which
		// replaced the real one, so a failure deep inside a program reported
		// no useful frames at all.
		return nil, ev.wrapBuiltinError(err)
	}
	return value, nil
}

// wrapBuiltinError attaches the current location and call stack to an error
// from the standard library, leaving a raised error value alone: that is a
// value the program may catch, not a failure of the call.
func (ev *Evaluator) wrapBuiltinError(err error) error {
	switch e := err.(type) {
	case *types.ErrorValue:
		return e
	case *RuntimeError:
		if len(e.CallStack) == 1 && e.CallStack[0] == "<stdlib>" {
			return ev.runtimeError(e.Message)
		}
		return e
	}
	return ev.runtimeError(err.Error())
}
