package do

import (
	"time"

	"github.com/bwmarrin/snowflake"
	"gorm.io/gorm"
)

// NmspacesRole 租户角色表，对应数据库表 moon_nmspaces_roles
type NmspacesRole struct {
	// NamespaceUID 所属租户 UID
	NamespaceUID snowflake.ID `gorm:"column:namespace_uid;type:bigint unsigned;not null;default:0;primaryKey" json:"namespaceUID"`
	// RoleCode 租户角色编码（同租户内唯一）
	RoleCode string `gorm:"column:role_code;type:varchar(100);not null;default:'';primaryKey" json:"roleCode"`
	// RoleName 租户角色名称
	RoleName string `gorm:"column:role_name;type:varchar(100);not null;default:''" json:"roleName"`
	// RoleDataScope 数据权限范围：1-公司级 2-部门级 3-个人级
	RoleDataScope int32 `gorm:"column:role_data_scope;type:tinyint;not null;default:3" json:"roleDataScope"`
	// RoleStatus 角色状态：0-启用 1-禁用
	RoleStatus int32 `gorm:"column:role_status;type:tinyint;not null;default:0" json:"roleStatus"`
	// RoleSortOrder 排序值
	RoleSortOrder int32 `gorm:"column:role_sort_order;type:int;not null;default:0" json:"roleSortOrder"`
	// RoleRemark 角色备注
	RoleRemark string `gorm:"column:role_remark;type:varchar(200);not null;default:''" json:"roleRemark"`
	// Creator 创建人 UID
	Creator snowflake.ID `gorm:"column:creator;type:bigint unsigned;not null;default:0" json:"creator"`
	// CreatedAt 创建时间
	CreatedAt time.Time `gorm:"column:created_at;type:datetime;not null;autoCreateTime" json:"createdAt"`
	// UpdatedAt 更新时间
	UpdatedAt time.Time `gorm:"column:updated_at;type:datetime;not null;autoUpdateTime" json:"updatedAt"`
	// DeletedAt 软删除时间
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;type:datetime;index" json:"deletedAt"`
}

// TableName 指定表名
func (*NmspacesRole) TableName() string {
	return "moon_nmspaces_roles"
}
