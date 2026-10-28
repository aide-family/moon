package alertevent

import (
	"testing"
	"time"

	"github.com/aide-family/moon/pkg/merr"
)

func analysisTestLocation(t *testing.T, name string) *time.Location {
	t.Helper()
	loc, err := time.LoadLocation(name)
	if err != nil {
		t.Fatalf("load location %s: %v", name, err)
	}
	return loc
}

func TestAnalysisWindow(t *testing.T) {
	loc := analysisTestLocation(t, "Asia/Shanghai")
	now := time.Date(2026, 10, 27, 14, 30, 0, 0, loc)

	t.Run("window covers both days inclusively", func(t *testing.T) {
		from, to, err := analysisWindow("2026-10-01", "2026-10-27", loc, now)
		if err != nil {
			t.Fatalf("analysisWindow: %v", err)
		}
		wantFrom := time.Date(2026, 10, 1, 0, 0, 0, 0, loc)
		if !from.Equal(wantFrom) {
			t.Fatalf("from = %v, want %v", from, wantFrom)
		}
		wantTo := time.Date(2026, 10, 28, 0, 0, 0, 0, loc)
		if !to.Equal(wantTo) {
			t.Fatalf("to = %v, want %v", to, wantTo)
		}
	})

	t.Run("thirty days inclusive is allowed", func(t *testing.T) {
		if _, _, err := analysisWindow("2026-09-28", "2026-10-27", loc, now); err != nil {
			t.Fatalf("30-day range rejected: %v", err)
		}
	})

	t.Run("thirty one days is rejected", func(t *testing.T) {
		if _, _, err := analysisWindow("2026-09-27", "2026-10-27", loc, now); !merr.IsInvalid(err) {
			t.Fatalf("31-day range err = %v, want invalid", err)
		}
	})

	t.Run("to later than today in the requested time zone is rejected", func(t *testing.T) {
		if _, _, err := analysisWindow("2026-10-01", "2026-10-28", loc, now); !merr.IsInvalid(err) {
			t.Fatalf("future range err = %v, want invalid", err)
		}
	})

	t.Run("today is computed in the requested time zone", func(t *testing.T) {
		// UTC 还是 27 日晚上，上海已经是 28 日。
		utcNow := time.Date(2026, 10, 27, 23, 30, 0, 0, time.UTC)
		if _, _, err := analysisWindow("2026-10-01", "2026-10-28", loc, utcNow); err != nil {
			t.Fatalf("local today should be 2026-10-28: %v", err)
		}
		if _, _, err := analysisWindow("2026-10-01", "2026-10-29", loc, utcNow); !merr.IsInvalid(err) {
			t.Fatalf("err = %v, want invalid", err)
		}
	})

	t.Run("day count is calendar based across DST", func(t *testing.T) {
		newYork := analysisTestLocation(t, "America/New_York")
		dstNow := time.Date(2026, 3, 10, 9, 0, 0, 0, newYork)
		// 2026-02-09 到 2026-03-10 是 30 个自然日，尽管中间有春季夏令时切换。
		if _, _, err := analysisWindow("2026-02-09", "2026-03-10", newYork, dstNow); err != nil {
			t.Fatalf("DST range rejected: %v", err)
		}
		// 2026-02-08 到 2026-03-10 是 31 个自然日，小时数不到 744 也必须拒绝。
		if _, _, err := analysisWindow("2026-02-08", "2026-03-10", newYork, dstNow); !merr.IsInvalid(err) {
			t.Fatalf("DST 31-day range err = %v, want invalid", err)
		}
	})

	t.Run("rejects bad input", func(t *testing.T) {
		cases := []struct{ from, to string }{
			{"", "2026-10-27"},
			{"2026-10-01", ""},
			{"2026/10/01", "2026-10-27"},
			{"2026-10-27", "2026-10-01"},
		}
		for _, item := range cases {
			if _, _, err := analysisWindow(item.from, item.to, loc, now); !merr.IsInvalid(err) {
				t.Fatalf("analysisWindow(%q, %q) err = %v, want invalid", item.from, item.to, err)
			}
		}
	})
}

