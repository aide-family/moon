package biz

import (
	"context"
	"time"

	goddessv1 "github.com/aide-family/goddess/pkg/api/v1"
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
	userAlertPageRepo repository.UserAlertPage,
	levelRepo repository.Level,
	memberRepo repository.Member,
	helper *klog.Helper,
) *AlertBiz {
	return &AlertBiz{
		alertPageRepo:     alertPageRepo,
		alertEventRepo:    alertEventRepo,
		userAlertPageRepo: userAlertPageRepo,
		levelRepo:         levelRepo,
		memberRepo:        memberRepo,
		helper:            klog.NewHelper(klog.With(helper.Logger(), "biz", "alert")),
	}
}

type AlertBiz struct {
	alertPageRepo     repository.AlertPage
	alertEventRepo    repository.AlertEvent
	userAlertPageRepo repository.UserAlertPage
	levelRepo         repository.Level
	memberRepo        repository.Member
	helper            *klog.Helper
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

	result, err := b.alertEventRepo.ListRealtimeAlert(ctx, req, page.Filter)
	if err != nil {
		b.helper.Errorw("msg", "list realtime alert failed", "error", err)
		return nil, merr.ErrorInternalServer("list realtime alert failed").WithCause(err)
	}
	return result, nil
}

func (b *AlertBiz) InterveneAlert(ctx context.Context, req *bo.InterveneAlertBo) error {
	if req == nil || req.UID.Int64() <= 0 {
		return merr.ErrorParams("uid must be greater than 0")
	}
	if req.IntervenedByUser.Int64() <= 0 {
		return merr.ErrorParams("intervenedByUser must be greater than 0")
	}

	// Single intervene: do not choose member on the client; resolve memberId by current userId.
	mlist, err := b.memberRepo.ListMember(ctx, &goddessv1.ListMemberRequest{
		Page:     1,
		PageSize: 1,
		UserUID:  req.IntervenedByUser.Int64(),
	})
	if err != nil {
		b.helper.Errorw("msg", "list member failed", "error", err, "userUID", req.IntervenedByUser.Int64())
		return merr.ErrorInternalServer("list member failed").WithCause(err)
	}
	items := mlist.GetItems()
	if len(items) == 0 {
		return merr.ErrorNotFound("member not found")
	}
	req.IntervenedBy = snowflake.ParseInt64(items[0].GetUid())
	req.IntervenedByName = items[0].GetName()

	if err := b.alertEventRepo.InterveneAlert(ctx, req); err != nil {
		if merr.IsNotFound(err) {
			return merr.ErrorNotFound("alert event not found")
		}
		b.helper.Errorw("msg", "intervene alert failed", "error", err, "uid", req.UID.Int64())
		return merr.ErrorInternalServer("intervene alert failed").WithCause(err)
	}
	return nil
}

func (b *AlertBiz) BatchInterveneAlert(ctx context.Context, req *bo.BatchInterveneAlertBo) error {
	if req == nil || len(req.UIDs) == 0 {
		return merr.ErrorParams("uids are required")
	}
	if req.IntervenedBy.Int64() <= 0 {
		return merr.ErrorParams("intervenedMemberUid must be greater than 0")
	}
	member, err := b.memberRepo.GetMember(ctx, &goddessv1.GetMemberRequest{Uid: req.IntervenedBy.Int64()})
	if err != nil {
		b.helper.Errorw("msg", "get member failed", "error", err, "memberUID", req.IntervenedBy.Int64())
		return merr.ErrorInternalServer("get member failed").WithCause(err)
	}
	req.IntervenedByName = member.GetName()

	for _, uid := range req.UIDs {
		if uid.Int64() <= 0 {
			return merr.ErrorParams("uid must be greater than 0")
		}
	}
	if err := b.alertEventRepo.BatchInterveneAlert(ctx, req); err != nil {
		if merr.IsNotFound(err) {
			return merr.ErrorNotFound("alert event not found")
		}
		b.helper.Errorw("msg", "batch intervene alert failed", "error", err)
		return merr.ErrorInternalServer("batch intervene alert failed").WithCause(err)
	}
	return nil
}

func (b *AlertBiz) SuppressAlert(ctx context.Context, req *bo.SuppressAlertBo) error {
	if err := b.alertEventRepo.SuppressAlert(ctx, req); err != nil {
		if merr.IsNotFound(err) {
			return merr.ErrorNotFound("alert event not found")
		}
		b.helper.Errorw("msg", "suppress alert failed", "error", err, "uid", req.UID.Int64())
		return merr.ErrorInternalServer("suppress alert failed").WithCause(err)
	}
	return nil
}

func (b *AlertBiz) RecoverAlert(ctx context.Context, req *bo.RecoverAlertBo) error {
	if err := b.alertEventRepo.RecoverAlert(ctx, req); err != nil {
		if merr.IsNotFound(err) {
			return merr.ErrorNotFound("alert event not found")
		}
		b.helper.Errorw("msg", "recover alert failed", "error", err, "uid", req.UID.Int64())
		return merr.ErrorInternalServer("recover alert failed").WithCause(err)
	}
	return nil
}

func (b *AlertBiz) GetAlertEvent(ctx context.Context, uid snowflake.ID) (*bo.AlertEventItemBo, error) {
	item, err := b.alertEventRepo.GetAlertEvent(ctx, uid)
	if err != nil {
		if merr.IsNotFound(err) {
			return nil, merr.ErrorNotFound("alert event not found")
		}
		b.helper.Errorw("msg", "get alert event failed", "error", err, "uid", uid.Int64())
		return nil, merr.ErrorInternalServer("get alert event failed").WithCause(err)
	}
	return item, nil
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

	stats := &bo.AlertStatisticsBo{
		TotalActiveCount:    totalActive,
		CountByLevel:        byLevel,
		TodayRecoveredCount: todayRecovered,
		CountByAlertPage:    nil,
	}

	userUID := contextx.GetUserUID(ctx)
	userPageUIDs, err := b.userAlertPageRepo.GetUserAlertPageUIDs(ctx, userUID)
	if err != nil {
		b.helper.Errorw("msg", "get user alert page uids failed", "error", err)
		return nil, merr.ErrorInternalServer("get user alert page uids failed").WithCause(err)
	}
	if len(userPageUIDs) == 0 {
		return stats, nil
	}
	userPageList, err := b.alertPageRepo.GetAlertPagesByUIDs(ctx, userPageUIDs)
	if err != nil {
		b.helper.Errorw("msg", "get alert pages by uids failed", "error", err)
		return nil, merr.ErrorInternalServer("get alert pages by uids failed").WithCause(err)
	}
	stats.CountByAlertPage = make([]bo.AlertPageCountBo, 0, len(userPageList))
	for _, page := range userPageList {
		var filter *bo.AlertPageFilterBo
		if page.Filter != nil {
			filter = page.Filter
		}
		cnt, err := b.alertEventRepo.CountActiveAlerts(ctx, startAt, endAt, filter)
		if err != nil {
			b.helper.Errorw("msg", "count active alerts for page failed", "error", err, "alertPageUID", page.UID.Int64())
			continue
		}
		stats.CountByAlertPage = append(stats.CountByAlertPage, bo.AlertPageCountBo{
			AlertPageUID:  page.UID,
			AlertPageName: page.Name,
			Count:         cnt,
		})
	}

	return stats, nil
}
