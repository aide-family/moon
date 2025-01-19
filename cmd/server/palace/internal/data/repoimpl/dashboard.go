package repoimpl

import (
	"context"
	"fmt"
	"strings"

	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/repository"
	"github.com/aide-family/moon/cmd/server/palace/internal/data"
	"github.com/aide-family/moon/pkg/helper/middleware"
	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/palace/model"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel/bizquery"
	"github.com/aide-family/moon/pkg/util/types"

	"github.com/go-kratos/kratos/v2/errors"
	"gorm.io/gen"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// NewDashboardRepository 创建仪表盘操作实现
func NewDashboardRepository(data *data.Data) repository.Dashboard {
	return &dashboardRepositoryImpl{data: data}
}

type dashboardRepositoryImpl struct {
	data *data.Data
}

// AddChart implements repository.Dashboard.
func (d *dashboardRepositoryImpl) AddChart(ctx context.Context, params *bo.AddChartParams) error {
	bizQuery, err := getBizQuery(ctx, d.data)
	if err != nil {
		return err
	}
	chartModel := params.ChartItem.ToModel(ctx)
	count, err := bizQuery.DashboardChart.WithContext(ctx).Where(bizQuery.DashboardChart.DashboardID.Eq(params.DashboardID)).Count()
	if err != nil {
		return err
	}
	chartModel.Sort = uint32(count)
	return bizQuery.DashboardChart.WithContext(ctx).Create(chartModel)
}

// DeleteChart implements repository.Dashboard.
func (d *dashboardRepositoryImpl) DeleteChart(ctx context.Context, params *bo.DeleteChartParams) error {
	bizQuery, err := getBizQuery(ctx, d.data)
	if err != nil {
		return err
	}
	_, err = bizQuery.DashboardChart.WithContext(ctx).Where(
		bizQuery.DashboardChart.ID.Eq(params.ChartID),
		bizQuery.DashboardChart.DashboardID.Eq(params.DashboardID),
	).Delete()
	return err
}

// GetChart implements repository.Dashboard.
func (d *dashboardRepositoryImpl) GetChart(ctx context.Context, params *bo.GetChartParams) (*bizmodel.DashboardChart, error) {
	bizQuery, err := getBizQuery(ctx, d.data)
	if err != nil {
		return nil, err
	}
	chartModel, err := bizQuery.DashboardChart.WithContext(ctx).Where(
		bizQuery.DashboardChart.ID.Eq(params.ChartID),
		bizQuery.DashboardChart.DashboardID.Eq(params.DashboardID),
	).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, merr.ErrorI18nToastDashboardChartNotFound(ctx).WithCause(err)
		}
		return nil, merr.ErrorI18nNotificationSystemError(ctx).WithCause(err)
	}
	return chartModel, nil
}

// ListChart implements repository.Dashboard.
func (d *dashboardRepositoryImpl) ListChart(ctx context.Context, params *bo.ListChartParams) ([]*bizmodel.DashboardChart, error) {
	bizQuery, err := getBizQuery(ctx, d.data)
	if err != nil {
		return nil, err
	}
	wheres := make([]gen.Condition, 0, 4)
	wheres = append(wheres, bizQuery.DashboardChart.DashboardID.Eq(params.DashboardID))
	if !types.TextIsNull(params.Keyword) {
		wheres = append(wheres, bizQuery.DashboardChart.Name.Like(params.Keyword))
	}
	if !params.Status.IsUnknown() {
		wheres = append(wheres, bizQuery.DashboardChart.Status.Eq(params.Status.GetValue()))
	}
	chartQuery := bizQuery.DashboardChart.WithContext(ctx)
	if len(wheres) > 0 {
		chartQuery = chartQuery.Where(wheres...)
	}
	if chartQuery, err = types.WithPageQuery(chartQuery, params.Page); err != nil {
		return nil, err
	}
	return chartQuery.Find()
}

// BatchUpdateChartStatus implements repository.Dashboard.
func (d *dashboardRepositoryImpl) BatchUpdateChartStatus(ctx context.Context, params *bo.BatchUpdateChartStatusParams) error {
	if len(params.ChartIDs) == 0 || params.Status.IsUnknown() {
		return nil
	}
	bizQuery, err := getBizQuery(ctx, d.data)
	if err != nil {
		return err
	}
	_, err = bizQuery.DashboardChart.WithContext(ctx).Where(
		bizQuery.DashboardChart.ID.In(params.ChartIDs...),
		bizQuery.DashboardChart.DashboardID.Eq(params.DashboardID),
	).UpdateSimple(bizQuery.DashboardChart.Status.Value(params.Status.GetValue()))
	return err
}

