package repository

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"github.com/aide-family/moon/app/prom_server/internal/biz/bo"
	"github.com/aide-family/moon/app/prom_server/internal/biz/do/basescopes"
)

var _ NotifyTemplateRepo = (*UnimplementedNotifyTemplateRepo)(nil)

type (
	NotifyTemplateRepo interface {
		unimplementedNotifyTemplateRepo()
		// Get 获取通知模板详情
		Get(ctx context.Context, scopes ...basescopes.ScopeMethod) (*bo.NotifyTemplateBO, error)
		// Find 获取通知模板列表
		Find(ctx context.Context, scopes ...basescopes.ScopeMethod) ([]*bo.NotifyTemplateBO, error)
		// Count 获取通知模板总数
		Count(ctx context.Context, scopes ...basescopes.ScopeMethod) (int64, error)
		// List 获取通知模板列表
		List(ctx context.Context, pgInfo bo.Pagination, scopes ...basescopes.ScopeMethod) ([]*bo.NotifyTemplateBO, error)
		// Create 创建通知模板
		Create(ctx context.Context, notifyTemplate *bo.NotifyTemplateBO) (*bo.NotifyTemplateBO, error)
		// Update 更新通知模板
		Update(ctx context.Context, notifyTemplate *bo.NotifyTemplateBO, scopes ...basescopes.ScopeMethod) error
		// Delete 删除通知模板
		Delete(ctx context.Context, scopes ...basescopes.ScopeMethod) error
	}

	UnimplementedNotifyTemplateRepo struct{}
)

func (UnimplementedNotifyTemplateRepo) unimplementedNotifyTemplateRepo() {}

func (UnimplementedNotifyTemplateRepo) Get(_ context.Context, _ ...basescopes.ScopeMethod) (*bo.NotifyTemplateBO, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Get not implemented")
}

func (UnimplementedNotifyTemplateRepo) Find(_ context.Context, _ ...basescopes.ScopeMethod) ([]*bo.NotifyTemplateBO, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Find not implemented")
}

func (UnimplementedNotifyTemplateRepo) Count(_ context.Context, _ ...basescopes.ScopeMethod) (int64, error) {
	return 0, status.Errorf(codes.Unimplemented, "method Count not implemented")
}

func (UnimplementedNotifyTemplateRepo) List(_ context.Context, _ bo.Pagination, _ ...basescopes.ScopeMethod) ([]*bo.NotifyTemplateBO, error) {
	return nil, status.Errorf(codes.Unimplemented, "method List not implemented")
}

func (UnimplementedNotifyTemplateRepo) Create(_ context.Context, _ *bo.NotifyTemplateBO) (*bo.NotifyTemplateBO, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Create not implemented")
}

func (UnimplementedNotifyTemplateRepo) Update(_ context.Context, _ *bo.NotifyTemplateBO, _ ...basescopes.ScopeMethod) error {
	return status.Errorf(codes.Unimplemented, "method Update not implemented")
}

func (UnimplementedNotifyTemplateRepo) Delete(_ context.Context, _ ...basescopes.ScopeMethod) error {
	return status.Errorf(codes.Unimplemented, "method Delete not implemented")
}
