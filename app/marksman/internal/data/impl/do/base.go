// Package do is the data package for the internal data.
package do

import (
	"time"

	"github.com/bwmarrin/snowflake"
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
