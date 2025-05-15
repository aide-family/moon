package system

import (
	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/biz/vobj"
)

var _ do.UserConfigTheme = (*UserConfigTheme)(nil)

const (
	tableNameUserConfig = "sys_user_config_themes"
)

// UserConfigTheme 用户配置基础结构
type UserConfigTheme struct {
	do.CreatorModel
	ThemeMode    vobj.ThemeMode   `gorm:"column:theme_mode;type:tinyint(2);not null;comment:主题模式" json:"themeMode"`
	ThemeLayout  vobj.ThemeLayout `gorm:"column:theme_layout;type:tinyint(2);not null;comment:主题布局" json:"themeLayout"`
	PrimaryColor string           `gorm:"column:primary_color;type:varchar(20);not null;comment:主题主色" json:"primaryColor"`
	TimeZone     string           `gorm:"column:time_zone;type:varchar(20);not null;comment:时区" json:"timeZone"`
}

func (u *UserConfigTheme) GetThemeMode() vobj.ThemeMode {
	if u == nil {
		return vobj.ThemeModeUnknown
	}
	return u.ThemeMode
}

func (u *UserConfigTheme) GetPrimaryColor() string {
	if u == nil {
		return ""
	}
	return u.PrimaryColor
}

func (u *UserConfigTheme) GetThemeLayout() vobj.ThemeLayout {
	if u == nil {
		return vobj.ThemeLayoutUnknown
	}
	return u.ThemeLayout
}

func (u *UserConfigTheme) GetTimeZone() string {
	if u == nil {
		return ""
	}
	return u.TimeZone
}

func (u *UserConfigTheme) TableName() string {
	return tableNameUserConfig
}
