package biz

import (
	"context"

	"github.com/aide-family/moon/cmd/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/palace/internal/biz/repository"
	"github.com/aide-family/moon/pkg/merr"
)

func NewRealtime(realtimeRepo repository.Realtime) *Realtime {
	return &Realtime{
		realtimeRepo: realtimeRepo,
	}
}

type Realtime struct {
	realtimeRepo repository.Realtime
}

func (r *Realtime) SaveAlert(ctx context.Context, alert *bo.Alert) error {
	if err := alert.Validate(); err != nil {
		return err
	}
	getRealtimeParams := &bo.GetAlertParams{
		TeamID:      alert.TeamID,
		Fingerprint: alert.Fingerprint,
		StartsAt:    alert.StartsAt,
	}
	exists, err := r.realtimeRepo.Exists(ctx, getRealtimeParams)
	if err != nil {
		return err
	}
	if exists {
		return r.realtimeRepo.UpdateAlert(ctx, alert)
	}
	return r.realtimeRepo.CreateAlert(ctx, alert)
}

func (r *Realtime) ListAlerts(ctx context.Context, params *bo.ListAlertParams) (*bo.ListAlertReply, error) {
	if len(params.TimeRange) != 2 {
		return nil, merr.ErrorInvalidArgument("time range must be 2")
	}
	return r.realtimeRepo.ListAlerts(ctx, params)
}
