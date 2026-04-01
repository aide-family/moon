package do

import (
	"time"

	"github.com/aide-family/magicbox/enum"
	"github.com/aide-family/magicbox/safety"
	"github.com/bwmarrin/snowflake"
	"gorm.io/gorm"
)

// SSHCommandAudit stores pending or completed reviews for command create/update proposals.
type SSHCommandAudit struct {
	BaseModel

	DeletedAt       gorm.DeletedAt              `gorm:"column:deleted_at;index"`
	TargetCommandID snowflake.ID                `gorm:"column:target_command_id;index"`
	Kind            enum.SSHCommandAuditKind   `gorm:"column:kind;index;default:0"`
	Status          enum.SSHCommandAuditStatus `gorm:"column:status;index:ssh_command_audit__status;default:0"`
	Name            string                      `gorm:"column:name;size:191"`
	Description     string                      `gorm:"column:description;size:512"`
	Content         string                      `gorm:"column:content;type:text"`
	WorkDir         string                      `gorm:"column:work_dir;size:512"`
	Env             *safety.Map[string, string] `gorm:"column:env;type:json;serializer:json"`
	RejectReason    string                      `gorm:"column:reject_reason;size:1024"`
	Reviewer        snowflake.ID                `gorm:"column:reviewer"`
	ReviewedAt      *time.Time                  `gorm:"column:reviewed_at"`
}

func (SSHCommandAudit) TableName() string {
	return "ssh_command_audits"
}
