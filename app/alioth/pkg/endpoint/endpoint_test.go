package endpoint

import "testing"

func TestNormalize(t *testing.T) {
	cases := []struct {
		raw  string
		want string
	}{
		{raw: " http://prometheus.moon.com/ ", want: "http://prometheus.moon.com"},
		{raw: "HTTP://Prometheus.Moon.com:9090/", want: "http://prometheus.moon.com:9090"},
		{raw: "https://metrics.example.com/prometheus/", want: "https://metrics.example.com/prometheus"},
		{raw: "http://victoriametrics.moon.com/select/0/prometheus", want: "http://victoriametrics.moon.com/select/0/prometheus"},
	}
	for _, item := range cases {
		if got := Normalize(item.raw); got != item.want {
			t.Fatalf("Normalize(%q) = %q, want %q", item.raw, got, item.want)
		}
	}
}
