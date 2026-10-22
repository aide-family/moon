package biz

import (
	"context"

	"github.com/google/wire"

	"github.com/aide-family/Alioth/app/megrez/internal/biz/eval"
	"github.com/aide-family/Alioth/pkg/healthcheck"
)

var ProviderSet = wire.NewSet(NewHealthProber, eval.ProviderSet)

// HealthProber 提供 liveness 与 readiness 两个视角。
// megrez 不连业务库，readiness 只检查 Redis（评测锁）是否可用。
type HealthProber struct {
	ready healthcheck.Checker
}

func NewHealthProber(ready healthcheck.Checker) *HealthProber {
	return &HealthProber{ready: ready}
}

func (prober *HealthProber) Status() string {
	return "ok"
}

// Ready 供 /readyz 使用；Redis 不可用时评测锁无法续期，应立即摘掉流量并停止领取任务。
func (prober *HealthProber) Ready(ctx context.Context) error {
	if prober == nil || prober.ready == nil {
		return nil
	}
	return prober.ready.Check(ctx)
}
