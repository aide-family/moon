package imodel

import (
	"github.com/aide-family/moon/pkg/util/types"
	"gorm.io/plugin/soft_delete"
)

// IBaseModel 基础模型
type IBaseModel interface {
	// GetCreatedAt 获取创建时间
	GetCreatedAt() *types.Time
	// GetUpdatedAt 获取更新时间
	GetUpdatedAt() *types.Time
	// GetDeletedAt 获取删除时间
	GetDeletedAt() soft_delete.DeletedAt
	// GetCreatorID 获取创建者ID
	GetCreatorID() uint32
}

// IEasyModel 简单模型
type IEasyModel interface {
	// GetID 获取ID
	GetID() uint32
	// GetCreatedAt 获取创建时间
	GetCreatedAt() *types.Time
	// GetUpdatedAt 获取更新时间
	GetUpdatedAt() *types.Time
	// GetDeletedAt 获取删除时间
	GetDeletedAt() soft_delete.DeletedAt
}

// IAllFieldModel 所有字段模型
type IAllFieldModel interface {
	IEasyModel
	IBaseModel
}
