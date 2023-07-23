package data

import (
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"prometheus-manager/api/strategy"
	"prometheus-manager/apps/node/internal/biz"
	"prometheus-manager/apps/node/internal/conf"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(
	NewData,
	NewPushRepo,
	wire.Bind(new(biz.IPushRepo), new(*PushRepo)),
	NewPullRepo,
	wire.Bind(new(biz.IPullRepo), new(*PullRepo)),
	NewLoadRepo,
	wire.Bind(new(biz.ILoadRepo), new(*LoadRepo)),
	NewPingRepo,
	wire.Bind(new(biz.IPingRepo), new(*PingRepo)),
)

var (
	strategies []*strategy.Strategy
)

// Data .
type Data struct {
	// TODO wrapped database client
}

// NewData .
func NewData(strategy *conf.Strategy, logger log.Logger) (*Data, func(), error) {
	cleanup := func() {
		log.NewHelper(logger).Info("closing the data resources")
	}
	return &Data{}, cleanup, nil
}
