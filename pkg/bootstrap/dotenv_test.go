package bootstrap

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadDotEnvFromConfigDirectoryAncestor(t *testing.T) {
	root := t.TempDir()
	configDir := filepath.Join(root, "app", "alkaid", "configs")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatal(err)
	}
	envPath := filepath.Join(root, ".env")
	if err := os.WriteFile(envPath, []byte("BOOTSTRAP_DOTENV_TEST=from-file\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	const key = "BOOTSTRAP_DOTENV_TEST"
	previous, wasSet := os.LookupEnv(key)
	_ = os.Unsetenv(key)
	t.Cleanup(func() {
		if wasSet {
			_ = os.Setenv(key, previous)
		} else {
			_ = os.Unsetenv(key)
		}
	})

	loaded, err := loadDotEnv(filepath.Join(configDir, "config.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if loaded != envPath {
		t.Fatalf("loadDotEnv() path = %q, want %q", loaded, envPath)
	}
	if got := os.Getenv(key); got != "from-file" {
		t.Fatalf("dotenv value = %q, want %q", got, "from-file")
	}
}

func TestLoadDotEnvPreservesProcessEnvironment(t *testing.T) {
	root := t.TempDir()
	configDir := filepath.Join(root, "configs")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".env"), []byte("BOOTSTRAP_DOTENV_PRECEDENCE=file-value\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("BOOTSTRAP_DOTENV_PRECEDENCE", "process-value")

	if _, err := loadDotEnv(filepath.Join(configDir, "config.yaml")); err != nil {
		t.Fatal(err)
	}
	if got := os.Getenv("BOOTSTRAP_DOTENV_PRECEDENCE"); got != "process-value" {
		t.Fatalf("process environment = %q, want %q", got, "process-value")
	}
}

func TestParseDotEnvCommonSyntax(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".env")
	content := "# comment\nexport TOKEN=abc123\nPLAIN=value # trailing comment\nSINGLE='literal $value'\nDOUBLE=\"line\\nvalue\"\nHASH=abc#123\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	got, err := parseDotEnv(path)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]string{
		"TOKEN":  "abc123",
		"PLAIN":  "value",
		"SINGLE": "literal $value",
		"DOUBLE": "line\nvalue",
		"HASH":   "abc#123",
	}
	for key, expected := range want {
		if got[key] != expected {
			t.Errorf("%s = %q, want %q", key, got[key], expected)
		}
	}
}

func TestFindDotEnvMissingIsOptional(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	path, err := findDotEnv(filepath.Join(root, "configs", "config.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if path != "" {
		t.Fatalf("findDotEnv() = %q, want no dotenv file", path)
	}
}
