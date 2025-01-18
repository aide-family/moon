package repository

import (
	"context"

	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
)

// Dashboard 仪表盘相关操作接口
type Dashboard interface {
	// AddDashboard 添加仪表盘
	AddDashboard(ctx context.Context, req *bo.AddDashboardParams) error

	// DeleteDashboard 删除仪表盘
	DeleteDashboard(ctx context.Context, req *bo.DeleteDashboardParams) error

	// UpdateDashboard 更新仪表盘
	UpdateDashboard(ctx context.Context, req *bo.UpdateDashboardParams) error

	// GetDashboard 获取仪表盘详情
	GetDashboard(ctx context.Context, id uint32) (*bizmodel.Dashboard, error)

	// ListDashboard 获取仪表盘列表
	ListDashboard(ctx context.Context, params *bo.ListDashboardParams) ([]*bizmodel.Dashboard, error)

	// BatchUpdateDashboardStatus 批量更新仪表盘状态
	BatchUpdateDashboardStatus(ctx context.Context, params *bo.BatchUpdateDashboardStatusParams) error

	// AddChart 添加图表
	AddChart(ctx context.Context, params *bo.AddChartParams) error

	// UpdateChart 更新图表
	UpdateChart(ctx context.Context, params *bo.UpdateChartParams) error

	// DeleteChart 删除图表
	DeleteChart(ctx context.Context, params *bo.DeleteChartParams) error

	// GetChart 获取图表
	GetChart(ctx context.Context, params *bo.GetChartParams) (*bizmodel.DashboardChart, error)

	// ListChart 获取图表列表
	ListChart(ctx context.Context, params *bo.ListChartParams) ([]*bizmodel.DashboardChart, error)

	// BatchUpdateChartStatus 批量更新图表状态
	BatchUpdateChartStatus(ctx context.Context, params *bo.BatchUpdateChartStatusParams) error
}
