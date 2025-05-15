package impl

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/moon-monitor/moon/cmd/rabbit/internal/biz/repository"
	"github.com/moon-monitor/moon/cmd/rabbit/internal/data"
)

func NewPingRepo(data *data.Data, logger log.Logger) repository.Health {
	return &pingRepo{
		Data:   data,
		helper: log.NewHelper(log.With(logger, "module", "data.repo.ping")),
	}
}

type pingRepo struct {
	*data.Data

	helper *log.Helper
}

func (r *pingRepo) PingCache(ctx context.Context) error {
	return r.GetCache().Client().Ping(ctx).Err()
}
