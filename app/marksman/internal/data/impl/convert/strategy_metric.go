package convert

import (
	"github.com/aide-family/magicbox/safety"
	"github.com/aide-family/marksman/internal/biz/bo"
	"github.com/aide-family/marksman/internal/data/impl/do"
)

func ToStrategyMetricItemBo(m *do.StrategyMetric, levels []*bo.StrategyMetricLevelItemBo) *bo.StrategyMetricItemBo {
	if m == nil {
		return nil
	}
	return &bo.StrategyMetricItemBo{
		StrategyUID:    m.StrategyUID,
		Expr:           m.Expr,
		Labels:         m.Labels.Map(),
		Summary:        m.Summary,
		Description:    m.Description,
		Status:         m.Status,
		DatasourceUIDs: m.DatasourceUIDs.List(),
		Levels:         levels,
		CreatedAt:      m.CreatedAt,
		UpdatedAt:      m.UpdatedAt,
	}
}

func ToStrategyMetricLevelItemBo(m *do.StrategyMetricLevel, levelDo *do.Level) *bo.StrategyMetricLevelItemBo {
	if m == nil {
		return nil
	}
	return &bo.StrategyMetricLevelItemBo{
		UID:         m.ID,
		StrategyUID: m.StrategyUID,
		Level:       ToLevelItemBo(levelDo),
		Mode:        m.Mode,
		Condition:   m.Condition,
		Values:      m.Values.List(),
		DurationSec: m.DurationSec,
		Status:      m.Status,
	}
}

func ToStrategyMetricDo(req *bo.SaveStrategyMetricBo) *do.StrategyMetric {
	if req == nil {
		return nil
	}
	return &do.StrategyMetric{
		StrategyUID:    req.StrategyUID,
		Expr:           req.Expr,
		Labels:         safety.NewMap(req.Labels),
		Summary:        req.Summary,
		Description:    req.Description,
		Status:         req.Status,
		DatasourceUIDs: safety.NewSlice(req.DatasourceUIDs),
	}
}

func ToStrategyMetricLevelDo(req *bo.SaveStrategyMetricLevelBo) *do.StrategyMetricLevel {
	if req == nil {
		return nil
	}
	return &do.StrategyMetricLevel{
		StrategyUID: req.StrategyUID,
		LevelUID:    req.LevelUID,
		Mode:        req.Mode,
		Condition:   req.Condition,
		Values:      safety.NewSlice(req.Values),
		DurationSec: req.DurationSec,
		Status:      req.Status,
	}
}
