package do

// SysFunction 系统功能信息表，对应数据库表 moon_sys_functions
// 联合主键：(menu_code, function_code)
type SysFunction struct {
	// MenuCode 菜单编码
	MenuCode string `gorm:"column:menu_code;type:varchar(100);not null;primaryKey" json:"menuCode"`
	// FunctionCode 功能编码（如 user:create）
	FunctionCode string `gorm:"column:function_code;type:varchar(100);not null;primaryKey" json:"functionCode"`
	// FunctionName 功能名称（如 新增用户）
	FunctionName string `gorm:"column:function_name;type:varchar(100);not null;default:''" json:"functionName"`
	// FunctionSortOrder 排序值
	FunctionSortOrder int32 `gorm:"column:function_sort_order;type:int;not null;default:0" json:"functionSortOrder"`
	// FunctionStatus 状态：0-启用 1-禁用
	FunctionStatus int32 `gorm:"column:function_status;type:tinyint;not null;default:0" json:"functionStatus"`
	// FunctionRemark 备注
	FunctionRemark string `gorm:"column:function_remark;type:varchar(200);not null;default:''" json:"functionRemark"`
	// Creator 创建人 UID
	Creator int64 `gorm:"column:creator;type:bigint;not null;default:0" json:"creator"`
	// CreatedAt 创建时间
	CreatedAt string `gorm:"column:created_at;type:datetime;not null;default:''" json:"createdAt"`
	// UpdatedAt 更新时间
	UpdatedAt string `gorm:"column:updated_at;type:datetime;not null;default:''" json:"updatedAt"`
	// DeletedAt 软删除时间
	DeletedAt *string `gorm:"column:deleted_at;type:datetime;index" json:"deletedAt"`
}

// TableName 指定表名
func (*SysFunction) TableName() string {
	return "moon_sys_functions"
}
