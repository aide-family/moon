package bo

import (
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
	"github.com/aide-family/moon/pkg/watch"
)

var _ watch.Indexer = (*Alarm)(nil)
var _ watch.Indexer = (*Alert)(nil)

type (
	// Alarm alarm detail info
	Alarm struct {
		Receiver          string           `json:"receiver"`
		Status            vobj.AlertStatus `json:"status"`
		Alerts            []*Alert         `json:"alerts"`
		GroupLabels       *vobj.Labels     `json:"groupLabels"`
		CommonLabels      *vobj.Labels     `json:"commonLabels"`
		CommonAnnotations vobj.Annotations `json:"commonAnnotations"`
		ExternalURL       string           `json:"externalURL"`
		Version           string           `json:"version"`
		GroupKey          string           `json:"groupKey"`
		TruncatedAlerts   int32            `json:"truncatedAlerts"`
	}

	alarmInfo struct {
		Receiver          string            `json:"receiver"`
		Status            string            `json:"status"`
		Alerts            []*alertInfo      `json:"alerts"`
		GroupLabels       map[string]string `json:"groupLabels"`
		CommonLabels      map[string]string `json:"commonLabels"`
		CommonAnnotations map[string]string `json:"commonAnnotations"`
		ExternalURL       string            `json:"externalURL"`
		Version           string            `json:"version"`
		GroupKey          string            `json:"groupKey"`
		Fingerprint       string            `json:"fingerprint"`
	}

	// Alert alert detail info
	Alert struct {
		Status       vobj.AlertStatus `json:"status"`
		Labels       *vobj.Labels     `json:"labels"`
		Annotations  vobj.Annotations `json:"annotations"`
		StartsAt     *types.Time      `json:"startsAt"`
		EndsAt       *types.Time      `json:"endsAt"`
		GeneratorURL string           `json:"generatorURL"`
		Fingerprint  string           `json:"fingerprint"`
		Value        float64          `json:"value"`
	}

	alertInfo struct {
		Status       string            `json:"status"`
		Labels       map[string]string `json:"labels"`
		Annotations  map[string]string `json:"annotations"`
		StartsAt     string            `json:"startsAt"`
		EndsAt       string            `json:"endsAt"`
		GeneratorURL string            `json:"generatorURL"`
		Fingerprint  string            `json:"fingerprint"`
		Value        float64           `json:"value"`
	}
)

// NewAlertWithAlertStrInfo create alert from alert string
func NewAlertWithAlertStrInfo(info string) (*Alert, error) {
	var a alertInfo
	if err := types.Unmarshal([]byte(info), &a); err != nil {
		return nil, err
	}
	return &Alert{
		Status:       vobj.ToAlertStatus(a.Status),
		Labels:       vobj.NewLabels(a.Labels),
		Annotations:  a.Annotations,
		StartsAt:     types.NewTimeByString(a.StartsAt),
		EndsAt:       types.NewTimeByString(a.EndsAt),
		GeneratorURL: a.GeneratorURL,
		Fingerprint:  a.Fingerprint,
		Value:        a.Value,
	}, nil
}

func (a *Alarm) String() string {
	alarm := alarmInfo{
		Receiver: a.Receiver,
		Status:   a.Status.String(),
		Alerts: types.SliceTo(a.Alerts, func(alert *Alert) *alertInfo {
			return &alertInfo{
				Status:       alert.Status.String(),
				Labels:       alert.Labels.Map(),
				Annotations:  alert.Annotations,
				StartsAt:     alert.StartsAt.String(),
				EndsAt:       alert.EndsAt.String(),
				GeneratorURL: alert.GeneratorURL,
				Fingerprint:  alert.Fingerprint,
				Value:        alert.Value,
			}
		}),
		GroupLabels:       a.GroupLabels.Map(),
		CommonLabels:      a.CommonLabels.Map(),
		CommonAnnotations: a.CommonAnnotations,
		ExternalURL:       a.ExternalURL,
		Version:           a.Version,
		GroupKey:          a.GroupKey,
	}
	bs, _ := types.Marshal(alarm)
	return string(bs)
}

// GetFingerprint gen alert fingerprint
func (a *Alert) GetFingerprint() string {
	fingerprint := a.Fingerprint
	if types.TextIsNull(fingerprint) {
		// 唯一索引+告警时间生成唯一告警指纹
		fingerprint = types.MD5(a.Index() + a.StartsAt.String())
	}
	return fingerprint
}

// GetExternalURL gen alert external url
func (a *Alert) GetExternalURL() string {
	// TODO 生成图表链接
	return a.GeneratorURL
}

func (a *Alert) String() string {
	alert := alertInfo{
		Status:       a.Status.String(),
		Labels:       a.Labels.Map(),
		Annotations:  a.Annotations,
		StartsAt:     a.StartsAt.String(),
		EndsAt:       a.EndsAt.String(),
		GeneratorURL: a.GeneratorURL,
		Fingerprint:  a.Fingerprint,
		Value:        a.Value,
	}
	bs, _ := types.Marshal(alert)
	return string(bs)
}

// Index gen alert index
func (a *Alert) Index() string {
	return "houyi:alert:" + types.MD5(a.Labels.String())
}

// Index gen alarm index
func (a *Alarm) Index() string {
	return "houyi:alarm:" + types.MD5(a.GroupLabels.String())
}

// Message gen alarm message
func (a *Alarm) Message() *watch.Message {
	return watch.NewMessage(a, vobj.TopicAlarm, watch.WithMessageRetryMax(3))
}

// Message gen alert message
func (a *Alert) Message() *watch.Message {
	return watch.NewMessage(a, vobj.TopicAlert, watch.WithMessageRetryMax(3))
}
