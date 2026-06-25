package service

import (
	"context"

	"github.com/aide-family/goddess/internal/biz"
	"github.com/aide-family/goddess/internal/biz/bo"
	"github.com/aide-family/goddess/internal/data/impl/convert"
	apiv1 "github.com/aide-family/goddess/pkg/api/v1"
	"github.com/aide-family/magicbox/contextx"
)

// NewSysRoleMenuService 创建系统角色菜单服务实例
func NewSysRoleMenuService(sysRoleMenuBiz *biz.SysRoleMenu) *SysRoleMenuService {
	return &SysRoleMenuService{
		sysRoleMenuBiz: sysRoleMenuBiz,
	}
}

// SysRoleMenuService 系统角色菜单服务，实现 proto 定义的 SysRoleMenuServer 接口
type SysRoleMenuService struct {
	apiv1.UnimplementedSysRoleMenuServer

	sysRoleMenuBiz *biz.SysRoleMenu
}

// ListRoleMenu 查询系统角色拥有的菜单列表
func (s *SysRoleMenuService) ListRoleMenu(ctx context.Context, req *apiv1.ListRoleMenuRequest) (*apiv1.ListRoleMenuReply, error) {
	listBo := bo.NewListRoleMenuBo(req)
	items, err := s.sysRoleMenuBiz.ListRoleMenu(ctx, listBo)
	if err != nil {
		return nil, err
	}

	if req.AsTree {
		// 转换为 SysMenuItemBo 再用 BuildMenuTree 构建树
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
		return &apiv1.ListRoleMenuReply{
			Tree: tree,
		}, nil
	}

	apiItems := make([]*apiv1.RoleMenuItem, 0, len(items))
	for _, item := range items {
		apiItems = append(apiItems, convert.RoleMenuItemBoToAPIV1(item))
	}
	return &apiv1.ListRoleMenuReply{
		Items: apiItems,
	}, nil
}

// GrantRoleMenus 授予系统角色菜单
func (s *SysRoleMenuService) GrantRoleMenus(ctx context.Context, req *apiv1.GrantRoleMenusRequest) (*apiv1.GrantRoleMenusReply, error) {
	grantBo := bo.NewGrantRoleMenusBo(req)
	grantBo.Creator = int64(contextx.GetUserUID(ctx))
	result, err := s.sysRoleMenuBiz.GrantRoleMenus(ctx, grantBo)
	if err != nil {
		return nil, err
	}
	return &apiv1.GrantRoleMenusReply{
		GrantedCount:       result.GrantedCount,
		AutoAddedMenuCodes: result.AutoAddedCodes,
	}, nil
}

// RevokeRoleMenus 撤销系统角色菜单
func (s *SysRoleMenuService) RevokeRoleMenus(ctx context.Context, req *apiv1.RevokeRoleMenusRequest) (*apiv1.RevokeRoleMenusReply, error) {
	revokeBo := bo.NewRevokeRoleMenusBo(req)
	result, err := s.sysRoleMenuBiz.RevokeRoleMenus(ctx, revokeBo)
	if err != nil {
		return nil, err
	}
	return &apiv1.RevokeRoleMenusReply{
		RevokedCount:            result.RevokedCount,
		CascadeRevokedMenuCodes: result.CascadeRevokedCodes,
	}, nil
}
