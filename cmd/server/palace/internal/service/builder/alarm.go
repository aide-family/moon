package builder

import (
	"context"

	"github.com/aide-family/moon/api"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/pkg/label"
	"github.com/aide-family/moon/pkg/util/types"
)

var _ IAlarmModuleBuilder = (*alarmModuleBuilder)(nil)

type (
	alarmModuleBuilder struct {
		ctx context.Context
	}

	// IAlarmModuleBuilder 告警模块构建器
	IAlarmModuleBuilder interface {
		WithCreateAlarmRawInfoRequest(*api.AlarmItem) ICreateAlarmRawInfoRequestBuilder
	}

	// ICreateAlarmRawInfoRequestBuilder 创建告警原始数据请求构建器
	ICreateAlarmRawInfoRequestBuilder interface {
		ToBo() *bo.CreateAlarmHookRawParams
	}

	createAlarmRequestBuilder struct {
		ctx context.Context
		*api.AlarmItem
	}
)

func (a *createAlarmRequestBuilder) ToBo() *bo.CreateAlarmHookRawParams {
	if types.IsNil(a) || types.IsNil(a.AlarmItem) {
		return nil
	}
	return &bo.CreateAlarmHookRawParams{
		Receiver:          a.GetReceiver(),
		Status:            a.GetStatus(),
		GroupLabels:       label.NewLabels(a.GetGroupLabels()),
		CommonLabels:      label.NewLabels(a.GetCommonLabels()),
		CommonAnnotations: a.GetCommonAnnotations(),
		ExternalURL:       a.GetExternalURL(),
		Version:           a.GetVersion(),
		GroupKey:          a.GetGroupKey(),
		TruncatedAlerts:   a.GetTruncatedAlerts(),
		Alerts: types.SliceTo(a.Alerts, func(item *api.AlertItem) *bo.AlertItemRawParams {
			return &bo.AlertItemRawParams{
				Status:       item.GetStatus(),
				Labels:       item.GetLabels(),
				Annotations:  item.GetAnnotations(),
				StartsAt:     item.GetStartsAt(),
				EndsAt:       item.GetEndsAt(),
				GeneratorURL: item.GetGeneratorURL(),
				Fingerprint:  item.GetFingerprint(),
				Value:        item.GetValue(),
			}
		}),
		TeamID:     label.NewLabels(a.GetGroupLabels()).GetTeamID(),
		StrategyID: label.NewLabels(a.GetCommonLabels()).GetStrategyID(),
		LevelID:    label.NewLabels(a.GetGroupLabels()).GetLevelID(),
	}
}

func (a *alarmModuleBuilder) WithCreateAlarmRawInfoRequest(item *api.AlarmItem) ICreateAlarmRawInfoRequestBuilder {
	return &createAlarmRequestBuilder{
		ctx:       a.ctx,
		AlarmItem: item,
	}
}
