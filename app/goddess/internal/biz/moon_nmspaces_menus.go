package biz

import (
	"context"

	klog "github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/magicbox/merr"
	"github.com/bwmarrin/snowflake"

	"github.com/aide-family/goddess/internal/biz/bo"
	"github.com/aide-family/goddess/internal/biz/repository"
)

// NewNmspacesMenu 创建租户菜单业务逻辑实例
func NewNmspacesMenu(nmspacesMenuRepo repository.NmspacesMenu, helper *klog.Helper) *NmspacesMenuBiz {
	return &NmspacesMenuBiz{
		nmspacesMenuRepo: nmspacesMenuRepo,
		helper:           klog.NewHelper(klog.With(helper.Logger(), "biz", "nmspaces_menu")),
	}
}

// NmspacesMenuBiz 租户菜单业务逻辑
type NmspacesMenuBiz struct {
	helper           *klog.Helper
	nmspacesMenuRepo repository.NmspacesMenu
}

// ListNmspacesMenu 分页查询租户菜单列表
func (s *NmspacesMenuBiz) ListNmspacesMenu(ctx context.Context, req *bo.ListNmspacesMenuBo) ([]*bo.NmspacesMenuItemBo, int64, error) {
	items, total, err := s.nmspacesMenuRepo.ListNmspacesMenu(ctx, req)
	if err != nil {
		s.helper.Errorw("msg", "list nmspaces menu failed", "error", err, "req", req)
		return nil, 0, merr.ErrorInternalServer("list nmspaces menu failed").WithCause(err)
	}
	return items, total, nil
}

// GetNmspacesMenu 查询租户菜单详情
func (s *NmspacesMenuBiz) GetNmspacesMenu(ctx context.Context, namespaceUID snowflake.ID, menuCode string) (*bo.NmspacesMenuItemBo, error) {
	result, err := s.nmspacesMenuRepo.GetNmspacesMenu(ctx, namespaceUID, menuCode)
	if err != nil {
		s.helper.Errorw("msg", "get nmspaces menu failed", "error", err, "namespaceUID", namespaceUID, "menuCode", menuCode)
		return nil, merr.ErrorInternalServer("get nmspaces menu failed").WithCause(err)
	}
	if result == nil {
		return nil, merr.ErrorNotFound("menu %s not found in namespace %d", menuCode, namespaceUID)
	}
	return result, nil
}

// CreateNmspacesMenu 新增租户菜单
func (s *NmspacesMenuBiz) CreateNmspacesMenu(ctx context.Context, req *bo.CreateNmspacesMenuBo) error {
	// 校验 menuCode 不存在（含软删除行）
	exists, err := s.nmspacesMenuRepo.ExistsByNamespaceAndMenuCode(ctx, req.NamespaceUID, req.MenuCode)
	if err != nil {
		s.helper.Errorw("msg", "check menuCode existence failed", "error", err, "menuCode", req.MenuCode)
		return merr.ErrorInternalServer("check menuCode failed").WithCause(err)
	}
	if exists {
		return merr.ErrorBadRequest("menuCode %s already exists in this namespace", req.MenuCode)
	}

	// 校验同一父菜单下 menuName 唯一（未删除行）
	nameExists, err := s.nmspacesMenuRepo.ExistsByNamespaceAndParentAndName(ctx, req.NamespaceUID, req.ParentMenuCode, req.MenuName)
	if err != nil {
		s.helper.Errorw("msg", "check menuName uniqueness failed", "error", err, "menuName", req.MenuName)
		return merr.ErrorInternalServer("check menuName failed").WithCause(err)
	}
	if nameExists {
		return merr.ErrorBadRequest("menuName %s already exists under the same parent", req.MenuName)
	}

	if err := s.nmspacesMenuRepo.CreateNmspacesMenu(ctx, req); err != nil {
		s.helper.Errorw("msg", "create nmspaces menu failed", "error", err, "req", req)
		return merr.ErrorInternalServer("create nmspaces menu failed").WithCause(err)
	}
	return nil
}

