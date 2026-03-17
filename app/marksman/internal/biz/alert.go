package biz

import (
	"context"
	"time"

	"github.com/aide-family/magicbox/contextx"
	"github.com/aide-family/magicbox/merr"
	"github.com/bwmarrin/snowflake"
	klog "github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/marksman/internal/biz/bo"
	"github.com/aide-family/marksman/internal/biz/repository"
)

func NewAlert(
	alertPageRepo repository.AlertPage,
	alertEventRepo repository.AlertEvent,
	helper *klog.Helper,
) *AlertBiz {
	return &AlertBiz{
		alertPageRepo:  alertPageRepo,
		alertEventRepo: alertEventRepo,
		helper:         klog.NewHelper(klog.With(helper.Logger(), "biz", "alert")),
	}
}

type AlertBiz struct {
	alertPageRepo  repository.AlertPage
	alertEventRepo repository.AlertEvent
	helper         *klog.Helper
}

func (b *AlertBiz) ListRealtimeAlert(ctx context.Context, req *bo.ListRealtimeAlertBo) (*bo.PageResponseBo[*bo.AlertEventItemBo], error) {
	page, err := b.alertPageRepo.GetAlertPage(ctx, req.AlertPageUID)
	if err != nil {
		if merr.IsNotFound(err) {
			return nil, merr.ErrorNotFound("alert page not found")
		}
		b.helper.Errorw("msg", "get alert page failed", "error", err, "alertPageUID", req.AlertPageUID.Int64())
		return nil, merr.ErrorInternalServer("get alert page failed").WithCause(err)
	}
	var filter *bo.AlertPageFilterBo
	if page.Filter != nil {
		filter = page.Filter
	}
	result, err := b.alertEventRepo.ListRealtimeAlert(ctx, req, filter)
	if err != nil {
		b.helper.Errorw("msg", "list realtime alert failed", "error", err)
		return nil, merr.ErrorInternalServer("list realtime alert failed").WithCause(err)
	}
	return result, nil
}

func (b *AlertBiz) InterveneAlert(ctx context.Context, uid snowflake.ID) error {
	userUID := contextx.GetUserUID(ctx)
	if err := b.alertEventRepo.InterveneAlert(ctx, uid, userUID); err != nil {
		if merr.IsNotFound(err) {
			return merr.ErrorNotFound("alert event not found")
		}
		b.helper.Errorw("msg", "intervene alert failed", "error", err, "uid", uid.Int64())
		return merr.ErrorInternalServer("intervene alert failed").WithCause(err)
	}
	return nil
}

func (b *AlertBiz) SuppressAlert(ctx context.Context, uid snowflake.ID, suppressUntil time.Time) error {
	if err := b.alertEventRepo.SuppressAlert(ctx, uid, suppressUntil); err != nil {
		if merr.IsNotFound(err) {
			return merr.ErrorNotFound("alert event not found")
		}
		b.helper.Errorw("msg", "suppress alert failed", "error", err, "uid", uid.Int64())
		return merr.ErrorInternalServer("suppress alert failed").WithCause(err)
	}
	return nil
}

func (b *AlertBiz) RecoverAlert(ctx context.Context, uid snowflake.ID) error {
	userUID := contextx.GetUserUID(ctx)
	if err := b.alertEventRepo.RecoverAlert(ctx, uid, userUID); err != nil {
		if merr.IsNotFound(err) {
			return merr.ErrorNotFound("alert event not found")
		}
		b.helper.Errorw("msg", "recover alert failed", "error", err, "uid", uid.Int64())
		return merr.ErrorInternalServer("recover alert failed").WithCause(err)
	}
	return nil
}
