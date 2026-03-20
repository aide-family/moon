package convert

import (
	"context"

	"github.com/aide-family/magicbox/contextx"
	"github.com/aide-family/magicbox/merr"
	"github.com/aide-family/magicbox/safety"
	"github.com/aide-family/marksman/internal/biz/bo"
	"github.com/aide-family/marksman/internal/data/impl/do"
)

func ToStrategyMetricItemBo(m *do.StrategyMetric) *bo.StrategyMetricItemBo {
	if m == nil {
		return nil
	}
	levels := make([]*bo.StrategyMetricLevelItemBo, 0, len(m.StrategyLevels))
	for _, level := range m.StrategyLevels {
		levels = append(levels, ToStrategyMetricLevelItemBo(level))
	}
	return &bo.StrategyMetricItemBo{
		UID:            m.ID,
		StrategyUID:    m.StrategyUID,
		Strategy:       ToStrategyItemBo(m.Strategy),
		Expr:           m.Expr,
		Labels:         m.Labels.Map(),
		Summary:        m.Summary,
		Description:    m.Description,
		DatasourceUIDs: m.DatasourceUIDs.List(),
		Levels:         levels,
		CreatedAt:      m.CreatedAt,
		UpdatedAt:      m.UpdatedAt,
	}
}

func ToStrategyMetricLevelItemBo(m *do.StrategyMetricLevel) *bo.StrategyMetricLevelItemBo {
	if m == nil {
		return nil
	}
	return &bo.StrategyMetricLevelItemBo{
		LevelUID:    m.LevelUID,
		StrategyUID: m.StrategyUID,
		Level:       ToLevelItemBo(m.Level),
		Mode:        m.Mode,
		Condition:   m.Condition,
		Values:      m.Values.List(),
		DurationSec: m.DurationSec,
		Status:      m.Status,
	}
}

func ToStrategyMetricDo(ctx context.Context, req *bo.SaveStrategyMetricBo) *do.StrategyMetric {
	if req == nil {
		return nil
	}
	model := &do.StrategyMetric{
		StrategyUID:    req.StrategyUID,
		Expr:           req.Expr,
		Labels:         safety.NewMap(req.Labels),
		Summary:        req.Summary,
		Description:    req.Description,
		DatasourceUIDs: safety.NewSlice(req.DatasourceUIDs),
	}
	model.WithNamespace(contextx.GetNamespace(ctx)).WithCreator(contextx.GetUserUID(ctx))
	return model
}

func ToStrategyMetricLevelDo(ctx context.Context, req *bo.SaveStrategyMetricLevelBo) *do.StrategyMetricLevel {
	if req == nil {
		return nil
	}
	model := &do.StrategyMetricLevel{
		StrategyUID: req.StrategyUID,
		LevelUID:    req.LevelUID,
		Mode:        req.Mode,
		Condition:   req.Condition,
		Values:      safety.NewSlice(req.Values),
		DurationSec: req.DurationSec,
		Status:      req.Status,
	}
	model.WithNamespace(contextx.GetNamespace(ctx)).WithCreator(contextx.GetUserUID(ctx))
	return model
}

func ToEvaluateMetricStrategyBo(
	strategyMetric *do.StrategyMetric,
	strategyLevel *do.StrategyMetricLevel,
	datasource *do.Datasource,
) (*bo.EvaluateMetricStrategyBo, error) {
	if strategyMetric == nil {
		return nil, merr.ErrorParams("strategy metric is required")
	}
	namespaceUID := strategyMetric.NamespaceUID
	if namespaceUID == 0 {
		return nil, merr.ErrorParams("namespace UID is required")
	}
	strategy := strategyMetric.Strategy
	if strategy == nil {
		return nil, merr.ErrorParams("strategy is required")
	}
	strategyGroup := strategy.StrategyGroup
	if strategyGroup == nil {
		return nil, merr.ErrorParams("strategy group is required")
	}
	if strategyLevel == nil {
		return nil, merr.ErrorParams("strategy level is required")
	}
	if datasource == nil {
		return nil, merr.ErrorParams("datasource is required")
	}

	return bo.NewEvaluateMetricStrategyBo(
		namespaceUID,
		ToStrategyGroupItemBo(strategyGroup),
		ToStrategyItemBo(strategy),
		ToStrategyMetricItemBo(strategyMetric),
		ToStrategyMetricLevelItemBo(strategyLevel),
		ToDatasourceItemBo(datasource),
	)
}
