package bizmodel

import (
	"encoding/json"

	"github.com/aide-family/moon/pkg/palace/imodel"
	"github.com/aide-family/moon/pkg/palace/model"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"

	"gorm.io/plugin/soft_delete"
)

var _ imodel.IDict = (*SysDict)(nil)

// TableNameSysDict 字典数据表
const TableNameSysDict = "sys_dict"

type SysDict struct {
	model.AllFieldModel

	Name         string        `gorm:"column:name;type:varchar(100);not null;uniqueIndex:idx__p__name__dict,priority:1;comment:字典名称"`
	Value        string        `gorm:"column:value;type:varchar(100);not null;default:'';comment:字典键值"`
	DictType     vobj.DictType `gorm:"column:dict_type;type:tinyint;not null;uniqueIndex:idx__p__name__dict,priority:2;index:idx__dict,priority:1;comment:字典类型"`
	ColorType    string        `gorm:"column:color_type;type:varchar(32);not null;default:warning;comment:颜色类型"`
	CssClass     string        `gorm:"column:css_class;type:varchar(100);not null;default:#165DFF;comment:css 样式"`
	Icon         string        `gorm:"column:icon;type:varchar(500);default:'';comment:图标"`
	ImageUrl     string        `gorm:"column:image_url;type:varchar(500);default:'';comment:图片url"`
	Status       vobj.Status   `gorm:"column:status;type:tinyint;not null;default:1;comment:状态 1：开启 2:关闭"`
	LanguageCode string        `gorm:"column:language_code;type:varchar(10);not null;default:zh-CN;comment:语言：zh-CN:中文 en-US:英文"`
	Remark       string        `gorm:"column:remark;type:varchar(500);not null;comment:字典备注"`
	//DeletedAt    soft_delete.DeletedAt `gorm:"column:deleted_at;type:bigint;not null;uniqueIndex:idx__p__del__dict,priority:3;index:idx__p__name__dict,priority:2;index:idx__dict,priority:1;default:0;" json:"deleted_at"`
}

func (c *SysDict) GetDeletedAt() soft_delete.DeletedAt {
	if types.IsNil(c) {
		return 0
	}
	return c.DeletedAt
}

func (c *SysDict) GetID() uint32 {
	if types.IsNil(c) {
		return 0
	}
	return c.ID
}

func (c *SysDict) GetCreatedAt() *types.Time {
	if types.IsNil(c) {
		return &types.Time{}
	}
	return &c.CreatedAt
}

func (c *SysDict) GetUpdatedAt() *types.Time {
	if types.IsNil(c) {
		return &types.Time{}
	}
	return &c.UpdatedAt
}

func (c *SysDict) GetCreatorID() uint32 {
	if types.IsNil(c) {
		return 0
	}
	return c.CreatorID
}

func (c *SysDict) GetValue() string {
	if types.IsNil(c) {
		return ""
	}
	return c.Value
}

func (c *SysDict) GetDictType() vobj.DictType {
	if types.IsNil(c) {
		return vobj.DictTypeUnknown
	}
	return c.DictType
}

func (c *SysDict) GetColorType() string {
	if types.IsNil(c) {
		return ""
	}
	return c.ColorType
}

func (c *SysDict) GetCssClass() string {
	if types.IsNil(c) {
		return ""
	}
	return c.CssClass
}

func (c *SysDict) GetIcon() string {
	if types.IsNil(c) {
		return ""
	}
	return c.Icon
}

func (c *SysDict) GetImageUrl() string {
	if types.IsNil(c) {
		return ""
	}
	return c.ImageUrl
}

func (c *SysDict) GetStatus() vobj.Status {
	if types.IsNil(c) {
		return vobj.StatusUnknown
	}
	return c.Status
}

func (c *SysDict) GetLanguageCode() string {
	if types.IsNil(c) {
		return ""
	}
	return c.LanguageCode
}

func (c *SysDict) GetRemark() string {
	if types.IsNil(c) {
		return ""
	}
	return c.Remark
}

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
	return TableNameSysDict
}
