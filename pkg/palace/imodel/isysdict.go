package imodel

import (
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
	"gorm.io/plugin/soft_delete"
)

// IDict 系统字典通用接口
type IDict interface {
	// GetID 字典ID
	GetID() uint32
	// GetName 字典名称
	GetName() string
	// GetValue 字典键值
	GetValue() string
	// GetDictType 字典类型
	GetDictType() vobj.DictType
	// GetColorType 字典颜色
	GetColorType() string
	// GetCSSClass 字典样式
	GetCSSClass() string
	// GetIcon 字典图标
	GetIcon() string
	// GetImageURL 字典图片
	GetImageURL() string
	// GetStatus 字典状态
	GetStatus() vobj.Status
	// GetLanguageCode 字典语言
	GetLanguageCode() string
	// GetRemark 字典备注
	GetRemark() string
	// GetCreatedAt 创建时间
	GetCreatedAt() *types.Time
	// GetUpdatedAt 更新时间
	GetUpdatedAt() *types.Time
	// GetCreatorID 创建者ID
	GetCreatorID() uint32
	// GetDeletedAt 软删除时间戳
	GetDeletedAt() soft_delete.DeletedAt
}
