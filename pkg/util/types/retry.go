package types

import (
	"time"
)

// Retry 函数用于实现重试机制
func Retry(f func() error, maxRetries int, maxRetryDuration time.Duration) error {
	var err error
	for i := 0; i < maxRetries; i++ {
		if err = f(); err == nil {
			return nil
		}
		time.Sleep(maxRetryDuration) // 等待一段时间后重试
	}
	return err
}
