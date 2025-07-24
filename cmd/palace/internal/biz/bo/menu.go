package bo

import (
	"github.com/aide-family/moon/cmd/palace/internal/biz/vobj"
)

type SaveMenuRequest struct {
	MenuID        uint32
	Name          string
	MenuPath      string
	MenuIcon      string
	MenuType      vobj.MenuType
	MenuCategory  vobj.MenuCategory
	APIPath       string
	Status        vobj.GlobalStatus
	ProcessType   vobj.MenuProcessType
	ParentID      uint32
	RelyOnBrother bool
	Sort          uint32
}

type GetMenuTreeParams struct {
	MenuCategory vobj.MenuCategory
	MenuTypes    []vobj.MenuType
}
