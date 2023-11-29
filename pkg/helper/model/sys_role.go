package model

import (
	query "github.com/aide-cloud/gorm-normalize"
)

const TableNameRole = "sys_roles"

// SysRole 角色表
type SysRole struct {
	query.BaseModel
	Remark string     `gorm:"column:remark;type:varchar(255);not null;comment:备注" json:"remark"`
	Name   string     `gorm:"column:name;type:varchar(64);not null;uniqueIndex:idx_name,priority:1;comment:角色名称" json:"name"`
	Status int32      `gorm:"column:status;type:tinyint;not null;default:1;comment:状态" json:"status"`
	Users  []*SysUser `gorm:"many2many:sys_user_roles;comment:用户角色" json:"users"`
}

// TableName 表名
func (*SysRole) TableName() string {
	return TableNameRole
}
