package biz

import (
	"context"

	klog "github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/magicbox/merr"

	"github.com/aide-family/goddess/internal/biz/bo"
	"github.com/aide-family/goddess/internal/biz/repository"
)

// NewSysMenu 创建系统菜单业务逻辑实例
func NewSysMenu(sysMenuRepo repository.SysMenu, helper *klog.Helper) *SysMenu {
	return &SysMenu{
		sysMenuRepo: sysMenuRepo,
		helper:      klog.NewHelper(klog.With(helper.Logger(), "biz", "sys_menu")),
	}
}

// SysMenu 系统菜单业务逻辑
type SysMenu struct {
	helper      *klog.Helper
	sysMenuRepo repository.SysMenu
}

// ListSysMenu 分页查询系统菜单列表
func (s *SysMenu) ListSysMenu(ctx context.Context, req *bo.ListSysMenuBo) ([]*bo.SysMenuItemBo, int64, error) {
	items, total, err := s.sysMenuRepo.ListSysMenu(ctx, req)
	if err != nil {
		s.helper.Errorw("msg", "list sys menu failed", "error", err, "req", req)
		return nil, 0, merr.ErrorInternalServer("list sys menu failed").WithCause(err)
	}
	return items, total, nil
}

// TreeSysMenu 查询系统菜单树
func (s *SysMenu) TreeSysMenu(ctx context.Context, req *bo.TreeSysMenuBo) ([]*bo.SysMenuItemBo, error) {
	items, err := s.sysMenuRepo.TreeSysMenu(ctx, req)
	if err != nil {
		s.helper.Errorw("msg", "tree sys menu failed", "error", err)
		return nil, merr.ErrorInternalServer("tree sys menu failed").WithCause(err)
	}
	return items, nil
}

// GetSysMenu 根据 menuCode 查询菜单详情
func (s *SysMenu) GetSysMenu(ctx context.Context, menuCode string) (*bo.SysMenuItemBo, error) {
	result, err := s.sysMenuRepo.GetSysMenu(ctx, menuCode)
	if err != nil {
		s.helper.Errorw("msg", "get sys menu failed", "error", err, "menuCode", menuCode)
		return nil, merr.ErrorInternalServer("get sys menu failed").WithCause(err)
	}
	if result == nil {
		return nil, merr.ErrorNotFound("menu %s not found", menuCode)
	}
	return result, nil
}

// CountChildren 统计子菜单数量
func (s *SysMenu) CountChildren(ctx context.Context, parentMenuCode string) (int64, error) {
	count, err := s.sysMenuRepo.CountChildren(ctx, parentMenuCode)
	if err != nil {
		s.helper.Errorw("msg", "count children failed", "error", err, "parentMenuCode", parentMenuCode)
		return 0, merr.ErrorInternalServer("count children failed").WithCause(err)
	}
	return count, nil
}

// CreateSysMenu 新增系统菜单
func (s *SysMenu) CreateSysMenu(ctx context.Context, req *bo.CreateSysMenuBo) error {
	// 校验 menuCode 全局唯一（含软删除行）
	exists, err := s.sysMenuRepo.ExistsByMenuCode(ctx, req.MenuCode)
	if err != nil {
		s.helper.Errorw("msg", "check menuCode existence failed", "error", err, "menuCode", req.MenuCode)
		return merr.ErrorInternalServer("check menuCode failed").WithCause(err)
	}
	if exists {
		return merr.ErrorBadRequest("menuCode %s already exists", req.MenuCode)
	}

	// 若 parentMenuCode 非空，校验父菜单存在且未删除
	if req.ParentMenuCode != "" {
		parentExists, err := s.sysMenuRepo.ExistsByParentMenuCode(ctx, req.ParentMenuCode)
		if err != nil {
			s.helper.Errorw("msg", "check parent menu failed", "error", err, "parentMenuCode", req.ParentMenuCode)
			return merr.ErrorInternalServer("check parent menu failed").WithCause(err)
		}
		if !parentExists {
			return merr.ErrorNotFound("parent menu %s not found", req.ParentMenuCode)
		}
	}

	// 校验同一 parentMenuCode 下 menuName 唯一
	nameExists, err := s.sysMenuRepo.ExistsByParentAndName(ctx, req.ParentMenuCode, req.MenuName)
	if err != nil {
		s.helper.Errorw("msg", "check menuName uniqueness failed", "error", err, "menuName", req.MenuName)
		return merr.ErrorInternalServer("check menuName failed").WithCause(err)
	}
	if nameExists {
		return merr.ErrorBadRequest("menuName %s already exists under parent", req.MenuName)
	}

	// 插入菜单
	if err := s.sysMenuRepo.CreateSysMenu(ctx, req); err != nil {
		s.helper.Errorw("msg", "create sys menu failed", "error", err, "req", req)
		return merr.ErrorInternalServer("create sys menu failed").WithCause(err)
	}
	return nil
}

