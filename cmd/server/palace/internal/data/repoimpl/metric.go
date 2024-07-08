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
)

func NewMetricRepository(data *data.Data) repository.Metric {
	return &metricRepositoryImpl{data: data}
}

type metricRepositoryImpl struct {
	data *data.Data
}

func (m *metricRepositoryImpl) MetricLabelCount(ctx context.Context, id uint32) (uint32, error) {
	teamQuery, err := getBizDB(ctx, m.data)
	if err != nil {
		return 0, err
	}
	total, err := teamQuery.MetricLabel.
		WithContext(ctx).
		Where(teamQuery.MetricLabel.MetricID.Eq(id)).Count()
	return uint32(total), err
}

func (m *metricRepositoryImpl) Update(ctx context.Context, params *bo.UpdateMetricParams) error {
	teamQuery, err := getBizDB(ctx, m.data)
	if err != nil {
		return err
	}
	_, err = teamQuery.MetricLabel.
		WithContext(ctx).
		Where(teamQuery.MetricLabel.MetricID.Eq(params.ID)).
		UpdateColumnSimple(
			teamQuery.DatasourceMetric.Unit.Value(params.Unit),
			teamQuery.DatasourceMetric.Remark.Value(params.Remark),
		)
	return err
}

func (m *metricRepositoryImpl) Get(ctx context.Context, id uint32) (*bizmodel.DatasourceMetric, error) {
	teamQuery, err := getBizDB(ctx, m.data)
	if err != nil {
		return nil, err
	}
	return teamQuery.DatasourceMetric.
		WithContext(ctx).
		Where(teamQuery.DatasourceMetric.ID.Eq(id)).
		First()
}

func (m *metricRepositoryImpl) GetWithRelation(ctx context.Context, id uint32) (*bizmodel.DatasourceMetric, error) {
	teamQuery, err := getBizDB(ctx, m.data)
	if err != nil {
		return nil, err
	}
	return teamQuery.DatasourceMetric.
		WithContext(ctx).
		Where(teamQuery.DatasourceMetric.ID.Eq(id)).
		Preload(teamQuery.DatasourceMetric.Labels.LabelValues).
		First()
}

func (m *metricRepositoryImpl) Delete(ctx context.Context, id uint32) error {
	teamQuery, err := getBizDB(ctx, m.data)
	if err != nil {
		return err
	}

	var labelIds []uint32
	// 查询所有label id
	err = teamQuery.MetricLabel.
		WithContext(ctx).
		Where(teamQuery.MetricLabel.MetricID.Eq(id)).
		Select(teamQuery.MetricLabel.ID).
		Scan(&labelIds)
	if err != nil {
		return err
	}
	// 查询所有的label value ids
	var labelValueIds []uint32
	if len(labelIds) > 0 {
		err = teamQuery.MetricLabelValue.
			WithContext(ctx).
			Where(teamQuery.MetricLabelValue.LabelID.In(labelIds...)).
			Select(teamQuery.MetricLabelValue.ID).
			Scan(&labelValueIds)
		if err != nil {
			return err
		}
	}

	metric := &bizmodel.DatasourceMetric{
		BaseModel: model.BaseModel{ID: id},
	}
	return teamQuery.Transaction(func(tx *bizquery.Query) error {
		cnt, err := tx.DatasourceMetric.WithContext(ctx).
			Select(field.AssociationFields).
			Delete(metric)
		if err != nil {
			return err
		}

		if cnt.RowsAffected == 0 {
			return nil
		}

		// 删除关联数据
		_, err = tx.MetricLabelValue.WithContext(ctx).Where(tx.MetricLabelValue.ID.In(labelValueIds...)).Delete()
		return err
	})
}

func (m *metricRepositoryImpl) List(ctx context.Context, params *bo.QueryMetricListParams) ([]*bizmodel.DatasourceMetric, error) {
	q, err := getBizDB(ctx, m.data)
	if err != nil {
		return nil, err
	}
	qq := q.DatasourceMetric.WithContext(ctx)
	var wheres []gen.Condition
	if !types.TextIsNull(params.Keyword) {
		wheres = append(wheres, q.DatasourceMetric.Name.Like(params.Keyword))
	}
	if params.DatasourceID > 0 {
		wheres = append(wheres, q.DatasourceMetric.DatasourceID.Eq(params.DatasourceID))
	}
	if !params.MetricType.IsUnknown() {
		wheres = append(wheres, q.DatasourceMetric.Category.Eq(params.MetricType.GetValue()))
	}
	qq = qq.Where(wheres...)
	if err := types.WithPageQuery[bizquery.IDatasourceMetricDo](qq, params.Page); err != nil {
		return nil, err
	}
	return qq.Order(bizquery.Datasource.ID.Desc()).Find()
}

func (m *metricRepositoryImpl) Select(ctx context.Context, params *bo.QueryMetricListParams) ([]*bizmodel.DatasourceMetric, error) {
	q, err := getBizDB(ctx, m.data)
	if err != nil {
		return nil, err
	}
	qq := q.DatasourceMetric.WithContext(ctx)
	qq.Select(q.DatasourceMetric.ID, q.DatasourceMetric.Name, q.DatasourceMetric.Unit, q.DatasourceMetric.DeletedAt)
	var wheres []gen.Condition
	if !types.TextIsNull(params.Keyword) {
		wheres = append(wheres, q.DatasourceMetric.Name.Like(params.Keyword))
	}
	if params.DatasourceID > 0 {
		wheres = append(wheres, q.DatasourceMetric.DatasourceID.Eq(params.DatasourceID))
	}
	if !params.MetricType.IsUnknown() {
		wheres = append(wheres, q.DatasourceMetric.Category.Eq(params.MetricType.GetValue()))
	}
	qq = qq.Where(wheres...)
	if !types.IsNil(params.Page) {
		page := params.Page
		total, err := qq.Count()
		if !types.IsNil(err) {
			return nil, err
		}
		params.Page.SetTotal(int(total))
		pageNum, pageSize := page.GetPageNum(), page.GetPageSize()
		if pageNum <= 1 {
			qq = qq.Limit(pageSize)
		} else {
			qq = qq.Offset((pageNum - 1) * pageSize).Limit(pageSize)
		}
	}
	return qq.Find()
}
