package server

import (
	"encoding/json"
	"net/http"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	khttp "github.com/go-kratos/kratos/v2/transport/http"

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
	uc *auth.Usecase,
	logger log.Logger,
) (*khttp.Server, error) {
	if c == nil || c.Http == nil || c.Http.Addr == "" {
		return nil, merr.Server("http addr is empty")
	}
	opts := []khttp.ServerOption{
		khttp.Address(c.Http.Addr),
		khttp.Middleware(recovery.Recovery(), authenticate(uc)),
		khttp.Logger(logger),
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
	return srv, nil
}
