package server

import (
	nHttp "net/http"

	"github.com/aide-family/moon/api/ping"
	"github.com/aide-family/moon/app/kubemoon/internal/conf"
	"github.com/aide-family/moon/app/kubemoon/internal/data"
	"github.com/aide-family/moon/app/kubemoon/internal/service"
	"github.com/aide-family/moon/pkg/helper/middler"
	"github.com/aide-family/moon/pkg/helper/prom"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/auth/jwt"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/selector"
	"github.com/go-kratos/kratos/v2/middleware/validate"
	"github.com/go-kratos/kratos/v2/transport/http"
	jwtv4 "github.com/golang-jwt/jwt/v4"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// NewHTTPServer new an HTTP server.
func NewHTTPServer(
	c *conf.Server,
	d *data.Data,
	apiWhite *conf.ApiWhite,
	pingService *service.PingService,
	logger log.Logger,
) *http.Server {
	logHelper := log.NewHelper(log.With(logger, "module", "http"))
	defer logHelper.Info("NewHTTPServer done")

	jwt.WithSigningMethod(jwtv4.SigningMethodHS256)
	jwt.WithClaims(func() jwtv4.Claims { return &jwtv4.RegisteredClaims{} })

	allApi := apiWhite.GetAll()
	jwtApis := append(allApi, apiWhite.GetJwtApi()...)

	jwtMiddle := selector.Server(
		middler.JwtServer(),
		middler.MustLogin(d.Cache()),
	).Match(middler.NewWhiteListMatcher(jwtApis)).Build()

	var opts = []http.ServerOption{
		http.Filter(middler.Cors(), middler.Context(), middler.LocalHttpRequestFilter(), middler.ParseRequestForKubernetes()),
		//http.Filter(middler.Cors(), middler.Context(), middler.LocalHttpRequestFilter(), middler.ParseRequestForKubernetes(), ProxyToKubernetes(nil)),
		http.Middleware(
			middler.IpMetric(prom.IpMetricCounter),
			recovery.Recovery(),
			middler.Logging(logger),
			jwtMiddle,
			validate.Validator(),
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

	srv.HandlePrefix("/metrics", promhttp.Handler())
	// doc
	srv.HandlePrefix("/doc/", nHttp.StripPrefix("/doc/", nHttp.FileServer(nHttp.Dir("../../third_party/swagger_ui"))))

	ping.RegisterPingHTTPServer(srv, pingService)

	srv.Route("/api").POST("/upload", func(ctx http.Context) error {
		return ctx.Result(nHttp.StatusOK, "ok")
	})

	return srv
}
