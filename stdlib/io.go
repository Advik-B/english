package stdlib

import (
	"bufio"
	"github.com/Advik-B/english/astvm"
	"fmt"
	"os"
	"strings"
)

func evalIO(name string, args []vm.Value) (vm.Value, error) {
	switch name {
	case "ask":
		if len(args) > 0 {
			fmt.Print(vm.ToString(args[0]))
		}
		reader := bufio.NewReader(os.Stdin)
		line, err := reader.ReadString('\n')
		if err != nil && len(line) == 0 {
			return "", nil
		}
		return strings.TrimRight(line, "\r\n"), nil
	}
	return nil, vm.NewRuntimeError("unknown IO function: " + name)
}

func registerIOFunctions(env *vm.Environment) {
	env.DefineFunction("ask", &vm.FunctionValue{Name: "ask", Parameters: []string{"prompt"}, Body: nil, Closure: env})
}
