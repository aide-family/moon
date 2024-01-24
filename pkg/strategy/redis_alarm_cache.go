package strategy

import (
	"context"
	"encoding/json"

	"github.com/redis/go-redis/v9"
	"prometheus-manager/pkg/helper/consts"
)

var _ AlarmCache = (*redisAlarmCache)(nil)

type redisAlarmCache struct {
	cache *redis.Client
}

func (l *redisAlarmCache) Get(ruleId uint32) (*Alarm, bool) {
	res, err := l.cache.Get(context.Background(), consts.AlarmTmpCache.KeyUint32(ruleId).String()).Result()
	if err != nil {
		return nil, false
	}
	var alarm Alarm
	if err = json.Unmarshal([]byte(res), &alarm); err != nil {
		return nil, false
	}
	return &alarm, true
}

func (l *redisAlarmCache) Set(ruleId uint32, alarm *Alarm) bool {
	alarmBytes, err := json.Marshal(alarm)
	if err != nil {
		return false
	}
	return l.cache.Set(context.Background(), consts.AlarmTmpCache.KeyUint32(ruleId).String(), string(alarmBytes), 0).Err() == nil
}

func (l *redisAlarmCache) Remove(ruleId uint32) bool {
	return l.cache.Del(context.Background(), consts.AlarmTmpCache.KeyUint32(ruleId).String()).Err() == nil
}

func (l *redisAlarmCache) SetNotifyAlert(alert *Alert) bool {
	alertBytes, err := json.Marshal(alert)
	if err != nil {
		return false
	}
	return l.cache.Set(context.Background(), consts.NotifyAlarmCache.Key(alert.Fingerprint).String(), string(alertBytes), 0).Err() == nil
}

func (l *redisAlarmCache) RemoveNotifyAlert(alert *Alert) bool {
	return l.cache.Del(context.Background(), consts.NotifyAlarmCache.Key(alert.Fingerprint).String()).Err() == nil
}

func (l *redisAlarmCache) GetNotifyAlert(alert *Alert) (*Alert, bool) {
	alertStr, err := l.cache.Get(context.Background(), consts.NotifyAlarmCache.Key(alert.Fingerprint).String()).Result()
	if err != nil {
		return nil, false
	}
	var alert2 Alert
	if err = json.Unmarshal([]byte(alertStr), &alert2); err != nil {
		return nil, false
	}
	return &alert2, true
}

func (l *redisAlarmCache) RangeNotifyAlerts(f func(*Alert)) {
	res, err := l.cache.Keys(context.Background(), consts.NotifyAlarmCache.String()).Result()
	if err != nil {
		return
	}
	for _, v := range res {
		alertStr, err := l.cache.Get(context.Background(), v).Result()
		if err != nil {
			continue
		}
		var alert2 Alert
		if err = json.Unmarshal([]byte(alertStr), &alert2); err != nil {
			continue
		}
		f(&alert2)
	}
}

func NewRedisAlarmCache(cache *redis.Client) AlarmCache {
	return &redisAlarmCache{
		cache: cache,
	}
}
