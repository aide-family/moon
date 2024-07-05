package repository

import (
	"context"

	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/pkg/palace/model"
	"github.com/aide-family/moon/pkg/vobj"
)

type Menu interface {
	// Create 创建系统菜单
	Create(ctx context.Context, menu *bo.CreateMenuParams) (*model.SysMenu, error)
	// BatchCreate 批量创建系统菜单
	BatchCreate(ctx context.Context, menus []*bo.CreateMenuParams) error

	// UpdateById 更新系统菜单
	UpdateById(ctx context.Context, user *bo.UpdateMenuParams) error
	// DeleteById 删除系统菜单
	DeleteById(ctx context.Context, id uint32) error
	// GetById 根据id获取系统菜单
	GetById(ctx context.Context, id uint32) (*model.SysMenu, error)
	// FindByPage 分页查询系统菜单
	FindByPage(ctx context.Context, params *bo.QueryMenuListParams) ([]*model.SysMenu, error)
	// ListAll 获取所有系统菜单
	ListAll(ctx context.Context) ([]*model.SysMenu, error)
	// UpdateStatusByIds 更新系统菜单状态
	UpdateStatusByIds(ctx context.Context, status vobj.Status, ids ...uint32) error
	// UpdateTypeByIds 更新系统菜单类型
	UpdateTypeByIds(ctx context.Context, menuType vobj.MenuType, ids ...uint32) error
}
