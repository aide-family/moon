package data

import (
	"context"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/aide-family/Alioth/app/alioth/internal/conf"
	"github.com/aide-family/Alioth/app/alioth/pkg/bo"
)

func TestAnalysisHelpers(t *testing.T) {
	t.Run("merge counts adds duplicate groups across sources", func(t *testing.T) {
		merged := mergeAnalysisCounts(
			map[string]int64{"FIRING": 2, "RECOVERED": 1},
			map[string]int64{"FIRING": 3, "RECOVERED_BY_MANUAL": 4},
		)
		if merged["FIRING"] != 5 || merged["RECOVERED"] != 1 || merged["RECOVERED_BY_MANUAL"] != 4 {
			t.Fatalf("merged = %+v", merged)
		}
	})

	t.Run("fingerprint totals ignore the empty fingerprint", func(t *testing.T) {
		events, unique, repeated := analysisFingerprintTotals(map[string]int64{
			"": 3, "fp-a": 4, "fp-b": 1,
		})
		if events != 8 {
			t.Fatalf("events = %d, want 8", events)
		}
		// 空指纹不计入去重数；重复事件数 = 每个指纹计数减 1 之和。
		if unique != 2 || repeated != 3 {
			t.Fatalf("unique/repeated = %d/%d, want 2/3", unique, repeated)
		}
	})

	t.Run("sorted counts order by events then key", func(t *testing.T) {
		items := sortedAnalysisKeyCounts(map[string]int64{"b": 2, "a": 2, "c": 5})
		if len(items) != 3 || items[0].key != "c" || items[1].key != "a" || items[2].key != "b" {
			t.Fatalf("items = %+v", items)
		}
	})

	t.Run("top fingerprints drop singletons and cut by limit", func(t *testing.T) {
		groups := map[string]analysisFingerprintGroup{
			"fp-one":   {events: 1, ruleUID: "rule-1"},
			"fp-two":   {events: 2, ruleUID: "rule-1"},
			"fp-four":  {events: 4, ruleUID: "rule-2"},
			"fp-three": {events: 3, ruleUID: "rule-2"},
		}
		items := topAnalysisFingerprints(groups, 2)
		if len(items) != 2 || items[0].Fingerprint != "fp-four" || items[1].Fingerprint != "fp-three" {
			t.Fatalf("items = %+v", items)
		}
		all := topAnalysisFingerprints(groups, 10)
		if len(all) != 3 {
			t.Fatalf("singleton fingerprint must be dropped: %+v", all)
		}
		if all[2].Fingerprint != "fp-two" {
			t.Fatalf("tie breaks by fingerprint: %+v", all)
		}
	})

	t.Run("merge fingerprint groups keeps hot rule uid", func(t *testing.T) {
		merged := mergeAnalysisFingerprintGroups(
			map[string]analysisFingerprintGroup{"fp-a": {events: 1, ruleUID: "hot-rule"}},
			map[string]analysisFingerprintGroup{"fp-a": {events: 2, ruleUID: "archive-rule"}, "fp-b": {events: 1, ruleUID: "archive-rule"}},
		)
		if got := merged["fp-a"]; got.events != 3 || got.ruleUID != "hot-rule" {
			t.Fatalf("fp-a = %+v", got)
		}
		if got := merged["fp-b"]; got.ruleUID != "archive-rule" {
			t.Fatalf("fp-b = %+v", got)
		}
	})

	t.Run("recovery prefers archive and falls back to hot", func(t *testing.T) {
		archived := map[string]bo.AlertAnalysisRecoveryStats{
			"RECOVERED": {MedianSeconds: 10, P90Seconds: 30, Samples: 2},
		}
		hot := map[string]bo.AlertAnalysisRecoveryStats{
			"RECOVERED":           {MedianSeconds: 99, P90Seconds: 99, Samples: 9},
			"RECOVERED_BY_MANUAL": {MedianSeconds: 5, P90Seconds: 6, Samples: 1},
		}
		if got := analysisRecoveryOf(archived, hot, "RECOVERED"); got.MedianSeconds != 10 || got.Samples != 2 {
			t.Fatalf("auto recovery = %+v, want archive", got)
		}
		// 归档没有该类样本时退回主库，避免分位数恒为 0。
		if got := analysisRecoveryOf(archived, hot, "RECOVERED_BY_MANUAL"); got.MedianSeconds != 5 || got.Samples != 1 {
			t.Fatalf("manual recovery = %+v, want hot fallback", got)
		}
		// 两边都没有样本时返回零值。
		if got := analysisRecoveryOf(archived, hot, "MISSING"); got != (bo.AlertAnalysisRecoveryStats{}) {
			t.Fatalf("missing recovery = %+v, want zero", got)
		}
	})

	t.Run("granularity falls back to day", func(t *testing.T) {
		for _, value := range []string{"hour", "day", "week", "month"} {
			if got := analysisGranularity(value); got != value {
				t.Fatalf("analysisGranularity(%q) = %q", value, got)
			}
		}
		if got := analysisGranularity("minute"); got != "day" {
			t.Fatalf("analysisGranularity(minute) = %q, want day", got)
		}
	})

	t.Run("bucket counts merge by unix second", func(t *testing.T) {
		merged := mergeAnalysisBucketCounts(
			map[int64]int64{100: 2, 200: 1},
			map[int64]int64{100: 3},
		)
		if merged[100] != 5 || merged[200] != 1 {
			t.Fatalf("merged = %+v", merged)
		}
	})
}

