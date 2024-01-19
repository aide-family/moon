package strategy

import (
	"sync"
)

var _ AlarmCache = (*defaultAlarmCache)(nil)

type defaultAlarmCache struct {
	alarms      *sync.Map
	notifyAlarm *sync.Map
}

// NewDefaultAlarmCache 创建默认的告警缓存
func NewDefaultAlarmCache() AlarmCache {
	return &defaultAlarmCache{
		alarms:      new(sync.Map),
		notifyAlarm: new(sync.Map),
	}
}

func (l *defaultAlarmCache) RangeNotifyAlerts(f func(*Alert)) {
	l.notifyAlarm.Range(func(key, value any) bool {
		alertItem, ok := value.(*Alert)
		if ok {
			f(alertItem)
		}
		return true
	})
}

func (l *defaultAlarmCache) Remove(ruleId uint32) bool {
	if _, exist := l.alarms.Load(ruleId); exist {
		l.alarms.Delete(ruleId)
		return true
	}
	return false
}

func (l *defaultAlarmCache) Get(ruleId uint32) (*Alarm, bool) {
	v, exist := l.alarms.Load(ruleId)
	if exist {
		if alarm, ok := v.(*Alarm); ok {
			return alarm, ok
		}
	}
	return nil, false
}

func (l *defaultAlarmCache) Set(ruleId uint32, alarm *Alarm) bool {
	l.alarms.Store(ruleId, alarm)
	return true
}

func (l *defaultAlarmCache) SetNotifyAlert(alert *Alert) bool {
	l.notifyAlarm.Store(alert.Fingerprint, alert)
	return true
}

func (l *defaultAlarmCache) RemoveNotifyAlert(alert *Alert) bool {
	if _, ok := l.notifyAlarm.Load(alert.Fingerprint); ok {
		l.notifyAlarm.Delete(alert.Fingerprint)
		return true
	}
	return false
}

func (l *defaultAlarmCache) GetNotifyAlert(alert *Alert) (*Alert, bool) {
	if v, exist := l.notifyAlarm.Load(alert.Fingerprint); exist {
		if alarm, ok := v.(*Alert); ok {
			return alarm, ok
		}
	}
	return nil, false
}
