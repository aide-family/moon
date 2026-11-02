package repository

import (
	"context"
	"time"

	"github.com/aide-family/moon/app/alioth/pkg/bo"
)

// StatusPageRepo 管理状态页配置，并为公开链接提供聚合展示数据。
type StatusPageRepo interface {
	CreateStatusPage(ctx context.Context, in bo.StatusPage) (*bo.StatusPage, error)
	UpdateStatusPage(ctx context.Context, in bo.StatusPage) (*bo.StatusPage, error)
	UpdateStatusPageStatus(ctx context.Context, namespaceUID, uid, status string) (*bo.StatusPage, error)
	RegenerateStatusPageToken(ctx context.Context, namespaceUID, uid string) (*bo.StatusPage, error)
	DeleteStatusPage(ctx context.Context, namespaceUID, uid string) error
	GetStatusPage(ctx context.Context, namespaceUID, uid string) (*bo.StatusPage, error)
	ListStatusPages(ctx context.Context, namespaceUID string, filter bo.StatusPageListFilter, page, pageSize int) ([]bo.StatusPage, int, error)
	// PublicStatusPage 按 token 返回公开展示数据；token 无效、页面停用或租户停用时返回 NotFound。
	PublicStatusPage(ctx context.Context, token string, now time.Time) (*bo.PublicStatusPage, error)
}