// 指纹去重必须走 group by 的 hash 聚合，不能退化成 count(DISTINCT) 的外部排序；
// 恢复时长必须用最近秩口径的 percentile_disc(0.9)，与其他查询的分位数使用范围分开。
func TestAnalysisSQLShape(t *testing.T) {
	if sql := analysisFingerprintSQL("alert_events"); strings.Contains(strings.ToUpper(sql), "DISTINCT") {
		t.Fatalf("fingerprint summary must not use count(distinct): %s", sql)
	}
	if sql := analysisFingerprintGroupSQL("alert_events"); strings.Contains(strings.ToUpper(sql), "DISTINCT") {
		t.Fatalf("top fingerprints must not use count(distinct): %s", sql)
	}
	recovery := analysisRecoverySQL("alert_events")
	if !strings.Contains(recovery, "percentile_cont(0.5)") {
		t.Fatalf("recovery median must use percentile_cont(0.5): %s", recovery)
	}
	if !strings.Contains(recovery, "percentile_disc(0.9)") {
		t.Fatalf("recovery p90 must use percentile_disc(0.9): %s", recovery)
	}
	if !strings.Contains(recovery, "ends_at IS NOT NULL") {
		t.Fatalf("recovery must only read closed events: %s", recovery)
	}
	// 其他聚合不能带分位数。
	for name, sql := range map[string]string{
		"summary":      analysisFingerprintSQL("alert_events"),
		"fingerprints": analysisFingerprintGroupSQL("alert_events"),
	} {
		if strings.Contains(sql, "percentile_") {
			t.Fatalf("%s query must not use percentiles: %s", name, sql)
		}
	}
}

