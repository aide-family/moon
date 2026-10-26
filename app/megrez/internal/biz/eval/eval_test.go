package eval

import (
	"context"
	"log/slog"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/aide-family/Alioth/app/megrez/internal/data/repository"
	"github.com/aide-family/Alioth/app/megrez/pkg/bo"
	"github.com/aide-family/Alioth/pkg/alerttmpl"
	"github.com/aide-family/Alioth/pkg/merr"
)

func TestHitsConditions(t *testing.T) {
	now := time.Date(2026, 10, 15, 0, 0, 0, 0, time.UTC)
	points := []bo.Point{
		{At: now, Value: 2, Raw: "2"},
		{At: now.Add(15 * time.Second), Value: 3, Raw: "3"},
	}
	rule := bo.RuleSnapshot{UID: "rule-1", DatasourceUID: "ds-1", Summary: "cpu", Levels: []bo.Level{{
		Condition: "GT", Sample: "MAX", Threshold: "1", LevelUID: "high",
	}}}
	hits := Hits(rule, []bo.Sample{{Labels: map[string]string{"job": "api"}, Points: points}})
	if len(hits) != 1 || hits[0].Value != "3" || hits[0].LevelUID != "high" || hits[0].Fingerprint == "" {
		t.Fatalf("max hit = %+v", hits)
	}
	rule.Levels[0].Sample = "MIN"
	rule.Levels[0].Threshold = "2"
	if len(Hits(rule, []bo.Sample{{Points: points}})) != 0 {
		t.Fatal("min should miss")
	}
	rule.Levels[0].Sample = "FOR"
	rule.Levels[0].Threshold = "1"
	rule.Levels[0].ForSeconds = 15
	got := Hits(rule, []bo.Sample{{Labels: map[string]string{"job": "api"}, Points: points}})
	if len(got) != 1 || got[0].Value != "3" {
		t.Fatalf("for hit = %+v", got)
	}
	rule.Summary = "测试告警, value: {{ $value }} job {{ $labels.job }}"
	rule.Description = "测试告警, value: {{ .Value }}"
	rule.Levels[0].Condition = "GT"
	rule.Levels[0].Sample = "MAX"
	rule.Levels[0].Threshold = "1"
	filled := Hits(rule, []bo.Sample{{Labels: map[string]string{"job": "api"}, Points: points}})
	if len(filled) != 1 || filled[0].Summary != "测试告警, value: 3 job api" || filled[0].Description != "测试告警, value: 3" {
		t.Fatalf("template hit = %+v", filled)
	}
	// {{ . }} 是采样值，不是 Prometheus 的 AlertTemplateData 结构体。
	rule.Description = "测试告警, value: {{ . }}"
	filled = Hits(rule, []bo.Sample{{Labels: map[string]string{"job": "api"}, Points: points}})
	if len(filled) != 1 || filled[0].Description != "测试告警, value: 3" || strings.Contains(filled[0].Description, "map[") {
		t.Fatalf("dot template = %+v", filled)
	}
	rule.Description = "up指标测试，告警明细 {{ .labels }} ; value: {{ $value }}"
	canonicalLabels := `{"__name__":"up","instance":"localhost:9090","job":"prometheus"}`
	filled = Hits(rule, []bo.Sample{{Labels: map[string]string{"__name__": "up", "instance": "localhost:9090", "job": "prometheus"}, Points: points}})
	if len(filled) != 1 || filled[0].Description != "up指标测试，告警明细 "+canonicalLabels+" ; value: 3" || filled[0].Labels != canonicalLabels {
		t.Fatalf("event labels and description template diverged = %+v", filled)
	}
	// 规则标签参与 fingerprint 和事件标签，修改规则标签会产生新事件。
	rule.Labels = `{"cluster":"prod"}`
	withRuleLabels := Hits(rule, []bo.Sample{{Labels: map[string]string{"job": "api"}, Points: points}})
	if len(withRuleLabels) != 1 || !strings.Contains(withRuleLabels[0].Labels, `"cluster":"prod"`) {
		t.Fatalf("rule labels = %+v", withRuleLabels)
	}
	if withRuleLabels[0].Fingerprint == filled[0].Fingerprint {
		t.Fatal("rule label change must change fingerprint")
	}
	rule.Labels = ""

	rule.Endpoint = "https://prometheus.example/base"
	rule.Description = `{{ $labels.instance }} {{ printf "%.2f" $value }} {{ humanize1024 2048 }} {{ $externalURL }}`
	filled = Hits(rule, []bo.Sample{{Labels: map[string]string{"instance": "localhost:9090"}, Points: points}})
	if len(filled) != 1 || filled[0].Description != "localhost:9090 3.00 2ki https://prometheus.example/base" {
		t.Fatalf("Prometheus native template = %+v", filled)
	}
	rule.Description = `{{ with query "up" }}{{ with first . }}{{ .Labels.instance }}={{ value . }}{{ end }}{{ end }}`
	query := func(context.Context, string, time.Time) ([]alerttmpl.Sample, error) {
		return []alerttmpl.Sample{{Labels: map[string]string{"instance": "query-target"}, Value: 9}}, nil
	}
	filled = HitsWithQuery(context.Background(), rule, []bo.Sample{{Labels: map[string]string{"instance": "localhost:9090"}, Points: points}}, now, query)
	if len(filled) != 1 || filled[0].Description != "query-target=9" {
		t.Fatalf("Prometheus query template = %+v", filled)
	}
	rule.Summary = "{{"
	if got := Hits(rule, []bo.Sample{{Labels: map[string]string{"job": "api"}, Points: points}}); len(got) != 1 || !strings.HasPrefix(got[0].Summary, "<error expanding template:") {
		t.Fatalf("broken template = %+v", got)
	}
	rule.Summary = "cpu"
	rule.Description = ""
	rule.Levels[0].Sample = "FOR"
	rule.Levels[0].Condition = "IN"
	rule.Levels[0].Threshold = "3,9"
	if len(Hits(rule, []bo.Sample{{Points: points}})) != 0 {
		t.Fatal("for in should miss when a point is outside")
	}
}

