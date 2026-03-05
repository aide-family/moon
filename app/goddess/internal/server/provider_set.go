// Package server is a server package for kratos.
package server

import (
	"context"
	"embed"
	nethttp "net/http"
	"strings"

	magicboxv1 "github.com/aide-family/magicbox/api/v1"
	"github.com/aide-family/magicbox/auth/basic"
	"github.com/aide-family/magicbox/oauth"
	"github.com/go-kratos/kratos/v2/transport"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/google/wire"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/aide-family/goddess/internal/conf"
	"github.com/aide-family/goddess/internal/service"
	goddessv1 "github.com/aide-family/goddess/pkg/api/v1"
)

//go:embed swagger
var docFS embed.FS

var (
	ProviderSetServerAll  = wire.NewSet(NewHTTPServer, NewGRPCServer, RegisterService)
	ProviderSetServerHTTP = wire.NewSet(NewHTTPServer, RegisterHTTPService)
	ProviderSetServerGRPC = wire.NewSet(NewGRPCServer, RegisterGRPCService)
)

type Server interface {
	transport.Server
	Name() string
	Instance() transport.Server
}

type server struct {
	transport.Server
	name string
}

func (s *server) Name() string {
	return s.name
}

func (s *server) Instance() transport.Server {
	return s.Server
}

func newServer(name string, srv transport.Server) Server {
	return &server{
		Server: srv,
		name:   name,
	}
}

type Servers []Server

func BindSwagger(httpSrv *http.Server, bc *conf.Bootstrap) {
	binding := basic.HandlerBinding{
		Name:      "Swagger",
		Enabled:   strings.EqualFold(bc.GetEnableSwagger(), "true"),
		BasicAuth: bc.GetSwaggerBasicAuth(),
		Handler:   nethttp.StripPrefix("/doc/", nethttp.FileServer(nethttp.FS(docFS))),
		Path:      "/doc/",
	}
	basic.BindHandlerWithAuth(httpSrv, binding)
}

func BindMetrics(httpSrv *http.Server, bc *conf.Bootstrap) {
	binding := basic.HandlerBinding{
		Name:      "Metrics",
		Enabled:   strings.EqualFold(bc.GetEnableMetrics(), "true"),
		BasicAuth: bc.GetMetricsBasicAuth(),
		Handler:   promhttp.Handler(),
		Path:      "/metrics",
	}
	basic.BindHandlerWithAuth(httpSrv, binding)
}

// RegisterService registers the service.
func RegisterService(
	c *conf.Bootstrap,
	httpSrv *http.Server,
	grpcSrv *grpc.Server,
	authService *service.AuthService,
	healthService *service.HealthService,
	namespaceService *service.NamespaceService,
	userService *service.UserService,
	memberService *service.MemberService,
	selfService *service.SelfService,
	captchaService *service.CaptchaService,
) Servers {
	var srvs Servers

	srvs = append(srvs, RegisterHTTPService(c, httpSrv,
		authService,
		healthService,
		namespaceService,
		userService,
		memberService,
		selfService,
		captchaService,
	)...)
	srvs = append(srvs, RegisterGRPCService(c, grpcSrv,
		authService,
		healthService,
		namespaceService,
		userService,
		memberService,
		selfService,
		captchaService,
	)...)
	return srvs
}

// RegisterHTTPService registers only HTTP service.
func RegisterHTTPService(
	c *conf.Bootstrap,
	httpSrv *http.Server,
	authService *service.AuthService,
	healthService *service.HealthService,
	namespaceService *service.NamespaceService,
	userService *service.UserService,
	memberService *service.MemberService,
	selfService *service.SelfService,
	captchaService *service.CaptchaService,
) Servers {
	magicboxv1.RegisterHealthHTTPServer(httpSrv, healthService)
	goddessv1.RegisterAuthServiceHTTPServer(httpSrv, authService)
	goddessv1.RegisterNamespaceHTTPServer(httpSrv, namespaceService)
	goddessv1.RegisterUserHTTPServer(httpSrv, userService)
	goddessv1.RegisterMemberHTTPServer(httpSrv, memberService)
	goddessv1.RegisterSelfHTTPServer(httpSrv, selfService)
	goddessv1.RegisterCaptchaHTTPServer(httpSrv, captchaService)

	oauth2Handler := oauth.NewOAuth2Handler(c.GetOauth2(), func(ctx context.Context, req *oauth.OAuth2LoginRequest) (string, error) {
		reply, err := authService.OAuth2Login(ctx, req)
		if err != nil {
			return "", err
		}
		return reply.Token, nil
	})
	if err := oauth2Handler.Handler(httpSrv); err != nil {
		panic(err)
	}
	return Servers{newServer("http", httpSrv)}
}

// RegisterGRPCService registers only gRPC service.
func RegisterGRPCService(
	c *conf.Bootstrap,
	grpcSrv *grpc.Server,
	authService *service.AuthService,
	healthService *service.HealthService,
	namespaceService *service.NamespaceService,
	userService *service.UserService,
	memberService *service.MemberService,
	selfService *service.SelfService,
	captchaService *service.CaptchaService,
) Servers {
	magicboxv1.RegisterHealthServer(grpcSrv, healthService)
	goddessv1.RegisterAuthServiceServer(grpcSrv, authService)
	goddessv1.RegisterNamespaceServer(grpcSrv, namespaceService)
	goddessv1.RegisterUserServer(grpcSrv, userService)
	goddessv1.RegisterMemberServer(grpcSrv, memberService)
	goddessv1.RegisterSelfServer(grpcSrv, selfService)
	goddessv1.RegisterCaptchaServer(grpcSrv, captchaService)
	return Servers{newServer("grpc", grpcSrv)}
}

var namespaceAllowList = []string{
	goddessv1.OperationNamespaceCreateNamespace,
	goddessv1.OperationNamespaceUpdateNamespace,
	goddessv1.OperationNamespaceUpdateNamespaceStatus,
	goddessv1.OperationNamespaceDeleteNamespace,
	goddessv1.OperationNamespaceGetNamespace,
	goddessv1.OperationNamespaceListNamespace,
	goddessv1.OperationNamespaceSelectNamespace,
	magicboxv1.OperationHealthHealthCheck,
	goddessv1.OperationSelfInfo,
	goddessv1.OperationSelfNamespaces,
	goddessv1.OperationSelfChangeEmail,
	goddessv1.OperationSelfChangeAvatar,
	goddessv1.OperationSelfRefreshToken,
	goddessv1.OperationUserGetUser,
	goddessv1.OperationUserListUser,
	goddessv1.OperationUserSelectUser,
	goddessv1.OperationUserBanUser,
	goddessv1.OperationUserPermitUser,
}

var authAllowList = []string{
	magicboxv1.OperationHealthHealthCheck,
	oauth.OperationOAuth2Reports,
	oauth.OperationOAuth2Login,
	oauth.OperationOAuth2Callback,
	goddessv1.OperationAuthServiceEmailLogin,
	goddessv1.OperationAuthServiceSendEmailLoginCode,
	goddessv1.OperationCaptchaGetCaptcha,
	goddessv1.OperationNamespaceGetNamespaceSimple,
}