// 分析接口的 SQL 只在真库上才能验证参数绑定、跨周表 UNION ALL 与扫描类型。
// 这组测试只发 SELECT，不会写入任何数据。
//
//	ALIOTH_TEST_DATABASE_URL='postgres://alioth:alioth@127.0.0.1:5432/alioth?sslmode=disable' \
//	  go test ./internal/data -run TestAlertAnalysisAgainstRealDatabase -v
func TestAlertAnalysisAgainstRealDatabase(t *testing.T) {
	dsn := testDatabaseURL(t)
	deps, cleanup, err := NewDependencies(
		&conf.Data{Database: &conf.Data_Database{Source: dsn}},
		&conf.Auth{EncryptKey: "dev-only-encrypt-key"},
		slog.Default(),
	)
	if err != nil {
		t.Fatalf("open dependencies: %v", err)
	}
	defer cleanup()
	repo := &alertEventRepo{newBase(deps, slog.Default())}
	// 同时构造一个没有历史分库的仓储，用来验证降级路径。
	hotOnlyRepo := &alertEventRepo{newBase(&Dependencies{db: deps.db}, slog.Default())}
	ctx := context.Background()

	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		t.Fatalf("load location: %v", err)
	}
	now := time.Now().In(loc)
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	to := today.AddDate(0, 0, 1)
	from := today.AddDate(0, 0, -29)
	scope := bo.AlertAnalysisScope{From: from, To: to, Location: loc}

	var namespaceUID string
	if err := deps.db.QueryRowContext(ctx, `SELECT namespace_uid FROM alert_events
		WHERE starts_at >= $1 AND starts_at < $2 GROUP BY 1 ORDER BY count(*) DESC LIMIT 1`,
		from, to).Scan(&namespaceUID); err != nil {
		t.Skipf("no alert events in the analysis window: %v", err)
	}

	summary, err := repo.AnalysisSummary(ctx, namespaceUID, scope)
	if err != nil {
		t.Fatalf("AnalysisSummary: %v", err)
	}
	if summary.Events == 0 {
		t.Skip("no historical events in the analysis window")
	}
	if !summary.Complete {
		t.Fatal("Complete = false, want true when history storage is available")
	}
	if summary.UniqueFingerprints < 0 || summary.UniqueFingerprints > summary.Events {
		t.Fatalf("unique fingerprints = %d, events = %d", summary.UniqueFingerprints, summary.Events)
	}
	if summary.Repeated < 0 || summary.Repeated > summary.Events {
		t.Fatalf("repeated = %d, events = %d", summary.Repeated, summary.Events)
	}
	if summary.TimeZone != "Asia/Shanghai" || !summary.From.Equal(from) || !summary.To.Equal(to) {
		t.Fatalf("window = %s %v..%v", summary.TimeZone, summary.From, summary.To)
	}

	status, err := repo.AnalysisStatus(ctx, namespaceUID, scope)
	if err != nil {
		t.Fatalf("AnalysisStatus: %v", err)
	}
	if !status.Complete || status.Total != summary.Events {
		t.Fatalf("status total = %d (complete=%v), events = %d", status.Total, status.Complete, summary.Events)
	}
	for _, item := range status.Items {
		if item.Status == "" || item.Events <= 0 {
			t.Fatalf("invalid status row: %+v", item)
		}
	}

	level, err := repo.AnalysisLevel(ctx, namespaceUID, scope)
	if err != nil {
		t.Fatalf("AnalysisLevel: %v", err)
	}
	if !level.Complete || level.Total != summary.Events {
		t.Fatalf("level total = %d (complete=%v), events = %d", level.Total, level.Complete, summary.Events)
	}

	for _, granularity := range []string{"hour", "day", "week", "month"} {
		trend, err := repo.AnalysisTrend(ctx, namespaceUID, scope, granularity)
		if err != nil {
			t.Fatalf("AnalysisTrend(%s): %v", granularity, err)
		}
		if !trend.Complete {
			t.Fatalf("trend(%s) complete = false", granularity)
		}
		var bucketTotal int64
		previous := time.Time{}
		for _, bucket := range trend.Buckets {
			if bucket.Events <= 0 {
				t.Fatalf("empty bucket returned: %+v", bucket)
			}
			if !previous.IsZero() && !bucket.Start.After(previous) {
				t.Fatalf("buckets not ascending: %v then %v", previous, bucket.Start)
			}
			previous = bucket.Start
			bucketTotal += bucket.Events
			local := bucket.Start.In(loc)
			switch granularity {
			case "hour":
				if local.Minute() != 0 || local.Second() != 0 {
					t.Fatalf("hour bucket not aligned: %v", local)
				}
			case "day":
				if local.Hour() != 0 || local.Minute() != 0 {
					t.Fatalf("day bucket not local midnight: %v", local)
				}
			case "week":
				if local.Weekday() != time.Monday || local.Hour() != 0 {
					t.Fatalf("week bucket not Monday midnight: %v", local)
				}
			case "month":
				if local.Day() != 1 || local.Hour() != 0 {
					t.Fatalf("month bucket not first day midnight: %v", local)
				}
			}
		}
		if bucketTotal != summary.Events {
			t.Fatalf("trend(%s) total = %d, events = %d", granularity, bucketTotal, summary.Events)
		}
	}

	fingerprints, err := repo.AnalysisFingerprints(ctx, namespaceUID, scope, 5)
	if err != nil {
		t.Fatalf("AnalysisFingerprints: %v", err)
	}
	if !fingerprints.Complete || len(fingerprints.Items) > 5 {
		t.Fatalf("fingerprints = %d items (complete=%v)", len(fingerprints.Items), fingerprints.Complete)
	}
	for index, item := range fingerprints.Items {
		if item.Fingerprint == "" || item.Events <= 1 {
			t.Fatalf("invalid fingerprint row: %+v", item)
		}
		if index > 0 {
			previous := fingerprints.Items[index-1]
			if previous.Events < item.Events ||
				(previous.Events == item.Events && previous.Fingerprint > item.Fingerprint) {
				t.Fatalf("fingerprints not sorted: %+v then %+v", previous, item)
			}
		}
	}

	recovery, err := repo.AnalysisRecovery(ctx, namespaceUID, scope)
	if err != nil {
		t.Fatalf("AnalysisRecovery: %v", err)
	}
	if !recovery.Complete {
		t.Fatal("recovery complete = false")
	}
	for name, stats := range map[string]bo.AlertAnalysisRecoveryStats{
		"auto": recovery.Auto, "manual": recovery.Manual,
	} {
		if stats.Samples < 0 || stats.Samples > summary.Events {
			t.Fatalf("%s samples = %d, events = %d", name, stats.Samples, summary.Events)
		}
		if stats.Samples > 0 && (stats.P90Seconds < stats.MedianSeconds || stats.MedianSeconds < 0) {
			t.Fatalf("%s percentiles = %+v", name, stats)
		}
	}

	// 状态过滤在整段聚合之后叠加：RECOVERED 只能是事件总数的子集。
	filtered := scope
	filtered.Status = "RECOVERED"
	recoveredStatus, err := repo.AnalysisStatus(ctx, namespaceUID, filtered)
	if err != nil {
		t.Fatalf("AnalysisStatus(RECOVERED): %v", err)
	}
	if recoveredStatus.Total > summary.Events {
		t.Fatalf("filtered total = %d > events = %d", recoveredStatus.Total, summary.Events)
	}
	for _, item := range recoveredStatus.Items {
		if item.Status != "RECOVERED" {
			t.Fatalf("status filter leaked %q", item.Status)
		}
	}

	// 历史分库不可用时只返回主库结果，Complete 为假，不整体报错。
	hotOnly, err := hotOnlyRepo.AnalysisSummary(ctx, namespaceUID, scope)
	if err != nil {
		t.Fatalf("AnalysisSummary without history: %v", err)
	}
	if hotOnly.Complete {
		t.Fatal("hot-only summary complete = true, want false")
	}
	if hotOnly.Events > summary.Events {
		t.Fatalf("hot-only events = %d > merged events = %d", hotOnly.Events, summary.Events)
	}
}
