package alertevent

import (
	"testing"
	"time"
)

func TestInsightRange(t *testing.T) {
	now := time.Date(2026, 10, 25, 14, 30, 0, 0, time.FixedZone("CST", 8*3600))

	t.Run("default days uses UTC when time zone is empty", func(t *testing.T) {
		from, to, loc, err := insightRange(0, "", now)
		if err != nil {
			t.Fatalf("insightRange: %v", err)
		}
		if loc != time.UTC {
			t.Fatalf("loc = %v, want UTC", loc)
		}
		want := time.Date(2026, 10, 19, 0, 0, 0, 0, time.UTC)
		if !from.Equal(want) {
			t.Fatalf("from = %v, want %v", from, want)
		}
		if !to.Equal(now.UTC()) {
			t.Fatalf("to = %v, want %v", to, now.UTC())
		}
	})

	t.Run("local midnight for the requested zone", func(t *testing.T) {
		from, _, loc, err := insightRange(3, "Asia/Shanghai", now)
		if err != nil {
			t.Fatalf("insightRange: %v", err)
		}
		want := time.Date(2026, 10, 23, 0, 0, 0, 0, loc)
		if !from.Equal(want) {
			t.Fatalf("from = %v, want %v", from, want)
		}
		if hour, minute := from.In(loc).Hour(), from.In(loc).Minute(); hour != 0 || minute != 0 {
			t.Fatalf("from is not local midnight: %v", from.In(loc))
		}
	})

	t.Run("rejects out of range days", func(t *testing.T) {
		if _, _, _, err := insightRange(31, "", now); err == nil {
			t.Fatal("days=31 should be rejected")
		}
		if _, _, _, err := insightRange(-1, "", now); err == nil {
			t.Fatal("days=-1 should be rejected")
		}
	})

	t.Run("rejects unknown time zone", func(t *testing.T) {
		if _, _, _, err := insightRange(7, "Mars/Olympus", now); err == nil {
			t.Fatal("unknown time zone should be rejected")
		}
	})
}

// 夏令时切换日不能靠 24 小时累加：2026-03-08 纽约只有 23 小时。
// 区间用本地零点回退，保证 N 天就是 N 个自然日。
func TestInsightRangeAcrossDST(t *testing.T) {
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatalf("load location: %v", err)
	}
	now := time.Date(2026, 3, 10, 9, 0, 0, 0, loc)
	from, to, _, err := insightRange(3, "America/New_York", now)
	if err != nil {
		t.Fatalf("insightRange: %v", err)
	}
	want := time.Date(2026, 3, 8, 0, 0, 0, 0, loc)
	if !from.Equal(want) {
		t.Fatalf("from = %v, want %v", from, want)
	}
	// 3 月 8 日当天 23 小时，整段区间不是整 72 小时。
	if hours := to.Sub(from).Hours(); hours >= 72 {
		t.Fatalf("hours = %v, want less than 72 across the spring-forward day", hours)
	}
	// 逐日推进仍然是三个自然日。
	days := 0
	for cursor := from; cursor.Before(to); cursor = cursor.AddDate(0, 0, 1) {
		days++
	}
	if days != 3 {
		t.Fatalf("days = %d, want 3", days)
	}
}
