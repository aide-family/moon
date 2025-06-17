package bo

import (
	"github.com/aide-family/moon/cmd/palace/internal/biz/vobj"
)

type SaveMenuRequest struct {
	MenuId        uint32
	Name          string
	MenuPath      string
	MenuIcon      string
	MenuType      vobj.MenuType
	MenuCategory  vobj.MenuCategory
	ApiPath       string
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
