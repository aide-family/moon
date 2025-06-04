package impl

import (
	"context"

	"github.com/aide-family/moon/cmd/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/biz/do/system"
	"github.com/aide-family/moon/cmd/palace/internal/biz/repository"
	"github.com/aide-family/moon/cmd/palace/internal/biz/vobj"
	"github.com/aide-family/moon/cmd/palace/internal/data"
	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/util/slices"
	"github.com/aide-family/moon/pkg/util/validate"
	"github.com/go-kratos/kratos/v2/errors"
	"gorm.io/gen"
	"gorm.io/gen/field"
	"gorm.io/gorm"
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
	menuDos, err := menu.WithContext(ctx).Where(wrappers...).Find()
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
	menuDo, err := menu.WithContext(ctx).Where(menu.ID.In(ids...)).Find()
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
	menuDos, err := menu.WithContext(ctx).Find()
	if err != nil {
		return nil, err
	}
	return slices.Map(menuDos, func(menu *system.Menu) do.Menu { return menu }), nil
}

func (m *menuRepoImpl) GetMenuByOperation(ctx context.Context, operation string) (do.Menu, error) {
	mainQuery := getMainQuery(ctx, m)
	menu := mainQuery.Menu
	menuDo, err := menu.WithContext(ctx).Where(menu.ApiPath.Eq(operation)).First()
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
		ApiPath:       menu.ApiPath,
		Status:        menu.Status,
		ProcessType:   menu.ProcessType,
		ParentID:      menu.ParentID,
		RelyOnBrother: menu.RelyOnBrother,
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
		menuMutation.ApiPath.Value(menu.ApiPath),
		menuMutation.Status.Value(menu.Status.GetValue()),
		menuMutation.ProcessType.Value(menu.ProcessType.GetValue()),
		menuMutation.ParentID.Value(menu.ParentID),
		menuMutation.RelyOnBrother.Value(menu.RelyOnBrother),
	}
	wrappers := []gen.Condition{
		menuMutation.ID.Eq(menu.MenuId),
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

func (m *menuRepoImpl) List(ctx context.Context, req *bo.ListMenuParams) (*bo.ListMenuReply, error) {
	mainQuery := getMainQuery(ctx, m)
	menu := mainQuery.Menu
	wrapper := menu.Where()
	if !req.MenuType.IsUnknown() {
		wrapper = wrapper.Where(menu.MenuType.Eq(req.MenuType.GetValue()))
	}
	if !req.MenuCategory.IsUnknown() {
		wrapper = wrapper.Where(menu.MenuCategory.Eq(req.MenuCategory.GetValue()))
	}
	if !req.Status.IsUnknown() {
		wrapper = wrapper.Where(menu.Status.Eq(req.Status.GetValue()))
	}
	if !req.ProcessType.IsUnknown() {
		wrapper = wrapper.Where(menu.ProcessType.Eq(req.ProcessType.GetValue()))
	}
	if validate.TextIsNotNull(req.Keyword) {
		or := []gen.Condition{
			menu.Name.Like(req.Keyword),
			menu.MenuPath.Like(req.Keyword),
			menu.ApiPath.Like(req.Keyword),
			menu.MenuIcon.Like(req.Keyword),
		}
		wrapper = wrapper.Where(wrapper.Or(or...))
	}
	if req.TeamID > 0 {
		wrapper = wrapper.Where(menu.MenuType.Eq(vobj.MenuTypeMenuTeam.GetValue()))
	}
	menuDos, err := wrapper.WithContext(ctx).Find()
	if err != nil {
		return nil, err
	}
	return req.ToListReply(slices.Map(menuDos, func(menu *system.Menu) do.Menu { return menu })), nil
}
