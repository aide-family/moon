package grpctls

import (
	"os"
	"path/filepath"
	"testing"
)

func TestClientDisabled(t *testing.T) {
	cfg, err := ClientFromEnv("TEST_GRPC_TLS")
	if err != nil || cfg != nil {
		t.Fatalf("cfg=%v err=%v, want nil nil", cfg, err)
	}
}

func TestClientRequiresCertAndKey(t *testing.T) {
	t.Setenv("TEST_GRPC_TLS_CERT", "/tmp/client.crt")
	if _, err := ClientFromEnv("TEST_GRPC_TLS"); err == nil {
		t.Fatal("cert without key was accepted")
	}
}

func TestClientRejectsBrokenCA(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "ca.pem")
	if err := os.WriteFile(path, []byte("not a certificate"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("TEST_GRPC_TLS_CA", path)
	if _, err := ClientFromEnv("TEST_GRPC_TLS"); err == nil {
		t.Fatal("invalid CA was accepted")
	}
}

func TestServerRequiresCertAndKey(t *testing.T) {
	t.Setenv("ALIOTH_GRPC_TLS_CERT", "/tmp/server.crt")
	if _, err := ServerFromEnv(); err == nil {
		t.Fatal("server cert without key was accepted")
	}
}

func TestServerDisabled(t *testing.T) {
	cfg, err := ServerFromEnv()
	if err != nil || cfg != nil {
		t.Fatalf("cfg=%v err=%v, want nil nil", cfg, err)
	}
}
