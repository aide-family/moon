package do

import (
	"github.com/aide-family/magicbox/safety"
	"gorm.io/gorm"
)

// SSHCommand is an approved SSH command template stored after audit approval.
type SSHCommand struct {
	BaseModel

	DeletedAt   gorm.DeletedAt              `gorm:"column:deleted_at;index"`
	Name        string                      `gorm:"column:name;size:191;uniqueIndex:ssh_command__name"`
	Description string                      `gorm:"column:description;size:512"`
	Content     string                      `gorm:"column:content;type:text"`
	WorkDir     string                      `gorm:"column:work_dir;size:512"`
	Env         *safety.Map[string, string] `gorm:"column:env;type:json;serializer:json"`
	Disabled    bool                        `gorm:"column:disabled;default:false"`
}

func (SSHCommand) TableName() string {
	return "ssh_commands"
}
