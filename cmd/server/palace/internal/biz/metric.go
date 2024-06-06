package biz

import (
	"context"

	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/repository"
	"github.com/aide-family/moon/pkg/helper/model/bizmodel"
	"github.com/aide-family/moon/pkg/types"
)

func NewMetricBiz(metricRepository repository.Metric) *MetricBiz {
	return &MetricBiz{
		metricRepository: metricRepository,
	}
}

// MetricBiz 指标业务
type MetricBiz struct {
	metricRepository repository.Metric
}

// UpdateMetricByID 通过ID修改指标信息
func (b *MetricBiz) UpdateMetricByID(ctx context.Context, params *bo.UpdateMetricParams) error {
	return b.metricRepository.Update(ctx, params)
}

// GetMetricByID 通过ID获取指标信息
func (b *MetricBiz) GetMetricByID(ctx context.Context, params *bo.GetMetricParams) (*bizmodel.DatasourceMetric, error) {
	if params.WithRelation {
		return b.metricRepository.GetWithRelation(ctx, params.ID)
	}
	return b.metricRepository.Get(ctx, params.ID)
}

// ListMetric 获取指标列表
func (b *MetricBiz) ListMetric(ctx context.Context, params *bo.QueryMetricListParams) ([]*bizmodel.DatasourceMetric, error) {
	return b.metricRepository.List(ctx, params)
}

// SelectMetric 获取指标列表
func (b *MetricBiz) SelectMetric(ctx context.Context, params *bo.QueryMetricListParams) ([]*bo.SelectOptionBo, error) {
	list, err := b.metricRepository.Select(ctx, params)
	if err != nil {
		return nil, err
	}
	return types.SliceTo(list, func(item *bizmodel.DatasourceMetric) *bo.SelectOptionBo {
		return bo.NewDatasourceMetricOptionBuild(item).ToSelectOption()
	}), nil
}

// DeleteMetricByID 通过ID删除指标信息
func (b *MetricBiz) DeleteMetricByID(ctx context.Context, id uint32) error {
	return b.metricRepository.Delete(ctx, id)
}

// GetMetricLabelCount 获取指标标签数量
func (b *MetricBiz) GetMetricLabelCount(ctx context.Context, metricId uint32) (uint32, error) {
	return b.metricRepository.MetricLabelCount(ctx, metricId)
}
