package team

import (
	"github.com/moon-monitor/moon/cmd/palace/internal/biz/do"
	"github.com/moon-monitor/moon/cmd/palace/internal/biz/vobj"
	"github.com/moon-monitor/moon/pkg/util/slices"
)

var _ do.Strategy = (*Strategy)(nil)

const tableNameStrategy = "team_strategies"

type Strategy struct {
	do.TeamModel
	StrategyGroupID uint32            `gorm:"column:strategy_group_id;type:int unsigned;not null;comment:组ID" json:"strategyGroupID"`
	Name            string            `gorm:"column:name;type:varchar(64);not null;comment:名称" json:"name"`
	Remark          string            `gorm:"column:remark;type:varchar(255);not null;comment:备注" json:"remark"`
	Status          vobj.GlobalStatus `gorm:"column:status;type:tinyint(2);not null;comment:状态" json:"status"`
	StrategyType    vobj.StrategyType `gorm:"column:strategy_type;type:tinyint(2);not null;comment:类型" json:"strategyType"`
	StrategyGroup   *StrategyGroup    `gorm:"foreignKey:StrategyGroupID;references:ID" json:"strategyGroup"`
	Notices         []*NoticeGroup    `gorm:"many2many:team_strategy_notice_groups" json:"notices"`
}

func (s *Strategy) GetStrategyGroupID() uint32 {
	if s == nil {
		return 0
	}
	return s.StrategyGroupID
}

func (s *Strategy) GetStrategyGroup() do.StrategyGroup {
	if s == nil || s.StrategyGroup == nil {
		return nil
	}
	return s.StrategyGroup
}

func (s *Strategy) GetStatus() vobj.GlobalStatus {
	if s == nil {
		return vobj.GlobalStatusUnknown
	}
	return s.Status
}

func (s *Strategy) GetName() string {
	if s == nil {
		return ""
	}
	return s.Name
}

func (s *Strategy) GetRemark() string {
	if s == nil {
		return ""
	}
	return s.Remark
}

func (s *Strategy) GetNotices() []do.NoticeGroup {
	if s == nil {
		return nil
	}
	return slices.Map(s.Notices, func(v *NoticeGroup) do.NoticeGroup { return v })
}

func (s *Strategy) GetStrategyType() vobj.StrategyType {
	if s == nil {
		return vobj.StrategyTypeUnknown
	}
	return vobj.StrategyType(s.StrategyType)
}

func (s *Strategy) TableName() string {
	return tableNameStrategy
}
