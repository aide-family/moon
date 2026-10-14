package bootstrap

import "testing"

func TestExpand(t *testing.T) {
	t.Setenv("ALIOTH_HTTP_ADDR", "127.0.0.1:18000")
	got := Expand("addr: ${ALIOTH_HTTP_ADDR:0.0.0.0:8000}\nredis: ${MISSING_REDIS:127.0.0.1:6379}\n")
	want := "addr: 127.0.0.1:18000\nredis: 127.0.0.1:6379\n"
	if got != want {
		t.Fatalf("Expand() = %q, want %q", got, want)
	}
}

func TestExpandEmptyOverridesDefault(t *testing.T) {
	t.Setenv("ALIOTH_REDIS_PASSWORD", "")
	got := Expand("password: ${ALIOTH_REDIS_PASSWORD:secret}")
	if got != "password: " {
		t.Fatalf("Expand() = %q", got)
	}
}
