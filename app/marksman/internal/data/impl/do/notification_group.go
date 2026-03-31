package do

import (
	"github.com/aide-family/magicbox/enum"
	"github.com/aide-family/magicbox/safety"
	"github.com/bwmarrin/snowflake"
	"gorm.io/gorm"
)

const (
	TableNameNotificationGroup = "notification_groups"
)

type NotificationGroup struct {
	BaseModel
	DeletedAt    gorm.DeletedAt                     `gorm:"column:deleted_at;uniqueIndex:idx__notification_groups__namespace_uid__deleted_at__name"`
	NamespaceUID snowflake.ID                       `gorm:"column:namespace_uid;uniqueIndex:idx__notification_groups__namespace_uid__deleted_at__name"`
	Name         string                             `gorm:"column:name;type:varchar(100);uniqueIndex:idx__notification_groups__namespace_uid__deleted_at__name"`
	Remark       string                             `gorm:"column:remark;type:varchar(100);default:''"`
	Metadata     *safety.Map[string, string]        `gorm:"column:metadata;type:json;"`
	Status       enum.GlobalStatus                  `gorm:"column:status;type:tinyint;default:0"`
	Members      *safety.Slice[*NotificationMember] `gorm:"column:members;type:json;"`
	Webhooks     *safety.Slice[int64]               `gorm:"column:webhooks;type:json;"`
	Templates    *safety.Slice[int64]               `gorm:"column:templates;type:json;"`
	EmailConfigs *safety.Slice[int64]               `gorm:"column:email_configs;type:json;"`
}

func (NotificationGroup) TableName() string {
	return TableNameNotificationGroup
}

func (n *NotificationGroup) WithNamespace(namespace snowflake.ID) *NotificationGroup {
	n.NamespaceUID = namespace
	return n
}

type NotificationMember struct {
	MemberUID int64 `json:"member_uid"`
	IsEmail   bool  `json:"is_email"`
	IsPhone   bool  `json:"is_phone"`
}
