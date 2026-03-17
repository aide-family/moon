package repository

import (
	"context"
	"time"

	"github.com/bwmarrin/snowflake"

	"github.com/aide-family/marksman/internal/biz/bo"
)

type AlertEvent interface {
	CreateAlertEvent(ctx context.Context, req *bo.AlertEventBo, strategyGroupUID snowflake.ID) (snowflake.ID, error)
	GetAlertEvent(ctx context.Context, uid snowflake.ID) (*bo.AlertEventItemBo, error)
	GetAlertEventByFingerprint(ctx context.Context, uid snowflake.ID, fingerprint string) (*bo.AlertEventItemBo, error)
	ListRealtimeAlert(ctx context.Context, req *bo.ListRealtimeAlertBo, pageFilter *bo.AlertPageFilterBo) (*bo.PageResponseBo[*bo.AlertEventItemBo], error)
	InterveneAlert(ctx context.Context, uid snowflake.ID, by snowflake.ID) error
	SuppressAlert(ctx context.Context, uid snowflake.ID, until time.Time) error
	RecoverAlert(ctx context.Context, uid snowflake.ID, by snowflake.ID) error
	AutoRecoverAlert(ctx context.Context, uid snowflake.ID) error
}
