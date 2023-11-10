package prombiz

import (
	"context"

	query "github.com/aide-cloud/gorm-normalize"
	"github.com/go-kratos/kratos/v2/log"
	"prometheus-manager/app/prom_server/internal/biz"
)

type (
	StrategyBiz struct {
		log *log.Helper

		strategyRepo StrategyRepo
	}

	StrategyRepo interface {
		// CreateStrategy 创建策略
		CreateStrategy(ctx context.Context, strategy *biz.StrategyDO) (*biz.StrategyDO, error)
		// UpdateStrategyById 通过id更新策略
		UpdateStrategyById(ctx context.Context, id uint, strategy *biz.StrategyDO) (*biz.StrategyDO, error)
		// BatchUpdateStrategyStatusByIds 通过id批量更新策略状态
		BatchUpdateStrategyStatusByIds(ctx context.Context, status int32, ids []uint) error
		// DeleteStrategyByIds 通过id删除策略
		DeleteStrategyByIds(ctx context.Context, id ...uint) error
		// GetStrategyById 通过id获取策略详情
		GetStrategyById(ctx context.Context, id uint) (*biz.StrategyDO, error)
		// ListStrategy 获取策略列表
		ListStrategy(ctx context.Context, pgInfo query.Pagination, scopes ...query.ScopeMethod) ([]*biz.StrategyDO, error)
	}
)

// NewStrategyBiz 创建策略业务对象
func NewStrategyBiz(strategyRepo StrategyRepo, logger log.Logger) *StrategyBiz {
	return &StrategyBiz{
		log: log.NewHelper(log.With(logger, "module", "strategy")),

		strategyRepo: strategyRepo,
	}
}
