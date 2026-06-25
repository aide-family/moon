package do

import (
	"time"

	"github.com/bwmarrin/snowflake"
	"gorm.io/gorm"
)

// NmspacesFunction 租户功能信息表，对应数据库表 moon_nmspaces_functions
// 联合主键：(namespace_uid, menu_code, function_code)
type NmspacesFunction struct {
	// NamespaceUID 所属租户 UID
	NamespaceUID snowflake.ID `gorm:"column:namespace_uid;type:bigint unsigned;not null;default:0;primaryKey" json:"namespaceUID"`
	// MenuCode 菜单编码
	MenuCode string `gorm:"column:menu_code;type:varchar(100);not null;default:'';primaryKey" json:"menuCode"`
	// FunctionCode 功能编码
	FunctionCode string `gorm:"column:function_code;type:varchar(100);not null;default:'';primaryKey" json:"functionCode"`
	// FunctionName 功能名称
	FunctionName string `gorm:"column:function_name;type:varchar(100);not null;default:''" json:"functionName"`
	// FunctionSortOrder 排序值
	FunctionSortOrder int32 `gorm:"column:function_sort_order;type:int;not null;default:0" json:"functionSortOrder"`
	// FunctionStatus 状态：0-启用 1-禁用
	FunctionStatus int32 `gorm:"column:function_status;type:tinyint;not null;default:0" json:"functionStatus"`
	// FunctionRemark 备注
	FunctionRemark string `gorm:"column:function_remark;type:varchar(200);not null;default:''" json:"functionRemark"`
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
func (*NmspacesFunction) TableName() string {
	return "moon_nmspaces_functions"
}
