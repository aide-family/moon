package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/moon/cmd/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/biz/repository"
)

// NewDashboard creates a new dashboard business logic.
func NewDashboardBiz(
	dashboardRepo repository.Dashboard,
	dashboardChartRepo repository.DashboardChart,
	transaction repository.Transaction,
	logger log.Logger,
) *Dashboard {
	return &Dashboard{
		dashboardRepo:      dashboardRepo,
		dashboardChartRepo: dashboardChartRepo,
		transaction:        transaction,
		helper:             log.NewHelper(log.With(logger, "module", "biz.dashboard")),
	}
}

// Dashboard is a dashboard business logic implementation.
type Dashboard struct {
	dashboardRepo      repository.Dashboard
	dashboardChartRepo repository.DashboardChart
	transaction        repository.Transaction
	helper             *log.Helper
}

// SaveDashboard saves a dashboard.
func (b *Dashboard) SaveDashboard(ctx context.Context, req *bo.SaveDashboardReq) error {
	if req.GetID() == 0 {
		return b.dashboardRepo.CreateDashboard(ctx, req)
	}
	return b.dashboardRepo.UpdateDashboard(ctx, req)
}

// DeleteDashboard deletes a dashboard.
func (b *Dashboard) DeleteDashboard(ctx context.Context, id uint32) error {
	return b.transaction.BizExec(ctx, func(ctx context.Context) error {
		if err := b.dashboardChartRepo.DeleteDashboardChartByDashboardID(ctx, id); err != nil {
			return err
		}
		return b.dashboardRepo.DeleteDashboard(ctx, id)
	})
}

// GetDashboard gets a dashboard.
func (b *Dashboard) GetDashboard(ctx context.Context, id uint32) (do.Dashboard, error) {
	return b.dashboardRepo.GetDashboard(ctx, id)
}

// ListDashboard lists dashboards.
func (b *Dashboard) ListDashboard(ctx context.Context, req *bo.ListDashboardReq) (*bo.ListDashboardReply, error) {
	return b.dashboardRepo.ListDashboards(ctx, req)
}

// SelectTeamDashboard selects team dashboards.
func (b *Dashboard) SelectTeamDashboard(ctx context.Context, req *bo.SelectTeamDashboardReq) (*bo.SelectTeamDashboardReply, error) {
	return b.dashboardRepo.SelectTeamDashboard(ctx, req)
}

// BatchUpdateDashboardStatus updates multiple dashboards' status.
func (b *Dashboard) BatchUpdateDashboardStatus(ctx context.Context, req *bo.BatchUpdateDashboardStatusReq) error {
	return b.dashboardRepo.BatchUpdateDashboardStatus(ctx, req)
}

// SaveDashboardChart saves a dashboard chart.
func (b *Dashboard) SaveDashboardChart(ctx context.Context, req *bo.SaveDashboardChartReq) error {
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
func (b *Dashboard) DeleteDashboardChart(ctx context.Context, req *bo.OperateOneDashboardChartReq) error {
	return b.dashboardChartRepo.DeleteDashboardChart(ctx, req)
}

// GetDashboardChart gets a dashboard chart.
func (b *Dashboard) GetDashboardChart(ctx context.Context, req *bo.OperateOneDashboardChartReq) (do.DashboardChart, error) {
	return b.dashboardChartRepo.GetDashboardChart(ctx, req)
}

// ListDashboardCharts lists dashboard charts.
func (b *Dashboard) ListDashboardCharts(ctx context.Context, req *bo.ListDashboardChartReq) (*bo.ListDashboardChartReply, error) {
	return b.dashboardChartRepo.ListDashboardCharts(ctx, req)
}

// SelectTeamDashboardChart selects team dashboard charts.
func (b *Dashboard) SelectTeamDashboardChart(ctx context.Context, req *bo.SelectTeamDashboardChartReq) (*bo.SelectTeamDashboardChartReply, error) {
	return b.dashboardChartRepo.SelectTeamDashboardChart(ctx, req)
}

// BatchUpdateDashboardChartStatus updates multiple dashboard charts' status.
func (b *Dashboard) BatchUpdateDashboardChartStatus(ctx context.Context, req *bo.BatchUpdateDashboardChartStatusReq) error {
	return b.dashboardChartRepo.BatchUpdateDashboardChartStatus(ctx, req)
}
