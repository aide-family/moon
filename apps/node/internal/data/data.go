package data

import (
	"prometheus-manager/pkg/conn"
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
	NewAlertRepo,
	wire.Bind(new(biz.IAlertRepo), new(*AlertRepo)),
)

var (
	strategies []*strategy.StrategyDir
	loadTime   time.Time
)

const (
	loadModuleName  = "data/load"
	pingModuleName  = "data/ping"
	pullModuleName  = "data/pull"
	pushModuleName  = "data/push"
	alertModuleName = "data/alert"
)

// Data .
type Data struct {
	strategy      *conf.Strategy
	kafkaProducer *conn.KafkaProducer
	kafkaConf     *conf.Kafka
}

// NewData .
func NewData(strategy *conf.Strategy, kafkaConf *conf.Kafka, logger log.Logger) (*Data, func(), error) {
	logHelper := log.NewHelper(log.With(logger, "module", "data"))
	kafkaProducer, err := conn.NewKafkaProducer(kafkaConf.GetEndpoints(), logger)
	if err != nil {
		return nil, nil, err
	}
	cleanup := func() {
		logHelper.Info("closing the data resources")
		kafkaProducer.Producer.Close()
	}
	return &Data{
		strategy:      strategy,
		kafkaProducer: kafkaProducer,
		kafkaConf:     kafkaConf,
	}, cleanup, nil
}
