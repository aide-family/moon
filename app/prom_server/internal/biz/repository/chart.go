package repository

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"prometheus-manager/app/prom_server/internal/biz/bo"
	"prometheus-manager/app/prom_server/internal/biz/do/basescopes"
)

var _ ChartRepo = (*UnimplementedChartRepo)(nil)

type (
	ChartRepo interface {
		unimplementedChartRepo()
		// Create 创建图表
		Create(ctx context.Context, chart *bo.MyChartBO) (*bo.MyChartBO, error)
		// Get 获取图表
		Get(ctx context.Context, scopes ...basescopes.ScopeMethod) (*bo.MyChartBO, error)
		// Find 查询图表
		Find(ctx context.Context, scopes ...basescopes.ScopeMethod) ([]*bo.MyChartBO, error)
		// List 获取图表列表
		List(ctx context.Context, pgInfo basescopes.Pagination, scopes ...basescopes.ScopeMethod) ([]*bo.MyChartBO, error)
		// Delete 删除图表
		Delete(ctx context.Context, scopes ...basescopes.ScopeMethod) error
		// Update 更新图表
		Update(ctx context.Context, chart *bo.MyChartBO, scopes ...basescopes.ScopeMethod) (*bo.MyChartBO, error)
	}

	UnimplementedChartRepo struct{}
)

func (UnimplementedChartRepo) unimplementedChartRepo() {}

func (UnimplementedChartRepo) Create(_ context.Context, _ *bo.MyChartBO) (*bo.MyChartBO, error) {
	return nil, status.Error(codes.Unimplemented, "method Create not implemented")
}

func (UnimplementedChartRepo) Get(_ context.Context, _ ...basescopes.ScopeMethod) (*bo.MyChartBO, error) {
	return nil, status.Error(codes.Unimplemented, "method Get not implemented")
}

func (UnimplementedChartRepo) Find(_ context.Context, _ ...basescopes.ScopeMethod) ([]*bo.MyChartBO, error) {
	return nil, status.Error(codes.Unimplemented, "method Find not implemented")
}

func (UnimplementedChartRepo) List(_ context.Context, _ basescopes.Pagination, _ ...basescopes.ScopeMethod) ([]*bo.MyChartBO, error) {
	return nil, status.Error(codes.Unimplemented, "method List not implemented")
}

func (UnimplementedChartRepo) Delete(_ context.Context, _ ...basescopes.ScopeMethod) error {
	return status.Error(codes.Unimplemented, "method Delete not implemented")
}

func (UnimplementedChartRepo) Update(_ context.Context, _ *bo.MyChartBO, _ ...basescopes.ScopeMethod) (*bo.MyChartBO, error) {
	return nil, status.Error(codes.Unimplemented, "method Update not implemented")
}
