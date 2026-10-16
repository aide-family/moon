package server

import (
	"encoding/json"
	"log/slog"
	"net/http"

	_ "github.com/go-kratos/kratos/contrib/encoding/json/v3"

	"github.com/go-kratos/kratos/v3/middleware/recovery"
	khttp "github.com/go-kratos/kratos/v3/transport/http"

	v1 "github.com/aide-family/Alioth/api/alioth/v1"
	"github.com/aide-family/Alioth/app/alioth/internal/biz/auth"
	"github.com/aide-family/Alioth/app/alioth/internal/conf"
	"github.com/aide-family/Alioth/app/alioth/internal/service"
	"github.com/aide-family/Alioth/pkg/merr"
)

func NewHTTPServer(
	c *conf.Server,
	health *service.HealthService,
	captcha *service.CaptchaService,
	authn *service.AuthService,
	self *service.SelfService,
	namespaces *service.NamespaceService,
	members *service.MemberService,
	users *service.UserService,
	roles *service.RoleService,
	datasources *service.DatasourceService,
	queries *service.MetricQueryService,
	rules *service.AlertRuleService,
	levels *service.AlertLevelService,
	alertPages *service.AlertPageService,
	events *service.AlertEventService,
	channels *service.NotificationChannelService,
	templates *service.NotificationTemplateService,
	groups *service.NotificationGroupService,
	logs *service.MessageLogService,
	agents *service.AgentService,
	sshCommands *service.SSHCommandService,
	sftpSvc *service.SFTPService,
	k8sSvc *service.K8sProxyService,
	uc *auth.Usecase,
	logger *slog.Logger,
) (*khttp.Server, error) {
	if c == nil || c.Http == nil || c.Http.Addr == "" {
		return nil, merr.Server("http addr is empty")
	}
	opts := []khttp.ServerOption{
		khttp.Address(c.Http.Addr),
		khttp.Middleware(recovery.Recovery(recovery.WithLogger(logger)), authenticate(uc)),
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
	v1.RegisterHealthServiceHTTPServer(srv, health)
	v1.RegisterCaptchaServiceHTTPServer(srv, captcha)
	v1.RegisterAuthServiceHTTPServer(srv, authn)
	v1.RegisterSelfServiceHTTPServer(srv, self)
	v1.RegisterNamespaceServiceHTTPServer(srv, namespaces)
	v1.RegisterMemberServiceHTTPServer(srv, members)
	v1.RegisterUserServiceHTTPServer(srv, users)
	v1.RegisterRoleServiceHTTPServer(srv, roles)
	v1.RegisterDatasourceServiceHTTPServer(srv, datasources)
	v1.RegisterMetricQueryServiceHTTPServer(srv, queries)
	v1.RegisterAlertRuleServiceHTTPServer(srv, rules)
	v1.RegisterAlertLevelServiceHTTPServer(srv, levels)
	v1.RegisterAlertPageServiceHTTPServer(srv, alertPages)
	v1.RegisterAlertEventServiceHTTPServer(srv, events)
	v1.RegisterNotificationChannelServiceHTTPServer(srv, channels)
	v1.RegisterNotificationTemplateServiceHTTPServer(srv, templates)
	v1.RegisterNotificationGroupServiceHTTPServer(srv, groups)
	v1.RegisterMessageLogServiceHTTPServer(srv, logs)
	v1.RegisterAgentServiceHTTPServer(srv, agents)
	v1.RegisterSSHCommandServiceHTTPServer(srv, sshCommands)
	v1.RegisterSFTPServiceHTTPServer(srv, sftpSvc)
	v1.RegisterK8SProxyServiceHTTPServer(srv, k8sSvc)
	return srv, nil
}
