package health

import (
	"context"

	"github.com/google/wire"

	"github.com/aide-family/moon/pkg/healthcheck"
)

var ProviderSet = wire.NewSet(NewProber)

// Prober 提供 liveness 与 readiness 两个视角：
// liveness 只看进程（永远 ok），readiness 检查 PostgreSQL/Redis 是否可用。
type Prober struct {
	ready healthcheck.Checker
}

func NewProber(ready healthcheck.Checker) *Prober {
	return &Prober{ready: ready}
}

func (prober *Prober) Status() string {
	return "ok"
}

// Ready 供 /readyz 使用。依赖不可用时返回错误，调用方应回 503 让 K8s 摘掉该副本。
func (prober *Prober) Ready(ctx context.Context) error {
	if prober == nil || prober.ready == nil {
		return nil
	}
	return prober.ready.Check(ctx)
}
