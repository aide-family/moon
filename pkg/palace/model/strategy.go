package model

import (
	"encoding/json"

	"github.com/aide-family/moon/pkg/vobj"
)

const TableNameStrategy = "strategies"

type Strategy struct {
	AllFieldModel
	// 模板ID
	StrategyTemplateID uint32 `gorm:"column:strategy_template_id;type:int unsigned;not null;comment:策略模板ID" json:"strategy_template_id"`
	// 策略模板
	StrategyTemplate *StrategyTemplate `gorm:"foreignKey:StrategyTemplateID" json:"strategy_template"`

	Alert       string           `gorm:"column:alert;type:varchar(64);not null;comment:策略名称" json:"alert"`
	Expr        string           `gorm:"column:expr;type:text;not null;comment:告警表达式" json:"expr"`
	Status      vobj.Status      `gorm:"column:status;type:int;not null;comment:策略状态" json:"status"`
	Remark      string           `gorm:"column:remark;type:varchar(255);not null;comment:备注" json:"remark"`
	Labels      vobj.Labels      `gorm:"column:labels;type:JSON;not null;comment:标签" json:"labels"`
	Annotations vobj.Annotations `gorm:"column:annotations;type:JSON;not null;comment:注解" json:"annotations"`
	// 告警等级数据
	StrategyLevelTemplates []*StrategyLevelTemplate `gorm:"foreignKey:StrategyID" json:"strategy_level_templates"`

	// 公共通知对象
}

// String json string
func (c *Strategy) String() string {
	bs, _ := json.Marshal(c)
	return string(bs)
}

func (c *Strategy) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, c)
}

func (c *Strategy) MarshalBinary() (data []byte, err error) {
	return json.Marshal(c)
}

// TableName Strategy's table name
func (*Strategy) TableName() string {
	return TableNameStrategy
}
