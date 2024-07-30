package menu

import (
	"context"

	"github.com/aide-family/moon/api/admin"
	menuapi "github.com/aide-family/moon/api/admin/menu"
	"github.com/aide-family/moon/api/merr"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/palace/internal/service/build"
	"github.com/aide-family/moon/pkg/palace/model"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
)

// Service 菜单服务
type Service struct {
	menuapi.UnimplementedMenuServer
	menuBiz *biz.MenuBiz
}

// NewMenuService 创建菜单服务
func NewMenuService(menuBiz *biz.MenuBiz) *Service {
	return &Service{
		menuBiz: menuBiz,
	}
}

// BatchCreateMenu 批量创建菜单
func (m *Service) BatchCreateMenu(ctx context.Context, req *menuapi.BatchCreateMenuRequest) (*menuapi.BatchCreateMenuReply, error) {
	params := build.NewBuilder().WithBatchCreateMenuBo(req).ToBatchCreateMenuBO()

	err := m.menuBiz.BatchCreateMenu(ctx, params)
	if err != nil {
		return nil, err
	}
	return &menuapi.BatchCreateMenuReply{}, nil
}

// UpdateMenu 更新菜单
func (m *Service) UpdateMenu(ctx context.Context, req *menuapi.UpdateMenuRequest) (*menuapi.UpdateMenuReply, error) {
	updateParams := build.NewBuilder().WithUpdateMenuBo(req).ToUpdateMenuBO()
	if err := m.menuBiz.UpdateMenu(ctx, updateParams); !types.IsNil(err) {
		return nil, err
	}
	return &menuapi.UpdateMenuReply{}, nil
}

// DeleteMenu 删除菜单
func (m *Service) DeleteMenu(ctx context.Context, req *menuapi.DeleteMenuRequest) (*menuapi.DeleteMenuReply, error) {
	err := m.menuBiz.DeleteMenu(ctx, req.GetId())
	if err != nil {
		return nil, merr.ErrorI18nSystemErr(ctx).WithCause(err)
	}
	return &menuapi.DeleteMenuReply{}, nil
}

// GetMenu 获取菜单
func (m *Service) GetMenu(ctx context.Context, req *menuapi.GetMenuRequest) (*menuapi.GetMenuReply, error) {
	menu, err := m.menuBiz.GetMenu(ctx, req.GetId())
	if !types.IsNil(err) {
		return nil, err
	}
	resMenu := build.NewBuilder().WithAPIMenu(menu).ToAPI()
	return &menuapi.GetMenuReply{
		Menu: resMenu,
	}, nil
}

// TreeMenu 菜单树
func (m *Service) TreeMenu(ctx context.Context, req *menuapi.TreeMenuRequest) (*menuapi.TreeMenuReply, error) {
	dbMenuList, err := m.menuBiz.MenuAllList(ctx)
	if err != nil {
		return nil, err
	}
	menus := types.SliceTo(dbMenuList, func(menu *model.SysMenu) *admin.Menu {
		return build.NewBuilder().WithAPIMenu(menu).ToAPI()
	})
	menuTrees := build.NewBuilder().WithAPIMenuTree(menus, 0).ToTree()
	return &menuapi.TreeMenuReply{MenuTree: menuTrees}, nil
}

// MenuListPage 菜单列表
func (m *Service) MenuListPage(ctx context.Context, req *menuapi.ListMenuRequest) (*menuapi.ListMenuReply, error) {
	queryParams := &bo.QueryMenuListParams{
		Keyword:  req.GetKeyword(),
		Page:     types.NewPagination(req.GetPagination()),
		Status:   vobj.Status(req.GetStatus()),
		MenuType: vobj.MenuType(req.GetMenuType()),
	}
	menuPage, err := m.menuBiz.ListMenuPage(ctx, queryParams)
	if !types.IsNil(err) {
		return nil, err
	}
	return &menuapi.ListMenuReply{
		List: types.SliceTo(menuPage, func(menu *model.SysMenu) *admin.Menu {
			return build.NewBuilder().WithAPIMenu(menu).ToAPI()
		}),
		Pagination: build.NewPageBuilder(queryParams.Page).ToAPI(),
	}, nil
}

// BatchUpdateDictStatus 批量更新菜单状态
func (m *Service) BatchUpdateDictStatus(ctx context.Context, req *menuapi.BatchUpdateMenuStatusRequest) (*menuapi.BatchUpdateMenuStatusReply, error) {
	params := &bo.UpdateMenuStatusParams{
		IDs:    req.Ids,
		Status: vobj.Status(req.Status),
	}
	err := m.menuBiz.UpdateMenuStatus(ctx, params)
	if err != nil {
		return nil, merr.ErrorI18nSystemErr(ctx).WithCause(err)
	}
	return &menuapi.BatchUpdateMenuStatusReply{}, nil
}

// BatchUpdateMenuType 批量更新菜单类型
func (m *Service) BatchUpdateMenuType(ctx context.Context, req *menuapi.BatchUpdateMenuTypeRequest) (*menuapi.BatchUpdateMenuTypeReply, error) {
	params := &bo.UpdateMenuTypeParams{
		IDs:  req.Ids,
		Type: vobj.MenuType(req.GetType()),
	}
	err := m.menuBiz.UpdateMenuTypes(ctx, params)
	if err != nil {
		return nil, merr.ErrorI18nSystemErr(ctx).WithCause(err)
	}
	return &menuapi.BatchUpdateMenuTypeReply{}, nil
}
