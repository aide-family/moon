package repository

import (
	"context"

	"github.com/aide-family/moon/app/prom_server/internal/biz/bo"
	"github.com/aide-family/moon/app/prom_server/internal/biz/do"
	"github.com/aide-family/moon/app/prom_server/internal/biz/vobj"
)

type SysDictRepo interface {
	// CreateDict 创建字典
	CreateDict(ctx context.Context, dict *bo.CreateSysDictBo) (*do.SysDict, error)
	// UpdateDictById 通过id更新字典
	UpdateDictById(ctx context.Context, id uint32, dict *bo.UpdateSysDictBo) (*do.SysDict, error)
	// BatchUpdateDictStatusByIds 通过id批量更新字典状态
	BatchUpdateDictStatusByIds(ctx context.Context, status vobj.Status, ids []uint32) error
	// DeleteDictByIds 通过id删除字典
	DeleteDictByIds(ctx context.Context, id ...uint32) error
	// GetDictById 通过id获取字典详情
	GetDictById(ctx context.Context, id uint32) (*do.SysDict, error)
	GetDictByIds(ctx context.Context, ids ...uint32) ([]*do.SysDict, error)
	// ListDict 获取字典列表
	ListDict(ctx context.Context, params *bo.ListSysDictBo) ([]*do.SysDict, error)
	SelectDict(ctx context.Context, params *bo.SelectSysDictBo) ([]*do.SysDict, error)
}
