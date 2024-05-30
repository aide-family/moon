package server

import (
	nHttp "net/http"

	"github.com/aide-cloud/moon/cmd/server/houyi/internal/houyiconf"
	"github.com/aide-cloud/moon/pkg/env"
	"github.com/aide-cloud/moon/pkg/helper/middleware"
	"github.com/aide-cloud/moon/pkg/log"

	"github.com/bufbuild/protovalidate-go"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/http"
)

// NewHTTPServer new an HTTP server.
func NewHTTPServer(bc *houyiconf.Bootstrap) *http.Server {
	c := bc.GetServer()

	var opts = []http.ServerOption{
		http.Filter(middleware.Cors()),
		http.Middleware(
			// TODO 开发完再开启
			recovery.Recovery(),
			middleware.Logging(log.GetLogger()),
			middleware.Validate(protovalidate.WithFailFast(true)),
		),
	}
	if c.Http.Network != "" {
		opts = append(opts, http.Network(c.Http.Network))
	}
	if c.Http.Addr != "" {
		opts = append(opts, http.Address(c.Http.Addr))
	}
	if c.Http.Timeout != nil {
		opts = append(opts, http.Timeout(c.Http.Timeout.AsDuration()))
	}
	srv := http.NewServer(opts...)

	if env.IsDev() || env.IsTest() || env.IsLocal() {
		// doc
		srv.HandlePrefix("/doc/", nHttp.StripPrefix("/doc/", nHttp.FileServer(nHttp.Dir("./third_party/swagger_ui"))))
	}

	return srv
}
