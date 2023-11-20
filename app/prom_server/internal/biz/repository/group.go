package repository

import (
	"context"

	query "github.com/aide-cloud/gorm-normalize"
	"prometheus-manager/app/prom_server/internal/biz/dobo"
)

type (
	StrategyGroupRepo interface {
		Create(ctx context.Context, strategyGroup *dobo.StrategyGroupDO) (*dobo.StrategyGroupDO, error)
		UpdateById(ctx context.Context, id uint, strategyGroup *dobo.StrategyGroupDO) (*dobo.StrategyGroupDO, error)
		BatchUpdateStatus(ctx context.Context, status int32, ids []uint) error
		DeleteByIds(ctx context.Context, ids ...uint) error
		GetById(ctx context.Context, id uint) (*dobo.StrategyGroupDO, error)
		List(ctx context.Context, pgInfo query.Pagination, scopes ...query.ScopeMethod) ([]*dobo.StrategyGroupDO, error)
	}
)
