package biz

import (
	"context"

	klog "github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/magicbox/merr"

	"github.com/aide-family/goddess/internal/biz/bo"
	"github.com/aide-family/goddess/internal/biz/repository"
)

// NewSysUserMenu 创建系统用户菜单直授业务逻辑实例
func NewSysUserMenu(
	sysUserMenuRepo repository.SysUserMenu,
	sysMenuRepo repository.SysMenu,
	userRepo repository.User,
	helper *klog.Helper,
) *SysUserMenu {
	return &SysUserMenu{
		sysUserMenuRepo: sysUserMenuRepo,
		sysMenuRepo:     sysMenuRepo,
		userRepo:        userRepo,
		helper:          klog.NewHelper(klog.With(helper.Logger(), "biz", "sys_user_menu")),
	}
}

// SysUserMenu 系统用户菜单直授业务逻辑
type SysUserMenu struct {
	helper          *klog.Helper
	sysUserMenuRepo repository.SysUserMenu
	sysMenuRepo     repository.SysMenu
	userRepo        repository.User
}

// ListUserMenu 查询用户直授菜单列表
func (s *SysUserMenu) ListUserMenu(ctx context.Context, req *bo.ListUserMenuBo) ([]*bo.UserMenuItemBo, error) {
	items, err := s.sysUserMenuRepo.ListUserMenu(ctx, req)
	if err != nil {
		s.helper.Errorw("msg", "list user menu failed", "error", err, "req", req)
		return nil, merr.ErrorInternalServer("list user menu failed").WithCause(err)
	}
	return items, nil
}

// GrantUserMenus 直授用户菜单
func (s *SysUserMenu) GrantUserMenus(ctx context.Context, req *bo.GrantUserMenusBo) (*bo.GrantResultBo, error) {
	// 1. 校验 userUID 在 users 表中存在
	user, err := s.userRepo.GetUser(ctx, req.UserUID)
	if err != nil {
		s.helper.Errorw("msg", "check user existence failed", "error", err, "userUID", req.UserUID)
		return nil, merr.ErrorInternalServer("check user %d failed", req.UserUID).WithCause(err)
	}
	if user == nil {
		return nil, merr.ErrorNotFound("user %d not found", req.UserUID)
	}

	// 2. 校验所有 menuCodes 在 moon_sys_menus 中存在且未删除
	for _, menuCode := range req.MenuCodes {
		exists, err := s.sysMenuRepo.ExistsByMenuCodeNotDeleted(ctx, menuCode)
		if err != nil {
			s.helper.Errorw("msg", "check menu existence failed", "error", err, "menuCode", menuCode)
			return nil, merr.ErrorInternalServer("check menu %s failed", menuCode).WithCause(err)
		}
		if !exists {
			return nil, merr.ErrorNotFound("menu %s not found", menuCode)
		}
	}

	// 3. 自动补全祖先菜单
	finalMenuCodes := make(map[string]struct{})
	for _, mc := range req.MenuCodes {
		finalMenuCodes[mc] = struct{}{}
	}

	var autoAdded []string
	if req.AutoAddAncestors {
		for _, mc := range req.MenuCodes {
			ancestors, err := s.collectAncestors(ctx, mc)
			if err != nil {
				return nil, err
			}
			for _, a := range ancestors {
				if _, exists := finalMenuCodes[a]; !exists {
					finalMenuCodes[a] = struct{}{}
					autoAdded = append(autoAdded, a)
				}
			}
		}
	}

	// 4. 转为 slice 并批量插入
	allMenuCodes := make([]string, 0, len(finalMenuCodes))
	for mc := range finalMenuCodes {
		allMenuCodes = append(allMenuCodes, mc)
	}

	req.MenuCodes = allMenuCodes
	grantedCount, err := s.sysUserMenuRepo.GrantUserMenus(ctx, req)
	if err != nil {
		s.helper.Errorw("msg", "grant user menus failed", "error", err, "req", req)
		return nil, merr.ErrorInternalServer("grant user menus failed").WithCause(err)
	}

	return &bo.GrantResultBo{
		GrantedCount:   grantedCount,
		AutoAddedCodes: autoAdded,
	}, nil
}

