package safety

import (
	"context"
	"time"

	klog "github.com/go-kratos/kratos/v2/log"
)

func Go(ctx context.Context, name string, f func(context.Context) error) {
	helper := klog.NewHelper(klog.With(klog.GetLogger(), "func", name))

	start := time.Now()

	go func() {
		defer func() {
			if r := recover(); r != nil {
				helper.Errorf("panic: %v", r)
			}
			helper.Debugw("completed, cost: %v", time.Since(start))
		}()

		if err := f(ctx); err != nil {
			helper.Errorf("run error: %v", err)
		}
	}()
}
