package repository

import (
	"context"

	"github.com/aide-cloud/moon/pkg/helper/model/bizmodel"
)

// DatasourceMetric .
type DatasourceMetric interface {
	// CreateMetrics 创建指标
	CreateMetrics(ctx context.Context, metrics ...*bizmodel.DatasourceMetric) error
}
