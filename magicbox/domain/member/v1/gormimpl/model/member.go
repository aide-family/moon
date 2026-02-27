// Package model is the model for the member service.
package model

import (
	"errors"
	"time"

	"github.com/bwmarrin/snowflake"
	"gorm.io/gorm"

	"github.com/aide-family/magicbox/enum"
	"github.com/aide-family/magicbox/hello"
)

type Member struct {
	ID           snowflake.ID      `gorm:"column:id;not null;primaryKey"`
	CreatedAt    time.Time         `gorm:"column:created_at;not null;"`
	UpdatedAt    time.Time         `gorm:"column:updated_at;not null;"`
	DeletedAt    gorm.DeletedAt    `gorm:"column:deleted_at;index"`
	Creator      snowflake.ID      `gorm:"column:creator;not null;index"`
	NamespaceUID snowflake.ID      `gorm:"column:namespace_uid;not null;uniqueIndex:idx__member__namespace_uid__user_uid"`
	UserUID      snowflake.ID      `gorm:"column:user_uid;not null;uniqueIndex:idx__member__namespace_uid__user_uid"`
	Status       enum.MemberStatus `gorm:"column:status;not null;default:1"`
	Name         string            `gorm:"column:name;size:191;not null;default:''"`
	Nickname     string            `gorm:"column:nickname;size:191;not null;default:''"`
	Avatar       string            `gorm:"column:avatar;size:191;not null;default:''"`
	Remark       string            `gorm:"column:remark;size:191;not null;default:''"`
	Email        string            `gorm:"column:email;size:191;not null;default:''"`
	Phone        string            `gorm:"column:phone;size:191;not null;default:''"`
}

func (Member) TableName() string {
	return "members"
}

func (m *Member) BeforeCreate(tx *gorm.DB) error {
	if m.Creator == 0 {
		return errors.New("creator is required")
	}
	node, err := snowflake.NewNode(hello.NodeID())
	if err != nil {
		return err
	}
	m.ID = node.Generate()
	if m.Status <= 0 {
		m.Status = enum.MemberStatus_JOINED
	}
	return nil
}
