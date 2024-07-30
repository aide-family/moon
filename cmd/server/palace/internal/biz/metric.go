package biz

import (
	"context"

	"github.com/aide-family/moon/api/merr"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/repository"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
)

// NewMetricBiz 创建指标业务
func NewMetricBiz(metricRepository repository.Metric, lock repository.Lock) *MetricBiz {
	return &MetricBiz{
		metricRepository: metricRepository,
		lock:             lock,
	}
}

// MetricBiz 指标业务
type MetricBiz struct {
	metricRepository repository.Metric
	lock             repository.Lock
}

// UpdateMetricByID 通过ID修改指标信息
func (b *MetricBiz) UpdateMetricByID(ctx context.Context, params *bo.UpdateMetricParams) error {
	if err := b.metricRepository.Update(ctx, params); !types.IsNil(err) {
		return merr.ErrorI18nSystemErr(ctx).WithCause(err)
	}
	return nil
}

// GetMetricByID 通过ID获取指标信息
func (b *MetricBiz) GetMetricByID(ctx context.Context, params *bo.GetMetricParams) (detail *bizmodel.DatasourceMetric, err error) {
	if params.WithRelation {
		detail, err = b.metricRepository.GetWithRelation(ctx, params.ID)
	} else {
		detail, err = b.metricRepository.Get(ctx, params.ID)
	}
	if !types.IsNil(err) {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, merr.ErrorI18nMetricNotFoundErr(ctx)
		}
		return nil, merr.ErrorI18nSystemErr(ctx).WithCause(err)
	}
	return
}

// ListMetric 获取指标列表
func (b *MetricBiz) ListMetric(ctx context.Context, params *bo.QueryMetricListParams) ([]*bizmodel.DatasourceMetric, error) {
	list, err := b.metricRepository.List(ctx, params)
	if !types.IsNil(err) {
		return nil, merr.ErrorI18nSystemErr(ctx).WithCause(err)
	}
	return list, nil
}

// SelectMetric 获取指标列表
func (b *MetricBiz) SelectMetric(ctx context.Context, params *bo.QueryMetricListParams) ([]*bo.SelectOptionBo, error) {
	list, err := b.metricRepository.Select(ctx, params)
	if !types.IsNil(err) {
		return nil, merr.ErrorI18nSystemErr(ctx).WithCause(err)
	}
	return types.SliceTo(list, func(item *bizmodel.DatasourceMetric) *bo.SelectOptionBo {
		return bo.NewDatasourceMetricOptionBuild(item).ToSelectOption()
	}), nil
}

// DeleteMetricByID 通过ID删除指标信息
func (b *MetricBiz) DeleteMetricByID(ctx context.Context, id uint32) error {
	if err := b.metricRepository.Delete(ctx, id); !types.IsNil(err) {
		return merr.ErrorI18nSystemErr(ctx).WithCause(err)
	}
	return nil
}

// GetMetricLabelCount 获取指标标签数量
func (b *MetricBiz) GetMetricLabelCount(ctx context.Context, metricID uint32) (uint32, error) {
	count, err := b.metricRepository.MetricLabelCount(ctx, metricID)
	if !types.IsNil(err) {
		return 0, merr.ErrorI18nSystemErr(ctx).WithCause(err)
	}
	return count, nil
}

// CreateMetric 创建指标信息
func (b *MetricBiz) CreateMetric(ctx context.Context, params *bo.CreateMetricParams) error {
	if params.Done {
		defer func() {
			if err := b.lock.UnLock(ctx, syncDatasourceMetaKey(params.DatasourceID)); !types.IsNil(err) {
				log.Errorw("unlock err", err)
			}
		}()
	}

	if err := b.metricRepository.CreateMetrics(ctx, params.TeamID, params.ToModel()); !types.IsNil(err) {
		return merr.ErrorI18nSystemErr(ctx).WithCause(err)
	}
	return nil
}
