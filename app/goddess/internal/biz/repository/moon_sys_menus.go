package repository

import (
	"context"

	"github.com/aide-family/goddess/internal/biz/bo"
)

// SysMenu 系统菜单仓储接口
type SysMenu interface {
	// ListSysMenu 分页查询系统菜单列表
	ListSysMenu(ctx context.Context, req *bo.ListSysMenuBo) ([]*bo.SysMenuItemBo, int64, error)
	// TreeSysMenu 查询全量菜单（内存构建树）
	TreeSysMenu(ctx context.Context, req *bo.TreeSysMenuBo) ([]*bo.SysMenuItemBo, error)
	// GetSysMenu 根据 menuCode 查询菜单详情
	GetSysMenu(ctx context.Context, menuCode string) (*bo.SysMenuItemBo, error)
	// CountChildren 统计子菜单数量
	CountChildren(ctx context.Context, parentMenuCode string) (int64, error)
	// ExistsByMenuCode 检查 menuCode 是否已存在（含软删除行）
	ExistsByMenuCode(ctx context.Context, menuCode string) (bool, error)
	// ExistsByMenuCodeNotDeleted 检查 menuCode 是否在未删除行中存在
	ExistsByMenuCodeNotDeleted(ctx context.Context, menuCode string) (bool, error)
	// ExistsByParentAndName 检查同一 parentMenuCode 下 menuName 是否已存在（未删除行）
	ExistsByParentAndName(ctx context.Context, parentMenuCode, menuName string) (bool, error)
	// ExistsByParentMenuCode 检查 parentMenuCode 是否存在且未删除
	ExistsByParentMenuCode(ctx context.Context, parentMenuCode string) (bool, error)
	// CreateSysMenu 新增系统菜单
	CreateSysMenu(ctx context.Context, req *bo.CreateSysMenuBo) error
	// UpdateSysMenu 编辑系统菜单
	UpdateSysMenu(ctx context.Context, req *bo.UpdateSysMenuBo) error
	// DeleteSysMenu 删除系统菜单（软删除）
	DeleteSysMenu(ctx context.Context, menuCode string) error
}
