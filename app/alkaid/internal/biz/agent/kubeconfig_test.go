package agent

import (
	"os"
	"path/filepath"
	"testing"
)

func TestKubeConfigConnection(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config")
	content := `apiVersion: v1
kind: Config
current-context: test
clusters:
  - name: test
    cluster:
      server: https://127.0.0.1:6443
      insecure-skip-tls-verify: true
users:
  - name: test
    user:
      token: test-token
contexts:
  - name: test
    context:
      cluster: test
      user: test
`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("ALKAID_KUBECONFIG", path)
	conn, err := k8sConnectionFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	if conn.base.Host != "127.0.0.1:6443" || conn.token != "test-token" {
		t.Fatalf("conn = %+v", conn)
	}
}

func TestKubeConfigRejectsInvalidFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config")
	if err := os.WriteFile(path, []byte("not yaml: ["), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("ALKAID_KUBECONFIG", path)
	if _, err := k8sConnectionFromEnv(); err == nil {
		t.Fatal("invalid kubeconfig was accepted")
	}
}

func TestDefaultK8sConnection(t *testing.T) {
	t.Setenv("ALKAID_KUBECONFIG", "")
	t.Setenv("ALKAID_K8S_API", "https://127.0.0.1:6443")
	conn, err := k8sConnectionFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	if conn.base.Host != "127.0.0.1:6443" {
		t.Fatalf("base = %s", conn.base)
	}
}

func TestSSHHostKeyCallback(t *testing.T) {
	t.Setenv("ALKAID_SSH_KNOWN_HOSTS", "")
	if _, err := sshHostKeyCallback(); err != nil {
		t.Fatalf("empty known_hosts should fall back: %v", err)
	}
	t.Setenv("ALKAID_SSH_KNOWN_HOSTS", filepath.Join(t.TempDir(), "missing"))
	if _, err := sshHostKeyCallback(); err == nil {
		t.Fatal("missing known_hosts was accepted")
	}
}