// UpdateNmspacesMenu 编辑租户菜单
func (s *NmspacesMenuBiz) UpdateNmspacesMenu(ctx context.Context, req *bo.UpdateNmspacesMenuBo) error {
	// 校验菜单存在且未删除
	exists, err := s.nmspacesMenuRepo.ExistsByNamespaceAndMenuCodeNotDeleted(ctx, req.NamespaceUID, req.MenuCode)
	if err != nil {
		s.helper.Errorw("msg", "check menu existence failed", "error", err, "menuCode", req.MenuCode)
		return merr.ErrorInternalServer("check menu existence failed").WithCause(err)
	}
	if !exists {
		return merr.ErrorNotFound("menu %s not found in namespace %d", req.MenuCode, req.NamespaceUID)
	}

	// 若修改父菜单，校验新父菜单存在
	if req.ParentMenuCode != nil && *req.ParentMenuCode != "" {
		parentExists, err := s.nmspacesMenuRepo.ExistsByNamespaceAndParentMenuCode(ctx, req.NamespaceUID, *req.ParentMenuCode)
		if err != nil {
			s.helper.Errorw("msg", "check parent menu failed", "error", err, "parentMenuCode", *req.ParentMenuCode)
			return merr.ErrorInternalServer("check parent menu failed").WithCause(err)
		}
		if !parentExists {
			return merr.ErrorBadRequest("parentMenuCode %s not found in this namespace", *req.ParentMenuCode)
		}
	}

	// 若修改 menuName，校验同一父菜单下唯一
	if req.MenuName != nil {
		parentCode := ""
		if req.ParentMenuCode != nil {
			parentCode = *req.ParentMenuCode
		}
		nameExists, err := s.nmspacesMenuRepo.ExistsByNamespaceAndParentAndName(ctx, req.NamespaceUID, parentCode, *req.MenuName)
		if err != nil {
			s.helper.Errorw("msg", "check menuName uniqueness failed", "error", err, "menuName", *req.MenuName)
			return merr.ErrorInternalServer("check menuName failed").WithCause(err)
		}
		if nameExists {
			return merr.ErrorBadRequest("menuName %s already exists under the same parent", *req.MenuName)
		}
	}

	if err := s.nmspacesMenuRepo.UpdateNmspacesMenu(ctx, req); err != nil {
		s.helper.Errorw("msg", "update nmspaces menu failed", "error", err, "req", req)
		return merr.ErrorInternalServer("update nmspaces menu failed").WithCause(err)
	}
	return nil
}

// DeleteNmspacesMenu 删除租户菜单（软删除）
func (s *NmspacesMenuBiz) DeleteNmspacesMenu(ctx context.Context, namespaceUID snowflake.ID, menuCode string) error {
	// 校验菜单存在且未删除
	exists, err := s.nmspacesMenuRepo.ExistsByNamespaceAndMenuCodeNotDeleted(ctx, namespaceUID, menuCode)
	if err != nil {
		s.helper.Errorw("msg", "check menu existence failed", "error", err, "menuCode", menuCode)
		return merr.ErrorInternalServer("check menu existence failed").WithCause(err)
	}
	if !exists {
		return merr.ErrorNotFound("menu %s not found in namespace %d", menuCode, namespaceUID)
	}

	// 级联软删除子菜单：获取当前租户全量菜单，收集所有子孙节点，从叶子向根逐层删除
	allMenus, _, err := s.nmspacesMenuRepo.ListNmspacesMenu(ctx, &bo.ListNmspacesMenuBo{
		PageRequestBo: &bo.PageRequestBo{Page: 0, PageSize: 0},
		NamespaceUID:  namespaceUID,
	})
	if err != nil {
		s.helper.Errorw("msg", "list all menus for cascade delete failed", "error", err)
		return merr.ErrorInternalServer("list all menus for cascade delete failed").WithCause(err)
	}

	// 构建 parentMenuCode → childrenMenuCodes 映射
	childrenMap := make(map[string][]string)
	for _, menu := range allMenus {
		childrenMap[menu.ParentMenuCode] = append(childrenMap[menu.ParentMenuCode], menu.MenuCode)
	}

	// BFS 收集所有子孙菜单，反序从叶子节点开始删除
	descendants := collectDescendants(menuCode, childrenMap)
	for i := len(descendants) - 1; i >= 0; i-- {
		if err := s.nmspacesMenuRepo.DeleteNmspacesMenu(ctx, namespaceUID, descendants[i]); err != nil {
			s.helper.Errorw("msg", "cascade delete child menu failed", "error", err, "menuCode", descendants[i])
			return merr.ErrorInternalServer("cascade delete child menu %s failed", descendants[i]).WithCause(err)
		}
	}

	// 删除目标菜单本身
	if err := s.nmspacesMenuRepo.DeleteNmspacesMenu(ctx, namespaceUID, menuCode); err != nil {
		s.helper.Errorw("msg", "delete nmspaces menu failed", "error", err, "menuCode", menuCode)
		return merr.ErrorInternalServer("delete nmspaces menu failed").WithCause(err)
	}
	return nil
}
