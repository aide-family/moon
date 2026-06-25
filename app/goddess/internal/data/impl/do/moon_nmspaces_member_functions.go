package do

import (
	"time"

	"github.com/bwmarrin/snowflake"
)

// NmspacesMemberFunction 租户成员功能直授关联表，对应数据库表 moon_nmspaces_member_functions
// 联合主键：(namespace_uid, member_uid, menu_code, function_code)，纯关联表，硬删除
type NmspacesMemberFunction struct {
	// NamespaceUID 所属租户 UID
	NamespaceUID snowflake.ID `gorm:"column:namespace_uid;type:bigint unsigned;not null;default:0;primaryKey" json:"namespaceUID"`
	// MemberUID 成员 UID
	MemberUID snowflake.ID `gorm:"column:member_uid;type:bigint unsigned;not null;default:0;primaryKey" json:"memberUID"`
	// MenuCode 菜单编码
	MenuCode string `gorm:"column:menu_code;type:varchar(100);not null;default:'';primaryKey" json:"menuCode"`
	// FunctionCode 功能编码
	FunctionCode string `gorm:"column:function_code;type:varchar(100);not null;default:'';primaryKey" json:"functionCode"`
	// Creator 授权人 UID
	Creator snowflake.ID `gorm:"column:creator;type:bigint unsigned;not null;default:0" json:"creator"`
	// CreatedAt 授权时间
	CreatedAt time.Time `gorm:"column:created_at;type:datetime;not null;autoCreateTime" json:"createdAt"`
}

// TableName 指定表名
func (*NmspacesMemberFunction) TableName() string {
	return "moon_nmspaces_member_functions"
}
