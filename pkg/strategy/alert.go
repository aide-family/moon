package strategy

import (
	"context"
	"strconv"
	"time"

	"golang.org/x/sync/errgroup"
	"prometheus-manager/pkg/util/times"
)

var _ Alerter = (*Alerting)(nil)

type Alerter interface {
	Eval(ctx context.Context) ([]*Alarm, error)
}

type (
	Alerting struct {
		datasourceName DatasourceName
		groups         []*Group

		alarmCache AlarmCache
	}

	AlertingOption func(*Alerting)
)

// NewAlerting 初始化策略告警实例
func NewAlerting(groups ...*Group) Alerter {
	a := &Alerting{
		datasourceName: PrometheusDatasource,
		groups:         groups,
		alarmCache:     NewDefaultAlarmCache(),
	}

	return a
}

func (a *Alerting) Eval(ctx context.Context) ([]*Alarm, error) {
	eg := new(errgroup.Group)
	eg.SetLimit(100)
	alarms := NewAlarmList()
	timeNowUnix := time.Now().Unix()
	for _, groupItem := range a.groups {
		group := &*groupItem
		for _, strategyItem := range group.Rules {
			strategyInfo := &*strategyItem
			eg.Go(func() error {
				datasource := NewDatasource(a.datasourceName, strategyInfo.Endpoint())
				queryResponse, err := datasource.Query(ctx, strategyInfo.Expr, timeNowUnix)
				if err != nil {
					return err
				}
				newAlarmInfo := NewAlarm(group, strategyInfo, queryResponse.Data.Result)
				// 获取该策略下所有已经产生的告警数据
				existAlarmInfo, exist := a.alarmCache.Get(strategyInfo.Id)
				if !exist {
					// 不存在历史数据, 则直接把新告警数据缓存到alarmCache
					a.alarmCache.Set(strategyInfo.Id, newAlarmInfo)
					alarms.Append(newAlarmInfo)
					return nil
				}

				// 比较两次告警数据, 新数据需要加入alerts, 旧数据需要删除, 并标记为告警恢复
				usableAlarmInfo := a.mergeAlarm(strategyInfo, newAlarmInfo, existAlarmInfo)
				alarms.Append(usableAlarmInfo)
				return nil
			})
		}
	}

	if err := eg.Wait(); err != nil {
		return nil, err
	}

	return alarms.List(), nil
}

// mergeAlarm 根据新告警数据和旧告警数据, 生成告警数据
func (a *Alerting) mergeAlarm(ruleInfo *Rule, newAlarmInfo, existAlarmInfo *Alarm) *Alarm {
	if newAlarmInfo == nil || existAlarmInfo == nil {
		return nil
	}

	alarm := &*newAlarmInfo
	existAlertMap := make(map[string]*Alert)
	for _, alert := range existAlarmInfo.Alerts {
		existAlertMap[alert.Fingerprint] = alert
	}
	// 初始化一个最大值
	alarm.Alerts = make([]*Alert, 0, len(newAlarmInfo.Alerts)+len(existAlarmInfo.Alerts))
	// 把新告警set到缓存中
	alarm.Alerts = append(alarm.Alerts, newAlarmInfo.Alerts...)
	a.alarmCache.Set(ruleInfo.Id, alarm)

	newAlertMap := make(map[string]*Alert)
	for _, alert := range newAlarmInfo.Alerts {
		alertTmp := &*alert
		newAlertMap[alert.Fingerprint] = alertTmp
	}
	timeUnix := time.Now().Unix()
	endsAt := time.Unix(timeUnix, 0).Format(times.ParseLayout)
	for _, alert := range existAlarmInfo.Alerts {
		if alertTmp, ok := newAlertMap[alert.Fingerprint]; !ok {
			alertTmp = &*alert
			// 如果告警不存在, 则告警已经恢复, 告警恢复
			alertTmp.Status = AlarmStatusResolved
			alertTmp.EndsAt = endsAt
			alarm.Alerts = append(alarm.Alerts, alertTmp)
		}
	}

	return alarm
}

// BuildDuration 字符串转为api时间
func BuildDuration(duration string) time.Duration {
	durationLen := len(duration)
	if duration == "" || durationLen < 2 {
		return 0
	}
	value, _ := strconv.Atoi(duration[:durationLen-1])
	// 获取字符串最后一个字符
	unit := string(duration[durationLen-1])
	switch unit {
	case "s":
		return time.Duration(value) * time.Second
	case "m":
		return time.Duration(value) * time.Minute
	case "h":
		return time.Duration(value) * time.Hour
	case "d":
		return time.Duration(value) * time.Hour * 24
	default:
		return time.Duration(value) * time.Second
	}
}

// WithDatasourceName 设置数据源名称
func WithDatasourceName(datasourceName DatasourceName) AlertingOption {
	return func(a *Alerting) {
		a.datasourceName = datasourceName
	}
}

// WithAlarmCache 设置告警缓存
func WithAlarmCache(alarmCache AlarmCache) AlertingOption {
	return func(a *Alerting) {
		a.alarmCache = alarmCache
	}
}
