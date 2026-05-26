package do

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/bwmarrin/snowflake"
	"gorm.io/gorm"

	"github.com/aide-family/magicbox/enum"
	"github.com/aide-family/magicbox/safety"
	"github.com/aide-family/magicbox/strutil"
)

const (
	TableNameAlertRecord       = "alert_records"
	TableNameAlertSubscription = "alert_subscriptions"
)

type AlertSubscriptionMember struct {
	MemberUID int64 `json:"member_uid"`
	IsEmail   bool  `json:"is_email"`
	IsSMS     bool  `json:"is_sms"`
	IsPhone   bool  `json:"is_phone"`
}

type AlertSubscriptionMembers []AlertSubscriptionMember

func (c AlertSubscriptionMembers) Value() (driver.Value, error) {
	if c == nil {
		return nil, nil
	}
	return json.Marshal(c)
}

func (c *AlertSubscriptionMembers) Scan(value any) error {
	if value == nil {
		*c = nil
		return nil
	}
	var b []byte
	switch v := value.(type) {
	case []byte:
		b = v
	case string:
		b = []byte(v)
	default:
		return errors.New("alert_subscription_members: expected []byte or string")
	}
	if len(b) == 0 {
		*c = nil
		return nil
	}
	return json.Unmarshal(b, c)
}

type AlertRecord struct {
	BaseModel
	NamespaceUID snowflake.ID                `gorm:"column:namespace_uid;index:idx_alert_records_namespace_uid_created_at"`
	Source       string                      `gorm:"column:source;size:100"`
	Receiver     string                      `gorm:"column:receiver;size:200"`
	Status       string                      `gorm:"column:status;size:32;index:idx_alert_records_status"`
	Fingerprint  string                      `gorm:"column:fingerprint;size:191;index:idx_alert_records_fingerprint"`
	GroupKey     string                      `gorm:"column:group_key;size:500"`
	StartsAt     time.Time                   `gorm:"column:starts_at;index:idx_alert_records_starts_at"`
	EndsAt       *time.Time                  `gorm:"column:ends_at"`
	GeneratorURL string                      `gorm:"column:generator_url;size:1000"`
	Labels       *safety.Map[string, string] `gorm:"column:labels;type:json;"`
	Annotations  *safety.Map[string, string] `gorm:"column:annotations;type:json;"`
	Raw          strutil.EncryptString       `gorm:"column:raw"`
}

func (AlertRecord) TableName() string {
	return TableNameAlertRecord
}

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
