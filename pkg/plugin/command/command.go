package command

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// ExecPython executes a Python script
//
//	script: The Python script content
//	args: Arguments to pass to the script
//	Returns the execution output and any error
func ExecPython(script string, args ...string) (string, error) {
	return execScript("python", script, args...)
}

// ExecPython3 executes a Python3 script
//
//	script: The Python3 script content
//	args: Arguments to pass to the script
//	Returns the execution output and any error
func ExecPython3(script string, args ...string) (string, error) {
	return execScript("python3", script, args...)
}

// ExecShell executes a Shell script
//
//	script: The Shell script content
//	args: Arguments to pass to the script
//	Returns the execution output and any error
func ExecShell(script string, args ...string) (string, error) {
	return execScript("sh", script, args...)
}

// ExecBash executes a Bash script
//
//	script: The Bash script content
//	args: Arguments to pass to the script
//	Returns the execution output and any error
func ExecBash(script string, args ...string) (string, error) {
	return execScript("bash", script, args...)
}

// execScript is a generic script execution function
func execScript(interpreter, script string, args ...string) (string, error) {
	// Check if the interpreter exists
	if _, err := exec.LookPath(interpreter); err != nil {
		return "", fmt.Errorf("%s not found: %v", interpreter, err)
	}

	// Create the command
	cmd := exec.Command(interpreter, append([]string{"-c", script}, args...)...)

	// Capture output
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Execute the command
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("execution failed: %v, stderr: %s", err, stderr.String())
	}

	// Return standard output
	return strings.TrimSpace(stdout.String()), nil
}
