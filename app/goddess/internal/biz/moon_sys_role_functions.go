package biz

import (
	"context"

	klog "github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/magicbox/merr"

	"github.com/aide-family/goddess/internal/biz/bo"
	"github.com/aide-family/goddess/internal/biz/repository"
)

// NewSysRoleFunction 创建系统角色功能业务逻辑实例
func NewSysRoleFunction(
	sysRoleFunctionRepo repository.SysRoleFunction,
	sysFunctionRepo repository.SysFunction,
	sysRoleMenuRepo repository.SysRoleMenu,
	sysRoleRepo repository.SysRole,
	helper *klog.Helper,
) *SysRoleFunction {
	return &SysRoleFunction{
		sysRoleFunctionRepo: sysRoleFunctionRepo,
		sysFunctionRepo:     sysFunctionRepo,
		sysRoleMenuRepo:     sysRoleMenuRepo,
		sysRoleRepo:         sysRoleRepo,
		helper:              klog.NewHelper(klog.With(helper.Logger(), "biz", "sys_role_function")),
	}
}

// SysRoleFunction 系统角色功能业务逻辑
type SysRoleFunction struct {
	helper              *klog.Helper
	sysRoleFunctionRepo repository.SysRoleFunction
	sysFunctionRepo     repository.SysFunction
	sysRoleMenuRepo     repository.SysRoleMenu
	sysRoleRepo         repository.SysRole
}

// ListRoleFunction 分页查询角色的功能列表
func (s *SysRoleFunction) ListRoleFunction(ctx context.Context, req *bo.ListRoleFunctionBo) ([]*bo.RoleFunctionItemBo, int64, error) {
	items, total, err := s.sysRoleFunctionRepo.ListRoleFunction(ctx, req)
	if err != nil {
		s.helper.Errorw("msg", "list role function failed", "error", err, "req", req)
		return nil, 0, merr.ErrorInternalServer("list role function failed").WithCause(err)
	}
	return items, total, nil
}

// GrantRoleFunctions 授予角色功能
func (s *SysRoleFunction) GrantRoleFunctions(ctx context.Context, req *bo.GrantRoleFunctionsBo) (*bo.GrantFunctionResultBo, error) {
	// 1. 校验 roleCode 存在且启用
	role, err := s.sysRoleRepo.GetSysRole(ctx, req.RoleCode)
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

	// 2. 校验每个 functionKey 在 moon_sys_functions 中存在且未删除
	for _, fk := range req.FunctionKeys {
		exists, err := s.sysFunctionRepo.ExistsByMenuCodeAndFunctionCodeNotDeleted(ctx, fk.MenuCode, fk.FunctionCode)
		if err != nil {
			s.helper.Errorw("msg", "check function existence failed", "error", err, "menuCode", fk.MenuCode, "functionCode", fk.FunctionCode)
			return nil, merr.ErrorInternalServer("check function %s/%s failed", fk.MenuCode, fk.FunctionCode).WithCause(err)
		}
		if !exists {
			return nil, merr.ErrorNotFound("function %s/%s not found or deleted", fk.MenuCode, fk.FunctionCode)
		}
	}

	// 3. 批量插入功能授权
	grantedCount, err := s.sysRoleFunctionRepo.GrantRoleFunctions(ctx, req)
	if err != nil {
		s.helper.Errorw("msg", "grant role functions failed", "error", err, "req", req)
		return nil, merr.ErrorInternalServer("grant role functions failed").WithCause(err)
	}

	// 4. 自动追加菜单授权（复用 SysRoleMenu.ExistsByRoleCodeAndMenuCode + GrantRoleMenus）
	var autoGrantedMenus []string
	if req.AutoGrantMenu {
		// 收集需要补全的菜单（去重）
		menuSet := make(map[string]struct{})
		for _, fk := range req.FunctionKeys {
			menuSet[fk.MenuCode] = struct{}{}
		}

		var menusToAdd []string
		for menuCode := range menuSet {
			exists, err := s.sysRoleMenuRepo.ExistsByRoleCodeAndMenuCode(ctx, req.RoleCode, menuCode)
			if err != nil {
				s.helper.Errorw("msg", "check role menu existence failed", "error", err, "roleCode", req.RoleCode, "menuCode", menuCode)
				return nil, merr.ErrorInternalServer("check role menu failed").WithCause(err)
			}
			if !exists {
				menusToAdd = append(menusToAdd, menuCode)
			}
		}

		if len(menusToAdd) > 0 {
			grantMenuBo := &bo.GrantRoleMenusBo{
				RoleCode:  req.RoleCode,
				MenuCodes: menusToAdd,
			}
			_, err := s.sysRoleMenuRepo.GrantRoleMenus(ctx, grantMenuBo)
			if err != nil {
				s.helper.Errorw("msg", "auto grant menus failed", "error", err, "menus", menusToAdd)
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

// RevokeRoleFunctions 撤销角色功能
func (s *SysRoleFunction) RevokeRoleFunctions(ctx context.Context, req *bo.RevokeRoleFunctionsBo) (*bo.RevokeFunctionResultBo, error) {
	// 1. 校验 roleCode 存在
	role, err := s.sysRoleRepo.GetSysRole(ctx, req.RoleCode)
	if err != nil {
		s.helper.Errorw("msg", "check role existence failed", "error", err, "roleCode", req.RoleCode)
		return nil, merr.ErrorInternalServer("check role %s failed", req.RoleCode).WithCause(err)
	}
	if role == nil {
		return nil, merr.ErrorNotFound("role %s not found", req.RoleCode)
	}

	// 2. 执行硬删除功能授权
	revokedCount, err := s.sysRoleFunctionRepo.RevokeRoleFunctions(ctx, req)
	if err != nil {
		s.helper.Errorw("msg", "revoke role functions failed", "error", err, "req", req)
		return nil, merr.ErrorInternalServer("revoke role functions failed").WithCause(err)
	}

	// 3. 自动撤销无功能的菜单授权（可选）
	var autoRevokedMenus []string
	if req.AutoRevokeMenu {
		// 收集涉及的菜单
		menuSet := make(map[string]struct{})
		for _, fk := range req.FunctionKeys {
			menuSet[fk.MenuCode] = struct{}{}
		}

		for menuCode := range menuSet {
			// 检查该角色在该菜单下是否还有功能
			remainingCount, err := s.sysRoleFunctionRepo.CountByRoleCodeAndMenuCode(ctx, req.RoleCode, menuCode)
			if err != nil {
				s.helper.Errorw("msg", "count remaining functions failed", "error", err, "roleCode", req.RoleCode, "menuCode", menuCode)
				return nil, merr.ErrorInternalServer("count remaining functions failed").WithCause(err)
			}
			if remainingCount == 0 {
				// 该菜单下无功能，撤销菜单授权
				revokeMenuBo := &bo.RevokeRoleMenusBo{
					RoleCode:  req.RoleCode,
					MenuCodes: []string{menuCode},
				}
				_, err := s.sysRoleMenuRepo.RevokeRoleMenus(ctx, revokeMenuBo)
				if err != nil {
					s.helper.Errorw("msg", "auto revoke menu failed", "error", err, "roleCode", req.RoleCode, "menuCode", menuCode)
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
