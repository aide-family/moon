// Package do contains GORM models for jade_tree persistence.
package do

import (
	"time"

	"github.com/aide-family/magicbox/hello"
	"github.com/aide-family/magicbox/merr"
	"github.com/bwmarrin/snowflake"
	"gorm.io/gorm"
)

func Models() []any {
	return []any{
		&MachineInfo{},
		&SSHCommand{},
		&SSHCommandAudit{},
		&ProbeTask{},
	}
}

type BaseModel struct {
	ID        snowflake.ID `gorm:"column:id;primaryKey"`
	CreatedAt time.Time    `gorm:"column:created_at;"`
	UpdatedAt time.Time    `gorm:"column:updated_at;"`
	Creator   snowflake.ID `gorm:"column:creator"`
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

func (b *BaseModel) WithCreator(creator snowflake.ID) *BaseModel {
	b.Creator = creator
	return b
}
