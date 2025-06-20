package command_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aide-family/moon/pkg/plugin/command"
)

// example usage
func TestCommand(t *testing.T) {
	ctx := context.Background()
	// exec python script
	pyOutput, err := command.ExecPython(ctx, "print 'Hello from Python!'")
	if err != nil {
		fmt.Printf("Python error: %v\n", err)
	} else {
		fmt.Printf("Python output: %s\n", pyOutput)
	}

	// exec python3 script
	py3Output, err := command.ExecPython3(ctx, "print('Hello from Python3!')")
	if err != nil {
		fmt.Printf("Python3 error: %v\n", err)
	} else {
		fmt.Printf("Python3 output: %s\n", py3Output)
	}

	// exec shell script
	shellOutput, err := command.ExecShell(ctx, "echo 'Hello from Shell!'")
	if err != nil {
		fmt.Printf("Shell error: %v\n", err)
	} else {
		fmt.Printf("Shell output: %s\n", shellOutput)
	}

	// exec bash script
	bashOutput, err := command.ExecBash(ctx, "echo 'Hello from Bash!'")
	if err != nil {
		fmt.Printf("Bash error: %v\n", err)
	} else {
		fmt.Printf("Bash output: %s\n", bashOutput)
	}

	// exec python script with parameters
	sumPy := `
import sys
a = int(sys.argv[1])
b = int(sys.argv[2])
print(a + b)
`
	sumOutput, err := command.ExecPython3(ctx, sumPy, "5", "7")
	if err != nil {
		fmt.Printf("Sum calculation error: %v\n", err)
	} else {
		fmt.Printf("Sum result: %s\n", sumOutput)
	}
}
