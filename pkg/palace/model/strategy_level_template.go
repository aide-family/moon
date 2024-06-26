package model

import (
	"encoding/json"

	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
	"google.golang.org/protobuf/types/known/durationpb"

	"gorm.io/plugin/soft_delete"
)

const TableNameStrategyLevelTemplates = "strategy_level_templates"

type StrategyLevelTemplate struct {
	ID        uint32                `gorm:"column:id;type:int unsigned;primaryKey;autoIncrement:true" json:"id"`
	CreatedAt types.Time            `gorm:"column:created_at;type:timestamp;not null;default:CURRENT_TIMESTAMP;comment:创建时间" json:"created_at"` // 创建时间
	UpdatedAt types.Time            `gorm:"column:updated_at;type:timestamp;not null;default:CURRENT_TIMESTAMP;comment:更新时间" json:"updated_at"` // 更新时间
	DeletedAt soft_delete.DeletedAt `gorm:"column:deleted_at;type:bigint;not null;comment:删除时间" json:"deleted_at"`

	// 所属策略
	StrategyID       uint32            `gorm:"column:strategy_id;type:int unsigned;not null;comment:策略ID" json:"strategy_id"`
	StrategyTemplate *StrategyTemplate `gorm:"foreignKey:StrategyID" json:"strategy_template"`

	// 持续时间
	Duration durationpb.Duration `gorm:"column:duration;type:varchar(64);not null;comment:告警持续时间" json:"duration"`
	// 持续次数
	Count uint32 `gorm:"column:count;type:int unsigned;not null;comment:持续次数" json:"count"`
	// 持续事件类型
	SustainType vobj.Sustain `gorm:"column:sustain_type;type:int(11);not null;comment:持续类型" json:"sustain_type"`
	// 执行频率
	Interval durationpb.Duration `gorm:"column:interval;type:varchar(64);not null;comment:执行频率" json:"interval"`
	// 条件
	Condition string `gorm:"column:condition;type:varchar(2);not null;comment:条件" json:"condition"`
	// 阈值
	Threshold float64 `gorm:"column:threshold;type:text;not null;comment:阈值" json:"threshold"`
	// 告警等级
	LevelID uint32              `gorm:"column:level_id;type:int unsigned;not null;comment:告警等级" json:"level_id"`
	Level   *StrategyAlarmLevel `gorm:"foreignKey:LevelID" json:"level"`

	// 状态
	Status    vobj.Status `gorm:"column:status;type:int;not null;comment:策略状态" json:"status"`
	CreatorID uint32      `gorm:"column:creator;type:int unsigned;not null;comment:创建者" json:"creator_id"`
	Creator   *SysUser    `gorm:"foreignKey:CreatorID" json:"creator"`
}

// String json string
func (c *StrategyLevelTemplate) String() string {
	bs, _ := json.Marshal(c)
	return string(bs)
}

func (c *StrategyLevelTemplate) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, c)
}

func (c *StrategyLevelTemplate) MarshalBinary() (data []byte, err error) {
	return json.Marshal(c)
}

// TableName StrategyLevelTemplate's table name
func (*StrategyLevelTemplate) TableName() string {
	return TableNameStrategyLevelTemplates
}
