package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/aide-family/moon/app/demo/internal/biz/bo"
	"github.com/aide-family/moon/app/demo/internal/biz/repository"
)

// PingBiz is a Ping useCase.
type PingBiz struct {
	repo repository.PingRepo
	log  *log.Helper
}

// NewPingBiz new a Ping useCase.
func NewPingBiz(repo repository.PingRepo, logger log.Logger) *PingBiz {
	return &PingBiz{repo: repo, log: log.NewHelper(logger)}
}

// Ping creates a Ping, and returns the new Ping.
func (l *PingBiz) Ping(ctx context.Context, g *bo.Ping) (*bo.Ping, error) {
	l.log.WithContext(ctx).Infof("CreatePing: %v", g.Hello)
	return l.repo.Ping(ctx, g)
}
