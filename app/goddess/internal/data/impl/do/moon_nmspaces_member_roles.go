package do

import (
	"time"

	"github.com/bwmarrin/snowflake"
)

// NmspacesMemberRole 租户成员角色权限关联表，对应数据库表 moon_nmspaces_member_roles
// 联合主键：(namespace_uid, member_uid, role_code)，纯关联表，硬删除
type NmspacesMemberRole struct {
	// NamespaceUID 所属租户 UID
	NamespaceUID snowflake.ID `gorm:"column:namespace_uid;type:bigint unsigned;not null;default:0;primaryKey" json:"namespaceUID"`
	// MemberUID 成员 UID
	MemberUID snowflake.ID `gorm:"column:member_uid;type:bigint unsigned;not null;default:0;primaryKey" json:"memberUID"`
	// RoleCode 角色编码
	RoleCode string `gorm:"column:role_code;type:varchar(100);not null;default:'';primaryKey" json:"roleCode"`
	// Grantor 授权人 UID
	Grantor snowflake.ID `gorm:"column:grantor;type:bigint unsigned;not null;default:0" json:"grantor"`
	// CreatedAt 授权时间
	CreatedAt time.Time `gorm:"column:created_at;type:datetime;not null;autoCreateTime" json:"createdAt"`
}

// TableName 指定表名
func (*NmspacesMemberRole) TableName() string {
	return "moon_nmspaces_member_roles"
}
