package bo

import (
	"github.com/aide-family/moon/pkg/conf"
)

// CacheConfigParams 缓存配置参数
type CacheConfigParams struct {
	Receivers map[string]*conf.Receiver
	Templates map[string]string
}

// CacheConfigKey 缓存配置key
const CacheConfigKey = "rabbit:cache:config"
