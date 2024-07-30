package model

import (
	"context"

	"github.com/aide-family/moon/pkg/helper/middleware"
	"github.com/aide-family/moon/pkg/util/types"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"
)

// BaseModel gorm基础模型
type BaseModel struct {
	ctx context.Context `gorm:"-"`

	CreatedAt types.Time            `gorm:"column:created_at;type:timestamp;not null;default:CURRENT_TIMESTAMP;comment:创建时间" json:"created_at"`
	UpdatedAt types.Time            `gorm:"column:updated_at;type:timestamp;not null;default:CURRENT_TIMESTAMP;comment:更新时间" json:"updated_at"`
	DeletedAt soft_delete.DeletedAt `gorm:"column:deleted_at;type:bigint;not null;default:0;" json:"deleted_at"`

	// 创建人
	CreatorID uint32 `gorm:"column:creator;type:int unsigned;not null;comment:创建者" json:"creator_id"`
}

// AllFieldModel gorm包含所有字段的模型
type AllFieldModel struct {
	ID uint32 `gorm:"column:id;type:int unsigned;primaryKey;autoIncrement:true" json:"id"`
	BaseModel
}

// EasyModel gorm包含基础字段的模型
type EasyModel struct {
	ID        uint32                `gorm:"column:id;type:int unsigned;primaryKey;autoIncrement:true" json:"id"`
	CreatedAt types.Time            `gorm:"column:created_at;type:timestamp;not null;default:CURRENT_TIMESTAMP;comment:创建时间" json:"created_at"`
	UpdatedAt types.Time            `gorm:"column:updated_at;type:timestamp;not null;default:CURRENT_TIMESTAMP;comment:更新时间" json:"updated_at"`
	DeletedAt soft_delete.DeletedAt `gorm:"column:deleted_at;type:bigint;not null;default:0;" json:"deleted_at"`
}

// WithContext 获取上下文
func (u *BaseModel) WithContext(ctx context.Context) *BaseModel {
	u.ctx = ctx
	return u
}

// BeforeCreate 创建前的hook
func (u *BaseModel) BeforeCreate(_ *gorm.DB) (err error) {
	if u.ctx == nil {
		return
	}
	claims, ok := middleware.ParseJwtClaims(u.ctx)
	if !ok {
		return
	}
	u.CreatorID = claims.GetUser()
	return
}

// GetContext 获取上下文
func (u *BaseModel) GetContext() context.Context {
	if types.IsNil(u.ctx) {
		return context.TODO()
	}
	return u.ctx
}
