// Package evaluator provides a metric evaluator.
package evaluator

import (
	"github.com/aide-family/magicbox/server/cron"
	"github.com/aide-family/marksman/internal/biz/bo"
)

func NewMetricEvaluator(info *bo.EvaluateMetricStrategyBo) cron.CronJob {
	return &metricEvaluator{info: info}
}

type metricEvaluator struct {
	info *bo.EvaluateMetricStrategyBo
}

// Index implements [cron.CronJob].
func (m *metricEvaluator) Index() string {
	panic("unimplemented")
}

// IsImmediate implements [cron.CronJob].
func (m *metricEvaluator) IsImmediate() bool {
	panic("unimplemented")
}

// Run implements [cron.CronJob].
func (m *metricEvaluator) Run() {
	panic("unimplemented")
}

// Spec implements [cron.CronJob].
func (m *metricEvaluator) Spec() cron.CronSpec {
	panic("unimplemented")
}
