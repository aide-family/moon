package vobj

import (
	"github.com/aide-family/moon/pkg/plugin/cache"
)

const (
	EmailCacheKey       cache.K = "rabbit:config:email"
	SmsCacheKey         cache.K = "rabbit:config:sms"
	HookCacheKey        cache.K = "rabbit:config:hook"
	SendLockKey         cache.K = "rabbit:send:lock"
	NoticeUserCacheKey  cache.K = "rabbit:config:notice:user"
	NoticeGroupCacheKey cache.K = "rabbit:config:notice:group"
)
