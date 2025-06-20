package command

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"slices"
	"strings"
)

type Interpreter string

const (
	Python  Interpreter = "python"
	Python3 Interpreter = "python3"
	Shell   Interpreter = "sh"
	Bash    Interpreter = "bash"
	py                  = "py"
	py3                 = "py3"
	shell               = "shell"
)

var (
	interpreters = []Interpreter{Python, Python3, Shell, Bash}
)

func (i Interpreter) IsUnknown() bool {
	return !slices.Contains(interpreters, i)
}

// ExecPython executes a Python script
//
//	script: The Python script content
//	args: Arguments to pass to the script
//	Returns the execution output and any error
func ExecPython(ctx context.Context, script string, args ...string) (string, error) {
	return execScript(ctx, Python, script, args...)
}

// ExecPython3 executes a Python3 script
//
//	script: The Python3 script content
//	args: Arguments to pass to the script
//	Returns the execution output and any error
func ExecPython3(ctx context.Context, script string, args ...string) (string, error) {
	return execScript(ctx, Python3, script, args...)
}

// ExecShell executes a Shell script
//
//	script: The Shell script content
//	args: Arguments to pass to the script
//	Returns the execution output and any error
func ExecShell(ctx context.Context, script string, args ...string) (string, error) {
	return execScript(ctx, Shell, script, args...)
}

// ExecBash executes a Bash script
//
//	script: The Bash script content
//	args: Arguments to pass to the script
//	Returns the execution output and any error
func ExecBash(ctx context.Context, script string, args ...string) (string, error) {
	return execScript(ctx, Bash, script, args...)
}

// ExecScript executes a script with a given interpreter
//
//	interpreter: The interpreter to use
//	script: The script content
//	args: Arguments to pass to the script
//	Returns the execution output and any error
func ExecScript(ctx context.Context, interpreter Interpreter, script string, args ...string) (string, error) {
	if interpreter.IsUnknown() {
		return "", fmt.Errorf("interpreter is unknown, valid interpreters:	 %v", interpreters)
	}
	return execScript(ctx, interpreter, script, args...)
}

// execScript is a generic script execution function
func execScript(ctx context.Context, interpreter Interpreter, script string, args ...string) (string, error) {
	// Check if the interpreter exists
	if _, err := exec.LookPath(string(interpreter)); err != nil {
		return "", fmt.Errorf("%s not found: %v", interpreter, err)
	}

	// Create the command
	cmd := exec.CommandContext(ctx, string(interpreter), append([]string{"-c", script}, args...)...)

	// Capture output
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Execute the command
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("execution failed: %v, stderr: %s", err, stderr.String())
	}

	// Return standard output
	return strings.TrimSpace(stdout.String()), nil
}
