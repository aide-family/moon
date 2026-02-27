package do

import (
	"github.com/bwmarrin/snowflake"
	"gorm.io/gorm"

	"github.com/aide-family/magicbox/enum"
	"github.com/aide-family/magicbox/safety"
	"github.com/aide-family/magicbox/strutil"
)

type RecipientGroup struct {
	BaseModel
	DeletedAt    gorm.DeletedAt              `gorm:"column:deleted_at;uniqueIndex:recipient_group__namespace_uid__name"`
	NamespaceUID snowflake.ID                `gorm:"column:namespace_uid;uniqueIndex:recipient_group__namespace_uid__name"`
	Name         string                      `gorm:"column:name;size:191;uniqueIndex:recipient_group__namespace_uid__name"`
	Metadata     *safety.Map[string, string] `gorm:"column:metadata;type:json;"`
	Status       enum.GlobalStatus           `gorm:"column:status;default:0"`
	Templates    []*Template                 `gorm:"many2many:recipient_group__templates;"`
	EmailConfigs []*EmailConfig              `gorm:"many2many:recipient_group__email_configs;"`
	Webhooks     []*WebhookConfig            `gorm:"many2many:recipient_group__webhooks;"`
	Members      []*RecipientMember          `gorm:"many2many:recipient_group__members;"`
}

func (RecipientGroup) TableName() string {
	return "recipient_groups"
}

type RecipientMember struct {
	BaseModel

	NamespaceUID snowflake.ID          `gorm:"column:namespace_uid;uniqueIndex:recipient_member__namespace_uid__user_uid"`
	UserUID      snowflake.ID          `gorm:"column:user_uid;uniqueIndex:recipient_member__namespace_uid__user_uid"`
	Email        strutil.EncryptString `gorm:"column:email"`
	Phone        strutil.EncryptString `gorm:"column:phone"`
	Status       enum.MemberStatus     `gorm:"column:status;default:1"`
	Name         string                `gorm:"column:name"`
	Nickname     string                `gorm:"column:nickname"`
	Avatar       string                `gorm:"column:avatar"`
	Remark       string                `gorm:"column:remark"`
}

func (RecipientMember) TableName() string {
	return "recipient_members"
}
