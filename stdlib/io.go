package stdlib

import (
	"fmt"

	vm "github.com/Advik-B/english/astvm"
	"github.com/Advik-B/english/runtime"
)

func evalIO(name string, args []vm.Value) (vm.Value, error) {
	switch name {
	case "ask":
		if len(args) > 0 {
			fmt.Print(vm.ToString(args[0]))
		}
		// One reader for the process; see runtime.ReadLine.
		return runtime.ReadLine(), nil
	}
	return nil, vm.NewRuntimeError("unknown IO function: " + name)
}
