package repository

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"prometheus-manager/app/prom_server/internal/biz/bo"
	"prometheus-manager/app/prom_server/internal/biz/do/basescopes"
)

var _ DashboardRepo = (*UnimplementedDashboardRepo)(nil)

type (
	DashboardRepo interface {
		unimplementedDashboardRepo()
		// Create 创建dashboard
		Create(ctx context.Context, dashboard *bo.MyDashboardConfigBO) (*bo.MyDashboardConfigBO, error)
		// Get 获取dashboard
		Get(ctx context.Context, scopes ...basescopes.ScopeMethod) (*bo.MyDashboardConfigBO, error)
		// Find 查询dashboard
		Find(ctx context.Context, scopes ...basescopes.ScopeMethod) ([]*bo.MyDashboardConfigBO, error)
		// List 获取dashboard列表
		List(ctx context.Context, pgInfo basescopes.Pagination, scopes ...basescopes.ScopeMethod) ([]*bo.MyDashboardConfigBO, error)
		// Delete 删除dashboard
		Delete(ctx context.Context, scopes ...basescopes.ScopeMethod) error
		// Update 更新dashboard
		Update(ctx context.Context, dashboard *bo.MyDashboardConfigBO, scopes ...basescopes.ScopeMethod) (*bo.MyDashboardConfigBO, error)
	}

	UnimplementedDashboardRepo struct{}
)

func (UnimplementedDashboardRepo) unimplementedDashboardRepo() {}

func (UnimplementedDashboardRepo) Create(_ context.Context, _ *bo.MyDashboardConfigBO) (*bo.MyDashboardConfigBO, error) {
	return nil, status.Error(codes.Unimplemented, "method Create not implemented")
}

func (UnimplementedDashboardRepo) Get(_ context.Context, _ ...basescopes.ScopeMethod) (*bo.MyDashboardConfigBO, error) {
	return nil, status.Error(codes.Unimplemented, "method Get not implemented")
}

func (UnimplementedDashboardRepo) Find(_ context.Context, _ ...basescopes.ScopeMethod) ([]*bo.MyDashboardConfigBO, error) {
	return nil, status.Error(codes.Unimplemented, "method Find not implemented")
}

func (UnimplementedDashboardRepo) List(_ context.Context, _ basescopes.Pagination, _ ...basescopes.ScopeMethod) ([]*bo.MyDashboardConfigBO, error) {
	return nil, status.Error(codes.Unimplemented, "method List not implemented")
}

func (UnimplementedDashboardRepo) Delete(_ context.Context, _ ...basescopes.ScopeMethod) error {
	return status.Error(codes.Unimplemented, "method Delete not implemented")
}

func (UnimplementedDashboardRepo) Update(_ context.Context, _ *bo.MyDashboardConfigBO, _ ...basescopes.ScopeMethod) (*bo.MyDashboardConfigBO, error) {
	return nil, status.Error(codes.Unimplemented, "method Update not implemented")
}
