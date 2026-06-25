package biz

import (
	"context"

	klog "github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/magicbox/merr"

	"github.com/aide-family/goddess/internal/biz/bo"
	"github.com/aide-family/goddess/internal/biz/repository"
)

// NewNmspacesRoleFunction 创建租户角色功能业务逻辑实例
func NewNmspacesRoleFunction(
	nmspacesRoleFunctionRepo repository.NmspacesRoleFunction,
	nmspacesFunctionRepo repository.NmspacesFunction,
	nmspacesRoleMenuRepo repository.NmspacesRoleMenu,
	nmspacesRoleRepo repository.NmspacesRole,
	helper *klog.Helper,
) *NmspacesRoleFunctionBiz {
	return &NmspacesRoleFunctionBiz{
		nmspacesRoleFunctionRepo: nmspacesRoleFunctionRepo,
		nmspacesFunctionRepo:     nmspacesFunctionRepo,
		nmspacesRoleMenuRepo:     nmspacesRoleMenuRepo,
		nmspacesRoleRepo:         nmspacesRoleRepo,
		helper:                   klog.NewHelper(klog.With(helper.Logger(), "biz", "nmspaces_role_function")),
	}
}

// NmspacesRoleFunctionBiz 租户角色功能业务逻辑
type NmspacesRoleFunctionBiz struct {
	helper                   *klog.Helper
	nmspacesRoleFunctionRepo repository.NmspacesRoleFunction
	nmspacesFunctionRepo     repository.NmspacesFunction
	nmspacesRoleMenuRepo     repository.NmspacesRoleMenu
	nmspacesRoleRepo         repository.NmspacesRole
}

// ListNmspacesRoleFunction 分页查询角色的功能列表
func (s *NmspacesRoleFunctionBiz) ListNmspacesRoleFunction(ctx context.Context, req *bo.ListNmspacesRoleFunctionBo) ([]*bo.NmspacesRoleFunctionItemBo, int64, error) {
	items, total, err := s.nmspacesRoleFunctionRepo.ListNmspacesRoleFunction(ctx, req)
	if err != nil {
		s.helper.Errorw("msg", "list nmspaces role function failed", "error", err, "req", req)
		return nil, 0, merr.ErrorInternalServer("list nmspaces role function failed").WithCause(err)
	}
	return items, total, nil
}

// GrantNmspacesRoleFunctions 授予角色功能
func (s *NmspacesRoleFunctionBiz) GrantNmspacesRoleFunctions(ctx context.Context, req *bo.GrantNmspacesRoleFunctionsBo) (*bo.GrantFunctionResultBo, error) {
	// 1. 校验 roleCode 存在且启用
	role, err := s.nmspacesRoleRepo.GetNmspacesRole(ctx, req.NamespaceUID, req.RoleCode)
	if err != nil {
		s.helper.Errorw("msg", "check role existence failed", "error", err, "roleCode", req.RoleCode)
		return nil, merr.ErrorInternalServer("check role %s failed", req.RoleCode).WithCause(err)
	}
	if role == nil {
		return nil, merr.ErrorNotFound("role %s not found", req.RoleCode)
	}
	if role.RoleStatus != 1 {
		return nil, merr.ErrorBadRequest("role %s is not enabled", req.RoleCode)
	}

	// 2. 校验每个 functionKey 在当前租户存在且未删除
	for _, fk := range req.FunctionKeys {
		exists, err := s.nmspacesFunctionRepo.ExistsByNamespaceAndMenuCodeAndFunctionCodeNotDeleted(ctx, req.NamespaceUID, fk.MenuCode, fk.FunctionCode)
		if err != nil {
			s.helper.Errorw("msg", "check function existence failed", "error", err)
			return nil, merr.ErrorInternalServer("check function %s/%s failed", fk.MenuCode, fk.FunctionCode).WithCause(err)
		}
		if !exists {
			return nil, merr.ErrorNotFound("function %s/%s not found or deleted", fk.MenuCode, fk.FunctionCode)
		}
	}

	// 3. 批量插入功能授权
	grantedCount, err := s.nmspacesRoleFunctionRepo.GrantNmspacesRoleFunctions(ctx, req)
	if err != nil {
		s.helper.Errorw("msg", "grant nmspaces role functions failed", "error", err, "req", req)
		return nil, merr.ErrorInternalServer("grant nmspaces role functions failed").WithCause(err)
	}

	// 4. 自动追加菜单授权
	var autoGrantedMenus []string
	if req.AutoGrantMenu {
		menuSet := make(map[string]struct{})
		for _, fk := range req.FunctionKeys {
			menuSet[fk.MenuCode] = struct{}{}
		}

		var menusToAdd []string
		for menuCode := range menuSet {
			exists, err := s.nmspacesRoleMenuRepo.ExistsByNamespaceAndRoleCodeAndMenuCode(ctx, req.NamespaceUID, req.RoleCode, menuCode)
			if err != nil {
				s.helper.Errorw("msg", "check role menu existence failed", "error", err)
				return nil, merr.ErrorInternalServer("check role menu failed").WithCause(err)
			}
			if !exists {
				menusToAdd = append(menusToAdd, menuCode)
			}
		}

		if len(menusToAdd) > 0 {
			grantMenuBo := &bo.GrantNmspacesRoleMenusBo{
				NamespaceUID: req.NamespaceUID,
				RoleCode:     req.RoleCode,
				MenuCodes:    menusToAdd,
				Creator:      req.Creator,
			}
			_, err := s.nmspacesRoleMenuRepo.GrantNmspacesRoleMenus(ctx, grantMenuBo)
			if err != nil {
				s.helper.Errorw("msg", "auto grant menus failed", "error", err)
				return nil, merr.ErrorInternalServer("auto grant menus failed").WithCause(err)
			}
			autoGrantedMenus = menusToAdd
		}
	}

	return &bo.GrantFunctionResultBo{
		GrantedCount:     grantedCount,
		AutoGrantedMenus: autoGrantedMenus,
	}, nil
}

