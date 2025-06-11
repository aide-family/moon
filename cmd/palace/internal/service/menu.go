package service

import (
	"context"

	"github.com/aide-family/moon/cmd/palace/internal/biz"
	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/service/build"
	api "github.com/aide-family/moon/pkg/api/palace"
	palacecommon "github.com/aide-family/moon/pkg/api/palace/common"
	"github.com/aide-family/moon/pkg/merr"
)

func NewMenuService(menuBiz *biz.Menu) *MenuService {
	return &MenuService{
		menuBiz: menuBiz,
	}
}

type MenuService struct {
	api.UnimplementedMenuServer
	menuBiz *biz.Menu
}

func (s *MenuService) SaveMenu(ctx context.Context, req *api.SaveMenuRequest) (*palacecommon.EmptyReply, error) {
	menu := build.ToSaveMenuRequest(req)

	if err := s.menuBiz.SaveMenu(ctx, menu); err != nil {
		return nil, err
	}
	return &palacecommon.EmptyReply{}, nil
}

func (s *MenuService) GetMenu(ctx context.Context, req *api.GetMenuRequest) (*palacecommon.MenuTreeItem, error) {
	menu, err := s.menuBiz.GetMenu(ctx, req.MenuId)
	if err != nil {
		return nil, err
	}
	return build.ToMenuTreeItem(menu), nil
}

func (s *MenuService) GetMenuTree(ctx context.Context, req *palacecommon.EmptyRequest) (*api.GetMenuTreeReply, error) {
	menus, err := s.menuBiz.SystemMenus(ctx)
	if err != nil {
		return nil, err
	}
	return &api.GetMenuTreeReply{
		Menus: build.ToMenuTree(menus),
	}, nil
}

func (s *MenuService) GetTeamMenuTree(ctx context.Context, req *palacecommon.EmptyRequest) (*api.GetMenuTreeReply, error) {
	menus, err := s.menuBiz.TeamMenus(ctx)
	if err != nil {
		return nil, err
	}
	return &api.GetMenuTreeReply{
		Menus: build.ToMenuTree(menus),
	}, nil
}

func (s *MenuService) GetMenuByOperation(ctx context.Context, operation string) (do.Menu, error) {
	menu, err := s.menuBiz.GetMenuByOperation(ctx, operation)
	if err != nil {
		if merr.IsNotFound(err) {
			return nil, merr.ErrorResourceNotOpen("menu")
		}
		return nil, err
	}
	return menu, nil
}
