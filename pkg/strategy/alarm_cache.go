package strategy

import (
	"context"
	"encoding/json"

	"github.com/aide-family/moon/pkg/helper/consts"
	"github.com/aide-family/moon/pkg/util/cache"
)

var _ AlarmCache = (*globalAlarmCache)(nil)

type globalAlarmCache struct {
	cache cache.GlobalCache
}

func (l *globalAlarmCache) Get(ruleId uint32) (*Alarm, bool) {
	res, err := l.cache.Get(context.Background(), consts.AlarmTmpCache.KeyUint32(ruleId).String())
	if err != nil {
		return nil, false
	}
	var alarm Alarm
	if err = json.Unmarshal(res, &alarm); err != nil {
		return nil, false
	}
	return &alarm, true
}

func (l *globalAlarmCache) Set(ruleId uint32, alarm *Alarm) bool {
	alarmBytes, err := json.Marshal(alarm)
	if err != nil {
		return false
	}
	return l.cache.Set(context.Background(), consts.AlarmTmpCache.KeyUint32(ruleId).String(), alarmBytes, 0) == nil
}

func (l *globalAlarmCache) Remove(ruleId uint32) bool {
	return l.cache.Del(context.Background(), consts.AlarmTmpCache.KeyUint32(ruleId).String()) == nil
}

func (l *globalAlarmCache) SetNotifyAlert(alert *Alert) bool {
	alertBytes, err := json.Marshal(alert)
	if err != nil {
		return false
	}
	args := [][]byte{[]byte(alert.Fingerprint), alertBytes}
	return l.cache.HSet(context.Background(), consts.NotifyAlarmCache.String(), args...) == nil
}

func (l *globalAlarmCache) RemoveNotifyAlert(alert *Alert) bool {
	return l.cache.HDel(context.Background(), consts.NotifyAlarmCache.String(), alert.Fingerprint) == nil
}

func (l *globalAlarmCache) GetNotifyAlert(alert *Alert) (*Alert, bool) {
	alertStr, err := l.cache.HGet(context.Background(), consts.NotifyAlarmCache.String(), alert.Fingerprint)
	if err != nil {
		return nil, false
	}
	var alert2 Alert
	if err = json.Unmarshal(alertStr, &alert2); err != nil {
		return nil, false
	}
	return &alert2, true
}

func (l *globalAlarmCache) RangeNotifyAlerts(f func(*Alert)) {
	alerts, err := l.cache.HGetAll(context.Background(), consts.NotifyAlarmCache.String())
	if err != nil {
		return
	}
	for _, alert := range alerts {
		var alert2 Alert
		if err = json.Unmarshal(alert, &alert2); err != nil {
			break
		}
		f(&alert2)
	}
}

func NewAlarmCache(cache cache.GlobalCache) AlarmCache {
	return &globalAlarmCache{
		cache: cache,
	}
}
