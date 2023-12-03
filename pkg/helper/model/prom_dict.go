package model

import (
	query "github.com/aide-cloud/gorm-normalize"
	"prometheus-manager/pkg/helper/valueobj"
)

const TableNamePromDict = "prom_dict"

// PromDict 系统的字典管理
type PromDict struct {
	query.BaseModel
	Name     string            `gorm:"column:name;type:varchar(64);not null;uniqueIndex:idx__name__category,priority:1;comment:字典名称"`
	Category valueobj.Category `gorm:"column:category;type:tinyint;not null;uniqueIndex:idx__name__category,priority:2;index:idx__category,priority:1;comment:字典类型"`
	Color    string            `gorm:"column:color;type:varchar(32);not null;default:#165DFF;comment:字典tag颜色"`
	Status   valueobj.Status   `gorm:"column:status;type:tinyint;not null;default:1;comment:状态"`
	Remark   string            `gorm:"column:remark;type:varchar(255);not null;comment:字典备注"`
}

// TableName PromDict table name
func (*PromDict) TableName() string {
	return TableNamePromDict
}
