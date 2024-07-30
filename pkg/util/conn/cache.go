package conn

import (
	"context"
	"time"
)

// Cache cache通用接口
type Cache interface {
	// Get 获取缓存
	Get(ctx context.Context, key string) (string, error)
	// Set 设置缓存
	Set(ctx context.Context, key string, value string, expiration time.Duration) error
	// Delete 删除缓存
	Delete(ctx context.Context, key string) error
	// Close 关闭缓存
	Close() error
	// Exist 判断缓存是否存在
	Exist(ctx context.Context, key string) bool
}
