package do

import (
	"gorm.io/gorm"
	"prometheus-manager/app/prom_server/internal/biz/do/basescopes"
	"prometheus-manager/app/prom_server/internal/biz/vo"
)

const TableNameRole = "sys_roles"

const (
	SysRoleFieldRemark      = "remark"
	SysRoleFieldName        = "name"
	SysRoleFieldStatus      = "status"
	SysRolePreloadFieldUser = "Users"
	SysRolePreloadFieldApis = "Apis"
)

// SysRolePreloadUsers 预加载用户
func SysRolePreloadUsers(userIds ...uint32) basescopes.ScopeMethod {
	return func(db *gorm.DB) *gorm.DB {
		if len(userIds) > 0 {
			return db.Preload(SysRolePreloadFieldUser, basescopes.WhereInColumn(basescopes.BaseFieldID, userIds...))
		}
		return db.Preload(SysRolePreloadFieldUser)
	}
}

// SysRolePreloadApis 预加载api
func SysRolePreloadApis(apiIds ...uint32) basescopes.ScopeMethod {
	return func(db *gorm.DB) *gorm.DB {
		if len(apiIds) > 0 {
			return db.Preload(SysRolePreloadFieldApis, basescopes.WhereInColumn(basescopes.BaseFieldID, apiIds...))
		}
		return db.Preload(SysRolePreloadFieldApis)
	}
}

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
