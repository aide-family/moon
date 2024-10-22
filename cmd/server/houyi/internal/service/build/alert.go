package build

import (
	"github.com/aide-family/moon/api"
	"github.com/aide-family/moon/cmd/server/houyi/internal/biz/bo"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
)

// AlarmAPIBuilder alarm api builder
type AlarmAPIBuilder struct {
	*api.AlarmItem
}

// NewAlarmAPIBuilder new alarm api builder
func NewAlarmAPIBuilder(alarm *api.AlarmItem) *AlarmAPIBuilder {
	return &AlarmAPIBuilder{AlarmItem: alarm}
}

// ToBo to bo
func (a *AlarmAPIBuilder) ToBo() *bo.Alarm {
	if types.IsNil(a) || types.IsNil(a.AlarmItem) {
		return nil
	}
	alerts := types.SliceTo(a.GetAlerts(), func(item *api.AlertItem) *bo.Alert {
		return NewAlertAPIBuilder(item).ToBo()
	})
	return &bo.Alarm{
		Alerts:            alerts,
		CommonAnnotations: a.GetCommonAnnotations(),
		CommonLabels:      vobj.NewLabels(a.GetCommonLabels()),
		ExternalURL:       a.GetExternalURL(),
		GroupKey:          a.GetGroupKey(),
		GroupLabels:       vobj.NewLabels(a.GetGroupLabels()),
		Receiver:          a.GetReceiver(),
		Status:            vobj.ToAlertStatus(a.GetStatus()),
		TruncatedAlerts:   a.GetTruncatedAlerts(),
		Version:           a.GetVersion(),
	}
}

// AlertAPIBuilder alert builder
type AlertAPIBuilder struct {
	*api.AlertItem
}

// NewAlertAPIBuilder new alert builder
func NewAlertAPIBuilder(alert *api.AlertItem) *AlertAPIBuilder {
	return &AlertAPIBuilder{AlertItem: alert}
}

// ToBo to bo
func (a *AlertAPIBuilder) ToBo() *bo.Alert {
	if types.IsNil(a) || types.IsNil(a.AlertItem) {
		return nil
	}
	return &bo.Alert{
		Status:       vobj.ToAlertStatus(a.GetStatus()),
		Labels:       vobj.NewLabels(a.GetLabels()),
		Annotations:  a.Annotations,
		StartsAt:     types.NewTimeByString(a.GetStartsAt()),
		EndsAt:       types.NewTimeByString(a.GetEndsAt()),
		GeneratorURL: a.GetGeneratorURL(),
		Fingerprint:  a.GetFingerprint(),
		Value:        a.GetValue(),
	}
}
