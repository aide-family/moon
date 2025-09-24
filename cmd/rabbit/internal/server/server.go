// Package server is a server package for kratos.
package server

import (
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/google/wire"

	"github.com/aide-family/moon/cmd/rabbit/internal/conf"
	"github.com/aide-family/moon/cmd/rabbit/internal/service"
	"github.com/aide-family/moon/pkg/api/common"
	rabbitv1 "github.com/aide-family/moon/pkg/api/rabbit/v1"
	"github.com/aide-family/moon/pkg/plugin/server"
	"github.com/aide-family/moon/pkg/plugin/server/ticker_server"
)

// ProviderSetServer is server providers.
var ProviderSetServer = wire.NewSet(NewGRPCServer, NewHTTPServer, NewTicker, RegisterService)

// RegisterService registers the service.
func RegisterService(
	c *conf.Bootstrap,
	rpcSrv *grpc.Server,
	httpSrv *http.Server,
	tickerSrv *ticker_server.Ticker,
	healthService *service.HealthService,
	sendService *service.SendService,
	syncService *service.SyncService,
	alertService *service.AlertService,
) server.Servers {
	common.RegisterHealthServer(rpcSrv, healthService)
	common.RegisterHealthHTTPServer(httpSrv, healthService)
	rabbitv1.RegisterSendServer(rpcSrv, sendService)
	rabbitv1.RegisterSyncServer(rpcSrv, syncService)
	rabbitv1.RegisterSendHTTPServer(httpSrv, sendService)
	rabbitv1.RegisterSyncHTTPServer(httpSrv, syncService)
	rabbitv1.RegisterAlertServer(rpcSrv, alertService)
	rabbitv1.RegisterAlertHTTPServer(httpSrv, alertService)

	return server.Servers{rpcSrv, httpSrv, tickerSrv}
}
