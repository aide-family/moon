package model

import (
	query "github.com/aide-cloud/gorm-normalize"
)

const TableNameSysApi = "sys_apis"

// SysApi 系统api
type SysApi struct {
	query.BaseModel
	Name   string `gorm:"column:name;type:varchar(64);not null;uniqueIndex:idx_name,priority:1;comment:api名称" json:"name"`
	Path   string `gorm:"column:path;type:varchar(255);not null;uniqueIndex:idx_path,priority:1;comment:api路径" json:"path"`
	Method string `gorm:"column:method;type:varchar(16);not null;comment:请求方法" json:"method"`
	Status int32  `gorm:"column:status;type:tinyint;not null;default:1;comment:状态" json:"status"`
	Remark string `gorm:"column:remark;type:varchar(255);not null;comment:备注" json:"remark"`
}

// TableName 表名
func (*SysApi) TableName() string {
	return TableNameSysApi
}
