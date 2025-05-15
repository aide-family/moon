package server

import (
	"embed"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/logging"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/selector"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/transport/http"

	"github.com/aide-family/moon/cmd/laurel/internal/conf"
	"github.com/aide-family/moon/cmd/laurel/internal/helper/middleware"
	"github.com/aide-family/moon/pkg/hello"
	"github.com/aide-family/moon/pkg/i18n"
	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/metric"
	"github.com/aide-family/moon/pkg/middler"
	"github.com/aide-family/moon/pkg/util/docs"
)

//go:embed swagger
var docFS embed.FS

// NewHTTPServer new an HTTP server.
func NewHTTPServer(bc *conf.Bootstrap, logger log.Logger) *http.Server {
	serverConf := bc.GetServer()
	httpConf := serverConf.GetHttp()
	jwtConf := bc.GetAuth().GetJwt()
	i18nConf := bc.GetI18N()
	bundle := i18n.New(i18nConf)
	merr.RegisterGlobalLocalizer(merr.NewLocalizer(bundle))

	authMiddleware := selector.Server(
		middleware.JwtServer(jwtConf.GetSignKey()),
	).Match(middler.AllowListMatcher(jwtConf.GetAllowOperations()...)).Build()
	opts := []http.ServerOption{
		http.Filter(middler.Cors(httpConf)),
		http.Middleware(
			recovery.Recovery(),
			merr.I18n(),
			logging.Server(logger),
			tracing.Server(),
			metric.Server(hello.GetEnv().Name()),
			authMiddleware,
			middler.Validate(),
		),
	}
	if httpConf.GetNetwork() != "" {
		opts = append(opts, http.Network(httpConf.GetNetwork()))
	}
	if httpConf.GetAddr() != "" {
		opts = append(opts, http.Address(httpConf.GetAddr()))
	}
	if httpConf.GetTimeout() != nil {
		opts = append(opts, http.Timeout(httpConf.GetTimeout().AsDuration()))
	}
	srv := http.NewServer(opts...)

	docs.RegisterDocs(srv, docFS, bc.IsDev())
	metric.RegisterRoutes(srv)
	return srv
}
