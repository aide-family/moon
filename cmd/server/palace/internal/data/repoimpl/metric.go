package repoimpl

import (
	"context"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/repository"
	"github.com/aide-family/moon/cmd/server/palace/internal/data"
	"github.com/aide-family/moon/pkg/helper/middleware"
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
		AllFieldModel: bizmodel.AllFieldModel{AllFieldModel: model.AllFieldModel{ID: id}},
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

func (m *metricRepositoryImpl) CreateMetrics(ctx context.Context, params *bo.CreateMetricParams) error {
	teamID := params.TeamID
	ctx = middleware.WithTeamIDContextKey(ctx, teamID)
	metric := createDatasourceMetricParamToModel(ctx, params)
	metric.WithContext(ctx)
	// 根据指标名称查询指标
	bizDB, err := m.data.GetBizGormDB(teamID)
	if !types.IsNil(err) {
		return err
	}
	bizQuery := bizquery.Use(bizDB)

	// metric 所更新的字段
	metricCol := []string{
		bizQuery.DatasourceMetric.DatasourceID.ColumnName().String(),
		bizQuery.DatasourceMetric.Name.ColumnName().String(),
		bizQuery.DatasourceMetric.Category.ColumnName().String(),
		bizQuery.DatasourceMetric.Unit.ColumnName().String(),
		bizQuery.DatasourceMetric.Remark.ColumnName().String(),
		bizQuery.DatasourceMetric.LabelCount.ColumnName().String(),
		bizQuery.DatasourceMetric.TeamID.ColumnName().String(),
		bizQuery.DatasourceMetric.DeletedAt.ColumnName().String(),
	}

	// metric 更新条件
	metricWrapper := []clause.Column{
		{Name: bizQuery.DatasourceMetric.DatasourceID.ColumnName().String()},
		{Name: bizQuery.DatasourceMetric.Name.ColumnName().String()},
		{Name: bizQuery.DatasourceMetric.Category.ColumnName().String()},
		{Name: bizQuery.DatasourceMetric.DeletedAt.ColumnName().String()},
	}

	// label 所更新的字段
	labelCol := []string{
		bizQuery.MetricLabel.Name.ColumnName().String(),
		bizQuery.MetricLabel.LabelValues.ColumnName().String(),
		bizQuery.MetricLabel.MetricID.ColumnName().String(),
		bizQuery.MetricLabel.TeamID.ColumnName().String(),
		bizQuery.MetricLabel.Remark.ColumnName().String(),
	}

	// label 更新条件
	labelWrapper := []clause.Column{
		{Name: bizQuery.MetricLabel.Name.ColumnName().String()},
		{Name: bizQuery.MetricLabel.MetricID.ColumnName().String()},
		{Name: bizQuery.MetricLabel.DeletedAt.ColumnName().String()},
	}

	return bizQuery.Transaction(func(tx *bizquery.Query) error {
		if err := bizQuery.DatasourceMetric.WithContext(ctx).Clauses(
			clause.OnConflict{
				Columns:   metricWrapper,
				DoUpdates: clause.AssignmentColumns(metricCol)},
		).Create(metric); !types.IsNil(err) {
			return err
		}

		metricID := metric.ID
		// select db labels
		datasourceMetric, err := m.GetWithRelation(ctx, metricID)
		if types.IsNotNil(err) {
			return err
		}

		if types.IsNotNil(datasourceMetric) && types.IsNotNil(datasourceMetric.Labels) {
			labelsMap := types.ToMap(datasourceMetric.Labels, func(item *bizmodel.MetricLabel) string {
				return item.Name
			})
			params.Metric.MapLabels = labelsMap
		}

		labels := createMetricLabelParamToModels(ctx, params, metricID)

		if _, err := bizQuery.DatasourceMetric.WithContext(ctx).Where(bizQuery.DatasourceMetric.ID.Eq(metricID)).
			UpdateColumn(bizQuery.DatasourceMetric.LabelCount, len(labels)); types.IsNotNil(err) {
			return err
		}

		return bizQuery.MetricLabel.WithContext(ctx).Clauses(clause.OnConflict{Columns: labelWrapper, DoUpdates: clause.AssignmentColumns(labelCol)}).Create(labels...)
	})
}

func createDatasourceMetricParamToModel(ctx context.Context, params *bo.CreateMetricParams) *bizmodel.DatasourceMetric {
	if types.IsNil(params) || types.IsNil(params.Metric) {
		return nil
	}

	datasourceMetric := &bizmodel.DatasourceMetric{
		Name:         params.Metric.Name,
		Category:     params.Metric.Type,
		Unit:         params.Metric.Unit,
		Remark:       params.Metric.Help,
		DatasourceID: params.DatasourceID,
	}
	datasourceMetric.WithContext(ctx)
	return datasourceMetric
}

func createMetricLabelParamToModels(ctx context.Context, params *bo.CreateMetricParams, metricId uint32) []*bizmodel.MetricLabel {
	if types.IsNil(params) || types.IsNil(params.Metric) {
		return nil
	}
	return types.SliceTo(params.Metric.Labels, func(label *bo.MetricLabel) *bizmodel.MetricLabel {
		var values []string
		mapLabels := params.Metric.GetMapLabels()
		if types.IsNotNil(mapLabels[label.Name].GetLabelValues()) {
			values = types.MergeSliceWithUnique(label.Values, mapLabels[label.Name].GetLabelValues())
		} else {
			values = label.Values
		}
		bs, _ := types.Marshal(values)
		metricLabel := &bizmodel.MetricLabel{
			MetricID:    metricId,
			Name:        label.Name,
			LabelValues: string(bs),
		}
		metricLabel.WithContext(ctx)
		return metricLabel
	})
}
