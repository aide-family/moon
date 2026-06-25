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

// NewNmspacesRoleService 创建租户角色服务实例
func NewNmspacesRoleService(nmspacesRoleBiz *biz.NmspacesRoleBiz) *NmspacesRoleService {
	return &NmspacesRoleService{
		nmspacesRoleBiz: nmspacesRoleBiz,
	}
}

// NmspacesRoleService 租户角色服务
type NmspacesRoleService struct {
	apiv1.UnimplementedNmspacesRoleServer

	nmspacesRoleBiz *biz.NmspacesRoleBiz
}

// getNamespaceUID 从请求头 X-Namespace 中获取租户 UID
func (s *NmspacesRoleService) getNamespaceUID(ctx context.Context) (snowflake.ID, error) {
	if tr, ok := transport.FromServerContext(ctx); ok {
		namespace := tr.RequestHeader().Get(cnst.HTTPHeaderXNamespace)
		namespaceUID, err := snowflake.ParseString(namespace)
		if err == nil && namespaceUID > 0 {
			return namespaceUID, nil
		}
	}
	return 0, merr.ErrorForbidden("namespace is required, please set %s header", cnst.HTTPHeaderXNamespace)
}

// ListNmspacesRole 分页查询租户角色列表
func (s *NmspacesRoleService) ListNmspacesRole(ctx context.Context, req *apiv1.ListNmspacesRoleRequest) (*apiv1.ListNmspacesRoleReply, error) {
	namespaceUID, err := s.getNamespaceUID(ctx)
	if err != nil {
		return nil, err
	}
	listBo := bo.NewListNmspacesRoleBo(namespaceUID, req)
	page, err := s.nmspacesRoleBiz.ListNmspacesRole(ctx, listBo)
	if err != nil {
		return nil, err
	}
	return bo.ToAPIV1ListNmspacesRoleReply(page), nil
}

// GetNmspacesRole 查询租户角色详情
func (s *NmspacesRoleService) GetNmspacesRole(ctx context.Context, req *apiv1.GetNmspacesRoleRequest) (*apiv1.GetNmspacesRoleReply, error) {
	namespaceUID, err := s.getNamespaceUID(ctx)
	if err != nil {
		return nil, err
	}
	result, err := s.nmspacesRoleBiz.GetNmspacesRole(ctx, namespaceUID, req.RoleCode)
	if err != nil {
		return nil, err
	}
	return &apiv1.GetNmspacesRoleReply{
		Role: result.ToAPIV1NmspacesRoleItem(),
	}, nil
}

// CreateNmspacesRole 新增租户角色
func (s *NmspacesRoleService) CreateNmspacesRole(ctx context.Context, req *apiv1.CreateNmspacesRoleRequest) (*apiv1.CreateNmspacesRoleReply, error) {
	namespaceUID, err := s.getNamespaceUID(ctx)
	if err != nil {
		return nil, err
	}
	creator := contextx.GetUserUID(ctx)
	createBo := bo.NewCreateNmspacesRoleBo(namespaceUID, creator, req)
	if err := s.nmspacesRoleBiz.CreateNmspacesRole(ctx, createBo); err != nil {
		return nil, err
	}
	return &apiv1.CreateNmspacesRoleReply{
		RoleCode: req.RoleCode,
	}, nil
}

// UpdateNmspacesRole 编辑租户角色
func (s *NmspacesRoleService) UpdateNmspacesRole(ctx context.Context, req *apiv1.UpdateNmspacesRoleRequest) (*apiv1.UpdateNmspacesRoleReply, error) {
	namespaceUID, err := s.getNamespaceUID(ctx)
	if err != nil {
		return nil, err
	}
	updateBo := bo.NewUpdateNmspacesRoleBo(namespaceUID, req)
	if err := s.nmspacesRoleBiz.UpdateNmspacesRole(ctx, updateBo); err != nil {
		return nil, err
	}
	return &apiv1.UpdateNmspacesRoleReply{
		RoleCode: req.RoleCode,
	}, nil
}

// DeleteNmspacesRole 删除租户角色
func (s *NmspacesRoleService) DeleteNmspacesRole(ctx context.Context, req *apiv1.DeleteNmspacesRoleRequest) (*apiv1.DeleteNmspacesRoleReply, error) {
	namespaceUID, err := s.getNamespaceUID(ctx)
	if err != nil {
		return nil, err
	}
	if err := s.nmspacesRoleBiz.DeleteNmspacesRole(ctx, namespaceUID, req.RoleCode); err != nil {
		return nil, err
	}
	return &apiv1.DeleteNmspacesRoleReply{
		Success: true,
	}, nil
}
