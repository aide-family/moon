// Package evaluator provides a metric evaluator.
package evaluator

import (
	"context"
	"fmt"
	"time"

	"github.com/aide-family/magicbox/contextx"
	"github.com/aide-family/magicbox/enum"
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
)

// NewMetricEvaluator creates a cron job that evaluates the given metric strategy and sends alert events when conditions are met.
func NewMetricEvaluator(
	querier repository.MetricDatasourceQuerier,
	alertCh repository.AlertEventChannel,
	info *bo.EvaluateMetricStrategyBo,
) cron.CronJob {
	return &metricEvaluator{
		querier: querier,
		alertCh: alertCh,
		info:    info,
	}
}

type metricEvaluator struct {
	querier repository.MetricDatasourceQuerier
	alertCh repository.AlertEventChannel
	info    *bo.EvaluateMetricStrategyBo
}

// Index implements [cron.CronJob].
func (m *metricEvaluator) Index() string {
	levelUID := int64(0)
	if m.info.Level != nil {
		levelUID = m.info.Level.UID.Int64()
	}
	return fmt.Sprintf("metric-%d-%d-%d-%d", m.info.NamespaceUID.Int64(), m.info.StrategyUID.Int64(), levelUID, m.info.Datasource.UID.Int64())
}

// IsImmediate implements [cron.CronJob].
func (m *metricEvaluator) IsImmediate() bool {
	return false
}

// Run implements [cron.CronJob].
func (m *metricEvaluator) Run() {
	ctx, cancel := context.WithTimeout(context.Background(), defaultQueryTimeout)
	defer cancel()
	ctx = contextx.WithNamespace(ctx, m.info.NamespaceUID)

	end := time.Now()
	dur := m.info.DurationSec
	if dur < 60 {
		dur = 60
	}
	start := end.Add(-time.Duration(dur*2) * time.Second)
	queryRange := prometheusv1.Range{
		Start: start,
		End:   end,
		Step:  time.Duration(defaultStepSeconds) * time.Second,
	}

	matrix, err := m.querier.QueryRange(ctx, m.info.Datasource, m.info.Expr, queryRange)
	if err != nil {
		klog.Errorw("msg", "metric evaluate query failed", "error", err, "strategyUID", m.info.StrategyUID.Int64(), "expr", m.info.Expr)
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
		labels := make(map[string]string)
		for k, v := range series.Metric {
			labels[string(k)] = string(v)
		}
		ev := &bo.AlertEventBo{
			StrategyUID:   m.info.StrategyUID,
			NamespaceUID:  m.info.NamespaceUID,
			Level:         m.info.Level,
			Summary:       m.info.Summary,
			Description:   m.info.Description,
			Expr:          m.info.Expr,
			FiredAt:       end,
			Value:         float64(lastVal.Value),
			Labels:        labels,
			DatasourceUID: m.info.Datasource.UID,
		}
		m.alertCh.Send(ev)
	}
}

// satisfiesCondition returns whether the metric value v satisfies ConditionMetric with strategy Values (thresholds).
func (m *metricEvaluator) satisfiesCondition(v float64) bool {
	vals := m.info.Values
	cond := m.info.Condition
	if len(vals) == 0 && cond != enum.ConditionMetric_CONDITION_METRIC_UNKNOWN {
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
	if len(m.info.Values) > 1 {
		n := int(m.info.Values[1])
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
	switch m.info.Mode {
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

// Spec implements [cron.CronJob].
func (m *metricEvaluator) Spec() cron.CronSpec {
	interval := defaultEvaluateInterval
	if m.info.DurationSec > 0 {
		interval = time.Duration(m.info.DurationSec) * time.Second
	}
	return cron.CronSpecEvery(interval)
}
