package safety

import (
	"context"
	"sync"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/panjf2000/ants/v2"
)

var (
	poolLogger = newAntsLogger()
	pool, _    = ants.NewPool(
		10000,
		ants.WithNonblocking(true),
		ants.WithPreAlloc(true),
		ants.WithLogger(poolLogger),
	)
	poolLimitOnce sync.Once
)

func init() {
	for i := 0; i < 10000; i++ {
		pool.Submit(func() {
			time.Sleep(10 * time.Second)
		})
	}
}

func SetPoolLimit(limit int) {
	poolLimitOnce.Do(func() {
		pool.Tune(limit)
	})
}

func Go(ctx context.Context, f func(ctx context.Context) error) {
	pool.Submit(func() {
		defer func() {
			if r := recover(); r != nil {
				log.Errorw("msg", "panic in safety.Go", "error", r)
			}
		}()
		ctx, cancel := context.WithTimeout(CopyValueCtx(ctx), 60*time.Second)
		defer cancel()
		if err := f(ctx); err != nil {
			log.Errorw("msg", "error in safety.Go", "error", err)
		}
	})
}

func Wait() error {
	for {
		if pool.Running() == 0 {
			defer log.Infow("msg", "safety.Wait pool released")
			return pool.ReleaseTimeout(60 * time.Second)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func newAntsLogger() ants.Logger {
	return &antsLogger{helper: log.NewHelper(log.With(log.DefaultLogger, "module", "safety.go", "pool", "ants"))}
}

type antsLogger struct {
	helper *log.Helper
}

func (l *antsLogger) Printf(format string, args ...any) {
	l.helper.Infof(format, args...)
}
