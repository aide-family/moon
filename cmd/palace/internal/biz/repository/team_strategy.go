package repository

import (
	"context"

	"github.com/aide-family/moon/cmd/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
)

type TeamStrategy interface {
	NameExists(ctx context.Context, name string, strategyId uint32) error
	Create(ctx context.Context, params bo.CreateTeamStrategyParams) error
	Update(ctx context.Context, params bo.UpdateTeamStrategyParams) error
	Delete(ctx context.Context, params *bo.OperateTeamStrategyParams) error
	UpdateStatus(ctx context.Context, params *bo.UpdateTeamStrategiesStatusParams) error
	List(ctx context.Context, params *bo.ListTeamStrategyParams) (*bo.ListTeamStrategyReply, error)
	Subscribe(ctx context.Context, params bo.SubscribeTeamStrategy) error
	SubscribeList(ctx context.Context, params *bo.SubscribeTeamStrategiesParams) (*bo.SubscribeTeamStrategiesReply, error)
	Get(ctx context.Context, params *bo.OperateTeamStrategyParams) (do.Strategy, error)
	GetByName(ctx context.Context, name string) (do.Strategy, error)
	FindByStrategiesGroupId(ctx context.Context, strategyGroupId uint32) ([]do.Strategy, error)
	DeleteByStrategyIds(ctx context.Context, strategyIds ...uint32) error
}

type TeamStrategyMetric interface {
	Create(ctx context.Context, params bo.CreateTeamMetricStrategyParams) error
	Update(ctx context.Context, params bo.UpdateTeamMetricStrategyParams) error
	Get(ctx context.Context, params *bo.OperateTeamStrategyParams) (do.StrategyMetric, error)
	Delete(ctx context.Context, params *bo.OperateTeamStrategyParams) error
	DeleteByStrategyIds(ctx context.Context, strategyIds ...uint32) error
}

type TeamStrategyMetricLevel interface {
	Create(ctx context.Context, params bo.SaveTeamMetricStrategyLevel) error
	Update(ctx context.Context, params bo.SaveTeamMetricStrategyLevel) error
	Delete(ctx context.Context, strategyMetricLevelId uint32) error
	DeleteByStrategyIds(ctx context.Context, strategyIds ...uint32) error
	List(ctx context.Context, params *bo.ListTeamMetricStrategyLevelsParams) (*bo.ListTeamMetricStrategyLevelsReply, error)
	UpdateStatus(ctx context.Context, params *bo.UpdateTeamMetricStrategyLevelStatusParams) error
	GetByLevelId(ctx context.Context, params *bo.OperateTeamStrategyLevelParams) (do.StrategyMetricRule, error)
	Get(ctx context.Context, strategyMetricLevelId uint32) (do.StrategyMetricRule, error)
}
