package repository

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"github.com/aide-family/moon/app/prom_server/internal/biz/bo"
	"github.com/aide-family/moon/app/prom_server/internal/biz/do"
	"github.com/aide-family/moon/app/prom_server/internal/biz/do/basescopes"
)

var _ PageRepo = (*UnimplementedPageRepo)(nil)

type (
	PageRepo interface {
		mustEmbedUnimplemented()
		// GetStrategyIds 获取策略id列表
		GetStrategyIds(ctx context.Context, scopes ...basescopes.ScopeMethod) ([]uint32, error)
		GetPromStrategyAlarmPage(ctx context.Context, scopes ...basescopes.ScopeMethod) ([]*do.PromStrategyAlarmPage, error)
		// UserPageList 获取用户页面列表
		UserPageList(ctx context.Context, userId uint32) ([]*bo.DictBO, error)
		// BindUserPages 绑定用户页面
		BindUserPages(ctx context.Context, userId uint32, pageIds []uint32) error
	}

	UnimplementedPageRepo struct{}
)

func (UnimplementedPageRepo) BindUserPages(_ context.Context, _ uint32, _ []uint32) error {
	return status.Error(codes.Unimplemented, "method BindUserPages not implemented")
}

func (UnimplementedPageRepo) UserPageList(_ context.Context, _ uint32) ([]*bo.DictBO, error) {
	return nil, status.Error(codes.Unimplemented, "method UserPageList not implemented")
}

func (UnimplementedPageRepo) GetPromStrategyAlarmPage(_ context.Context, _ ...basescopes.ScopeMethod) ([]*do.PromStrategyAlarmPage, error) {
	return nil, status.Error(codes.Unimplemented, "method GetPromStrategyAlarmPage not implemented")
}

func (UnimplementedPageRepo) GetStrategyIds(_ context.Context, _ ...basescopes.ScopeMethod) ([]uint32, error) {
	return nil, status.Error(codes.Unimplemented, "method GetStrategyIds not implemented")
}

func (UnimplementedPageRepo) mustEmbedUnimplemented() {}
