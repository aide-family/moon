package consts

import (
	"strconv"
	"strings"
)

type RedisKey string

const (
	// PromGroupDeleteKey 删除规则, 用于记录删除的ID列表数据
	PromGroupDeleteKey RedisKey = "prom:group:delete"
	// PromGroupChangeKey 更新规则, 用于记录更新的ID列表数据
	PromGroupChangeKey RedisKey = "prom:group:change"
	// UserRoleKey 用户角色缓存
	UserRoleKey RedisKey = "user:role"
	// AuthCaptchaKey 验证码缓存
	AuthCaptchaKey RedisKey = "auth:captcha"
	// UserLogoutKey 用户退出缓存
	UserLogoutKey RedisKey = "user:logout"
	// UserRolesKey 用户角色缓存
	UserRolesKey RedisKey = "user:roles:rbac"
	// APICacheKey 接口缓存
	APICacheKey RedisKey = "api:cache"
	// RoleDisabledKey 角色禁用列表缓存
	RoleDisabledKey RedisKey = "role:disabled:hash"
	// AlarmRealtimeCacheById 告警实时数据缓存
	AlarmRealtimeCacheById RedisKey = "alarm:realtime:id"
	// AlarmTmpCache 告警临时缓存
	AlarmTmpCache RedisKey = "alarm:realtime:tmp"
	// NotifyAlarmCache 告警缓存
	NotifyAlarmCache RedisKey = "alarm:realtime:notify"
	// AgentNames agent名称缓存
	AgentNames RedisKey = "agent:names"
	// StrategyGroups 策略组缓存
	StrategyGroups RedisKey = "strategy:groups"
	// ChangeGroupIds 更新规则, 用于记录更新的ID列表数据
	ChangeGroupIds RedisKey = "prom:group:change:ids"
	// AlarmNotifyCache 告警通知缓存
	AlarmNotifyCache RedisKey = "alarm:notify:alert"
)

func (r RedisKey) String() string {
	return string(r)
}

func (r RedisKey) Key(args ...string) RedisKey {
	return RedisKey(strings.Join(append([]string{r.String()}, args...), ":"))
}

func (r RedisKey) KeyUint32(args ...uint32) RedisKey {
	var s []string
	for _, v := range args {
		s = append(s, strconv.FormatInt(int64(v), 10))
	}
	return r.Key(s...)
}

func (r RedisKey) KeyInt(args ...uint) RedisKey {
	var s []string
	for _, v := range args {
		s = append(s, strconv.FormatInt(int64(v), 10))
	}
	return r.Key(s...)
}
