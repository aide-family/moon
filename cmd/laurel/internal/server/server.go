package server

import (
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/google/wire"

	"github.com/aide-family/moon/cmd/laurel/internal/conf"
	"github.com/aide-family/moon/cmd/laurel/internal/service"
	"github.com/aide-family/moon/pkg/api/common"
	apiv1 "github.com/aide-family/moon/pkg/api/laurel/v1"
	"github.com/aide-family/moon/pkg/plugin/server"
	"github.com/aide-family/moon/pkg/plugin/server/ticker_server"
)

// ProviderSetServer is server providers.
var ProviderSetServer = wire.NewSet(NewGRPCServer, NewHTTPServer, NewTicker, RegisterService)

// RegisterService register service
func RegisterService(
	c *conf.Bootstrap,
	rpcSrv *grpc.Server,
	httpSrv *http.Server,
	tickerSrv *ticker_server.Ticker,
	healthService *service.HealthService,
	metricService *service.MetricService,
) server.Servers {
	common.RegisterHealthServer(rpcSrv, healthService)
	common.RegisterHealthHTTPServer(httpSrv, healthService)
	apiv1.RegisterMetricServer(rpcSrv, metricService)
	apiv1.RegisterMetricHTTPServer(httpSrv, metricService)

	return server.Servers{rpcSrv, httpSrv, tickerSrv}
}