// RevokeUserMenus 撤销用户直授菜单
func (s *SysUserMenu) RevokeUserMenus(ctx context.Context, req *bo.RevokeUserMenusBo) (*bo.RevokeResultBo, error) {
	// 1. 校验 userUID 存在
	user, err := s.userRepo.GetUser(ctx, req.UserUID)
	if err != nil {
		s.helper.Errorw("msg", "check user existence failed", "error", err, "userUID", req.UserUID)
		return nil, merr.ErrorInternalServer("check user %d failed", req.UserUID).WithCause(err)
	}
	if user == nil {
		return nil, merr.ErrorNotFound("user %d not found", req.UserUID)
	}

	// 2. 扩展要撤销的菜单列表
	revokeSet := make(map[string]struct{})
	for _, mc := range req.MenuCodes {
		revokeSet[mc] = struct{}{}
	}

	var cascadeRevoked []string
	if req.CascadeDescendants {
		allMenus, err := s.sysMenuRepo.TreeSysMenu(ctx, &bo.TreeSysMenuBo{})
		if err != nil {
			s.helper.Errorw("msg", "list all menus for cascade failed", "error", err)
			return nil, merr.ErrorInternalServer("list all menus failed").WithCause(err)
		}

		childrenMap := make(map[string][]string)
		for _, menu := range allMenus {
			childrenMap[menu.ParentMenuCode] = append(childrenMap[menu.ParentMenuCode], menu.MenuCode)
		}

		for _, mc := range req.MenuCodes {
			descendants := collectDescendants(mc, childrenMap)
			for _, d := range descendants {
				if _, exists := revokeSet[d]; !exists {
					revokeSet[d] = struct{}{}
					cascadeRevoked = append(cascadeRevoked, d)
				}
			}
		}
	}

	// 3. 转为 slice 并执行删除
	allMenuCodes := make([]string, 0, len(revokeSet))
	for mc := range revokeSet {
		allMenuCodes = append(allMenuCodes, mc)
	}

	req.MenuCodes = allMenuCodes
	revokedCount, err := s.sysUserMenuRepo.RevokeUserMenus(ctx, req)
	if err != nil {
		s.helper.Errorw("msg", "revoke user menus failed", "error", err, "req", req)
		return nil, merr.ErrorInternalServer("revoke user menus failed").WithCause(err)
	}

	return &bo.RevokeResultBo{
		RevokedCount:        revokedCount,
		CascadeRevokedCodes: cascadeRevoked,
	}, nil
}

// collectAncestors 递归收集祖先菜单（不含自身），从根到叶返回
func (s *SysUserMenu) collectAncestors(ctx context.Context, menuCode string) ([]string, error) {
	var ancestors []string
	current := menuCode
	for {
		menu, err := s.sysMenuRepo.GetSysMenu(ctx, current)
		if err != nil {
			return nil, merr.ErrorInternalServer("get menu %s failed", current).WithCause(err)
		}
		if menu == nil {
			break
		}
		if menu.ParentMenuCode == "" {
			break
		}
		ancestors = append(ancestors, menu.ParentMenuCode)
		current = menu.ParentMenuCode
	}
	for i, j := 0, len(ancestors)-1; i < j; i, j = i+1, j-1 {
		ancestors[i], ancestors[j] = ancestors[j], ancestors[i]
	}
	return ancestors, nil
}

// collectDescendants BFS 收集所有子孙菜单 code（不含自身），包级共享函数
func collectDescendants(parentCode string, childrenMap map[string][]string) []string {
	var result []string
	queue := childrenMap[parentCode]
	for len(queue) > 0 {
		code := queue[0]
		queue = queue[1:]
		result = append(result, code)
		queue = append(queue, childrenMap[code]...)
	}
	return result
}
