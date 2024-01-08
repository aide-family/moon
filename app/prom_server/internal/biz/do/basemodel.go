package do

import (
	"time"

	"gorm.io/plugin/soft_delete"
)

type BaseModel struct {
	ID        uint32                `gorm:"primary_key" json:"id"`
	CreatedAt time.Time             `gorm:"column:created_at;type:timestamp;not null;default:CURRENT_TIMESTAMP;comment:创建时间" json:"-"`
	UpdatedAt time.Time             `gorm:"column:updated_at;type:timestamp;not null;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP;comment:更新时间" json:"-"`
	DeletedAt soft_delete.DeletedAt `gorm:"column:deleted_at;type:bigint;not null;default:0;" json:"-"`
}
