package bootstrap

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// loadDotEnv loads the nearest .env file at or above the bootstrap config file.
// Existing process environment variables always take precedence.
func loadDotEnv(configPath string) (string, error) {
	path, err := findDotEnv(configPath)
	if err != nil || path == "" {
		return path, err
	}

	values, err := parseDotEnv(path)
	if err != nil {
		return path, err
	}
	for key, value := range values {
		if _, exists := os.LookupEnv(key); exists {
			continue
		}
		if err := os.Setenv(key, value); err != nil {
			return path, fmt.Errorf("set dotenv variable %s: %w", key, err)
		}
	}
	return path, nil
}

// findDotEnv searches from the config directory upward. It stops after checking
// the repository root when a .git entry is found, avoiding unrelated parent .env files.
func findDotEnv(configPath string) (string, error) {
	absPath, err := filepath.Abs(configPath)
	if err != nil {
		return "", err
	}
	dir := filepath.Dir(absPath)
	for {
		candidate := filepath.Join(dir, ".env")
		info, statErr := os.Stat(candidate)
		if statErr == nil {
			if !info.Mode().IsRegular() {
				return "", fmt.Errorf("dotenv path is not a regular file: %s", candidate)
			}
			return candidate, nil
		}
		if !os.IsNotExist(statErr) {
			return "", fmt.Errorf("inspect dotenv file: %w", statErr)
		}

		if _, gitErr := os.Stat(filepath.Join(dir, ".git")); gitErr == nil {
			return "", nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", nil
		}
		dir = parent
	}
}

// parseDotEnv handles the common dotenv syntax: KEY=value, optional export,
// comments, and single-, double-, or backtick-quoted values.
func parseDotEnv(path string) (map[string]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	values := make(map[string]string)
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 4096), 1024*1024)
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		key, value, ok, err := parseDotEnvLine(scanner.Text())
		if err != nil {
			return nil, fmt.Errorf("parse %s:%d: %w", path, lineNumber, err)
		}
		if ok {
			values[key] = value
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read dotenv file %s: %w", path, err)
	}
	return values, nil
}

func parseDotEnvLine(line string) (key, value string, ok bool, err error) {
	line = strings.TrimSpace(line)
	if line == "" || strings.HasPrefix(line, "#") {
		return "", "", false, nil
	}
	if strings.HasPrefix(line, "export ") {
		line = strings.TrimSpace(strings.TrimPrefix(line, "export "))
	}
	equal := strings.IndexByte(line, '=')
	if equal < 0 {
		return "", "", false, fmt.Errorf("expected KEY=value")
	}
	key = strings.TrimSpace(line[:equal])
	if !validEnvKey(key) {
		return "", "", false, fmt.Errorf("invalid variable name")
	}
	raw := strings.TrimSpace(line[equal+1:])
	if raw == "" {
		return key, "", true, nil
	}

	quote := raw[0]
	if quote == '\'' || quote == '"' || quote == '`' {
		end := -1
		escaped := false
		for i := 1; i < len(raw); i++ {
			if quote == '"' && raw[i] == '\\' && !escaped {
				escaped = true
				continue
			}
			if raw[i] == quote && !escaped {
				end = i
				break
			}
			escaped = false
		}
		if end < 0 {
			return "", "", false, fmt.Errorf("unterminated quoted value")
		}
		tail := strings.TrimSpace(raw[end+1:])
		if tail != "" && !strings.HasPrefix(tail, "#") {
			return "", "", false, fmt.Errorf("unexpected text after quoted value")
		}
		quoted := raw[1:end]
		switch quote {
		case '"':
			value, err = strconv.Unquote(raw[:end+1])
			if err != nil {
				return "", "", false, fmt.Errorf("invalid quoted value")
			}
		case '\'':
			value = quoted
		case '`':
			value = quoted
		}
		return key, value, true, nil
	}

	// A hash starts a comment only when separated from the value by whitespace.
	for i := 1; i < len(raw); i++ {
		if raw[i] == '#' && (raw[i-1] == ' ' || raw[i-1] == '\t') {
			raw = raw[:i]
			break
		}
	}
	return key, strings.TrimSpace(raw), true, nil
}

func validEnvKey(key string) bool {
	if key == "" {
		return false
	}
	for i, r := range key {
		if (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') || r == '_' || (i > 0 && r >= '0' && r <= '9') {
			continue
		}
		return false
	}
	return true
}
