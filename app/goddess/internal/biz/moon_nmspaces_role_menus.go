package biz

import (
	"context"

	klog "github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/magicbox/merr"
	"github.com/bwmarrin/snowflake"

	"github.com/aide-family/goddess/internal/biz/bo"
	"github.com/aide-family/goddess/internal/biz/repository"
)

// NewNmspacesRoleMenu 创建租户角色菜单业务逻辑实例
func NewNmspacesRoleMenu(
	nmspacesRoleMenuRepo repository.NmspacesRoleMenu,
	nmspacesMenuRepo repository.NmspacesMenu,
	nmspacesRoleRepo repository.NmspacesRole,
	helper *klog.Helper,
) *NmspacesRoleMenuBiz {
	return &NmspacesRoleMenuBiz{
		nmspacesRoleMenuRepo: nmspacesRoleMenuRepo,
		nmspacesMenuRepo:     nmspacesMenuRepo,
		nmspacesRoleRepo:     nmspacesRoleRepo,
		helper:               klog.NewHelper(klog.With(helper.Logger(), "biz", "nmspaces_role_menu")),
	}
}

// NmspacesRoleMenuBiz 租户角色菜单业务逻辑
type NmspacesRoleMenuBiz struct {
	helper               *klog.Helper
	nmspacesRoleMenuRepo repository.NmspacesRoleMenu
	nmspacesMenuRepo     repository.NmspacesMenu
	nmspacesRoleRepo     repository.NmspacesRole
}

// ListNmspacesRoleMenu 查询角色菜单列表
func (s *NmspacesRoleMenuBiz) ListNmspacesRoleMenu(ctx context.Context, req *bo.ListNmspacesRoleMenuBo) ([]*bo.NmspacesRoleMenuItemBo, error) {
	items, err := s.nmspacesRoleMenuRepo.ListNmspacesRoleMenu(ctx, req)
	if err != nil {
		s.helper.Errorw("msg", "list nmspaces role menu failed", "error", err, "req", req)
		return nil, merr.ErrorInternalServer("list nmspaces role menu failed").WithCause(err)
	}
	return items, nil
}

// GrantNmspacesRoleMenus 授予角色菜单
func (s *NmspacesRoleMenuBiz) GrantNmspacesRoleMenus(ctx context.Context, req *bo.GrantNmspacesRoleMenusBo) (*bo.GrantResultBo, error) {
	// 1. 校验 roleCode 在当前租户存在且启用
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

	// 2. 校验所有 menuCodes 在当前租户存在且未删除
	for _, menuCode := range req.MenuCodes {
		exists, err := s.nmspacesMenuRepo.ExistsByNamespaceAndMenuCodeNotDeleted(ctx, req.NamespaceUID, menuCode)
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
			ancestors, err := s.collectNmspacesAncestors(ctx, req.NamespaceUID, mc)
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
	grantedCount, err := s.nmspacesRoleMenuRepo.GrantNmspacesRoleMenus(ctx, req)
	if err != nil {
		s.helper.Errorw("msg", "grant nmspaces role menus failed", "error", err, "req", req)
		return nil, merr.ErrorInternalServer("grant nmspaces role menus failed").WithCause(err)
	}

	return &bo.GrantResultBo{
		GrantedCount:   grantedCount,
		AutoAddedCodes: autoAdded,
	}, nil
}

// RevokeNmspacesRoleMenus 撤销角色菜单
func (s *NmspacesRoleMenuBiz) RevokeNmspacesRoleMenus(ctx context.Context, req *bo.RevokeNmspacesRoleMenusBo) (*bo.RevokeResultBo, error) {
	// 1. 校验 roleCode 存在
	role, err := s.nmspacesRoleRepo.GetNmspacesRole(ctx, req.NamespaceUID, req.RoleCode)
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
		// 获取当前租户全量菜单
		allMenus, _, err := s.nmspacesMenuRepo.ListNmspacesMenu(ctx, &bo.ListNmspacesMenuBo{
			PageRequestBo: &bo.PageRequestBo{Page: 0, PageSize: 0},
			NamespaceUID:  req.NamespaceUID,
		})
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
	revokedCount, err := s.nmspacesRoleMenuRepo.RevokeNmspacesRoleMenus(ctx, req)
	if err != nil {
		s.helper.Errorw("msg", "revoke nmspaces role menus failed", "error", err, "req", req)
		return nil, merr.ErrorInternalServer("revoke nmspaces role menus failed").WithCause(err)
	}

	return &bo.RevokeResultBo{
		RevokedCount:        revokedCount,
		CascadeRevokedCodes: cascadeRevoked,
	}, nil
}

// collectNmspacesAncestors 递归收集祖先菜单（不含自身），从根到叶返回
func (s *NmspacesRoleMenuBiz) collectNmspacesAncestors(ctx context.Context, namespaceUID snowflake.ID, menuCode string) ([]string, error) {
	var ancestors []string
	current := menuCode
	for {
		menu, err := s.nmspacesMenuRepo.GetNmspacesMenu(ctx, namespaceUID, current)
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
