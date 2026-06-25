package do

import (
	"time"

	"github.com/bwmarrin/snowflake"
)

// SysUserRole 系统用户角色权限关联表，对应数据库表 moon_sys_user_roles
// 联合主键：(user_uid, role_code)，纯关联表，硬删除
type SysUserRole struct {
	// UserUID 用户 UID
	UserUID snowflake.ID `gorm:"column:user_uid;type:bigint;not null;primaryKey" json:"userUID"`
	// RoleCode 角色编码
	RoleCode string `gorm:"column:role_code;type:varchar(64);not null;primaryKey" json:"roleCode"`
	// Creator 授权人 UID
	Creator snowflake.ID `gorm:"column:creator;type:bigint;not null;default:0" json:"creator"`
	// CreatedAt 授权时间
	CreatedAt time.Time `gorm:"column:created_at;type:datetime;not null;autoCreateTime" json:"createdAt"`
}

// TableName 指定表名
func (*SysUserRole) TableName() string {
	return "moon_sys_user_roles"
}
