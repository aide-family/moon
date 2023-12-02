package model

import (
	query "github.com/aide-cloud/gorm-normalize"
	"gorm.io/gorm/schema"
)

var _ schema.Tabler = (*SysApi)(nil)

const TableNameSysApi = "sys_apis"

// SysApi 系统api
type SysApi struct {
	query.BaseModel
	Name   string `gorm:"column:name;type:varchar(64);not null;uniqueIndex:idx__name,priority:1;comment:api名称"`
	Path   string `gorm:"column:path;type:varchar(255);not null;uniqueIndex:idx__path,priority:1;comment:api路径"`
	Method string `gorm:"column:method;type:varchar(16);not null;comment:请求方法"`
	Status int32  `gorm:"column:status;type:tinyint;not null;default:1;comment:状态"`
	Remark string `gorm:"column:remark;type:varchar(255);not null;comment:备注"`
}

// TableName 表名
func (SysApi) TableName() string {
	return TableNameSysApi
}
