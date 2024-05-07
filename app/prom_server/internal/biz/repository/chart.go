package repository

import (
	"context"

	"github.com/aide-family/moon/app/prom_server/internal/biz/bo"
	"github.com/aide-family/moon/app/prom_server/internal/biz/do"
	"github.com/aide-family/moon/app/prom_server/internal/biz/do/basescopes"
)

type (
	ChartRepo interface {
		// Create 创建图表
		Create(ctx context.Context, chart *do.MyChart) (*do.MyChart, error)
		// Get 获取图表
		Get(ctx context.Context, scopes ...basescopes.ScopeMethod) (*do.MyChart, error)
		// Find 查询图表
		Find(ctx context.Context, scopes ...basescopes.ScopeMethod) ([]*do.MyChart, error)
		// List 获取图表列表
		List(ctx context.Context, pgInfo bo.Pagination, scopes ...basescopes.ScopeMethod) ([]*do.MyChart, error)
		// Delete 删除图表
		Delete(ctx context.Context, scopes ...basescopes.ScopeMethod) error
		// Update 更新图表
		Update(ctx context.Context, chart *do.MyChart, scopes ...basescopes.ScopeMethod) (*do.MyChart, error)
	}
)
