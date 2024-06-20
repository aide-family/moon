package biz

import (
	"context"

	"github.com/aide-family/moon/cmd/server/houyi/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/houyi/internal/biz/repository"
	"github.com/aide-family/moon/pkg/houyi/datasource/metric"
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

// Query 查询数据
func (b *MetricBiz) Query(ctx context.Context, req *bo.QueryQLParams) ([]*metric.QueryResponse, error) {
	return b.metricRepository.Query(ctx, req)
}
