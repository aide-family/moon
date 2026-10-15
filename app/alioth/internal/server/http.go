package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	v1 "github.com/aide-family/Alioth/api/alioth/v1"
	"github.com/aide-family/Alioth/app/alioth/internal/biz"
	"github.com/aide-family/Alioth/app/alioth/internal/conf"
	"github.com/aide-family/Alioth/app/alioth/internal/service"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	khttp "github.com/go-kratos/kratos/v2/transport/http"
)

func NewHTTPServer(c *conf.Server, health *service.HealthService, identity *service.IdentityService, uc *biz.IdentityUsecase, logger log.Logger) (*khttp.Server, error) {
	if c == nil || c.Http == nil || c.Http.Addr == "" {
		return nil, fmt.Errorf("http addr is empty")
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
	v1.RegisterCaptchaServiceHTTPServer(srv, identity)
	v1.RegisterAuthServiceHTTPServer(srv, identity)
	v1.RegisterSelfServiceHTTPServer(srv, identity)
	v1.RegisterNamespaceServiceHTTPServer(srv, identity)
	v1.RegisterMemberServiceHTTPServer(srv, identity)
	v1.RegisterUserServiceHTTPServer(srv, identity)
	v1.RegisterRoleServiceHTTPServer(srv, identity)
	return srv, nil
}
