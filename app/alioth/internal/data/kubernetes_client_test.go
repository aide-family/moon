package data

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	kerrors "github.com/go-kratos/kratos/v3/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aide-family/moon/pkg/merr"
)

func testLogger() *slog.Logger {
	return slog.New(slog.DiscardHandler)
}

// testKubeconfig 生成一份指向测试服务器的 kubeconfig。
func testKubeconfig(server, contextName, clusterName string) string {
	return fmt.Sprintf(`apiVersion: v1
kind: Config
clusters:
- name: %s
  cluster:
    server: %s
    insecure-skip-tls-verify: true
contexts:
- name: %s
  context:
    cluster: %s
    user: tester
current-context: %s
users:
- name: tester
  user: {}
`, clusterName, server, contextName, clusterName, contextName)
}

func TestParseKubeconfig(t *testing.T) {
	t.Run("yaml", func(t *testing.T) {
		config, clusterName, contextName, err := parseKubeconfig(testKubeconfig("https://api.example.com:6443", "prod", "prod-cluster"))
		if err != nil {
			t.Fatalf("parseKubeconfig: %v", err)
		}
		if config.Host != "https://api.example.com:6443" {
			t.Fatalf("host = %q", config.Host)
		}
		if clusterName != "prod-cluster" || contextName != "prod" {
			t.Fatalf("cluster=%q context=%q", clusterName, contextName)
		}
		if config.Timeout != kubeRequestTimeout {
			t.Fatalf("timeout = %v", config.Timeout)
		}
	})

	t.Run("json", func(t *testing.T) {
		raw := `{
			"apiVersion": "v1",
			"kind": "Config",
			"clusters": [{"name": "json-cluster", "cluster": {"server": "https://api.example.com:6443", "insecure-skip-tls-verify": true}}],
			"contexts": [{"name": "json-context", "context": {"cluster": "json-cluster", "user": "tester"}}],
			"current-context": "json-context",
			"users": [{"name": "tester", "user": {}}]
		}`
		_, clusterName, contextName, err := parseKubeconfig(raw)
		if err != nil {
			t.Fatalf("parseKubeconfig: %v", err)
		}
		if clusterName != "json-cluster" || contextName != "json-context" {
			t.Fatalf("cluster=%q context=%q", clusterName, contextName)
		}
	})

	cases := []struct {
		name       string
		kubeconfig string
	}{
		{name: "empty", kubeconfig: "   "},
		{name: "not yaml", kubeconfig: "this is not a kubeconfig"},
		{name: "missing current context", kubeconfig: `apiVersion: v1
kind: Config
clusters:
- name: c1
  cluster:
    server: https://api.example.com`},
		{name: "unknown context", kubeconfig: `apiVersion: v1
kind: Config
clusters:
- name: c1
  cluster:
    server: https://api.example.com
contexts:
- name: ctx
  context:
    cluster: c1
current-context: missing`},
		{name: "context without cluster", kubeconfig: `apiVersion: v1
kind: Config
clusters:
- name: c1
  cluster:
    server: https://api.example.com
contexts:
- name: ctx
  context:
    user: tester
current-context: ctx`},
		{name: "cluster not defined", kubeconfig: `apiVersion: v1
kind: Config
clusters:
- name: c1
  cluster:
    server: https://api.example.com
contexts:
- name: ctx
  context:
    cluster: missing
current-context: ctx`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, _, _, err := parseKubeconfig(tc.kubeconfig); !merr.IsInvalid(err) {
				t.Fatalf("err = %v, want invalid", err)
			}
		})
	}
}

func TestParseKubeconfigRejectsUnsafeCredentials(t *testing.T) {
	base := func(userBlock string) string {
		return fmt.Sprintf(`apiVersion: v1
kind: Config
clusters:
- name: c1
  cluster:
    server: https://api.example.com:6443
    insecure-skip-tls-verify: true
contexts:
- name: ctx
  context:
    cluster: c1
    user: tester
current-context: ctx
users:
- name: tester
  user:
%s`, userBlock)
	}
	cases := []struct {
		name       string
		kubeconfig string
	}{
		{name: "exec", kubeconfig: base(`    exec:
      apiVersion: client.authentication.k8s.io/v1
      command: /bin/sh
      args: ["-c", "true"]`)},
		{name: "auth provider", kubeconfig: base(`    auth-provider:
      name: gcp`)},
		{name: "token file", kubeconfig: base(`    tokenFile: /proc/self/environ`)},
		{name: "client certificate file", kubeconfig: base(`    client-certificate: /etc/passwd`)},
		{name: "client key file", kubeconfig: base(`    client-key: /etc/passwd`)},
		{name: "proxy url", kubeconfig: `apiVersion: v1
kind: Config
clusters:
- name: c1
  cluster:
    server: https://api.example.com:6443
    proxy-url: http://127.0.0.1:9
contexts:
- name: ctx
  context:
    cluster: c1
current-context: ctx`},
		{name: "certificate authority file", kubeconfig: `apiVersion: v1
kind: Config
clusters:
- name: c1
  cluster:
    server: https://api.example.com:6443
    certificate-authority: /etc/passwd
contexts:
- name: ctx
  context:
    cluster: c1
current-context: ctx`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, _, _, err := parseKubeconfig(tc.kubeconfig); !merr.IsInvalid(err) {
				t.Fatalf("err = %v, want invalid", err)
			}
		})
	}
}

