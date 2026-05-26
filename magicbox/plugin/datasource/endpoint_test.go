package datasource

import "testing"

func TestNormalizeMetricEndpoint(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"http://localhost:9090", "http://localhost:9090"},
		{"http://localhost:9090/", "http://localhost:9090"},
		{"http://localhost:9090/api/v1", "http://localhost:9090"},
		{"http://localhost:9090/api/v1/", "http://localhost:9090"},
		{"http://vm:8481/select/0/prometheus", "http://vm:8481/select/0/prometheus"},
	}
	for _, tt := range tests {
		if got := NormalizeMetricEndpoint(tt.in); got != tt.want {
			t.Errorf("NormalizeMetricEndpoint(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}
