package biz

import (
	"context"

	"github.com/aide-cloud/moon/cmd/server/houyi/internal/biz/bo"
	"github.com/aide-cloud/moon/cmd/server/houyi/internal/biz/repository"
)

func NewMetricBiz(metricRepository repository.Metric) *MetricBiz {
	return &MetricBiz{
		metricRepository: metricRepository,
	}
}

// MetricBiz .
type MetricBiz struct {
	metricRepository repository.Metric
}

// SyncMetrics 同步数据源元数据
func (b *MetricBiz) SyncMetrics(ctx context.Context, datasourceInfo *bo.GetMetricsParams) ([]*bo.MetricDetail, error) {
	return b.metricRepository.GetMetrics(ctx, datasourceInfo)
}
