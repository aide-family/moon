package repository

import (
	"context"
	"time"

	"github.com/aide-family/moon/cmd/laurel/internal/biz/bo"
	"github.com/aide-family/moon/cmd/laurel/internal/biz/vobj"
)

type Cache interface {
	Lock(ctx context.Context, key string, expiration time.Duration) (bool, error)
	Unlock(ctx context.Context, key string) error

	StorageMetric(ctx context.Context, metrics ...bo.MetricVec) error
	GetCounterMetrics(ctx context.Context, names ...string) ([]*bo.CounterMetricVec, error)
	GetGaugeMetrics(ctx context.Context, names ...string) ([]*bo.GaugeMetricVec, error)
	GetHistogramMetrics(ctx context.Context, names ...string) ([]*bo.HistogramMetricVec, error)
	GetSummaryMetrics(ctx context.Context, names ...string) ([]*bo.SummaryMetricVec, error)
	GetMetric(ctx context.Context, metricType vobj.MetricType, metricName string) (bo.MetricVec, error)
}
