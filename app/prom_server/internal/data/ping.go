package data

import (
	"context"

	"prometheus-manager/app/prom_server/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
)

type PingRepo struct {
	data *Data
	log  *log.Helper
}

func (l *PingRepo) Ping(ctx context.Context, g *biz.Ping) (*biz.Ping, error) {
	return &biz.Ping{Hello: "world"}, nil
}

// NewPingRepo .
func NewPingRepo(data *Data, logger log.Logger) biz.PingRepo {
	return &PingRepo{
		data: data,
		log:  log.NewHelper(log.With(logger, "module", "data/Ping")),
	}
}
