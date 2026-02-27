// Package do is the do package for the auth service.
package do

import (
	"time"

	"github.com/aide-family/magicbox/enum"
	"github.com/aide-family/magicbox/hello"
	"github.com/bwmarrin/snowflake"
	"gorm.io/gorm"
)

type User struct {
	ID        uint32         `gorm:"column:id;primaryKey;autoIncrement"`
	UID       snowflake.ID   `gorm:"column:uid;not null;uniqueIndex"`
	CreatedAt time.Time      `gorm:"column:created_at;type:datetime;not null;"`
	UpdatedAt time.Time      `gorm:"column:updated_at;type:datetime;not null;"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;type:datetime;index"`

	Name     string          `gorm:"column:name;type:varchar(100);not null;index:idx__user__name"`
	Nickname string          `gorm:"column:nickname;type:varchar(200);not null;default:''"`
	Email    string          `gorm:"column:email;type:varchar(100);not null;uniqueIndex:idx__user__email;default:''"`
	Avatar   string          `gorm:"column:avatar;type:varchar(100);not null;default:''"`
	Remark   string          `gorm:"column:remark;type:varchar(100);not null;default:''"`
	Status   enum.UserStatus `gorm:"column:status;type:tinyint;not null;default:0"`
}

func (User) TableName() string {
	return "users"
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.Status <= 0 {
		u.Status = enum.UserStatus_ACTIVE
	}
	node, err := snowflake.NewNode(hello.NodeID())
	if err != nil {
		return err
	}
	u.UID = node.Generate()
	return nil
}

type OAuth2User struct {
	ID        uint32       `gorm:"column:id;primaryKey;autoIncrement"`
	CreatedAt time.Time    `gorm:"column:created_at;type:datetime;not null;"`
	UpdatedAt time.Time    `gorm:"column:updated_at;type:datetime;not null;"`
	OpenID    string       `gorm:"column:open_id;type:varchar(100);not null;uniqueIndex:idx__oauth2_user__app__open_id"`
	Name      string       `gorm:"column:name;type:varchar(100);not null;default:''"`
	Nickname  string       `gorm:"column:nickname;type:varchar(100);not null;default:''"`
	Remark    string       `gorm:"column:remark;type:varchar(200);not null;default:''"`
	Email     string       `gorm:"column:email;type:varchar(100);not null;default:''"`
	Avatar    string       `gorm:"column:avatar;type:varchar(100);not null;default:''"`
	APP       string       `gorm:"column:app;type:varchar(100);not null;uniqueIndex:idx__oauth2_user__app__open_id"`
	Raw       []byte       `gorm:"column:raw;type:json;"`
	UID       snowflake.ID `gorm:"column:user_id;not null;index:idx__oauth2_user__user_uid"`
	User      *User        `gorm:"foreignKey:UID;references:UID"`
}

func (OAuth2User) TableName() string {
	return "user_oauth2s"
}

func (u *OAuth2User) BeforeCreate(tx *gorm.DB) error {
	node, err := snowflake.NewNode(hello.NodeID())
	if err != nil {
		return err
	}
	u.UID = node.Generate()
	return nil
}
