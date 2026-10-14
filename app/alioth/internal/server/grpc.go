package server

import (
	"fmt"

	v1 "github.com/aide-family/Alioth/api/alioth/v1"
	"github.com/aide-family/Alioth/app/alioth/internal/conf"
	"github.com/aide-family/Alioth/app/alioth/internal/service"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	kgrpc "github.com/go-kratos/kratos/v2/transport/grpc"
)

func NewGRPCServer(c *conf.Server, health *service.HealthService, logger log.Logger) (*kgrpc.Server, error) {
	if c == nil || c.Grpc == nil || c.Grpc.Addr == "" {
		return nil, fmt.Errorf("grpc addr is empty")
	}
	opts := []kgrpc.ServerOption{
		kgrpc.Address(c.Grpc.Addr),
		kgrpc.Middleware(recovery.Recovery()),
		kgrpc.Logger(logger),
	}
	if c.Grpc.Network != "" {
		opts = append(opts, kgrpc.Network(c.Grpc.Network))
	}
	if c.Grpc.Timeout != nil {
		opts = append(opts, kgrpc.Timeout(c.Grpc.Timeout.AsDuration()))
	}
	srv := kgrpc.NewServer(opts...)
	v1.RegisterHealthServiceServer(srv, health)
	return srv, nil
}
