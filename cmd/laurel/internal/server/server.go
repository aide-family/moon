package server

import (
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/google/wire"

	"github.com/moon-monitor/moon/cmd/laurel/internal/conf"
	"github.com/moon-monitor/moon/cmd/laurel/internal/service"
	"github.com/moon-monitor/moon/pkg/api/common"
	apiv1 "github.com/moon-monitor/moon/pkg/api/laurel/v1"
	"github.com/moon-monitor/moon/pkg/plugin/server"
)

// ProviderSetServer is server providers.
var ProviderSetServer = wire.NewSet(NewGRPCServer, NewHTTPServer, NewTicker, RegisterService)

// RegisterService register service
func RegisterService(
	c *conf.Bootstrap,
	rpcSrv *grpc.Server,
	httpSrv *http.Server,
	tickerSrv *server.Ticker,
	healthService *service.HealthService,
	metricService *service.MetricService,
) server.Servers {
	common.RegisterHealthServer(rpcSrv, healthService)
	common.RegisterHealthHTTPServer(httpSrv, healthService)
	apiv1.RegisterMetricServer(rpcSrv, metricService)
	apiv1.RegisterMetricHTTPServer(httpSrv, metricService)

	return server.Servers{rpcSrv, httpSrv, tickerSrv}
}
