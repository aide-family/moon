package builder

import (
	"context"

	"github.com/aide-family/moon/api"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
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

	// ICreateAlarmItemRequestBuilder 创建报警项请求构建器
	ICreateAlarmItemRequestBuilder interface {
	}

	createAlarmItemRequestBuilder struct {
		ctx context.Context
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
		GroupLabels:       vobj.NewLabels(a.GetGroupLabels()),
		CommonLabels:      vobj.NewLabels(a.GetCommonLabels()),
		CommonAnnotations: a.GetCommonAnnotations(),
		ExternalURL:       a.GetExternalURL(),
		Version:           a.GetVersion(),
		GroupKey:          a.GetGroupKey(),
		TruncatedAlerts:   a.GetTruncatedAlerts(),
		Fingerprint:       a.GetFingerprint(),
		Alerts: types.SliceTo(a.Alerts, func(item *api.AlertItem) *bo.CreateAlarmItemParams {
			startsAt, _ := types.StrToInt64(item.StartsAt)
			endsAt, _ := types.StrToInt64(item.EndsAt)
			return &bo.CreateAlarmItemParams{
				Status:       item.GetStatus(),
				Labels:       item.GetLabels(),
				Annotations:  item.GetAnnotations(),
				StartsAt:     startsAt,
				EndsAt:       endsAt,
				GeneratorURL: item.GetGeneratorURL(),
				Fingerprint:  item.GetFingerprint(),
				StrategyID:   vobj.NewLabels(a.GetCommonLabels()).GetStrategyID(),
				LevelID:      vobj.NewLabels(a.GetGroupLabels()).GetLevelID(),
			}
		}),
	}
}

func (a *alarmModuleBuilder) WithCreateAlarmRawInfoRequest(item *api.AlarmItem) ICreateAlarmRawInfoRequestBuilder {
	return &createAlarmRequestBuilder{
		ctx:       a.ctx,
		AlarmItem: item,
	}
}
