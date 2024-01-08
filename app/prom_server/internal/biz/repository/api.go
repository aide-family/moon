package repository

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"prometheus-manager/app/prom_server/internal/biz/do/basescopes"

	"prometheus-manager/app/prom_server/internal/biz/bo"
)

var _ ApiRepo = (*UnimplementedApiRepo)(nil)

type (
	ApiRepo interface {
		unimplementedApiRepo()
		// Create 创建api
		Create(ctx context.Context, apiBOList ...*bo.ApiBO) ([]*bo.ApiBO, error)
		// Get 获取api
		Get(ctx context.Context, scopes ...basescopes.ScopeMethod) (*bo.ApiBO, error)
		// Find 查询api
		Find(ctx context.Context, scopes ...basescopes.ScopeMethod) ([]*bo.ApiBO, error)
		// List 获取api列表
		List(ctx context.Context, pgInfo basescopes.Pagination, scopes ...basescopes.ScopeMethod) ([]*bo.ApiBO, error)
		// Delete 删除api
		Delete(ctx context.Context, scopes ...basescopes.ScopeMethod) error
		// Update 更新api
		Update(ctx context.Context, apiBO *bo.ApiBO, scopes ...basescopes.ScopeMethod) (*bo.ApiBO, error)
		// UpdateAll 更新所有api
		UpdateAll(ctx context.Context, apiBO *bo.ApiBO, scopes ...basescopes.ScopeMethod) error
	}

	UnimplementedApiRepo struct{}
)

func (UnimplementedApiRepo) UpdateAll(_ context.Context, _ *bo.ApiBO, _ ...basescopes.ScopeMethod) error {
	return status.Error(codes.Unimplemented, "method UpdateAll not implemented")
}

func (UnimplementedApiRepo) Find(_ context.Context, _ ...basescopes.ScopeMethod) ([]*bo.ApiBO, error) {
	return nil, status.Error(codes.Unimplemented, "method Find not implemented")
}

func (UnimplementedApiRepo) Create(_ context.Context, _ ...*bo.ApiBO) ([]*bo.ApiBO, error) {
	return nil, status.Error(codes.Unimplemented, "method Create not implemented")
}

func (UnimplementedApiRepo) Get(_ context.Context, _ ...basescopes.ScopeMethod) (*bo.ApiBO, error) {
	return nil, status.Error(codes.Unimplemented, "method Get not implemented")
}

func (UnimplementedApiRepo) List(_ context.Context, _ basescopes.Pagination, _ ...basescopes.ScopeMethod) ([]*bo.ApiBO, error) {
	return nil, status.Error(codes.Unimplemented, "method List not implemented")
}

func (UnimplementedApiRepo) Delete(_ context.Context, _ ...basescopes.ScopeMethod) error {
	return status.Error(codes.Unimplemented, "method Delete not implemented")
}

func (UnimplementedApiRepo) Update(_ context.Context, _ *bo.ApiBO, _ ...basescopes.ScopeMethod) (*bo.ApiBO, error) {
	return nil, status.Error(codes.Unimplemented, "method Update not implemented")
}

func (UnimplementedApiRepo) unimplementedApiRepo() {}
