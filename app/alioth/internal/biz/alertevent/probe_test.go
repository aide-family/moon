package alertevent

import (
	"testing"
	"time"

	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/pkg/alertjudge"
)

// TestProbeLevelWindow 固定判定窗口：持续类要覆盖 for_seconds 加一个周期，
// 其余只看最近一个周期。窗口取大了会让刚恢复的目标继续命中。
func TestProbeLevelWindow(t *testing.T) {
	interval := 15 * time.Second
	cases := []struct {
		level bo.AlertRuleLevel
		want  time.Duration
	}{
		{bo.AlertRuleLevel{Sample: "FOR", ForSeconds: 0}, interval},
		{bo.AlertRuleLevel{Sample: "FOR", ForSeconds: 30}, 45 * time.Second},
		{bo.AlertRuleLevel{Sample: "MAX", ForSeconds: 300}, interval},
		{bo.AlertRuleLevel{Sample: "MIN"}, interval},
	}
	for _, item := range cases {
		if got := probeLevelWindow(item.level, interval); got != item.want {
			t.Fatalf("probeLevelWindow(%+v) = %v, want %v", item.level, got, item.want)
		}
	}
}

// TestFilterProbePoints 只保留窗口内的点，避免旧的成功点把持续条件一直否掉。
func TestFilterProbePoints(t *testing.T) {
	slot := time.Date(2026, 10, 29, 5, 39, 0, 0, time.UTC)
	points := []alertjudge.Point{
		{Value: 1, At: slot.Add(-30 * time.Second)},
		{Value: 0, At: slot.Add(-15 * time.Second)},
		{Value: 0, At: slot},
	}
	windowed := filterProbePoints(points, slot.Add(-15*time.Second))
	if len(windowed) != 2 {
		t.Fatalf("expected 2 points in window, got %d", len(windowed))
	}
	level := alertjudge.Level{Condition: "EQ", Sample: "FOR", Threshold: "0"}
	if _, _, ok := alertjudge.Match(level, windowed, 15); !ok {
		t.Fatal("two consecutive failures should hit")
	}
	if _, _, ok := alertjudge.Match(level, points, 15); ok {
		t.Fatal("window containing an old success should not hit")
	}
}
