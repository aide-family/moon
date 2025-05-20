package build

import (
	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/pkg/api/palace/common"
)

func ToMenuTree(menus []do.Menu) []*common.MenuTreeItem {
	menuMap := make(map[uint32]do.Menu)
	for _, menu := range menus {
		menuMap[menu.GetID()] = menu
	}

	// 构建树形结构
	var roots []*common.MenuTreeItem
	for _, menu := range menus {
		if menu.GetParentID() == 0 {
			roots = append(roots, convertMenuToTreeItemWithMap(menu, menuMap))
		}
	}

	return roots
}

func ToMenuTreeItem(menu do.Menu) *common.MenuTreeItem {
	return convertMenuToTreeItemWithMap(menu, nil)
}

func convertMenuToTreeItemWithMap(menu do.Menu, menuMap map[uint32]do.Menu) *common.MenuTreeItem {
	treeItem := &common.MenuTreeItem{
		MenuId:          menu.GetID(),
		Name:            menu.GetName(),
		Status:          common.GlobalStatus(menu.GetStatus().GetValue()),
		Children:        nil,
		MenuType:        common.MenuType(menu.GetMenuType().GetValue()),
		MenuPath:        menu.GetMenuPath(),
		ApiPath:         menu.GetApiPath(),
		MenuIcon:        menu.GetMenuIcon(),
		MenuCategory:    common.MenuCategory(menu.GetMenuCategory().GetValue()),
		ProcessType:     common.MenuProcessType(menu.GetProcessType().GetValue()),
		ParentId:        menu.GetParentID(),
		IsRelyOnBrother: menu.IsRelyOnBrother(),
	}

	// 查找所有子菜单
	for _, m := range menuMap {
		if m.GetParentID() == menu.GetID() {
			if treeItem.Children == nil {
				treeItem.Children = make([]*common.MenuTreeItem, 0)
			}
			treeItem.Children = append(treeItem.Children, convertMenuToTreeItemWithMap(m, menuMap))
		}
	}

	return treeItem
}
