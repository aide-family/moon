package repository

import (
	"context"
	"time"

	"github.com/aide-family/moon/app/alioth/pkg/bo"
)

// AlertMaintenanceRepo 管理维护窗口，并在启停/删除时同步解除或应用抑制。
type AlertMaintenanceRepo interface {
	CreateWindow(ctx context.Context, in bo.AlertMaintenanceWindow) (*bo.AlertMaintenanceWindow, error)
	UpdateWindow(ctx context.Context, in bo.AlertMaintenanceWindow) (*bo.AlertMaintenanceWindow, error)
	UpdateWindowStatus(ctx context.Context, namespaceUID, uid, status string) (*bo.AlertMaintenanceWindow, error)
	DeleteWindow(ctx context.Context, namespaceUID, uid string) error
	GetWindow(ctx context.Context, namespaceUID, uid string) (*bo.AlertMaintenanceWindow, error)
	ListWindows(ctx context.Context, namespaceUID string, filter bo.AlertMaintenanceWindowFilter, page, pageSize int) ([]bo.AlertMaintenanceWindow, int, error)
	ListActiveWindows(ctx context.Context, namespaceUID string, at time.Time) ([]bo.AlertMaintenanceWindow, error)
}
