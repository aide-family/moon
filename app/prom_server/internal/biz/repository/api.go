package repository

import (
	"context"

	query "github.com/aide-cloud/gorm-normalize"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"prometheus-manager/app/prom_server/internal/biz/dobo"
)

var _ ApiRepo = (*UnimplementedApiRepo)(nil)

type (
	ApiRepo interface {
		unimplementedApiRepo()
		// Create 创建api
		Create(ctx context.Context, apiDoList ...*dobo.ApiDO) ([]*dobo.ApiDO, error)
		// Get 获取api
		Get(ctx context.Context, scopes ...query.ScopeMethod) (*dobo.ApiDO, error)
		// List 获取api列表
		List(ctx context.Context, pgInfo query.Pagination, scopes ...query.ScopeMethod) ([]*dobo.ApiDO, error)
		// Delete 删除api
		Delete(ctx context.Context, scopes ...query.ScopeMethod) error
		// Update 更新api
		Update(ctx context.Context, apiDo *dobo.ApiDO, scopes ...query.ScopeMethod) (*dobo.ApiDO, error)
	}

	UnimplementedApiRepo struct{}
)

func (UnimplementedApiRepo) Create(ctx context.Context, apiDoList ...*dobo.ApiDO) ([]*dobo.ApiDO, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Create not implemented")
}

func (UnimplementedApiRepo) Get(ctx context.Context, scopes ...query.ScopeMethod) (*dobo.ApiDO, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Get not implemented")
}

func (UnimplementedApiRepo) List(ctx context.Context, pgInfo query.Pagination, scopes ...query.ScopeMethod) ([]*dobo.ApiDO, error) {
	return nil, status.Errorf(codes.Unimplemented, "method List not implemented")
}

func (UnimplementedApiRepo) Delete(ctx context.Context, scopes ...query.ScopeMethod) error {
	return status.Errorf(codes.Unimplemented, "method Delete not implemented")
}

func (UnimplementedApiRepo) Update(ctx context.Context, apiDo *dobo.ApiDO, scopes ...query.ScopeMethod) (*dobo.ApiDO, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Update not implemented")
}

func (UnimplementedApiRepo) unimplementedApiRepo() {}
