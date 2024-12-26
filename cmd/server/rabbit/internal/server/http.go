package server

import (
	nHttp "net/http"

	"github.com/aide-family/moon/cmd/server/rabbit/internal/rabbitconf"
	"github.com/aide-family/moon/pkg/env"
	"github.com/aide-family/moon/pkg/helper/metric"
	"github.com/aide-family/moon/pkg/helper/middleware"
	"github.com/aide-family/moon/pkg/plugin/mlog"
	"github.com/bufbuild/protovalidate-go"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/http"
)

// NewHTTPServer new an HTTP server.
func NewHTTPServer(bc *rabbitconf.Bootstrap) *http.Server {
	c := bc.GetHttp()

	opts := []http.ServerOption{
		http.Filter(middleware.Cors()),
		http.Middleware(
			recovery.Recovery(recovery.WithHandler(mlog.RecoveryHandle)),
			middleware.Logging(log.GetLogger()),
			middleware.Validate(protovalidate.WithFailFast(false)),
		),
	}
	if c.GetNetwork() != "" {
		opts = append(opts, http.Network(c.GetNetwork()))
	}
	if c.GetAddr() != "" {
		opts = append(opts, http.Address(c.GetAddr()))
	}
	if c.GetTimeout() != nil {
		opts = append(opts, http.Timeout(c.GetTimeout().AsDuration()))
	}
	srv := http.NewServer(opts...)

	// metrics
	srv.Handle("/metrics", metric.NewMetricHandler(bc.GetMetricsToken()))

	if env.IsDev() || env.IsTest() || env.IsLocal() {
		// doc
		srv.HandlePrefix("/doc/", nHttp.StripPrefix("/doc/", nHttp.FileServer(nHttp.Dir("./third_party/swagger_ui"))))
	}

	return srv
}
