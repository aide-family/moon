package biz

import (
	"context"

	"github.com/google/wire"

	"github.com/aide-family/Alioth/app/alkaid/internal/biz/agent"
	"github.com/aide-family/Alioth/pkg/healthcheck"
)

var ProviderSet = wire.NewSet(NewHealthUsecase, agent.ProviderSet)

// HealthUsecase 提供 liveness 与 readiness 两个视角。
// alkaid 的“就绪”以能否连上 alioth 为准，由 Connector 上报；这里不额外探测依赖，
// 否则控制面滚动更新时会把所有 Agent 一起摘掉。
type HealthUsecase struct {
	ready healthcheck.Checker
}

func NewHealthUsecase(ready healthcheck.Checker) *HealthUsecase {
	return &HealthUsecase{ready: ready}
}

func (uc *HealthUsecase) Status() string {
	return "ok"
}

func (uc *HealthUsecase) Ready(ctx context.Context) error {
	if uc == nil || uc.ready == nil {
		return nil
	}
	return uc.ready.Check(ctx)
}
