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

// NewNmspacesMemberRoleService 创建租户成员角色服务实例
func NewNmspacesMemberRoleService(nmspacesMemberRoleBiz *biz.NmspacesMemberRoleBiz) *NmspacesMemberRoleService {
	return &NmspacesMemberRoleService{
		nmspacesMemberRoleBiz: nmspacesMemberRoleBiz,
	}
}

// NmspacesMemberRoleService 租户成员角色服务
type NmspacesMemberRoleService struct {
	apiv1.UnimplementedNmspacesMemberRoleServer

	nmspacesMemberRoleBiz *biz.NmspacesMemberRoleBiz
}

// getNamespaceUID 从请求头 X-Namespace 中获取租户 UID
func (s *NmspacesMemberRoleService) getNamespaceUID(ctx context.Context) (snowflake.ID, error) {
	if tr, ok := transport.FromServerContext(ctx); ok {
		namespace := tr.RequestHeader().Get(cnst.HTTPHeaderXNamespace)
		namespaceUID, err := snowflake.ParseString(namespace)
		if err == nil && namespaceUID > 0 {
			return namespaceUID, nil
		}
	}
	return 0, merr.ErrorForbidden("namespace is required, please set %s header", cnst.HTTPHeaderXNamespace)
}

// ListNmspacesMemberRole 分页查询成员的角色列表
func (s *NmspacesMemberRoleService) ListNmspacesMemberRole(ctx context.Context, req *apiv1.ListNmspacesMemberRoleRequest) (*apiv1.ListNmspacesMemberRoleReply, error) {
	namespaceUID, err := s.getNamespaceUID(ctx)
	if err != nil {
		return nil, err
	}
	listBo := bo.NewListNmspacesMemberRoleBo(namespaceUID, req)
	items, total, err := s.nmspacesMemberRoleBiz.ListNmspacesMemberRole(ctx, listBo)
	if err != nil {
		return nil, err
	}
	apiItems := make([]*apiv1.NmspacesMemberRoleItem, 0, len(items))
	for _, item := range items {
		apiItems = append(apiItems, item.ToAPIV1NmspacesMemberRoleItem())
	}
	return &apiv1.ListNmspacesMemberRoleReply{
		Items:    apiItems,
		Total:    total,
		Page:     listBo.PageRequestBo.Page,
		PageSize: listBo.PageRequestBo.PageSize,
	}, nil
}

// BindNmspacesMemberRoles 绑定成员角色
func (s *NmspacesMemberRoleService) BindNmspacesMemberRoles(ctx context.Context, req *apiv1.BindNmspacesMemberRolesRequest) (*apiv1.BindNmspacesMemberRolesReply, error) {
	namespaceUID, err := s.getNamespaceUID(ctx)
	if err != nil {
		return nil, err
	}
	grantor := contextx.GetUserUID(ctx)
	bindBo := bo.NewBindNmspacesMemberRolesBo(namespaceUID, grantor, req)
	boundCount, err := s.nmspacesMemberRoleBiz.BindNmspacesMemberRoles(ctx, bindBo)
	if err != nil {
		return nil, err
	}
	return &apiv1.BindNmspacesMemberRolesReply{
		BoundCount: boundCount,
	}, nil
}

// UnbindNmspacesMemberRoles 解绑成员角色
func (s *NmspacesMemberRoleService) UnbindNmspacesMemberRoles(ctx context.Context, req *apiv1.UnbindNmspacesMemberRolesRequest) (*apiv1.UnbindNmspacesMemberRolesReply, error) {
	namespaceUID, err := s.getNamespaceUID(ctx)
	if err != nil {
		return nil, err
	}
	unbindBo := bo.NewUnbindNmspacesMemberRolesBo(namespaceUID, req)
	unboundCount, err := s.nmspacesMemberRoleBiz.UnbindNmspacesMemberRoles(ctx, unbindBo)
	if err != nil {
		return nil, err
	}
	return &apiv1.UnbindNmspacesMemberRolesReply{
		UnboundCount: unboundCount,
	}, nil
}
