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
