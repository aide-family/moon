package conn

import (
	"github.com/alicebob/miniredis/v2"
)

// NewMiniRedis 创建一个内存中的redis服务
func NewMiniRedis() (*miniredis.Miniredis, error) {
	return miniredis.Run()
}
