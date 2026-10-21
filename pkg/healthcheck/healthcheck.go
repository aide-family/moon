package healthcheck

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// CheckFunc 探测一个依赖，返回错误表示未就绪。
type CheckFunc func(ctx context.Context) error

// Checker 是 readiness 需要的接口；nil 实现视为永远就绪。
type Checker interface {
	Check(ctx context.Context) error
}

type namedCheck struct {
	name string
	fn   CheckFunc
}

// Checks 按注册顺序并行探测依赖，任一失败即视为未就绪。
// 单个探测有自己的超时，避免 DB/Redis 卡住时把探针一起拖死。
type Checks struct {
	timeout time.Duration
	items   []namedCheck
}

// New 创建探测集合。timeout 是单个依赖的超时，<=0 时取 1 秒。
func New(timeout time.Duration) *Checks {
	if timeout <= 0 {
		timeout = time.Second
	}
	return &Checks{timeout: timeout}
}

// Add 注册一个依赖探测；nil 函数会被忽略。
func (c *Checks) Add(name string, fn CheckFunc) {
	if c == nil || fn == nil {
		return
	}
	c.items = append(c.items, namedCheck{name: name, fn: fn})
}

// Check 并行探测所有依赖，返回第一个失败项（带依赖名）。
func (c *Checks) Check(ctx context.Context) error {
	if c == nil || len(c.items) == 0 {
		return nil
	}
	var (
		wg    sync.WaitGroup
		mu    sync.Mutex
		first error
	)
	for _, item := range c.items {
		wg.Add(1)
		go func(item namedCheck) {
			defer wg.Done()
			checkCtx, cancel := context.WithTimeout(ctx, c.timeout)
			defer cancel()
			if err := item.fn(checkCtx); err != nil {
				mu.Lock()
				if first == nil {
					first = fmt.Errorf("%s: %w", item.name, err)
				}
				mu.Unlock()
			}
		}(item)
	}
	wg.Wait()
	return first
}
