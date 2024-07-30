package model

import (
	"encoding/json"

	"github.com/aide-family/moon/pkg/palace/imodel"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"

	"gorm.io/plugin/soft_delete"
)

var _ imodel.IDict = (*SysDict)(nil)

// tableNameSysDict 字典数据表
const tableNameSysDict = "sys_dict"

// SysDict 字典数据
type SysDict struct {
	AllFieldModel
	Name         string        `gorm:"column:name;type:varchar(100);not null;uniqueIndex:idx__p__name__dict,priority:1;comment:字典名称"`
	Value        string        `gorm:"column:value;type:varchar(100);not null;default:'';comment:字典键值"`
	DictType     vobj.DictType `gorm:"column:dict_type;type:tinyint;not null;uniqueIndex:idx__p__name__dict,priority:2;index:idx__dict,priority:1;comment:字典类型"`
	ColorType    string        `gorm:"column:color_type;type:varchar(32);not null;default:warning;comment:颜色类型"`
	CSSClass     string        `gorm:"column:css_class;type:varchar(100);not null;default:#165DFF;comment:css 样式"`
	Icon         string        `gorm:"column:icon;type:varchar(500);default:'';comment:图标"`
	ImageURL     string        `gorm:"column:image_url;type:varchar(500);default:'';comment:图片url"`
	Status       vobj.Status   `gorm:"column:status;type:tinyint;not null;default:1;comment:状态 1：开启 2:关闭"`
	LanguageCode string        `gorm:"column:language_code;type:varchar(10);not null;default:zh-CN;comment:语言：zh-CN:中文 en-US:英文"`
	Remark       string        `gorm:"column:remark;type:varchar(500);not null;comment:字典备注"`
}

// GetDeletedAt get deleted at
func (c *SysDict) GetDeletedAt() soft_delete.DeletedAt {
	if types.IsNil(c) {
		return 0
	}
	return c.DeletedAt
}

// GetID get id
func (c *SysDict) GetID() uint32 {
	if types.IsNil(c) {
		return 0
	}
	return c.ID
}

// GetCreatedAt get created at
func (c *SysDict) GetCreatedAt() *types.Time {
	if types.IsNil(c) {
		return &types.Time{}
	}
	return &c.CreatedAt
}

// GetUpdatedAt get updated at
func (c *SysDict) GetUpdatedAt() *types.Time {
	if types.IsNil(c) {
		return &types.Time{}
	}
	return &c.UpdatedAt
}

// GetCreatorID get creator id
func (c *SysDict) GetCreatorID() uint32 {
	if types.IsNil(c) {
		return 0
	}
	return c.CreatorID
}

// GetValue get value
func (c *SysDict) GetValue() string {
	if types.IsNil(c) {
		return ""
	}
	return c.Value
}

// GetDictType get dict type
func (c *SysDict) GetDictType() vobj.DictType {
	if types.IsNil(c) {
		return vobj.DictTypeUnknown
	}
	return c.DictType
}

// GetColorType get color type
func (c *SysDict) GetColorType() string {
	if types.IsNil(c) {
		return ""
	}
	return c.ColorType
}

// GetCSSClass get css class
func (c *SysDict) GetCSSClass() string {
	if types.IsNil(c) {
		return ""
	}
	return c.CSSClass
}

// GetIcon get icon
func (c *SysDict) GetIcon() string {
	if types.IsNil(c) {
		return ""
	}
	return c.Icon
}

// GetImageURL get image url
func (c *SysDict) GetImageURL() string {
	if types.IsNil(c) {
		return ""
	}
	return c.ImageURL
}

// GetStatus get status
func (c *SysDict) GetStatus() vobj.Status {
	if types.IsNil(c) {
		return vobj.StatusUnknown
	}
	return c.Status
}

// GetLanguageCode get language code
func (c *SysDict) GetLanguageCode() string {
	if types.IsNil(c) {
		return ""
	}
	return c.LanguageCode
}

// GetRemark get remark
func (c *SysDict) GetRemark() string {
	if types.IsNil(c) {
		return ""
	}
	return c.Remark
}

// GetName get name
func (c *SysDict) GetName() string {
	if types.IsNil(c) {
		return ""
	}
	return c.Name
}

// String json string
func (c *SysDict) String() string {
	bs, _ := json.Marshal(c)
	return string(bs)
}

// TableName SysDict's table name
func (*SysDict) TableName() string {
	return tableNameSysDict
}
