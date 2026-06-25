package biz

import (
	"context"

	klog "github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/magicbox/merr"
	"github.com/bwmarrin/snowflake"

	"github.com/aide-family/goddess/internal/biz/bo"
	"github.com/aide-family/goddess/internal/biz/repository"
)

// NewNmspacesRole 创建租户角色业务逻辑实例
func NewNmspacesRole(nmspacesRoleRepo repository.NmspacesRole, helper *klog.Helper) *NmspacesRoleBiz {
	return &NmspacesRoleBiz{
		nmspacesRoleRepo: nmspacesRoleRepo,
		helper:           klog.NewHelper(klog.With(helper.Logger(), "biz", "nmspaces_role")),
	}
}

// NmspacesRoleBiz 租户角色业务逻辑
type NmspacesRoleBiz struct {
	helper           *klog.Helper
	nmspacesRoleRepo repository.NmspacesRole
}

// ListNmspacesRole 分页查询租户角色列表
func (s *NmspacesRoleBiz) ListNmspacesRole(ctx context.Context, req *bo.ListNmspacesRoleBo) (*bo.PageResponseBo[*bo.NmspacesRoleItemBo], error) {
	page, err := s.nmspacesRoleRepo.ListNmspacesRole(ctx, req)
	if err != nil {
		s.helper.Errorw("msg", "list nmspaces role failed", "error", err, "req", req)
		return nil, merr.ErrorInternalServer("list nmspaces role failed").WithCause(err)
	}
	return page, nil
}

// GetNmspacesRole 查询租户角色详情
func (s *NmspacesRoleBiz) GetNmspacesRole(ctx context.Context, namespaceUID snowflake.ID, roleCode string) (*bo.NmspacesRoleItemBo, error) {
	result, err := s.nmspacesRoleRepo.GetNmspacesRole(ctx, namespaceUID, roleCode)
	if err != nil {
		s.helper.Errorw("msg", "get nmspaces role failed", "error", err, "namespaceUID", namespaceUID, "roleCode", roleCode)
		return nil, merr.ErrorInternalServer("get nmspaces role failed").WithCause(err)
	}
	if result == nil {
		return nil, merr.ErrorNotFound("role %s not found in namespace %d", roleCode, namespaceUID)
	}
	return result, nil
}

// CreateNmspacesRole 新增租户角色
func (s *NmspacesRoleBiz) CreateNmspacesRole(ctx context.Context, req *bo.CreateNmspacesRoleBo) error {
	// 校验 roleCode 不存在（含软删除行）
	exists, err := s.nmspacesRoleRepo.ExistsByNamespaceAndRoleCode(ctx, req.NamespaceUID, req.RoleCode)
	if err != nil {
		s.helper.Errorw("msg", "check roleCode existence failed", "error", err, "roleCode", req.RoleCode)
		return merr.ErrorInternalServer("check roleCode failed").WithCause(err)
	}
	if exists {
		return merr.ErrorBadRequest("roleCode %s already exists in this namespace", req.RoleCode)
	}

	// 校验 roleName 在未删除行中唯一
	nameExists, err := s.nmspacesRoleRepo.ExistsByNamespaceAndRoleName(ctx, req.NamespaceUID, req.RoleName)
	if err != nil {
		s.helper.Errorw("msg", "check roleName uniqueness failed", "error", err, "roleName", req.RoleName)
		return merr.ErrorInternalServer("check roleName failed").WithCause(err)
	}
	if nameExists {
		return merr.ErrorBadRequest("roleName %s already exists in this namespace", req.RoleName)
	}

	if err := s.nmspacesRoleRepo.CreateNmspacesRole(ctx, req); err != nil {
		s.helper.Errorw("msg", "create nmspaces role failed", "error", err, "req", req)
		return merr.ErrorInternalServer("create nmspaces role failed").WithCause(err)
	}
	return nil
}

// UpdateNmspacesRole 编辑租户角色
func (s *NmspacesRoleBiz) UpdateNmspacesRole(ctx context.Context, req *bo.UpdateNmspacesRoleBo) error {
	// 校验角色存在且未删除
	exists, err := s.nmspacesRoleRepo.ExistsByNamespaceAndRoleCodeNotDeleted(ctx, req.NamespaceUID, req.RoleCode)
	if err != nil {
		s.helper.Errorw("msg", "check role existence failed", "error", err, "roleCode", req.RoleCode)
		return merr.ErrorInternalServer("check role existence failed").WithCause(err)
	}
	if !exists {
		return merr.ErrorNotFound("role %s not found in namespace %d", req.RoleCode, req.NamespaceUID)
	}

	// 若修改 roleName，校验未删除行中唯一
	if req.RoleName != nil {
		nameExists, err := s.nmspacesRoleRepo.ExistsByNamespaceAndRoleName(ctx, req.NamespaceUID, *req.RoleName)
		if err != nil {
			s.helper.Errorw("msg", "check roleName uniqueness failed", "error", err, "roleName", *req.RoleName)
			return merr.ErrorInternalServer("check roleName failed").WithCause(err)
		}
		if nameExists {
			return merr.ErrorBadRequest("roleName %s already exists in this namespace", *req.RoleName)
		}
	}

	if err := s.nmspacesRoleRepo.UpdateNmspacesRole(ctx, req); err != nil {
		s.helper.Errorw("msg", "update nmspaces role failed", "error", err, "req", req)
		return merr.ErrorInternalServer("update nmspaces role failed").WithCause(err)
	}
	return nil
}

// DeleteNmspacesRole 删除租户角色（软删除）
func (s *NmspacesRoleBiz) DeleteNmspacesRole(ctx context.Context, namespaceUID snowflake.ID, roleCode string) error {
	// 校验角色存在且未删除
	exists, err := s.nmspacesRoleRepo.ExistsByNamespaceAndRoleCodeNotDeleted(ctx, namespaceUID, roleCode)
	if err != nil {
		s.helper.Errorw("msg", "check role existence failed", "error", err, "roleCode", roleCode)
		return merr.ErrorInternalServer("check role existence failed").WithCause(err)
	}
	if !exists {
		return merr.ErrorNotFound("role %s not found in namespace %d", roleCode, namespaceUID)
	}

	// 执行软删除
	if err := s.nmspacesRoleRepo.DeleteNmspacesRole(ctx, namespaceUID, roleCode); err != nil {
		s.helper.Errorw("msg", "delete nmspaces role failed", "error", err, "roleCode", roleCode)
		return merr.ErrorInternalServer("delete nmspaces role failed").WithCause(err)
	}
	return nil
}
