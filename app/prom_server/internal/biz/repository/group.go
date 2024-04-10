package repository

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"github.com/aide-family/moon/app/prom_server/internal/biz/do/basescopes"
	"github.com/aide-family/moon/app/prom_server/internal/biz/vobj"

	"github.com/aide-family/moon/app/prom_server/internal/biz/bo"
)

var _ StrategyGroupRepo = (*UnimplementedStrategyGroupRepo)(nil)

type (
	StrategyGroupRepo interface {
		mustEmbedUnimplemented()
		Create(ctx context.Context, strategyGroup *bo.StrategyGroupBO) (*bo.StrategyGroupBO, error)
		UpdateById(ctx context.Context, id uint32, strategyGroup *bo.StrategyGroupBO) (*bo.StrategyGroupBO, error)
		BatchUpdateStatus(ctx context.Context, status vobj.Status, ids []uint32) error
		DeleteByIds(ctx context.Context, ids ...uint32) error
		GetById(ctx context.Context, id uint32) (*bo.StrategyGroupBO, error)
		GetByParams(ctx context.Context, scopes ...basescopes.ScopeMethod) ([]*bo.StrategyGroupBO, error)
		List(ctx context.Context, pgInfo bo.Pagination, scopes ...basescopes.ScopeMethod) ([]*bo.StrategyGroupBO, error)
		ListAllLimit(ctx context.Context, limit int, scopes ...basescopes.ScopeMethod) ([]*bo.StrategyGroupBO, error)
		UpdateStrategyCount(ctx context.Context, ids ...uint32) error
		UpdateEnableStrategyCount(ctx context.Context, ids ...uint32) error
		BatchCreate(ctx context.Context, strategyGroups []*bo.StrategyGroupBO) ([]*bo.StrategyGroupBO, error)
	}

	UnimplementedStrategyGroupRepo struct{}
)

func (UnimplementedStrategyGroupRepo) GetByParams(_ context.Context, _ ...basescopes.ScopeMethod) ([]*bo.StrategyGroupBO, error) {
	return nil, status.Error(codes.Unimplemented, "method GetByParams not implemented")
}

func (UnimplementedStrategyGroupRepo) BatchCreate(_ context.Context, _ []*bo.StrategyGroupBO) ([]*bo.StrategyGroupBO, error) {
	return nil, status.Error(codes.Unimplemented, "method BatchCreate not implemented")
}

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

func (UnimplementedStrategyGroupRepo) BatchUpdateStatus(_ context.Context, _ vobj.Status, _ []uint32) error {
	return status.Error(codes.Unimplemented, "method BatchUpdateStatus not implemented")
}

func (UnimplementedStrategyGroupRepo) DeleteByIds(_ context.Context, _ ...uint32) error {
	return status.Error(codes.Unimplemented, "method DeleteByIds not implemented")
}

func (UnimplementedStrategyGroupRepo) GetById(_ context.Context, _ uint32) (*bo.StrategyGroupBO, error) {
	return nil, status.Error(codes.Unimplemented, "method GetById not implemented")
}

func (UnimplementedStrategyGroupRepo) List(_ context.Context, _ bo.Pagination, _ ...basescopes.ScopeMethod) ([]*bo.StrategyGroupBO, error) {
	return nil, status.Error(codes.Unimplemented, "method List not implemented")
}
