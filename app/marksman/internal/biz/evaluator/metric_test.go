package evaluator

import (
	"context"
	"testing"
	"time"

	"github.com/aide-family/magicbox/enum"
	"github.com/aide-family/magicbox/server/cron"
	"github.com/bwmarrin/snowflake"
	prometheusv1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"

	"github.com/aide-family/marksman/internal/biz/bo"
)

const testDatasourceURL = "http://localhost:9090"

func Test_countTrue(t *testing.T) {
	tests := []struct {
		name string
		b    []bool
		want int
	}{
		{"empty", nil, 0},
		{"empty slice", []bool{}, 0},
		{"all false", []bool{false, false}, 0},
		{"all true", []bool{true, true, true}, 3},
		{"mixed", []bool{true, false, true, true, false}, 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := countTrue(tt.b); got != tt.want {
				t.Errorf("countTrue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_maxConsecutiveTrue(t *testing.T) {
	tests := []struct {
		name string
		b    []bool
		want int
	}{
		{"empty", nil, 0},
		{"empty slice", []bool{}, 0},
		{"all false", []bool{false, false}, 0},
		{"all true", []bool{true, true, true}, 3},
		{"one run", []bool{true, true, false, true}, 2},
		{"two runs", []bool{true, true, false, true, true, true}, 3},
		{"single true", []bool{false, true, false}, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := maxConsecutiveTrue(tt.b); got != tt.want {
				t.Errorf("maxConsecutiveTrue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_metricEvaluator_satisfiesCondition(t *testing.T) {
	nsID := snowflake.ParseInt64(100)
	strategyID := snowflake.ParseInt64(200)
	dsID := snowflake.ParseInt64(300)
	ds := &bo.DatasourceItemBo{UID: dsID, URL: testDatasourceURL}
	info := &bo.EvaluateMetricStrategyBo{
		StrategyUID:  strategyID,
		NamespaceUID: nsID,
		Datasource:   ds,
		Values:       []float64{10, 20}, // threshold 10; for IN [10,20]
	}
	m := &metricEvaluator{info: info}

	tests := []struct {
		name      string
		condition enum.ConditionMetric
		v         float64
		want      bool
	}{
		{"EQ true", enum.ConditionMetric_CONDITION_METRIC_EQ, 10, true},
		{"EQ false", enum.ConditionMetric_CONDITION_METRIC_EQ, 11, false},
		{"NE true", enum.ConditionMetric_CONDITION_METRIC_NE, 11, true},
		{"NE false", enum.ConditionMetric_CONDITION_METRIC_NE, 10, false},
		{"GT true", enum.ConditionMetric_CONDITION_METRIC_GT, 11, true},
		{"GT false", enum.ConditionMetric_CONDITION_METRIC_GT, 10, false},
		{"GTE true eq", enum.ConditionMetric_CONDITION_METRIC_GTE, 10, true},
		{"GTE true gt", enum.ConditionMetric_CONDITION_METRIC_GTE, 11, true},
		{"GTE false", enum.ConditionMetric_CONDITION_METRIC_GTE, 9, false},
		{"LT true", enum.ConditionMetric_CONDITION_METRIC_LT, 9, true},
		{"LT false", enum.ConditionMetric_CONDITION_METRIC_LT, 10, false},
		{"LTE true eq", enum.ConditionMetric_CONDITION_METRIC_LTE, 10, true},
		{"LTE false", enum.ConditionMetric_CONDITION_METRIC_LTE, 11, false},
		{"IN in", enum.ConditionMetric_CONDITION_METRIC_IN, 15, true},
		{"IN low", enum.ConditionMetric_CONDITION_METRIC_IN, 10, true},
		{"IN high", enum.ConditionMetric_CONDITION_METRIC_IN, 20, true},
		{"IN out", enum.ConditionMetric_CONDITION_METRIC_IN, 21, false},
		{"NOT_IN out", enum.ConditionMetric_CONDITION_METRIC_NOT_IN, 21, true},
		{"NOT_IN in", enum.ConditionMetric_CONDITION_METRIC_NOT_IN, 15, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m.info.Condition = tt.condition
			if got := m.satisfiesCondition(tt.v); got != tt.want {
				t.Errorf("satisfiesCondition(%v) = %v, want %v", tt.v, got, tt.want)
			}
		})
	}
}

func Test_metricEvaluator_shouldFireBySampleMode(t *testing.T) {
	nsID := snowflake.ParseInt64(100)
	strategyID := snowflake.ParseInt64(200)
	dsID := snowflake.ParseInt64(300)
	ds := &bo.DatasourceItemBo{UID: dsID, URL: testDatasourceURL}
	level := &bo.LevelItemBo{UID: snowflake.ParseInt64(400), Name: "critical"}

	baseInfo := &bo.EvaluateMetricStrategyBo{
		StrategyUID:  strategyID,
		NamespaceUID: nsID,
		Datasource:   ds,
		Level:        level,
		Values:       []float64{5, 3}, // threshold 5, n=3 for mode
	}
	m := &metricEvaluator{info: baseInfo}

	tests := []struct {
		name      string
		mode      enum.SampleMode
		values    []float64 // override info.Values for n
		satisfied []bool
		want      bool
	}{
		// FOR: n consecutive
		{"FOR 3 consecutive yes", enum.SampleMode_SAMPLE_MODE_FOR, []float64{5, 3}, []bool{false, true, true, true, false}, true},
		{"FOR 3 consecutive no", enum.SampleMode_SAMPLE_MODE_FOR, []float64{5, 3}, []bool{true, true, false, true}, false},
		{"FOR default 1", enum.SampleMode_SAMPLE_MODE_FOR, []float64{5}, []bool{false, true, false}, true},
		// MAX: count > n
		{"MAX more than 2 fire", enum.SampleMode_SAMPLE_MODE_MAX, []float64{5, 2}, []bool{true, true, true}, true},
		{"MAX at most 2 no fire", enum.SampleMode_SAMPLE_MODE_MAX, []float64{5, 2}, []bool{true, true}, false},
		{"MAX zero fire on any", enum.SampleMode_SAMPLE_MODE_MAX, []float64{5, 0}, []bool{true}, true},
		// MIN: count >= n
		{"MIN at least 2 fire", enum.SampleMode_SAMPLE_MODE_MIN, []float64{5, 2}, []bool{true, false, true}, true},
		{"MIN less than 2 no fire", enum.SampleMode_SAMPLE_MODE_MIN, []float64{5, 2}, []bool{true}, false},
		{"MIN default 1", enum.SampleMode_SAMPLE_MODE_MIN, []float64{5}, []bool{false, true}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m.info.Mode = tt.mode
			m.info.Values = tt.values
			if got := m.shouldFireBySampleMode(tt.satisfied); got != tt.want {
				t.Errorf("shouldFireBySampleMode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_metricEvaluator_Index(t *testing.T) {
	nsID := snowflake.ParseInt64(100)
	strategyID := snowflake.ParseInt64(200)
	dsID := snowflake.ParseInt64(300)
	levelID := snowflake.ParseInt64(400)
	ds := &bo.DatasourceItemBo{UID: dsID, URL: testDatasourceURL}
	level := &bo.LevelItemBo{UID: levelID, Name: "warn"}
	info := &bo.EvaluateMetricStrategyBo{
		StrategyUID:  strategyID,
		NamespaceUID: nsID,
		Datasource:   ds,
		Level:        level,
	}
	m := &metricEvaluator{info: info}
	got := m.Index()
	want := "metric-100-200-400-300"
	if got != want {
		t.Errorf("Index() = %v, want %v", got, want)
	}
	// without level
	m.info.Level = nil
	got2 := m.Index()
	want2 := "metric-100-200-0-300"
	if got2 != want2 {
		t.Errorf("Index() without level = %v, want %v", got2, want2)
	}
}

func Test_metricEvaluator_Spec(t *testing.T) {
	info := &bo.EvaluateMetricStrategyBo{
		DurationSec: 0,
	}
	m := &metricEvaluator{info: info}
	got := m.Spec()
	if got != cron.CronSpecEvery(defaultEvaluateInterval) {
		t.Errorf("Spec() with 0 duration = %v", got)
	}
	info.DurationSec = 120
	got = m.Spec()
	want := cron.CronSpecEvery(120 * time.Second)
	if got != want {
		t.Errorf("Spec() = %v, want %v", got, want)
	}
}

// mockQuerier returns a fixed matrix for Run tests (datasource URL is set to localhost:9090 in info).
type mockQuerier struct {
	matrix model.Matrix
	err    error
}

func (q *mockQuerier) ListMetrics(context.Context, *bo.DatasourceItemBo) ([]*bo.MetricSummaryItemBo, error) {
	return nil, nil
}

func (q *mockQuerier) GetMetricDetail(context.Context, *bo.DatasourceItemBo, string) (*bo.MetricDetailItemBo, error) {
	return nil, nil
}

func (q *mockQuerier) QueryRange(_ context.Context, _ *bo.DatasourceItemBo, _ string, _ prometheusv1.Range) (model.Matrix, error) {
	if q.err != nil {
		return nil, q.err
	}
	return q.matrix, nil
}

// mockAlertChannel records events sent for Run tests.
type mockAlertChannel struct {
	events []*bo.AlertEventBo
}

func (c *mockAlertChannel) Send(event *bo.AlertEventBo) {
	c.events = append(c.events, event)
}

func (c *mockAlertChannel) GetChannel() <-chan *bo.AlertEventBo {
	return nil // not used in unit test
}

func Test_metricEvaluator_Run(t *testing.T) {
	nsID := snowflake.ParseInt64(100)
	strategyID := snowflake.ParseInt64(200)
	dsID := snowflake.ParseInt64(300)
	levelID := snowflake.ParseInt64(400)
	ds := &bo.DatasourceItemBo{UID: dsID, URL: testDatasourceURL}
	level := &bo.LevelItemBo{UID: levelID, Name: "critical"}
	// Condition: value > 5; SampleMode MIN with n=2 (at least 2 samples satisfy)
	// Series with values 6,6,6 -> all 3 satisfy -> count>=2 -> fire
	info := &bo.EvaluateMetricStrategyBo{
		StrategyUID:  strategyID,
		NamespaceUID: nsID,
		Datasource:   ds,
		Level:        level,
		Summary:      "test summary",
		Expr:         "up",
		Mode:         enum.SampleMode_SAMPLE_MODE_MIN,
		Condition:    enum.ConditionMetric_CONDITION_METRIC_GT,
		Values:       []float64{5, 2},
		DurationSec:  120,
	}
	matrix := model.Matrix{
		{
			Metric: model.Metric{"job": "test"},
			Values: []model.SamplePair{
				{Timestamp: model.TimeFromUnix(1000), Value: 6},
				{Timestamp: model.TimeFromUnix(1015), Value: 6},
				{Timestamp: model.TimeFromUnix(1030), Value: 6},
			},
		},
	}
	querier := &mockQuerier{matrix: matrix}
	alertCh := &mockAlertChannel{}
	eval := NewMetricEvaluator(querier, alertCh, info)
	eval.(*metricEvaluator).Run()

	if len(alertCh.events) != 1 {
		t.Fatalf("Run() sent %d events, want 1", len(alertCh.events))
	}
	ev := alertCh.events[0]
	if ev.StrategyUID != strategyID || ev.NamespaceUID != nsID {
		t.Errorf("event strategy/namespace = %v/%v", ev.StrategyUID, ev.NamespaceUID)
	}
	if ev.Summary != "test summary" {
		t.Errorf("event summary = %q", ev.Summary)
	}
	if ev.Value != 6 {
		t.Errorf("event value = %v, want 6", ev.Value)
	}
	if ev.DatasourceUID != dsID {
		t.Errorf("event datasource = %v", ev.DatasourceUID)
	}
}

func Test_metricEvaluator_Run_noFire(t *testing.T) {
	ds := &bo.DatasourceItemBo{UID: snowflake.ParseInt64(300), URL: testDatasourceURL}
	info := &bo.EvaluateMetricStrategyBo{
		StrategyUID:  snowflake.ParseInt64(200),
		NamespaceUID: snowflake.ParseInt64(100),
		Datasource:   ds,
		Level:        &bo.LevelItemBo{UID: snowflake.ParseInt64(400)},
		Mode:         enum.SampleMode_SAMPLE_MODE_MIN,
		Condition:    enum.ConditionMetric_CONDITION_METRIC_GT,
		Values:       []float64{5, 2}, // need at least 2 samples > 5
		DurationSec:  120,
	}
	// Only one sample > 5 -> count 1 < 2 -> no fire
	matrix := model.Matrix{
		{
			Metric: model.Metric{"job": "test"},
			Values: []model.SamplePair{
				{Timestamp: model.TimeFromUnix(1000), Value: 3},
				{Timestamp: model.TimeFromUnix(1015), Value: 6}, // only one
			},
		},
	}
	querier := &mockQuerier{matrix: matrix}
	alertCh := &mockAlertChannel{}
	eval := NewMetricEvaluator(querier, alertCh, info)
	eval.(*metricEvaluator).Run()

	if len(alertCh.events) != 0 {
		t.Errorf("Run() sent %d events, want 0", len(alertCh.events))
	}
}

func Test_metricEvaluator_IsImmediate(t *testing.T) {
	eval := NewMetricEvaluator(nil, nil, &bo.EvaluateMetricStrategyBo{Datasource: &bo.DatasourceItemBo{URL: testDatasourceURL}})
	if eval.IsImmediate() != false {
		t.Error("IsImmediate() should be false")
	}
}
