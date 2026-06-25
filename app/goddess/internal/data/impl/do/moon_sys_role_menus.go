package do

import (
	"time"
)

// SysRoleMenu 系统角色菜单权限关联表，对应数据库表 moon_sys_role_menus
// 联合主键：(roleCode, menuCode)，纯关联表，硬删除
type SysRoleMenu struct {
	// RoleCode 角色编码
	RoleCode string `gorm:"column:role_code;type:varchar(64);not null;primaryKey" json:"roleCode"`
	// MenuCode 菜单编码
	MenuCode string `gorm:"column:menu_code;type:varchar(100);not null;primaryKey" json:"menuCode"`
	// Creator 授权人 UID
	Creator int64 `gorm:"column:creator;type:bigint;not null;default:0" json:"creator"`
	// CreatedAt 授权时间
	CreatedAt time.Time `gorm:"column:created_at;type:datetime;not null;autoCreateTime" json:"createdAt"`
}

// TableName 指定表名
func (*SysRoleMenu) TableName() string {
	return "moon_sys_role_menus"
}
