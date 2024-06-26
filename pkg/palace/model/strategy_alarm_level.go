package model

import (
	"context"
	"encoding/json"

	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"

	"gorm.io/gen"
	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"
)

const TableNameStrategyAlarmLevel = "strategy_alarm_levels"

// StrategyAlarmLevel mapped from table <strategy_levels>
type StrategyAlarmLevel struct {
	ID        uint32                `gorm:"column:id;type:int unsigned;primaryKey;autoIncrement:true" json:"id"`
	CreatedAt types.Time            `gorm:"column:created_at;type:timestamp;not null;default:CURRENT_TIMESTAMP;comment:创建时间" json:"created_at"` // 创建时间
	UpdatedAt types.Time            `gorm:"column:updated_at;type:timestamp;not null;default:CURRENT_TIMESTAMP;comment:更新时间" json:"updated_at"` // 更新时间
	DeletedAt soft_delete.DeletedAt `gorm:"column:deleted_at;type:bigint;not null;comment:删除时间" json:"deleted_at"`

	Name   string      `gorm:"column:name;type:varchar(64);not null;uniqueIndex:idx__level__name,priority:1;comment:api名称" json:"name"` // api名称
	Status vobj.Status `gorm:"column:status;type:tinyint;not null;comment:状态" json:"status"`                                            // 状态
	Level  int         `gorm:"column:level;type:int;not null;comment:告警等级" json:"level"`
	Color  string      `gorm:"column:color;type:varchar(64);not null;comment:颜色" json:"color"`
}

// String json string
func (c *StrategyAlarmLevel) String() string {
	bs, _ := json.Marshal(c)
	return string(bs)
}

func (c *StrategyAlarmLevel) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, c)
}

func (c *StrategyAlarmLevel) MarshalBinary() (data []byte, err error) {
	return json.Marshal(c)
}

// Create func
func (c *StrategyAlarmLevel) Create(ctx context.Context, tx *gorm.DB) error {
	return tx.WithContext(ctx).Create(c).Error
}

// Update func
func (c *StrategyAlarmLevel) Update(ctx context.Context, tx *gorm.DB, conds []gen.Condition) error {
	return tx.WithContext(ctx).Model(c).Where(conds).Updates(c).Error
}

// Delete func
func (c *StrategyAlarmLevel) Delete(ctx context.Context, tx *gorm.DB, conds []gen.Condition) error {
	return tx.WithContext(ctx).Where(conds).Delete(c).Error
}

// TableName StrategyAlarmLevel's table name
func (*StrategyAlarmLevel) TableName() string {
	return TableNameStrategyAlarmLevel
}
