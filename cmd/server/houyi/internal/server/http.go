package server

import (
	nHttp "net/http"

	"github.com/aide-family/moon/cmd/server/houyi/internal/houyiconf"
	"github.com/aide-family/moon/pkg/env"
	"github.com/aide-family/moon/pkg/helper/metric"
	"github.com/aide-family/moon/pkg/helper/middleware"
	"github.com/aide-family/moon/pkg/plugin/slog"
	"github.com/aide-family/moon/pkg/util/types"

	"github.com/bufbuild/protovalidate-go"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/http"
)

// NewHTTPServer new an HTTP server.
func NewHTTPServer(bc *houyiconf.Bootstrap) *http.Server {
	var opts = []http.ServerOption{
		http.Filter(middleware.Cors()),
		http.Middleware(
			recovery.Recovery(recovery.WithHandler(slog.RecoveryHandle)),
			middleware.Logging(log.GetLogger()),
			middleware.Validate(protovalidate.WithFailFast(false)),
		),
	}
	httpConf := bc.GetHttp()
	if !types.IsNil(httpConf) {
		if httpConf.GetNetwork() != "" {
			opts = append(opts, http.Network(httpConf.GetNetwork()))
		}
		if httpConf.GetAddr() != "" {
			opts = append(opts, http.Address(httpConf.GetAddr()))
		}
		if httpConf.GetTimeout() != nil {
			opts = append(opts, http.Timeout(httpConf.GetTimeout().AsDuration()))
		}
	}

	srv := http.NewServer(opts...)
	// metrics
	srv.Handle("/metrics", metric.NewMetricHandler(bc.GetMetricsToken()))

	if env.IsDev() || env.IsTest() || env.IsLocal() {
		srv.HandlePrefix("/doc/", nHttp.StripPrefix("/doc/", nHttp.FileServer(nHttp.Dir("./third_party/swagger_ui"))))
	}

	return srv
}
