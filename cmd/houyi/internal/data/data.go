package data

import (
	"github.com/aide-family/moon/pkg/plugin/server"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"

	"github.com/aide-family/moon/cmd/houyi/internal/biz/bo"
	"github.com/aide-family/moon/cmd/houyi/internal/conf"
	"github.com/aide-family/moon/pkg/plugin/cache"
	"github.com/aide-family/moon/pkg/plugin/datasource"
	"github.com/aide-family/moon/pkg/util/safety"
)

// ProviderSetData is a set of data providers.
var ProviderSetData = wire.NewSet(New)

func New(c *conf.Bootstrap, logger log.Logger) (*Data, func(), error) {
	var err error
	dataConf := c.GetData()
	eventBusConf := c.GetEventBus()
	palaceConfig := c.GetPalace()
	initConfig := &server.InitConfig{
		MicroConfig: palaceConfig,
		Registry:    c.GetRegistry(),
	}
	palaceServer, err := InitClient(initConfig)
	if err != nil {
		return nil, nil, err
	}
	cacheServer, err := cache.NewCache(c.GetCache())
	if err != nil {
		return nil, nil, err
	}

	data := &Data{
		dataConf:            dataConf,
		cache:               cacheServer,
		helper:              log.NewHelper(log.With(logger, "module", "data")),
		palaceServer:        palaceServer,
		metricDatasource:    safety.NewMap[string, datasource.Metric](),
		StrategyJobEventBus: make(chan bo.StrategyJob, eventBusConf.GetStrategyJobEventBusMaxCap()),
		AlertJobEventBus:    make(chan bo.AlertJob, eventBusConf.GetAlertEventJobBusMaxCap()),
		AlertEventBus:       make(chan bo.Alert, eventBusConf.GetAlertEventBusMaxCap()),
	}

	cleanup := func() {
		log.NewHelper(logger).Info("closing the data resources")
		if err = data.cache.Close(); err != nil {
			log.NewHelper(logger).Errorw("method", "close cache", "err", err)
		}
		close(data.StrategyJobEventBus)
	}
	return data, cleanup, nil
}

type Data struct {
	dataConf            *conf.Data
	cache               cache.Cache
	helper              *log.Helper
	palaceServer        *Server
	metricDatasource    *safety.Map[string, datasource.Metric]
	StrategyJobEventBus chan bo.StrategyJob
	AlertJobEventBus    chan bo.AlertJob
	AlertEventBus       chan bo.Alert
}

func (d *Data) GetPlaceServer() *Server {
	return d.palaceServer
}

func (d *Data) GetCache() cache.Cache {
	return d.cache
}

func (d *Data) GetMetricDatasource(id string) (datasource.Metric, bool) {
	return d.metricDatasource.Get(id)
}

func (d *Data) SetMetricDatasource(id string, metric datasource.Metric) {
	d.metricDatasource.Set(id, metric)
}

func (d *Data) InStrategyJobEventBus() chan<- bo.StrategyJob {
	return d.StrategyJobEventBus
}

func (d *Data) OutStrategyJobEventBus() <-chan bo.StrategyJob {
	return d.StrategyJobEventBus
}

func (d *Data) InAlertJobEventBus() chan<- bo.AlertJob {
	return d.AlertJobEventBus
}

func (d *Data) OutAlertJobEventBus() <-chan bo.AlertJob {
	return d.AlertJobEventBus
}

func (d *Data) InAlertEventBus() chan<- bo.Alert {
	return d.AlertEventBus
}

func (d *Data) OutAlertEventBus() <-chan bo.Alert {
	return d.AlertEventBus
}
