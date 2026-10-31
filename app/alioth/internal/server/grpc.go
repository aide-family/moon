package server

import (
	"log/slog"

	"github.com/go-kratos/kratos/v3/middleware/recovery"
	kgrpc "github.com/go-kratos/kratos/v3/transport/grpc"

	v1 "github.com/aide-family/moon/api/alioth/v1"
	megrezv1 "github.com/aide-family/moon/api/megrez/v1"
	merakv1 "github.com/aide-family/moon/api/merak/v1"
	"github.com/aide-family/moon/app/alioth/internal/biz/auth"
	"github.com/aide-family/moon/app/alioth/internal/conf"
	"github.com/aide-family/moon/app/alioth/internal/service"
	"github.com/aide-family/moon/pkg/grpctls"
	"github.com/aide-family/moon/pkg/merr"
)

func NewGRPCServer(
	c *conf.Server,
	health *service.HealthService,
	captcha *service.CaptchaService,
	authn *service.AuthService,
	self *service.SelfService,
	namespaces *service.NamespaceService,
	members *service.MemberService,
	invitations *service.InvitationService,
	messages *service.SystemMessageService,
	users *service.UserService,
	roles *service.RoleService,
	datasources *service.DatasourceService,
	probeTargets *service.ProbeTargetService,
	queries *service.MetricQueryService,
	logQueries *service.LogsQueryService,
	rules *service.AlertRuleService,
	dictionaries *service.AlertDictionaryService,
	alertPages *service.AlertPageService,
	events *service.AlertEventService,
	eval *service.EvalService,
	channels *service.NotificationChannelService,
	templates *service.NotificationTemplateService,
	groups *service.NotificationGroupService,
	subscriptions *service.NotificationSubscriptionService,
	onCalls *service.NotificationOnCallService,
	logs *service.MessageLogService,
	workers *service.WorkerService,
	maintenance *service.MaintenanceService,
	notify *service.NotifyService,
	agents *service.AgentService,
	sshCommands *service.SSHCommandService,
	authenticator *auth.Authenticator,
	logger *slog.Logger,
) (*kgrpc.Server, error) {
	if c == nil || c.Grpc == nil || c.Grpc.Addr == "" {
		return nil, merr.Server("grpc addr is empty")
	}
	opts := []kgrpc.ServerOption{
		kgrpc.Address(c.Grpc.Addr),
		kgrpc.Middleware(recovery.Recovery(recovery.WithLogger(logger)), authenticate(authenticator), protoValidator()),
		kgrpc.StreamMiddleware(recovery.Recovery(recovery.WithLogger(logger)), authenticate(authenticator)),
	}
	if c.Grpc.Network != "" {
		opts = append(opts, kgrpc.Network(c.Grpc.Network))
	}
	if c.Grpc.Timeout != nil {
		opts = append(opts, kgrpc.Timeout(c.Grpc.Timeout.AsDuration()))
	}
	tlsConfig, err := grpctls.ServerFromEnv()
	if err != nil {
		return nil, err
	}
	if tlsConfig != nil {
		opts = append(opts, kgrpc.TLSConfig(tlsConfig))
		logger.Info("grpc TLS enabled")
	}
	srv := kgrpc.NewServer(opts...)
	v1.RegisterHealthServiceServer(srv, health)
	v1.RegisterCaptchaServiceServer(srv, captcha)
	v1.RegisterAuthServiceServer(srv, authn)
	v1.RegisterSelfServiceServer(srv, self)
	v1.RegisterNamespaceServiceServer(srv, namespaces)
	v1.RegisterMemberServiceServer(srv, members)
	v1.RegisterInvitationServiceServer(srv, invitations)
	v1.RegisterSystemMessageServiceServer(srv, messages)
	v1.RegisterUserServiceServer(srv, users)
	v1.RegisterRoleServiceServer(srv, roles)
	v1.RegisterDatasourceServiceServer(srv, datasources)
	v1.RegisterProbeTargetServiceServer(srv, probeTargets)
	v1.RegisterMetricQueryServiceServer(srv, queries)
	v1.RegisterLogsQueryServiceServer(srv, logQueries)
	v1.RegisterAlertRuleServiceServer(srv, rules)
	v1.RegisterAlertDictionaryServiceServer(srv, dictionaries)
	v1.RegisterAlertPageServiceServer(srv, alertPages)
	v1.RegisterAlertEventServiceServer(srv, events)
	v1.RegisterNotificationChannelServiceServer(srv, channels)
	v1.RegisterNotificationTemplateServiceServer(srv, templates)
	v1.RegisterNotificationGroupServiceServer(srv, groups)
	v1.RegisterNotificationSubscriptionServiceServer(srv, subscriptions)
	v1.RegisterNotificationOnCallServiceServer(srv, onCalls)
	v1.RegisterMessageLogServiceServer(srv, logs)
	v1.RegisterWorkerServiceServer(srv, workers)
	v1.RegisterMaintenanceServiceServer(srv, maintenance)
	megrezv1.RegisterEvalServiceServer(srv, eval)
	merakv1.RegisterNotifyServiceServer(srv, notify)
	v1.RegisterAgentServiceServer(srv, agents)
	v1.RegisterSSHCommandServiceServer(srv, sshCommands)
	return srv, nil
}
