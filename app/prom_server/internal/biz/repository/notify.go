package repository

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"github.com/aide-family/moon/app/prom_server/internal/biz/do/basescopes"

	"github.com/aide-family/moon/app/prom_server/internal/biz/bo"
)

var _ NotifyRepo = (*UnimplementedNotifyRepo)(nil)

type (
	NotifyRepo interface {
		mustEmbedUnimplemented()
		// Get 获取通知详情
		Get(ctx context.Context, scopes ...basescopes.ScopeMethod) (*bo.NotifyBO, error)
		// Find 获取通知列表
		Find(ctx context.Context, scopes ...basescopes.ScopeMethod) ([]*bo.NotifyBO, error)
		// Count 获取通知总数
		Count(ctx context.Context, scopes ...basescopes.ScopeMethod) (int64, error)
		// List 获取通知列表
		List(ctx context.Context, pgInfo bo.Pagination, scopes ...basescopes.ScopeMethod) ([]*bo.NotifyBO, error)
		// Create 创建通知
		Create(ctx context.Context, notify *bo.NotifyBO) (*bo.NotifyBO, error)
		// Update 更新通知
		Update(ctx context.Context, notify *bo.NotifyBO, scopes ...basescopes.ScopeMethod) error
		// Delete 删除通知
		Delete(ctx context.Context, scopes ...basescopes.ScopeMethod) error
	}

	UnimplementedNotifyRepo struct{}
)

func (UnimplementedNotifyRepo) mustEmbedUnimplemented() {}

func (UnimplementedNotifyRepo) Get(_ context.Context, _ ...basescopes.ScopeMethod) (*bo.NotifyBO, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Get not implemented")
}

func (UnimplementedNotifyRepo) Find(_ context.Context, _ ...basescopes.ScopeMethod) ([]*bo.NotifyBO, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Find not implemented")
}

func (UnimplementedNotifyRepo) Count(_ context.Context, _ ...basescopes.ScopeMethod) (int64, error) {
	return 0, status.Errorf(codes.Unimplemented, "method Count not implemented")
}

func (UnimplementedNotifyRepo) List(_ context.Context, _ bo.Pagination, _ ...basescopes.ScopeMethod) ([]*bo.NotifyBO, error) {
	return nil, status.Errorf(codes.Unimplemented, "method List not implemented")
}

func (UnimplementedNotifyRepo) Create(_ context.Context, _ *bo.NotifyBO) (*bo.NotifyBO, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Create not implemented")
}

func (UnimplementedNotifyRepo) Update(_ context.Context, _ *bo.NotifyBO, _ ...basescopes.ScopeMethod) error {
	return status.Errorf(codes.Unimplemented, "method Update not implemented")
}

func (UnimplementedNotifyRepo) Delete(_ context.Context, _ ...basescopes.ScopeMethod) error {
	return status.Errorf(codes.Unimplemented, "method Delete not implemented")
}
