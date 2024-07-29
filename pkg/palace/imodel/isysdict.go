package imodel

import (
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
	"gorm.io/plugin/soft_delete"
)

type IDict interface {
	GetID() uint32
	GetName() string
	GetValue() string
	GetDictType() vobj.DictType
	GetColorType() string
	GetCssClass() string
	GetIcon() string
	GetImageUrl() string
	GetStatus() vobj.Status
	GetLanguageCode() string
	GetRemark() string
	GetCreatedAt() *types.Time
	GetUpdatedAt() *types.Time
	GetCreatorID() uint32
	GetDeletedAt() soft_delete.DeletedAt
}
