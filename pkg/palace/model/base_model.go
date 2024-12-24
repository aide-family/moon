package model

import (
	"context"

	"github.com/aide-family/moon/pkg/helper/middleware"
	"github.com/aide-family/moon/pkg/palace/imodel"
	"github.com/aide-family/moon/pkg/util/types"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"
)

var _ imodel.IBaseModel = (*BaseModel)(nil)

// BaseModel gorm基础模型
type BaseModel struct {
	ctx context.Context `gorm:"-"`

	CreatedAt *types.Time           `gorm:"column:created_at;type:timestamp;not null;default:CURRENT_TIMESTAMP;comment:创建时间" json:"created_at,omitempty"`
	UpdatedAt *types.Time           `gorm:"column:updated_at;type:timestamp;not null;default:CURRENT_TIMESTAMP;comment:更新时间" json:"updated_at,omitempty"`
	DeletedAt soft_delete.DeletedAt `gorm:"column:deleted_at;type:bigint;not null;default:0;" json:"deleted_at,omitempty"`

	// 创建人
	CreatorID uint32 `gorm:"column:creator;type:int unsigned;not null;comment:创建者" json:"creator_id,omitempty"`
}

// GetCreatedAt 获取创建时间
func (u *BaseModel) GetCreatedAt() *types.Time {
	if types.IsNil(u) {
		return &types.Time{}
	}
	return u.CreatedAt
}

// GetUpdatedAt 获取更新时间
func (u *BaseModel) GetUpdatedAt() *types.Time {
	if types.IsNil(u) {
		return &types.Time{}
	}

	return u.UpdatedAt
}

// GetDeletedAt 获取删除时间
func (u *BaseModel) GetDeletedAt() soft_delete.DeletedAt {
	if types.IsNil(u) {
		return 0
	}

	return u.DeletedAt
}

// GetCreatorID 获取创建者ID
func (u *BaseModel) GetCreatorID() uint32 {
	if types.IsNil(u) {
		return 0
	}

	return u.CreatorID
}

var _ imodel.IAllFieldModel = (*AllFieldModel)(nil)

// AllFieldModel gorm包含所有字段的模型
type AllFieldModel struct {
	ID uint32 `gorm:"column:id;primaryKey;autoIncrement" json:"id,omitempty"`
	BaseModel
}

// GetID 获取ID
func (a *AllFieldModel) GetID() uint32 {
	if types.IsNil(a) {
		return 0
	}
	return a.ID
}

var _ imodel.IEasyModel = (*EasyModel)(nil)

// EasyModel gorm包含基础字段的模型
type EasyModel struct {
	ID        uint32                `gorm:"column:id;primaryKey;autoIncrement" json:"id,omitempty"`
	CreatedAt *types.Time           `gorm:"column:created_at;type:timestamp;not null;default:CURRENT_TIMESTAMP;comment:创建时间" json:"created_at,omitempty"`
	UpdatedAt *types.Time           `gorm:"column:updated_at;type:timestamp;not null;default:CURRENT_TIMESTAMP;comment:更新时间" json:"updated_at,omitempty"`
	DeletedAt soft_delete.DeletedAt `gorm:"column:deleted_at;type:bigint;not null;default:0;" json:"deleted_at,omitempty"`
}

// GetID 获取ID
func (e *EasyModel) GetID() uint32 {
	if types.IsNil(e) {
		return 0
	}

	return e.ID
}

// GetCreatedAt 获取创建时间
func (e *EasyModel) GetCreatedAt() *types.Time {
	if types.IsNil(e) {
		return &types.Time{}
	}

	return e.CreatedAt
}

// GetUpdatedAt 获取更新时间
func (e *EasyModel) GetUpdatedAt() *types.Time {
	if types.IsNil(e) {
		return &types.Time{}
	}

	return e.UpdatedAt
}

// GetDeletedAt 获取删除时间
func (e *EasyModel) GetDeletedAt() soft_delete.DeletedAt {
	if types.IsNil(e) {
		return 0
	}

	return e.DeletedAt
}

// WithContext 设置上下文
func (u *BaseModel) WithContext(ctx context.Context) *BaseModel {
	u.ctx = ctx
	return u
}

// BeforeCreate 创建前的hook
func (u *BaseModel) BeforeCreate(_ *gorm.DB) (err error) {
	if u.ctx == nil {
		return
	}

	u.CreatorID = middleware.GetUserID(u.GetContext())
	return
}

// GetContext 获取上下文
func (u *BaseModel) GetContext() context.Context {
	if types.IsNil(u.ctx) {
		return context.TODO()
	}
	return u.ctx
}
