package do

import (
	"time"

	"gorm.io/gorm"
)

// SysMenu 系统菜单表，对应数据库表 moon_sys_menus
type SysMenu struct {
	// MenuCode 菜单编码，业务主键（如 user:manage）
	MenuCode string `gorm:"column:menu_code;type:varchar(64);not null;primaryKey" json:"menuCode"`
	// ParentMenuCode 父菜单编码，为空表示顶级菜单
	ParentMenuCode string `gorm:"column:parent_menu_code;type:varchar(64);not null;default:'';index" json:"parentMenuCode"`
	// MenuName 菜单名称
	MenuName string `gorm:"column:menu_name;type:varchar(100);not null;default:''" json:"menuName"`
	// MenuPath 路由路径
	MenuPath string `gorm:"column:menu_path;type:varchar(255);not null;default:''" json:"menuPath"`
	// MenuIcon 菜单图标
	MenuIcon string `gorm:"column:menu_icon;type:varchar(100);not null;default:''" json:"menuIcon"`
	// MenuComponent 前端组件路径
	MenuComponent string `gorm:"column:menu_component;type:varchar(255);not null;default:''" json:"menuComponent"`
	// MenuRedirect 重定向路径
	MenuRedirect string `gorm:"column:menu_redirect;type:varchar(255);not null;default:''" json:"menuRedirect"`
	// MenuIsHidden 是否隐藏：0-显示 1-隐藏
	MenuIsHidden int32 `gorm:"column:menu_is_hidden;type:tinyint;not null;default:0" json:"menuIsHidden"`
	// MenuIsKeepAlive 是否缓存：0-不缓存 1-缓存
	MenuIsKeepAlive int32 `gorm:"column:menu_is_keep_alive;type:tinyint;not null;default:0" json:"menuIsKeepAlive"`
	// MenuSortOrder 排序序号
	MenuSortOrder int32 `gorm:"column:menu_sort_order;type:int;not null;default:0" json:"menuSortOrder"`
	// MenuStatus 菜单状态：0-未知 1-启用 2-禁用
	MenuStatus int32 `gorm:"column:menu_status;type:tinyint;not null;default:1" json:"menuStatus"`
	// MenuRemark 菜单备注
	MenuRemark string `gorm:"column:menu_remark;type:varchar(255);not null;default:''" json:"menuRemark"`
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
func (*SysMenu) TableName() string {
	return "moon_sys_menus"
}
