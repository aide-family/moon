package repository

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"prometheus-manager/app/prom_server/internal/biz/do/basescopes"
	"prometheus-manager/app/prom_server/internal/biz/vo"

	"prometheus-manager/app/prom_server/internal/biz/bo"
)

var _ StrategyRepo = (*UnimplementedStrategyRepo)(nil)

type (
	StrategyRepo interface {
		mustEmbedUnimplemented()
		// CreateStrategy 创建策略
		CreateStrategy(ctx context.Context, strategy *bo.StrategyBO) (*bo.StrategyBO, error)
		// UpdateStrategyById 通过id更新策略
		UpdateStrategyById(ctx context.Context, id uint32, strategy *bo.StrategyBO) (*bo.StrategyBO, error)
		// BatchUpdateStrategyStatusByIds 通过id批量更新策略状态
		BatchUpdateStrategyStatusByIds(ctx context.Context, status vo.Status, ids []uint32) error
		// DeleteStrategyByIds 通过id删除策略
		DeleteStrategyByIds(ctx context.Context, id ...uint32) error
		// GetStrategyById 通过id获取策略详情
		GetStrategyById(ctx context.Context, id uint32, wheres ...basescopes.ScopeMethod) (*bo.StrategyBO, error)
		// ListStrategy 获取策略列表
		ListStrategy(ctx context.Context, pgInfo basescopes.Pagination, scopes ...basescopes.ScopeMethod) ([]*bo.StrategyBO, error)
		// ListStrategyByIds 通过id列表获取策略列表
		ListStrategyByIds(ctx context.Context, ids []uint32) ([]*bo.StrategyBO, error)
		// List 获取策略列表
		List(ctx context.Context, wheres ...basescopes.ScopeMethod) ([]*bo.StrategyBO, error)
	}

	UnimplementedStrategyRepo struct{}
)

func (UnimplementedStrategyRepo) mustEmbedUnimplemented() {}

func (UnimplementedStrategyRepo) List(_ context.Context, _ ...basescopes.ScopeMethod) ([]*bo.StrategyBO, error) {
	return nil, status.Error(codes.Unimplemented, "method List not implemented")
}

func (UnimplementedStrategyRepo) CreateStrategy(_ context.Context, _ *bo.StrategyBO) (*bo.StrategyBO, error) {
	return nil, status.Error(codes.Unimplemented, "method CreateStrategy not implemented")
}

func (UnimplementedStrategyRepo) UpdateStrategyById(_ context.Context, _ uint32, _ *bo.StrategyBO) (*bo.StrategyBO, error) {
	return nil, status.Error(codes.Unimplemented, "method UpdateStrategyById not implemented")
}

func (UnimplementedStrategyRepo) BatchUpdateStrategyStatusByIds(_ context.Context, _ vo.Status, _ []uint32) error {
	return status.Error(codes.Unimplemented, "method BatchUpdateStrategyStatusByIds not implemented")
}

func (UnimplementedStrategyRepo) DeleteStrategyByIds(_ context.Context, _ ...uint32) error {
	return status.Error(codes.Unimplemented, "method DeleteStrategyByIds not implemented")
}

func (UnimplementedStrategyRepo) GetStrategyById(_ context.Context, _ uint32, _ ...basescopes.ScopeMethod) (*bo.StrategyBO, error) {
	return nil, status.Error(codes.Unimplemented, "method GetStrategyById not implemented")
}

func (UnimplementedStrategyRepo) ListStrategy(_ context.Context, _ basescopes.Pagination, _ ...basescopes.ScopeMethod) ([]*bo.StrategyBO, error) {
	return nil, status.Error(codes.Unimplemented, "method ListStrategy not implemented")
}

func (UnimplementedStrategyRepo) ListStrategyByIds(_ context.Context, _ []uint32) ([]*bo.StrategyBO, error) {
	return nil, status.Error(codes.Unimplemented, "method ListStrategyByIds not implemented")
}
