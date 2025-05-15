package repository

import (
	"context"

	"github.com/moon-monitor/moon/cmd/palace/internal/biz/bo"
	"github.com/moon-monitor/moon/cmd/palace/internal/biz/do"
)

type Realtime interface {
	Exists(ctx context.Context, alert *bo.GetAlertParams) (bool, error)
	GetAlert(ctx context.Context, alert *bo.GetAlertParams) (do.Realtime, error)
	CreateAlert(ctx context.Context, alert *bo.Alert) error
	UpdateAlert(ctx context.Context, alert *bo.Alert) error
	ListAlerts(ctx context.Context, params *bo.ListAlertParams) (*bo.ListAlertReply, error)
}
