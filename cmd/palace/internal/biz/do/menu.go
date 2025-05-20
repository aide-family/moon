package do

import (
	"github.com/aide-family/moon/cmd/palace/internal/biz/vobj"
)

type Menu interface {
	Base
	GetName() string
	GetMenuPath() string
	GetMenuIcon() string
	GetMenuType() vobj.MenuType
	GetMenuCategory() vobj.MenuCategory
	GetApiPath() string
	GetStatus() vobj.GlobalStatus
	GetProcessType() vobj.MenuProcessType
	GetParentID() uint32
	GetParent() Menu
	IsRelyOnBrother() bool
}
