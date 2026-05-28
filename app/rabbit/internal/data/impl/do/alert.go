package do

import (
	"github.com/bwmarrin/snowflake"
	"gorm.io/gorm"

	"github.com/aide-family/magicbox/enum"
	"github.com/aide-family/magicbox/safety"
)

const (
	TableNameAlertSubscription = "alert_subscriptions"
)

type AlertSubscriptionMember = NotificationMember
type AlertSubscriptionMembers = NotificationMembers

type AlertSubscription struct {
	BaseModel
	DeletedAt                gorm.DeletedAt                `gorm:"column:deleted_at;uniqueIndex:alert_subscription__namespace_uid__name"`
	NamespaceUID             snowflake.ID                  `gorm:"column:namespace_uid;uniqueIndex:alert_subscription__namespace_uid__name"`
	Name                     string                        `gorm:"column:name;size:191;uniqueIndex:alert_subscription__namespace_uid__name"`
	Remark                   string                        `gorm:"column:remark;size:500"`
	Labels                   *safety.Map[string, string]   `gorm:"column:labels;type:json;"`
	ExcludeLabels            *safety.Map[string, string]   `gorm:"column:exclude_labels;type:json;"`
	RecipientGroupUIDs       *safety.Slice[int64]          `gorm:"column:recipient_group_uids;type:json;"`
	Members                  AlertSubscriptionMembers      `gorm:"column:members;type:json;"`
	DirectMemberEmailConfig  snowflake.ID                  `gorm:"column:direct_member_email_config_uid"`
	DirectMemberTemplateUID  snowflake.ID                  `gorm:"column:direct_member_template_uid"`
	Status                   enum.GlobalStatus             `gorm:"column:status;default:0"`
}

func (AlertSubscription) TableName() string {
	return TableNameAlertSubscription
}
