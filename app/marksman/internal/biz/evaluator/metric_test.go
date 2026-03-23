package evaluator

import (
	"testing"
	"time"
)

func TestCalculateMetricQueryStep_UsesDefaultForSmallWindow(t *testing.T) {
	step := calculateMetricQueryStep(2 * time.Minute)
	if step != 15*time.Second {
		t.Fatalf("unexpected step for small window: got %s, want %s", step, 15*time.Second)
	}
}

func TestCalculateMetricQueryStep_IncreasesForLargeWindow(t *testing.T) {
	window := 7 * 24 * time.Hour
	step := calculateMetricQueryStep(window)
	if step <= 15*time.Second {
		t.Fatalf("step should be increased for large window, got %s", step)
	}

	points := int(window/step) + 1
	if points > maxQueryRangePoints {
		t.Fatalf("too many points: got %d, limit %d", points, maxQueryRangePoints)
	}
}