func TestAnalysisFocus(t *testing.T) {
	loc := analysisTestLocation(t, "Asia/Shanghai")
	from := time.Date(2026, 10, 1, 0, 0, 0, 0, loc)
	to := time.Date(2026, 10, 28, 0, 0, 0, 0, loc)

	t.Run("no focus keeps the whole range", func(t *testing.T) {
		gotFrom, gotTo, err := analysisFocus(from, to, "", "")
		if err != nil {
			t.Fatalf("analysisFocus: %v", err)
		}
		if !gotFrom.Equal(from) || !gotTo.Equal(to) {
			t.Fatalf("range = %v..%v, want %v..%v", gotFrom, gotTo, from, to)
		}
	})

	t.Run("valid focus replaces the range", func(t *testing.T) {
		gotFrom, gotTo, err := analysisFocus(from, to, "2026-10-05T00:00:00+08:00", "2026-10-06T00:00:00+08:00")
		if err != nil {
			t.Fatalf("analysisFocus: %v", err)
		}
		if gotFrom.Format(time.RFC3339) != "2026-10-05T00:00:00+08:00" {
			t.Fatalf("from = %v", gotFrom)
		}
		if gotTo.Format(time.RFC3339) != "2026-10-06T00:00:00+08:00" {
			t.Fatalf("to = %v", gotTo)
		}
	})

	t.Run("focus may touch both window edges", func(t *testing.T) {
		if _, _, err := analysisFocus(from, to,
			from.Format(time.RFC3339), to.Format(time.RFC3339)); err != nil {
			t.Fatalf("edge focus rejected: %v", err)
		}
	})

	t.Run("rejects invalid focus", func(t *testing.T) {
		cases := []struct{ start, end string }{
			{"2026-10-05T00:00:00+08:00", ""},
			{"", "2026-10-06T00:00:00+08:00"},
			{"2026-10-06T00:00:00+08:00", "2026-10-06T00:00:00+08:00"},
			{"2026-10-06T00:00:00+08:00", "2026-10-05T00:00:00+08:00"},
			{"2026-09-30T00:00:00+08:00", "2026-10-06T00:00:00+08:00"},
			{"2026-10-05T00:00:00+08:00", "2026-10-28T01:00:00+08:00"},
			{"not-a-time", "2026-10-06T00:00:00+08:00"},
		}
		for _, item := range cases {
			if _, _, err := analysisFocus(from, to, item.start, item.end); !merr.IsInvalid(err) {
				t.Fatalf("analysisFocus(%q, %q) err = %v, want invalid", item.start, item.end, err)
			}
		}
	})
}

func TestAnalysisStatus(t *testing.T) {
	for _, value := range []string{"", "RECOVERED", "RECOVERED_BY_MANUAL", "SUPPRESSING"} {
		if _, err := analysisStatus(value); err != nil {
			t.Fatalf("analysisStatus(%q) = %v, want nil", value, err)
		}
	}
	for _, value := range []string{"FIRING", "suppressing", "SUPPRESSED"} {
		if _, err := analysisStatus(value); !merr.IsInvalid(err) {
			t.Fatalf("analysisStatus(%q) err = %v, want invalid", value, err)
		}
	}
}

