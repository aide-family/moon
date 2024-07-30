package biz

import (
	"context"

	"github.com/aide-family/moon/cmd/server/houyi/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/houyi/internal/biz/repository"
)

// NewAlertBiz new AlertBiz
func NewAlertBiz(alertRepository repository.Alert) *AlertBiz {
	return &AlertBiz{
		alertRepository: alertRepository,
	}
}

// AlertBiz .
type AlertBiz struct {
	alertRepository repository.Alert
}

// SaveAlarm 保存告警数据
func (a *AlertBiz) SaveAlarm(ctx context.Context, alarm *bo.Alarm) error {
	return a.alertRepository.SaveAlarm(ctx, alarm)
}
