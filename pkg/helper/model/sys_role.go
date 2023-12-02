package model

import (
	query "github.com/aide-cloud/gorm-normalize"
)

const TableNameRole = "sys_roles"

// SysRole 角色表
type SysRole struct {
	query.BaseModel
	Remark string     `gorm:"column:remark;type:varchar(255);not null;comment:备注"`
	Name   string     `gorm:"column:name;type:varchar(64);not null;uniqueIndex:idx__name,priority:1;comment:角色名称"`
	Status int32      `gorm:"column:status;type:tinyint;not null;default:1;comment:状态"`
	Users  []*SysUser `gorm:"many2many:sys_user_roles;comment:用户角色"`
}

// TableName 表名
func (*SysRole) TableName() string {
	return TableNameRole
}
