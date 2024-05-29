package server

import (
	"context"
	nHttp "net/http"

	"github.com/aide-cloud/moon/pkg/log"
	"github.com/bufbuild/protovalidate-go"
	"github.com/go-kratos/kratos/v2/transport/http"

	authorizationapi "github.com/aide-cloud/moon/api/admin/authorization"
	conf "github.com/aide-cloud/moon/cmd/server/palace/internal/palaceconf"
	"github.com/aide-cloud/moon/cmd/server/palace/internal/service/authorization"
	"github.com/aide-cloud/moon/pkg/env"
	"github.com/aide-cloud/moon/pkg/helper/middleware"
)

// NewHTTPServer new an HTTP server.
func NewHTTPServer(bc *conf.Bootstrap, authService *authorization.Service) *http.Server {
	c := bc.GetServer()

	apiWhiteList := bc.GetServer().GetJwt().GetWhiteList()
	rbacApiWhiteList := append(apiWhiteList, bc.GetServer().GetJwt().GetRbacWhiteList()...)
	// 验证是否登录
	authMiddleware := middleware.Server(
		middleware.JwtServer(),
		middleware.JwtLoginMiddleware(func(ctx context.Context) (bool, error) {
			checkRes, err := authService.CheckToken(ctx, &authorizationapi.CheckTokenRequest{})
			if err != nil {
				return false, err
			}
			return checkRes.GetIsLogin(), nil
		}),
	).Match(middleware.NewWhiteListMatcher(apiWhiteList)).Build()

	// 验证是否有数据权限
	rbacMiddleware := middleware.Server(middleware.Rbac(func(ctx context.Context, operation string) (bool, error) {
		permission, err := authService.CheckPermission(ctx, &authorizationapi.CheckPermissionRequest{
			Operation: operation,
		})
		if err != nil {
			return false, err
		}
		return permission.GetHasPermission(), nil
	})).Match(middleware.NewWhiteListMatcher(rbacApiWhiteList)).Build()

	var opts = []http.ServerOption{
		http.Filter(middleware.Cors()),
		http.Middleware(
			// TODO 开发完再开启
			//recovery.Recovery(),
			middleware.Logging(log.GetLogger()),
			authMiddleware,
			rbacMiddleware,
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