func TestAnalysisScopeConvertsSuppressingAndFocus(t *testing.T) {
	loc := analysisTestLocation(t, "Asia/Shanghai")
	now := time.Date(2026, 10, 27, 14, 30, 0, 0, loc)

	scope, err := analysisScope(AnalysisQuery{
		From: "2026-10-01", To: "2026-10-27", TimeZone: "Asia/Shanghai",
		Status: "SUPPRESSING", FocusStart: "2026-10-05T00:00:00+08:00", FocusEnd: "2026-10-06T00:00:00+08:00",
	}, now)
	if err != nil {
		t.Fatalf("analysisScope: %v", err)
	}
	// SUPPRESSING 在库里对应仍为 FIRING 且仍在抑制中的事件。
	if scope.Status != "FIRING" {
		t.Fatalf("status = %q, want FIRING", scope.Status)
	}
	if scope.Location == nil || scope.Location.String() != "Asia/Shanghai" {
		t.Fatalf("location = %v", scope.Location)
	}
	if scope.From.Format(time.RFC3339) != "2026-10-05T00:00:00+08:00" {
		t.Fatalf("from = %v", scope.From)
	}
	if scope.To.Format(time.RFC3339) != "2026-10-06T00:00:00+08:00" {
		t.Fatalf("to = %v", scope.To)
	}
}

func TestAnalysisScopeDefaultsToUTC(t *testing.T) {
	scope, err := analysisScope(AnalysisQuery{From: "2026-10-01", To: "2026-10-02"}, time.Now())
	if err != nil {
		t.Fatalf("analysisScope: %v", err)
	}
	if scope.Location != time.UTC {
		t.Fatalf("location = %v, want UTC", scope.Location)
	}
}

func TestAnalysisScopeRejectsUnknownTimeZone(t *testing.T) {
	_, err := analysisScope(AnalysisQuery{From: "2026-10-01", To: "2026-10-02", TimeZone: "Mars/Olympus"}, time.Now())
	if !merr.IsInvalid(err) {
		t.Fatalf("err = %v, want invalid", err)
	}
}

func TestAnalysisGranularity(t *testing.T) {
	valid := map[string]string{"": "day", "hour": "hour", "day": "day", "week": "week", "month": "month"}
	for value, want := range valid {
		got, err := analysisGranularity(value)
		if err != nil || got != want {
			t.Fatalf("analysisGranularity(%q) = %q, %v; want %q", value, got, err, want)
		}
	}
	for _, value := range []string{"minute", "DAY", "years"} {
		if _, err := analysisGranularity(value); !merr.IsInvalid(err) {
			t.Fatalf("analysisGranularity(%q) err = %v, want invalid", value, err)
		}
	}
}

func TestAnalysisFingerprintLimit(t *testing.T) {
	if limit, err := analysisFingerprintLimit(0); err != nil || limit != int(AnalysisDefaultFingerprintLimit) {
		t.Fatalf("limit(0) = %d, %v", limit, err)
	}
	for _, value := range []int32{1, 5, 20} {
		limit, err := analysisFingerprintLimit(value)
		if err != nil || limit != int(value) {
			t.Fatalf("limit(%d) = %d, %v", value, limit, err)
		}
	}
	for _, value := range []int32{-1, 21, 100} {
		if _, err := analysisFingerprintLimit(value); !merr.IsInvalid(err) {
			t.Fatalf("limit(%d) err = %v, want invalid", value, err)
		}
	}
}

func TestAnalysisTrendIgnoresFocus(t *testing.T) {
	loc := analysisTestLocation(t, "Asia/Shanghai")
	// AnalysisTrend 会把 focus 清空后再校验：只传半个聚焦窗口也不影响趋势请求。
	query := AnalysisQuery{
		From: "2026-10-01", To: "2026-10-27", TimeZone: "Asia/Shanghai",
		FocusStart: "2026-10-05T00:00:00+08:00",
	}
	query.FocusStart, query.FocusEnd = "", ""
	scope, err := analysisScope(query, time.Date(2026, 10, 27, 14, 30, 0, 0, loc))
	if err != nil {
		t.Fatalf("analysisScope: %v", err)
	}
	wantFrom := time.Date(2026, 10, 1, 0, 0, 0, 0, loc)
	if !scope.From.Equal(wantFrom) {
		t.Fatalf("from = %v, want %v", scope.From, wantFrom)
	}
}
