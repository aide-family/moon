package conf

import (
	"github.com/google/wire"
	"sync"
)

// ProviderSet is conf providers.
var ProviderSet = wire.NewSet(
	wire.FieldsOf(new(*Bootstrap), "Data"),
	wire.FieldsOf(new(*Bootstrap), "Trace"),
	wire.FieldsOf(new(*Bootstrap), "Server"),
	wire.FieldsOf(new(*Bootstrap), "Nodes"),
)

// 全局配置获取
var c *Bootstrap

// 保证只执行一次
var once sync.Once

// Get 获取配置
func Get() *Bootstrap {
	return c
}

// Set 设置配置, 单例
func Set(b *Bootstrap) {
	once.Do(func() {
		c = b
	})
}
