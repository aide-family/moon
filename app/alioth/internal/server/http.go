package server

import (
	"encoding/json"
	"log/slog"
	"net/http"

	_ "github.com/go-kratos/kratos/contrib/encoding/json/v3"

	"github.com/go-kratos/kratos/v3/middleware/recovery"
	khttp "github.com/go-kratos/kratos/v3/transport/http"

	v1 "github.com/aide-family/moon/api/alioth/v1"
	"github.com/aide-family/moon/app/alioth/internal/biz/auth"
	"github.com/aide-family/moon/app/alioth/internal/conf"
	"github.com/aide-family/moon/app/alioth/internal/service"
	"github.com/aide-family/moon/pkg/merr"
)

func NewHTTPServer(
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
	queries *service.MetricQueryService,
	logQueries *service.LogsQueryService,
	rules *service.AlertRuleService,
	dictionaries *service.AlertDictionaryService,
	alertPages *service.AlertPageService,
	events *service.AlertEventService,
	channels *service.NotificationChannelService,
	templates *service.NotificationTemplateService,
	groups *service.NotificationGroupService,
	subscriptions *service.NotificationSubscriptionService,
	onCalls *service.NotificationOnCallService,
	logs *service.MessageLogService,
	workers *service.WorkerService,
	agents *service.AgentService,
	sshCommands *service.SSHCommandService,
	clusters *service.ClusterService,
	kubeResources *service.KubernetesService,
	authenticator *auth.Authenticator,
	logger *slog.Logger,
) (*khttp.Server, error) {
	if c == nil || c.Http == nil || c.Http.Addr == "" {
		return nil, merr.Server("http addr is empty")
	}
	opts := []khttp.ServerOption{
		khttp.Address(c.Http.Addr),
		khttp.Middleware(recovery.Recovery(recovery.WithLogger(logger)), authenticate(authenticator), protoValidator()),
	}
	if c.Http.Network != "" {
		opts = append(opts, khttp.Network(c.Http.Network))
	}
	if c.Http.Timeout != nil {
		opts = append(opts, khttp.Timeout(c.Http.Timeout.AsDuration()))
	}
	srv := khttp.NewServer(opts...)
	srv.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		reply, err := health.Check(r.Context(), &v1.CheckRequest{})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(reply)
	})
	// liveness 与 readiness 分开：依赖抖动只摘流量，不重启进程。
	srv.HandleFunc("/livez", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})
	srv.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := health.Ready(r.Context()); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "unavailable", "error": err.Error()})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})
	v1.RegisterHealthServiceHTTPServer(srv, health)
	v1.RegisterCaptchaServiceHTTPServer(srv, captcha)
	v1.RegisterAuthServiceHTTPServer(srv, authn)
	v1.RegisterSelfServiceHTTPServer(srv, self)
	v1.RegisterNamespaceServiceHTTPServer(srv, namespaces)
	v1.RegisterMemberServiceHTTPServer(srv, members)
	v1.RegisterInvitationServiceHTTPServer(srv, invitations)
	v1.RegisterSystemMessageServiceHTTPServer(srv, messages)
	v1.RegisterUserServiceHTTPServer(srv, users)
	v1.RegisterRoleServiceHTTPServer(srv, roles)
	v1.RegisterDatasourceServiceHTTPServer(srv, datasources)
	v1.RegisterMetricQueryServiceHTTPServer(srv, queries)
	v1.RegisterLogsQueryServiceHTTPServer(srv, logQueries)
	v1.RegisterAlertRuleServiceHTTPServer(srv, rules)
	v1.RegisterAlertDictionaryServiceHTTPServer(srv, dictionaries)
	v1.RegisterAlertPageServiceHTTPServer(srv, alertPages)
	v1.RegisterAlertEventServiceHTTPServer(srv, events)
	v1.RegisterNotificationChannelServiceHTTPServer(srv, channels)
	v1.RegisterNotificationTemplateServiceHTTPServer(srv, templates)
	v1.RegisterNotificationGroupServiceHTTPServer(srv, groups)
	v1.RegisterNotificationSubscriptionServiceHTTPServer(srv, subscriptions)
	v1.RegisterNotificationOnCallServiceHTTPServer(srv, onCalls)
	v1.RegisterMessageLogServiceHTTPServer(srv, logs)
	v1.RegisterWorkerServiceHTTPServer(srv, workers)
	v1.RegisterAgentServiceHTTPServer(srv, agents)
	v1.RegisterSSHCommandServiceHTTPServer(srv, sshCommands)
	v1.RegisterClusterServiceHTTPServer(srv, clusters)
	v1.RegisterKubernetesServiceHTTPServer(srv, kubeResources)
	srv.HandleFunc("/v1/agent/{uid}/shell", agents.ServeShell)
	return srv, nil
}
