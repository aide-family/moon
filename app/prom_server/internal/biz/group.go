package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"prometheus-manager/app/prom_server/internal/biz/do/basescopes"
	"prometheus-manager/app/prom_server/internal/biz/vo"

	"prometheus-manager/api"
	"prometheus-manager/app/prom_server/internal/biz/bo"
	"prometheus-manager/app/prom_server/internal/biz/repository"
)

type (
	StrategyGroupBiz struct {
		log *log.Helper

		strategyGroupRepo repository.StrategyGroupRepo
	}
)

func NewStrategyGroupBiz(strategyGroupRepo repository.StrategyGroupRepo, logger log.Logger) *StrategyGroupBiz {
	return &StrategyGroupBiz{
		log:               log.NewHelper(log.With(logger, "module", "biz.strategyGroup")),
		strategyGroupRepo: strategyGroupRepo,
	}
}

func (l *StrategyGroupBiz) Create(ctx context.Context, strategyGroup *bo.StrategyGroupBO) (*bo.StrategyGroupBO, error) {
	strategyGroup, err := l.strategyGroupRepo.Create(ctx, strategyGroup)
	if err != nil {
		return nil, err
	}
	return strategyGroup, nil
}

// BatchCreate 批量创建
func (l *StrategyGroupBiz) BatchCreate(ctx context.Context, strategyGroups []*bo.StrategyGroupBO) ([]*bo.StrategyGroupBO, error) {
	strategyGroups, err := l.strategyGroupRepo.BatchCreate(ctx, strategyGroups)
	if err != nil {
		return nil, err
	}
	return strategyGroups, nil
}

func (l *StrategyGroupBiz) UpdateById(ctx context.Context, strategyGroup *bo.StrategyGroupBO) (*bo.StrategyGroupBO, error) {
	strategyGroupBO, err := l.strategyGroupRepo.UpdateById(ctx, strategyGroup.Id, strategyGroup)
	if err != nil {
		return nil, err
	}
	return strategyGroupBO, nil
}

func (l *StrategyGroupBiz) BatchUpdateStatus(ctx context.Context, status api.Status, ids []uint32) error {
	if err := l.strategyGroupRepo.BatchUpdateStatus(ctx, vo.Status(status), ids); err != nil {
		return err
	}
	return nil
}

func (l *StrategyGroupBiz) DeleteByIds(ctx context.Context, ids ...uint32) error {
	if err := l.strategyGroupRepo.DeleteByIds(ctx, ids...); err != nil {
		return err
	}
	return nil
}

func (l *StrategyGroupBiz) GetById(ctx context.Context, id uint32) (*bo.StrategyGroupBO, error) {
	strategyGroupBO, err := l.strategyGroupRepo.GetById(ctx, id)
	if err != nil {
		return nil, err
	}
	return strategyGroupBO, nil
}

func (l *StrategyGroupBiz) List(ctx context.Context, pgInfo basescopes.Pagination, scopes ...basescopes.ScopeMethod) ([]*bo.StrategyGroupBO, error) {
	strategyGroupBoList, err := l.strategyGroupRepo.List(ctx, pgInfo, scopes...)
	if err != nil {
		return nil, err
	}
	return strategyGroupBoList, nil
}

func (l *StrategyGroupBiz) ListAllLimit(ctx context.Context, limit int, scopes ...basescopes.ScopeMethod) ([]*bo.StrategyGroupBO, error) {
	return l.strategyGroupRepo.ListAllLimit(ctx, limit, scopes...)
}
