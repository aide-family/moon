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

	"github.com/aide-family/moon/pkg/merr"
)

// Health 是各服务健康服务的最小接口。
type Health interface {
	Status(context.Context) string
}

// Readiness 检查依赖是否可用。liveness 不能调它：依赖抖动不应触发重启。
type Readiness interface {
	Ready(context.Context) error
}

// NewServer 构造健康检查 HTTP 服务。
// /livez 只证明进程存活；/readyz 检查依赖；/health 保留旧契约给探针和脚本。
func NewServer(addr, network string, timeout time.Duration, health Health, ready Readiness, logger *slog.Logger) (*khttp.Server, error) {
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
		writeJSON(w, http.StatusOK, map[string]string{"status": health.Status(r.Context())})
	})
	srv.HandleFunc("/livez", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	srv.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		if ready == nil {
			writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
			return
		}
		if err := ready.Ready(r.Context()); err != nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{
				"status": "unavailable",
				"error":  err.Error(),
			})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	return srv, nil
}

func writeJSON(w http.ResponseWriter, code int, body map[string]string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(body)
}
