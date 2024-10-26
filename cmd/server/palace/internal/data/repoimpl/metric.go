package repoimpl

import (
	"context"

	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/repository"
	"github.com/aide-family/moon/cmd/server/palace/internal/data"
	"github.com/aide-family/moon/pkg/palace/model"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel/bizquery"
	"github.com/aide-family/moon/pkg/util/types"
	"gorm.io/gen"
	"gorm.io/gen/field"
	"gorm.io/gorm/clause"
)

// NewMetricRepository 创建指标仓库
func NewMetricRepository(data *data.Data) repository.Metric {
	return &metricRepositoryImpl{data: data}
}

type metricRepositoryImpl struct {
	data *data.Data
}

func (m *metricRepositoryImpl) MetricLabelCount(ctx context.Context, id uint32) (uint32, error) {
	bizQuery, err := getBizQuery(ctx, m.data)
	if err != nil {
		return 0, err
	}
	total, err := bizQuery.MetricLabel.
		WithContext(ctx).
		Where(bizQuery.MetricLabel.MetricID.Eq(id)).Count()
	return uint32(total), err
}

func (m *metricRepositoryImpl) Update(ctx context.Context, params *bo.UpdateMetricParams) error {
	bizQuery, err := getBizQuery(ctx, m.data)
	if err != nil {
		return err
	}
	_, err = bizQuery.MetricLabel.
		WithContext(ctx).
		Where(bizQuery.MetricLabel.MetricID.Eq(params.ID)).
		UpdateColumnSimple(
			bizQuery.DatasourceMetric.Unit.Value(params.Unit),
			bizQuery.DatasourceMetric.Remark.Value(params.Remark),
		)
	return err
}

func (m *metricRepositoryImpl) Get(ctx context.Context, id uint32) (*bizmodel.DatasourceMetric, error) {
	bizQuery, err := getBizQuery(ctx, m.data)
	if err != nil {
		return nil, err
	}
	return bizQuery.DatasourceMetric.
		WithContext(ctx).
		Where(bizQuery.DatasourceMetric.ID.Eq(id)).
		Preload(field.Associations).
		First()
}

func (m *metricRepositoryImpl) GetWithRelation(ctx context.Context, id uint32) (*bizmodel.DatasourceMetric, error) {
	bizQuery, err := getBizQuery(ctx, m.data)
	if err != nil {
		return nil, err
	}
	return bizQuery.DatasourceMetric.
		WithContext(ctx).
		Where(bizQuery.DatasourceMetric.ID.Eq(id)).
		Preload(bizQuery.DatasourceMetric.Labels).
		First()
}

func (m *metricRepositoryImpl) Delete(ctx context.Context, id uint32) error {
	bizQuery, err := getBizQuery(ctx, m.data)
	if err != nil {
		return err
	}

	var labelIds []uint32
	// 查询所有label id
	err = bizQuery.MetricLabel.
		WithContext(ctx).
		Where(bizQuery.MetricLabel.MetricID.Eq(id)).
		Select(bizQuery.MetricLabel.ID).
		Scan(&labelIds)
	if err != nil {
		return err
	}

	metric := &bizmodel.DatasourceMetric{
		AllFieldModel: model.AllFieldModel{ID: id},
	}
	return bizQuery.Transaction(func(tx *bizquery.Query) error {
		cnt, err := tx.DatasourceMetric.WithContext(ctx).
			Select(field.AssociationFields).
			Delete(metric)
		if err != nil {
			return err
		}

		if cnt.RowsAffected == 0 {
			return nil
		}

		return err
	})
}

func (m *metricRepositoryImpl) List(ctx context.Context, params *bo.QueryMetricListParams) ([]*bizmodel.DatasourceMetric, error) {
	bizQuery, err := getBizQuery(ctx, m.data)
	if err != nil {
		return nil, err
	}
	metricQuery := bizQuery.DatasourceMetric.WithContext(ctx)
	var wheres []gen.Condition
	if !types.TextIsNull(params.Keyword) {
		wheres = append(wheres, bizQuery.DatasourceMetric.Name.Like(params.Keyword))
	}
	if params.DatasourceID > 0 {
		wheres = append(wheres, bizQuery.DatasourceMetric.DatasourceID.Eq(params.DatasourceID))
	}
	if !params.MetricType.IsUnknown() {
		wheres = append(wheres, bizQuery.DatasourceMetric.Category.Eq(params.MetricType.GetValue()))
	}
	metricQuery = metricQuery.Where(wheres...)
	if metricQuery, err = types.WithPageQuery(metricQuery, params.Page); err != nil {
		return nil, err
	}
	return metricQuery.Order(bizQuery.DatasourceMetric.ID.Desc()).Find()
}

func (m *metricRepositoryImpl) Select(ctx context.Context, params *bo.QueryMetricListParams) ([]*bizmodel.DatasourceMetric, error) {
	bizQuery, err := getBizQuery(ctx, m.data)
	if err != nil {
		return nil, err
	}
	metricQuery := bizQuery.DatasourceMetric.WithContext(ctx)
	metricQuery.Select(bizQuery.DatasourceMetric.ID, bizQuery.DatasourceMetric.Name, bizQuery.DatasourceMetric.Unit, bizQuery.DatasourceMetric.DeletedAt)
	var wheres []gen.Condition
	if !types.TextIsNull(params.Keyword) {
		wheres = append(wheres, bizQuery.DatasourceMetric.Name.Like(params.Keyword))
	}
	if params.DatasourceID > 0 {
		wheres = append(wheres, bizQuery.DatasourceMetric.DatasourceID.Eq(params.DatasourceID))
	}
	if !params.MetricType.IsUnknown() {
		wheres = append(wheres, bizQuery.DatasourceMetric.Category.Eq(params.MetricType.GetValue()))
	}
	metricQuery = metricQuery.Where(wheres...)
	if metricQuery, err = types.WithPageQuery(metricQuery, params.Page); err != nil {
		return nil, err
	}
	return metricQuery.Find()
}

func (m *metricRepositoryImpl) CreateMetrics(ctx context.Context, teamID uint32, metric *bizmodel.DatasourceMetric) error {
	// 根据指标名称查询指标
	bizDB, err := m.data.GetBizGormDB(teamID)
	if !types.IsNil(err) {
		return err
	}
	bizQuery := bizquery.Use(bizDB)

	return bizQuery.DatasourceMetric.WithContext(ctx).Clauses(
		clause.OnConflict{DoNothing: true},
	).Create(metric)
}
