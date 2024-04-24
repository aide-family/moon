package data

import (
	"github.com/aide-family/moon/app/prom_agent/internal/conf"
	"github.com/aide-family/moon/pkg"
	"github.com/aide-family/moon/pkg/conn"
	"github.com/aide-family/moon/pkg/servers"
	"github.com/aide-family/moon/pkg/strategy"
	"github.com/aide-family/moon/pkg/util/cache"
	"github.com/aide-family/moon/pkg/util/interflow"
	"github.com/aide-family/moon/pkg/util/interflow/build"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// ProviderSetData is data providers.
var ProviderSetData = wire.NewSet(NewData, NewPingRepo)

// Data .
type Data struct {
	cache             cache.GlobalCache
	interflowInstance interflow.AgentInterflow

	log *log.Helper
}

func (d *Data) Cache() cache.GlobalCache {
	return d.cache
}

func (d *Data) Interflow() interflow.AgentInterflow {
	return d.interflowInstance
}

// NewData .
func NewData(c *conf.Bootstrap, logger log.Logger) (*Data, func(), error) {
	d := &Data{
		log: log.NewHelper(log.With(logger, "module", "data")),
	}
	redisConf := c.GetData().GetRedis()
	if redisConf != nil {
		d.cache = cache.NewRedisGlobalCache(conn.NewRedisClient(redisConf))
	} else {
		globalCache, err := cache.NewNutsDbCache()
		if err != nil {
			return nil, nil, err
		}
		d.cache = globalCache
	}

	interflowConf := c.GetInterflow()
	hookConf := conf.BuilderInterflowHook(interflowConf.GetHook())
	interflowOpts := []build.AgentInterflowOption{
		build.WithAgentHttpNetwork(hookConf.GetHttp()),
		build.WithAgentGrpcNetwork(hookConf.GetGrpc()),
		build.WithAgentLogger(logger),
	}
	if !pkg.IsNil(interflowConf.GetMq()) {
		kafkaConf := interflowConf.GetMq().GetKafka()
		if !pkg.IsNil(kafkaConf) {
			kafkaMqServer, err := servers.NewKafkaMQServer(kafkaConf, logger)
			if err != nil {
				return nil, nil, err
			}
			interflowOpts = append(interflowOpts, build.WithAgentKafka(kafkaMqServer))
		}
	}

	interflowInstance, err := build.NewAgentInterflow(interflowOpts...)
	if err != nil {
		return nil, nil, err
	}
	d.interflowInstance = interflowInstance

	// 注册全局告警缓存组件
	alarmCache := strategy.NewAlarmCache(d.Cache())
	strategy.SetAlarmCache(alarmCache)

	cleanup := func() {
		if err = d.Cache().Close(); err != nil {
			d.log.Errorf("close redis error: %v", err)
		}
		if err = d.Interflow().Close(); err != nil {
			d.log.Errorf("close interflow error: %v", err)
		}
		d.log.Info("closing the data resources")
	}
	return d, cleanup, nil
}
