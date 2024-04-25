package strategy

import (
	"context"
	"time"

	"github.com/aide-family/moon/pkg/agent"
	"github.com/aide-family/moon/pkg/util/hash"
	"github.com/go-kratos/kratos/v2/log"
)

var _ Ruler = (*EvalRule)(nil)

func (e *EvalRule) Eval(ctx context.Context) (*agent.Alarm, error) {
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
	firingAlarm := e.NewFiringAlarm(ctx, value)
	return e.mergeAlarm(ctx, firingAlarm), nil
}

// mergeAlarm 合并告警
func (e *EvalRule) mergeAlarm(ctx context.Context, firingAlarm *agent.Alarm) *agent.Alarm {
	// 获取存在的告警
	existAlarmInfo := &agent.Alarm{
		Status:            agent.AlarmStatusResolved,
		Alerts:            make([]*agent.Alert, 0),
		GroupLabels:       e.GetLabels(),
		CommonLabels:      e.GetLabels(),
		CommonAnnotations: e.GetAnnotations(),
	}
	cache := agent.GetGlobalCache().WithContext(ctx)

	if err := cache.Get(e.GetID(), existAlarmInfo); err != nil {
		log.Warnw("mergeAlarm get rule alarm cache err", err)
		return firingAlarm
	}

	firingAlarmMap := make(map[string]*agent.Alert, len(firingAlarm.GetAlerts()))
	for _, firingAlert := range firingAlarm.GetAlerts() {
		firingAlertTmp := firingAlert
		firingAlarmMap[firingAlertTmp.GetFingerprint()] = firingAlertTmp
	}

	resolvedAlerts := make([]*agent.Alert, 0, len(existAlarmInfo.GetAlerts()))
	for _, existAlert := range existAlarmInfo.GetAlerts() {
		existAlertTmp := existAlert
		if _, ok := firingAlarmMap[existAlertTmp.GetFingerprint()]; ok {
			continue
		}
		existAlertTmp.Status = agent.AlarmStatusResolved
		existAlertTmp.EndsAt = e.eventsAt.Format(time.RFC3339)
		resolvedAlerts = append(resolvedAlerts, existAlertTmp)
		// 删除缓存
		if err := cache.Delete(existAlertTmp.GetFingerprint()); err != nil {
			log.Warnw("mergeAlarm delete rule alarm cache err", err)
		}
	}
	if len(firingAlarm.GetAlerts()) == 0 {
		// 删除缓存
		if err := cache.Delete(e.GetID()); err != nil {
			log.Warnw("mergeAlarm delete rule alarm cache err", err)
		}
	}

	existAlarmInfo.Alerts = append(existAlarmInfo.Alerts, firingAlarm.GetAlerts()...)
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
	if len(alarmInfo.GetAlerts()) == 0 {
		return nil
	}

	cache := agent.GetGlobalCache().WithContext(ctx)
	if err := cache.Set(e.GetID(), alarmInfo, 0); err != nil {
		log.Warnw("rule eval alarm cache err", err)
	}

	return alarmInfo
}

// NewFiringAlert 创建警报
func (e *EvalRule) NewFiringAlert(ctx context.Context, result *agent.Result) (*agent.Alert, bool) {
	// 生成告警标识
	fingerprint := hash.MD5(result.GetMetric().String() + ":" + e.GetAlert() + ":" + e.GetLabels().String())
	log.Debugw("fingerprint", fingerprint, "metric", result.GetMetric().String())
	cache := agent.GetGlobalCache().WithContext(ctx)
	agentInfo := agent.Alert{
		StartsAt: e.eventsAt.Format(time.RFC3339),
	}

	if err := cache.Get(fingerprint, &agentInfo); err != nil {
		log.Warnw("NewFiringAlert get rule alarm cache err", err)
	}

	log.Debugw("agentInfo", agentInfo)
	log.Debugw("eventsAt", e.eventsAt.Format(time.RFC3339))
	log.Debugw("startsAt", agentInfo.StartsAt)

	labels := e.GetLabels().Append(result.GetMetric())
	annotations := e.GetAnnotations()
	metricValue := result.GetValue()

	annotationsVals := map[string]any{
		"value":    metricValue,
		"labels":   labels,
		"eventsAt": agentInfo.StartsAt,
	}

	for k, v := range annotations {
		annotations[k] = agent.Formatter(v, annotationsVals)
	}

	agentInfo = agent.Alert{
		Status:      agent.AlarmStatusFiring,
		Labels:      labels,
		Annotations: annotations,
		StartsAt:    agentInfo.StartsAt,
		EndsAt:      "",
		// TODO 生成平台图标链接
		GeneratorURL: "",
		Fingerprint:  fingerprint,
	}

	// 缓存具体的告警事件
	if err := cache.Set(fingerprint, agentInfo, 0); err != nil {
		log.Warnw("set cache err", err)
	}

	startAt, err := time.Parse(time.RFC3339, agentInfo.StartsAt)
	if err != nil {
		log.Warnw("parse time err", err)
		return nil, false
	}
	return &agentInfo, e.ForEventsAt(startAt)
}

func (e *EvalRule) initAlarm() *agent.Alarm {
	return &agent.Alarm{
		Alerts:            make([]*agent.Alert, 0),
		GroupLabels:       e.GetLabels(),
		CommonLabels:      e.GetAnnotations(),
		CommonAnnotations: e.GetAnnotations(),
	}
}
