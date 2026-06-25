package biz

import (
	"context"

	klog "github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/magicbox/merr"

	"github.com/aide-family/goddess/internal/biz/bo"
	"github.com/aide-family/goddess/internal/biz/repository"
)

// NewNmspacesMemberRole 创建租户成员角色业务逻辑实例
func NewNmspacesMemberRole(
	nmspacesMemberRoleRepo repository.NmspacesMemberRole,
	nmspacesRoleRepo repository.NmspacesRole,
	memberRepo repository.Member,
	helper *klog.Helper,
) *NmspacesMemberRoleBiz {
	return &NmspacesMemberRoleBiz{
		nmspacesMemberRoleRepo: nmspacesMemberRoleRepo,
		nmspacesRoleRepo:       nmspacesRoleRepo,
		memberRepo:             memberRepo,
		helper:                 klog.NewHelper(klog.With(helper.Logger(), "biz", "nmspaces_member_role")),
	}
}

// NmspacesMemberRoleBiz 租户成员角色业务逻辑
type NmspacesMemberRoleBiz struct {
	helper                 *klog.Helper
	nmspacesMemberRoleRepo repository.NmspacesMemberRole
	nmspacesRoleRepo       repository.NmspacesRole
	memberRepo             repository.Member
}

// ListNmspacesMemberRole 分页查询成员的角色列表
func (s *NmspacesMemberRoleBiz) ListNmspacesMemberRole(ctx context.Context, req *bo.ListNmspacesMemberRoleBo) ([]*bo.NmspacesMemberRoleItemBo, int64, error) {
	items, total, err := s.nmspacesMemberRoleRepo.ListNmspacesMemberRole(ctx, req)
	if err != nil {
		s.helper.Errorw("msg", "list nmspaces member role failed", "error", err, "req", req)
		return nil, 0, merr.ErrorInternalServer("list nmspaces member role failed").WithCause(err)
	}
	return items, total, nil
}

// BindNmspacesMemberRoles 批量绑定成员角色
func (s *NmspacesMemberRoleBiz) BindNmspacesMemberRoles(ctx context.Context, req *bo.BindNmspacesMemberRolesBo) (int32, error) {
	// 1. 校验 memberUID 存在于 members 表
	member, err := s.memberRepo.GetMember(ctx, req.MemberUID)
	if err != nil {
		s.helper.Errorw("msg", "check member existence failed", "error", err, "memberUID", req.MemberUID)
		return 0, merr.ErrorNotFound("member %d not found", req.MemberUID)
	}
	if member == nil {
		return 0, merr.ErrorNotFound("member %d not found", req.MemberUID)
	}

	// 2. 校验所有 roleCodes 在当前租户存在且启用
	for _, roleCode := range req.RoleCodes {
		role, err := s.nmspacesRoleRepo.GetNmspacesRole(ctx, req.NamespaceUID, roleCode)
		if err != nil {
			s.helper.Errorw("msg", "check role existence failed", "error", err, "roleCode", roleCode)
			return 0, merr.ErrorInternalServer("check role %s failed", roleCode).WithCause(err)
		}
		if role == nil {
			return 0, merr.ErrorNotFound("role %s not found", roleCode)
		}
		if role.RoleStatus != 1 {
			return 0, merr.ErrorBadRequest("role %s is not enabled", roleCode)
		}
	}

	// 3. 批量插入
	boundCount, err := s.nmspacesMemberRoleRepo.BindNmspacesMemberRoles(ctx, req)
	if err != nil {
		s.helper.Errorw("msg", "bind nmspaces member roles failed", "error", err, "req", req)
		return 0, merr.ErrorInternalServer("bind nmspaces member roles failed").WithCause(err)
	}
	return boundCount, nil
}

// UnbindNmspacesMemberRoles 批量解绑成员角色
func (s *NmspacesMemberRoleBiz) UnbindNmspacesMemberRoles(ctx context.Context, req *bo.UnbindNmspacesMemberRolesBo) (int32, error) {
	// 1. 校验 memberUID 存在
	member, err := s.memberRepo.GetMember(ctx, req.MemberUID)
	if err != nil {
		s.helper.Errorw("msg", "check member existence failed", "error", err, "memberUID", req.MemberUID)
		return 0, merr.ErrorNotFound("member %d not found", req.MemberUID)
	}
	if member == nil {
		return 0, merr.ErrorNotFound("member %d not found", req.MemberUID)
	}

	// 2. 校验解绑后该成员仍至少保留一个角色
	currentCount, err := s.nmspacesMemberRoleRepo.CountByNamespaceAndMemberUID(ctx, req.NamespaceUID, req.MemberUID)
	if err != nil {
		s.helper.Errorw("msg", "count member roles failed", "error", err)
		return 0, merr.ErrorInternalServer("count member roles failed").WithCause(err)
	}

	removeCount, err := s.nmspacesMemberRoleRepo.CountByNamespaceAndMemberUIDAndRoleCodes(ctx, req.NamespaceUID, req.MemberUID, req.RoleCodes)
	if err != nil {
		s.helper.Errorw("msg", "count member roles by codes failed", "error", err)
		return 0, merr.ErrorInternalServer("count member roles by codes failed").WithCause(err)
	}

	if currentCount-removeCount < 1 {
		return 0, merr.ErrorBadRequest("member %d must retain at least one role", req.MemberUID)
	}

	// 3. 硬删除
	unboundCount, err := s.nmspacesMemberRoleRepo.UnbindNmspacesMemberRoles(ctx, req)
	if err != nil {
		s.helper.Errorw("msg", "unbind nmspaces member roles failed", "error", err, "req", req)
		return 0, merr.ErrorInternalServer("unbind nmspaces member roles failed").WithCause(err)
	}
	return unboundCount, nil
}
