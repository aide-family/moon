// Package model is the model package for the namespace service.
package model

import (
	"errors"
	"time"

	"github.com/aide-family/magicbox/enum"
	"github.com/aide-family/magicbox/hello"
	"github.com/aide-family/magicbox/safety"

	"github.com/bwmarrin/snowflake"
	"gorm.io/gorm"
)

func Models() []any {
	return []any{
		&Namespace{},
	}
}

type Namespace struct {
	ID        snowflake.ID                `gorm:"column:id;not null;primaryKey"`
	CreatedAt time.Time                   `gorm:"column:created_at;not null;"`
	UpdatedAt time.Time                   `gorm:"column:updated_at;not null;"`
	Creator   snowflake.ID                `gorm:"column:creator;not null;index"`
	DeletedAt gorm.DeletedAt              `gorm:"column:deleted_at;uniqueIndex:idx__namespace__name__deleted_at"`
	Name      string                      `gorm:"column:name;type:varchar(100);not null;uniqueIndex:idx__namespace__name__deleted_at"`
	Metadata  *safety.Map[string, string] `gorm:"column:metadata;type:json;"`
	Status    enum.GlobalStatus           `gorm:"column:status;type:int;not null;default:0"`
	Remark    string                      `gorm:"column:remark;type:varchar(1000);not null;default:''"`
}

func (Namespace) TableName() string {
	return "namespaces"
}

func (n *Namespace) WithCreator(creator snowflake.ID) *Namespace {
	n.Creator = creator
	return n
}

func (n *Namespace) BeforeCreate(tx *gorm.DB) (err error) {
	if n.Creator == 0 {
		return errors.New("creator is required")
	}

	node, err := snowflake.NewNode(hello.NodeID())
	if err != nil {
		return err
	}
	n.ID = node.Generate()
	if n.Status <= 0 {
		return errors.New("status is required")
	}
	return
}
