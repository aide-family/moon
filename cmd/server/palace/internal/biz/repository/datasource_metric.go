package repository

import (
	"context"

	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
)

// DatasourceMetric 数据源指标接口
type DatasourceMetric interface {
	// CreateMetrics 创建指标
	CreateMetrics(context.Context, ...*bizmodel.DatasourceMetric) error

	// CreateMetricsNoAuth 创建指标(不鉴权)
	CreateMetricsNoAuth(context.Context, uint32, ...*bizmodel.DatasourceMetric) error
}
