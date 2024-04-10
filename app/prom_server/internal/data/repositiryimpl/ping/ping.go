package ping

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/moon/app/prom_server/internal/biz/bo"
	"github.com/aide-family/moon/app/prom_server/internal/biz/repository"
	"github.com/aide-family/moon/app/prom_server/internal/data"
)

var _ repository.PingRepo = (*pingRepoImpl)(nil)

type pingRepoImpl struct {
	repository.UnimplementedPingRepo
	data *data.Data
	log  *log.Helper
}

func (l *pingRepoImpl) Ping(ctx context.Context, g *bo.Ping) (*bo.Ping, error) {
	return &bo.Ping{Hello: "world"}, nil
}

// NewPingRepo .
func NewPingRepo(data *data.Data, logger log.Logger) repository.PingRepo {
	return &pingRepoImpl{
		data: data,
		log:  log.NewHelper(log.With(logger, "module", "data/Ping")),
	}
}
