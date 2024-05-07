package repository

import (
	"context"

	"github.com/aide-family/moon/app/prom_server/internal/biz/bo"
	"github.com/aide-family/moon/app/prom_server/internal/biz/do"
	"github.com/aide-family/moon/app/prom_server/internal/biz/do/basescopes"
)

type (
	DashboardRepo interface {
		// Create 创建dashboard
		Create(ctx context.Context, dashboard *bo.CreateMyDashboardBO) (*do.MyDashboardConfig, error)
		// Get 获取dashboard
		Get(ctx context.Context, scopes ...basescopes.ScopeMethod) (*do.MyDashboardConfig, error)
		// Find 查询dashboard
		Find(ctx context.Context, scopes ...basescopes.ScopeMethod) ([]*do.MyDashboardConfig, error)
		// List 获取dashboard列表
		List(ctx context.Context, pgInfo bo.Pagination, scopes ...basescopes.ScopeMethod) ([]*do.MyDashboardConfig, error)
		// Delete 删除dashboard
		Delete(ctx context.Context, scopes ...basescopes.ScopeMethod) error
		// Update 更新dashboard
		Update(ctx context.Context, dashboard *bo.UpdateMyDashboardBO, scopes ...basescopes.ScopeMethod) (*do.MyDashboardConfig, error)
	}
)
