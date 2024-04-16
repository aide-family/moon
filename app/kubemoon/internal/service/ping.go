package service

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/moon/api/ping"
)

// PingService is a Ping service.
type PingService struct {
	ping.UnimplementedPingServer

	log *log.Helper
}

// NewPingService new a Ping service.
func NewPingService(logger log.Logger) *PingService {
	return &PingService{
		log: log.NewHelper(log.With(logger, "module", "service/ping")),
	}
}

// Check implements ping.Check
func (s *PingService) Check(ctx context.Context, in *ping.PingRequest) (*ping.PingReply, error) {
	return &ping.PingReply{
		Name:      in.GetName(),
		Version:   "0.0.1",
		Namespace: "default",
		Metadata:  nil,
	}, nil
}