type fakeHold struct{ renew error }

func (h fakeHold) Renew(context.Context, time.Duration) error { return h.renew }
func (h fakeHold) Release(context.Context) error              { return nil }

type fakeGateway struct {
	reports  []bo.EvalReport
	statuses []bo.WorkerStatus
}

func (g *fakeGateway) Claim(context.Context) ([]bo.RuleSnapshot, error) { return nil, nil }
func (g *fakeGateway) Report(_ context.Context, report bo.EvalReport) error {
	g.reports = append(g.reports, report)
	return nil
}
func (g *fakeGateway) ReportStatus(_ context.Context, status bo.WorkerStatus) error {
	g.statuses = append(g.statuses, status)
	return nil
}

type fakeQuery struct{}

func (fakeQuery) QueryRange(context.Context, bo.RuleSnapshot, int64, int64, int64) ([]bo.Sample, error) {
	return []bo.Sample{}, nil
}

type successLocker struct{}

func (successLocker) Available(context.Context) error { return nil }
func (successLocker) Acquire(context.Context, string, time.Duration) (repository.Hold, error) {
	return fakeHold{}, nil
}

func TestEvalOneReportsTaskLifecycle(t *testing.T) {
	now := time.Date(2026, 10, 21, 8, 0, 0, 0, time.UTC)
	gateway := &fakeGateway{}
	worker := &Worker{
		gateway: gateway, locks: successLocker{}, query: fakeQuery{}, log: slog.Default(),
		now: func() time.Time { return now }, instanceUID: "host-1", hostname: "host",
	}
	if err := worker.evalOne(context.Background(), bo.RuleSnapshot{UID: "rule-1", Name: "CPU high", Driver: "unsupported"}); err != nil {
		t.Fatal(err)
	}
	if len(gateway.statuses) != 2 {
		t.Fatalf("reported %d status transitions, want active and idle", len(gateway.statuses))
	}
	active, idle := gateway.statuses[0], gateway.statuses[1]
	if !active.Busy || active.TaskUID != "rule-1" || active.TaskName != "CPU high" || !active.TaskStartedAt.Equal(now) {
		t.Fatalf("active status = %+v", active)
	}
	if idle.Busy || idle.TaskUID != "" || idle.TaskName != "" || !idle.TaskStartedAt.IsZero() {
		t.Fatalf("idle status = %+v", idle)
	}
}

