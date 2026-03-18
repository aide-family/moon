package repository

import (
	"context"

	"github.com/bwmarrin/snowflake"

	"github.com/aide-family/marksman/internal/biz/bo"
)

type StrategyMetric interface {
	CreateStrategyMetric(ctx context.Context, req *bo.SaveStrategyMetricBo) error
	UpdateStrategyMetric(ctx context.Context, req *bo.SaveStrategyMetricBo) error
	GetStrategyMetric(ctx context.Context, strategyUID snowflake.ID) (*bo.StrategyMetricItemBo, error)
	CreateStrategyMetricLevel(ctx context.Context, req *bo.SaveStrategyMetricLevelBo) error
	UpdateStrategyMetricLevel(ctx context.Context, req *bo.SaveStrategyMetricLevelBo) error
	GetStrategyMetricLevelByStrategyAndLevel(ctx context.Context, strategyUID snowflake.ID, levelUID snowflake.ID) (*bo.StrategyMetricLevelItemBo, error)
	UpdateStrategyMetricLevelStatus(ctx context.Context, req *bo.UpdateStrategyMetricLevelStatusBo) error
	DeleteStrategyMetricLevel(ctx context.Context, uid snowflake.ID, strategyUID snowflake.ID) error
	DeleteStrategyMetricLevelsByStrategyUID(ctx context.Context, strategyUID snowflake.ID) error
	DeleteStrategyMetricByStrategyUID(ctx context.Context, strategyUID snowflake.ID) error
	HasStrategyMetricData(ctx context.Context, strategyUID snowflake.ID) (bool, error)
	HasStrategyMetricLevelData(ctx context.Context, strategyUID snowflake.ID) (bool, error)
	LevelReferencedByStrategyMetricLevel(ctx context.Context, levelUID snowflake.ID) (bool, error)
	GetEvaluateMetricStrategies(ctx context.Context) ([]*bo.EvaluateMetricStrategyBo, error)
}
