package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
)

// Ping is a Ping model.
type Ping struct {
	Hello string
}

// PingRepo is a Greater repo.
type PingRepo interface {
	Ping(ctx context.Context, g *Ping) (*Ping, error)
}

// PingUseCase is a Ping useCase.
type PingUseCase struct {
	repo PingRepo
	log  *log.Helper
}

// NewPingUseCase new a Ping useCase.
func NewPingUseCase(repo PingRepo, logger log.Logger) *PingUseCase {
	return &PingUseCase{repo: repo, log: log.NewHelper(logger)}
}

// Ping creates a Ping, and returns the new Ping.
func (l *PingUseCase) Ping(ctx context.Context, g *Ping) (*Ping, error) {
	l.log.WithContext(ctx).Infof("CreatePing: %v", g.Hello)
	return l.repo.Ping(ctx, g)
}
