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
		Receiver          string            `json:"receiver"`
		Status            vobj.AlertStatus  `json:"status"`
		Alerts            []*Alert          `json:"alerts"`
		GroupLabels       *vobj.Labels      `json:"groupLabels"`
		CommonLabels      *vobj.Labels      `json:"commonLabels"`
		CommonAnnotations *vobj.Annotations `json:"commonAnnotations"`
		ExternalURL       string            `json:"externalURL"`
		Version           string            `json:"version"`
		GroupKey          string            `json:"groupKey"`
		TruncatedAlerts   int32             `json:"truncatedAlerts"`
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

// MarshalBinary 将告警信息编码为二进制
func (a *Alert) MarshalBinary() (data []byte, err error) {
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
	return types.Marshal(alert)
}

// UnmarshalBinary 将告警信息解码为告警结构体
func (a *Alert) UnmarshalBinary(data []byte) error {
	var alert alertInfo
	if err := types.Unmarshal(data, &alert); err != nil {
		return err
	}
	a.Status = vobj.ToAlertStatus(alert.Status)
	a.Labels = vobj.NewLabels(alert.Labels)
	a.Annotations = alert.Annotations
	a.StartsAt = types.NewTimeByString(alert.StartsAt)
	a.EndsAt = types.NewTimeByString(alert.EndsAt)
	a.GeneratorURL = alert.GeneratorURL
	a.Fingerprint = alert.Fingerprint
	a.Value = alert.Value
	return nil
}

// NewAlertWithAlertStrInfo 从告警字符串信息创建告警结构体
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

// String 将告警信息转换为字符串
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
		CommonAnnotations: a.CommonAnnotations.Map(),
		ExternalURL:       a.ExternalURL,
		Version:           a.Version,
		GroupKey:          a.GroupKey,
	}
	bs, _ := types.Marshal(alarm)
	return string(bs)
}

// GetFingerprint 生成告警指纹
func (a *Alert) GetFingerprint() string {
	fingerprint := a.Fingerprint
	if types.TextIsNull(fingerprint) {
		// 唯一索引+告警时间生成唯一告警指纹
		fingerprint = types.MD5(types.TextJoin(a.Index(), a.StartsAt.String()))
	}
	return fingerprint
}

// GetExternalURL 生成告警外部链接
func (a *Alert) GetExternalURL() string {
	// TODO 生成图表链接
	return a.GeneratorURL
}

func (a *Alert) String() string {
	bs, _ := a.MarshalBinary()
	return string(bs)
}

// Index 生成告警索引
func (a *Alert) Index() string {
	return types.TextJoin("houyi:alert:", types.MD5(a.Labels.String()))
}

// Index 生成告警索引
func (a *Alarm) Index() string {
	return types.TextJoin("houyi:alarm:", types.MD5(a.GroupLabels.String()))
}

// Message 生成告警消息
func (a *Alarm) Message() *watch.Message {
	return watch.NewMessage(a, vobj.TopicAlarm, watch.WithMessageRetryMax(10))
}

// Message 生成告警消息
func (a *Alert) Message() *watch.Message {
	return watch.NewMessage(a, vobj.TopicAlert, watch.WithMessageRetryMax(10))
}

// PushedFlag 生成推送标识
func (a *Alert) PushedFlag() string {
	return types.TextJoin("houyi:alert:pushed:", a.Status.String(), ":", a.GetFingerprint())
}
