// Package server is a server package for kratos.
package server

import (
	"embed"
	nethttp "net/http"
	"strings"

	goddessv1 "github.com/aide-family/goddess/pkg/api/v1"
	magicboxapiv1 "github.com/aide-family/magicbox/api/v1"
	"github.com/aide-family/magicbox/auth/basic"
	"github.com/aide-family/magicbox/oauth"
	"github.com/go-kratos/kratos/v2/transport"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/google/wire"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/aide-family/marksman/internal/biz/collector"
	"github.com/aide-family/marksman/internal/biz/repository"
	"github.com/aide-family/marksman/internal/conf"
	"github.com/aide-family/marksman/internal/service"
	apiv1 "github.com/aide-family/marksman/pkg/api/v1"
)

//go:embed swagger
var docFS embed.FS

var (
	ProviderSetServerAll  = wire.NewSet(NewHTTPServer, NewGRPCServer, RegisterDatasourceMetrics, RegisterService)
	ProviderSetServerHTTP = wire.NewSet(NewHTTPServer, RegisterHTTPService)
	ProviderSetServerGRPC = wire.NewSet(NewGRPCServer, RegisterGRPCService)
)

// datasourceMetricsReg ensures the datasource status collector is registered with Prometheus when wire builds the graph.
type datasourceMetricsReg struct{}

// RegisterDatasourceMetrics registers the datasource status collector (HTTP probe, marksman_datasource_status) with the default Prometheus registry.
func RegisterDatasourceMetrics(datasourceRepo repository.Datasource) (datasourceMetricsReg, error) {
	lister := collector.NewDatasourceListerFromRepo(datasourceRepo)
	prometheus.MustRegister(collector.NewDatasourceCollector(lister))
	return datasourceMetricsReg{}, nil
}

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
	_ datasourceMetricsReg,
	c *conf.Bootstrap,
	httpSrv *http.Server,
	grpcSrv *grpc.Server,
	authService *service.AuthService,
	healthService *service.HealthService,
	namespaceService *service.NamespaceService,
	selfService *service.SelfService,
	userService *service.UserService,
	memberService *service.MemberService,
	captchaService *service.CaptchaService,
	levelService *service.LevelService,
	datasourceService *service.DatasourceService,
	strategyService *service.StrategyService,
	strategyMetricService *service.StrategyMetricService,
) Servers {
	var srvs Servers

	srvs = append(srvs, RegisterHTTPService(c, httpSrv,
		authService,
		healthService,
		namespaceService,
		selfService,
		userService,
		memberService,
		captchaService,
		levelService,
		datasourceService,
		strategyService,
		strategyMetricService,
	)...)
	srvs = append(srvs, RegisterGRPCService(c,
		grpcSrv,
		healthService,
		namespaceService,
		authService,
		selfService,
		userService,
		memberService,
		captchaService,
		levelService,
		datasourceService,
		strategyService,
		strategyMetricService,
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
	selfService *service.SelfService,
	userService *service.UserService,
	memberService *service.MemberService,
	captchaService *service.CaptchaService,
	levelService *service.LevelService,
	datasourceService *service.DatasourceService,
	strategyService *service.StrategyService,
	strategyMetricService *service.StrategyMetricService,
) Servers {
	magicboxapiv1.RegisterHealthHTTPServer(httpSrv, healthService)
	goddessv1.RegisterAuthServiceHTTPServer(httpSrv, authService)
	goddessv1.RegisterNamespaceHTTPServer(httpSrv, namespaceService)
	goddessv1.RegisterSelfHTTPServer(httpSrv, selfService)
	goddessv1.RegisterUserHTTPServer(httpSrv, userService)
	goddessv1.RegisterMemberHTTPServer(httpSrv, memberService)
	goddessv1.RegisterCaptchaHTTPServer(httpSrv, captchaService)
	apiv1.RegisterLevelHTTPServer(httpSrv, levelService)
	apiv1.RegisterDatasourceHTTPServer(httpSrv, datasourceService)
	apiv1.RegisterStrategyHTTPServer(httpSrv, strategyService)
	apiv1.RegisterStrategyMetricHTTPServer(httpSrv, strategyMetricService)

	oauth2Handler := oauth.NewOAuth2Handler(c.GetOauth2(), authService.Login)
	if err := oauth2Handler.Handler(httpSrv); err != nil {
		panic(err)
	}
	return Servers{newServer("http", httpSrv)}
}

// RegisterGRPCService registers only gRPC service.
func RegisterGRPCService(
	c *conf.Bootstrap,
	grpcSrv *grpc.Server,
	healthService *service.HealthService,
	namespaceService *service.NamespaceService,
	authService *service.AuthService,
	selfService *service.SelfService,
	userService *service.UserService,
	memberService *service.MemberService,
	captchaService *service.CaptchaService,
	levelService *service.LevelService,
	datasourceService *service.DatasourceService,
	strategyService *service.StrategyService,
	strategyMetricService *service.StrategyMetricService,
) Servers {
	magicboxapiv1.RegisterHealthServer(grpcSrv, healthService)
	goddessv1.RegisterAuthServiceServer(grpcSrv, authService)
	goddessv1.RegisterNamespaceServer(grpcSrv, namespaceService)
	goddessv1.RegisterSelfServer(grpcSrv, selfService)
	goddessv1.RegisterUserServer(grpcSrv, userService)
	goddessv1.RegisterMemberServer(grpcSrv, memberService)
	goddessv1.RegisterCaptchaServer(grpcSrv, captchaService)
	apiv1.RegisterLevelServer(grpcSrv, levelService)
	apiv1.RegisterDatasourceServer(grpcSrv, datasourceService)
	apiv1.RegisterStrategyServer(grpcSrv, strategyService)
	apiv1.RegisterStrategyMetricServer(grpcSrv, strategyMetricService)
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
	magicboxapiv1.OperationHealthHealthCheck,
	oauth.OperationOAuth2Reports,
	oauth.OperationOAuth2Login,
	oauth.OperationOAuth2Callback,
	goddessv1.OperationAuthServiceEmailLogin,
	goddessv1.OperationAuthServiceSendEmailLoginCode,
	goddessv1.OperationCaptchaGetCaptcha,
	goddessv1.OperationNamespaceGetNamespaceSimple,
}
