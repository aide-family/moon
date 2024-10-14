package menu

import (
	"context"

	menuapi "github.com/aide-family/moon/api/admin/menu"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz"
	"github.com/aide-family/moon/cmd/server/palace/internal/service/builder"
	"github.com/aide-family/moon/pkg/util/types"
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
	params := builder.NewParamsBuild().WithContext(ctx).MenuModuleBuilder().WithBatchCreateMenuRequest(req).ToBos()
	err := m.menuBiz.BatchCreateMenu(ctx, params)
	if err != nil {
		return nil, err
	}
	return &menuapi.BatchCreateMenuReply{}, nil
}

// UpdateMenu 更新菜单
func (m *Service) UpdateMenu(ctx context.Context, req *menuapi.UpdateMenuRequest) (*menuapi.UpdateMenuReply, error) {
	updateParams := builder.NewParamsBuild().WithContext(ctx).MenuModuleBuilder().WithUpdateMenuRequest(req).ToBo()
	if err := m.menuBiz.UpdateMenu(ctx, updateParams); !types.IsNil(err) {
		return nil, err
	}
	return &menuapi.UpdateMenuReply{}, nil
}

// DeleteMenu 删除菜单
func (m *Service) DeleteMenu(ctx context.Context, req *menuapi.DeleteMenuRequest) (*menuapi.DeleteMenuReply, error) {
	err := m.menuBiz.DeleteMenu(ctx, req.GetId())
	if err != nil {
		return nil, err
	}
	return &menuapi.DeleteMenuReply{}, nil
}

// GetMenu 获取菜单
func (m *Service) GetMenu(ctx context.Context, req *menuapi.GetMenuRequest) (*menuapi.GetMenuReply, error) {
	menuDetail, err := m.menuBiz.GetMenu(ctx, req.GetId())
	if !types.IsNil(err) {
		return nil, err
	}

	return &menuapi.GetMenuReply{
		Menu: builder.NewParamsBuild().WithContext(ctx).MenuModuleBuilder().DoMenuBuilder().ToAPI(menuDetail),
	}, nil
}

// TreeMenu 菜单树
func (m *Service) TreeMenu(ctx context.Context, req *menuapi.TreeMenuRequest) (*menuapi.TreeMenuReply, error) {
	dbMenuList, err := m.menuBiz.MenuAllList(ctx)
	if err != nil {
		return nil, err
	}

	menuTrees := builder.NewParamsBuild().WithContext(ctx).MenuModuleBuilder().DoMenuBuilder().ToAPITree(dbMenuList)
	return &menuapi.TreeMenuReply{MenuTree: menuTrees}, nil
}

// MenuListPage 菜单列表
func (m *Service) MenuListPage(ctx context.Context, req *menuapi.ListMenuRequest) (*menuapi.ListMenuReply, error) {
	queryParams := builder.NewParamsBuild().WithContext(ctx).MenuModuleBuilder().WithListMenuRequest(req).ToBo()
	menuPage, err := m.menuBiz.ListMenuPage(ctx, queryParams)
	if !types.IsNil(err) {
		return nil, err
	}
	return &menuapi.ListMenuReply{
		List:       builder.NewParamsBuild().WithContext(ctx).MenuModuleBuilder().DoMenuBuilder().ToAPIs(menuPage),
		Pagination: builder.NewParamsBuild().WithContext(ctx).PaginationModuleBuilder().ToAPI(queryParams.Page),
	}, nil
}

// BatchUpdateDictStatus 批量更新菜单状态
func (m *Service) BatchUpdateDictStatus(ctx context.Context, req *menuapi.BatchUpdateMenuStatusRequest) (*menuapi.BatchUpdateMenuStatusReply, error) {
	params := builder.NewParamsBuild().WithContext(ctx).MenuModuleBuilder().WithBatchUpdateMenuStatusRequest(req).ToBo()
	err := m.menuBiz.UpdateMenuStatus(ctx, params)
	if err != nil {
		return nil, err
	}
	return &menuapi.BatchUpdateMenuStatusReply{}, nil
}

// BatchUpdateMenuType 批量更新菜单类型
func (m *Service) BatchUpdateMenuType(ctx context.Context, req *menuapi.BatchUpdateMenuTypeRequest) (*menuapi.BatchUpdateMenuTypeReply, error) {
	params := builder.NewParamsBuild().WithContext(ctx).MenuModuleBuilder().WithUpdateMenuTypeParams(req).ToBo()
	if err := m.menuBiz.UpdateMenuTypes(ctx, params); err != nil {
		return nil, err
	}
	return &menuapi.BatchUpdateMenuTypeReply{}, nil
}
