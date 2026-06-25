package do

import (
	"time"

	"github.com/bwmarrin/snowflake"
)

// SysUserMenu 系统用户菜单直授关联表，对应数据库表 moon_sys_user_menus
// 联合主键：(user_uid, menu_code)，纯关联表，硬删除
type SysUserMenu struct {
	// UserUID 用户 UID
	UserUID snowflake.ID `gorm:"column:user_uid;type:bigint;not null;primaryKey" json:"userUID"`
	// MenuCode 菜单编码
	MenuCode string `gorm:"column:menu_code;type:varchar(100);not null;primaryKey" json:"menuCode"`
	// Creator 授权人 UID
	Creator snowflake.ID `gorm:"column:creator;type:bigint;not null;default:0" json:"creator"`
	// CreatedAt 授权时间
	CreatedAt time.Time `gorm:"column:created_at;type:datetime;not null;autoCreateTime" json:"createdAt"`
}

// TableName 指定表名
func (*SysUserMenu) TableName() string {
	return "moon_sys_user_menus"
}
