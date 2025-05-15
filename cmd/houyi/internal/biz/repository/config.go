package repository

import (
	"context"

	"github.com/moon-monitor/moon/cmd/houyi/internal/biz/bo"
)

type Config interface {
	GetMetricDatasourceConfig(ctx context.Context, id string) (bo.MetricDatasourceConfig, bool)
	SetMetricDatasourceConfig(ctx context.Context, configs ...bo.MetricDatasourceConfig) error

	SetMetricRules(ctx context.Context, rules ...bo.MetricRule) error
	GetMetricRules(ctx context.Context) ([]bo.MetricRule, error)
	GetMetricRule(ctx context.Context, id string) (bo.MetricRule, bool)
	DeleteMetricRules(ctx context.Context, ids ...string) error
}
