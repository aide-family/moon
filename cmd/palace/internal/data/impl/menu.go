package impl

import (
	"context"

	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/biz/do/system"
	"github.com/aide-family/moon/cmd/palace/internal/biz/repository"
	"github.com/aide-family/moon/cmd/palace/internal/biz/vobj"
	"github.com/aide-family/moon/cmd/palace/internal/data"
	"github.com/aide-family/moon/pkg/util/slices"
	"gorm.io/gen"
)

func NewMenuRepo(d *data.Data) repository.Menu {
	return &menuImpl{
		Data: d,
	}
}

type menuImpl struct {
	*data.Data
}

// FindMenusByType implements repository.Menu.
func (m *menuImpl) FindMenusByType(ctx context.Context, menuType vobj.MenuType) ([]do.Menu, error) {
	mainQuery := getMainQuery(ctx, m)
	menu := mainQuery.Menu
	wrappers := []gen.Condition{
		menu.Status.Eq(vobj.GlobalStatusEnable.GetValue()),
		menu.MenuType.Eq(menuType.GetValue()),
	}
	menuDos, err := menu.WithContext(ctx).Where(wrappers...).Find()
	if err != nil {
		return nil, err
	}
	return slices.Map(menuDos, func(menu *system.Menu) do.Menu { return menu }), nil
}

func (m *menuImpl) Find(ctx context.Context, ids []uint32) ([]do.Menu, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	mainQuery := getMainQuery(ctx, m)
	menu := mainQuery.Menu
	menuDo, err := menu.WithContext(ctx).Where(menu.ID.In(ids...)).Find()
	if err != nil {
		return nil, err
	}
	menus := slices.Map(menuDo, func(menu *system.Menu) do.Menu { return menu })
	return menus, nil
}

func (m *menuImpl) GetMenuByOperation(ctx context.Context, operation string) (do.Menu, error) {
	mainQuery := getMainQuery(ctx, m)
	menu := mainQuery.Menu
	menuDo, err := menu.WithContext(ctx).Where(menu.ApiPath.Eq(operation)).First()
	if err != nil {
		return nil, menuNotFound(err)
	}
	return menuDo, nil
}
