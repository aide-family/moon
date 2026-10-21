package biz

import (
	"context"

	"github.com/google/wire"

	"github.com/aide-family/Alioth/app/merak/internal/biz/notify"
	"github.com/aide-family/Alioth/pkg/healthcheck"
)

var ProviderSet = wire.NewSet(NewHealthUsecase, notify.ProviderSet)

// HealthUsecase 提供 liveness 与 readiness 两个视角。
// merak 不连业务库，readiness 只检查 Redis（投递锁）是否可用。
type HealthUsecase struct {
	ready healthcheck.Checker
}

func NewHealthUsecase(ready healthcheck.Checker) *HealthUsecase {
	return &HealthUsecase{ready: ready}
}

func (uc *HealthUsecase) Status() string {
	return "ok"
}

// Ready 供 /readyz 使用；投递锁不可用时继续领任务只会重复发送。
func (uc *HealthUsecase) Ready(ctx context.Context) error {
	if uc == nil || uc.ready == nil {
		return nil
	}
	return uc.ready.Check(ctx)
}
