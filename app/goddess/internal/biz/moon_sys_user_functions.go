package biz

import (
	"context"

	klog "github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/magicbox/merr"

	"github.com/aide-family/goddess/internal/biz/bo"
	"github.com/aide-family/goddess/internal/biz/repository"
)

// NewSysUserFunction 创建系统用户功能直授业务逻辑实例
func NewSysUserFunction(
	sysUserFunctionRepo repository.SysUserFunction,
	sysFunctionRepo repository.SysFunction,
	sysUserMenuRepo repository.SysUserMenu,
	userRepo repository.User,
	helper *klog.Helper,
) *SysUserFunction {
	return &SysUserFunction{
		sysUserFunctionRepo: sysUserFunctionRepo,
		sysFunctionRepo:     sysFunctionRepo,
		sysUserMenuRepo:     sysUserMenuRepo,
		userRepo:            userRepo,
		helper:              klog.NewHelper(klog.With(helper.Logger(), "biz", "sys_user_function")),
	}
}

// SysUserFunction 系统用户功能直授业务逻辑
type SysUserFunction struct {
	helper              *klog.Helper
	sysUserFunctionRepo repository.SysUserFunction
	sysFunctionRepo     repository.SysFunction
	sysUserMenuRepo     repository.SysUserMenu
	userRepo            repository.User
}

// ListUserFunction 查询用户直授功能列表
func (s *SysUserFunction) ListUserFunction(ctx context.Context, req *bo.ListUserFunctionBo) ([]*bo.UserFunctionItemBo, int64, error) {
	items, total, err := s.sysUserFunctionRepo.ListUserFunction(ctx, req)
	if err != nil {
		s.helper.Errorw("msg", "list user function failed", "error", err, "req", req)
		return nil, 0, merr.ErrorInternalServer("list user function failed").WithCause(err)
	}
	return items, total, nil
}

// GrantUserFunctions 直授用户功能
func (s *SysUserFunction) GrantUserFunctions(ctx context.Context, req *bo.GrantUserFunctionsBo) (*bo.GrantUserFunctionResultBo, error) {
	// 1. 校验 userUID 存在
	user, err := s.userRepo.GetUser(ctx, req.UserUID)
	if err != nil {
		s.helper.Errorw("msg", "check user existence failed", "error", err, "userUID", req.UserUID)
		return nil, merr.ErrorInternalServer("check user %d failed", req.UserUID).WithCause(err)
	}
	if user == nil {
		return nil, merr.ErrorNotFound("user %d not found", req.UserUID)
	}

	// 2. 校验每个 functionKey 在 moon_sys_functions 中存在且未删除
	for _, fk := range req.FunctionKeys {
		exists, err := s.sysFunctionRepo.ExistsByMenuCodeAndFunctionCodeNotDeleted(ctx, fk.MenuCode, fk.FunctionCode)
		if err != nil {
			s.helper.Errorw("msg", "check function existence failed", "error", err, "menuCode", fk.MenuCode, "functionCode", fk.FunctionCode)
			return nil, merr.ErrorInternalServer("check function %s/%s failed", fk.MenuCode, fk.FunctionCode).WithCause(err)
		}
		if !exists {
			return nil, merr.ErrorNotFound("function %s/%s not found", fk.MenuCode, fk.FunctionCode)
		}
	}

	// 3. 批量插入
	grantedCount, err := s.sysUserFunctionRepo.GrantUserFunctions(ctx, req)
	if err != nil {
		s.helper.Errorw("msg", "grant user functions failed", "error", err, "req", req)
		return nil, merr.ErrorInternalServer("grant user functions failed").WithCause(err)
	}

	// 4. autoGrantMenu: 自动追加对应菜单直授
	var autoGrantedMenus []string
	if req.AutoGrantMenu {
		menuCodeSet := make(map[string]struct{})
		for _, fk := range req.FunctionKeys {
			menuCodeSet[fk.MenuCode] = struct{}{}
		}

		var toGrant []string
		for menuCode := range menuCodeSet {
			exists, err := s.sysUserMenuRepo.ExistsByUserUIDAndMenuCode(ctx, req.UserUID, menuCode)
			if err != nil {
				s.helper.Errorw("msg", "check user menu existence failed", "error", err)
				return nil, merr.ErrorInternalServer("check user menu failed").WithCause(err)
			}
			if !exists {
				toGrant = append(toGrant, menuCode)
			}
		}

		if len(toGrant) > 0 {
			grantReq := &bo.GrantUserMenusBo{
				UserUID:   req.UserUID,
				MenuCodes: toGrant,
			}
			_, err := s.sysUserMenuRepo.GrantUserMenus(ctx, grantReq)
			if err != nil {
				s.helper.Errorw("msg", "auto grant user menus failed", "error", err)
				return nil, merr.ErrorInternalServer("auto grant user menus failed").WithCause(err)
			}
			autoGrantedMenus = toGrant
		}
	}

	return &bo.GrantUserFunctionResultBo{
		GrantedCount:     grantedCount,
		AutoGrantedMenus: autoGrantedMenus,
	}, nil
}

// RevokeUserFunctions 撤销用户直授功能
func (s *SysUserFunction) RevokeUserFunctions(ctx context.Context, req *bo.RevokeUserFunctionsBo) (*bo.RevokeUserFunctionResultBo, error) {
	// 1. 校验 userUID 存在
	user, err := s.userRepo.GetUser(ctx, req.UserUID)
	if err != nil {
		s.helper.Errorw("msg", "check user existence failed", "error", err, "userUID", req.UserUID)
		return nil, merr.ErrorInternalServer("check user %d failed", req.UserUID).WithCause(err)
	}
	if user == nil {
		return nil, merr.ErrorNotFound("user %d not found", req.UserUID)
	}

	// 2. 硬删除
	revokedCount, err := s.sysUserFunctionRepo.RevokeUserFunctions(ctx, req)
	if err != nil {
		s.helper.Errorw("msg", "revoke user functions failed", "error", err, "req", req)
		return nil, merr.ErrorInternalServer("revoke user functions failed").WithCause(err)
	}

	// 3. autoRevokeMenu: 检查被撤销功能对应的菜单是否还有其他功能，若无则自动撤销菜单
	var autoRevokedMenus []string
	if req.AutoRevokeMenu {
		menuCodeSet := make(map[string]struct{})
		for _, fk := range req.FunctionKeys {
			menuCodeSet[fk.MenuCode] = struct{}{}
		}

		for menuCode := range menuCodeSet {
			remaining, err := s.sysUserFunctionRepo.CountByUserUIDAndMenuCode(ctx, req.UserUID, menuCode)
			if err != nil {
				s.helper.Errorw("msg", "count remaining functions failed", "error", err)
				return nil, merr.ErrorInternalServer("count remaining functions failed").WithCause(err)
			}
			if remaining == 0 {
				revokeMenuReq := &bo.RevokeUserMenusBo{
					UserUID:   req.UserUID,
					MenuCodes: []string{menuCode},
				}
				_, err := s.sysUserMenuRepo.RevokeUserMenus(ctx, revokeMenuReq)
				if err != nil {
					s.helper.Errorw("msg", "auto revoke user menu failed", "error", err)
					return nil, merr.ErrorInternalServer("auto revoke user menu failed").WithCause(err)
				}
				autoRevokedMenus = append(autoRevokedMenus, menuCode)
			}
		}
	}

	return &bo.RevokeUserFunctionResultBo{
		RevokedCount:     revokedCount,
		AutoRevokedMenus: autoRevokedMenus,
	}, nil
}
