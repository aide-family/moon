package repository

import (
	"context"

	query "github.com/aide-cloud/gorm-normalize"
	"prometheus-manager/app/prom_server/internal/biz/dobo"
)

type (
	PromDictRepo interface {
		// CreateDict 创建字典
		CreateDict(ctx context.Context, dict *dobo.DictDO) (*dobo.DictDO, error)
		// UpdateDictById 通过id更新字典
		UpdateDictById(ctx context.Context, id uint, dict *dobo.DictDO) (*dobo.DictDO, error)
		// BatchUpdateDictStatusByIds 通过id批量更新字典状态
		BatchUpdateDictStatusByIds(ctx context.Context, status int32, ids []uint) error
		// DeleteDictByIds 通过id删除字典
		DeleteDictByIds(ctx context.Context, id ...uint) error
		// GetDictById 通过id获取字典详情
		GetDictById(ctx context.Context, id uint) (*dobo.DictDO, error)
		// ListDict 获取字典列表
		ListDict(ctx context.Context, pgInfo query.Pagination, scopes ...query.ScopeMethod) ([]*dobo.DictDO, error)
	}
)
