package command

import (
	"path/filepath"
	"strings"
	"time"
)

func getFilename(file string) string {
	return strings.TrimSuffix(filepath.Base(file), filepath.Ext(file))
}

// FileType returns the file type of the file
func GetFileTypeByPrefix(file string, index int) Interpreter {
	// trim extension
	filename := getFilename(file)
	parts := strings.Split(filename, "_")
	if len(parts) < index+1 {
		return ""
	}
	interpreter := parts[index]
	switch Interpreter(interpreter) {
	case Python:
		return Python
	case Python3:
		return Python3
	case Shell:
		return Shell
	case Bash:
		return Bash
	}
	return ""
}

func GetIntervalByPrefix(file string, index int) time.Duration {
	filename := getFilename(file)
	parts := strings.Split(filename, "_")
	if len(parts) < index+1 {
		return 0
	}
	// Convert 10s, 5s to time.Duration
	interval, err := time.ParseDuration(parts[index])
	if err != nil {
		return 0
	}
	return interval
}
