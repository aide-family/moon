package goroutine

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/aide-family/moon/pkg/util/after"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport"
	"golang.org/x/sync/errgroup"
)

var _ transport.Server = (*GoRoutine)(nil)

var globalG *GoRoutine
var gOnce sync.Once

// Go 暴露到外界到可控写成方法
var Go = func(f func()) {
	globalG.Go(f)
}

// New 创建一个协程池
//
//	limit: 协程池限制
//	multiple: 缓冲区大小
func New(limit, multiple int) *GoRoutine {
	return &GoRoutine{
		limit:    limit,
		multiple: multiple,
	}
}

type (
	GoRoutine struct {
		lock sync.Mutex
		stop int32
		eg   errgroup.Group
		// 协程池限制, limit为最大任务数量， multiple 为缓冲区大小
		limit, multiple int
		task            chan func()
	}
)

func (g *GoRoutine) Go(f func()) {
	if g.isStop() {
		return
	}

	g.task <- f
}

func (g *GoRoutine) isStop() bool {
	if atomic.LoadInt32(&g.stop) == 1 {
		log.Warnw("method", "GoRoutine.Go", "warn", "协程池已停止", "stop", true)
	}
	return atomic.LoadInt32(&g.stop) == 1
}

func (g *GoRoutine) Start(_ context.Context) error {
	g.lock.Lock()
	defer g.lock.Unlock()
	atomic.StoreInt32(&g.stop, 0)
	g.eg.SetLimit(g.limit)
	g.task = make(chan func(), g.limit*g.multiple)
	gOnce.Do(func() {
		globalG = g
	})

	go func() {
		defer after.RecoverX()
		for f := range g.task {
			g.eg.Go(func() error {
				defer after.RecoverX()
				f()
				return nil
			})
		}
	}()
	return nil
}

func (g *GoRoutine) Stop(_ context.Context) error {
	g.lock.Lock()
	defer g.lock.Unlock()
	atomic.StoreInt32(&g.stop, 1)
	defer func() {
		close(g.task)
	}()
	for len(g.task) > 0 {
	}
	return g.eg.Wait()
}
