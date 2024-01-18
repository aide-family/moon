package strategy

import (
	"sync"
)

var _ AlarmCache = (*defaultAlarmCache)(nil)

type defaultAlarmCache struct {
	alarms sync.Map
}

// NewDefaultAlarmCache 创建默认的告警缓存
func NewDefaultAlarmCache() AlarmCache {
	return &defaultAlarmCache{}
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
