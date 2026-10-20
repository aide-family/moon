// Package healthserver 提供 megrez/merak/alkaid 共用的最小健康检查 HTTP 服务。
package healthserver

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-kratos/kratos/v3/middleware/recovery"
	khttp "github.com/go-kratos/kratos/v3/transport/http"

	"github.com/aide-family/Alioth/pkg/merr"
)

// Health 是各服务健康服务的最小接口。
type Health interface {
	Status(context.Context) string
}

// New 构造只暴露 /health 的 HTTP 服务。
func New(addr, network string, timeout time.Duration, health Health, logger *slog.Logger) (*khttp.Server, error) {
	if addr == "" {
		return nil, merr.Server("http addr is empty")
	}
	opts := []khttp.ServerOption{
		khttp.Address(addr),
		khttp.Middleware(recovery.Recovery(recovery.WithLogger(logger))),
	}
	if network != "" {
		opts = append(opts, khttp.Network(network))
	}
	if timeout > 0 {
		opts = append(opts, khttp.Timeout(timeout))
	}
	srv := khttp.NewServer(opts...)
	srv.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": health.Status(r.Context())})
	})
	return srv, nil
}
