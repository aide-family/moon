package bo

import (
	"time"

	"github.com/bwmarrin/snowflake"
)

// AlertEventBo is the business object for an alert event produced by metric strategy evaluation.
type AlertEventBo struct {
	StrategyUID   snowflake.ID
	NamespaceUID  snowflake.ID
	Level         *LevelItemBo
	Summary       string
	Description   string
	Expr          string
	FiredAt       time.Time
	Value         float64
	Labels        map[string]string
	DatasourceUID snowflake.ID
}
