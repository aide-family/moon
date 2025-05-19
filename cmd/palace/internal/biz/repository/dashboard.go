package repository

import (
	"context"

	"github.com/aide-family/moon/cmd/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
)

type Dashboard interface {
	CreateDashboard(ctx context.Context, dashboard bo.Dashboard) error

	UpdateDashboard(ctx context.Context, dashboard bo.Dashboard) error

	// DeleteDashboard delete dashboard by id
	DeleteDashboard(ctx context.Context, id uint32) error

	// GetDashboard get dashboard by id
	GetDashboard(ctx context.Context, id uint32) (do.Dashboard, error)

	// ListDashboards list dashboards with filter
	ListDashboards(ctx context.Context, req *bo.ListDashboardReq) (*bo.ListDashboardReply, error)

	// SelectTeamDashboard select team dashboard
	SelectTeamDashboard(ctx context.Context, req *bo.SelectTeamDashboardReq) (*bo.SelectTeamDashboardReply, error)

	// BatchUpdateDashboardStatus update multiple dashboards status
	BatchUpdateDashboardStatus(ctx context.Context, req *bo.BatchUpdateDashboardStatusReq) error
}

type DashboardChart interface {
	CreateDashboardChart(ctx context.Context, chart bo.DashboardChart) error

	UpdateDashboardChart(ctx context.Context, chart bo.DashboardChart) error

	// DeleteDashboardChart delete dashboard chart by id
	DeleteDashboardChart(ctx context.Context, req *bo.OperateOneDashboardChartReq) error

	// GetDashboardChart get dashboard chart by id
	GetDashboardChart(ctx context.Context, req *bo.OperateOneDashboardChartReq) (do.DashboardChart, error)

	// ListDashboardCharts list dashboard charts with filter
	ListDashboardCharts(ctx context.Context, req *bo.ListDashboardChartReq) (*bo.ListDashboardChartReply, error)

	// SelectTeamDashboardChart select team dashboard chart
	SelectTeamDashboardChart(ctx context.Context, req *bo.SelectTeamDashboardChartReq) (*bo.SelectTeamDashboardChartReply, error)

	// BatchUpdateDashboardChartStatus update multiple dashboard charts status
	BatchUpdateDashboardChartStatus(ctx context.Context, req *bo.BatchUpdateDashboardChartStatusReq) error

	DeleteDashboardChartByDashboardID(ctx context.Context, dashboardID uint32) error
}
