package upstream

import (
	"testing"

	kerrors "github.com/go-kratos/kratos/v3/errors"

	"github.com/aide-family/Alioth/pkg/merr"
)

func TestResolveKeepsEndpointPathPrefix(t *testing.T) {
	cases := []struct {
		name     string
		endpoint string
		apiPath  string
		want     string
	}{
		{
			name:     "root endpoint",
			endpoint: "http://127.0.0.1:9090",
			apiPath:  "/api/v1/query?query=up",
			want:     "http://127.0.0.1:9090/api/v1/query?query=up",
		},
		{
			name:     "victoriametrics cluster prefix",
			endpoint: "http://vmselect:8481/select/0/prometheus",
			apiPath:  "/api/v1/query_range?query=up",
			want:     "http://vmselect:8481/select/0/prometheus/api/v1/query_range?query=up",
		},
		{
			name:     "endpoint already contains api version",
			endpoint: "https://metrics.example.com/prometheus/api/v1/",
			apiPath:  "/api/v1/label/__name__/values",
			want:     "https://metrics.example.com/prometheus/api/v1/label/__name__/values",
		},
		{
			name:     "api path already contains the prefix",
			endpoint: "http://vmselect:8481/select/0/prometheus",
			apiPath:  "/select/0/prometheus/api/v1/query",
			want:     "http://vmselect:8481/select/0/prometheus/api/v1/query",
		},
		{
			name:     "query keeps dots and slashes",
			endpoint: "http://127.0.0.1:9090",
			apiPath:  "/api/v1/query?query=foo%2Ebar",
			want:     "http://127.0.0.1:9090/api/v1/query?query=foo%2Ebar",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Resolve(tc.endpoint, tc.apiPath)
			if err != nil {
				t.Fatal(err)
			}
			if got.String() != tc.want {
				t.Fatalf("Resolve() = %q, want %q", got.String(), tc.want)
			}
		})
	}
}

func TestResolveNamesInvalidDatasourceEndpoint(t *testing.T) {
	_, err := Resolve("htp://es.moon.com", "/_search")
	if !merr.IsInvalid(err) || kerrors.FromError(err).Message != "invalid datasource endpoint" {
		t.Fatalf("Resolve() error = %v", err)
	}
}

func TestResolveRejectsInvalidInput(t *testing.T) {
	cases := []struct {
		name     string
		endpoint string
		apiPath  string
	}{
		{name: "missing scheme", endpoint: "127.0.0.1:9090", apiPath: "/api/v1/query"},
		{name: "unsupported scheme", endpoint: "ftp://127.0.0.1", apiPath: "/api/v1/query"},
		{name: "missing host", endpoint: "http:///metrics", apiPath: "/api/v1/query"},
		{name: "relative api path", endpoint: "http://127.0.0.1:9090", apiPath: "api/v1/query"},
		{name: "path traversal", endpoint: "http://127.0.0.1:9090", apiPath: "/api/../admin"},
		{name: "endpoint path traversal", endpoint: "http://127.0.0.1:9090/a/../api/v1", apiPath: "/api/v1/query"},
		{name: "absolute reference", endpoint: "http://127.0.0.1:9090", apiPath: "//evil.example.com/api/v1/query"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := Resolve(tc.endpoint, tc.apiPath); !merr.IsInvalid(err) {
				t.Fatalf("Resolve() error = %v", err)
			}
		})
	}
}
