package command

import (
	"bytes"
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
func ExecPython(script string, args ...string) (string, error) {
	return execScript(Python, script, args...)
}

// ExecPython3 executes a Python3 script
//
//	script: The Python3 script content
//	args: Arguments to pass to the script
//	Returns the execution output and any error
func ExecPython3(script string, args ...string) (string, error) {
	return execScript(Python3, script, args...)
}

// ExecShell executes a Shell script
//
//	script: The Shell script content
//	args: Arguments to pass to the script
//	Returns the execution output and any error
func ExecShell(script string, args ...string) (string, error) {
	return execScript(Shell, script, args...)
}

// ExecBash executes a Bash script
//
//	script: The Bash script content
//	args: Arguments to pass to the script
//	Returns the execution output and any error
func ExecBash(script string, args ...string) (string, error) {
	return execScript(Bash, script, args...)
}

// ExecScript executes a script with a given interpreter
//
//	interpreter: The interpreter to use
//	script: The script content
//	args: Arguments to pass to the script
//	Returns the execution output and any error
func ExecScript(interpreter Interpreter, script string, args ...string) (string, error) {
	if interpreter.IsUnknown() {
		return "", fmt.Errorf("interpreter is unknown, valid interpreters:	 %v", interpreters)
	}
	return execScript(interpreter, script, args...)
}

// execScript is a generic script execution function
func execScript(interpreter Interpreter, script string, args ...string) (string, error) {
	// Check if the interpreter exists
	if _, err := exec.LookPath(string(interpreter)); err != nil {
		return "", fmt.Errorf("%s not found: %v", interpreter, err)
	}

	// Create the command
	cmd := exec.Command(string(interpreter), append([]string{"-c", script}, args...)...)

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
