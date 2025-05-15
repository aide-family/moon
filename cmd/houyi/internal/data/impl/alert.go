package impl

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/moon/cmd/houyi/internal/biz/bo"
	"github.com/aide-family/moon/cmd/houyi/internal/biz/do"
	"github.com/aide-family/moon/cmd/houyi/internal/biz/repository"
	"github.com/aide-family/moon/cmd/houyi/internal/biz/vobj"
	"github.com/aide-family/moon/cmd/houyi/internal/data"
	"github.com/aide-family/moon/pkg/api/common"
	"github.com/aide-family/moon/pkg/util/timex"
)

func NewAlertRepo(data *data.Data, logger log.Logger) repository.Alert {
	return &alertImpl{
		Data:   data,
		helper: log.NewHelper(log.With(logger, "module", "data.repo.alert")),
	}
}

type alertImpl struct {
	*data.Data
	helper *log.Helper
}

func (a *alertImpl) Delete(ctx context.Context, fingerprint string) error {
	key := vobj.AlertEventCacheKey.Key()
	if err := a.GetCache().Client().HDel(ctx, key, fingerprint).Err(); err != nil {
		a.helper.WithContext(ctx).Warnw("method", "DeleteAlert", "err", err)
		return err
	}
	return nil
}

func (a *alertImpl) GetAll(ctx context.Context) ([]bo.Alert, error) {
	key := vobj.AlertEventCacheKey.Key()
	exist, err := a.GetCache().Client().Exists(ctx, key).Result()
	if err != nil {
		a.helper.WithContext(ctx).Warnw("method", "GetAllAlerts", "err", err)
		return nil, err
	}
	if exist == 0 {
		return nil, nil
	}
	alertMap, err := a.GetCache().Client().HGetAll(ctx, key).Result()
	if err != nil {
		a.helper.WithContext(ctx).Warnw("method", "GetAllAlerts", "err", err)
		return nil, err
	}
	alerts := make([]bo.Alert, 0, len(alertMap))
	for _, v := range alertMap {
		var alert do.Alert
		if err := alert.UnmarshalBinary([]byte(v)); err != nil {
			a.helper.WithContext(ctx).Warnw("method", "UnmarshalBinary", "err", err)
			continue
		}
		alerts = append(alerts, &alert)
	}
	return alerts, nil
}

func (a *alertImpl) Get(ctx context.Context, fingerprint string) (bo.Alert, bool) {
	key := vobj.AlertEventCacheKey.Key()
	exist, err := a.GetCache().Client().HExists(ctx, key, fingerprint).Result()
	if err != nil {
		a.helper.WithContext(ctx).Warnw("method", "GetAlert", "err", err)
		return nil, false
	}
	if !exist {
		return nil, false
	}
	var alert do.Alert
	if err := a.GetCache().Client().HGet(ctx, key, fingerprint).Scan(&alert); err != nil {
		a.helper.WithContext(ctx).Warnw("method", "GetAlert", "err", err)
		return nil, false
	}
	return &alert, true
}

func (a *alertImpl) Save(ctx context.Context, alerts ...bo.Alert) error {
	if len(alerts) == 0 {
		return nil
	}
	key := vobj.AlertEventCacheKey.Key()
	alertMap := make(map[string]any, len(alerts))
	for _, alert := range alerts {
		fingerprint := alert.GetFingerprint()
		item := &do.Alert{
			Status:       alert.GetStatus(),
			Labels:       alert.GetLabels(),
			Annotations:  alert.GetAnnotations(),
			StartsAt:     alert.GetStartsAt(),
			EndsAt:       alert.GetEndsAt(),
			GeneratorURL: alert.GetGeneratorURL(),
			Fingerprint:  fingerprint,
			Value:        alert.GetValue(),
			Duration:     alert.GetDuration(),
			LastUpdated:  timex.Now(),
		}

		alertMap[fingerprint] = a.oldAlert(ctx, item)
	}
	if err := a.GetCache().Client().HSet(ctx, key, alertMap).Err(); err != nil {
		a.helper.WithContext(ctx).Warnw("method", "SaveAlert", "err", err)
		return err
	}
	return nil
}

func (a *alertImpl) oldAlert(ctx context.Context, newAlert *do.Alert) *do.Alert {
	key := vobj.AlertEventCacheKey.Key()
	exist, err := a.GetCache().Client().HExists(ctx, key, newAlert.GetFingerprint()).Result()
	if err != nil {
		a.helper.WithContext(ctx).Warnw("method", "SaveAlert", "err", err)
		return newAlert
	}
	if !exist {
		return newAlert
	}
	var oldAlert do.Alert
	if err := a.GetCache().Client().HGet(ctx, key, newAlert.GetFingerprint()).Scan(&oldAlert); err != nil {
		a.helper.WithContext(ctx).Warnw("method", "SaveAlert", "err", err)
		return newAlert
	}
	newAlert.Status = common.AlertStatus_firing
	newAlert.StartsAt = oldAlert.StartsAt

	return newAlert
}
