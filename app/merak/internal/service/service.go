package service

import (
	"context"

	"github.com/google/wire"

	"github.com/aide-family/Alioth/app/merak/internal/biz"
)

var ProviderSet = wire.NewSet(NewHealthService)

type HealthService struct {
	prober *biz.HealthProber
}

func NewHealthService(prober *biz.HealthProber) *HealthService {
	return &HealthService{prober: prober}
}

func (s *HealthService) Status(context.Context) string {
	return s.prober.Status()
}

// Ready 供 /readyz 使用；依赖不可用时返回错误，探针会摘掉该副本。
func (s *HealthService) Ready(ctx context.Context) error {
	return s.prober.Ready(ctx)
}
