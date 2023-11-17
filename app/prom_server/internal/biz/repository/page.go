package repository

import (
	"context"

	query "github.com/aide-cloud/gorm-normalize"
	"prometheus-manager/app/prom_server/internal/biz/dobo"
)

type (
	PageRepo interface {
		// CreatePage 创建页面
		CreatePage(ctx context.Context, pageDo *dobo.AlarmPageDO) (*dobo.AlarmPageDO, error)
		// UpdatePageById 通过id更新页面
		UpdatePageById(ctx context.Context, id uint, pageDo *dobo.AlarmPageDO) (*dobo.AlarmPageDO, error)
		// BatchUpdatePageStatusByIds 通过id批量更新页面状态
		BatchUpdatePageStatusByIds(ctx context.Context, status int32, ids []uint) error
		// DeletePageByIds 通过id删除页面
		DeletePageByIds(ctx context.Context, id ...uint) error
		// GetPageById 通过id获取页面详情
		GetPageById(ctx context.Context, id uint) (*dobo.AlarmPageDO, error)
		// ListPage 获取页面列表
		ListPage(ctx context.Context, pgInfo query.Pagination, scopes ...query.ScopeMethod) ([]*dobo.AlarmPageDO, error)
	}
)
