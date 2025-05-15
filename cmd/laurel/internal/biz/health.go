package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/moon/cmd/laurel/internal/biz/repository"
)

func NewHealthBiz(cacheRepo repository.Cache, pingRepo repository.Health, logger log.Logger) *HealthBiz {
	return &HealthBiz{
		cacheRepo: cacheRepo,
		pingRepo:  pingRepo,
		helper:    log.NewHelper(log.With(logger, "module", "biz.health")),
	}
}

type HealthBiz struct {
	cacheRepo repository.Cache
	pingRepo  repository.Health

	helper *log.Helper
}

func (b *HealthBiz) Check(ctx context.Context) error {
	if err := b.pingRepo.PingCache(ctx); err != nil {
		return err
	}
	return nil
}
