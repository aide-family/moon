package imodel

import (
	"github.com/aide-family/moon/pkg/util/types"
	"gorm.io/plugin/soft_delete"
)

type IBaseModel interface {
	GetCreatedAt() *types.Time
	GetUpdatedAt() *types.Time
	GetDeletedAt() soft_delete.DeletedAt
	GetCreatorID() uint32
}

type IEasyModel interface {
	GetID() uint32
	GetCreatedAt() *types.Time
	GetUpdatedAt() *types.Time
	GetDeletedAt() soft_delete.DeletedAt
}

type IAllFieldModel interface {
	IEasyModel
	IBaseModel
}
