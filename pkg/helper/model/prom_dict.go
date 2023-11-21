package model

import (
	query "github.com/aide-cloud/gorm-normalize"
)

const TableNamePromDict = "prom_dict"

// PromDict mapped from table <prom_dict>
type PromDict struct {
	query.BaseModel
	Name     string `gorm:"column:name;type:varchar(64);not null;uniqueIndex:idx__name__category,priority:1;comment:字典名称" json:"name"`                                    // 字典名称
	Category int32  `gorm:"column:category;type:tinyint;not null;uniqueIndex:idx__name__category,priority:2;index:idx__category,priority:1;comment:字典类型" json:"category"` // 字典类型
	Color    string `gorm:"column:color;type:varchar(32);not null;default:#165DFF;comment:字典tag颜色" json:"color"`                                                          // 字典tag颜色
	Status   int32  `gorm:"column:status;type:tinyint;not null;default:1;comment:状态" json:"status"`                                                                       // 状态
	Remark   string `gorm:"column:remark;type:varchar(255);not null;comment:字典备注" json:"remark"`                                                                          // 字典备注 // 删除时间
}

// TableName PromDict table name
func (*PromDict) TableName() string {
	return TableNamePromDict
}
