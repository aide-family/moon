package do

import (
	"time"

	"gorm.io/plugin/soft_delete"
)

type BaseModel struct {
	ID        uint32                `gorm:"primary_key" json:"id"`
	CreatedAt time.Time             `gorm:"column:created_at;type:timestamp;not null;default:CURRENT_TIMESTAMP;comment:创建时间" json:"-"`
	UpdatedAt time.Time             `gorm:"column:updated_at;type:timestamp;not null;default:CURRENT_TIMESTAMP;comment:更新时间" json:"-"`
	DeletedAt soft_delete.DeletedAt `gorm:"column:deleted_at;type:bigint;not null;default:0;" json:"-"`
}

func (l *BaseModel) GetID() uint32 {
	if l == nil {
		return 0
	}
	return l.ID
}

func (l *BaseModel) GetCreatedAt() time.Time {
	if l == nil {
		return time.Time{}
	}
	return l.CreatedAt
}

func (l *BaseModel) GetUpdatedAt() time.Time {
	if l == nil {
		return time.Time{}
	}
	return l.UpdatedAt
}

func (l *BaseModel) GetDeletedAt() soft_delete.DeletedAt {
	if l == nil {
		return 0
	}
	return l.DeletedAt
}
