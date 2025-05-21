package event

import (
	"context"
	"time"

	"github.com/aide-family/moon/cmd/houyi/internal/biz/vobj"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/robfig/cron/v3"

	"github.com/aide-family/moon/cmd/houyi/internal/biz/bo"
	"github.com/aide-family/moon/cmd/houyi/internal/biz/repository"
	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/plugin/server"
	"github.com/aide-family/moon/pkg/util/timex"
	"github.com/aide-family/moon/pkg/util/validate"
)

func NewAlertJob(alert bo.Alert, opts ...AlertJobOption) (bo.AlertJob, error) {
	a := &alertJob{
		alert: alert,
	}
	for _, opt := range opts {
		if err := opt(a); err != nil {
			return nil, err
		}
	}
	checkOpts := []*checkItem{
		{"alertRepo", a.alertRepo},
		{"eventBusRepo", a.eventBusRepo},
		{"cacheRepo", a.cacheRepo},
		{"helper", a.helper},
	}
	return a, checkList(checkOpts...)
}

type AlertJobOption func(*alertJob) error

func WithAlertJobAlertRepo(alertRepo repository.Alert) AlertJobOption {
	return func(a *alertJob) error {
		if validate.IsNil(alertRepo) {
			return merr.ErrorInternalServer("alertRepo is nil")
		}
		a.alertRepo = alertRepo
		return nil
	}
}

func WithAlertJobEventBusRepo(eventBusRepo repository.EventBus) AlertJobOption {
	return func(a *alertJob) error {
		if validate.IsNil(eventBusRepo) {
			return merr.ErrorInternalServer("eventBusRepo is nil")
		}
		a.eventBusRepo = eventBusRepo
		return nil
	}
}

func WithAlertJobHelper(logger log.Logger) AlertJobOption {
	return func(a *alertJob) error {
		if validate.IsNil(logger) {
			return merr.ErrorInternalServer("logger is nil")
		}
		a.helper = log.NewHelper(log.With(logger, "module", "event.alert", "jobKey", a.alert.GetFingerprint()))
		return nil
	}
}

func WithAlertJobCacheRepo(cacheRepo repository.Cache) AlertJobOption {
	return func(a *alertJob) error {
		if validate.IsNil(cacheRepo) {
			return merr.ErrorInternalServer("cacheRepo is nil")
		}
		a.cacheRepo = cacheRepo
		return nil
	}
}

type alertJob struct {
	alert bo.Alert

	id           cron.EntryID
	alertRepo    repository.Alert
	eventBusRepo repository.EventBus
	cacheRepo    repository.Cache

	helper *log.Helper
}

func (a *alertJob) GetAlert() bo.Alert {
	return a.alert
}

func (a *alertJob) isSustaining() (alert bo.Alert, sustaining bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	defer func() {
		a.helper.Debugw("sustaining", sustaining)
		if sustaining {
			return
		}
		if err := a.alertRepo.Delete(ctx, a.alert.GetFingerprint()); err != nil {
			a.helper.Warnw("msg", "delete alertInfo error", "error", err)
		}
	}()
	alertInfo, ok := a.alertRepo.Get(ctx, a.alert.GetFingerprint())
	if !ok {
		return a.alert, false
	}
	return alertInfo, alertInfo.GetLastUpdated().Add(a.alert.GetDuration()).After(timex.Now())
}

func (a *alertJob) Run() {
	lockKey := vobj.AlertJobLockKey.Key(a.alert.GetFingerprint())
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	locked, err := a.cacheRepo.Lock(ctx, lockKey, a.alert.GetDuration())
	if err != nil {
		a.helper.Errorw("msg", "lock error", "error", err)
		return
	}
	if !locked {
		return
	}
	defer func(cacheRepo repository.Cache, ctx context.Context, key string) {
		if err := cacheRepo.Unlock(ctx, key); err != nil {
			a.helper.Warnw("err", err, "msg", "unlock error")
		}
	}(a.cacheRepo, ctx, lockKey)

	alertInfo, ok := a.isSustaining()
	if !ok {
		alertInfo.Resolved()
		a.alert = alertInfo
		a.eventBusRepo.InAlertJobEventBus() <- a
		a.eventBusRepo.InAlertEventBus() <- alertInfo
		return
	}
	if alertInfo.IsFiring() {
		a.helper.Debugw("msg", "alert is firing")
		a.eventBusRepo.InAlertEventBus() <- alertInfo
	}
}

func (a *alertJob) ID() cron.EntryID {
	if a == nil {
		return 0
	}
	return a.id
}

func (a *alertJob) Index() string {
	return a.alert.GetFingerprint()
}

func (a *alertJob) Spec() server.CronSpec {
	if a == nil {
		return server.CronSpecEvery(1 * time.Minute)
	}
	return server.CronSpecEvery(a.alert.GetDuration())
}

func (a *alertJob) WithID(id cron.EntryID) server.CronJob {
	a.id = id
	return a
}

// IsImmediate implements server.CronJob.
func (a *alertJob) IsImmediate() bool {
	return false
}
