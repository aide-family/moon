package model

import (
	query "github.com/aide-cloud/gorm-normalize"
)

const TableNameSystemUser = "sys_users"

// SysUser 用户表
type SysUser struct {
	query.BaseModel
	Username string `gorm:"column:username;type:varchar(64);not null;uniqueIndex:idx_username,priority:1;comment:用户名" json:"username"`
	Password string `gorm:"column:password;type:varchar(255);not null;comment:密码" json:"password"`
	Email    string `gorm:"column:email;type:varchar(64);not null;uniqueIndex:idx_email,priority:1;comment:邮箱" json:"email"`
	Phone    string `gorm:"column:phone;type:varchar(64);not null;uniqueIndex:idx_phone,priority:1;comment:手机号" json:"phone"`
	Status   int32  `gorm:"column:status;type:tinyint;not null;default:1;comment:状态" json:"status"`
	Remark   string `gorm:"column:remark;type:varchar(255);not null;comment:备注" json:"remark"`
	Avatar   string `gorm:"column:avatar;type:varchar(255);not null;comment:头像" json:"avatar"`
	Salt     string `gorm:"column:salt;type:varchar(16);not null;comment:盐" json:"salt"`
}

// TableName 表名
func (*SysUser) TableName() string {
	return TableNameSystemUser
}
