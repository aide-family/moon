package server

import (
	"fmt"
	"io"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/google/wire"

	"github.com/aide-family/moon/cmd/rabbit/internal/conf"
	"github.com/aide-family/moon/cmd/rabbit/internal/service"
	"github.com/aide-family/moon/pkg/api/common"
	rabbitv1 "github.com/aide-family/moon/pkg/api/rabbit/v1"
	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/plugin/server"
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
	registerCustomerHookTest(httpSrv, c.IsDev())
	return server.Servers{rpcSrv, httpSrv, tickerSrv}
}

func registerCustomerHookTest(srv *http.Server, isDev bool) {
	if !isDev {
		return
	}
	srv.Route("/hook/test").POST("custom", func(ctx http.Context) error {
		// basic auth
		username, password, ok := ctx.Request().BasicAuth()
		if !ok || username != "moon-rabbit" || password != "moon-rabbit" {
			return merr.ErrorUnauthorized("basic auth error").WithMetadata(map[string]string{
				"exist":    fmt.Sprintf("%v", ok),
				"username": username,
				"password": password,
			})
		}
		body, err := io.ReadAll(ctx.Request().Body)
		if err != nil {
			return err
		}
		defer func(Body io.ReadCloser) {
			if err := Body.Close(); err != nil {
				log.Warnf("close request body failed: %v", err)
			}
		}(ctx.Request().Body)
		log.Infof("hook test: %s", string(body))
		return ctx.JSON(200, map[string]string{
			"message": "ok",
			"body":    string(body),
		})
	})
}
