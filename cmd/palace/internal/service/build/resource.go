package build

import (
	"github.com/aide-family/moon/cmd/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/biz/vobj"
	"github.com/aide-family/moon/pkg/api/palace"
	"github.com/aide-family/moon/pkg/api/palace/common"
	"github.com/aide-family/moon/pkg/util/slices"
	"github.com/aide-family/moon/pkg/util/timex"
	"github.com/aide-family/moon/pkg/util/validate"
)

func ToResourceItem(resource do.Resource) *common.ResourceItem {
	if resource == nil {
		return nil
	}
	return &common.ResourceItem{
		ResourceId: resource.GetID(),
		Name:       resource.GetName(),
		Path:       resource.GetPath(),
		Status:     common.GlobalStatus(resource.GetStatus().GetValue()),
		Remark:     resource.GetRemark(),
		CreatedAt:  timex.Format(resource.GetCreatedAt()),
		UpdatedAt:  timex.Format(resource.GetUpdatedAt()),
		Allow:      common.ResourceAllow(resource.GetAllow().GetValue()),
		Menus:      ToMenuTree(resource.GetMenus()),
	}
}

func ToResourceItems(resources []do.Resource) []*common.ResourceItem {
	return slices.Map(resources, ToResourceItem)
}

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
		MenuId:    menu.GetID(),
		Name:      menu.GetName(),
		Path:      menu.GetPath(),
		Status:    common.GlobalStatus(menu.GetStatus().GetValue()),
		Icon:      menu.GetIcon(),
		Children:  nil,
		MenuType:  common.MenuType(menu.GetType().GetValue()),
		Resources: ToResourceItems(menu.GetResources()),
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

func ToSaveMenuReq(req *palace.SaveMenuRequest) *bo.SaveMenuReq {
	if validate.IsNil(req) {
		return nil
	}
	return &bo.SaveMenuReq{
		ID:          req.GetMenuId(),
		Name:        req.GetName(),
		Path:        req.GetPath(),
		Status:      vobj.GlobalStatus(req.GetStatus()),
		Icon:        req.GetIcon(),
		ParentID:    req.GetParentId(),
		Type:        vobj.MenuType(req.GetMenuType()),
		ResourceIds: req.GetResourceIds(),
	}
}

func ToSaveResourceReq(req *palace.SaveResourceRequest) *bo.SaveResourceReq {
	if validate.IsNil(req) {
		return nil
	}
	return &bo.SaveResourceReq{
		ID:     req.GetResourceId(),
		Name:   req.GetName(),
		Path:   req.GetPath(),
		Status: vobj.GlobalStatus(req.GetStatus()),
		Allow:  vobj.ResourceAllow(req.GetAllow()),
		Remark: req.GetRemark(),
	}
}
