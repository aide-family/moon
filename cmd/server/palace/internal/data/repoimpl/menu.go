package repoimpl

import (
	"context"

	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/repository"
	"github.com/aide-family/moon/cmd/server/palace/internal/data"
	"github.com/aide-family/moon/pkg/palace/model"
	"github.com/aide-family/moon/pkg/palace/model/query"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"

	"gorm.io/gen"
)

// NewMenuRepository 创建菜单仓库
func NewMenuRepository(data *data.Data) repository.Menu {
	return &menuRepositoryImpl{
		data: data,
	}
}

type menuRepositoryImpl struct {
	data *data.Data
}

func (m *menuRepositoryImpl) Create(ctx context.Context, menu *bo.CreateMenuParams) (*model.SysMenu, error) {
	menuModel := createMenuParamsToModel(ctx, menu)
	menuModel.WithContext(ctx)
	queryWrapper := query.Use(m.data.GetMainDB(ctx)).WithContext(ctx).SysMenu
	if err := queryWrapper.Create(menuModel); !types.IsNil(err) {
		return nil, err
	}
	return menuModel, nil
}

func (m *menuRepositoryImpl) BatchCreate(ctx context.Context, menus []*bo.CreateMenuParams) error {
	menuModels := types.SliceToWithFilter(menus, func(item *bo.CreateMenuParams) (*model.SysMenu, bool) {
		if types.IsNil(item) || types.TextIsNull(item.Name) {
			return nil, false
		}
		return createMenuParamsToModel(ctx, item), true
	})
	mainQuery := query.Use(m.data.GetMainDB(ctx)).WithContext(ctx)
	return mainQuery.SysMenu.CreateInBatches(menuModels, 10)
}

func (m *menuRepositoryImpl) UpdateByID(ctx context.Context, menu *bo.UpdateMenuParams) error {
	updateParam := menu.UpdateParam
	mainQuery := query.Use(m.data.GetMainDB(ctx))
	_, err := mainQuery.WithContext(ctx).SysMenu.Where(mainQuery.SysMenu.ID.Eq(menu.ID)).UpdateSimple(
		mainQuery.SysMenu.Name.Value(updateParam.Name),
		mainQuery.SysMenu.Component.Value(updateParam.Component),
		mainQuery.SysMenu.Path.Value(updateParam.Path),
		mainQuery.SysMenu.Icon.Value(updateParam.Icon),
		mainQuery.SysMenu.Permission.Value(updateParam.Permission),
		mainQuery.SysMenu.Level.Value(updateParam.Level),
	)
	return err
}

func (m *menuRepositoryImpl) DeleteByID(ctx context.Context, id uint32) error {
	mainQuery := query.Use(m.data.GetMainDB(ctx))
	_, err := mainQuery.WithContext(ctx).SysDict.Where(mainQuery.SysMenu.ID.Eq(id)).Delete()
	return err
}

func (m *menuRepositoryImpl) GetByID(ctx context.Context, id uint32) (*model.SysMenu, error) {
	mainQuery := query.Use(m.data.GetMainDB(ctx))
	return mainQuery.SysMenu.WithContext(ctx).Where(mainQuery.SysMenu.ID.Eq(id)).First()
}

func (m *menuRepositoryImpl) ListAll(ctx context.Context) ([]*model.SysMenu, error) {
	mainQuery := query.Use(m.data.GetMainDB(ctx))
	menus, err := mainQuery.SysMenu.WithContext(ctx).Order(mainQuery.SysMenu.Sort.Asc()).Find()
	if err != nil {
		return nil, err
	}
	return menus, nil
}

func (m *menuRepositoryImpl) FindByPage(ctx context.Context, params *bo.QueryMenuListParams) ([]*model.SysMenu, error) {
	mainQuery := query.Use(m.data.GetMainDB(ctx))
	queryWrapper := mainQuery.SysMenu.WithContext(ctx)
	var wheres []gen.Condition
	if !params.Status.IsUnknown() {
		wheres = append(wheres, mainQuery.SysMenu.Status.Eq(params.Status.GetValue()))
	}

	if !params.MenuType.IsUnknown() {
		wheres = append(wheres, mainQuery.SysMenu.Type.Eq(params.MenuType.GetValue()))
	}

	if !types.TextIsNull(params.Keyword) {
		queryWrapper = queryWrapper.Or(
			mainQuery.SysMenu.Name.Like(params.Keyword),
			mainQuery.SysMenu.Path.Like(params.Keyword),
			mainQuery.SysMenu.EnName.Like(params.Keyword),
		)
	}
	queryWrapper = queryWrapper.Where(wheres...)
	if err := types.WithPageQuery[query.ISysMenuDo](queryWrapper, params.Page); err != nil {
		return nil, err
	}
	return queryWrapper.Order(mainQuery.SysMenu.ID.Desc()).Find()
}

func (m *menuRepositoryImpl) UpdateStatusByIds(ctx context.Context, status vobj.Status, ids ...uint32) error {
	mainQuery := query.Use(m.data.GetMainDB(ctx))
	_, err := mainQuery.WithContext(ctx).SysMenu.Where(mainQuery.SysMenu.ID.In(ids...)).Update(mainQuery.SysMenu.Status, status)
	return err
}

func (m *menuRepositoryImpl) UpdateTypeByIds(ctx context.Context, menuType vobj.MenuType, ids ...uint32) error {
	mainQuery := query.Use(m.data.GetMainDB(ctx))
	_, err := mainQuery.WithContext(ctx).SysMenu.Where(mainQuery.SysMenu.ID.In(ids...)).Update(mainQuery.SysMenu.Type, menuType)
	return err
}

func createMenuParamsToModel(ctx context.Context, param *bo.CreateMenuParams) *model.SysMenu {
	if types.IsNil(param) {
		return nil
	}

	menu := model.SysMenu{
		Name:       param.Name,
		Path:       param.Path,
		Icon:       param.Icon,
		Type:       param.Type,
		Sort:       param.Sort,
		ParentID:   param.ParentID,
		Status:     param.Status,
		Permission: param.Permission,
		EnName:     param.EnName,
	}
	menu.WithContext(ctx)
	return &menu
}
