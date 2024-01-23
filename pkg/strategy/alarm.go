package strategy

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"text/template"
	"time"

	"prometheus-manager/pkg/after"
	"prometheus-manager/pkg/util/hash"
	"prometheus-manager/pkg/util/times"
)

type (
	AlarmStatus string

	Alarm struct {
		// 接收者
		Receiver string `json:"receiver"`
		// 报警状态
		Status AlarmStatus `json:"status"`
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
		Status AlarmStatus `json:"status"`
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

	Alarms struct {
		Alarms []*Alarm
		lock   sync.RWMutex
	}

	AlarmCache interface {
		// Get 获取当前规则下所有告警数据
		Get(ruleId uint32) (*Alarm, bool)
		// Set 当前规则下告警数据
		Set(ruleId uint32, alarm *Alarm) bool
		// Remove 删除告警
		Remove(ruleId uint32) bool
		// SetNotifyAlert 当前规则下告警数据
		SetNotifyAlert(alert *Alert) bool
		// RemoveNotifyAlert 删除告警
		RemoveNotifyAlert(alert *Alert) bool
		// GetNotifyAlert 获取告警数据
		GetNotifyAlert(alert *Alert) (*Alert, bool)
		// RangeNotifyAlerts range notify alerts
		RangeNotifyAlerts(f func(*Alert))
	}
)

const (
	// AlarmStatusFiring firing
	AlarmStatusFiring AlarmStatus = "firing"
	// AlarmStatusResolved resolved
	AlarmStatusResolved AlarmStatus = "resolved"
)

// Append append alarm
func (l *Alarms) Append(alarm *Alarm) {
	if alarm == nil {
		return
	}
	l.lock.Lock()
	defer l.lock.Unlock()
	l.Alarms = append(l.Alarms, alarm)
}

// List  alarm list
func (l *Alarms) List() []*Alarm {
	l.lock.RLock()
	defer l.lock.RUnlock()
	list := make([]*Alarm, 0, len(l.Alarms))
	// 处理alert指纹
	for _, alarm := range l.Alarms {
		if alarm == nil {
			continue
		}
		alarmTmp := *alarm
		alerts := make([]*Alert, 0, len(alarm.Alerts))
		for _, alert := range alarm.Alerts {
			alertTmp := *alert
			alertTmp.Fingerprint = hash.MD5(alertTmp.Fingerprint + hash.MD5(alertTmp.StartsAt))
			alerts = append(alerts, &alertTmp)
		}
		alarmTmp.Alerts = alerts
		list = append(list, &alarmTmp)
	}

	return list
}

func NewAlarmList(alarms ...*Alarm) *Alarms {
	return &Alarms{
		Alarms: alarms,
	}
}

func NewAlarm(group *Group, rule *Rule, results []*Result) *Alarm {
	alarmInfo := &Alarm{
		Receiver: group.Name,
		Status:   AlarmStatusFiring,
		Alerts:   make([]*Alert, 0, len(results)),
		GroupLabels: map[string]string{
			MetricGroupName: group.Name,
			MetricGroupId:   strconv.Itoa(int(group.Id)),
			MetricAlert:     rule.Alert,
			MetricAlertId:   strconv.Itoa(int(rule.Id)),
		},
		// 公共标签
		CommonLabels: rule.Labels,
		// 公共注解
		CommonAnnotations: rule.Annotations,
		// TODO 生成前端可用链接
		ExternalURL: "",
		// TODO 显示正确的系统版本
		Version:  "",
		GroupKey: fmt.Sprintf("%s:%s", MetricGroupName, group.Name),
		// TODO 后面再考虑增加截断告警数
		TruncatedAlerts: 0,
	}

	allLabels := make(map[string]string)
	allLabels[MetricAlertId] = strconv.Itoa(int(rule.Id))
	for _, result := range results {
		for key, value := range result.Metric.Map() {
			allLabels[key] = value
		}
	}
	for key, value := range alarmInfo.GroupLabels {
		allLabels[key] = value
	}
	for key, value := range alarmInfo.CommonLabels {
		allLabels[key] = value
	}
	for key, value := range rule.Labels {
		allLabels[key] = value
	}

	// 获取该策略下所有已经产生的告警数据
	existAlarmInfo, exist := alarmCache.Get(rule.Id)
	existAlertMap := make(map[string]*Alert)
	if exist {
		for _, alert := range existAlarmInfo.Alerts {
			existAlertMap[alert.Fingerprint] = alert
		}
	}

	for _, result := range results {
		if len(result.Value) != 2 {
			continue
		}

		timeUnix := result.Value[0].(float64)
		metricValue := result.Value[1].(string)
		annotations := make(Annotations)
		for key, value := range rule.Annotations {
			formatStr := Formatter(value, map[string]any{
				"value":  metricValue,
				"labels": allLabels,
			})
			annotations[key] = formatStr
		}

		alert := &Alert{
			Status:       AlarmStatusFiring,
			Labels:       allLabels,
			Annotations:  annotations,
			GeneratorURL: "",
			Fingerprint:  result.Metric.MD5(),
		}
		if existAlert, ok := existAlertMap[alert.Fingerprint]; ok {
			alert.StartsAt = existAlert.StartsAt
		} else {
			alert.StartsAt = time.Unix(int64(timeUnix), 0).Format(times.ParseLayout)
		}
		alarmInfo.Alerts = append(alarmInfo.Alerts, alert)
	}

	return alarmInfo
}

// ReplaceString 替换字符串中的$为.
//
//		eg: {{ $labels.instance }} 的值大于 {{ $value }} {{ .labels.instance }} 的值大于 {{ .value }}
//	 如果{{}}中间存在$符号, 则替换成.符号
func replaceString(str string) (s string) {
	if str == "" {
		return ""
	}

	// 正则表达式匹配 {{ $... }} 形式的子串
	r := regexp.MustCompile(`\{\{\s*\$(.*?)\s*\}\}`)

	// 使用 ReplaceAllStringFunc 函数替换匹配到的内容
	s = r.ReplaceAllStringFunc(str, func(match string) string {
		// 去掉 {{ 和 }} 符号，保留内部的变量名并替换 $
		variable := strings.TrimSuffix(strings.TrimPrefix(match, "{{"), "}}")
		return fmt.Sprintf("{{ %s }}", strings.Replace(variable, "$", ".", 1))
	})

	return s
}

// Formatter 格式化告警文案
func Formatter(format string, data map[string]any) (s string) {
	formatStr := replaceString(format)
	if formatStr == "" || data == nil || len(data) == 0 {
		return ""
	}

	defer after.RecoverX()
	// 创建一个模板对象，定义模板字符串
	t, err := template.New("alert").Parse(formatStr)
	if err != nil {
		return format
	}
	tmpl := template.Must(t, err)

	// 执行模板并填充数据
	resultIoWriter := new(strings.Builder)

	if err := tmpl.Execute(resultIoWriter, data); err != nil {
		return format
	}
	return resultIoWriter.String()
}

// Bytes Alarm to bytes
func (a *Alarm) Bytes() []byte {
	bs, _ := json.Marshal(a)
	return bs
}
