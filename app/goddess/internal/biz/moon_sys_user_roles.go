package biz

import (
	"context"

	klog "github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/magicbox/merr"

	"github.com/aide-family/goddess/internal/biz/bo"
	"github.com/aide-family/goddess/internal/biz/repository"
)

// NewSysUserRole 创建系统用户角色业务逻辑实例
func NewSysUserRole(
	sysUserRoleRepo repository.SysUserRole,
	sysRoleRepo repository.SysRole,
	userRepo repository.User,
	helper *klog.Helper,
) *SysUserRole {
	return &SysUserRole{
		sysUserRoleRepo: sysUserRoleRepo,
		sysRoleRepo:     sysRoleRepo,
		userRepo:        userRepo,
		helper:          klog.NewHelper(klog.With(helper.Logger(), "biz", "sys_user_role")),
	}
}

// SysUserRole 系统用户角色业务逻辑
type SysUserRole struct {
	helper          *klog.Helper
	sysUserRoleRepo repository.SysUserRole
	sysRoleRepo     repository.SysRole
	userRepo        repository.User
}

// ListSysUserRole 分页查询用户的系统角色列表
func (s *SysUserRole) ListSysUserRole(ctx context.Context, req *bo.ListSysUserRoleBo) ([]*bo.SysUserRoleItemBo, int64, error) {
	items, total, err := s.sysUserRoleRepo.ListSysUserRole(ctx, req)
	if err != nil {
		s.helper.Errorw("msg", "list sys user role failed", "error", err, "req", req)
		return nil, 0, merr.ErrorInternalServer("list sys user role failed").WithCause(err)
	}
	return items, total, nil
}

// BindSysUserRoles 批量绑定用户角色
func (s *SysUserRole) BindSysUserRoles(ctx context.Context, req *bo.BindSysUserRolesBo) (int32, error) {
	// 1. 校验目标 userUID 存在于 users 表
	if _, err := s.userRepo.GetUser(ctx, req.UserUID); err != nil {
		s.helper.Errorw("msg", "check user existence failed", "error", err, "userUID", req.UserUID)
		return 0, merr.ErrorNotFound("user %d not found", req.UserUID)
	}

	// 2. 校验所有 roleCodes 在 moon_sys_roles 中存在且启用
	for _, roleCode := range req.RoleCodes {
		role, err := s.sysRoleRepo.GetSysRole(ctx, roleCode)
		if err != nil {
			s.helper.Errorw("msg", "check role existence failed", "error", err, "roleCode", roleCode)
			return 0, merr.ErrorInternalServer("check role %s failed", roleCode).WithCause(err)
		}
		if role == nil || role.SysRoleItemBo == nil {
			return 0, merr.ErrorNotFound("role %s not found", roleCode)
		}
		if role.RoleStatus != 1 {
			return 0, merr.ErrorBadRequest("role %s is not enabled", roleCode)
		}
	}

	// 3. 批量插入（冲突时忽略，幂等）
	boundCount, err := s.sysUserRoleRepo.BindSysUserRoles(ctx, req)
	if err != nil {
		s.helper.Errorw("msg", "bind sys user roles failed", "error", err, "req", req)
		return 0, merr.ErrorInternalServer("bind sys user roles failed").WithCause(err)
	}
	return boundCount, nil
}

// UnbindSysUserRoles 批量解绑用户角色
func (s *SysUserRole) UnbindSysUserRoles(ctx context.Context, req *bo.UnbindSysUserRolesBo) (int32, error) {
	// 1. 校验目标 userUID 存在于 users 表
	if _, err := s.userRepo.GetUser(ctx, req.UserUID); err != nil {
		s.helper.Errorw("msg", "check user existence failed", "error", err, "userUID", req.UserUID)
		return 0, merr.ErrorNotFound("user %d not found", req.UserUID)
	}

	// 2. 校验解绑后该用户仍至少保留一个系统角色（防止失权）
	currentCount, err := s.sysUserRoleRepo.CountByUserUID(ctx, req.UserUID)
	if err != nil {
		s.helper.Errorw("msg", "count user roles failed", "error", err, "userUID", req.UserUID)
		return 0, merr.ErrorInternalServer("count user roles failed").WithCause(err)
	}

	removeCount, err := s.sysUserRoleRepo.CountByUserUIDAndRoleCodes(ctx, req.UserUID, req.RoleCodes)
	if err != nil {
		s.helper.Errorw("msg", "count user roles by codes failed", "error", err, "userUID", req.UserUID, "roleCodes", req.RoleCodes)
		return 0, merr.ErrorInternalServer("count user roles by codes failed").WithCause(err)
	}

	if currentCount-removeCount < 1 {
		return 0, merr.ErrorBadRequest("user %d must retain at least one system role", req.UserUID)
	}

	// 3. 硬删除
	unboundCount, err := s.sysUserRoleRepo.UnbindSysUserRoles(ctx, req)
	if err != nil {
		s.helper.Errorw("msg", "unbind sys user roles failed", "error", err, "req", req)
		return 0, merr.ErrorInternalServer("unbind sys user roles failed").WithCause(err)
	}
	return unboundCount, nil
}
