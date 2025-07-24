package impl

import (
	"context"

	"github.com/go-kratos/kratos/v2/errors"
	"gorm.io/gen"
	"gorm.io/gen/field"
	"gorm.io/gorm"

	"github.com/aide-family/moon/cmd/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/biz/do/system"
	"github.com/aide-family/moon/cmd/palace/internal/biz/repository"
	"github.com/aide-family/moon/cmd/palace/internal/biz/vobj"
	"github.com/aide-family/moon/cmd/palace/internal/data"
	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/util/slices"
)

func NewMenuRepo(d *data.Data) repository.Menu {
	return &menuRepoImpl{
		Data: d,
	}
}

type menuRepoImpl struct {
	*data.Data
}

// FindMenusByType implements repository.Menu.
func (m *menuRepoImpl) FindMenusByType(ctx context.Context, menuType vobj.MenuType) ([]do.Menu, error) {
	mainQuery := getMainQuery(ctx, m)
	menu := mainQuery.Menu
	wrappers := []gen.Condition{
		menu.Status.Eq(vobj.GlobalStatusEnable.GetValue()),
		menu.MenuType.Eq(menuType.GetValue()),
	}
	menuDos, err := menu.WithContext(ctx).Where(wrappers...).Order(menu.Sort.Desc()).Find()
	if err != nil {
		return nil, err
	}
	return slices.Map(menuDos, func(menu *system.Menu) do.Menu { return menu }), nil
}

func (m *menuRepoImpl) Find(ctx context.Context, ids []uint32) ([]do.Menu, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	mainQuery := getMainQuery(ctx, m)
	menu := mainQuery.Menu
	menuDo, err := menu.WithContext(ctx).Where(menu.ID.In(ids...)).Order(menu.Sort.Desc()).Find()
	if err != nil {
		return nil, err
	}
	menus := slices.Map(menuDo, func(menu *system.Menu) do.Menu { return menu })
	return menus, nil
}

func (m *menuRepoImpl) FindAll(ctx context.Context, ids ...uint32) ([]do.Menu, error) {
	if len(ids) > 0 {
		return m.Find(ctx, ids)
	}
	mainQuery := getMainQuery(ctx, m)
	menu := mainQuery.Menu
	menuDos, err := menu.WithContext(ctx).Order(menu.Sort.Desc()).Find()
	if err != nil {
		return nil, err
	}
	return slices.Map(menuDos, func(menu *system.Menu) do.Menu { return menu }), nil
}

func (m *menuRepoImpl) GetMenuByOperation(ctx context.Context, operation string) (do.Menu, error) {
	mainQuery := getMainQuery(ctx, m)
	menu := mainQuery.Menu
	menuDo, err := menu.WithContext(ctx).Where(menu.APIPath.Eq(operation)).First()
	if err != nil {
		return nil, menuNotFound(err)
	}
	return menuDo, nil
}

func (m *menuRepoImpl) Create(ctx context.Context, menu *bo.SaveMenuRequest) error {
	mainQuery := getMainQuery(ctx, m)
	menuMutation := mainQuery.Menu
	systemMenu := &system.Menu{
		Name:          menu.Name,
		MenuPath:      menu.MenuPath,
		MenuIcon:      menu.MenuIcon,
		MenuType:      menu.MenuType,
		MenuCategory:  menu.MenuCategory,
		APIPath:       menu.APIPath,
		Status:        menu.Status,
		ProcessType:   menu.ProcessType,
		ParentID:      menu.ParentID,
		RelyOnBrother: menu.RelyOnBrother,
		Sort:          menu.Sort,
	}
	systemMenu.WithContext(ctx)
	return menuMutation.WithContext(ctx).Create(systemMenu)
}

func (m *menuRepoImpl) Update(ctx context.Context, menu *bo.SaveMenuRequest) error {
	mainQuery := getMainQuery(ctx, m)
	menuMutation := mainQuery.Menu
	mutations := []field.AssignExpr{
		menuMutation.Name.Value(menu.Name),
		menuMutation.MenuPath.Value(menu.MenuPath),
		menuMutation.MenuIcon.Value(menu.MenuIcon),
		menuMutation.MenuType.Value(menu.MenuType.GetValue()),
		menuMutation.MenuCategory.Value(menu.MenuCategory.GetValue()),
		menuMutation.APIPath.Value(menu.APIPath),
		menuMutation.Status.Value(menu.Status.GetValue()),
		menuMutation.ProcessType.Value(int8(menu.ProcessType)),
		menuMutation.ParentID.Value(menu.ParentID),
		menuMutation.RelyOnBrother.Value(menu.RelyOnBrother),
		menuMutation.Sort.Value(menu.Sort),
	}
	wrappers := []gen.Condition{
		menuMutation.ID.Eq(menu.MenuID),
	}
	_, err := menuMutation.WithContext(ctx).Where(wrappers...).UpdateColumnSimple(mutations...)
	return err
}

func (m *menuRepoImpl) ExistByName(ctx context.Context, name string, menuID uint32) error {
	mainQuery := getMainQuery(ctx, m)
	menu := mainQuery.Menu
	menuDo, err := menu.WithContext(ctx).Where(menu.Name.Eq(name), menu.ID.Neq(menuID)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}
	if menuID == menuDo.ID {
		return nil
	}
	return merr.ErrorExist("menu name already exists")
}

func (m *menuRepoImpl) FindMenus(ctx context.Context, params *bo.GetMenuTreeParams) ([]do.Menu, error) {
	mainQuery := getMainQuery(ctx, m)
	menu := mainQuery.Menu
	wrappers := []gen.Condition{
		menu.Status.Eq(vobj.GlobalStatusEnable.GetValue()),
	}
	if category := params.MenuCategory.GetValue(); category != 0 {
		wrappers = append(wrappers, menu.MenuCategory.Eq(category))
	}
	menuTypes := slices.MapFilter(params.MenuTypes, func(menuType vobj.MenuType) (int8, bool) {
		value := menuType.GetValue()
		return int8(value), value != 0
	})
	if len(menuTypes) > 0 {
		wrappers = append(wrappers, menu.MenuType.In(menuTypes...))
	}
	menuDos, err := menu.WithContext(ctx).Where(wrappers...).Order(menu.Sort.Desc()).Find()
	if err != nil {
		return nil, err
	}
	return slices.Map(menuDos, func(menu *system.Menu) do.Menu { return menu }), nil
}

func (m *menuRepoImpl) Delete(ctx context.Context, menuID uint32) error {
	mainQuery := getMainQuery(ctx, m)
	menu := mainQuery.Menu
	_, err := menu.WithContext(ctx).Where(menu.ID.Eq(menuID)).Delete()
	return err
}
