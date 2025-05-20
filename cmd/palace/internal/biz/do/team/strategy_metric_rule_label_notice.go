package team

import (
	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/pkg/util/slices"
)

var _ do.StrategyMetricRuleLabelNotice = (*StrategyMetricRuleLabelNotice)(nil)

const tableNameStrategyMetricRuleLabelNotice = "team_strategy_metric_rule_label_notices"

type StrategyMetricRuleLabelNotice struct {
	do.TeamModel
	StrategyMetricRuleID uint32         `gorm:"column:strategy_metric_rule_id;not null;int(10) unsigned;index:idx_strategy_metric_rule_id;comment:strategy metric rule ID"`
	LabelKey             string         `gorm:"column:label_key;type:varchar(64);not null;comment:label key" json:"labelKey"`
	LabelValue           string         `gorm:"column:label_value;type:varchar(255);not null;comment:label value" json:"labelValue"`
	Notices              []*NoticeGroup `gorm:"many2many:team_strategy_metric_rule_label_notice_notice_groups" json:"notices"`
}

func (s *StrategyMetricRuleLabelNotice) TableName() string {
	return tableNameStrategyMetricRuleLabelNotice
}

func (s *StrategyMetricRuleLabelNotice) GetStrategyMetricRuleID() uint32 {
	if s == nil {
		return 0
	}
	return s.StrategyMetricRuleID
}

func (s *StrategyMetricRuleLabelNotice) GetLabelKey() string {
	if s == nil {
		return ""
	}
	return s.LabelKey
}

func (s *StrategyMetricRuleLabelNotice) GetLabelValue() string {
	if s == nil {
		return ""
	}
	return s.LabelValue
}

func (s *StrategyMetricRuleLabelNotice) GetNotices() []do.NoticeGroup {
	if s == nil {
		return nil
	}
	return slices.Map(s.Notices, func(n *NoticeGroup) do.NoticeGroup { return n })
}