func TestRenewFailureDoesNotReport(t *testing.T) {
	gateway := &fakeGateway{}
	worker := &Worker{gateway: gateway, query: fakeQuery{}, log: slog.Default(), now: time.Now}
	err := worker.finish(context.Background(), fakeHold{renew: merr.Server("lock failed")}, bo.RuleSnapshot{
		UID: "rule-1", Driver: "METRICS_PROMETHEUS", EvalSlot: time.Now().UTC(),
	})
	if err == nil {
		t.Fatal("renew failure returned success")
	}
	if len(gateway.reports) != 0 {
		t.Fatalf("reported after renew failure: %+v", gateway.reports)
	}
}

type failLocker struct{}

func (failLocker) Available(context.Context) error { return merr.Server("redis down") }
func (failLocker) Acquire(context.Context, string, time.Duration) (repository.Hold, error) {
	return fakeHold{}, nil
}

func TestClaimFailureStretches(t *testing.T) {
	worker := &Worker{locks: failLocker{}, gateway: &fakeGateway{}, log: slog.Default(), now: time.Now}
	if !worker.once(context.Background()) {
		t.Fatal("redis failure should stretch")
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Nanosecond)
	defer cancel()
	time.Sleep(time.Millisecond)
	if !worker.once(ctx) {
		t.Fatal("deadline should stretch")
	}
	cancel()
	stopped, stop := context.WithCancel(context.Background())
	stop()
	if worker.once(stopped) {
		t.Fatal("shutdown should not stretch")
	}
}

func TestUnimplementedDriverReportsSkipped(t *testing.T) {
	gateway := &fakeGateway{}
	worker := &Worker{gateway: gateway, query: fakeQuery{}, log: slog.Default(), now: time.Now}
	slot := time.Now().UTC().Truncate(time.Second)
	for _, driver := range []string{"TRACE_JAEGER"} {
		gateway.reports = nil
		if err := worker.finish(context.Background(), fakeHold{}, bo.RuleSnapshot{UID: "rule-1", Driver: driver, EvalSlot: slot}); err != nil {
			t.Fatal(err)
		}
		if len(gateway.reports) != 1 || !gateway.reports[0].Skipped || len(gateway.reports[0].Hits) != 0 {
			t.Fatalf("driver %s report = %+v", driver, gateway.reports)
		}
	}
}

type recordingQuery struct {
	driver  string
	samples []bo.Sample
}

func (q *recordingQuery) QueryRange(_ context.Context, rule bo.RuleSnapshot, _, _, _ int64) ([]bo.Sample, error) {
	q.driver = rule.Driver
	return q.samples, nil
}

func TestLokiDriverEvaluatesNumericLogQLSamples(t *testing.T) {
	gateway := &fakeGateway{}
	query := &recordingQuery{samples: []bo.Sample{{
		Labels: map[string]string{"app": "api"},
		Points: []bo.Point{{At: time.Now(), Value: 3, Raw: "3"}},
	}}}
	worker := &Worker{gateway: gateway, query: query, log: slog.Default(), now: time.Now}
	rule := bo.RuleSnapshot{
		UID: "loki-rule", Driver: "LOGS_LOKI", DatasourceUID: "loki-ds", EvalSlot: time.Now().UTC(),
		Expr:   `sum by (app) (count_over_time({app="api"} |= "error" [5m]))`,
		Levels: []bo.Level{{Condition: "GT", Sample: "MAX", Threshold: "2", LevelUID: "high"}},
	}
	if err := worker.finish(context.Background(), fakeHold{}, rule); err != nil {
		t.Fatal(err)
	}
	if query.driver != "LOGS_LOKI" {
		t.Fatalf("query driver = %q", query.driver)
	}
	if len(gateway.reports) != 1 || gateway.reports[0].Skipped || len(gateway.reports[0].Hits) != 1 {
		t.Fatalf("Loki eval report = %+v", gateway.reports)
	}
	if gateway.reports[0].Hits[0].Value != "3" || gateway.reports[0].Hits[0].LevelUID != "high" {
		t.Fatalf("Loki hit = %+v", gateway.reports[0].Hits[0])
	}
}

