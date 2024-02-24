package do

import (
	"prometheus-manager/app/prom_server/internal/biz/vo"
)

const TableNameRole = "sys_roles"

// SysRole 角色表
type SysRole struct {
	BaseModel
	Remark string     `gorm:"column:remark;type:varchar(255);not null;comment:备注"`
	Name   string     `gorm:"column:name;type:varchar(64);not null;uniqueIndex:idx__sr__name,priority:1;comment:角色名称"`
	Status vo.Status  `gorm:"column:status;type:tinyint;not null;default:1;comment:状态"`
	Users  []*SysUser `gorm:"many2many:sys_user_roles;comment:用户角色"`
	Apis   []*SysAPI  `gorm:"many2many:sys_role_apis;comment:角色api"`
}

// TableName 表名
func (*SysRole) TableName() string {
	return TableNameRole
}

// GetUsers 获取用户
func (r *SysRole) GetUsers() []*SysUser {
	if r == nil {
		return nil
	}
	return r.Users
}

// GetApis 获取api
func (r *SysRole) GetApis() []*SysAPI {
	if r == nil {
		return nil
	}
	return r.Apis
}
