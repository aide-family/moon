package bo

import (
	"github.com/aide-family/moon/pkg/conf"
	"github.com/aide-family/moon/pkg/util/types"
)

// CacheConfigParams 缓存配置参数
type CacheConfigParams struct {
	Receivers map[string]*conf.Receiver `json:"receivers"`
	Templates map[string]string         `json:"templates"`
}

// String 字符串化
func (c *CacheConfigParams) String() string {
	bs, _ := types.Marshal(c)
	return string(bs)
}
