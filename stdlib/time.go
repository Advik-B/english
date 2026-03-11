package stdlib

import (
	"english/astvm"
	"fmt"
	"time"
)

// programStart records the time the stdlib was first loaded.
var programStart = time.Now()

func evalTime(name string, args []vm.Value) (vm.Value, error) {
	switch name {
	case "current_time":
		return time.Now().Format("2006-01-02 15:04:05"), nil

	case "elapsed_time":
		// Round to microsecond precision to avoid nanosecond float-noise.
		return time.Since(programStart).Round(time.Microsecond).Seconds(), nil

	case "sleep":
		if len(args) == 0 {
			return nil, fmt.Errorf("TypeError: sleep expects a number of seconds")
		}
		secs, err := requireNumber("sleep", args[0])
		if err != nil {
			return nil, err
		}
		if secs < 0 {
			return nil, fmt.Errorf("TypeError: sleep expects a non-negative number of seconds")
		}
		time.Sleep(time.Duration(secs * float64(time.Second)))
		return nil, nil
	}
	return nil, vm.NewRuntimeError("unknown time function: " + name)
}

func registerTimeFunctions(env *vm.Environment) {
	env.DefineFunction("current_time", &vm.FunctionValue{Name: "current_time", Parameters: []string{}, Body: nil, Closure: env})
	env.DefineFunction("elapsed_time", &vm.FunctionValue{Name: "elapsed_time", Parameters: []string{}, Body: nil, Closure: env})
	env.DefineFunction("sleep", &vm.FunctionValue{Name: "sleep", Parameters: []string{"seconds"}, Body: nil, Closure: env})
}
