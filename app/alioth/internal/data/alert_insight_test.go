package data

import (
	"testing"
	"time"
)

func shanghai(t *testing.T) *time.Location {
	t.Helper()
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		t.Fatalf("load location: %v", err)
	}
	return loc
}

func TestBuildInsightSeparatesNewFromClosed(t *testing.T) {
	loc := shanghai(t)
	from := time.Date(2026, 10, 23, 0, 0, 0, 0, loc)
	to := time.Date(2026, 10, 25, 14, 0, 0, 0, loc)

	// 23 日新增 6 条：其中 4 条 24 日闭环、2 条 25 日人工闭环。
	startedHot := []insightRow{
		{day: "2026-10-24", status: "FIRING", levelUID: "lvl-p1", events: 3},
		{day: "2026-10-24", status: "FIRING", levelUID: "", events: 1},
	}
	startedArchived := []insightRow{
		{day: "2026-10-23", status: "RECOVERED", levelUID: "lvl-p1", events: 4},
		{day: "2026-10-23", status: "RECOVERED_BY_MANUAL", levelUID: "lvl-p2", events: 2},
	}
	closedArchived := []insightRow{
		{day: "2026-10-24", status: "RECOVERED", levelUID: "lvl-p1", events: 4, samples: 4, p50: 120, p90: 900},
		{day: "2026-10-25", status: "RECOVERED_BY_MANUAL", levelUID: "lvl-p2", events: 2, samples: 2, p50: 60, p90: 300},
		{day: "2026-10-24", status: "", levelUID: "", events: 4, samples: 4, p50: 120, p90: 900},
		{day: "2026-10-25", status: "", levelUID: "", events: 2, samples: 2, p50: 60, p90: 300},
		{day: "", status: "", levelUID: "", events: 6, samples: 6, p50: 110, p90: 850},
	}

	got := buildInsight(startedHot, startedArchived, nil, closedArchived, from, to, loc, true)

	// to 是 25 日 14:00，今天的桶从 00:00 开始算，所以是 23、24、25 三天。
	if len(got.Days) != 3 {
		t.Fatalf("days = %d, want 3", len(got.Days))
	}
	// 23 日只有新增，没有闭环：闭环落在事件结束的那一天，不是开始的那一天。
	if got.Days[0].Events != 6 || got.Days[0].Closed != 0 {
		t.Fatalf("day 0 = %+v, want 6 new and 0 closed", got.Days[0])
	}
	// 24 日新增 4、闭环 4：同一天两根柱，但不是同一批事件。
	if got.Days[1].Events != 4 || got.Days[1].Closed != 4 || got.Days[1].P90Seconds != 900 {
		t.Fatalf("day 1 = %+v", got.Days[1])
	}
	if got.Days[2].Events != 0 || got.Days[2].Closed != 2 || got.Days[2].ClosedByManual != 2 {
		t.Fatalf("day 2 = %+v, want 0 new and 2 manual closes", got.Days[2])
	}
	if got.EventsTotal != 10 || got.ClosedTotal != 6 {
		t.Fatalf("totals = %d/%d, want 10 new and 6 closed", got.EventsTotal, got.ClosedTotal)
	}
	// 区间分位数取归档的汇总行；样本数为 0 时前端必须显示「—」。
	if got.P50Seconds != 110 || got.P90Seconds != 850 || got.RecoverySamples != 6 {
		t.Fatalf("range quantiles = %+v", got)
	}
	if !got.Complete {
		t.Fatal("complete = false, want true")
	}

	// 级别按事件数降序：lvl-p1 = 3(实时) + 4(归档) = 7、lvl-p2 = 2、未分级 = 1。
	if len(got.Levels) != 3 {
		t.Fatalf("levels = %+v, want 3 rows", got.Levels)
	}
	if got.Levels[0].LevelUID != "lvl-p1" || got.Levels[0].Events != 7 || got.Levels[0].Firing != 3 {
		t.Fatalf("first level = %+v", got.Levels[0])
	}
	if got.Levels[2].LevelUID != "" || got.Levels[2].Firing != 1 {
		t.Fatalf("unranked level = %+v", got.Levels[2])
	}
}

func TestBuildInsightFallsBackToHotQuantiles(t *testing.T) {
	loc := shanghai(t)
	from := time.Date(2026, 10, 24, 0, 0, 0, 0, loc)
	to := time.Date(2026, 10, 25, 0, 0, 0, 0, loc)

	startedHot := []insightRow{
		{day: "2026-10-24", status: "RECOVERED", levelUID: "lvl-p1", events: 2},
	}
	closedHot := []insightRow{
		{day: "2026-10-24", status: "RECOVERED", levelUID: "lvl-p1", events: 2, samples: 2, p50: 30, p90: 90},
		{day: "2026-10-24", status: "", levelUID: "", events: 2, samples: 2, p50: 30, p90: 90},
		{day: "", status: "", levelUID: "", events: 2, samples: 2, p50: 30, p90: 90},
	}

	got := buildInsight(startedHot, nil, closedHot, nil, from, to, loc, false)

	if got.Complete {
		t.Fatal("complete = true, want false without history storage")
	}
	if got.P50Seconds != 30 || got.P90Seconds != 90 || got.RecoverySamples != 2 {
		t.Fatalf("quantiles = %+v, want hot fallback", got)
	}
	if len(got.Days) != 1 || got.Days[0].P90Seconds != 90 || got.Days[0].Closed != 2 {
		t.Fatalf("day = %+v", got.Days)
	}
}

func TestHistoryTableOverlaps(t *testing.T) {
	cases := []struct {
		name  string
		table string
		from  time.Time
		to    time.Time
		want  bool
	}{
		{
			name: "same week", table: "alert_events_20261019",
			from: time.Date(2026, 10, 20, 0, 0, 0, 0, time.UTC),
			to:   time.Date(2026, 10, 21, 0, 0, 0, 0, time.UTC),
			want: true,
		},
		{
			name: "range starts before the week ends", table: "alert_events_20261019",
			from: time.Date(2026, 10, 25, 0, 0, 0, 0, time.UTC),
			to:   time.Date(2026, 10, 26, 0, 0, 0, 0, time.UTC),
			want: true,
		},
		{
			name: "week fully in the past", table: "alert_events_20261012",
			from: time.Date(2026, 10, 25, 0, 0, 0, 0, time.UTC),
			to:   time.Date(2026, 10, 26, 0, 0, 0, 0, time.UTC),
			want: false,
		},
		{
			name: "week fully in the future", table: "alert_events_20261102",
			from: time.Date(2026, 10, 25, 0, 0, 0, 0, time.UTC),
			to:   time.Date(2026, 10, 26, 0, 0, 0, 0, time.UTC),
			want: false,
		},
		{
			name: "non weekly table", table: "alert_events_summary",
			from: time.Date(2026, 10, 25, 0, 0, 0, 0, time.UTC),
			to:   time.Date(2026, 10, 26, 0, 0, 0, 0, time.UTC),
			want: false,
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			got := historyTableOverlaps(testCase.table, "alert_events_", testCase.from, testCase.to)
			if got != testCase.want {
				t.Fatalf("overlaps = %v, want %v", got, testCase.want)
			}
		})
	}
}
