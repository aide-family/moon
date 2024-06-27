package model

import (
	"context"
	"github.com/aide-family/moon/pkg/helper/model/base"
	"github.com/aide-family/moon/pkg/vobj"
	"gorm.io/gen"
	"gorm.io/gorm"
)

// 字典数据表
const TableNameSysDict = "sys_dict"

type SysDict struct {
	Name         string        `gorm:"column:name;type:varchar(100);not null;uniqueIndex:idx__p__name__dict,priority:1;comment:字典名称"`
	Value        string        `gorm:"column:value;type:varchar(100);not null;default:'';comment:字典键值"`
	DictType     vobj.DictType `gorm:"column:dict_type;type:tinyint;not null;uniqueIndex:idx__p__name__dict,priority:2;index:idx__dict,priority:1;comment:字典类型"`
	ColorType    string        `gorm:"column:color_type;type:varchar(32);not null;default:warning;comment:颜色类型"`
	CssClass     string        `gorm:"column:css_class;type:varchar(100);not null;default:#165DFF;comment:css 样式"`
	Icon         string        `gorm:"column:icon;type:varchar(500);default:'';comment:图标"`
	ImageUrl     string        `gorm:"column:image_url;type:varchar(500);default:'';comment:图片url"`
	Status       vobj.Status   `gorm:"column:status;type:tinyint;not null;default:1;comment:状态 1：开启 2:关闭"`
	LanguageCode string        `gorm:"column:language_code;type:varchar(10);not null;default:zh;comment:语言：zh:中文 en:英文"`
	Remark       string        `gorm:"column:remark;type:varchar(500);not null;comment:字典备注"`
	base.BaseModel
	base.BaseModelID
}

// Create func
func (c *SysDict) Create(ctx context.Context, tx *gorm.DB) error {
	return tx.WithContext(ctx).Create(c).Error
}

// Update func
func (c *SysDict) Update(ctx context.Context, tx *gorm.DB, conds []gen.Condition) error {
	return tx.WithContext(ctx).Model(c).Where(conds).Updates(c).Error
}

// Delete func
func (c *SysDict) Delete(ctx context.Context, tx *gorm.DB, conds []gen.Condition) error {
	return tx.WithContext(ctx).Where(conds).Delete(c).Error
}

// TableName SysDict's table name
func (*SysDict) TableName() string {
	return TableNameSysDict
}
