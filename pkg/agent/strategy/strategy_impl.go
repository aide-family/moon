package strategy

import (
	"context"
	"time"

	"github.com/aide-family/moon/pkg"
	"github.com/aide-family/moon/pkg/agent"
	"github.com/aide-family/moon/pkg/util/hash"
	"github.com/go-kratos/kratos/v2/log"
)

var _ Ruler = (*EvalRule)(nil)

func (e *EvalRule) Eval(ctx context.Context) ([]*agent.Alarm, error) {
	if e == nil {
		return nil, nil
	}
	timeNow := time.Now()
	e.SetEventsAt(timeNow)
	value, err := e.datasource.Query(ctx, e.GetExpr(), timeNow.Unix()-int64(10*time.Second.Seconds()))
	if err != nil {
		log.Debugw("rule eval err", err)
		return nil, err
	}
	alarmList := make([]*agent.Alarm, 0, 10)
	firingAlarm := e.NewFiringAlarm(ctx, value)
	if !pkg.IsNil(firingAlarm) {
		alarmList = append(alarmList, firingAlarm)
	}
	resoledAlarm := e.NewResolvedAlarm(ctx, firingAlarm)
	if !pkg.IsNil(resoledAlarm) {
		alarmList = append(alarmList, resoledAlarm)
	}
	//log.Debugw("alarmList", alarmList)
	return alarmList, nil
}

// NewResolvedAlarm 合并告警
func (e *EvalRule) NewResolvedAlarm(_ context.Context, firingAlarm *agent.Alarm) *agent.Alarm {
	cache := agent.GetGlobalCache()
	// 获取存在的告警
	existAlarmInfo := &agent.Alarm{
		Status:            agent.AlarmStatusResolved,
		Alerts:            make([]*agent.Alert, 0),
		GroupLabels:       e.GetLabels(),
		CommonLabels:      e.GetLabels(),
		CommonAnnotations: e.GetAnnotations(),
	}

	if err := cache.Get(e.GetID(), existAlarmInfo); err != nil {
		log.Warnw("mergeAlarm get rule alarm cache err", err)
		return nil
	}

	if pkg.IsNil(existAlarmInfo) || len(existAlarmInfo.GetAlerts()) == 0 {
		// 不存在历史告警数据，直接返回
		return nil
	}

	resolvedAlertList := make([]*agent.Alert, 0, len(existAlarmInfo.GetAlerts()))
	existAlarmInfo.Status = agent.AlarmStatusResolved
	if pkg.IsNil(firingAlarm) || len(firingAlarm.GetAlerts()) == 0 {
		// 全部为告警恢复事件
		for _, existAlert := range existAlarmInfo.GetAlerts() {
			if existAlert.Status == agent.AlarmStatusResolved {
				continue
			}
			startAt, err := time.Parse(time.RFC3339, existAlert.StartsAt)
			if err != nil {
				continue
			}
			if !e.ForEventsAt(startAt) {
				// 是因为还没有达到告警条件
				continue
			}
			resolvedAlert := existAlert
			resolvedAlert.Status = agent.AlarmStatusResolved
			resolvedAlert.EndsAt = e.eventsAt.Format(time.RFC3339)
			resolvedAlertList = append(resolvedAlertList, resolvedAlert)
			// 删除缓存
			if err = cache.Delete(resolvedAlert.GetFingerprint()); err != nil {
				log.Warnw("mergeAlarm delete rule alarm cache err", err)
			}
		}
		if err := cache.Delete(e.GetID()); err != nil {
			log.Warnw("mergeAlarm delete rule alarm cache err", err)
		}
	} else {
		firingAlarmMap := make(map[string]*agent.Alert, len(firingAlarm.GetAlerts()))
		for _, firingAlert := range firingAlarm.GetAlerts() {
			firingAlertTmp := firingAlert
			firingAlarmMap[firingAlertTmp.GetFingerprint()] = firingAlertTmp
		}

		for _, existAlert := range existAlarmInfo.GetAlerts() {
			existAlertTmp := existAlert
			if _, ok := firingAlarmMap[existAlertTmp.GetFingerprint()]; ok {
				continue
			}
			if existAlertTmp.Status == agent.AlarmStatusResolved {
				continue
			}

			existAlertTmp.Status = agent.AlarmStatusResolved
			existAlertTmp.EndsAt = e.eventsAt.Format(time.RFC3339)
			resolvedAlertList = append(resolvedAlertList, existAlertTmp)
			// 删除缓存
			if err := cache.Delete(existAlertTmp.GetFingerprint()); err != nil {
				log.Warnw("mergeAlarm delete rule alarm cache err", err)
			}
		}
		if err := cache.Set(e.GetID(), firingAlarm, 0); err != nil {
			log.Warnw("mergeAlarm set rule alarm cache err", err)
		}
	}
	existAlarmInfo.Alerts = resolvedAlertList

	return existAlarmInfo
}

