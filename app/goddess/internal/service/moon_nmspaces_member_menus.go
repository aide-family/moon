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

// NewNmspacesMemberMenuService 创建租户成员菜单直授服务实例
func NewNmspacesMemberMenuService(nmspacesMemberMenuBiz *biz.NmspacesMemberMenuBiz) *NmspacesMemberMenuService {
	return &NmspacesMemberMenuService{
		nmspacesMemberMenuBiz: nmspacesMemberMenuBiz,
	}
}

// NmspacesMemberMenuService 租户成员菜单直授服务
type NmspacesMemberMenuService struct {
	apiv1.UnimplementedNmspacesMemberMenuServer

	nmspacesMemberMenuBiz *biz.NmspacesMemberMenuBiz
}

// getNamespaceUID 从请求头 X-Namespace 中获取租户 UID
func (s *NmspacesMemberMenuService) getNamespaceUID(ctx context.Context) (snowflake.ID, error) {
	if tr, ok := transport.FromServerContext(ctx); ok {
		namespace := tr.RequestHeader().Get(cnst.HTTPHeaderXNamespace)
		namespaceUID, err := snowflake.ParseString(namespace)
		if err == nil && namespaceUID > 0 {
			return namespaceUID, nil
		}
	}
	return 0, merr.ErrorForbidden("namespace is required, please set %s header", cnst.HTTPHeaderXNamespace)
}

// ListNmspacesMemberMenu 查询成员直授菜单列表
func (s *NmspacesMemberMenuService) ListNmspacesMemberMenu(ctx context.Context, req *apiv1.ListNmspacesMemberMenuRequest) (*apiv1.ListNmspacesMemberMenuReply, error) {
	namespaceUID, err := s.getNamespaceUID(ctx)
	if err != nil {
		return nil, err
	}
	listBo := bo.NewListNmspacesMemberMenuBo(namespaceUID, req)
	items, err := s.nmspacesMemberMenuBiz.ListNmspacesMemberMenu(ctx, listBo)
	if err != nil {
		return nil, err
	}
	apiItems := make([]*apiv1.NmspacesMemberMenuItem, 0, len(items))
	for _, item := range items {
		apiItems = append(apiItems, &apiv1.NmspacesMemberMenuItem{
			MenuCode:       item.MenuCode,
			ParentMenuCode: item.ParentMenuCode,
			MenuName:       item.MenuName,
			MenuStatus:     item.MenuStatus,
		})
	}
	return &apiv1.ListNmspacesMemberMenuReply{
		Items: apiItems,
	}, nil
}

// GrantNmspacesMemberMenus 直授成员菜单
func (s *NmspacesMemberMenuService) GrantNmspacesMemberMenus(ctx context.Context, req *apiv1.GrantNmspacesMemberMenusRequest) (*apiv1.GrantNmspacesMemberMenusReply, error) {
	namespaceUID, err := s.getNamespaceUID(ctx)
	if err != nil {
		return nil, err
	}
	creator := contextx.GetUserUID(ctx)
	grantBo := bo.NewGrantNmspacesMemberMenusBo(namespaceUID, creator, req)
	result, err := s.nmspacesMemberMenuBiz.GrantNmspacesMemberMenus(ctx, grantBo)
	if err != nil {
		return nil, err
	}
	return &apiv1.GrantNmspacesMemberMenusReply{
		GrantedCount:       result.GrantedCount,
		AutoAddedMenuCodes: result.AutoAddedCodes,
	}, nil
}

// RevokeNmspacesMemberMenus 撤销成员直授菜单
func (s *NmspacesMemberMenuService) RevokeNmspacesMemberMenus(ctx context.Context, req *apiv1.RevokeNmspacesMemberMenusRequest) (*apiv1.RevokeNmspacesMemberMenusReply, error) {
	namespaceUID, err := s.getNamespaceUID(ctx)
	if err != nil {
		return nil, err
	}
	revokeBo := bo.NewRevokeNmspacesMemberMenusBo(namespaceUID, req)
	result, err := s.nmspacesMemberMenuBiz.RevokeNmspacesMemberMenus(ctx, revokeBo)
	if err != nil {
		return nil, err
	}
	return &apiv1.RevokeNmspacesMemberMenusReply{
		RevokedCount:            result.RevokedCount,
		CascadeRevokedMenuCodes: result.CascadeRevokedCodes,
	}, nil
}