func TestLokiEmptyMatrixReportsSuccessfulNoHit(t *testing.T) {
	gateway := &fakeGateway{}
	query := &recordingQuery{}
	worker := &Worker{gateway: gateway, query: query, log: slog.Default(), now: time.Now}
	err := worker.finish(context.Background(), fakeHold{}, bo.RuleSnapshot{
		UID: "loki-rule", Driver: "LOGS_LOKI", DatasourceUID: "loki-ds", EvalSlot: time.Now().UTC(),
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(gateway.reports) != 1 || gateway.reports[0].Skipped || len(gateway.reports[0].Hits) != 0 {
		t.Fatalf("empty Loki matrix report = %+v", gateway.reports)
	}
}

func TestVlogsDriverEvaluatesNumericStatsSamples(t *testing.T) {
	gateway := &fakeGateway{}
	points := []bo.Point{
		{At: time.Now().Add(-15 * time.Second), Value: 3, Raw: "3"},
		{At: time.Now(), Value: 4, Raw: "4"},
	}
	query := &recordingQuery{samples: []bo.Sample{{
		Labels: map[string]string{"app": "api", "level": "error"},
		Points: points,
	}}}
	worker := &Worker{gateway: gateway, query: query, log: slog.Default(), now: time.Now}
	rule := bo.RuleSnapshot{
		UID: "vlogs-rule", Driver: "LOGS_VLOGS", DatasourceUID: "vlogs-ds", EvalSlot: time.Now().UTC(),
		Expr:   `level:error | stats by (app, level) count() as value`,
		Levels: []bo.Level{{Condition: "GT", Sample: "MAX", Threshold: "2", LevelUID: "high"}},
	}
	if err := worker.finish(context.Background(), fakeHold{}, rule); err != nil {
		t.Fatal(err)
	}
	if query.driver != "LOGS_VLOGS" {
		t.Fatalf("query driver = %q", query.driver)
	}
	if len(gateway.reports) != 1 || gateway.reports[0].Skipped || len(gateway.reports[0].Hits) != 1 {
		t.Fatalf("VictoriaLogs eval report = %+v", gateway.reports)
	}
	hit := gateway.reports[0].Hits[0]
	if hit.Value != "4" || hit.LevelUID != "high" || !strings.Contains(hit.Labels, `"app":"api"`) {
		t.Fatalf("VictoriaLogs hit = %+v", hit)
	}
}

func TestVlogsEmptyStatsReportsSuccessfulNoHit(t *testing.T) {
	gateway := &fakeGateway{}
	query := &recordingQuery{}
	worker := &Worker{gateway: gateway, query: query, log: slog.Default(), now: time.Now}
	err := worker.finish(context.Background(), fakeHold{}, bo.RuleSnapshot{
		UID: "vlogs-rule", Driver: "LOGS_VLOGS", DatasourceUID: "vlogs-ds", EvalSlot: time.Now().UTC(),
		Expr: `level:error | stats by (app) count() as value`,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(gateway.reports) != 1 || gateway.reports[0].Skipped || len(gateway.reports[0].Hits) != 0 {
		t.Fatalf("empty VictoriaLogs stats report = %+v", gateway.reports)
	}
}

func TestVlogsWindowCoversConfiguredForDuration(t *testing.T) {
	now := time.Date(2026, 10, 22, 12, 0, 7, 0, time.UTC)
	rule := bo.RuleSnapshot{
		Driver: "LOGS_VLOGS",
		Expr:   `level:error | stats by (app) count() as value`,
		Levels: []bo.Level{{Condition: "GT", Sample: "FOR", Threshold: "1", ForSeconds: 300}},
	}
	start, end, step := window(rule, now)
	if step != 15 || end != now.Truncate(15*time.Second).Unix() || end-start != 315 {
		t.Fatalf("window = %d..%d step %d, want aligned 315s window", start, end, step)
	}

	rule.Levels = []bo.Level{{Condition: "GT", Sample: "FOR", Threshold: "1", ForSeconds: 100}}
	start, end, step = window(rule, now)
	if step != 15 || end%15 != 0 || start%15 != 0 || end-start != 120 {
		t.Fatalf("100s FOR window = %d..%d step %d, want aligned 120s window", start, end, step)
	}

	// 短于默认步长的持续窗口按 FOR 秒数切桶，避免 10 秒规则要等满两个 15 秒桶。
	rule.Levels = []bo.Level{{Condition: "GT", Sample: "FOR", Threshold: "0", ForSeconds: 10}}
	start, end, step = window(rule, now)
	if step != 10 || end%10 != 0 || start%10 != 0 || end-start != 20 {
		t.Fatalf("10s FOR window = %d..%d step %d, want aligned 20s window", start, end, step)
	}
	at := time.Unix(end, 0).UTC()
	points := []bo.Point{
		{At: at.Add(-10 * time.Second), Value: 1, Raw: "1"},
		{At: at.Add(-20 * time.Second), Value: 1, Raw: "1"},
	}
	// 采样按时间升序排列，和 vlogsQuery 返回顺序保持一致。
	slices.SortFunc(points, func(left, right bo.Point) int { return left.At.Compare(right.At) })
	hits := Hits(bo.RuleSnapshot{Driver: "LOGS_VLOGS", Levels: rule.Levels}, []bo.Sample{{Points: points, Step: step}})
	if len(hits) != 1 {
		t.Fatalf("10s FOR buckets = %+v, want one hit", hits)
	}

	rule.Levels = []bo.Level{{Condition: "GT", Sample: "MAX", Threshold: "1"}}
	start, end, step = window(rule, now)
	if step != 15 || end-start != 15 {
		t.Fatalf("max window = %d..%d step %d, want one full bucket", start, end, step)
	}

	start, end, _ = window(bo.RuleSnapshot{
		Driver: "LOGS_LOKI",
		Levels: []bo.Level{{Condition: "GT", Sample: "FOR", Threshold: "1", ForSeconds: 300}},
	}, now)
	if end != now.Unix() || end-start != 300 {
		t.Fatalf("Loki window changed by VictoriaLogs padding: %d..%d", start, end)
	}
}

func TestVlogsForRejectsMissingBuckets(t *testing.T) {
	at := time.Date(2026, 10, 22, 12, 0, 0, 0, time.UTC)
	rule := bo.RuleSnapshot{Levels: []bo.Level{{
		Condition: "GT", Sample: "FOR", Threshold: "0", ForSeconds: 30,
	}}}
	continuous := []bo.Sample{{
		Points: []bo.Point{
			{At: at, Value: 1, Raw: "1"},
			{At: at.Add(15 * time.Second), Value: 1, Raw: "1"},
			{At: at.Add(30 * time.Second), Value: 1, Raw: "1"},
		},
		Step: 15,
	}}
	if hits := Hits(rule, continuous); len(hits) != 1 {
		t.Fatalf("continuous buckets = %+v", hits)
	}
	gapped := []bo.Sample{{
		Points: []bo.Point{
			{At: at, Value: 1, Raw: "1"},
			{At: at.Add(45 * time.Second), Value: 1, Raw: "1"},
		},
		Step: 15,
	}}
	if hits := Hits(rule, gapped); len(hits) != 0 {
		t.Fatalf("missing bucket must not satisfy FOR: %+v", hits)
	}
	// 上游时间戳可能带亚秒抖动，不能因此把连续桶判成缺桶。
	jittered := []bo.Sample{{
		Points: []bo.Point{
			{At: at, Value: 1, Raw: "1"},
			{At: at.Add(15*time.Second + time.Millisecond), Value: 1, Raw: "1"},
			{At: at.Add(30*time.Second + 2*time.Millisecond), Value: 1, Raw: "1"},
		},
		Step: 15,
	}}
	if hits := Hits(rule, jittered); len(hits) != 1 {
		t.Fatalf("sub-second timestamp jitter must not break FOR: %+v", hits)
	}
}

func TestMetricWindowCoversVictoriaMetricsLatencyOffset(t *testing.T) {
	now := time.Date(2026, 10, 25, 3, 20, 7, 0, time.UTC)
	// 15 秒窗口结束于当前时刻时 VictoriaMetrics 返回空矩阵（默认 latencyOffset=30s），
	// FOR 10 秒和 MAX/MIN 规则都会拿不到采样；窗口至少要有两个 step。
	for _, level := range []bo.Level{
		{Condition: "GT", Sample: "FOR", Threshold: "10", ForSeconds: 10},
		{Condition: "GT", Sample: "MAX", Threshold: "10"},
		{Condition: "GT", Sample: "MIN", Threshold: "10"},
	} {
		start, end, step := window(bo.RuleSnapshot{Driver: "METRICS_VICTORIA_METRICS", Levels: []bo.Level{level}}, now)
		if step != 15 || end != now.Unix() || end-start != 30 {
			t.Fatalf("%s window = %d..%d step %d, want 30s", level.Sample, start, end, step)
		}
	}
	// Prometheus 没有延迟偏移，一个 step 的窗口保持原样。
	start, end, _ := window(bo.RuleSnapshot{Driver: "METRICS_PROMETHEUS", Levels: []bo.Level{{
		Condition: "GT", Sample: "FOR", Threshold: "10", ForSeconds: 10,
	}}}, now)
	if end-start != 15 {
		t.Fatalf("prometheus window = %d..%d, want 15s", start, end)
	}
}

func TestMetricWindowRoundsForSecondsToStep(t *testing.T) {
	now := time.Date(2026, 10, 25, 3, 20, 7, 0, time.UTC)
	// 20 秒 FOR 按 15 秒步长取点只覆盖 15 秒，首末跨度小于 ForSeconds，永远不命中。
	start, end, step := window(bo.RuleSnapshot{Driver: "METRICS_PROMETHEUS", Levels: []bo.Level{{
		Condition: "GT", Sample: "FOR", Threshold: "1", ForSeconds: 20,
	}}}, now)
	if step != 15 || end-start != 30 {
		t.Fatalf("20s FOR window = %d..%d step %d, want 30s", start, end, step)
	}
}

func TestElasticsearchDriverEvaluatesCountSamples(t *testing.T) {
	gateway := &fakeGateway{}
	now := time.Now()
	query := &recordingQuery{samples: []bo.Sample{{
		Points: []bo.Point{
			{At: now.Add(-15 * time.Second), Value: 0, Raw: "0"},
			{At: now, Value: 7, Raw: "7"},
		},
		Step: 15,
	}}}
	worker := &Worker{gateway: gateway, query: query, log: slog.Default(), now: time.Now}
	rule := bo.RuleSnapshot{
		UID: "es-rule", Driver: "LOGS_ELASTICSEARCH", DatasourceUID: "es-ds", EvalSlot: now.UTC(),
		Expr:   `{"match":{"level":"error"}}`,
		Levels: []bo.Level{{Condition: "GT", Sample: "MAX", Threshold: "0", LevelUID: "high"}},
	}
	if err := worker.finish(context.Background(), fakeHold{}, rule); err != nil {
		t.Fatal(err)
	}
	if query.driver != "LOGS_ELASTICSEARCH" {
		t.Fatalf("query driver = %q", query.driver)
	}
	if len(gateway.reports) != 1 || gateway.reports[0].Skipped || len(gateway.reports[0].Hits) != 1 {
		t.Fatalf("Elasticsearch eval report = %+v", gateway.reports)
	}
	hit := gateway.reports[0].Hits[0]
	if hit.Value != "7" || hit.LevelUID != "high" {
		t.Fatalf("Elasticsearch hit = %+v", hit)
	}
}

func TestElasticsearchWindowAlignsToBucketBoundary(t *testing.T) {
	now := time.Date(2026, 10, 25, 12, 0, 7, 0, time.UTC)
	aligned := now.Truncate(15 * time.Second).Unix()

	start, end, step := window(bo.RuleSnapshot{Driver: "LOGS_ELASTICSEARCH"}, now)
	if step != 15 || end != aligned || end-start != 15 {
		t.Fatalf("window = %d..%d step %d, want aligned 15s window", start, end, step)
	}

	// FOR 需要覆盖完整窗口：桶数是 seconds/step，跨度要再补一个桶。
	start, end, step = window(bo.RuleSnapshot{
		Driver: "LOGS_ELASTICSEARCH",
		Levels: []bo.Level{{Condition: "GT", Sample: "FOR", Threshold: "0", ForSeconds: 30}},
	}, now)
	if step != 15 || end != aligned || end-start != 45 {
		t.Fatalf("30s FOR window = %d..%d step %d, want aligned 45s window", start, end, step)
	}
}
