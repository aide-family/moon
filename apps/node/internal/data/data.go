package data

import (
	"time"

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
	strategies []*strategy.StrategyDir
	loadTime   time.Time
)

const (
	loadModuleName = "data/load"
	pingModuleName = "data/ping"
	pullModuleName = "data/pull"
	pushModuleName = "data/push"
)

// Data .
type Data struct {
	strategy *conf.Strategy
}

// NewData .
func NewData(strategy *conf.Strategy, logger log.Logger) (*Data, func(), error) {
	cleanup := func() {
		log.NewHelper(logger).Info("closing the data resources")
	}
	return &Data{
		strategy: strategy,
	}, cleanup, nil
}
