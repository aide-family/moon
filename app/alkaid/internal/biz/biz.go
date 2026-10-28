package biz

import (
	"context"

	"github.com/google/wire"

	"github.com/aide-family/moon/app/alkaid/internal/biz/agent"
	"github.com/aide-family/moon/pkg/healthcheck"
)

var ProviderSet = wire.NewSet(NewHealthProber, agent.ProviderSet)

// HealthProber 提供 liveness 与 readiness 两个视角。
// alkaid 的“就绪”以能否连上 alioth 为准，由 Connector 上报；这里不额外探测依赖，
// 否则控制面滚动更新时会把所有 Agent 一起摘掉。
type HealthProber struct {
	ready healthcheck.Checker
}

func NewHealthProber(ready healthcheck.Checker) *HealthProber {
	return &HealthProber{ready: ready}
}

func (prober *HealthProber) Status() string {
	return "ok"
}

func (prober *HealthProber) Ready(ctx context.Context) error {
	if prober == nil || prober.ready == nil {
		return nil
	}
	return prober.ready.Check(ctx)
}
