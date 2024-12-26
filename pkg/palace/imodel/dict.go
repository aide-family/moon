package imodel

import (
	"github.com/aide-family/moon/pkg/plugin/cache"
	"github.com/aide-family/moon/pkg/vobj"
)

// IDict 系统字典通用接口
type IDict interface {
	cache.IObjectSchema
	IAllFieldModel
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
	GetLanguageCode() vobj.Language
	// GetRemark 字典备注
	GetRemark() string
}
