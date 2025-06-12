package build

import (
	"context"

	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/biz/do/system"
	"github.com/aide-family/moon/pkg/util/slices"
	"github.com/aide-family/moon/pkg/util/validate"
)

func ToMenus(ctx context.Context, menus []do.Menu) []*system.Menu {
	return slices.MapFilter(menus, func(menu do.Menu) (*system.Menu, bool) {
		if validate.IsNil(menu) {
			return nil, false
		}
		return ToMenu(ctx, menu), true
	})
}

func ToMenu(ctx context.Context, menu do.Menu) *system.Menu {
	menuDo := &system.Menu{
		BaseModel:     ToBaseModel(ctx, menu),
		Name:          menu.GetName(),
		MenuPath:      menu.GetMenuPath(),
		MenuIcon:      menu.GetMenuIcon(),
		MenuType:      menu.GetMenuType(),
		MenuCategory:  menu.GetMenuCategory(),
		ApiPath:       menu.GetApiPath(),
		Status:        menu.GetStatus(),
		ProcessType:   menu.GetProcessType(),
		ParentID:      menu.GetParentID(),
		RelyOnBrother: menu.IsRelyOnBrother(),
	}
	if validate.IsNotNil(menu.GetParent()) {
		menuDo.Parent = ToMenu(ctx, menu.GetParent())
		menuDo.ParentID = menu.GetParent().GetID()
	}
	menuDo.WithContext(ctx)
	return menuDo
}
