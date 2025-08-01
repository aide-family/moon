package build

import (
	"github.com/aide-family/moon/cmd/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/biz/vobj"
	api "github.com/aide-family/moon/pkg/api/palace"
	"github.com/aide-family/moon/pkg/api/palace/common"
	"github.com/aide-family/moon/pkg/util/slices"
	"github.com/aide-family/moon/pkg/util/validate"
)

func ToMenuTree(menus []do.Menu) []*common.MenuTreeItem {
	if len(menus) == 0 {
		return nil
	}
	menus = slices.UniqueWithFunc(menus, func(v do.Menu) uint32 {
		return v.GetID()
	})
	menuMap := make(map[uint32]do.Menu)
	for _, menu := range menus {
		menuMap[menu.GetID()] = menu
	}

	// Build tree structure
	var roots []*common.MenuTreeItem
	for _, menu := range menus {
		if menu.GetParentID() == 0 {
			roots = append(roots, convertMenuToTreeItemWithChildren(menu, menuMap))
		}
	}

	return roots
}

func ToMenuTreeItem(menu do.Menu) *common.MenuTreeItem {
	if validate.IsNil(menu) {
		return nil
	}
	return convertMenuToTreeItemWithChildren(menu, nil)
}

func convertMenuToTreeItemWithChildren(menu do.Menu, menuMap map[uint32]do.Menu) *common.MenuTreeItem {
	if validate.IsNil(menu) {
		return nil
	}
	treeItem := &common.MenuTreeItem{
		MenuId:          menu.GetID(),
		Name:            menu.GetName(),
		Status:          common.GlobalStatus(menu.GetStatus().GetValue()),
		Children:        nil,
		MenuType:        common.MenuType(menu.GetMenuType().GetValue()),
		MenuPath:        menu.GetMenuPath(),
		ApiPath:         menu.GetAPIPath(),
		MenuIcon:        menu.GetMenuIcon(),
		MenuCategory:    common.MenuCategory(menu.GetMenuCategory().GetValue()),
		ProcessType:     common.MenuProcessType(menu.GetProcessType()),
		ParentId:        menu.GetParentID(),
		IsRelyOnBrother: menu.IsRelyOnBrother(),
		Sort:            menu.GetSort(),
	}

	for _, m := range menuMap {
		if m.GetParentID() == menu.GetID() {
			child := convertMenuToTreeItemWithChildren(m, menuMap)
			if treeItem.Children == nil {
				treeItem.Children = make([]*common.MenuTreeItem, 0)
			}
			treeItem.Children = append(treeItem.Children, child)
		}
	}

	return treeItem
}

func ToSaveMenuRequest(req *api.SaveMenuRequest) *bo.SaveMenuRequest {
	return &bo.SaveMenuRequest{
		Name:          req.GetName(),
		MenuPath:      req.GetMenuPath(),
		MenuIcon:      req.GetMenuIcon(),
		MenuType:      vobj.MenuType(req.GetMenuType()),
		MenuCategory:  vobj.MenuCategory(req.GetMenuCategory()),
		APIPath:       req.GetApiPath(),
		Status:        vobj.GlobalStatus(req.GetStatus()),
		ProcessType:   vobj.MenuProcessType(req.GetProcessType()),
		ParentID:      req.GetParentId(),
		RelyOnBrother: req.GetIsRelyOnBrother(),
		MenuID:        req.GetMenuId(),
		Sort:          req.GetSort(),
	}
}
