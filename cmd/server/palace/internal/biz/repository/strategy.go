package repository

import (
	"context"

	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
)

// Strategy 策略管理接口
type Strategy interface {
	// CreateStrategy 创建策略
	CreateStrategy(context.Context, *bo.CreateStrategyParams) (*bizmodel.Strategy, error)

	// UpdateByID 更新策略
	UpdateByID(context.Context, *bo.UpdateStrategyParams) error

	// GetByID 获取策略详情
	GetByID(context.Context, uint32) (*bizmodel.Strategy, error)

	// GetStrategyByIds 批量获取策略详情
	GetStrategyByIds(context.Context, []uint32) ([]*bizmodel.Strategy, error)

	// UpdateStatus 更新状态
	UpdateStatus(context.Context, *bo.UpdateStrategyStatusParams) error

	// FindByPage 策略分页列表
	FindByPage(context.Context, *bo.QueryStrategyListParams) ([]*bizmodel.Strategy, error)

	// DeleteByID 删除策略
	DeleteByID(context.Context, uint32) error

	// CopyStrategy 复制策略
	CopyStrategy(context.Context, uint32) (*bizmodel.Strategy, error)

	// Eval 策略评估
	Eval(context.Context, *bo.Strategy) (*bo.Alarm, error)
	// GetTeamStrategy 获取团队策略
	GetTeamStrategy(ctx context.Context, params *bo.GetTeamStrategyParams) (*bizmodel.Strategy, error)

	// GetTeamStrategyLevelByLevelID 获取团队策略等级
	GetTeamStrategyLevelByLevelID(ctx context.Context, params *bo.GetTeamStrategyLevelParams) (*bo.TeamStrategyLevelModel, error)

	// Sync 同步策略
	Sync(ctx context.Context, id uint32) error

	// GetStrategyMetricLevels 获取Metric策略等级
	GetStrategyMetricLevels(ctx context.Context, strategyID []uint32) ([]*bizmodel.StrategyMetricsLevel, error)

	// GetStrategyMQLevels 获取MQ策略等级
	GetStrategyMQLevels(ctx context.Context, strategyIds []uint32) ([]*bizmodel.StrategyMQLevel, error)

	// GetStrategyLevels 获取策略等级
	GetStrategyLevels(ctx context.Context, strategyIds []uint32) ([]*bizmodel.StrategyLevels, error)
}
