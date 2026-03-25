// Package evaluator provides a metric evaluator.
package evaluator

import (
	"context"
	"maps"
	"time"

	"github.com/aide-family/magicbox/contextx"
	"github.com/aide-family/magicbox/enum"
	"github.com/aide-family/magicbox/server/cron"
	"github.com/aide-family/magicbox/strutil"
	klog "github.com/go-kratos/kratos/v2/log"
	prometheusv1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"

	"github.com/aide-family/marksman/internal/biz/bo"
	"github.com/aide-family/marksman/internal/biz/repository"
)

const (
	defaultEvaluateInterval = time.Minute
	defaultQueryTimeout     = 15 * time.Second
	defaultStepSeconds      = 15
	// Keep a safe margin under Prometheus hard limit (11,000 points).
	maxQueryRangePoints = 10000
	// EvaluatorTypeMetric is the evaluator type for metric strategy; used on alert events and evaluator_snapshots.
	EvaluatorTypeMetric = "metric"
)

// alertTemplateData is the data passed to Go templates for Summary, Description and Labels.
type alertTemplateData struct {
	Strategy     *alertTemplateInfo
	OriginLabels map[string]string
	Labels       map[string]string
	Value        float64
	FiredAt      time.Time
}

type alertTemplateInfo struct {
	NamespaceUID      int64
	StrategyGroupUID  int64
	StrategyGroupName string
	StrategyUID       int64
	StrategyName      string
	LevelUID          int64
	LevelName         string
	DatasourceUID     int64
	DatasourceName    string
	Labels            map[string]string
	Threshold         []float64
	SampleMode        enum.SampleMode
	Condition         enum.ConditionMetric
	DurationSec       int64
}

// NewMetricEvaluator creates a cron job that evaluates the given metric strategy and sends alert events when conditions are met.
func NewMetricEvaluator(
	metricDatasourceQuerierRepo repository.MetricDatasourceQuerier,
	alertEventChannelRepo repository.AlertEventChannel,
	info *bo.EvaluateMetricStrategyBo,
) cron.CronJob {
	return &metricEvaluator{
		metricDatasourceQuerierRepo: metricDatasourceQuerierRepo,
		alertEventChannelRepo:       alertEventChannelRepo,
		info:                        info,
		cachedSnapshotJSON:          info.MarshalEvaluatorSnapshotJSON(),
	}
}

type metricEvaluator struct {
	metricDatasourceQuerierRepo repository.MetricDatasourceQuerier
	alertEventChannelRepo       repository.AlertEventChannel
	info                        *bo.EvaluateMetricStrategyBo
	cachedSnapshotJSON          string // pre-serialized evaluator snapshot, same for all events from this evaluator
}

// Index implements [cron.CronJob].
func (m *metricEvaluator) Index() string {
	if m.info == nil {
		return ""
	}
	return m.info.BuildMetricEvaluatorIndex()
}

// IsImmediate implements [cron.CronJob].
func (m *metricEvaluator) IsImmediate() bool {
	return false
}

