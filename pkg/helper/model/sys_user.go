package model

import (
	query "github.com/aide-cloud/gorm-normalize"
)

const TableNameSystemUser = "sys_users"

// SysUser 用户表
type SysUser struct {
	query.BaseModel
	Username string     `gorm:"column:username;type:varchar(64);not null;uniqueIndex:idx__username,priority:1;comment:用户名"`
	Nickname string     `gorm:"column:nickname;type:varchar(64);not null;comment:昵称"`
	Password string     `gorm:"column:password;type:varchar(255);not null;comment:密码"`
	Email    string     `gorm:"column:email;type:varchar(64);not null;uniqueIndex:idx__email,priority:1;comment:邮箱"`
	Phone    string     `gorm:"column:phone;type:varchar(64);not null;uniqueIndex:idx__phone,priority:1;comment:手机号"`
	Status   int32      `gorm:"column:status;type:tinyint;not null;default:1;comment:状态"`
	Remark   string     `gorm:"column:remark;type:varchar(255);not null;comment:备注"`
	Avatar   string     `gorm:"column:avatar;type:varchar(255);not null;comment:头像"`
	Salt     string     `gorm:"column:salt;type:varchar(16);not null;comment:盐"`
	Gender   int32      `gorm:"column:gender;type:tinyint;not null;default:0;comment:性别"`
	Roles    []*SysRole `gorm:"many2many:sys_user_roles;comment:用户角色"`
}

// TableName 表名
func (*SysUser) TableName() string {
	return TableNameSystemUser
}
