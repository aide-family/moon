package do

import (
	"time"

	"github.com/bwmarrin/snowflake"
	"gorm.io/gorm"
)

// NmspacesMenu 租户菜单表，对应数据库表 moon_nmspaces_menus
type NmspacesMenu struct {
	// NamespaceUID 所属租户 UID
	NamespaceUID snowflake.ID `gorm:"column:namespace_uid;type:bigint unsigned;not null;default:0;primaryKey" json:"namespaceUID"`
	// MenuCode 菜单编码（同租户内唯一）
	MenuCode string `gorm:"column:menu_code;type:varchar(100);not null;default:'';primaryKey" json:"menuCode"`
	// ParentMenuCode 父菜单编码，空=根节点
	ParentMenuCode string `gorm:"column:parent_menu_code;type:varchar(100);not null;default:''" json:"parentMenuCode"`
	// MenuName 菜单名称
	MenuName string `gorm:"column:menu_name;type:varchar(100);not null;default:''" json:"menuName"`
	// MenuPath 前端路由路径
	MenuPath string `gorm:"column:menu_path;type:varchar(200);not null;default:''" json:"menuPath"`
	// MenuIcon 菜单图标
	MenuIcon string `gorm:"column:menu_icon;type:varchar(100);not null;default:''" json:"menuIcon"`
	// MenuComponent 前端组件路径
	MenuComponent string `gorm:"column:menu_component;type:varchar(200);not null;default:''" json:"menuComponent"`
	// MenuRedirect 重定向地址
	MenuRedirect string `gorm:"column:menu_redirect;type:varchar(200);not null;default:''" json:"menuRedirect"`
	// MenuIsHidden 是否隐藏
	MenuIsHidden int32 `gorm:"column:menu_is_hidden;type:tinyint;not null;default:0" json:"menuIsHidden"`
	// MenuIsKeepAlive 是否缓存页面
	MenuIsKeepAlive int32 `gorm:"column:menu_is_keep_alive;type:tinyint;not null;default:1" json:"menuIsKeepAlive"`
	// MenuSortOrder 排序值
	MenuSortOrder int32 `gorm:"column:menu_sort_order;type:int;not null;default:0" json:"menuSortOrder"`
	// MenuStatus 菜单状态
	MenuStatus int32 `gorm:"column:menu_status;type:tinyint;not null;default:0" json:"menuStatus"`
	// MenuRemark 备注
	MenuRemark string `gorm:"column:menu_remark;type:varchar(200);not null;default:''" json:"menuRemark"`
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
func (*NmspacesMenu) TableName() string {
	return "moon_nmspaces_menus"
}
