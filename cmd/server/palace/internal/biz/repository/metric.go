package repository

import (
	"context"

	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
)

// Metric .
type Metric interface {
	// Get 查询指标详情
	Get(context.Context, uint32) (*bizmodel.DatasourceMetric, error)

	// GetWithRelation 查询指标详情(关联其他属性)
	GetWithRelation(context.Context, uint32) (*bizmodel.DatasourceMetric, error)

	// Delete 删除指标
	Delete(context.Context, uint32) error

	// List 查询指标列表
	List(context.Context, *bo.QueryMetricListParams) ([]*bizmodel.DatasourceMetric, error)

	// Select 查询指标列表(不关联其他属性)
	Select(context.Context, *bo.QueryMetricListParams) ([]*bizmodel.DatasourceMetric, error)

	// Update 更新指标
	Update(context.Context, *bo.UpdateMetricParams) error

	// MetricLabelCount 指标标签数量
	MetricLabelCount(context.Context, uint32) (uint32, error)

	// CreateMetrics 创建指标
	CreateMetrics(context.Context, uint32, *bizmodel.DatasourceMetric) error
}
