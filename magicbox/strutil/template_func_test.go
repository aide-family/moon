package strutil

import "testing"

func TestTemplateDefault(t *testing.T) {
	if got := templateDefault("N/A", ""); got != "N/A" {
		t.Fatalf("templateDefault with empty string = %v, want N/A", got)
	}
	if got := templateDefault("N/A", "HighCPU"); got != "HighCPU" {
		t.Fatalf("templateDefault with value = %v, want HighCPU", got)
	}
	if got := templateDefault("N/A", "", "fallback"); got != "fallback" {
		t.Fatalf("templateDefault with multiple overrides = %v, want fallback", got)
	}
}
