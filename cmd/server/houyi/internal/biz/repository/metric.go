package repository

import (
	"context"

	"github.com/aide-cloud/moon/cmd/server/houyi/internal/biz/bo"
)

// Metric .
type Metric interface {
	// GetMetrics 获取指标列表
	GetMetrics(ctx context.Context, datasourceInfo *bo.GetMetricsParams) ([]*bo.MetricDetail, error)
}
