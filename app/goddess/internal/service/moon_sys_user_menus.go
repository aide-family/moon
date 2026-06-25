package service

import (
	"context"

	"github.com/aide-family/goddess/internal/biz"
	"github.com/aide-family/goddess/internal/biz/bo"
	"github.com/aide-family/goddess/internal/data/impl/convert"
	apiv1 "github.com/aide-family/goddess/pkg/api/v1"
	"github.com/aide-family/magicbox/contextx"
)

// NewSysUserMenuService 创建系统用户菜单直授服务实例
func NewSysUserMenuService(sysUserMenuBiz *biz.SysUserMenu) *SysUserMenuService {
	return &SysUserMenuService{
		sysUserMenuBiz: sysUserMenuBiz,
	}
}

// SysUserMenuService 系统用户菜单直授服务
type SysUserMenuService struct {
	apiv1.UnimplementedSysUserMenuServer

	sysUserMenuBiz *biz.SysUserMenu
}

// ListUserMenu 查询用户直授菜单列表
func (s *SysUserMenuService) ListUserMenu(ctx context.Context, req *apiv1.ListUserMenuRequest) (*apiv1.ListUserMenuReply, error) {
	listBo := bo.NewListUserMenuBo(req)
	items, err := s.sysUserMenuBiz.ListUserMenu(ctx, listBo)
	if err != nil {
		return nil, err
	}

	if req.AsTree {
		menuItemBos := make([]*bo.SysMenuItemBo, 0, len(items))
		for _, item := range items {
			menuItemBos = append(menuItemBos, &bo.SysMenuItemBo{
				MenuCode:       item.MenuCode,
				ParentMenuCode: item.ParentMenuCode,
				MenuName:       item.MenuName,
				MenuSortOrder:  0,
				MenuStatus:     item.MenuStatus,
			})
		}
		tree := bo.BuildMenuTree(menuItemBos)
		return &apiv1.ListUserMenuReply{
			Tree: tree,
		}, nil
	}

	apiItems := make([]*apiv1.UserMenuItem, 0, len(items))
	for _, item := range items {
		apiItems = append(apiItems, convert.UserMenuItemBoToAPIV1(item))
	}
	return &apiv1.ListUserMenuReply{
		Items: apiItems,
	}, nil
}

// GrantUserMenus 直授系统菜单给用户
func (s *SysUserMenuService) GrantUserMenus(ctx context.Context, req *apiv1.GrantUserMenusRequest) (*apiv1.GrantUserMenusReply, error) {
	grantBo := bo.NewGrantUserMenusBo(req)
	grantBo.Creator = contextx.GetUserUID(ctx)
	result, err := s.sysUserMenuBiz.GrantUserMenus(ctx, grantBo)
	if err != nil {
		return nil, err
	}
	return &apiv1.GrantUserMenusReply{
		GrantedCount:       result.GrantedCount,
		AutoAddedMenuCodes: result.AutoAddedCodes,
	}, nil
}

// RevokeUserMenus 撤销直授系统菜单
func (s *SysUserMenuService) RevokeUserMenus(ctx context.Context, req *apiv1.RevokeUserMenusRequest) (*apiv1.RevokeUserMenusReply, error) {
	revokeBo := bo.NewRevokeUserMenusBo(req)
	result, err := s.sysUserMenuBiz.RevokeUserMenus(ctx, revokeBo)
	if err != nil {
		return nil, err
	}
	return &apiv1.RevokeUserMenusReply{
		RevokedCount:            result.RevokedCount,
		CascadeRevokedMenuCodes: result.CascadeRevokedCodes,
	}, nil
}
