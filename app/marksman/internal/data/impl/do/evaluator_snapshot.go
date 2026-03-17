package do

import (
	"time"

	"github.com/bwmarrin/snowflake"
	"gorm.io/gorm"

	"github.com/aide-family/magicbox/hello"
	"github.com/aide-family/magicbox/merr"
)

const TableNameEvaluatorSnapshot = "evaluator_snapshots"

// EvaluatorSnapshot stores one evaluator snapshot; multiple alert events reference it by ID to avoid duplicating JSON.
type EvaluatorSnapshot struct {
	ID            snowflake.ID `gorm:"column:id;primaryKey"`
	EvaluatorType string       `gorm:"column:evaluator_type;size:32;uniqueIndex:idx_evaluator_type_hash"`
	ContentHash   string       `gorm:"column:content_hash;size:64;uniqueIndex:idx_evaluator_type_hash"`
	SnapshotJSON  string       `gorm:"column:snapshot_json;type:text;default:''"`
	CreatedAt     time.Time    `gorm:"column:created_at"`
	UpdatedAt     time.Time    `gorm:"column:updated_at"`
}

func (EvaluatorSnapshot) TableName() string {
	return TableNameEvaluatorSnapshot
}

func (e *EvaluatorSnapshot) BeforeCreate(tx *gorm.DB) error {
	node, err := snowflake.NewNode(hello.NodeID())
	if err != nil {
		return merr.ErrorInternalServer("create snowflake node failed").WithCause(err)
	}
	e.ID = node.Generate()
	return nil
}
