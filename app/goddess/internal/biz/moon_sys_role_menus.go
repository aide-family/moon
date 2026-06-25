package biz

import (
	"context"

	klog "github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/magicbox/merr"

	"github.com/aide-family/goddess/internal/biz/bo"
	"github.com/aide-family/goddess/internal/biz/repository"
)

// NewSysRoleMenu 创建系统角色菜单业务逻辑实例
func NewSysRoleMenu(
	sysRoleMenuRepo repository.SysRoleMenu,
	sysMenuRepo repository.SysMenu,
	sysRoleRepo repository.SysRole,
	helper *klog.Helper,
) *SysRoleMenu {
	return &SysRoleMenu{
		sysRoleMenuRepo: sysRoleMenuRepo,
		sysMenuRepo:     sysMenuRepo,
		sysRoleRepo:     sysRoleRepo,
		helper:          klog.NewHelper(klog.With(helper.Logger(), "biz", "sys_role_menu")),
	}
}

// SysRoleMenu 系统角色菜单业务逻辑
type SysRoleMenu struct {
	helper          *klog.Helper
	sysRoleMenuRepo repository.SysRoleMenu
	sysMenuRepo     repository.SysMenu
	sysRoleRepo     repository.SysRole
}

// ListRoleMenu 查询角色菜单列表
func (s *SysRoleMenu) ListRoleMenu(ctx context.Context, req *bo.ListRoleMenuBo) ([]*bo.RoleMenuItemBo, error) {
	items, err := s.sysRoleMenuRepo.ListRoleMenu(ctx, req)
	if err != nil {
		s.helper.Errorw("msg", "list role menu failed", "error", err, "req", req)
		return nil, merr.ErrorInternalServer("list role menu failed").WithCause(err)
	}
	return items, nil
}

// GrantRoleMenus 授予角色菜单
func (s *SysRoleMenu) GrantRoleMenus(ctx context.Context, req *bo.GrantRoleMenusBo) (*bo.GrantResultBo, error) {
	// 1. 校验 roleCode 在 moon_sys_roles 中存在且启用
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
	grantedCount, err := s.sysRoleMenuRepo.GrantRoleMenus(ctx, req)
	if err != nil {
		s.helper.Errorw("msg", "grant role menus failed", "error", err, "req", req)
		return nil, merr.ErrorInternalServer("grant role menus failed").WithCause(err)
	}

	return &bo.GrantResultBo{
		GrantedCount:   grantedCount,
		AutoAddedCodes: autoAdded,
	}, nil
}

// RevokeRoleMenus 撤销角色菜单
func (s *SysRoleMenu) RevokeRoleMenus(ctx context.Context, req *bo.RevokeRoleMenusBo) (*bo.RevokeResultBo, error) {
	// 1. 校验 roleCode 存在
	role, err := s.sysRoleRepo.GetSysRole(ctx, req.RoleCode)
	if err != nil {
		s.helper.Errorw("msg", "check role existence failed", "error", err, "roleCode", req.RoleCode)
		return nil, merr.ErrorInternalServer("check role %s failed", req.RoleCode).WithCause(err)
	}
	if role == nil {
		return nil, merr.ErrorNotFound("role %s not found", req.RoleCode)
	}

	// 2. 扩展要撤销的菜单列表
	revokeSet := make(map[string]struct{})
	for _, mc := range req.MenuCodes {
		revokeSet[mc] = struct{}{}
	}

	var cascadeRevoked []string
	if req.CascadeDescendants {
		// 获取全量菜单，在内存中递归查找子孙
		allMenus, err := s.sysMenuRepo.TreeSysMenu(ctx, &bo.TreeSysMenuBo{})
		if err != nil {
			s.helper.Errorw("msg", "list all menus for cascade failed", "error", err)
			return nil, merr.ErrorInternalServer("list all menus failed").WithCause(err)
		}

		// 构建 parent → children 映射
		childrenMap := make(map[string][]string)
		for _, menu := range allMenus {
			childrenMap[menu.ParentMenuCode] = append(childrenMap[menu.ParentMenuCode], menu.MenuCode)
		}

		for _, mc := range req.MenuCodes {
			descendants := s.collectDescendants(mc, childrenMap)
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
	revokedCount, err := s.sysRoleMenuRepo.RevokeRoleMenus(ctx, req)
	if err != nil {
		s.helper.Errorw("msg", "revoke role menus failed", "error", err, "req", req)
		return nil, merr.ErrorInternalServer("revoke role menus failed").WithCause(err)
	}

	return &bo.RevokeResultBo{
		RevokedCount:        revokedCount,
		CascadeRevokedCodes: cascadeRevoked,
	}, nil
}

// collectAncestors 递归收集祖先菜单（不含自身），从顶向下返回
func (s *SysRoleMenu) collectAncestors(ctx context.Context, menuCode string) ([]string, error) {
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
	// 反转顺序（从根到叶），使前端可以按顺序展开
	for i, j := 0, len(ancestors)-1; i < j; i, j = i+1, j-1 {
		ancestors[i], ancestors[j] = ancestors[j], ancestors[i]
	}
	return ancestors, nil
}

// collectDescendants 递归收集所有子孙菜单 code（不含自身）
func (s *SysRoleMenu) collectDescendants(parentCode string, childrenMap map[string][]string) []string {
	var result []string
	queue := childrenMap[parentCode]
	for len(queue) > 0 {
		code := queue[0]
		queue = queue[1:]
		result = append(result, code)
		// 将子菜单的子菜单加入队列
		queue = append(queue, childrenMap[code]...)
	}
	return result
}
