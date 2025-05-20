package service

import (
	"context"

	"github.com/aide-family/moon/cmd/palace/internal/biz"
	api "github.com/aide-family/moon/pkg/api/palace"
	palacecommon "github.com/aide-family/moon/pkg/api/palace/common"
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
	return &palacecommon.EmptyReply{}, nil
}

func (s *MenuService) GetMenu(ctx context.Context, req *api.GetMenuRequest) (*palacecommon.MenuTreeItem, error) {
	return &palacecommon.MenuTreeItem{}, nil
}

func (s *MenuService) GetMenuTree(ctx context.Context, req *palacecommon.EmptyRequest) (*api.GetMenuTreeReply, error) {
	return &api.GetMenuTreeReply{}, nil
}

func (s *MenuService) GetTeamMenuTree(ctx context.Context, req *palacecommon.EmptyRequest) (*api.GetMenuTreeReply, error) {
	return &api.GetMenuTreeReply{}, nil
}
