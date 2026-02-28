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
	UpdateStrategyMetricLevelStatus(ctx context.Context, req *bo.UpdateStrategyMetricLevelStatusBo) error
	DeleteStrategyMetricLevel(ctx context.Context, uid snowflake.ID, strategyUID snowflake.ID) error
	GetStrategyMetricLevel(ctx context.Context, uid snowflake.ID, strategyUID snowflake.ID) (*bo.StrategyMetricLevelItemBo, error)
	StrategyMetricBindReceivers(ctx context.Context, req *bo.StrategyMetricBindReceiversBo) error
}