// UpdateChart implements repository.Dashboard.
func (d *dashboardRepositoryImpl) UpdateChart(ctx context.Context, params *bo.UpdateChartParams) error {
	bizQuery, err := getBizQuery(ctx, d.data)
	if err != nil {
		return err
	}
	chartModel := params.ChartItem.ToModel(ctx)
	_, err = bizQuery.DashboardChart.WithContext(ctx).Where(
		bizQuery.DashboardChart.ID.Eq(params.ChartItem.ID),
		bizQuery.DashboardChart.DashboardID.Eq(params.DashboardID),
	).UpdateSimple(
		bizQuery.DashboardChart.Name.Value(chartModel.Name),
		bizQuery.DashboardChart.Status.Value(chartModel.Status.GetValue()),
		bizQuery.DashboardChart.Remark.Value(chartModel.Remark),
		bizQuery.DashboardChart.Height.Value(chartModel.Height),
		bizQuery.DashboardChart.Width.Value(chartModel.Width),
		bizQuery.DashboardChart.URL.Value(chartModel.URL),
	)
	return err
}

func (d *dashboardRepositoryImpl) BatchUpdateDashboardStatus(ctx context.Context, params *bo.BatchUpdateDashboardStatusParams) error {
	if len(params.IDs) == 0 {
		return nil
	}
	bizQuery, err := getBizQuery(ctx, d.data)
	if err != nil {
		return err
	}
	_, err = bizQuery.WithContext(ctx).
		Dashboard.Where(bizQuery.Dashboard.ID.In(params.IDs...)).
		Update(bizQuery.Dashboard.Status, params.Status)
	return err
}

func (d *dashboardRepositoryImpl) BatchUpdateChartSort(ctx context.Context, dashboardID uint32, ids []uint32) error {
	bizQuery, err := getBizQuery(ctx, d.data)
	if err != nil {
		return err
	}

	// 使用 strings.Builder 构建 CASE 语句
	var caseStmt strings.Builder
	caseStmt.WriteString("CASE")
	for index, id := range ids {
		caseStmt.WriteString(fmt.Sprintf(" WHEN id = %d THEN %d", id, index))
	}
	caseStmt.WriteString(" END")

	// 执行批量更新
	_, err = bizQuery.DashboardChart.WithContext(ctx).
		Where(bizQuery.DashboardChart.DashboardID.Eq(dashboardID), bizQuery.DashboardChart.ID.In(ids...)).
		Update(bizQuery.DashboardChart.Sort, gorm.Expr(caseStmt.String()))

	return err
}

func (d *dashboardRepositoryImpl) ListSelfDashboard(ctx context.Context) ([]*bizmodel.DashboardSelf, error) {
	bizQuery, err := getBizQuery(ctx, d.data)
	if err != nil {
		return nil, err
	}
	dashboardSelf := bizQuery.DashboardSelf
	return dashboardSelf.WithContext(ctx).
		Where(dashboardSelf.UserID.Eq(middleware.GetUserID(ctx))).
		Preload(dashboardSelf.Dashboard).
		Order(dashboardSelf.Sort.Asc(), dashboardSelf.ID.Asc()).Find()
}

func (d *dashboardRepositoryImpl) UpdateSelfDashboard(ctx context.Context, ids []uint32) error {
	bizQuery, err := getBizQuery(ctx, d.data)
	if err != nil {
		return err
	}

	if len(ids) == 0 {
		_, err = bizQuery.DashboardSelf.WithContext(ctx).Where(bizQuery.DashboardSelf.UserID.Eq(middleware.GetUserID(ctx))).Delete()
		return err
	}
	list := make([]*bizmodel.DashboardSelf, 0, len(ids))
	for index, id := range ids {
		if id == 0 {
			continue
		}
		list = append(list, &bizmodel.DashboardSelf{
			AllFieldModel: bizmodel.AllFieldModel{},
			DashboardID:   id,
			UserID:        middleware.GetUserID(ctx),
			MemberID:      middleware.GetTeamMemberID(ctx),
			Sort:          uint32(index),
		})
	}
	dashboardSelf := bizQuery.DashboardSelf
	return dashboardSelf.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: dashboardSelf.DashboardID.ColumnName().String()}, {Name: dashboardSelf.UserID.ColumnName().String()}},
		DoUpdates: clause.AssignmentColumns([]string{"sort"}),
	}).Create(list...)
}

