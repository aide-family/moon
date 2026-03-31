package repository

import (
	"context"
	"time"

	"github.com/bwmarrin/snowflake"

	"github.com/aide-family/marksman/internal/biz/bo"
)

type AlertEvent interface {
	SaveAlertEvent(ctx context.Context, event *bo.AlertEventBo) (snowflake.ID, error)
	GetAlertEvent(ctx context.Context, uid snowflake.ID) (*bo.AlertEventItemBo, error)
	GetAlertEventByFingerprint(ctx context.Context, uid snowflake.ID, fingerprint string) (*bo.AlertEventItemBo, error)
	ListRealtimeAlert(ctx context.Context, req *bo.ListRealtimeAlertBo, pageFilter *bo.AlertPageFilterBo) (*bo.PageResponseBo[*bo.AlertEventItemBo], error)
	// CountActiveAlerts returns count of alerts with status FIRING in the time range; pageFilter optional.
	CountActiveAlerts(ctx context.Context, startAt, endAt time.Time, pageFilter *bo.AlertPageFilterBo) (int64, error)
	// CountActiveAlertsByLevel returns per-level counts for active alerts; LevelName is not set.
	CountActiveAlertsByLevel(ctx context.Context, startAt, endAt time.Time, pageFilter *bo.AlertPageFilterBo) ([]bo.LevelCountBo, error)
	// CountRecoveredAlertsSince returns count of alerts with status Recovered and recovered_at >= since.
	CountRecoveredAlertsSince(ctx context.Context, since time.Time) (int64, error)
	InterveneAlert(ctx context.Context, req *bo.InterveneAlertBo) error
	BatchInterveneAlert(ctx context.Context, req *bo.BatchInterveneAlertBo) error
	SuppressAlert(ctx context.Context, req *bo.SuppressAlertBo) error
	RecoverAlert(ctx context.Context, req *bo.RecoverAlertBo) error
	BatchRecoverAlert(ctx context.Context, req *bo.BatchRecoverAlertBo) error
	AutoRecoverAlert(ctx context.Context, uid snowflake.ID) error
}
