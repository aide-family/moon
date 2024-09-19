package server

import (
	"context"
	nHttp "net/http"

	authorizationapi "github.com/aide-family/moon/api/admin/authorization"
	"github.com/aide-family/moon/cmd/server/palace/internal/palaceconf"
	"github.com/aide-family/moon/cmd/server/palace/internal/service/authorization"
	"github.com/aide-family/moon/pkg/env"
	"github.com/aide-family/moon/pkg/helper/middleware"
	"github.com/aide-family/moon/pkg/util/log"
	"github.com/bufbuild/protovalidate-go"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/transport/http"
)

// NewHTTPServer new an HTTP server.
func NewHTTPServer(bc *palaceconf.Bootstrap, authService *authorization.Service) *http.Server {
	c := bc.GetServer()

	apiWhiteList := bc.GetServer().GetJwt().GetWhiteList()
	blackList := bc.GetServer().GetJwt().GetBlackList()
	rbacAPIWhiteList := append(apiWhiteList, bc.GetServer().GetJwt().GetRbacWhiteList()...)
	// 验证是否登录
	authMiddleware := middleware.Server(
		middleware.JwtServer(),
		middleware.JwtLoginMiddleware(func(ctx context.Context) (*authorizationapi.CheckTokenReply, error) {
			return authService.CheckToken(ctx, &authorizationapi.CheckTokenRequest{})
		}),
	).Match(middleware.NewWhiteListMatcher(apiWhiteList)).Build()

	// 验证是否有数据权限
	rbacMiddleware := middleware.Server(middleware.Rbac(func(ctx context.Context, operation string) (*authorizationapi.CheckPermissionReply, error) {
		return authService.CheckPermission(ctx, &authorizationapi.CheckPermissionRequest{
			Operation: operation,
		})
	})).Match(middleware.NewWhiteListMatcher(rbacAPIWhiteList)).Build()

	var opts = []http.ServerOption{
		http.Filter(middleware.Cors()),
		http.Middleware(
			recovery.Recovery(recovery.WithHandler(log.RecoveryHandle)),
			tracing.Server(),
			middleware.Logging(log.GetLogger()),
			middleware.I18N(),
			middleware.Forbidden(blackList...),
			authMiddleware,
			rbacMiddleware,
			middleware.Validate(protovalidate.WithFailFast(false)),
			middleware.SourceType(),
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
