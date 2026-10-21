package health

import (
	"context"

	"github.com/google/wire"

	"github.com/aide-family/Alioth/pkg/healthcheck"
)

var ProviderSet = wire.NewSet(NewUsecase)

// Usecase 提供 liveness 与 readiness 两个视角：
// liveness 只看进程（永远 ok），readiness 检查 PostgreSQL/Redis 是否可用。
type Usecase struct {
	ready healthcheck.Checker
}

func NewUsecase(ready healthcheck.Checker) *Usecase {
	return &Usecase{ready: ready}
}

func (uc *Usecase) Status() string {
	return "ok"
}

// Ready 供 /readyz 使用。依赖不可用时返回错误，调用方应回 503 让 K8s 摘掉该副本。
func (uc *Usecase) Ready(ctx context.Context) error {
	if uc == nil || uc.ready == nil {
		return nil
	}
	return uc.ready.Check(ctx)
}
