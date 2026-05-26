// Package evaluator provides a metric evaluator.
package evaluator

import (
	"context"
	"time"

	"github.com/aide-family/magicbox/contextx"
	"github.com/aide-family/magicbox/server/cron"
	klog "github.com/go-kratos/kratos/v2/log"
	prometheusv1 "github.com/prometheus/client_golang/api/prometheus/v1"

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

	rule := buildPrometheusAlertRule(m.info)
	if rule == nil || rule.Expr == "" {
		klog.Errorw("msg", "metric evaluate rule is invalid", "strategyUID", m.info.GetStrategyUID())
		return
	}

	end := time.Now()
	evalInterval := m.evaluateInterval()
	firingSeries, err := m.evaluatePrometheusRule(ctx, rule, evalInterval, end)
	if err != nil {
		ds := m.info.GetDatasource()
		klog.Errorw("msg", "metric evaluate query failed",
			"error", err,
			"strategyUID", m.info.GetStrategyUID(),
			"expr", rule.Expr,
			"datasourceUID", ds.UID,
			"datasourceName", ds.Name,
			"datasourceURL", ds.URL,
			"datasourceDriver", ds.Driver,
		)
		return
	}

	for _, series := range firingSeries {
		ruleLabels := expandRuleLabels(rule.Labels, series.labels)
		alertLabels := mergePrometheusAlertLabels(series.labels, ruleLabels)
		summary := executePrometheusTemplate(rule.Annotations[annotationSummary], alertLabels, series.value)
		description := executePrometheusTemplate(rule.Annotations[annotationDescription], alertLabels, series.value)

		ev := &bo.AlertEventBo{
			StrategyUID:           m.info.GetStrategyUID(),
			NamespaceUID:          m.info.GetNamespaceUID(),
			LevelUID:              m.info.GetLevelUID(),
			Summary:               summary,
			Description:           description,
			Expr:                  rule.Expr,
			FiredAt:               end,
			Value:                 series.value,
			Labels:                alertLabels,
			DatasourceUID:         m.info.GetDatasource().UID,
			EvaluatorType:         EvaluatorTypeMetric,
			EvaluatorSnapshotJSON: m.cachedSnapshotJSON,
			Fingerprint:           prometheusAlertFingerprint(alertLabels),
			EvaluateDuration:      m.evaluateDurationForEvent(rule),
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

func (m *metricEvaluator) evaluatePrometheusRule(
	ctx context.Context,
	rule *prometheusAlertRule,
	evalInterval time.Duration,
	end time.Time,
) ([]firingSeries, error) {
	step := evalInterval
	if step <= 0 {
		step = defaultEvaluateInterval
	}
	if step < time.Second {
		step = time.Second
	}

	window := rule.For
	if window <= 0 {
		window = step
	}
	start := end.Add(-window - step)
	queryRange := prometheusv1.Range{
		Start: start,
		End:   end,
		Step:  calculateMetricQueryStep(end.Sub(start), step),
	}

	matrix, err := m.metricDatasourceQuerierRepo.QueryRange(ctx, m.info.GetDatasource(), rule.Expr, queryRange)
	if err != nil {
		return nil, err
	}
	return collectFiringSeries(matrix, rule.For, step, end), nil
}

func calculateMetricQueryStep(window, preferredStep time.Duration) time.Duration {
	step := preferredStep
	if step <= 0 {
		step = time.Duration(defaultStepSeconds) * time.Second
	}
	if window <= 0 {
		return step
	}

	// Prometheus enforces a per-series point limit. Compute the minimum step that
	// keeps points in range and then keep our preferred step when it is already large enough.
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
	return cron.CronSpecEvery(m.evaluateInterval())
}

func (m *metricEvaluator) evaluateInterval() time.Duration {
	interval := defaultEvaluateInterval
	if durationSec := m.info.GetDurationSec(); durationSec > 0 {
		interval = time.Duration(durationSec) * time.Second
	}
	return interval
}

func (m *metricEvaluator) evaluateDurationForEvent(rule *prometheusAlertRule) time.Duration {
	if rule.For > 0 {
		return rule.For
	}
	return time.Duration(m.info.GetDurationSec()) * time.Second
}
