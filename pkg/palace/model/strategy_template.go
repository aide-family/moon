package model

import (
	"encoding/json"

	"github.com/aide-family/moon/pkg/vobj"
)

const TableNameStrategyTemplate = "strategy_templates"

type StrategyTemplate struct {
	AllFieldModel
	Alert       string           `gorm:"column:alert;type:varchar(64);not null;comment:策略名称" json:"alert"`
	Expr        string           `gorm:"column:expr;type:text;not null;comment:告警表达式" json:"expr"`
	Status      vobj.Status      `gorm:"column:status;type:int;not null;comment:策略状态" json:"status"`
	Remark      string           `gorm:"column:remark;type:varchar(255);not null;comment:备注" json:"remark"`
	Labels      vobj.Labels      `gorm:"column:labels;type:JSON;not null;comment:标签" json:"labels"`
	Annotations vobj.Annotations `gorm:"column:annotations;type:JSON;not null;comment:注解" json:"annotations"`
	// 告警等级数据
	StrategyLevelTemplates []*StrategyLevelTemplate `gorm:"foreignKey:StrategyTemplateID" json:"strategy_level_templates"`
	//策略模板类型
	Categories []*SysDict `gorm:"many2many:strategy_template_categories"`
}

// String json string
func (c *StrategyTemplate) String() string {
	bs, _ := json.Marshal(c)
	return string(bs)
}

func (c *StrategyTemplate) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, c)
}

func (c *StrategyTemplate) MarshalBinary() (data []byte, err error) {
	return json.Marshal(c)
}

// TableName StrategyTemplate's table name
func (*StrategyTemplate) TableName() string {
	return TableNameStrategyTemplate
}
