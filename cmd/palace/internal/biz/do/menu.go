package do

import (
	"github.com/aide-family/moon/cmd/palace/internal/biz/vobj"
	"github.com/aide-family/moon/pkg/plugin/cache"
)

type Menu interface {
	Base
	cache.Object
	GetName() string
	GetMenuPath() string
	GetMenuIcon() string
	GetMenuType() vobj.MenuType
	GetMenuCategory() vobj.MenuCategory
	GetAPIPath() string
	GetStatus() vobj.GlobalStatus
	GetProcessType() vobj.MenuProcessType
	GetParentID() uint32
	GetParent() Menu
	IsRelyOnBrother() bool
	GetSort() uint32
}
