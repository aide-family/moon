package server

import (
	"encoding/json"
	"log/slog"
	"net/http"

	_ "github.com/go-kratos/kratos/contrib/encoding/json/v3"

	"github.com/go-kratos/kratos/v3/middleware/recovery"
	khttp "github.com/go-kratos/kratos/v3/transport/http"
	"github.com/google/wire"

	"github.com/aide-family/Alioth/app/megrez/internal/conf"
	"github.com/aide-family/Alioth/app/megrez/internal/service"
	"github.com/aide-family/Alioth/pkg/merr"
)

var ProviderSet = wire.NewSet(NewHTTPServer)

func NewHTTPServer(c *conf.Server, health *service.HealthService, logger *slog.Logger) (*khttp.Server, error) {
	if c == nil || c.Http == nil || c.Http.Addr == "" {
		return nil, merr.Server("http addr is empty")
	}
	opts := []khttp.ServerOption{
		khttp.Address(c.Http.Addr),
		khttp.Middleware(recovery.Recovery(recovery.WithLogger(logger))),
	}
	if c.Http.Network != "" {
		opts = append(opts, khttp.Network(c.Http.Network))
	}
	if c.Http.Timeout != nil {
		opts = append(opts, khttp.Timeout(c.Http.Timeout.AsDuration()))
	}
	srv := khttp.NewServer(opts...)
	srv.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": health.Status(r.Context())})
	})
	return srv, nil
}
