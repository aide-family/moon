package base

import (
	"gorm.io/plugin/soft_delete"
	"time"
)

type BaseModel struct {
	CreatedAt time.Time             `gorm:"column:created_at;type:timestamp;not null;default:CURRENT_TIMESTAMP;comment:创建时间" json:"-"`
	UpdatedAt time.Time             `gorm:"column:updated_at;type:timestamp;not null;default:CURRENT_TIMESTAMP;comment:更新时间" json:"-"`
	DeletedAt soft_delete.DeletedAt `gorm:"column:deleted_at;type:bigint;not null;default:0;" json:"-"`
}

type BaseModelID struct {
	ID uint32 `gorm:"column:id;type:int unsigned;primaryKey;autoIncrement:true" json:"id"`
}
