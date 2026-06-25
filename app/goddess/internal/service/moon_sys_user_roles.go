package service

import (
	"context"

	"github.com/aide-family/goddess/internal/biz"
	"github.com/aide-family/goddess/internal/biz/bo"
	"github.com/aide-family/goddess/internal/data/impl/convert"
	apiv1 "github.com/aide-family/goddess/pkg/api/v1"
	"github.com/aide-family/magicbox/contextx"
)

// NewSysUserRoleService 创建系统用户角色服务实例
func NewSysUserRoleService(sysUserRoleBiz *biz.SysUserRole) *SysUserRoleService {
	return &SysUserRoleService{
		sysUserRoleBiz: sysUserRoleBiz,
	}
}

// SysUserRoleService 系统用户角色服务，实现 proto 定义的 SysUserRoleServer 接口
type SysUserRoleService struct {
	apiv1.UnimplementedSysUserRoleServer

	sysUserRoleBiz *biz.SysUserRole
}

// ListSysUserRole 分页查询用户的系统角色列表
func (s *SysUserRoleService) ListSysUserRole(ctx context.Context, req *apiv1.ListSysUserRoleRequest) (*apiv1.ListSysUserRoleReply, error) {
	listBo := bo.NewListSysUserRoleBo(req)
	items, total, err := s.sysUserRoleBiz.ListSysUserRole(ctx, listBo)
	if err != nil {
		return nil, err
	}
	apiItems := make([]*apiv1.SysUserRoleItem, 0, len(items))
	for _, item := range items {
		apiItems = append(apiItems, convert.SysUserRoleItemBotoAPIV1(item))
	}
	return &apiv1.ListSysUserRoleReply{
		Items:    apiItems,
		Total:    total,
		Page:     listBo.PageRequestBo.Page,
		PageSize: listBo.PageRequestBo.PageSize,
	}, nil
}

// BindSysUserRoles 绑定系统用户角色
func (s *SysUserRoleService) BindSysUserRoles(ctx context.Context, req *apiv1.BindSysUserRolesRequest) (*apiv1.BindSysUserRolesReply, error) {
	bindBo := bo.NewBindSysUserRolesBo(req)
	bindBo.Creator = contextx.GetUserUID(ctx)
	boundCount, err := s.sysUserRoleBiz.BindSysUserRoles(ctx, bindBo)
	if err != nil {
		return nil, err
	}
	return &apiv1.BindSysUserRolesReply{
		BoundCount: boundCount,
	}, nil
}

// UnbindSysUserRoles 解绑系统用户角色
func (s *SysUserRoleService) UnbindSysUserRoles(ctx context.Context, req *apiv1.UnbindSysUserRolesRequest) (*apiv1.UnbindSysUserRolesReply, error) {
	unbindBo := bo.NewUnbindSysUserRolesBo(req)
	unboundCount, err := s.sysUserRoleBiz.UnbindSysUserRoles(ctx, unbindBo)
	if err != nil {
		return nil, err
	}
	return &apiv1.UnbindSysUserRolesReply{
		UnboundCount: unboundCount,
	}, nil
}
