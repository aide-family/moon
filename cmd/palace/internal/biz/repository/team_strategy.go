package repository

import (
	"context"

	"github.com/aide-family/moon/cmd/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
)

type TeamStrategy interface {
	NameExists(ctx context.Context, name string, strategyId uint32) error
	Create(ctx context.Context, params bo.CreateTeamStrategyParams) (uint32, error)
	Update(ctx context.Context, params bo.UpdateTeamStrategyParams) error
	Delete(ctx context.Context, strategyId uint32) error
	UpdateStatus(ctx context.Context, params *bo.UpdateTeamStrategiesStatusParams) error
	List(ctx context.Context, params *bo.ListTeamStrategyParams) (*bo.ListTeamStrategyReply, error)
	Subscribe(ctx context.Context, params *bo.SubscribeTeamStrategyParams) error
	SubscribeList(ctx context.Context, params *bo.SubscribeTeamStrategiesParams) (*bo.SubscribeTeamStrategiesReply, error)
	Get(ctx context.Context, strategyId uint32) (do.Strategy, error)
	GetByName(ctx context.Context, name string) (do.Strategy, error)
	FindByStrategiesGroupIds(ctx context.Context, strategyGroupIds ...uint32) ([]do.Strategy, error)
	DeleteByStrategyIds(ctx context.Context, strategyIds ...uint32) error
}

type TeamStrategyMetric interface {
	Create(ctx context.Context, params bo.CreateTeamMetricStrategyParams) error
	Update(ctx context.Context, params bo.UpdateTeamMetricStrategyParams) error
	Get(ctx context.Context, strategyMetricId uint32) (do.StrategyMetric, error)
	GetByStrategyId(ctx context.Context, strategyId uint32) (do.StrategyMetric, error)
	Delete(ctx context.Context, strategyMetricId uint32) error
	DeleteByStrategyIds(ctx context.Context, strategyIds ...uint32) error
	FindByStrategyIds(ctx context.Context, strategyIds []uint32) ([]do.StrategyMetric, error)
}

type TeamStrategyMetricLevel interface {
	Create(ctx context.Context, params bo.CreateTeamMetricStrategyLevelParams) error
	Update(ctx context.Context, params bo.UpdateTeamMetricStrategyLevelParams) error
	Delete(ctx context.Context, strategyMetricLevelIds []uint32) error
	DeleteByStrategyIds(ctx context.Context, strategyIds ...uint32) error
	List(ctx context.Context, params *bo.ListTeamMetricStrategyLevelsParams) (*bo.ListTeamMetricStrategyLevelsReply, error)
	UpdateStatus(ctx context.Context, params *bo.UpdateTeamMetricStrategyLevelStatusParams) error
	Get(ctx context.Context, strategyMetricLevelId uint32) (do.StrategyMetricRule, error)
	FindByIds(ctx context.Context, strategyMetricLevelIds []uint32) ([]do.StrategyMetricRule, error)
}
