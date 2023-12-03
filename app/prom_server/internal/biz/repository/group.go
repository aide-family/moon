package repository

import (
	"context"

	query "github.com/aide-cloud/gorm-normalize"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"prometheus-manager/app/prom_server/internal/biz/bo"
	"prometheus-manager/pkg/helper/valueobj"
)

var _ StrategyGroupRepo = (*UnimplementedStrategyGroupRepo)(nil)

type (
	StrategyGroupRepo interface {
		mustEmbedUnimplemented()
		Create(ctx context.Context, strategyGroup *bo.StrategyGroupBO) (*bo.StrategyGroupBO, error)
		UpdateById(ctx context.Context, id uint, strategyGroup *bo.StrategyGroupBO) (*bo.StrategyGroupBO, error)
		BatchUpdateStatus(ctx context.Context, status valueobj.Status, ids []uint) error
		DeleteByIds(ctx context.Context, ids ...uint) error
		GetById(ctx context.Context, id uint) (*bo.StrategyGroupBO, error)
		List(ctx context.Context, pgInfo query.Pagination, scopes ...query.ScopeMethod) ([]*bo.StrategyGroupBO, error)
	}

	UnimplementedStrategyGroupRepo struct{}
)

func (UnimplementedStrategyGroupRepo) mustEmbedUnimplemented() {}

func (UnimplementedStrategyGroupRepo) Create(_ context.Context, _ *bo.StrategyGroupBO) (*bo.StrategyGroupBO, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Create not implemented")
}

func (UnimplementedStrategyGroupRepo) UpdateById(_ context.Context, _ uint, _ *bo.StrategyGroupBO) (*bo.StrategyGroupBO, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateById not implemented")
}

func (UnimplementedStrategyGroupRepo) BatchUpdateStatus(_ context.Context, _ valueobj.Status, _ []uint) error {
	return status.Errorf(codes.Unimplemented, "method BatchUpdateStatus not implemented")
}

func (UnimplementedStrategyGroupRepo) DeleteByIds(_ context.Context, _ ...uint) error {
	return status.Errorf(codes.Unimplemented, "method DeleteByIds not implemented")
}

func (UnimplementedStrategyGroupRepo) GetById(_ context.Context, _ uint) (*bo.StrategyGroupBO, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetById not implemented")
}

func (UnimplementedStrategyGroupRepo) List(_ context.Context, _ query.Pagination, _ ...query.ScopeMethod) ([]*bo.StrategyGroupBO, error) {
	return nil, status.Errorf(codes.Unimplemented, "method List not implemented")
}