func TestProbeKubeconfig(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/version":
				writeJSON(w, `{"major":"1","minor":"35","gitVersion":"v1.35.5"}`)
			case "/api/v1/nodes":
				writeJSON(w, `{"apiVersion":"v1","kind":"NodeList","metadata":{},"items":[{"metadata":{"name":"node-a"}},{"metadata":{"name":"node-b"}}]}`)
			default:
				http.NotFound(w, r)
			}
		}))
		defer server.Close()
		probe, err := probeKubeconfig(context.Background(), testLogger(), testKubeconfig(server.URL, "ctx-a", "cluster-a"))
		if err != nil {
			t.Fatalf("probeKubeconfig: %v", err)
		}
		if probe.Server != server.URL || probe.Version != "v1.35.5" {
			t.Fatalf("server=%q version=%q", probe.Server, probe.Version)
		}
		if probe.ClusterName != "cluster-a" || probe.ContextName != "ctx-a" {
			t.Fatalf("cluster=%q context=%q", probe.ClusterName, probe.ContextName)
		}
		if probe.NodeCount != 2 {
			t.Fatalf("node count = %d", probe.NodeCount)
		}
	})

	t.Run("nodes forbidden", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/version":
				writeJSON(w, `{"gitVersion":"v1.30.0"}`)
			case "/api/v1/nodes":
				w.WriteHeader(http.StatusForbidden)
				writeJSON(w, `{"apiVersion":"v1","kind":"Status","status":"Failure","reason":"Forbidden","code":403,"message":"forbidden"}`)
			default:
				http.NotFound(w, r)
			}
		}))
		defer server.Close()
		probe, err := probeKubeconfig(context.Background(), testLogger(), testKubeconfig(server.URL, "ctx", "cluster"))
		if err != nil {
			t.Fatalf("probeKubeconfig: %v", err)
		}
		if probe.NodeCount != 0 || probe.Version != "v1.30.0" {
			t.Fatalf("node count=%d version=%q", probe.NodeCount, probe.Version)
		}
	})

	t.Run("version failure", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			writeJSON(w, `{"apiVersion":"v1","kind":"Status","status":"Failure","reason":"InternalError","code":500,"message":"boom"}`)
		}))
		defer server.Close()
		_, err := probeKubeconfig(context.Background(), testLogger(), testKubeconfig(server.URL, "ctx", "cluster"))
		if err == nil {
			t.Fatal("want error")
		}
		if reason := kerrors.FromError(err).Reason; reason != merr.ReasonServer {
			t.Fatalf("reason = %q, want %q", reason, merr.ReasonServer)
		}
	})

	t.Run("unreachable", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			writeJSON(w, `{"gitVersion":"v1.30.0"}`)
		}))
		url := server.URL
		server.Close()
		_, err := probeKubeconfig(context.Background(), testLogger(), testKubeconfig(url, "ctx", "cluster"))
		if err == nil {
			t.Fatal("want error")
		}
		if reason := kerrors.FromError(err).Reason; reason != merr.ReasonClient {
			t.Fatalf("reason = %q, want %q", reason, merr.ReasonClient)
		}
	})

	t.Run("invalid kubeconfig", func(t *testing.T) {
		_, err := probeKubeconfig(context.Background(), testLogger(), "garbage")
		if !merr.IsInvalid(err) {
			t.Fatalf("err = %v, want invalid", err)
		}
	})
}

func TestKubeErrorMapping(t *testing.T) {
	cases := []struct {
		name   string
		err    error
		reason string
	}{
		{name: "not found", err: newStatusError(404, metav1.StatusReasonNotFound), reason: merr.ReasonNotFound},
		// 集群侧 401/403 不能映射成平台未认证/无权限：前者会清掉登录态，
		// 后者会让前端跳到 /forbidden，用户看不到受限 kubeconfig 的真实原因。
		{name: "unauthorized", err: newStatusError(401, metav1.StatusReasonUnauthorized), reason: merr.ReasonInvalid},
		{name: "forbidden", err: newStatusError(403, metav1.StatusReasonForbidden), reason: merr.ReasonClient},
		{name: "invalid", err: newStatusError(422, metav1.StatusReasonInvalid), reason: merr.ReasonInvalid},
		{name: "conflict", err: newStatusError(409, metav1.StatusReasonConflict), reason: merr.ReasonConflict},
		{name: "too many requests", err: newStatusError(429, metav1.StatusReasonTooManyRequests), reason: merr.ReasonServer},
		{name: "service unavailable", err: newStatusError(503, metav1.StatusReasonServiceUnavailable), reason: merr.ReasonServer},
		{name: "server", err: newStatusError(500, metav1.StatusReasonInternalError), reason: merr.ReasonServer},
		{name: "timeout", err: context.DeadlineExceeded, reason: merr.ReasonClient},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mapped := kubeError(testLogger(), tc.err)
			if reason := kerrors.FromError(mapped).Reason; reason != tc.reason {
				t.Fatalf("reason = %q, want %q", reason, tc.reason)
			}
		})
	}
	if kubeError(testLogger(), nil) != nil {
		t.Fatal("nil error should stay nil")
	}
	already := merr.NotFound("not found")
	if kubeError(testLogger(), already) != already {
		t.Fatal("kratos error should pass through")
	}
}

// newStatusError 构造 client-go 风格的 API 错误。
func newStatusError(code int32, reason metav1.StatusReason) error {
	return &apierrors.StatusError{ErrStatus: metav1.Status{
		Status: metav1.StatusFailure, Code: code, Reason: reason, Message: string(reason),
	}}
}

func writeJSON(w http.ResponseWriter, body string) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(body))
}
