package do

import (
	"errors"
	"time"

	"github.com/aide-family/magicbox/enum"
	"github.com/aide-family/magicbox/hello"
	"github.com/bwmarrin/snowflake"
	"gorm.io/gorm"
)

type Member struct {
	ID           uint32            `gorm:"column:id;primaryKey;autoIncrement"`
	UID          snowflake.ID      `gorm:"column:uid;not null;uniqueIndex"`
	CreatedAt    time.Time         `gorm:"column:created_at;type:datetime;not null;"`
	UpdatedAt    time.Time         `gorm:"column:updated_at;type:datetime;not null;"`
	DeletedAt    gorm.DeletedAt    `gorm:"column:deleted_at;type:datetime;index"`
	Creator      snowflake.ID      `gorm:"column:creator;not null;index"`
	NamespaceUID snowflake.ID      `gorm:"column:namespace_uid;not null;index:idx__member__namespace_uid__user_uid"`
	UserUID      snowflake.ID      `gorm:"column:user_uid;not null;index:idx__member__namespace_uid__user_uid"`
	Status       enum.MemberStatus `gorm:"column:status;type:tinyint;not null;default:0"`
	Name         string            `gorm:"column:name;type:varchar(100);not null;default:''"`
	Nickname     string            `gorm:"column:nickname;type:varchar(100);not null;default:''"`
	Avatar       string            `gorm:"column:avatar;type:varchar(100);not null;default:''"`
	Remark       string            `gorm:"column:remark;type:varchar(100);not null;default:''"`
}

func (m *Member) TableName() string {
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
	m.UID = node.Generate()
	if m.Status <= 0 {
		m.Status = enum.MemberStatus_JOINED
	}
	return nil
}
