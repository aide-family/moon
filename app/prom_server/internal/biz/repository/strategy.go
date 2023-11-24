package repository

import (
	"context"

	query "github.com/aide-cloud/gorm-normalize"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"prometheus-manager/app/prom_server/internal/biz/dobo"
)

var _ StrategyRepo = (*UnimplementedStrategyRepo)(nil)

type (
	StrategyRepo interface {
		mustEmbedUnimplemented()
		// CreateStrategy 创建策略
		CreateStrategy(ctx context.Context, strategy *dobo.StrategyDO) (*dobo.StrategyDO, error)
		// UpdateStrategyById 通过id更新策略
		UpdateStrategyById(ctx context.Context, id uint, strategy *dobo.StrategyDO) (*dobo.StrategyDO, error)
		// BatchUpdateStrategyStatusByIds 通过id批量更新策略状态
		BatchUpdateStrategyStatusByIds(ctx context.Context, status int32, ids []uint) error
		// DeleteStrategyByIds 通过id删除策略
		DeleteStrategyByIds(ctx context.Context, id ...uint) error
		// GetStrategyById 通过id获取策略详情
		GetStrategyById(ctx context.Context, id uint) (*dobo.StrategyDO, error)
		// ListStrategy 获取策略列表
		ListStrategy(ctx context.Context, pgInfo query.Pagination, scopes ...query.ScopeMethod) ([]*dobo.StrategyDO, error)
		// ListStrategyByIds 通过id列表获取策略列表
		ListStrategyByIds(ctx context.Context, ids []uint) ([]*dobo.StrategyDO, error)
	}

	UnimplementedStrategyRepo struct{}
)

func (UnimplementedStrategyRepo) mustEmbedUnimplemented() {}

func (UnimplementedStrategyRepo) CreateStrategy(_ context.Context, _ *dobo.StrategyDO) (*dobo.StrategyDO, error) {
	//TODO implement me
	panic("implement me")
}

func (UnimplementedStrategyRepo) UpdateStrategyById(_ context.Context, _ uint, _ *dobo.StrategyDO) (*dobo.StrategyDO, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateStrategyById not implemented")
}

func (UnimplementedStrategyRepo) BatchUpdateStrategyStatusByIds(_ context.Context, _ int32, _ []uint) error {
	return status.Errorf(codes.Unimplemented, "method BatchUpdateStrategyStatusByIds not implemented")
}

func (UnimplementedStrategyRepo) DeleteStrategyByIds(_ context.Context, _ ...uint) error {
	return status.Errorf(codes.Unimplemented, "method DeleteStrategyByIds not implemented")
}

func (UnimplementedStrategyRepo) GetStrategyById(_ context.Context, _ uint) (*dobo.StrategyDO, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetStrategyById not implemented")
}

func (UnimplementedStrategyRepo) ListStrategy(_ context.Context, _ query.Pagination, _ ...query.ScopeMethod) ([]*dobo.StrategyDO, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListStrategy not implemented")
}

func (UnimplementedStrategyRepo) ListStrategyByIds(_ context.Context, _ []uint) ([]*dobo.StrategyDO, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListStrategyByIds not implemented")
}
