package do

import (
	"github.com/aide-family/magicbox/enum"
	"github.com/aide-family/magicbox/safety"
	"github.com/aide-family/magicbox/strutil"
	"github.com/bwmarrin/snowflake"
	"gorm.io/gorm"
)

type WebhookConfig struct {
	BaseModel

	DeletedAt    gorm.DeletedAt              `gorm:"column:deleted_at;uniqueIndex:webhook_config__namespace_uid__name"`
	NamespaceUID snowflake.ID                `gorm:"column:namespace_uid;uniqueIndex:webhook_config__namespace_uid__name"`
	Name         string                      `gorm:"column:name;size:191;uniqueIndex:webhook_config__namespace_uid__name"`
	App          enum.WebhookAPP             `gorm:"column:app;default:0"`
	URL          string                      `gorm:"column:url;size:191;uniqueIndex"`
	Method       enum.HTTPMethod             `gorm:"column:method;default:0"`
	Headers      *safety.Map[string, string] `gorm:"column:headers;type:json;"`
	Secret       strutil.EncryptString       `gorm:"column:secret;"`
	Status       enum.GlobalStatus           `gorm:"column:status;default:0"`
}

func (WebhookConfig) TableName() string {
	return "webhooks"
}
