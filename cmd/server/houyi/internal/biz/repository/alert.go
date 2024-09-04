package repository

import (
	"context"

	"github.com/aide-family/moon/cmd/server/houyi/internal/biz/bo"
)

// Alert .
type Alert interface {
	// SaveAlarm 保存告警
	SaveAlarm(ctx context.Context, alarm *bo.Alarm) error
	// PushAlarm 推送告警
	PushAlarm(ctx context.Context, alarm *bo.Alarm) error
}