// RevokeNmspacesRoleFunctions 撤销角色功能
func (s *NmspacesRoleFunctionBiz) RevokeNmspacesRoleFunctions(ctx context.Context, req *bo.RevokeNmspacesRoleFunctionsBo) (*bo.RevokeFunctionResultBo, error) {
	// 1. 校验 roleCode 存在
	role, err := s.nmspacesRoleRepo.GetNmspacesRole(ctx, req.NamespaceUID, req.RoleCode)
	if err != nil {
		s.helper.Errorw("msg", "check role existence failed", "error", err, "roleCode", req.RoleCode)
		return nil, merr.ErrorInternalServer("check role %s failed", req.RoleCode).WithCause(err)
	}
	if role == nil {
		return nil, merr.ErrorNotFound("role %s not found", req.RoleCode)
	}

	// 2. 执行硬删除功能授权
	revokedCount, err := s.nmspacesRoleFunctionRepo.RevokeNmspacesRoleFunctions(ctx, req)
	if err != nil {
		s.helper.Errorw("msg", "revoke nmspaces role functions failed", "error", err, "req", req)
		return nil, merr.ErrorInternalServer("revoke nmspaces role functions failed").WithCause(err)
	}

	// 3. 自动撤销无功能的菜单授权
	var autoRevokedMenus []string
	if req.AutoRevokeMenu {
		menuSet := make(map[string]struct{})
		for _, fk := range req.FunctionKeys {
			menuSet[fk.MenuCode] = struct{}{}
		}

		for menuCode := range menuSet {
			remainingCount, err := s.nmspacesRoleFunctionRepo.CountByNamespaceAndRoleCodeAndMenuCode(ctx, req.NamespaceUID, req.RoleCode, menuCode)
			if err != nil {
				s.helper.Errorw("msg", "count remaining functions failed", "error", err)
				return nil, merr.ErrorInternalServer("count remaining functions failed").WithCause(err)
			}
			if remainingCount == 0 {
				revokeMenuBo := &bo.RevokeNmspacesRoleMenusBo{
					NamespaceUID: req.NamespaceUID,
					RoleCode:     req.RoleCode,
					MenuCodes:    []string{menuCode},
				}
				_, err := s.nmspacesRoleMenuRepo.RevokeNmspacesRoleMenus(ctx, revokeMenuBo)
				if err != nil {
					s.helper.Errorw("msg", "auto revoke menu failed", "error", err)
					return nil, merr.ErrorInternalServer("auto revoke menu failed").WithCause(err)
				}
				autoRevokedMenus = append(autoRevokedMenus, menuCode)
			}
		}
	}

	return &bo.RevokeFunctionResultBo{
		RevokedCount:     revokedCount,
		AutoRevokedMenus: autoRevokedMenus,
	}, nil
}
