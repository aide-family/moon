package service

import (
	"context"

	"github.com/aide-family/magicbox/contextx"
	"github.com/aide-family/magicbox/merr"
	"github.com/aide-family/magicbox/strutil/cnst"
	"github.com/bwmarrin/snowflake"
	"github.com/go-kratos/kratos/v2/transport"

	"github.com/aide-family/goddess/internal/biz"
	"github.com/aide-family/goddess/internal/biz/bo"
	apiv1 "github.com/aide-family/goddess/pkg/api/v1"
)

// NewNmspacesRoleMenuService 创建租户角色菜单服务实例
func NewNmspacesRoleMenuService(nmspacesRoleMenuBiz *biz.NmspacesRoleMenuBiz) *NmspacesRoleMenuService {
	return &NmspacesRoleMenuService{
		nmspacesRoleMenuBiz: nmspacesRoleMenuBiz,
	}
}

// NmspacesRoleMenuService 租户角色菜单服务
type NmspacesRoleMenuService struct {
	apiv1.UnimplementedNmspacesRoleMenuServer

	nmspacesRoleMenuBiz *biz.NmspacesRoleMenuBiz
}

// getNamespaceUID 从请求头 X-Namespace 中获取租户 UID
func (s *NmspacesRoleMenuService) getNamespaceUID(ctx context.Context) (snowflake.ID, error) {
	if tr, ok := transport.FromServerContext(ctx); ok {
		namespace := tr.RequestHeader().Get(cnst.HTTPHeaderXNamespace)
		namespaceUID, err := snowflake.ParseString(namespace)
		if err == nil && namespaceUID > 0 {
			return namespaceUID, nil
		}
	}
	return 0, merr.ErrorForbidden("namespace is required, please set %s header", cnst.HTTPHeaderXNamespace)
}

// ListNmspacesRoleMenu 查询角色的菜单列表
func (s *NmspacesRoleMenuService) ListNmspacesRoleMenu(ctx context.Context, req *apiv1.ListNmspacesRoleMenuRequest) (*apiv1.ListNmspacesRoleMenuReply, error) {
	namespaceUID, err := s.getNamespaceUID(ctx)
	if err != nil {
		return nil, err
	}
	listBo := bo.NewListNmspacesRoleMenuBo(namespaceUID, req)
	items, err := s.nmspacesRoleMenuBiz.ListNmspacesRoleMenu(ctx, listBo)
	if err != nil {
		return nil, err
	}
	apiItems := make([]*apiv1.NmspacesRoleMenuItem, 0, len(items))
	for _, item := range items {
		apiItems = append(apiItems, &apiv1.NmspacesRoleMenuItem{
			MenuCode:       item.MenuCode,
			ParentMenuCode: item.ParentMenuCode,
			MenuName:       item.MenuName,
			MenuStatus:     item.MenuStatus,
		})
	}
	return &apiv1.ListNmspacesRoleMenuReply{
		Items: apiItems,
	}, nil
}

// GrantNmspacesRoleMenus 授予角色菜单
func (s *NmspacesRoleMenuService) GrantNmspacesRoleMenus(ctx context.Context, req *apiv1.GrantNmspacesRoleMenusRequest) (*apiv1.GrantNmspacesRoleMenusReply, error) {
	namespaceUID, err := s.getNamespaceUID(ctx)
	if err != nil {
		return nil, err
	}
	creator := contextx.GetUserUID(ctx)
	grantBo := bo.NewGrantNmspacesRoleMenusBo(namespaceUID, creator, req)
	result, err := s.nmspacesRoleMenuBiz.GrantNmspacesRoleMenus(ctx, grantBo)
	if err != nil {
		return nil, err
	}
	return &apiv1.GrantNmspacesRoleMenusReply{
		GrantedCount:       result.GrantedCount,
		AutoAddedMenuCodes: result.AutoAddedCodes,
	}, nil
}

// RevokeNmspacesRoleMenus 撤销角色菜单
func (s *NmspacesRoleMenuService) RevokeNmspacesRoleMenus(ctx context.Context, req *apiv1.RevokeNmspacesRoleMenusRequest) (*apiv1.RevokeNmspacesRoleMenusReply, error) {
	namespaceUID, err := s.getNamespaceUID(ctx)
	if err != nil {
		return nil, err
	}
	revokeBo := bo.NewRevokeNmspacesRoleMenusBo(namespaceUID, req)
	result, err := s.nmspacesRoleMenuBiz.RevokeNmspacesRoleMenus(ctx, revokeBo)
	if err != nil {
		return nil, err
	}
	return &apiv1.RevokeNmspacesRoleMenusReply{
		RevokedCount:            result.RevokedCount,
		CascadeRevokedMenuCodes: result.CascadeRevokedCodes,
	}, nil
}
