package repository

import (
	"context"

	query "github.com/aide-cloud/gorm-normalize"
	"prometheus-manager/app/prom_server/internal/biz/dobo"
)

type (
	StrategyRepo interface {
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
)
