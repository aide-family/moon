package repository

import (
	"context"

	query "github.com/aide-cloud/gorm-normalize"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"prometheus-manager/app/prom_server/internal/biz/dobo"
)

var _ StrategyGroupRepo = (*UnimplementedStrategyGroupRepo)(nil)

type (
	StrategyGroupRepo interface {
		mustEmbedUnimplemented()
		Create(ctx context.Context, strategyGroup *dobo.StrategyGroupDO) (*dobo.StrategyGroupDO, error)
		UpdateById(ctx context.Context, id uint, strategyGroup *dobo.StrategyGroupDO) (*dobo.StrategyGroupDO, error)
		BatchUpdateStatus(ctx context.Context, status int32, ids []uint) error
		DeleteByIds(ctx context.Context, ids ...uint) error
		GetById(ctx context.Context, id uint) (*dobo.StrategyGroupDO, error)
		List(ctx context.Context, pgInfo query.Pagination, scopes ...query.ScopeMethod) ([]*dobo.StrategyGroupDO, error)
	}

	UnimplementedStrategyGroupRepo struct{}
)

func (UnimplementedStrategyGroupRepo) mustEmbedUnimplemented() {}

func (UnimplementedStrategyGroupRepo) Create(_ context.Context, _ *dobo.StrategyGroupDO) (*dobo.StrategyGroupDO, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Create not implemented")
}

func (UnimplementedStrategyGroupRepo) UpdateById(_ context.Context, _ uint, _ *dobo.StrategyGroupDO) (*dobo.StrategyGroupDO, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateById not implemented")
}

func (UnimplementedStrategyGroupRepo) BatchUpdateStatus(_ context.Context, _ int32, _ []uint) error {
	return status.Errorf(codes.Unimplemented, "method BatchUpdateStatus not implemented")
}

func (UnimplementedStrategyGroupRepo) DeleteByIds(_ context.Context, _ ...uint) error {
	return status.Errorf(codes.Unimplemented, "method DeleteByIds not implemented")
}

func (UnimplementedStrategyGroupRepo) GetById(_ context.Context, _ uint) (*dobo.StrategyGroupDO, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetById not implemented")
}

func (UnimplementedStrategyGroupRepo) List(_ context.Context, _ query.Pagination, _ ...query.ScopeMethod) ([]*dobo.StrategyGroupDO, error) {
	return nil, status.Errorf(codes.Unimplemented, "method List not implemented")
}