// Run implements [cron.CronJob].
func (m *metricEvaluator) Run() {
	ctx, cancel := context.WithTimeout(context.Background(), defaultQueryTimeout)
	defer cancel()
	ctx = contextx.WithNamespace(ctx, m.info.GetNamespaceUID())

	end := time.Now()
	dur := time.Duration(m.info.GetDurationSec()) * time.Second
	if dur < 10*time.Second {
		dur = 10 * time.Second
	}
	start := end.Add(-dur * 2)
	queryRange := prometheusv1.Range{
		Start: start,
		End:   end,
		Step:  calculateMetricQueryStep(end.Sub(start)),
	}

	matrix, err := m.metricDatasourceQuerierRepo.QueryRange(ctx, m.info.GetDatasource(), m.info.GetExpr(), queryRange)
	if err != nil {
		klog.Errorw("msg", "metric evaluate query failed", "error", err, "strategyUID", m.info.GetStrategyUID())
		return
	}

	for _, series := range matrix {
		if len(series.Values) == 0 {
			continue
		}
		// 1) For each sample in the time window, decide if it satisfies ConditionMetric (value vs Values).
		satisfied := make([]bool, len(series.Values))
		for i, p := range series.Values {
			satisfied[i] = m.satisfiesCondition(float64(p.Value))
		}
		// 2) Apply SampleMode: FOR = n consecutive true, MAX = at most n times true, MIN = at least n times true.
		if !m.shouldFireBySampleMode(satisfied) {
			continue
		}
		// Emit one alert event per series (carry labels); use last sample value for event.
		lastVal := series.Values[len(series.Values)-1]
		tmlData := m.buildAlertTemplateData(series, float64(lastVal.Value), end)
		summary := m.fillStringTemplate(m.info.GetSummary(), tmlData)
		description := m.fillStringTemplate(m.info.GetDescription(), tmlData)

		ev := &bo.AlertEventBo{
			StrategyUID:           m.info.GetStrategyUID(),
			NamespaceUID:          m.info.GetNamespaceUID(),
			LevelUID:              m.info.GetLevelUID(),
			Summary:               summary,
			Description:           description,
			Expr:                  m.info.GetExpr(),
			FiredAt:               end,
			Value:                 float64(lastVal.Value),
			Labels:                tmlData.Labels,
			DatasourceUID:         m.info.GetDatasource().UID,
			EvaluatorType:         EvaluatorTypeMetric,
			EvaluatorSnapshotJSON: m.cachedSnapshotJSON,
			Fingerprint:           bo.BuildAlertFingerprint(m.Index(), tmlData.OriginLabels),
			EvaluateDuration:      dur,
			StrategyGroupUID:      m.info.GetStrategyGroupUID(),
			StrategyGroupName:     m.info.GetStrategyGroupName(),
			StrategyName:          m.info.GetStrategyName(),
			LevelName:             m.info.GetLevelName(),
			BgColor:               m.info.GetLevelBgColor(),
			DatasourceName:        m.info.GetDatasourceName(),
			DatasourceLevelName:   m.info.GetDatasourceLevelName(),
		}
		m.alertEventChannelRepo.Send(ev)
	}
}

// satisfiesCondition returns whether the metric value v satisfies ConditionMetric with strategy Values (thresholds).
func (m *metricEvaluator) satisfiesCondition(v float64) bool {
	vals := m.info.GetValues()
	cond := m.info.GetCondition()
	if len(vals) == 0 || cond == enum.ConditionMetric_CONDITION_METRIC_UNKNOWN {
		return false
	}
	threshold := vals[0]
	switch cond {
	case enum.ConditionMetric_CONDITION_METRIC_EQ:
		return v == threshold
	case enum.ConditionMetric_CONDITION_METRIC_NE:
		return v != threshold
	case enum.ConditionMetric_CONDITION_METRIC_GT:
		return v > threshold
	case enum.ConditionMetric_CONDITION_METRIC_GTE:
		return v >= threshold
	case enum.ConditionMetric_CONDITION_METRIC_LT:
		return v < threshold
	case enum.ConditionMetric_CONDITION_METRIC_LTE:
		return v <= threshold
	case enum.ConditionMetric_CONDITION_METRIC_IN:
		return len(vals) >= 2 && v >= vals[0] && v <= vals[1]
	case enum.ConditionMetric_CONDITION_METRIC_NOT_IN:
		return len(vals) < 2 || v < vals[0] || v > vals[1]
	default:
		return false
	}
}

// sampleModeN returns the "n" for SampleMode (FOR: consecutive count, MAX/MIN: count threshold). Values[0] is condition threshold; n is Values[1] when present.
func (m *metricEvaluator) sampleModeN() int {
	if values := m.info.GetValues(); len(values) > 1 {
		n := int(values[1])
		if n < 0 {
			n = 0
		}
		return n
	}
	return 0
}

// shouldFireBySampleMode decides whether to fire based on SampleMode within the time window:
// - FOR: "Occurs n times consecutively within m time" → fire if there are n consecutive samples satisfying the condition.
// - MAX: "Occurs at most n times within m time" → fire if condition holds in more than n samples.
// - MIN: "Occurs at least n times within m time" → fire if condition holds in at least n samples.
func (m *metricEvaluator) shouldFireBySampleMode(satisfied []bool) bool {
	n := m.sampleModeN()
	switch m.info.GetMode() {
	case enum.SampleMode_SAMPLE_MODE_FOR:
		// n consecutive times; if n not set, require at least 1
		required := n
		if required <= 0 {
			required = 1
		}
		return maxConsecutiveTrue(satisfied) >= required
	case enum.SampleMode_SAMPLE_MODE_MAX:
		// at most n times → fire when count > n
		count := countTrue(satisfied)
		return count > n
	case enum.SampleMode_SAMPLE_MODE_MIN:
		// at least n times → fire when count >= n; if n not set, require at least 1
		count := countTrue(satisfied)
		if n <= 0 {
			n = 1
		}
		return count >= n
	default:
		// unknown: treat as at least 1 time (same as FOR with n=1)
		return countTrue(satisfied) >= 1
	}
}

