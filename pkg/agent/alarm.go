package agent

import (
	"encoding"
	"encoding/json"
	"sort"
	"strings"

	"golang.org/x/exp/maps"
)

var _ encoding.BinaryUnmarshaler = (*Alarm)(nil)

type (
	Labels      map[string]string
	Annotations map[string]string

	Status string

	Alarm struct {
		// 接收者
		Receiver string `json:"receiver"`
		// 报警状态
		Status Status `json:"status"`
		// 告警列表
		Alerts []*Alert `json:"alerts"`
		// 告警组标签
		GroupLabels Labels `json:"groupLabels"`
		// 公共标签
		CommonLabels map[string]string `json:"commonLabels"`
		// 公共注解
		CommonAnnotations map[string]string `json:"commonAnnotations"`
		// 外部链接
		ExternalURL string `json:"externalURL"`
		// 版本
		Version string `json:"version"`
		// 告警组key
		GroupKey string `json:"groupKey"`
		// 截断告警数
		TruncatedAlerts int32 `json:"truncate"`
	}

	Alert struct {
		// 告警状态
		Status Status `json:"status"`
		// 标签
		Labels Labels `json:"labels"`
		// 注解
		Annotations Annotations `json:"annotations"`
		// 开始时间
		StartsAt string `json:"startsAt"`
		// 结束时间, 如果为空, 则表示未结束
		EndsAt string `json:"endsAt"`
		// 告警生成链接
		GeneratorURL string `json:"generatorURL"`
		// 指纹
		Fingerprint string `json:"fingerprint"`
	}
)

const (
	AlarmStatusFiring   Status = "firing"
	AlarmStatusResolved Status = "resolved"
)

func (a *Alarm) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, a)
}

// Append 追加label
func (l Labels) Append(labels ...map[string]string) Labels {
	if l == nil {
		return l
	}
	for _, label := range labels {
		for k, v := range label {
			l[k] = v
		}
	}
	return l
}

// String 转换为字符串
func (l Labels) String() string {
	if l == nil {
		return ""
	}

	str := strings.Builder{}
	str.WriteString("{")
	keys := maps.Keys(l)
	// 排序
	sort.Strings(keys)
	for _, key := range keys {
		k := key
		v := l[key]
		str.WriteString(k)
		str.WriteString("=")
		str.WriteString(v)
		str.WriteString(",")
	}
	return strings.TrimRight(str.String(), ",") + "}"
}

// GetAlerts 获取告警
func (a *Alarm) GetAlerts() []*Alert {
	if a == nil {
		return []*Alert{}
	}
	return a.Alerts
}

// GetFingerprint 获取指纹
func (a *Alert) GetFingerprint() string {
	if a == nil {
		return ""
	}
	return a.Fingerprint
}

// String Alarm 转换为json字符串
func (a *Alarm) String() string {
	if a == nil {
		return ""
	}
	bs, _ := json.Marshal(a)
	return string(bs)
}

// String Alert 转换为json字符串
func (a *Alert) String() string {
	if a == nil {
		return ""
	}
	bs, _ := json.Marshal(a)
	return string(bs)
}
