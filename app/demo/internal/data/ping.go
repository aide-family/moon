package data

import (
	"context"

	"github.com/aide-family/moon/app/demo/internal/biz/bo"
	"github.com/aide-family/moon/app/demo/internal/biz/repository"

	"github.com/go-kratos/kratos/v2/log"
)

type PingRepo struct {
	data *Data
	log  *log.Helper
}

func (l *PingRepo) Ping(ctx context.Context, g *bo.Ping) (*bo.Ping, error) {
	return &bo.Ping{Hello: "world"}, nil
}

// NewPingRepo .
func NewPingRepo(data *Data, logger log.Logger) repository.PingRepo {
	return &PingRepo{
		data: data,
		log:  log.NewHelper(log.With(logger, "module", "data.Ping")),
	}
}
