package bo

import (
	"github.com/aide-cloud/moon/api"
)

type CacheConfigParams struct {
	Receivers map[string]*api.Receiver
	Templates map[string]string
}

const CacheConfigKey = "rabbit:cache:config"
