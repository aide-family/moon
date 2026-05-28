package evaluator

import (
	"testing"
	"time"

	"github.com/aide-family/magicbox/enum"
	"github.com/aide-family/magicbox/strutil/cnst"
	"github.com/prometheus/common/model"

	"github.com/aide-family/marksman/internal/biz/bo"
)

func TestBuildComparisonExpr(t *testing.T) {
	tests := []struct {
		name   string
		expr   string
		cond   enum.ConditionMetric
		values []float64
		want   string
	}{
		{
			name: "greater than",
			expr: "rate(http_requests_total[5m])",
			cond: enum.ConditionMetric_CONDITION_METRIC_GT,
			values: []float64{0.5},
			want:   "(rate(http_requests_total[5m])) > 0.5",
		},
		{
			name: "in range",
			expr: "node_cpu_seconds_total",
			cond: enum.ConditionMetric_CONDITION_METRIC_IN,
			values: []float64{0.1, 0.9},
			want:   "((node_cpu_seconds_total) >= 0.1 and (node_cpu_seconds_total) <= 0.9)",
		},
		{
			name: "passthrough without condition",
			expr: "up == 0",
			cond: enum.ConditionMetric_CONDITION_METRIC_UNKNOWN,
			values: nil,
			want:   "up == 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildComparisonExpr(tt.expr, tt.cond, tt.values)
			if got != tt.want {
				t.Fatalf("buildComparisonExpr() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestBuildModeExpr(t *testing.T) {
	comparison := "(metric) > 1"
	got := buildModeExpr(comparison, enum.SampleMode_SAMPLE_MODE_MIN, []float64{1, 3}, 5*time.Minute)
	want := `count_over_time(((metric) > 1)[300s:]) >= 3`
	if got != want {
		t.Fatalf("buildModeExpr() = %q, want %q", got, want)
	}
}

func TestExecutePrometheusTemplate(t *testing.T) {
	got := executePrometheusTemplate(
		`High usage on {{ $labels.instance }} value={{ $value }}`,
		map[string]string{"instance": "host-a"},
		42,
	)
	want := "High usage on host-a value=42"
	if got != want {
		t.Fatalf("executePrometheusTemplate() = %q, want %q", got, want)
	}
}

func TestIsSeriesFiring_WithForDuration(t *testing.T) {
	end := time.Unix(600, 0)
	step := 60 * time.Second
	forDur := 3 * time.Minute

	series := &model.SampleStream{
		Metric: model.Metric{"instance": "a"},
		Values: []model.SamplePair{
			{Timestamp: model.TimeFromUnix(240), Value: 1},
			{Timestamp: model.TimeFromUnix(300), Value: 1},
			{Timestamp: model.TimeFromUnix(360), Value: 1},
			{Timestamp: model.TimeFromUnix(420), Value: 1},
			{Timestamp: model.TimeFromUnix(480), Value: 1},
			{Timestamp: model.TimeFromUnix(540), Value: 1},
			{Timestamp: model.TimeFromUnix(600), Value: 1},
		},
	}

	ok, value := isSeriesFiring(series, forDur, step, end)
	if !ok {
		t.Fatal("expected series to be firing")
	}
	if value != 1 {
		t.Fatalf("unexpected value: got %v, want 1", value)
	}
}

func TestBuildRuleLabelsContainsSystemFields(t *testing.T) {
	info, err := bo.NewEvaluateMetricStrategyBo(
		1,
		&bo.StrategyGroupItemBo{UID: 2, Name: "group-a"},
		&bo.StrategyItemBo{UID: 3, Name: "HighCPU"},
		&bo.StrategyMetricItemBo{
			Expr:   "cpu_usage",
			Labels: map[string]string{"team": "ops"},
		},
		&bo.StrategyMetricLevelItemBo{
			Level: &bo.LevelItemBo{UID: 4, Name: "critical"},
		},
		&bo.DatasourceItemBo{UID: 5, Name: "prom-main"},
	)
	if err != nil {
		t.Fatalf("NewEvaluateMetricStrategyBo failed: %v", err)
	}

	labels := buildRuleLabels(info)
	if labels[cnst.LabelAlertName] != "HighCPU" {
		t.Fatalf("alertname = %q, want HighCPU", labels[cnst.LabelAlertName])
	}
	if labels[cnst.LabelSeverity] != "critical" {
		t.Fatalf("severity = %q, want critical", labels[cnst.LabelSeverity])
	}
	if labels["team"] != "ops" {
		t.Fatalf("team label missing")
	}
	if labels[cnst.LabelStrategyUID] != "3" {
		t.Fatalf("strategy_uid = %q, want 3", labels[cnst.LabelStrategyUID])
	}
	if labels[cnst.LabelDatasourceUID] != "5" {
		t.Fatalf("datasource_uid = %q, want 5", labels[cnst.LabelDatasourceUID])
	}
}

func TestPrometheusAlertFingerprintIgnoresVolatileNameLabels(t *testing.T) {
	base := map[string]string{
		cnst.LabelNamespaceUID:  "1",
		cnst.LabelStrategyUID:   "3",
		cnst.LabelDatasourceUID: "10",
		"instance":              "host-a",
	}
	labelsA := mapsCloneString(base)
	labelsA[cnst.LabelAlertName] = "OldName"
	labelsA[cnst.LabelStrategyGroupName] = "group-a"
	labelsA[cnst.LabelSeverity] = "warning"
	labelsB := mapsCloneString(base)
	labelsB[cnst.LabelAlertName] = "NewName"
	labelsB[cnst.LabelStrategyGroupName] = "group-b"
	labelsB[cnst.LabelSeverity] = "critical"
	if prometheusAlertFingerprint(labelsA) != prometheusAlertFingerprint(labelsB) {
		t.Fatal("expected identical fingerprints when only volatile name labels differ")
	}
}

func TestPrometheusAlertFingerprintDiffersByDatasourceUID(t *testing.T) {
	base := map[string]string{
		cnst.LabelAlertName:   "HighCPU",
		cnst.LabelStrategyUID: "3",
		"instance":            "host-a",
	}
	labelsA := mapsCloneString(base)
	labelsA[cnst.LabelDatasourceUID] = "10"
	labelsB := mapsCloneString(base)
	labelsB[cnst.LabelDatasourceUID] = "20"
	if prometheusAlertFingerprint(labelsA) == prometheusAlertFingerprint(labelsB) {
		t.Fatal("expected different fingerprints for different datasource_uid")
	}
}

func mapsCloneString(m map[string]string) map[string]string {
	out := make(map[string]string, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}
