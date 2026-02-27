// Package model is the model package for the auth service.
package model

import (
	"time"

	"github.com/aide-family/magicbox/enum"
	"github.com/aide-family/magicbox/hello"
	"github.com/bwmarrin/snowflake"
	"gorm.io/gorm"
)

func Models() []any {
	return []any{
		&User{},
		&OAuth2User{},
	}
}

type User struct {
	ID        snowflake.ID    `gorm:"column:id;not null;primaryKey"`
	CreatedAt time.Time       `gorm:"column:created_at;not null;"`
	UpdatedAt time.Time       `gorm:"column:updated_at;not null;"`
	DeletedAt gorm.DeletedAt  `gorm:"column:deleted_at;uniqueIndex:idx__user__email__deleted_at"`
	Name      string          `gorm:"column:name;type:varchar(100);not null;default:''"`
	Nickname  string          `gorm:"column:nickname;type:varchar(100);not null;default:''"`
	Email     string          `gorm:"column:email;type:varchar(100);not null;uniqueIndex:idx__user__email__deleted_at"`
	Avatar    string          `gorm:"column:avatar;type:varchar(100);not null;default:''"`
	Remark    string          `gorm:"column:remark;type:varchar(100);not null;default:''"`
	Status    enum.UserStatus `gorm:"column:status;not null;default:1"`
}

func (User) TableName() string {
	return "users"
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == 0 {
		node, err := snowflake.NewNode(hello.NodeID())
		if err != nil {
			return err
		}
		u.ID = node.Generate()
	}
	return nil
}

type OAuth2User struct {
	ID        uint32       `gorm:"column:id;primaryKey;autoIncrement"`
	CreatedAt time.Time    `gorm:"column:created_at;not null;"`
	UpdatedAt time.Time    `gorm:"column:updated_at;not null;"`
	OpenID    string       `gorm:"column:open_id;type:varchar(100);not null;uniqueIndex:idx__oauth2_user__app__open_id"`
	Name      string       `gorm:"column:name;type:varchar(100);not null;default:''"`
	Email     string       `gorm:"column:email;type:varchar(100);not null;default:'';index:idx__oauth2_user__email"`
	Avatar    string       `gorm:"column:avatar;type:varchar(100);not null;default:''"`
	APP       string       `gorm:"column:app;type:varchar(100);not null;uniqueIndex:idx__oauth2_user__app__open_id"`
	Raw       []byte       `gorm:"column:raw;type:json;"`
	UserID    snowflake.ID `gorm:"column:user_id;not null;index:idx__oauth2_user__user_id"`
}

func (OAuth2User) TableName() string {
	return "user_oauth2s"
}
