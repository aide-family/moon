package repoimpl

import (
	"context"

	"github.com/aide-family/moon/api/merr"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/repository"
	"github.com/aide-family/moon/cmd/server/palace/internal/data"
	"github.com/aide-family/moon/cmd/server/palace/internal/service/build"
	"github.com/aide-family/moon/pkg/palace/model"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel/bizquery"
	"github.com/aide-family/moon/pkg/util/types"

	"github.com/go-kratos/kratos/v2/errors"
	"gorm.io/gen"
	"gorm.io/gorm"
)

// NewDashboardRepository 创建仪表盘操作实现
func NewDashboardRepository(data *data.Data) repository.Dashboard {
	return &dashboardRepositoryImpl{data: data}
}

type dashboardRepositoryImpl struct {
	data *data.Data
}

func (d *dashboardRepositoryImpl) AddDashboard(ctx context.Context, req *bo.AddDashboardParams) error {
	dashboardModuleBuilder := build.NewBuilder().DashboardModule().WithBoAddDashboardParams(req)
	dashboardModel := dashboardModuleBuilder.ToDo()
	bizQuery, err := getBizQuery(ctx, d.data)
	if err != nil {
		return err
	}
	return bizQuery.Transaction(func(tx *bizquery.Query) error {
		if err = tx.Dashboard.WithContext(ctx).Create(dashboardModel); err != nil {
			return err
		}
		strategyGroups := dashboardModuleBuilder.WithDashboardID(dashboardModel.ID).ToDoStrategyGroups()
		if err = tx.Dashboard.StrategyGroups.Model(dashboardModel).Append(strategyGroups...); err != nil {
			return err
		}
		charts := dashboardModuleBuilder.WithDashboardID(dashboardModel.ID).ToDoCharts()
		if err = tx.Dashboard.Charts.Model(dashboardModel).Append(charts...); err != nil {
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
		AllFieldModel: model.AllFieldModel{ID: req.ID},
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
	dashboardModuleBuilder := build.NewBuilder().DashboardModule().WithBoUpdateDashboardParams(req)
	dashboardModel := dashboardModuleBuilder.ToDo()
	strategyGroups := dashboardModuleBuilder.ToDoStrategyGroups()
	charts := dashboardModuleBuilder.ToDoCharts()
	bizQuery, err := getBizQuery(ctx, d.data)
	if err != nil {
		return err
	}

	return bizQuery.Transaction(func(tx *bizquery.Query) error {
		// 替换仪表盘图表
		if err = tx.Dashboard.Charts.Model(dashboardModel).Replace(charts...); err != nil {
			return err
		}
		// 替换策略组
		if err = tx.Dashboard.StrategyGroups.Model(dashboardModel).Replace(strategyGroups...); err != nil {
			return err
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

func (d *dashboardRepositoryImpl) GetDashboard(ctx context.Context, id uint32) (*bizmodel.Dashboard, error) {
	bizQuery, err := getBizQuery(ctx, d.data)
	if err != nil {
		return nil, err
	}
	detail, err := bizQuery.Dashboard.WithContext(ctx).Where(bizQuery.Dashboard.ID.Eq(id)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, merr.ErrorI18nDashboardDataNotFoundErr(ctx).WithCause(err)
		}
		return nil, merr.ErrorI18nSystemErr(ctx).WithCause(err)
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
	if err = types.WithPageQuery[bizquery.IDashboardDo](dashboardQuery, params.Page); err != nil {
		return nil, err
	}
	return dashboardQuery.Find()
}
