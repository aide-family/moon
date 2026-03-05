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
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/aide-family/rabbit/internal/conf"
	"github.com/aide-family/rabbit/internal/service"
	apiv1 "github.com/aide-family/rabbit/pkg/api/v1"
)

//go:embed swagger
var docFS embed.FS

var (
	ProviderSetServerAll  = wire.NewSet(NewHTTPServer, NewGRPCServer, NewJobServer, RegisterService)
	ProviderSetServerHTTP = wire.NewSet(NewHTTPServer, RegisterHTTPService)
	ProviderSetServerGRPC = wire.NewSet(NewGRPCServer, RegisterGRPCService)
	ProviderSetServerJob  = wire.NewSet(NewJobServer, RegisterJobService)
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

func RegisterJobService(jobSrv *JobServer) Servers {
	return Servers{newServer("job", jobSrv)}
}

// RegisterService registers the service.
func RegisterService(
	c *conf.Bootstrap,
	httpSrv *http.Server,
	grpcSrv *grpc.Server,
	jobSrv *JobServer,
	jobService *service.JobService,
	authService *service.AuthService,
	healthService *service.HealthService,
	namespaceService *service.NamespaceService,
	selfService *service.SelfService,
	userService *service.UserService,
	memberService *service.MemberService,
	captchaService *service.CaptchaService,
	emailService *service.EmailService,
	webhookService *service.WebhookService,
	senderService *service.SenderService,
	templateService *service.TemplateService,
	messageLogService *service.MessageLogService,
	recipientGroupService *service.RecipientGroupService,
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
		emailService,
		webhookService,
		senderService,
		templateService,
		messageLogService,
		recipientGroupService,
	)...)
	srvs = append(srvs, RegisterGRPCService(c, grpcSrv,
		healthService,
		namespaceService,
		authService,
		selfService,
		userService,
		memberService,
		captchaService,
		emailService,
		webhookService,
		senderService,
		templateService,
		messageLogService,
		recipientGroupService,
	)...)
	srvs = append(srvs, RegisterJobService(jobSrv)...)
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
	emailService *service.EmailService,
	webhookService *service.WebhookService,
	senderService *service.SenderService,
	templateService *service.TemplateService,
	messageLogService *service.MessageLogService,
	recipientGroupService *service.RecipientGroupService,
) Servers {
	magicboxapiv1.RegisterHealthHTTPServer(httpSrv, healthService)
	goddessv1.RegisterNamespaceHTTPServer(httpSrv, namespaceService)
	goddessv1.RegisterAuthServiceHTTPServer(httpSrv, authService)
	goddessv1.RegisterSelfHTTPServer(httpSrv, selfService)
	goddessv1.RegisterUserHTTPServer(httpSrv, userService)
	goddessv1.RegisterMemberHTTPServer(httpSrv, memberService)
	goddessv1.RegisterCaptchaHTTPServer(httpSrv, captchaService)
	apiv1.RegisterEmailHTTPServer(httpSrv, emailService)
	apiv1.RegisterWebhookHTTPServer(httpSrv, webhookService)
	apiv1.RegisterSenderHTTPServer(httpSrv, senderService)
	apiv1.RegisterTemplateHTTPServer(httpSrv, templateService)
	apiv1.RegisterMessageLogHTTPServer(httpSrv, messageLogService)
	apiv1.RegisterRecipientGroupServiceHTTPServer(httpSrv, recipientGroupService)

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
	emailService *service.EmailService,
	webhookService *service.WebhookService,
	senderService *service.SenderService,
	templateService *service.TemplateService,
	messageLogService *service.MessageLogService,
	recipientGroupService *service.RecipientGroupService,
) Servers {
	magicboxapiv1.RegisterHealthServer(grpcSrv, healthService)
	goddessv1.RegisterNamespaceServer(grpcSrv, namespaceService)
	goddessv1.RegisterAuthServiceServer(grpcSrv, authService)
	goddessv1.RegisterSelfServer(grpcSrv, selfService)
	goddessv1.RegisterUserServer(grpcSrv, userService)
	goddessv1.RegisterMemberServer(grpcSrv, memberService)
	goddessv1.RegisterCaptchaServer(grpcSrv, captchaService)
	apiv1.RegisterEmailServer(grpcSrv, emailService)
	apiv1.RegisterWebhookServer(grpcSrv, webhookService)
	apiv1.RegisterSenderServer(grpcSrv, senderService)
	apiv1.RegisterTemplateServer(grpcSrv, templateService)
	apiv1.RegisterMessageLogServer(grpcSrv, messageLogService)
	apiv1.RegisterRecipientGroupServiceServer(grpcSrv, recipientGroupService)
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
