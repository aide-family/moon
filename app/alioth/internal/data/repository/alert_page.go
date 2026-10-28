package repository

import (
	"context"

	"github.com/aide-family/moon/app/alioth/pkg/bo"
)

// AlertPageRepo 读写告警页面与个人 Tab 订阅。
type AlertPageRepo interface {
	CreateAlertPage(ctx context.Context, in bo.AlertPage) (*bo.AlertPage, error)
	UpdateAlertPage(ctx context.Context, in bo.AlertPage) (*bo.AlertPage, error)
	DeleteAlertPage(ctx context.Context, namespaceUID, uid string) error
	GetAlertPage(ctx context.Context, namespaceUID, uid string) (*bo.AlertPage, error)
	ListAlertPages(ctx context.Context, namespaceUID, ownerUID, scope, name string, page, pageSize int) ([]bo.AlertPage, int, error)
	ListSelectableByUIDs(ctx context.Context, namespaceUID, ownerUID string, uids []string) ([]bo.AlertPage, error)

	ListUserAlertPages(ctx context.Context, namespaceUID, userUID string) ([]bo.AlertPage, error)
	SetUserAlertPages(ctx context.Context, namespaceUID, userUID string, pageUIDs []string) error

	// GetAlertColumns 读取用户在某个实时视图上保存的自定义列。没有配置时返回空。
	GetAlertColumns(ctx context.Context, namespaceUID, userUID, scope string) ([]bo.AlertColumn, error)
	// SetAlertColumns 覆盖写入用户在某个实时视图上的自定义列。
	SetAlertColumns(ctx context.Context, namespaceUID, userUID, scope string, columns []bo.AlertColumn) error
}
