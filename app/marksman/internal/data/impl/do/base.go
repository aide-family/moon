// Package do is the data package for the internal data.
package do

import (
	"time"

	"github.com/bwmarrin/snowflake"
	"gorm.io/gorm"

	"github.com/aide-family/magicbox/hello"
	"github.com/aide-family/magicbox/merr"
)

func Models() []any {
	return []any{
		&Level{},
		&Datasource{},
		&StrategyGroup{},
		&Strategy{},
		&StrategyGroupReceiver{},
		&StrategyMetric{},
		&StrategyMetricLevel{},
		&StrategyMetricReceiver{},
		&AlertPage{},
		&EvaluatorSnapshot{},
		&AlertEvent{},
	}
}

type BaseModel struct {
	ID        snowflake.ID `gorm:"column:id;primaryKey"`
	CreatedAt time.Time    `gorm:"column:created_at;"`
	UpdatedAt time.Time    `gorm:"column:updated_at;"`
	Creator   snowflake.ID `gorm:"column:creator;index"`
}

func (b *BaseModel) WithCreator(creator snowflake.ID) *BaseModel {
	b.Creator = creator
	return b
}

func (b *BaseModel) BeforeCreate(tx *gorm.DB) (err error) {
	if b.Creator == 0 {
		return merr.ErrorInvalidArgument("creator is required")
	}
	node, err := snowflake.NewNode(hello.NodeID())
	if err != nil {
		return merr.ErrorInternalServer("create snowflake node failed").WithCause(err)
	}
	b.ID = node.Generate()
	return nil
}

// EventBaseModel is used for system-generated records (e.g. alert events) that have no Creator.
type EventBaseModel struct {
	ID        snowflake.ID `gorm:"column:id;primaryKey"`
	CreatedAt time.Time    `gorm:"column:created_at;"`
	UpdatedAt time.Time    `gorm:"column:updated_at;"`
}

func (e *EventBaseModel) BeforeCreate(tx *gorm.DB) (err error) {
	node, err := snowflake.NewNode(hello.NodeID())
	if err != nil {
		return merr.ErrorInternalServer("create snowflake node failed").WithCause(err)
	}
	e.ID = node.Generate()
	return nil
}
