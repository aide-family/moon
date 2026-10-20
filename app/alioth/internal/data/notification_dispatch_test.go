package data

import "testing"

func TestLabelsMatch(t *testing.T) {
	tests := []struct {
		name     string
		selector string
		labels   string
		want     bool
	}{
		{name: "empty selector", selector: "{}", labels: `{"env":"prod"}`, want: true},
		{name: "exact match", selector: `{"env":"prod"}`, labels: `{"env":"prod","team":"ops"}`, want: true},
		{name: "missing key", selector: `{"env":"prod"}`, labels: `{"team":"ops"}`, want: false},
		{name: "different value", selector: `{"env":"prod"}`, labels: `{"env":"dev"}`, want: false},
		{name: "invalid event labels", selector: `{"env":"prod"}`, labels: `not-json`, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := labelsMatch(tt.selector, tt.labels); got != tt.want {
				t.Fatalf("labelsMatch(%q, %q) = %v, want %v", tt.selector, tt.labels, got, tt.want)
			}
		})
	}
}

func TestContainsString(t *testing.T) {
	if !containsString([]string{"a", "b"}, "b") {
		t.Fatal("expected b")
	}
	if containsString([]string{"a"}, "b") {
		t.Fatal("unexpected b")
	}
}
