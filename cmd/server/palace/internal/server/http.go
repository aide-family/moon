package server

import (
	"context"
	nHttp "net/http"
	"time"

	authorizationapi "github.com/aide-family/moon/api/admin/authorization"
	"github.com/aide-family/moon/cmd/server/palace/internal/palaceconf"
	"github.com/aide-family/moon/cmd/server/palace/internal/service/authorization"
	"github.com/aide-family/moon/pkg/env"
	"github.com/aide-family/moon/pkg/helper/middleware"
	"github.com/aide-family/moon/pkg/plugin/mlog"

	"github.com/bufbuild/protovalidate-go"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/transport/http"
)

// NewHTTPServer new an HTTP server.
func NewHTTPServer(bc *palaceconf.Bootstrap, authService *authorization.Service) *http.Server {
	middleware.SetJwtConfig(bc.GetJwt())
	httpConf := bc.GetHttp()
	allowAPIList := bc.GetJwt().GetAllowList()
	// 验证是否登录
	authMiddleware := middleware.Server(
		middleware.JwtServer(),
		middleware.JwtLoginMiddleware(func(ctx context.Context) (*authorizationapi.CheckTokenReply, error) {
			return authService.CheckToken(ctx, &authorizationapi.CheckTokenRequest{})
		}),
		middleware.Rbac(func(ctx context.Context, operation string) (*authorizationapi.CheckPermissionReply, error) {
			return authService.CheckPermission(ctx, &authorizationapi.CheckPermissionRequest{Operation: operation})
		}),
	).Match(middleware.NewWhiteListMatcher(allowAPIList)).Build()

	timeoutMiddleware := middleware.
		Server(middleware.Timeout(httpConf.GetTimeout().AsDuration())).
		Match(middleware.NewWhiteListMatcher(nil)).
		Build()

	opts := []http.ServerOption{
		http.Filter(middleware.Cors()),
		http.Middleware(
			recovery.Recovery(recovery.WithHandler(mlog.RecoveryHandle)),
			timeoutMiddleware,
			tracing.Server(),
			middleware.Logging(log.GetLogger()),
			middleware.I18N(),
			middleware.SourceType(),
			authMiddleware,
			middleware.Validate(protovalidate.WithFailFast(false)),
		),
	}
	if httpConf.GetNetwork() != "" {
		opts = append(opts, http.Network(httpConf.GetNetwork()))
	}
	if httpConf.GetAddr() != "" {
		opts = append(opts, http.Address(httpConf.GetAddr()))
	}
	if httpConf.GetTimeout() != nil {
		opts = append(opts, http.Timeout(time.Hour*24))
	}
	srv := http.NewServer(opts...)

	if env.IsDev() || env.IsTest() || env.IsLocal() {
		// doc
		srv.HandlePrefix("/doc/", nHttp.StripPrefix("/doc/", nHttp.FileServer(nHttp.Dir("./third_party/swagger_ui"))))
	}

	return srv
}