// UpdateSysMenu 编辑系统菜单
func (s *SysMenu) UpdateSysMenu(ctx context.Context, req *bo.UpdateSysMenuBo) error {
	// 校验菜单存在且未删除
	exists, err := s.sysMenuRepo.ExistsByMenuCodeNotDeleted(ctx, req.MenuCode)
	if err != nil {
		s.helper.Errorw("msg", "check menu existence failed", "error", err, "menuCode", req.MenuCode)
		return merr.ErrorInternalServer("check menu existence failed").WithCause(err)
	}
	if !exists {
		return merr.ErrorNotFound("menu %s not found", req.MenuCode)
	}

	// 若修改 parentMenuCode，校验新父菜单存在且不能自引用
	if req.ParentMenuCode != nil {
		if *req.ParentMenuCode == req.MenuCode {
			return merr.ErrorBadRequest("menu cannot be its own parent")
		}
		if *req.ParentMenuCode != "" {
			parentExists, err := s.sysMenuRepo.ExistsByParentMenuCode(ctx, *req.ParentMenuCode)
			if err != nil {
				s.helper.Errorw("msg", "check parent menu failed", "error", err, "parentMenuCode", *req.ParentMenuCode)
				return merr.ErrorInternalServer("check parent menu failed").WithCause(err)
			}
			if !parentExists {
				return merr.ErrorNotFound("parent menu %s not found", *req.ParentMenuCode)
			}
		}
	}

	// 执行更新
	if err := s.sysMenuRepo.UpdateSysMenu(ctx, req); err != nil {
		s.helper.Errorw("msg", "update sys menu failed", "error", err, "req", req)
		return merr.ErrorInternalServer("update sys menu failed").WithCause(err)
	}
	return nil
}

// DeleteSysMenu 删除系统菜单（软删除）
func (s *SysMenu) DeleteSysMenu(ctx context.Context, menuCode string) error {
	// 校验菜单存在且未删除
	exists, err := s.sysMenuRepo.ExistsByMenuCodeNotDeleted(ctx, menuCode)
	if err != nil {
		s.helper.Errorw("msg", "check menu existence failed", "error", err, "menuCode", menuCode)
		return merr.ErrorInternalServer("check menu existence failed").WithCause(err)
	}
	if !exists {
		return merr.ErrorNotFound("menu %s not found", menuCode)
	}

	// 检查是否有子菜单
	childrenCount, err := s.sysMenuRepo.CountChildren(ctx, menuCode)
	if err != nil {
		s.helper.Errorw("msg", "count children failed", "error", err, "menuCode", menuCode)
		return merr.ErrorInternalServer("count children failed").WithCause(err)
	}
	if childrenCount > 0 {
		return merr.ErrorBadRequest("menu %s has %d children, delete them first", menuCode, childrenCount)
	}

	// 执行软删除
	if err := s.sysMenuRepo.DeleteSysMenu(ctx, menuCode); err != nil {
		s.helper.Errorw("msg", "delete sys menu failed", "error", err, "menuCode", menuCode)
		return merr.ErrorInternalServer("delete sys menu failed").WithCause(err)
	}
	return nil
}
