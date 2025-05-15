package repository

import (
	"context"

	"github.com/aide-family/moon/cmd/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
)

type TeamStrategy interface {
	Create(ctx context.Context, params bo.CreateTeamStrategyParams) (do.Strategy, error)
	Update(ctx context.Context, params bo.UpdateTeamStrategyParams) (do.Strategy, error)
	Delete(ctx context.Context, params *bo.OperateTeamStrategyParams) error
	UpdateStatus(ctx context.Context, params *bo.UpdateTeamStrategiesStatusParams) error
	List(ctx context.Context, params *bo.ListTeamStrategyParams) (*bo.ListTeamStrategyReply, error)
	Subscribe(ctx context.Context, params bo.SubscribeTeamStrategy) error
	SubscribeList(ctx context.Context, params *bo.SubscribeTeamStrategiesParams) (*bo.SubscribeTeamStrategiesReply, error)
	Get(ctx context.Context, params *bo.OperateTeamStrategyParams) (do.Strategy, error)
}

type TeamStrategyMetric interface {
	Create(ctx context.Context, params bo.CreateTeamMetricStrategyParams) (do.StrategyMetric, error)
	Update(ctx context.Context, params bo.UpdateTeamMetricStrategyParams) (do.StrategyMetric, error)
	Get(ctx context.Context, params *bo.OperateTeamStrategyParams) (do.StrategyMetric, error)
	Delete(ctx context.Context, params *bo.OperateTeamStrategyParams) error

	FindLevels(ctx context.Context, params *bo.FindTeamMetricStrategyLevelsParams) ([]do.StrategyMetricRule, error)
	CreateLevels(ctx context.Context, params bo.SaveTeamMetricStrategyLevels) ([]do.StrategyMetricRule, error)
	UpdateLevels(ctx context.Context, params bo.SaveTeamMetricStrategyLevels) ([]do.StrategyMetricRule, error)
	DeleteLevels(ctx context.Context, params *bo.OperateTeamStrategyParams) error
	DeleteUnUsedLevels(ctx context.Context, params *bo.DeleteUnUsedLevelsParams) error
}
