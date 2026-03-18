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
	levelRepo repository.Level,
	helper *klog.Helper,
) *AlertBiz {
	return &AlertBiz{
		alertPageRepo:  alertPageRepo,
		alertEventRepo: alertEventRepo,
		levelRepo:      levelRepo,
		helper:         klog.NewHelper(klog.With(helper.Logger(), "biz", "alert")),
	}
}

type AlertBiz struct {
	alertPageRepo  repository.AlertPage
	alertEventRepo repository.AlertEvent
	levelRepo      repository.Level
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

func (b *AlertBiz) GetAlertStatistics(ctx context.Context) (*bo.AlertStatisticsBo, error) {
	now := time.Now()
	startAt := now.Add(-bo.ListRealtimeAlertTimeRangeDefault)
	endAt := now

	totalActive, err := b.alertEventRepo.CountActiveAlerts(ctx, startAt, endAt, nil)
	if err != nil {
		b.helper.Errorw("msg", "count active alerts failed", "error", err)
		return nil, merr.ErrorInternalServer("count active alerts failed").WithCause(err)
	}

	byLevel, err := b.alertEventRepo.CountActiveAlertsByLevel(ctx, startAt, endAt, nil)
	if err != nil {
		b.helper.Errorw("msg", "count active alerts by level failed", "error", err)
		return nil, merr.ErrorInternalServer("count active alerts by level failed").WithCause(err)
	}
	levelUIDs := make([]int64, 0, len(byLevel))
	for _, c := range byLevel {
		levelUIDs = append(levelUIDs, c.LevelUID.Int64())
	}
	levelNames, err := b.levelRepo.GetLevelNamesByUIDs(ctx, levelUIDs)
	if err != nil {
		b.helper.Errorw("msg", "get level names failed", "error", err)
		return nil, merr.ErrorInternalServer("get level names failed").WithCause(err)
	}
	for i := range byLevel {
		byLevel[i].LevelName = levelNames[byLevel[i].LevelUID.Int64()]
	}

	startOfToday := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	todayRecovered, err := b.alertEventRepo.CountRecoveredAlertsSince(ctx, startOfToday)
	if err != nil {
		b.helper.Errorw("msg", "count recovered alerts since failed", "error", err)
		return nil, merr.ErrorInternalServer("count recovered alerts since failed").WithCause(err)
	}

	pageList, err := b.alertPageRepo.ListAlertPage(ctx, &bo.ListAlertPageBo{
		PageRequestBo: bo.NewPageRequestBo(1, 200),
	})
	if err != nil {
		b.helper.Errorw("msg", "list alert pages failed", "error", err)
		return nil, merr.ErrorInternalServer("list alert pages failed").WithCause(err)
	}
	byPage := make([]bo.AlertPageCountBo, 0, len(pageList.GetItems()))
	for _, page := range pageList.GetItems() {
		var filter *bo.AlertPageFilterBo
		if page.Filter != nil {
			filter = page.Filter
		}
		cnt, err := b.alertEventRepo.CountActiveAlerts(ctx, startAt, endAt, filter)
		if err != nil {
			b.helper.Errorw("msg", "count active alerts for page failed", "error", err, "alertPageUID", page.UID.Int64())
			continue
		}
		byPage = append(byPage, bo.AlertPageCountBo{
			AlertPageUID:  page.UID,
			AlertPageName: page.Name,
			Count:         cnt,
		})
	}

	return &bo.AlertStatisticsBo{
		TotalActiveCount:    totalActive,
		CountByLevel:        byLevel,
		TodayRecoveredCount: todayRecovered,
		CountByAlertPage:    byPage,
	}, nil
}