// buildAlertTemplateData builds template data from strategy info and the current series for alert templating.
func (m *metricEvaluator) buildAlertTemplateData(series *model.SampleStream, value float64, firedAt time.Time) *alertTemplateData {
	seriesLabels := make(map[string]string, len(series.Metric))
	for k, v := range series.Metric {
		seriesLabels[string(k)] = string(v)
	}
	info := &alertTemplateInfo{
		StrategyGroupUID:  m.info.GetStrategyGroupUID().Int64(),
		StrategyGroupName: m.info.GetStrategyGroupName(),
		StrategyUID:       m.info.GetStrategyUID().Int64(),
		StrategyName:      m.info.GetStrategyName(),
		LevelUID:          m.info.GetLevelUID().Int64(),
		LevelName:         m.info.GetLevelName(),
		DatasourceUID:     m.info.GetDatasourceUID().Int64(),
		DatasourceName:    m.info.GetDatasourceName(),
		NamespaceUID:      m.info.GetNamespaceUID().Int64(),
		Labels:            m.info.GetLabels(),
		Threshold:         m.info.GetValues(),
		SampleMode:        m.info.GetMode(),
		Condition:         m.info.GetCondition(),
		DurationSec:       m.info.GetDurationSec(),
	}
	return &alertTemplateData{
		Strategy:     info,
		OriginLabels: seriesLabels,
		Labels:       m.fillLabels(seriesLabels, info.Labels),
		Value:        value,
		FiredAt:      firedAt,
	}
}

// fillStringTemplate executes the template with data; on parse/execute error returns the original string.
func (m *metricEvaluator) fillStringTemplate(tmpl string, data any) string {
	if tmpl == "" {
		return ""
	}
	out, err := strutil.ExecuteTextTemplate(tmpl, data)
	if err != nil {
		klog.Debugw("msg", "alert template execute failed, use raw", "template", tmpl, "error", err)
		return tmpl
	}
	return out
}

// fillLabels merges series labels with strategy labels; each strategy label value is template-filled.
func (m *metricEvaluator) fillLabels(originLabels map[string]string, strategyLabels map[string]string) map[string]string {
	labels := maps.Clone(originLabels)
	for k, v := range strategyLabels {
		labels[k] = m.fillStringTemplate(v, originLabels)
	}
	return labels
}

func countTrue(b []bool) int {
	c := 0
	for _, v := range b {
		if v {
			c++
		}
	}
	return c
}

func maxConsecutiveTrue(b []bool) int {
	maxRun, cur := 0, 0
	for _, v := range b {
		if v {
			cur++
			if cur > maxRun {
				maxRun = cur
			}
		} else {
			cur = 0
		}
	}
	return maxRun
}

func calculateMetricQueryStep(window time.Duration) time.Duration {
	step := time.Duration(defaultStepSeconds) * time.Second
	if window <= 0 {
		return step
	}

	// Prometheus enforces a per-series point limit. Compute the minimum step that
	// keeps points in range and then keep our default when it is already large enough.
	maxIntervals := time.Duration(maxQueryRangePoints - 1)
	minStepByWindow := window / maxIntervals
	if window%maxIntervals != 0 {
		minStepByWindow++
	}
	if minStepByWindow > step {
		step = minStepByWindow
	}
	if step < time.Second {
		step = time.Second
	}
	return step
}

// Spec implements [cron.CronJob].
func (m *metricEvaluator) Spec() cron.CronSpec {
	interval := defaultEvaluateInterval
	if durationSec := m.info.GetDurationSec(); durationSec > 0 {
		interval = time.Duration(durationSec) * time.Second
	}
	return cron.CronSpecEvery(interval)
}
