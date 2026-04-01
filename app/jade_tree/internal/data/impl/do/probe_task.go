package do

import (
	"github.com/aide-family/magicbox/enum"
	"gorm.io/gorm"
)

type ProbeTask struct {
	BaseModel

	DeletedAt      gorm.DeletedAt    `gorm:"column:deleted_at;index"`
	Type           string            `gorm:"column:type;size:32;index"`
	Host           string            `gorm:"column:host;size:255"`
	Port           string            `gorm:"column:port;size:16"`
	URL            string            `gorm:"column:url;size:1024"`
	Name           string            `gorm:"column:name;size:255;index"`
	Status         enum.GlobalStatus `gorm:"column:status;index;default:1"`
	TimeoutSeconds int32             `gorm:"column:timeout_seconds;default:5"`
}

func (ProbeTask) TableName() string {
	return "probe_tasks"
}
