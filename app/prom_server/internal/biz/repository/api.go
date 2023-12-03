package repository

import (
	"context"

	query "github.com/aide-cloud/gorm-normalize"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"prometheus-manager/app/prom_server/internal/biz/bo"
)

var _ ApiRepo = (*UnimplementedApiRepo)(nil)

type (
	ApiRepo interface {
		unimplementedApiRepo()
		// Create 创建api
		Create(ctx context.Context, apiBOList ...*bo.ApiBO) ([]*bo.ApiBO, error)
		// Get 获取api
		Get(ctx context.Context, scopes ...query.ScopeMethod) (*bo.ApiBO, error)
		// Find 查询api
		Find(ctx context.Context, scopes ...query.ScopeMethod) ([]*bo.ApiBO, error)
		// List 获取api列表
		List(ctx context.Context, pgInfo query.Pagination, scopes ...query.ScopeMethod) ([]*bo.ApiBO, error)
		// Delete 删除api
		Delete(ctx context.Context, scopes ...query.ScopeMethod) error
		// Update 更新api
		Update(ctx context.Context, apiBO *bo.ApiBO, scopes ...query.ScopeMethod) (*bo.ApiBO, error)
	}

	UnimplementedApiRepo struct{}
)

func (UnimplementedApiRepo) Find(_ context.Context, _ ...query.ScopeMethod) ([]*bo.ApiBO, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Find not implemented")
}

func (UnimplementedApiRepo) Create(_ context.Context, _ ...*bo.ApiBO) ([]*bo.ApiBO, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Create not implemented")
}

func (UnimplementedApiRepo) Get(_ context.Context, _ ...query.ScopeMethod) (*bo.ApiBO, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Get not implemented")
}

func (UnimplementedApiRepo) List(_ context.Context, _ query.Pagination, _ ...query.ScopeMethod) ([]*bo.ApiBO, error) {
	return nil, status.Errorf(codes.Unimplemented, "method List not implemented")
}

func (UnimplementedApiRepo) Delete(_ context.Context, _ ...query.ScopeMethod) error {
	return status.Errorf(codes.Unimplemented, "method Delete not implemented")
}

func (UnimplementedApiRepo) Update(_ context.Context, _ *bo.ApiBO, _ ...query.ScopeMethod) (*bo.ApiBO, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Update not implemented")
}

func (UnimplementedApiRepo) unimplementedApiRepo() {}
