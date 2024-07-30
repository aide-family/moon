package build

import (
	"github.com/aide-family/moon/api"
	"github.com/aide-family/moon/cmd/server/houyi/internal/biz/bo"
	"github.com/aide-family/moon/pkg/util/types"
)

// AlarmBuilder alarm 构造器
type AlarmBuilder struct {
	*bo.Alarm
}

// NewAlarmBuilder 创建 alarm 构造器
func NewAlarmBuilder(alarm *bo.Alarm) *AlarmBuilder {
	return &AlarmBuilder{
		Alarm: alarm,
	}
}

// ToAPI 转换为 api 对象
func (a *AlarmBuilder) ToAPI() *api.AlarmItem {
	if types.IsNil(a) || types.IsNil(a.Alarm) {
		return nil
	}
	alarm := a.Alarm
	return &api.AlarmItem{
		Receiver: alarm.Receiver,
		Status:   alarm.Status.String(),
		Alerts: types.SliceTo(alarm.Alerts, func(alert *bo.Alert) *api.AlertItem {
			return NewAlertBuilder(alert).ToAPI()
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

// AlertBuilder alert 构造器
type AlertBuilder struct {
	*bo.Alert
}

// NewAlertBuilder 创建 alert 构造器
func NewAlertBuilder(alert *bo.Alert) *AlertBuilder {
	return &AlertBuilder{
		Alert: alert,
	}
}

// ToAPI 转换为 api 对象
func (a *AlertBuilder) ToAPI() *api.AlertItem {
	if types.IsNil(a) || types.IsNil(a.Alert) {
		return nil
	}
	alert := a.Alert
	return &api.AlertItem{
		Status:       alert.Status.String(),
		Labels:       alert.Labels.Map(),
		Annotations:  alert.Annotations,
		StartsAt:     alert.StartsAt.String(),
		EndsAt:       alert.EndsAt.String(),
		GeneratorURL: alert.GeneratorURL,
		Fingerprint:  alert.Fingerprint,
	}
}