func (d *dashboardRepositoryImpl) AddDashboard(ctx context.Context, req *bo.AddDashboardParams) error {
	dashboardModel := req.ToModel(ctx)
	bizQuery, err := getBizQuery(ctx, d.data)
	if err != nil {
		return err
	}
	return bizQuery.Transaction(func(tx *bizquery.Query) error {
		if err = tx.Dashboard.WithContext(ctx).Create(dashboardModel); err != nil {
			return err
		}
		strategyGroups := req.GetStrategyGroupDos()
		if err = tx.Dashboard.StrategyGroups.Model(dashboardModel).Append(strategyGroups...); err != nil {
			return err
		}
		return nil
	})
}

func (d *dashboardRepositoryImpl) DeleteDashboard(ctx context.Context, req *bo.DeleteDashboardParams) error {
	bizQuery, err := getBizQuery(ctx, d.data)
	if err != nil {
		return err
	}
	dashboardModel := &bizmodel.Dashboard{
		AllFieldModel: bizmodel.AllFieldModel{AllFieldModel: model.AllFieldModel{ID: req.ID}},
	}
	return bizQuery.Transaction(func(tx *bizquery.Query) error {
		if err = tx.Dashboard.Charts.Model(dashboardModel).Clear(); err != nil {
			return err
		}
		if err = tx.Dashboard.StrategyGroups.Model(dashboardModel).Clear(); err != nil {
			return err
		}
		_, err = tx.Dashboard.WithContext(ctx).Where(bizQuery.Dashboard.ID.Eq(req.ID), bizQuery.Dashboard.Status.Eq(req.Status.GetValue())).Delete()
		return err
	})
}

func (d *dashboardRepositoryImpl) UpdateDashboard(ctx context.Context, req *bo.UpdateDashboardParams) error {
	dashboardModel := req.ToModel(ctx)
	strategyGroups := req.Dashboard.GetStrategyGroupDos()

	bizQuery, err := getBizQuery(ctx, d.data)
	if err != nil {
		return err
	}

	return bizQuery.Transaction(func(tx *bizquery.Query) error {
		if len(strategyGroups) > 0 {
			// 替换策略组
			if err = tx.Dashboard.StrategyGroups.Model(dashboardModel).Replace(strategyGroups...); err != nil {
				return err
			}
		} else {
			// 删除策略组
			if err = tx.Dashboard.StrategyGroups.Model(dashboardModel).Clear(); err != nil {
				return err
			}
		}

		_, err = tx.Dashboard.WithContext(ctx).
			Where(bizQuery.Dashboard.ID.Eq(req.ID)).
			UpdateSimple(
				tx.Dashboard.Color.Value(dashboardModel.Color),
				tx.Dashboard.Name.Value(dashboardModel.Name),
				tx.Dashboard.Remark.Value(dashboardModel.Remark),
				tx.Dashboard.Status.Value(dashboardModel.Status.GetValue()),
			)
		return err
	})
}

func (d *dashboardRepositoryImpl) GetDashboard(ctx context.Context, params *bo.GetDashboardParams) (*bizmodel.Dashboard, error) {
	bizQuery, err := getBizQuery(ctx, d.data)
	if err != nil {
		return nil, err
	}
	dashboardQuery := bizQuery.Dashboard.WithContext(ctx).Where(bizQuery.Dashboard.ID.Eq(params.ID))
	if params.Charts {
		dashboardQuery = dashboardQuery.Preload(bizQuery.Dashboard.Charts)
	}
	detail, err := dashboardQuery.First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, merr.ErrorI18nToastDashboardNotFound(ctx).WithCause(err)
		}
		return nil, merr.ErrorI18nNotificationSystemError(ctx).WithCause(err)
	}
	if params.MyDashboard {
		detail.Charts = types.SliceToWithFilter(detail.Charts, func(chart *bizmodel.DashboardChart) (*bizmodel.DashboardChart, bool) {
			return chart, chart.Status.IsEnable()
		})
	}
	return detail, nil
}

func (d *dashboardRepositoryImpl) ListDashboard(ctx context.Context, params *bo.ListDashboardParams) ([]*bizmodel.Dashboard, error) {
	bizQuery, err := getBizQuery(ctx, d.data)
	if err != nil {
		return nil, err
	}
	wheres := make([]gen.Condition, 0, 2)
	if !types.TextIsNull(params.Keyword) {
		wheres = append(wheres, bizQuery.Dashboard.Name.Like(params.Keyword))
	}
	if !params.Status.IsUnknown() {
		wheres = append(wheres, bizQuery.Dashboard.Status.Eq(params.Status.GetValue()))
	}
	dashboardQuery := bizQuery.Dashboard.WithContext(ctx).Where(wheres...)
	if dashboardQuery, err = types.WithPageQuery(dashboardQuery, params.Page); err != nil {
		return nil, err
	}
	return dashboardQuery.Find()
}
