package types

import (
	"context"
	"time"
)

var _ context.Context = (*valueOnlyContext)(nil)

// valueOnlyContext 包装一个只有value的context
type valueOnlyContext struct{ context.Context }

func (valueOnlyContext) Deadline() (deadline time.Time, ok bool) { return }
func (valueOnlyContext) Done() <-chan struct{}                   { return nil }
func (valueOnlyContext) Err() error                              { return nil }

// CopyValueCtx 复制一个只有value的context
func CopyValueCtx(ctx context.Context) context.Context {
	return valueOnlyContext{ctx}
}
