package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"prometheus-manager/app/prom_server/internal/biz/bo"
	"prometheus-manager/app/prom_server/internal/biz/do"
	"prometheus-manager/app/prom_server/internal/biz/do/basescopes"
	"prometheus-manager/app/prom_server/internal/biz/repository"
	"prometheus-manager/app/prom_server/internal/biz/vobj"
)

type (
	DashboardBiz struct {
		log *log.Helper

		dashboardRepo repository.DashboardRepo
		chartRepo     repository.ChartRepo
		logX          repository.SysLogRepo
	}
)

func NewDashboardBiz(
	dashboardRepo repository.DashboardRepo,
	chartRepo repository.ChartRepo,
	logX repository.SysLogRepo,
	logger log.Logger,
) *DashboardBiz {
	return &DashboardBiz{
		log:           log.NewHelper(log.With(logger, "module", "biz.dashboard")),
		dashboardRepo: dashboardRepo,
		chartRepo:     chartRepo,
		logX:          logX,
	}
}

// CreateChart 创建图表
func (l *DashboardBiz) CreateChart(ctx context.Context, chartInfo *bo.MyChartBO) (*bo.MyChartBO, error) {
	newChartDetail, err := l.chartRepo.Create(ctx, chartInfo)
	if err != nil {
		return nil, err
	}

	l.logX.CreateSysLog(ctx, vobj.ActionCreate, &bo.SysLogBo{
		ModuleName: vobj.ModuleDashboardChart,
		ModuleId:   newChartDetail.Id,
		Content:    newChartDetail.String(),
		Title:      "创建图表",
	})
	return newChartDetail, nil
}

// UpdateChartById 更新图表
func (l *DashboardBiz) UpdateChartById(ctx context.Context, chartId uint32, chartInfo *bo.MyChartBO) (*bo.MyChartBO, error) {
	// 查询
	chartInfoDetail, err := l.chartRepo.Get(ctx, basescopes.InIds(chartId))
	if err != nil {
		return nil, err
	}
	newData, err := l.chartRepo.Update(ctx, chartInfo, basescopes.InIds(chartId))
	if err != nil {
		return nil, err
	}
	l.logX.CreateSysLog(ctx, vobj.ActionUpdate, &bo.SysLogBo{
		ModuleName: vobj.ModuleDashboardChart,
		ModuleId:   newData.Id,
		Content:    bo.NewChangeLogBo(chartInfoDetail, newData).String(),
		Title:      "更新图表",
	})
	return newData, nil
}

// DeleteChartById 删除图表
func (l *DashboardBiz) DeleteChartById(ctx context.Context, chartId uint32) error {
	// 查询
	chartInfoDetail, err := l.chartRepo.Get(ctx, basescopes.InIds(chartId))
	if err != nil {
		return err
	}
	if err = l.chartRepo.Delete(ctx, basescopes.InIds(chartId)); err != nil {
		return err
	}
	l.logX.CreateSysLog(ctx, vobj.ActionDelete, &bo.SysLogBo{
		ModuleName: vobj.ModuleDashboardChart,
		ModuleId:   chartId,
		Content:    chartInfoDetail.String(),
		Title:      "删除图表",
	})
	return nil
}

// GetChartDetail 获取图表详情
func (l *DashboardBiz) GetChartDetail(ctx context.Context, chartId uint32) (*bo.MyChartBO, error) {
	return l.chartRepo.Get(ctx, basescopes.InIds(chartId))
}

// ListChartByPage 查询图表列表
func (l *DashboardBiz) ListChartByPage(ctx context.Context, pgInfo bo.Pagination, scopes ...basescopes.ScopeMethod) ([]*bo.MyChartBO, error) {
	return l.chartRepo.List(ctx, pgInfo, scopes...)
}

// CreateDashboard 创建dashboard
func (l *DashboardBiz) CreateDashboard(ctx context.Context, dashboardInfo *bo.MyDashboardConfigBO) (*bo.MyDashboardConfigBO, error) {
	newData, err := l.dashboardRepo.Create(ctx, dashboardInfo)
	if err != nil {
		return nil, err
	}
	l.logX.CreateSysLog(ctx, vobj.ActionCreate, &bo.SysLogBo{
		ModuleName: vobj.ModuleDashboard,
		ModuleId:   newData.Id,
		Content:    newData.String(),
		Title:      "创建dashboard",
	})
	return newData, nil
}

// UpdateDashboardById 更新dashboard
func (l *DashboardBiz) UpdateDashboardById(ctx context.Context, dashboardId uint32, dashboardInfo *bo.MyDashboardConfigBO) (*bo.MyDashboardConfigBO, error) {
	// 查询
	dashboardInfoDetail, err := l.dashboardRepo.Get(ctx, basescopes.InIds(dashboardId))
	if err != nil {
		return nil, err
	}
	newData, err := l.dashboardRepo.Update(ctx, dashboardInfo, basescopes.InIds(dashboardId))
	if err != nil {
		return nil, err
	}
	l.logX.CreateSysLog(ctx, vobj.ActionUpdate, &bo.SysLogBo{
		ModuleName: vobj.ModuleDashboard,
		ModuleId:   newData.Id,
		Content:    bo.NewChangeLogBo(dashboardInfoDetail, newData).String(),
		Title:      "更新dashboard",
	})
	return newData, nil
}

// DeleteDashboardById 删除dashboard
func (l *DashboardBiz) DeleteDashboardById(ctx context.Context, dashboardId uint32) error {
	// 查询
	dashboardInfoDetail, err := l.dashboardRepo.Get(ctx, basescopes.InIds(dashboardId))
	if err != nil {
		return err
	}
	if err = l.dashboardRepo.Delete(ctx, basescopes.InIds(dashboardId)); err != nil {
		return err
	}
	l.logX.CreateSysLog(ctx, vobj.ActionDelete, &bo.SysLogBo{
		ModuleName: vobj.ModuleDashboard,
		ModuleId:   dashboardId,
		Content:    dashboardInfoDetail.String(),
		Title:      "删除dashboard",
	})
	return nil
}

// GetDashboardById 获取dashboard详情
func (l *DashboardBiz) GetDashboardById(ctx context.Context, dashboardId uint32) (*bo.MyDashboardConfigBO, error) {
	return l.dashboardRepo.Get(ctx, basescopes.InIds(dashboardId), do.MyDashboardConfigPreloadCharts())
}

// ListDashboard 查询dashboard列表
func (l *DashboardBiz) ListDashboard(ctx context.Context, req *bo.ListDashboardReq) ([]*bo.MyDashboardConfigBO, error) {
	wheres := []basescopes.ScopeMethod{
		basescopes.TitleLike(req.Keyword),
		basescopes.StatusEQ(req.Status),
		basescopes.CreatedAtDesc(),
		basescopes.DeletedAtDesc(),
	}
	return l.dashboardRepo.List(ctx, req.Page, wheres...)
}
