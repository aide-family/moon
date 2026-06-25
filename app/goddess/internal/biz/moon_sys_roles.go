package biz

import (
	"context"

	klog "github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/magicbox/merr"

	"github.com/aide-family/goddess/internal/biz/bo"
	"github.com/aide-family/goddess/internal/biz/repository"
)

// NewSysRole 创建系统角色业务逻辑实例
func NewSysRole(sysRoleRepo repository.SysRole, helper *klog.Helper) *SysRole {
	return &SysRole{
		sysRoleRepo: sysRoleRepo,
		helper:      klog.NewHelper(klog.With(helper.Logger(), "biz", "sys_role")),
	}
}

// SysRole 系统角色业务逻辑
type SysRole struct {
	helper      *klog.Helper
	sysRoleRepo repository.SysRole
}

// ListSysRole 分页查询系统角色列表
func (s *SysRole) ListSysRole(ctx context.Context, req *bo.ListSysRoleBo) (*bo.PageResponseBo[*bo.SysRoleItemBo], error) {
	page, err := s.sysRoleRepo.ListSysRole(ctx, req)
	if err != nil {
		s.helper.Errorw("msg", "list sys role failed", "error", err, "req", req)
		return nil, merr.ErrorInternalServer("list sys role failed").WithCause(err)
	}
	return page, nil
}

// GetSysRole 根据 roleCode 查询角色详情
func (s *SysRole) GetSysRole(ctx context.Context, roleCode string) (*bo.GetSysRoleBo, error) {
	result, err := s.sysRoleRepo.GetSysRole(ctx, roleCode)
	if err != nil {
		s.helper.Errorw("msg", "get sys role failed", "error", err, "roleCode", roleCode)
		return nil, merr.ErrorInternalServer("get sys role failed").WithCause(err)
	}
	if result == nil {
		return nil, merr.ErrorNotFound("role %s not found", roleCode)
	}
	return result, nil
}

// CreateSysRole 新增系统角色
func (s *SysRole) CreateSysRole(ctx context.Context, req *bo.CreateSysRoleBo) error {
	// 校验 roleCode 不存在（含软删除行，避免重复注册）
	exists, err := s.sysRoleRepo.ExistsByRoleCode(ctx, req.RoleCode)
	if err != nil {
		s.helper.Errorw("msg", "check roleCode existence failed", "error", err, "roleCode", req.RoleCode)
		return merr.ErrorInternalServer("check roleCode failed").WithCause(err)
	}
	if exists {
		return merr.ErrorBadRequest("roleCode %s already exists", req.RoleCode)
	}

	// 校验 roleName 在未删除行中唯一
	nameExists, err := s.sysRoleRepo.ExistsByRoleName(ctx, req.RoleName)
	if err != nil {
		s.helper.Errorw("msg", "check roleName uniqueness failed", "error", err, "roleName", req.RoleName)
		return merr.ErrorInternalServer("check roleName failed").WithCause(err)
	}
	if nameExists {
		return merr.ErrorBadRequest("roleName %s already exists", req.RoleName)
	}

	// 插入角色
	if err := s.sysRoleRepo.CreateSysRole(ctx, req); err != nil {
		s.helper.Errorw("msg", "create sys role failed", "error", err, "req", req)
		return merr.ErrorInternalServer("create sys role failed").WithCause(err)
	}
	return nil
}

// UpdateSysRole 编辑系统角色
func (s *SysRole) UpdateSysRole(ctx context.Context, req *bo.UpdateSysRoleBo) error {
	// 校验角色存在且未删除
	exists, err := s.sysRoleRepo.ExistsByRoleCodeNotDeleted(ctx, req.RoleCode)
	if err != nil {
		s.helper.Errorw("msg", "check role existence failed", "error", err, "roleCode", req.RoleCode)
		return merr.ErrorInternalServer("check role existence failed").WithCause(err)
	}
	if !exists {
		return merr.ErrorNotFound("role %s not found", req.RoleCode)
	}

	// 若修改 roleName，校验未删除行中唯一
	if req.RoleName != nil {
		nameExists, err := s.sysRoleRepo.ExistsByRoleName(ctx, *req.RoleName)
		if err != nil {
			s.helper.Errorw("msg", "check roleName uniqueness failed", "error", err, "roleName", *req.RoleName)
			return merr.ErrorInternalServer("check roleName failed").WithCause(err)
		}
		if nameExists {
			return merr.ErrorBadRequest("roleName %s already exists", *req.RoleName)
		}
	}

	// 执行更新
	if err := s.sysRoleRepo.UpdateSysRole(ctx, req); err != nil {
		s.helper.Errorw("msg", "update sys role failed", "error", err, "req", req)
		return merr.ErrorInternalServer("update sys role failed").WithCause(err)
	}
	return nil
}

// DeleteSysRole 删除系统角色（软删除，role_status 置为 0）
func (s *SysRole) DeleteSysRole(ctx context.Context, roleCode string) error {
	// 校验角色存在且未删除
	exists, err := s.sysRoleRepo.ExistsByRoleCodeNotDeleted(ctx, roleCode)
	if err != nil {
		s.helper.Errorw("msg", "check role existence failed", "error", err, "roleCode", roleCode)
		return merr.ErrorInternalServer("check role existence failed").WithCause(err)
	}
	if !exists {
		return merr.ErrorNotFound("role %s not found", roleCode)
	}

	// 执行软删除
	if err := s.sysRoleRepo.DeleteSysRole(ctx, roleCode); err != nil {
		s.helper.Errorw("msg", "delete sys role failed", "error", err, "roleCode", roleCode)
		return merr.ErrorInternalServer("delete sys role failed").WithCause(err)
	}
	return nil
}
