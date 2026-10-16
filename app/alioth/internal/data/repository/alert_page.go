package repository

import (
	"context"

	"github.com/aide-family/Alioth/app/alioth/pkg/bo"
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
}
