package alertevent

import (
	"testing"
	"time"
)

func TestValidateHistoryListQuery(t *testing.T) {
	from := time.Date(2026, 10, 24, 12, 0, 0, 0, time.UTC)
	to := from.Add(time.Hour)
	cases := []struct {
		name    string
		query   ListQuery
		wantErr bool
	}{
		{name: "history filters allowed", query: ListQuery{View: "HISTORY", StartsAtFrom: &from, StartsAtTo: &to, EventStatus: "RECOVERED", Suppression: "SUPPRESSED", RuleName: "API", DatasourceName: "prod"}},
		{name: "history-only filter rejected in realtime", query: ListQuery{View: "REALTIME", RuleName: "API"}, wantErr: true},
		{name: "invalid time range", query: ListQuery{View: "HISTORY", StartsAtFrom: &to, StartsAtTo: &from}, wantErr: true},
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
