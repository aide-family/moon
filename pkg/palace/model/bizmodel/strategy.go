package bizmodel

import (
	"github.com/aide-family/moon/pkg/label"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"

	"gorm.io/plugin/soft_delete"
)

const tableNameStrategy = "strategies"

// Strategy mapped from table <Strategy>
type Strategy struct {
	AllFieldModel
	// StrategyType 策略类型
	StrategyType vobj.StrategyType `gorm:"column:strategy_type;type:int;not null;comment:策略类型" json:"strategy_type"`
	// 模板ID, 用于标记是否从模板创建而来
	TemplateID uint32                `gorm:"column:strategy_template_id;type:int unsigned;not null;comment:策略模板ID" json:"template_id"`
	GroupID    uint32                `gorm:"column:group_id;type:int unsigned;not null;comment:策略规则组ID;uniqueIndex:idx__strategy__group_id__name,priority:2" json:"group_id"`
	DeletedAt  soft_delete.DeletedAt `gorm:"column:deleted_at;type:bigint;not null;default:0;uniqueIndex:idx__strategy__group_id__name,priority:3" json:"deleted_at"`
	// 策略模板来源（系统、团队）
	TemplateSource vobj.StrategyTemplateSource `gorm:"column:strategy_template_source;type:tinyint;not null;comment:策略模板来源（系统、团队）" json:"template_source"`
	Name           string                      `gorm:"column:alert;type:varchar(64);not null;comment:策略名称;uniqueIndex:idx__strategy__group_id__name,priority:1" json:"name"`
	Expr           string                      `gorm:"column:expr;type:text;not null;comment:告警表达式" json:"expr"`
	Labels         *label.Labels               `gorm:"column:labels;type:JSON;not null;comment:标签" json:"labels"`
	Annotations    *label.Annotations          `gorm:"column:annotations;type:JSON;not null;comment:注解" json:"annotations"`
	Remark         string                      `gorm:"column:remark;type:varchar(255);not null;comment:备注" json:"remark"`
	Status         vobj.Status                 `gorm:"column:status;type:int;not null;comment:策略状态" json:"status"`
	Datasource     []*Datasource               `gorm:"many2many:strategy_datasource;" json:"datasource"`
	// 策略类型
	Categories []*SysDict `gorm:"many2many:strategy_categories" json:"categories"`
	// 策略告警组
	AlarmNoticeGroups []*AlarmNoticeGroup `gorm:"many2many:strategies_alarm_groups;" json:"alarm_groups"`
	// 策略组
	Group *StrategyGroup `gorm:"foreignKey:GroupID" json:"group"`
	// 策略等级
	Level *StrategyLevel `gorm:"foreignKey:StrategyID" json:"level"`
}

// GetLevel 获取策略等级
func (c *Strategy) GetLevel() *StrategyLevel {
	if c == nil {
		return nil
	}
	return c.Level
}

// GetAlarmNoticeGroups 获取告警组
func (c *Strategy) GetAlarmNoticeGroups() []*AlarmNoticeGroup {
	if types.IsNil(c) {
		return nil
	}
	return c.AlarmNoticeGroups
}

// String json string
func (c *Strategy) String() string {
	bs, _ := types.Marshal(c)
	return string(bs)
}

// UnmarshalBinary redis存储实现
func (c *Strategy) UnmarshalBinary(data []byte) error {
	return types.Unmarshal(data, c)
}

// MarshalBinary redis存储实现
func (c *Strategy) MarshalBinary() (data []byte, err error) {
	return types.Marshal(c)
}

// TableName Strategy's table name
func (*Strategy) TableName() string {
	return tableNameStrategy
}
