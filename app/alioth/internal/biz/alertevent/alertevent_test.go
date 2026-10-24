package alertevent

import (
	"testing"
	"time"

	"github.com/aide-family/Alioth/app/alioth/pkg/bo"
)

func TestValidateHistoryListQuery(t *testing.T) {
	from := time.Date(2026, 10, 24, 12, 0, 0, 0, time.UTC)
	to := from.Add(time.Hour)
	// 前端一天是 [00:00:00, 23:59:59.999999999]，30 天整正好不超过上限，
	// 第 31 天就必须拒绝。
	thirtyDaysFrom := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	thirtyDaysTo := time.Date(2026, 10, 30, 23, 59, 59, 999999999, time.UTC)
	thirtyOneDaysTo := thirtyDaysTo.Add(24 * time.Hour)
	cases := []struct {
		name    string
		query   ListQuery
		wantErr bool
	}{
		{name: "history filters allowed", query: ListQuery{View: "HISTORY", StartsAtFrom: &from, StartsAtTo: &to, EventStatus: "RECOVERED", Suppression: "SUPPRESSED", RuleName: "API", DatasourceName: "prod"}},
		{name: "history-only filter rejected in realtime", query: ListQuery{View: "REALTIME", RuleName: "API"}, wantErr: true},
		{name: "invalid time range", query: ListQuery{View: "HISTORY", StartsAtFrom: &to, StartsAtTo: &from}, wantErr: true},
		{name: "30 days allowed", query: ListQuery{View: "HISTORY", StartsAtFrom: &thirtyDaysFrom, StartsAtTo: &thirtyDaysTo}},
		{name: "31 days rejected", query: ListQuery{View: "HISTORY", StartsAtFrom: &thirtyDaysFrom, StartsAtTo: &thirtyOneDaysTo}, wantErr: true},
		{name: "unbounded allowed", query: ListQuery{View: "HISTORY", StartsAtFrom: &from}},
		{name: "invalid status", query: ListQuery{View: "HISTORY", EventStatus: "UNKNOWN"}, wantErr: true},
		{name: "invalid suppression filter", query: ListQuery{View: "HISTORY", Suppression: "MAYBE"}, wantErr: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateHistoryListQuery(tc.query)
			if (err != nil) != tc.wantErr {
				t.Fatalf("validateHistoryListQuery() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}

func TestChartEnd(t *testing.T) {
	now := time.Date(2026, 10, 24, 12, 0, 0, 0, time.UTC)
	startsAt := now.Add(-48 * time.Hour)
	recoveredAt := now.Add(-24 * time.Hour)
	future := now.Add(time.Hour)

	cases := []struct {
		name  string
		event bo.AlertEvent
		want  time.Time
	}{
		{
			name:  "recovered event ends at recovery time",
			event: bo.AlertEvent{StartsAt: startsAt, EndsAt: &recoveredAt},
			want:  recoveredAt,
		},
		{
			name:  "firing event ends at now",
			event: bo.AlertEvent{StartsAt: startsAt},
			want:  now,
		},
		{
			name:  "zero ends_at ends at now",
			event: bo.AlertEvent{StartsAt: startsAt, EndsAt: &time.Time{}},
			want:  now,
		},
		{
			name:  "future ends_at falls back to now",
			event: bo.AlertEvent{StartsAt: startsAt, EndsAt: &future},
			want:  now,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := chartEnd(tc.event, now); !got.Equal(tc.want) {
				t.Fatalf("chartEnd() = %s, want %s", got, tc.want)
			}
		})
	}
}

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
