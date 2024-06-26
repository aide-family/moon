package build

import (
	"github.com/aide-family/moon/api"
	"github.com/aide-family/moon/api/admin"
	"github.com/aide-family/moon/pkg/palace/model"
)

type StrategyAlarmLevelBuilder struct {
	*model.StrategyAlarmLevel
}

func NewStrategyAlarmLevelBuilder(level *model.StrategyAlarmLevel) *StrategyAlarmLevelBuilder {
	return &StrategyAlarmLevelBuilder{
		StrategyAlarmLevel: level,
	}
}

func (b *StrategyAlarmLevelBuilder) ToApi() *admin.StrategyAlarmLevel {
	return &admin.StrategyAlarmLevel{
		Id:        b.ID,
		Name:      b.Name,
		Status:    api.Status(b.Status),
		CreatedAt: b.CreatedAt.String(),
		UpdatedAt: b.UpdatedAt.String(),
		Level:     int32(b.Level),
	}
}

// ToApiSelect 转换成Select
func (b *StrategyAlarmLevelBuilder) ToApiSelect() *admin.Select {
	return &admin.Select{
		Value:    b.ID,
		Label:    b.Name,
		Children: nil,
		Disabled: b.DeletedAt > 0 || !b.Status.IsEnable(),
		Extend:   nil,
	}
}
