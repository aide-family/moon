package service

import (
	"context"

	"github.com/aide-family/goddess/internal/biz"
	"github.com/aide-family/goddess/internal/biz/bo"
	apiv1 "github.com/aide-family/goddess/pkg/api/v1"
	"github.com/aide-family/magicbox/contextx"
)

// NewSysRoleService 创建系统角色服务实例
func NewSysRoleService(sysRoleBiz *biz.SysRole) *SysRoleService {
	return &SysRoleService{
		sysRoleBiz: sysRoleBiz,
	}
}

// SysRoleService 系统角色服务，实现 proto 定义的 SysRoleServer 接口
type SysRoleService struct {
	apiv1.UnimplementedSysRoleServer

	sysRoleBiz *biz.SysRole
}

// ListSysRole 分页查询系统角色列表
func (s *SysRoleService) ListSysRole(ctx context.Context, req *apiv1.ListSysRoleRequest) (*apiv1.ListSysRoleReply, error) {
	listBo := bo.NewListSysRoleBo(req)
	page, err := s.sysRoleBiz.ListSysRole(ctx, listBo)
	if err != nil {
		return nil, err
	}
	return bo.ToAPIV1ListSysRoleReply(page), nil
}

// GetSysRole 查询系统角色详情（含菜单和功能权限）
func (s *SysRoleService) GetSysRole(ctx context.Context, req *apiv1.GetSysRoleRequest) (*apiv1.GetSysRoleReply, error) {
	result, err := s.sysRoleBiz.GetSysRole(ctx, req.RoleCode)
	if err != nil {
		return nil, err
	}
	return result.ToAPIV1GetSysRoleReply(), nil
}

// CreateSysRole 新增系统角色
func (s *SysRoleService) CreateSysRole(ctx context.Context, req *apiv1.CreateSysRoleRequest) (*apiv1.CreateSysRoleReply, error) {
	createBo := bo.NewCreateSysRoleBo(req)
	createBo.Creator = int64(contextx.GetUserUID(ctx))
	if err := s.sysRoleBiz.CreateSysRole(ctx, createBo); err != nil {
		return nil, err
	}
	return &apiv1.CreateSysRoleReply{
		RoleCode: req.RoleCode,
	}, nil
}

// UpdateSysRole 编辑系统角色
func (s *SysRoleService) UpdateSysRole(ctx context.Context, req *apiv1.UpdateSysRoleRequest) (*apiv1.UpdateSysRoleReply, error) {
	updateBo := bo.NewUpdateSysRoleBo(req)
	if err := s.sysRoleBiz.UpdateSysRole(ctx, updateBo); err != nil {
		return nil, err
	}
	return &apiv1.UpdateSysRoleReply{
		RoleCode: req.RoleCode,
	}, nil
}

// DeleteSysRole 删除系统角色
func (s *SysRoleService) DeleteSysRole(ctx context.Context, req *apiv1.DeleteSysRoleRequest) (*apiv1.DeleteSysRoleReply, error) {
	if err := s.sysRoleBiz.DeleteSysRole(ctx, req.RoleCode); err != nil {
		return nil, err
	}
	return &apiv1.DeleteSysRoleReply{
		Success: true,
	}, nil
}
