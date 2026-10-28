package data

import (
	"strings"
	"testing"
)

func TestAvailabilityFingerprintIsStableAndPerDatasource(t *testing.T) {
	first := availabilityFingerprint("ds-1")
	if first != availabilityFingerprint("ds-1") {
		t.Fatal("fingerprint changed between calls")
	}
	if first == availabilityFingerprint("ds-2") {
		t.Fatal("different datasources share an availability fingerprint")
	}
	if len(first) == 0 {
		t.Fatal("empty availability fingerprint")
	}
}

func TestAvailabilitySummaryIncludesWindowAndCount(t *testing.T) {
	summary := availabilitySummary("prom-prod", 7, 5, 3)
	for _, want := range []string{"prom-prod", "5", "7", "3"} {
		if !strings.Contains(summary, want) {
			t.Fatalf("summary %q misses %q", summary, want)
		}
	}
}

func TestAvailabilityDescriptionKeepsLastFailure(t *testing.T) {
	description := availabilityDescription("upstream query failed")
	if !strings.Contains(description, "upstream query failed") {
		t.Fatalf("description = %q", description)
	}
	if got := availabilityDescription("   "); got == "" {
		t.Fatal("empty failure should still produce a description")
	}
	long := strings.Repeat("失", maxFailureMessageLength+10)
	if got := availabilityDescription(long); len([]rune(got)) > maxFailureMessageLength+32 {
		t.Fatalf("description not truncated: %d runes", len([]rune(got)))
	}
}

func TestTruncateTextKeepsRunes(t *testing.T) {
	if got := truncateText("  数据源不可用  ", 32); got != "数据源不可用" {
		t.Fatalf("trimmed = %q", got)
	}
	if got := truncateText("数据源不可用", 3); got != "数据源…" {
		t.Fatalf("truncated = %q", got)
	}
	if got := truncateText("abc", 0); got != "…" {
		t.Fatalf("zero limit = %q", got)
	}
}
