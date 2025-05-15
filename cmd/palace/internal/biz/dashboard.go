package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/moon/cmd/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/biz/repository"
)

// DashboardBiz is a dashboard business logic implementation.
type DashboardBiz struct {
	dashboardRepo      repository.Dashboard
	dashboardChartRepo repository.DashboardChart
	transaction        repository.Transaction
	log                *log.Helper
}

// NewDashboardBiz creates a new dashboard business logic.
func NewDashboardBiz(
	dashboardRepo repository.Dashboard,
	dashboardChartRepo repository.DashboardChart,
	transaction repository.Transaction,
	logger log.Logger,
) *DashboardBiz {
	return &DashboardBiz{
		dashboardRepo:      dashboardRepo,
		dashboardChartRepo: dashboardChartRepo,
		transaction:        transaction,
		log:                log.NewHelper(log.With(logger, "module", "biz.dashboard")),
	}
}

// SaveDashboard saves a dashboard.
func (b *DashboardBiz) SaveDashboard(ctx context.Context, req *bo.SaveDashboardReq) error {
	if req.GetID() == 0 {
		return b.dashboardRepo.CreateDashboard(ctx, req)
	}
	return b.dashboardRepo.UpdateDashboard(ctx, req)
}

// DeleteDashboard deletes a dashboard.
func (b *DashboardBiz) DeleteDashboard(ctx context.Context, id uint32) error {
	return b.transaction.BizExec(ctx, func(ctx context.Context) error {
		if err := b.dashboardChartRepo.DeleteDashboardChartByDashboardID(ctx, id); err != nil {
			return err
		}
		return b.dashboardRepo.DeleteDashboard(ctx, id)
	})
}

// GetDashboard gets a dashboard.
func (b *DashboardBiz) GetDashboard(ctx context.Context, id uint32) (do.Dashboard, error) {
	return b.dashboardRepo.GetDashboard(ctx, id)
}

// ListDashboard lists dashboards.
func (b *DashboardBiz) ListDashboard(ctx context.Context, req *bo.ListDashboardReq) (*bo.ListDashboardReply, error) {
	return b.dashboardRepo.ListDashboards(ctx, req)
}

// BatchUpdateDashboardStatus updates multiple dashboards' status.
func (b *DashboardBiz) BatchUpdateDashboardStatus(ctx context.Context, req *bo.BatchUpdateDashboardStatusReq) error {
	return b.dashboardRepo.BatchUpdateDashboardStatus(ctx, req)
}

// SaveDashboardChart saves a dashboard chart.
func (b *DashboardBiz) SaveDashboardChart(ctx context.Context, req *bo.SaveDashboardChartReq) error {
	if req.GetID() <= 0 {
		return b.dashboardChartRepo.CreateDashboardChart(ctx, req)
	}

	_, err := b.dashboardRepo.GetDashboard(ctx, req.GetDashboardID())
	if err != nil {
		return err
	}

	return b.dashboardChartRepo.UpdateDashboardChart(ctx, req)
}

// DeleteDashboardChart deletes a dashboard chart.
func (b *DashboardBiz) DeleteDashboardChart(ctx context.Context, req *bo.OperateOneDashboardChartReq) error {
	return b.dashboardChartRepo.DeleteDashboardChart(ctx, req)
}

// GetDashboardChart gets a dashboard chart.
func (b *DashboardBiz) GetDashboardChart(ctx context.Context, req *bo.OperateOneDashboardChartReq) (do.DashboardChart, error) {
	return b.dashboardChartRepo.GetDashboardChart(ctx, req)
}

// ListDashboardCharts lists dashboard charts.
func (b *DashboardBiz) ListDashboardCharts(ctx context.Context, req *bo.ListDashboardChartReq) (*bo.ListDashboardChartReply, error) {
	return b.dashboardChartRepo.ListDashboardCharts(ctx, req)
}

// BatchUpdateDashboardChartStatus updates multiple dashboard charts' status.
func (b *DashboardBiz) BatchUpdateDashboardChartStatus(ctx context.Context, req *bo.BatchUpdateDashboardChartStatusReq) error {
	return b.dashboardChartRepo.BatchUpdateDashboardChartStatus(ctx, req)
}
