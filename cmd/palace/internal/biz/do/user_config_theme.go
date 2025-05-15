package do

import (
	"github.com/aide-family/moon/cmd/palace/internal/biz/vobj"
)

// UserConfigTheme represents user theme configuration interface
type UserConfigTheme interface {
	Creator
	GetThemeMode() vobj.ThemeMode
	GetPrimaryColor() string
	GetThemeLayout() vobj.ThemeLayout
	GetTimeZone() string
}
