package agent

import (
	"fmt"
	"net/url"
	"path/filepath"
	"strings"
	"testing"

	v1 "github.com/aide-family/Alioth/api/alioth/v1"
)

func TestExecRunsLocallyWithoutSSHCredentials(t *testing.T) {
	workdir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	result := (&Connector{}).exec(&v1.AgentExecRequest{
		RequestUid: "request-1",
		Command:    `printf '%s:%s' "$(pwd)" "$AGENT_TEST_VALUE"`,
		Workdir:    workdir,
		Env:        `{"AGENT_TEST_VALUE":"available"}`,
	})
	if result.GetRequestUid() != "request-1" {
		t.Fatalf("request uid = %q, want request-1", result.GetRequestUid())
	}
	if result.GetExitCode() != 0 || result.GetError() != "" {
		t.Fatalf("execution failed: exit=%d error=%q output=%q", result.GetExitCode(), result.GetError(), result.GetOutput())
	}
	if got, want := strings.TrimSpace(result.GetOutput()), workdir+":available"; got != want {
		t.Fatalf("output = %q, want %q", got, want)
	}
}

func TestExecExpandsHomeWorkdir(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	result := (&Connector{}).exec(&v1.AgentExecRequest{
		RequestUid: "home-request",
		Command:    "pwd -P",
		Workdir:    "~",
	})
	if result.GetExitCode() != 0 || result.GetError() != "" {
		t.Fatalf("execution failed: exit=%d error=%q output=%q", result.GetExitCode(), result.GetError(), result.GetOutput())
	}
	canonicalHome, err := filepath.EvalSymlinks(home)
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.TrimSpace(result.GetOutput()); got != canonicalHome {
		t.Fatalf("output = %q, want home directory %q", got, canonicalHome)
	}
}

func TestK8STargetRejectsHostInjection(t *testing.T) {
	base, err := url.Parse("https://kubernetes.default.svc")
	if err != nil {
		t.Fatal(err)
	}
	bad := []string{
		"",
		"api/v1/pods",
		"@evil.example/x",
		"//evil.example/x",
		"https://evil.example/x",
		"/api/../secrets",
		`/api\v1\pods`,
	}
	for _, path := range bad {
		if _, err := k8sTarget(base, path, ""); err == nil {
			t.Fatalf("k8sTarget(%q) accepted an unsafe path", path)
		}
	}
}

func TestK8STargetKeepsBaseHost(t *testing.T) {
	base, err := url.Parse("https://kubernetes.default.svc")
	if err != nil {
		t.Fatal(err)
	}
	target, err := k8sTarget(base, "/api/v1/pods", "labelSelector=app%3Dapi")
	if err != nil {
		t.Fatal(err)
	}
	if target.Host != "kubernetes.default.svc" || target.Scheme != "https" {
		t.Fatalf("target = %s", target)
	}
	if target.Path != "/api/v1/pods" || target.RawQuery != "labelSelector=app%3Dapi" {
		t.Fatalf("target = %s", target)
	}
}

func TestK8SAllowedHeader(t *testing.T) {
	for _, name := range []string{"Accept", "content-type", "Accept-Encoding"} {
		if !k8sAllowedHeader(name) {
			t.Fatalf("%s should be allowed", name)
		}
	}
	for _, name := range []string{"Authorization", "Host", "Impersonate-User", "X-Forwarded-For"} {
		if k8sAllowedHeader(name) {
			t.Fatalf("%s should be rejected", name)
		}
	}
}

func TestParseEndpoints(t *testing.T) {
	got := parseEndpoints(" alioth-0:9000,alioth-1:9000 ,, ")
	if len(got) != 2 || got[0] != "alioth-0:9000" || got[1] != "alioth-1:9000" {
		t.Fatalf("parseEndpoints = %v", got)
	}
}

func TestExpandEndpoint(t *testing.T) {
	lookup := func(host string) ([]string, error) {
		if host != "alioth-headless" {
			t.Fatalf("unexpected host %s", host)
		}
		return []string{"10.0.0.1", "10.0.0.2"}, nil
	}
	got := expandEndpoint("alioth-headless:9000", lookup)
	if len(got) != 2 || got[0] != "10.0.0.1:9000" || got[1] != "10.0.0.2:9000" {
		t.Fatalf("expandEndpoint = %v", got)
	}
	if got := expandEndpoint("10.0.0.9:9000", lookup); len(got) != 1 || got[0] != "10.0.0.9:9000" {
		t.Fatalf("literal ip = %v", got)
	}
	failing := func(string) ([]string, error) { return nil, fmt.Errorf("dns down") }
	if got := expandEndpoint("alioth:9000", failing); len(got) != 1 || got[0] != "alioth:9000" {
		t.Fatalf("dns fallback = %v", got)
	}
}
