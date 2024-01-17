package strategy

import (
	"context"
	"strconv"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"golang.org/x/sync/errgroup"
	"prometheus-manager/pkg/util/times"
)

var _ Alerter = (*Alerting)(nil)

type Alerter interface {
	Eval(ctx context.Context) ([]*Alarm, error)
}

type Alerting struct {
	exitCh         chan struct{}
	log            *log.Helper
	datasourceName DatasourceName
	group          *Group

	alarmCache AlarmCache
}

func (a *Alerting) Eval(ctx context.Context) ([]*Alarm, error) {
	eg := new(errgroup.Group)
	eg.SetLimit(100)
	alarms := NewAlarmList()
	timeNowUnix := time.Now().Unix()
	for _, strategyItem := range a.group.Rules {
		strategyInfo := &*strategyItem
		eg.Go(func() error {
			datasource := NewDatasource(a.datasourceName, strategyInfo.Endpoint())
			queryResponse, err := datasource.Query(ctx, strategyInfo.Expr, timeNowUnix)
			if err != nil {
				return err
			}
			newAlarmInfo := NewAlarm(a.group, strategyInfo, queryResponse.Data.Result)
			// 获取该策略下所有已经产生的告警数据
			existAlarmInfo, exist := a.alarmCache.Get(strategyInfo.Id)
			if !exist {
				// 不存在历史数据, 则直接把新告警数据缓存到alarmCache
				if cacheOK := a.alarmCache.Set(strategyInfo.Id, newAlarmInfo); !cacheOK {
					// 没有缓存成功, 代表已经存在, 或者不需要处理, 则直接返回
					return nil
				}
				alarms.Append(newAlarmInfo)
				return nil
			}

			// 比较两次告警数据, 新数据需要加入alerts, 旧数据需要删除, 并标记为告警恢复
			usableAlarmInfo := a.generateAlarm(strategyInfo, newAlarmInfo, existAlarmInfo)
			alarms.Append(usableAlarmInfo)
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		return nil, err
	}

	return alarms.List(), nil
}

// generateAlarm 根据新告警数据和旧告警数据, 生成告警数据
func (a *Alerting) generateAlarm(ruleInfo *Rule, newAlarmInfo, existAlarmInfo *Alarm) *Alarm {
	if newAlarmInfo == nil || existAlarmInfo == nil {
		return nil
	}

	alarm := &*newAlarmInfo
	for key, value := range existAlarmInfo.GroupLabels {
		alarm.GroupLabels[key] = value
	}
	for key, value := range existAlarmInfo.CommonLabels {
		alarm.CommonLabels[key] = value
	}
	for key, value := range existAlarmInfo.CommonAnnotations {
		alarm.CommonAnnotations[key] = value
	}
	existAlertMap := make(map[string]*Alert)
	for _, alert := range existAlarmInfo.Alerts {
		existAlertMap[alert.Fingerprint] = alert
	}
	// 初始化一个最大值
	alarm.Alerts = make([]*Alert, 0, len(newAlarmInfo.Alerts)+len(existAlarmInfo.Alerts))
	for _, alert := range newAlarmInfo.Alerts {
		alertTmp := &*alert
		// 更新数据, 告警保持为告警状态,
		a.alarmCache.SetAlert(ruleInfo.Id, alertTmp)
		_, ok := existAlertMap[alert.Fingerprint]
		if ok {
			// 如果告警已经存在, 不再继续生成alarm
			continue
		}
		// 如果新告警不存在, 则需要告警
		if alertTmp.Status == AlarmStatusFiring {
			alarm.Alerts = append(alarm.Alerts, alertTmp)
		}
	}
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
			// 删除缓存中的告警数据
			a.alarmCache.RemoveAlert(ruleInfo.Id, alertTmp)
		}
	}

	return alarm
}

// NewAlerting 初始化策略告警实例
func NewAlerting(group *Group, alarmCache AlarmCache, logger *log.Helper) *Alerting {
	a := &Alerting{
		exitCh: make(chan struct{}),
		// TODO 多数据类型时候, 需要调整为可配置
		datasourceName: PrometheusDatasource,
		group:          group,
		log:            logger,
		alarmCache:     alarmCache,
	}

	return a
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
