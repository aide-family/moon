package alertevent

import (
	"testing"
	"time"
)

func TestChartWindow(t *testing.T) {
	now := time.Date(2026, 10, 21, 12, 0, 0, 0, time.UTC)
	start, end, step := chartWindow(now.Add(-time.Minute), now)
	if start != now.Add(-time.Minute).Unix() || end != now.Unix() || step != 15 {
		t.Fatalf("short window = %d %d %d", start, end, step)
	}

	start, end, step = chartWindow(now.Add(-24*time.Hour), now)
	if end-start > 300*step {
		t.Fatalf("too many points: span=%d step=%d", end-start, step)
	}
	if step%15 != 0 {
		t.Fatalf("step = %d, want multiple of 15", step)
	}
	if start >= end {
		t.Fatalf("window = %d..%d", start, end)
	}

	start, end, step = chartWindow(now.Add(time.Hour), now)
	if end != start+1 || step != 15 {
		t.Fatalf("future start = %d %d %d", start, end, step)
	}
}
