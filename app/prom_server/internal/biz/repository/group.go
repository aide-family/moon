package repository

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"prometheus-manager/app/prom_server/internal/biz/do/basescopes"
	"prometheus-manager/app/prom_server/internal/biz/vo"

	"prometheus-manager/app/prom_server/internal/biz/bo"
)

var _ StrategyGroupRepo = (*UnimplementedStrategyGroupRepo)(nil)

type (
	StrategyGroupRepo interface {
		mustEmbedUnimplemented()
		Create(ctx context.Context, strategyGroup *bo.StrategyGroupBO) (*bo.StrategyGroupBO, error)
		UpdateById(ctx context.Context, id uint32, strategyGroup *bo.StrategyGroupBO) (*bo.StrategyGroupBO, error)
		BatchUpdateStatus(ctx context.Context, status vo.Status, ids []uint32) error
		DeleteByIds(ctx context.Context, ids ...uint32) error
		GetById(ctx context.Context, id uint32) (*bo.StrategyGroupBO, error)
		List(ctx context.Context, pgInfo basescopes.Pagination, scopes ...basescopes.ScopeMethod) ([]*bo.StrategyGroupBO, error)
		ListAllLimit(ctx context.Context, limit int, scopes ...basescopes.ScopeMethod) ([]*bo.StrategyGroupBO, error)
		UpdateStrategyCount(ctx context.Context, ids ...uint32) error
		UpdateEnableStrategyCount(ctx context.Context, ids ...uint32) error
	}

	UnimplementedStrategyGroupRepo struct{}
)

func (UnimplementedStrategyGroupRepo) ListAllLimit(_ context.Context, _ int, _ ...basescopes.ScopeMethod) ([]*bo.StrategyGroupBO, error) {
	return nil, status.Error(codes.Unimplemented, "method ListAllLimit not implemented")
}

func (UnimplementedStrategyGroupRepo) UpdateEnableStrategyCount(_ context.Context, _ ...uint32) error {
	return status.Error(codes.Unimplemented, "method UpdateEnableStrategyCount not implemented")
}

func (UnimplementedStrategyGroupRepo) UpdateStrategyCount(_ context.Context, _ ...uint32) error {
	return status.Error(codes.Unimplemented, "method UpdateStrategyCount not implemented")
}

func (UnimplementedStrategyGroupRepo) mustEmbedUnimplemented() {}

func (UnimplementedStrategyGroupRepo) Create(_ context.Context, _ *bo.StrategyGroupBO) (*bo.StrategyGroupBO, error) {
	return nil, status.Error(codes.Unimplemented, "method Create not implemented")
}

func (UnimplementedStrategyGroupRepo) UpdateById(_ context.Context, _ uint32, _ *bo.StrategyGroupBO) (*bo.StrategyGroupBO, error) {
	return nil, status.Error(codes.Unimplemented, "method UpdateById not implemented")
}

func (UnimplementedStrategyGroupRepo) BatchUpdateStatus(_ context.Context, _ vo.Status, _ []uint32) error {
	return status.Error(codes.Unimplemented, "method BatchUpdateStatus not implemented")
}

func (UnimplementedStrategyGroupRepo) DeleteByIds(_ context.Context, _ ...uint32) error {
	return status.Error(codes.Unimplemented, "method DeleteByIds not implemented")
}

func (UnimplementedStrategyGroupRepo) GetById(_ context.Context, _ uint32) (*bo.StrategyGroupBO, error) {
	return nil, status.Error(codes.Unimplemented, "method GetById not implemented")
}

func (UnimplementedStrategyGroupRepo) List(_ context.Context, _ basescopes.Pagination, _ ...basescopes.ScopeMethod) ([]*bo.StrategyGroupBO, error) {
	return nil, status.Error(codes.Unimplemented, "method List not implemented")
}
