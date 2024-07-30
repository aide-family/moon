package repository

import (
	"context"

	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/pkg/palace/model"
	"github.com/aide-family/moon/pkg/vobj"
)

// Menu 系统菜单接口
type Menu interface {
	// Create 创建系统菜单
	Create(context.Context, *bo.CreateMenuParams) (*model.SysMenu, error)
	// BatchCreate 批量创建系统菜单
	BatchCreate(context.Context, []*bo.CreateMenuParams) error

	// UpdateById 更新系统菜单
	UpdateByID(context.Context, *bo.UpdateMenuParams) error
	// DeleteById 删除系统菜单
	DeleteByID(context.Context, uint32) error
	// GetByID 根据id获取系统菜单
	GetByID(context.Context, uint32) (*model.SysMenu, error)
	// FindByPage 分页查询系统菜单
	FindByPage(context.Context, *bo.QueryMenuListParams) ([]*model.SysMenu, error)
	// ListAll 获取所有系统菜单
	ListAll(context.Context) ([]*model.SysMenu, error)
	// UpdateStatusByIds 更新系统菜单状态
	UpdateStatusByIds(context.Context, vobj.Status, ...uint32) error
	// UpdateTypeByIds 更新系统菜单类型
	UpdateTypeByIds(context.Context, vobj.MenuType, ...uint32) error
}
