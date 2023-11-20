package biz

import (
	"context"

	query "github.com/aide-cloud/gorm-normalize"
	"github.com/go-kratos/kratos/v2/log"
	"prometheus-manager/api"
	"prometheus-manager/app/prom_server/internal/biz/dobo"
	"prometheus-manager/app/prom_server/internal/biz/repository"
	"prometheus-manager/pkg/util/slices"
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

func (l *StrategyGroupBiz) Create(ctx context.Context, strategyGroup *dobo.StrategyGroupBO) (*dobo.StrategyGroupBO, error) {
	strategyGroupDO := dobo.NewStrategyGroupBO(strategyGroup).DO().First()
	strategyGroupDO, err := l.strategyGroupRepo.Create(ctx, strategyGroupDO)
	if err != nil {
		return nil, err
	}
	return dobo.NewStrategyGroupDO(strategyGroupDO).BO().First(), nil
}

func (l *StrategyGroupBiz) UpdateById(ctx context.Context, strategyGroup *dobo.StrategyGroupBO) (*dobo.StrategyGroupBO, error) {
	strategyGroupDO := dobo.NewStrategyGroupBO(strategyGroup).DO().First()
	strategyGroupDO, err := l.strategyGroupRepo.UpdateById(ctx, strategyGroupDO.Id, strategyGroupDO)
	if err != nil {
		return nil, err
	}
	return dobo.NewStrategyGroupDO(strategyGroupDO).BO().First(), nil
}

func (l *StrategyGroupBiz) BatchUpdateStatus(ctx context.Context, status api.Status, ids []uint32) error {
	if err := l.strategyGroupRepo.BatchUpdateStatus(ctx, int32(status), slices.To(ids, func(t uint32) uint {
		return uint(t)
	})); err != nil {
		return err
	}
	return nil
}

func (l *StrategyGroupBiz) DeleteByIds(ctx context.Context, ids ...uint32) error {
	if err := l.strategyGroupRepo.DeleteByIds(ctx, slices.To(ids, func(t uint32) uint {
		return uint(t)
	})...); err != nil {
		return err
	}
	return nil
}

func (l *StrategyGroupBiz) GetById(ctx context.Context, id uint32) (*dobo.StrategyGroupBO, error) {
	strategyGroupDO, err := l.strategyGroupRepo.GetById(ctx, uint(id))
	if err != nil {
		return nil, err
	}
	return dobo.NewStrategyGroupDO(strategyGroupDO).BO().First(), nil
}

func (l *StrategyGroupBiz) List(ctx context.Context, pgInfo query.Pagination, scopes ...query.ScopeMethod) ([]*dobo.StrategyGroupBO, error) {
	strategyGroupDoList, err := l.strategyGroupRepo.List(ctx, pgInfo, scopes...)
	if err != nil {
		return nil, err
	}
	return dobo.NewStrategyGroupDO(strategyGroupDoList...).BO().List(), nil
}
