package server

import (
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/google/wire"

	"github.com/moon-monitor/moon/cmd/houyi/internal/conf"
	"github.com/moon-monitor/moon/cmd/houyi/internal/service"
	"github.com/moon-monitor/moon/pkg/api/common"
	houyiv1 "github.com/moon-monitor/moon/pkg/api/houyi/v1"
	"github.com/moon-monitor/moon/pkg/plugin/server"
)

// ProviderSetServer is server providers.
var ProviderSetServer = wire.NewSet(
	NewGRPCServer,
	NewHTTPServer,
	RegisterService,
	NewCronStrategyJobServer,
	NewCronAlertJobServer,
	NewEventBusServer,
	NewLoadTickerServer,
)

// RegisterService register service
func RegisterService(
	c *conf.Bootstrap,
	rpcSrv *grpc.Server,
	httpSrv *http.Server,
	cronStrategySrv *CronStrategyJobServer,
	cronAlertSrv *CronAlertJobServer,
	eventBusService *EventBusServer,
	loadTickerSrv *LoadTickerServer,
	healthService *service.HealthService,
	syncService *service.SyncService,
	alertService *service.AlertService,
	queryService *service.QueryService,
) server.Servers {
	common.RegisterHealthServer(rpcSrv, healthService)
	common.RegisterHealthHTTPServer(httpSrv, healthService)
	houyiv1.RegisterSyncServer(rpcSrv, syncService)
	houyiv1.RegisterSyncHTTPServer(httpSrv, syncService)
	houyiv1.RegisterAlertHTTPServer(httpSrv, alertService)
	houyiv1.RegisterQueryServer(rpcSrv, queryService)
	houyiv1.RegisterQueryHTTPServer(httpSrv, queryService)
	return server.Servers{
		rpcSrv,
		httpSrv,
		cronStrategySrv,
		cronAlertSrv,
		eventBusService,
		loadTickerSrv,
	}
}
