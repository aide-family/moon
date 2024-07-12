package build

import (
	"github.com/aide-family/moon/api"
	"github.com/aide-family/moon/cmd/server/houyi/internal/biz/bo"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
)

type AlarmApiBuilder struct {
	*api.Strategy
}

func NewAlarmApiBuilder(strategyInfo *api.Strategy) *AlarmApiBuilder {
	return &AlarmApiBuilder{
		Strategy: strategyInfo,
	}
}

// ToBo 转换为业务对象
func (a *AlarmApiBuilder) ToBo() *bo.Strategy {
	if types.IsNil(a) || types.IsNil(a.Strategy) {
		return nil
	}
	strategyInfo := a.Strategy
	return &bo.Strategy{
		ID:                         strategyInfo.GetId(),
		Alert:                      strategyInfo.GetAlert(),
		Expr:                       strategyInfo.GetExpr(),
		For:                        types.NewDuration(strategyInfo.GetFor()),
		Count:                      strategyInfo.GetCount(),
		SustainType:                vobj.Sustain(strategyInfo.GetSustainType()),
		MultiDatasourceSustainType: vobj.MultiDatasourceSustain(strategyInfo.GetMultiDatasourceSustainType()),
		Labels:                     vobj.NewLabels(strategyInfo.GetLabels()),
		Annotations:                strategyInfo.GetAnnotations(),
		Interval:                   types.NewDuration(strategyInfo.GetInterval()),
		Datasource: types.SliceTo(strategyInfo.GetDatasource(), func(ds *api.Datasource) *bo.Datasource {
			return NewDatasourceApiBuilder(ds).ToBo()
		}),
		Status:    vobj.Status(strategyInfo.GetStatus()),
		Step:      strategyInfo.GetStep(),
		Condition: vobj.Condition(strategyInfo.GetCondition()),
		Threshold: strategyInfo.GetThreshold(),
	}
}

type AlarmBuilder struct {
	*bo.Alarm
}

func NewAlarmBuilder(alarm *bo.Alarm) *AlarmBuilder {
	return &AlarmBuilder{
		Alarm: alarm,
	}
}

func (a *AlarmBuilder) ToApi() *api.Alarm {
	if types.IsNil(a) || types.IsNil(a.Alarm) {
		return nil
	}
	alarm := a.Alarm
	return &api.Alarm{
		Receiver: alarm.Receiver,
		Status:   alarm.Status.String(),
		Alerts: types.SliceTo(alarm.Alerts, func(alert *bo.Alert) *api.Alert {
			return NewAlertBuilder(alert).ToApi()
		}),
		GroupLabels:       alarm.GroupLabels.Map(),
		CommonLabels:      alarm.CommonLabels.Map(),
		CommonAnnotations: alarm.CommonAnnotations,
		ExternalURL:       alarm.ExternalURL,
		Version:           alarm.Version,
		GroupKey:          alarm.GroupKey,
		TruncatedAlerts:   alarm.TruncatedAlerts,
	}
}

type AlertBuilder struct {
	*bo.Alert
}

func NewAlertBuilder(alert *bo.Alert) *AlertBuilder {
	return &AlertBuilder{
		Alert: alert,
	}
}

func (a *AlertBuilder) ToApi() *api.Alert {
	if types.IsNil(a) || types.IsNil(a.Alert) {
		return nil
	}
	alert := a.Alert
	return &api.Alert{
		Status:       alert.Status.String(),
		Labels:       alert.Labels.Map(),
		Annotations:  alert.Annotations,
		StartsAt:     alert.StartsAt.String(),
		EndsAt:       alert.EndsAt.String(),
		GeneratorURL: alert.GeneratorURL,
		Fingerprint:  alert.Fingerprint,
	}
}
