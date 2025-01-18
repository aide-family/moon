package biz

import (
	"context"

	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/repository"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
)

// NewDashboardBiz 创建仪表盘业务
func NewDashboardBiz(dashboardRepository repository.Dashboard) *DashboardBiz {
	return &DashboardBiz{
		dashboardRepository: dashboardRepository,
	}
}

// DashboardBiz  仪表盘业务
type DashboardBiz struct {
	dashboardRepository repository.Dashboard
}

// CreateDashboard 创建仪表盘
func (b *DashboardBiz) CreateDashboard(ctx context.Context, req *bo.AddDashboardParams) error {
	return b.dashboardRepository.AddDashboard(ctx, req)
}

// UpdateDashboard 更新仪表盘
func (b *DashboardBiz) UpdateDashboard(ctx context.Context, req *bo.UpdateDashboardParams) error {
	return b.dashboardRepository.UpdateDashboard(ctx, req)
}

// DeleteDashboard 删除仪表盘
func (b *DashboardBiz) DeleteDashboard(ctx context.Context, req *bo.DeleteDashboardParams) error {
	return b.dashboardRepository.DeleteDashboard(ctx, req)
}

// GetDashboard 获取仪表盘
func (b *DashboardBiz) GetDashboard(ctx context.Context, id uint32) (*bizmodel.Dashboard, error) {
	return b.dashboardRepository.GetDashboard(ctx, id)
}

// ListDashboard 获取仪表盘列表
func (b *DashboardBiz) ListDashboard(ctx context.Context, params *bo.ListDashboardParams) ([]*bizmodel.Dashboard, error) {
	return b.dashboardRepository.ListDashboard(ctx, params)
}

// BatchUpdateDashboardStatus 批量更新仪表盘状态
func (b *DashboardBiz) BatchUpdateDashboardStatus(ctx context.Context, params *bo.BatchUpdateDashboardStatusParams) error {
	return b.dashboardRepository.BatchUpdateDashboardStatus(ctx, params)
}

// AddChart 添加图表
func (b *DashboardBiz) AddChart(ctx context.Context, params *bo.AddChartParams) error {
	return b.dashboardRepository.AddChart(ctx, params)
}

// UpdateChart 更新图表
func (b *DashboardBiz) UpdateChart(ctx context.Context, params *bo.UpdateChartParams) error {
	return b.dashboardRepository.UpdateChart(ctx, params)
}

// DeleteChart 删除图表
func (b *DashboardBiz) DeleteChart(ctx context.Context, params *bo.DeleteChartParams) error {
	return b.dashboardRepository.DeleteChart(ctx, params)
}

// ListChart 获取图表列表
func (b *DashboardBiz) ListChart(ctx context.Context, params *bo.ListChartParams) ([]*bizmodel.DashboardChart, error) {
	return b.dashboardRepository.ListChart(ctx, params)
}

// GetChart 获取图表
func (b *DashboardBiz) GetChart(ctx context.Context, params *bo.GetChartParams) (*bizmodel.DashboardChart, error) {
	return b.dashboardRepository.GetChart(ctx, params)
}

// BatchUpdateChartStatus 批量更新图表状态
func (b *DashboardBiz) BatchUpdateChartStatus(ctx context.Context, params *bo.BatchUpdateChartStatusParams) error {
	return b.dashboardRepository.BatchUpdateChartStatus(ctx, params)
}

// BatchUpdateChartSort 批量更新图表排序
func (b *DashboardBiz) BatchUpdateChartSort(ctx context.Context, dashboardID uint32, ids []uint32) error {
	return b.dashboardRepository.BatchUpdateChartSort(ctx, dashboardID, ids)
}
