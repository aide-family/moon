package biz

import (
	"context"

	"github.com/google/wire"

	"github.com/aide-family/Alioth/app/megrez/internal/biz/eval"
	"github.com/aide-family/Alioth/pkg/healthcheck"
)

var ProviderSet = wire.NewSet(NewHealthUsecase, eval.ProviderSet)

// HealthUsecase 提供 liveness 与 readiness 两个视角。
// megrez 不连业务库，readiness 只检查 Redis（评测锁）是否可用。
type HealthUsecase struct {
	ready healthcheck.Checker
}

func NewHealthUsecase(ready healthcheck.Checker) *HealthUsecase {
	return &HealthUsecase{ready: ready}
}

func (uc *HealthUsecase) Status() string {
	return "ok"
}

// Ready 供 /readyz 使用；Redis 不可用时评测锁无法续期，应立即摘掉流量并停止领取任务。
func (uc *HealthUsecase) Ready(ctx context.Context) error {
	if uc == nil || uc.ready == nil {
		return nil
	}
	return uc.ready.Check(ctx)
}
