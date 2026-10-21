package labels

import "testing"

func TestNormalizeMatchers(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		want    string
		wantErr bool
	}{
		{name: "empty", raw: "", want: "{}"},
		{name: "literal", raw: `{"env":"prod"}`, want: `{"env":"prod"}`},
		{name: "regex", raw: `{"instance":"10\\.0\\..*"}`, want: `{"instance":"10\\.0\\..*"}`},
		{name: "invalid json", raw: `not-json`, wantErr: true},
		{name: "invalid regex", raw: `{"env":"prod["}`, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NormalizeMatchers(tt.raw)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("NormalizeMatchers(%q) error = nil", tt.raw)
				}
				return
			}
			if err != nil {
				t.Fatalf("NormalizeMatchers(%q): %v", tt.raw, err)
			}
			if got != tt.want {
				t.Fatalf("NormalizeMatchers(%q) = %q, want %q", tt.raw, got, tt.want)
			}
		})
	}
}

func TestMatchersContained(t *testing.T) {
	tests := []struct {
		name     string
		matchers map[string]string
		target   map[string]string
		want     bool
	}{
		{name: "empty", matchers: map[string]string{}, target: map[string]string{"env": "prod"}, want: true},
		{name: "literal", matchers: map[string]string{"env": "prod"}, target: map[string]string{"env": "prod", "team": "ops"}, want: true},
		{name: "literal is fully anchored", matchers: map[string]string{"env": "prod"}, target: map[string]string{"env": "production"}, want: false},
		{name: "wildcard", matchers: map[string]string{"instance": `10\.0\..*`}, target: map[string]string{"instance": "10.0.0.1:9100"}, want: true},
		{name: "alternation", matchers: map[string]string{"env": "prod|staging"}, target: map[string]string{"env": "staging"}, want: true},
		{name: "different value", matchers: map[string]string{"env": "prod"}, target: map[string]string{"env": "dev"}, want: false},
		{name: "missing key", matchers: map[string]string{"env": ".*"}, target: map[string]string{"team": "ops"}, want: false},
		{name: "empty value requires the key", matchers: map[string]string{"env": ""}, target: map[string]string{"team": "ops"}, want: false},
		{name: "invalid pattern falls back to literal", matchers: map[string]string{"env": "prod["}, target: map[string]string{"env": "prod["}, want: true},
		{name: "invalid pattern does not match other values", matchers: map[string]string{"env": "prod["}, target: map[string]string{"env": "prod"}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MatchersContained(tt.matchers, tt.target); got != tt.want {
				t.Fatalf("MatchersContained(%v, %v) = %v, want %v", tt.matchers, tt.target, got, tt.want)
			}
		})
	}
}

func TestMatchersJSONContained(t *testing.T) {
	matched, err := MatchersJSONContained(`{"env":"pro.*"}`, `{"env":"prod"}`)
	if err != nil || !matched {
		t.Fatalf("regex match = %v, %v", matched, err)
	}
	if _, err := MatchersJSONContained(`{"env":"prod"}`, `not-json`); err == nil {
		t.Fatal("invalid target labels should return an error")
	}
	if matched, err := MatchersJSONContained("", `not-json`); err != nil || !matched {
		t.Fatalf("empty matchers should match regardless of target: %v, %v", matched, err)
	}
}
