package server

import (
	"log/slog"

	"github.com/go-kratos/kratos/v3/middleware/recovery"
	kgrpc "github.com/go-kratos/kratos/v3/transport/grpc"

	"github.com/aide-family/Alioth/app/alkaid/internal/conf"
	"github.com/aide-family/Alioth/pkg/merr"
)

func NewGRPCServer(c *conf.Server, logger *slog.Logger) (*kgrpc.Server, error) {
	if c == nil || c.Grpc == nil || c.Grpc.Addr == "" {
		return nil, merr.Server("grpc addr is empty")
	}
	opts := []kgrpc.ServerOption{
		kgrpc.Address(c.Grpc.Addr),
		kgrpc.Middleware(recovery.Recovery(recovery.WithLogger(logger))),
	}
	if c.Grpc.Network != "" {
		opts = append(opts, kgrpc.Network(c.Grpc.Network))
	}
	if c.Grpc.Timeout != nil {
		opts = append(opts, kgrpc.Timeout(c.Grpc.Timeout.AsDuration()))
	}
	return kgrpc.NewServer(opts...), nil
}