// NewFiringAlarm 创建告警
func (e *EvalRule) NewFiringAlarm(ctx context.Context, value *agent.QueryResponse) *agent.Alarm {
	alarmInfo := e.initAlarm()
	valueResult := value.GetData().GetResult()
	for _, result := range valueResult {
		alertInfo, fire := e.NewFiringAlert(ctx, result)
		if !fire {
			continue
		}
		alarmInfo.Alerts = append(alarmInfo.Alerts, alertInfo)
	}
	if pkg.IsNil(alarmInfo) || len(alarmInfo.GetAlerts()) == 0 {
		return nil
	}

	cache := agent.GetGlobalCache()

	var historyAlarm agent.Alarm
	alarmStoreInfo := *alarmInfo
	if err := cache.Get(e.GetID(), &historyAlarm); err == nil {
		alertMap := make(map[string]*agent.Alert, len(alarmInfo.GetAlerts()))
		for _, alert := range alarmInfo.GetAlerts() {
			alertMap[alert.GetFingerprint()] = alert
		}

		alertList := make([]*agent.Alert, 0, len(historyAlarm.GetAlerts()))
		// 如果存在则更新
		for _, alert := range historyAlarm.GetAlerts() {
			alertTmp := alert
			if _, ok := alertMap[alertTmp.GetFingerprint()]; ok {
				continue
			}
			alertTmp.EndsAt = e.eventsAt.Format(time.RFC3339)
			alertTmp.Status = agent.AlarmStatusResolved
			alertList = append(alertList, alertTmp)
		}
		alarmStoreInfo.Alerts = append(alarmInfo.Alerts, alertList...)
	}

	if err := cache.Set(e.GetID(), &alarmStoreInfo, 0); err != nil {
		log.Warnw("rule eval alarm cache err", err)
	}

	return alarmInfo
}

// NewFiringAlert 创建警报
func (e *EvalRule) NewFiringAlert(_ context.Context, result *agent.Result) (*agent.Alert, bool) {
	if result == nil {
		log.Debug("result is nil")
		return nil, false
	}
	// 生成告警标识
	fingerprint := hash.MD5(result.GetMetric().String() + ":" + e.GetID())
	cache := agent.GetGlobalCache()
	agentAlertInfo := new(agent.Alert)
	if !cache.Exists(fingerprint) {
		agentAlertInfo = &agent.Alert{
			StartsAt: time.Unix(int64(result.Ts), 0).Format(time.RFC3339),
			Status:   agent.AlarmStatusFiring,
			// TODO 生成平台图标链接
			GeneratorURL: "",
		}
		log.Debugw("new alert", agentAlertInfo)
	} else {
		if err := cache.Get(fingerprint, agentAlertInfo); err != nil {
			log.Warnw("get cache err", err)
			return nil, false
		}
	}

	labels := e.GetLabels().Append(result.GetMetric())
	annotations := e.GetAnnotations()
	metricValue := result.GetValue()

	annotationsVals := map[string]any{
		"value":    metricValue,
		"labels":   labels,
		"eventsAt": agentAlertInfo.StartsAt,
	}

	for k, v := range annotations {
		annotations[k] = agent.Formatter(v, annotationsVals)
	}

	agentAlertInfo.Labels = labels
	agentAlertInfo.Annotations = annotations
	// 拼接开始时间， 保证全局唯一
	agentAlertInfo.Fingerprint = fingerprint

	startAt, err := time.Parse(time.RFC3339, agentAlertInfo.StartsAt)
	if err != nil {
		log.Warnw("parse time err", err)
		return nil, false
	}

	// 缓存具体的告警事件
	if err = cache.Set(fingerprint, agentAlertInfo, 0); err != nil {
		log.Warnw("set cache err", err)
		return nil, false
	}

	if !e.ForEventsAt(startAt) {
		log.Debugw("match", "rule eval alarm not match", "startAt", startAt, "eventsAt", e.eventsAt)
		return nil, false
	}

	return agentAlertInfo, true
}

func (e *EvalRule) initAlarm() *agent.Alarm {
	return &agent.Alarm{
		Status:            agent.AlarmStatusFiring,
		Alerts:            make([]*agent.Alert, 0),
		GroupLabels:       e.GetLabels(),
		CommonLabels:      e.GetAnnotations(),
		CommonAnnotations: e.GetAnnotations(),
	}
}
