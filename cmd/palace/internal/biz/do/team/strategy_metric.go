package team

import (
	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/pkg/util/kv"
	"github.com/aide-family/moon/pkg/util/slices"
)

var _ do.StrategyMetric = (*StrategyMetric)(nil)

const tableNameStrategyMetrics = "team_strategy_metrics"

type StrategyMetric struct {
	do.TeamModel
	StrategyID          uint32                `gorm:"column:strategy_id;type:int unsigned;not null;comment:策略id" json:"strategyID"`
	Strategy            *Strategy             `gorm:"foreignKey:StrategyID;references:ID" json:"strategy"`
	Expr                string                `gorm:"column:expr;type:varchar(1024);not null;comment:表达式" json:"expr"`
	Labels              kv.StringMap          `gorm:"column:labels;type:json;not null;comment:标签" json:"labels"`
	Annotations         kv.StringMap          `gorm:"column:annotations;type:json;not null;comment:注解" json:"annotations"`
	StrategyMetricRules []*StrategyMetricRule `gorm:"foreignKey:StrategyMetricID;references:ID" json:"strategyMetricRules"`
	Datasource          []*DatasourceMetric   `gorm:"many2many:team_strategy_metric_datasource" json:"datasource"`
}

func (m *StrategyMetric) GetStrategyID() uint32 {
	if m == nil {
		return 0
	}
	return m.StrategyID
}

func (m *StrategyMetric) GetStrategy() do.Strategy {
	if m == nil || m.Strategy == nil {
		return nil
	}
	return m.Strategy
}

func (m *StrategyMetric) GetExpr() string {
	if m == nil {
		return ""
	}
	return m.Expr
}

func (m *StrategyMetric) GetLabels() kv.StringMap {
	if m == nil {
		return nil
	}
	return m.Labels
}

func (m *StrategyMetric) GetAnnotations() kv.StringMap {
	if m == nil {
		return nil
	}
	return m.Annotations
}

func (m *StrategyMetric) GetRules() []do.StrategyMetricRule {
	if m == nil {
		return nil
	}
	return slices.Map(m.StrategyMetricRules, func(r *StrategyMetricRule) do.StrategyMetricRule { return r })
}

func (m *StrategyMetric) GetDatasourceList() []do.DatasourceMetric {
	if m == nil {
		return nil
	}
	return slices.Map(m.Datasource, func(d *DatasourceMetric) do.DatasourceMetric { return d })
}

func (m *StrategyMetric) TableName() string {
	return tableNameStrategyMetrics
}
