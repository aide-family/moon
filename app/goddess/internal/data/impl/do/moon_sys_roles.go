package do

import (
	"time"

	"gorm.io/gorm"
)

// SysRole 系统角色表，对应数据库表 moon_sys_roles
type SysRole struct {
	// RoleCode 角色编码，业务主键（如 sys:admin）
	RoleCode string `gorm:"column:role_code;type:varchar(64);not null;primaryKey" json:"roleCode"`
	// RoleName 角色名称
	RoleName string `gorm:"column:role_name;type:varchar(100);not null;default:''" json:"roleName"`
	// RoleDataScope 数据权限范围：1-公司级 2-部门级 3-个人级
	RoleDataScope int32 `gorm:"column:role_data_scope;type:tinyint;not null;default:1" json:"roleDataScope"`
	// RoleStatus 角色状态：0-未知 1-启用 2-禁用
	RoleStatus int32 `gorm:"column:role_status;type:tinyint;not null;default:1" json:"roleStatus"`
	// RoleSortOrder 排序序号
	RoleSortOrder int32 `gorm:"column:role_sort_order;type:int;not null;default:0" json:"roleSortOrder"`
	// RoleRemark 角色备注
	RoleRemark string `gorm:"column:role_remark;type:varchar(255);not null;default:''" json:"roleRemark"`
	// Creator 创建者 UID
	Creator int64 `gorm:"column:creator;type:bigint;not null;default:0" json:"creator"`
	// CreatedAt 创建时间
	CreatedAt time.Time `gorm:"column:created_at;type:datetime;not null;autoCreateTime" json:"createdAt"`
	// UpdatedAt 更新时间
	UpdatedAt time.Time `gorm:"column:updated_at;type:datetime;not null;autoUpdateTime" json:"updatedAt"`
	// DeletedAt 软删除标记
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;type:datetime;index" json:"deletedAt"`
}

// TableName 指定表名
func (*SysRole) TableName() string {
	return "moon_sys_roles"
}
