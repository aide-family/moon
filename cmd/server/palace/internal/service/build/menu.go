package build

import (
	"github.com/aide-family/moon/api"
	"github.com/aide-family/moon/api/admin"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
	"github.com/aide-family/moon/pkg/util/types"
)

type MenuBuild struct {
	menuMap  map[uint32][]*bizmodel.SysTeamMenu
	parentID uint32
}

func NewMenuBuild(menuList []*bizmodel.SysTeamMenu, parentID uint32) *MenuBuild {
	menuMap := make(map[uint32][]*bizmodel.SysTeamMenu)
	// 按照父级ID分组
	for _, menu := range menuList {
		if _, ok := menuMap[menu.ParentID]; !ok {
			menuMap[menu.ParentID] = make([]*bizmodel.SysTeamMenu, 0)
		}
		menuMap[menu.ParentID] = append(menuMap[menu.ParentID], menu)
	}
	return &MenuBuild{
		menuMap:  menuMap,
		parentID: parentID,
	}
}

// ToTree 转换为树形菜单
func (b *MenuBuild) ToTree() []*admin.Menu {
	if types.IsNil(b) || types.IsNil(b.menuMap) || len(b.menuMap) == 0 {
		return nil
	}
	list := make([]*admin.Menu, 0)
	// 递归遍历
	for _, menu := range b.menuMap[b.parentID] {
		if menu.ParentID == b.parentID {
			list = append(list, b.ToApi(menu))
		}
	}
	return list
}

func (b *MenuBuild) ToApi(menu *bizmodel.SysTeamMenu) *admin.Menu {
	if types.IsNil(menu) {
		return nil
	}
	return &admin.Menu{
		Id:        menu.ID,
		Name:      menu.Name,
		Path:      menu.Path,
		Icon:      menu.Icon,
		Status:    api.Status(menu.Status),
		ParentId:  menu.ParentID,
		CreatedAt: menu.CreatedAt.String(),
		UpdatedAt: menu.UpdatedAt.String(),
		Level:     menu.Level,
		Children:  NewMenuBuild(b.menuMap[menu.ID], menu.ID).ToTree(),
	}
}
