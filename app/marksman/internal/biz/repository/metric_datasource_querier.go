package repository

import (
	"context"

	"github.com/aide-family/marksman/internal/biz/bo"
)

// MetricDatasourceQuerier provides metric list (no labels) and single-metric label detail.
type MetricDatasourceQuerier interface {
	// ListMetrics returns metric names with basic metadata (name, description, unit, type), no label detail.
	ListMetrics(ctx context.Context, ds *bo.DatasourceItemBo) ([]*bo.MetricSummaryItemBo, error)
	// GetMetricDetail returns one metric's full metadata including labels and label values.
	GetMetricDetail(ctx context.Context, ds *bo.DatasourceItemBo, metric string) (*bo.MetricDetailItemBo, error)
}
