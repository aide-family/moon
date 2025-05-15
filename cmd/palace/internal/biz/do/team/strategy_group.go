package team

import (
	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/biz/vobj"
	"github.com/aide-family/moon/pkg/util/slices"
)

var _ do.StrategyGroup = (*StrategyGroup)(nil)

const tableNameStrategyGroup = "team_strategy_groups"

type StrategyGroup struct {
	do.TeamModel
	Name       string            `gorm:"column:name;type:varchar(64);not null;comment:名称" json:"name"`
	Remark     string            `gorm:"column:remark;type:varchar(255);not null;comment:备注" json:"remark"`
	Status     vobj.GlobalStatus `gorm:"column:status;type:tinyint(2);not null;comment:状态" json:"status"`
	Strategies []*Strategy       `gorm:"foreignKey:StrategyGroupID;references:ID" json:"strategies"`
}

func (s *StrategyGroup) GetName() string {
	if s == nil {
		return ""
	}
	return s.Name
}

func (s *StrategyGroup) GetRemark() string {
	if s == nil {
		return ""
	}
	return s.Remark
}

func (s *StrategyGroup) GetStatus() vobj.GlobalStatus {
	if s == nil {
		return vobj.GlobalStatusUnknown
	}
	return s.Status
}

func (s *StrategyGroup) GetStrategies() []do.Strategy {
	if s == nil {
		return nil
	}
	return slices.Map(s.Strategies, func(v *Strategy) do.Strategy { return v })
}

func (s *StrategyGroup) TableName() string {
	return tableNameStrategyGroup
}
