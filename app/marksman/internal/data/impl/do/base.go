// Package do is the data package for the internal data.
package do

import (
	"errors"
	"time"

	"github.com/aide-family/magicbox/hello"
	"github.com/bwmarrin/snowflake"
	"gorm.io/gorm"
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
		return errors.New("creator is required")
	}
	node, err := snowflake.NewNode(hello.NodeID())
	if err != nil {
		return err
	}
	b.ID = node.Generate()
	return nil
}
