package repository

import (
	"context"

	"github.com/aide-family/moon/cmd/server/houyi/internal/biz/bo"
	"github.com/aide-family/moon/pkg/houyi/datasource/metric"
)

// Metric .
type Metric interface {
	// GetMetrics 获取指标列表
	GetMetrics(ctx context.Context, datasourceInfo *bo.GetMetricsParams) ([]*bo.MetricDetail, error)

	// Query 查询QL数据
	Query(ctx context.Context, req *bo.QueryQLParams) ([]*metric.QueryResponse, error)
}
