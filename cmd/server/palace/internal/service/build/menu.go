package build

import (
	"github.com/aide-family/moon/api"
	"github.com/aide-family/moon/api/admin"
	"github.com/aide-family/moon/pkg/palace/model"
	"github.com/aide-family/moon/pkg/util/types"
)

type MenuBuilder struct {
	Menu *model.SysMenu
}

func NewMenuBuilder(menu *model.SysMenu) *MenuBuilder {
	return &MenuBuilder{
		Menu: menu,
	}
}

func (b *MenuBuilder) ToApi() *admin.Menu {
	if types.IsNil(b) || types.IsNil(b.Menu) {
		return nil
	}
	return &admin.Menu{
		Id:         b.Menu.ID,
		Name:       b.Menu.Name,
		Path:       b.Menu.Path,
		Icon:       b.Menu.Icon,
		Status:     api.Status(b.Menu.Status),
		ParentId:   b.Menu.ParentID,
		Sort:       b.Menu.Sort,
		Type:       api.MenuType(b.Menu.Type),
		Level:      b.Menu.Level,
		Component:  b.Menu.Component,
		Permission: b.Menu.Permission,
		EnName:     b.Menu.EnName,
		CreatedAt:  b.Menu.CreatedAt.String(),
		UpdatedAt:  b.Menu.UpdatedAt.String(),
	}
}

type MenuTreeBuilder struct {
	menuMap  map[uint32][]*admin.Menu
	parentID uint32
}

func NewMenuTreeBuilder(menuList []*admin.Menu, parentID uint32) *MenuTreeBuilder {
	menuMap := make(map[uint32][]*admin.Menu)
	// 按照父级ID分组
	for _, menu := range menuList {
		if _, ok := menuMap[menu.GetParentId()]; !ok {
			menuMap[menu.GetParentId()] = make([]*admin.Menu, 0)
		}
		menuMap[menu.GetParentId()] = append(menuMap[menu.GetParentId()], menu)
	}
	return &MenuTreeBuilder{
		menuMap:  menuMap,
		parentID: parentID,
	}
}

// ToTree 转换为树形菜单
func (b *MenuTreeBuilder) ToTree() []*admin.MenuTree {
	if types.IsNil(b) || types.IsNil(b.menuMap) || len(b.menuMap) == 0 {
		return nil
	}
	list := make([]*admin.MenuTree, 0)
	// 递归遍历
	for _, menu := range b.menuMap[b.parentID] {
		if menu.ParentId == b.parentID {
			list = append(list, &admin.MenuTree{
				Id:        menu.GetId(),
				Name:      menu.GetName(),
				Path:      menu.GetPath(),
				Icon:      menu.GetIcon(),
				Status:    menu.GetStatus(),
				ParentId:  menu.GetParentId(),
				CreatedAt: menu.GetCreatedAt(),
				UpdatedAt: menu.GetUpdatedAt(),
				Level:     menu.GetLevel(),
				Sort:      menu.GetSort(),
				EnName:    menu.GetEnName(),
				Children:  NewMenuTreeBuilder(b.menuMap[menu.GetId()], menu.GetId()).ToTree(),
			})
		}
	}
	return list
}
