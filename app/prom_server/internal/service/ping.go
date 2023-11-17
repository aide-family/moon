package service

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"prometheus-manager/api/ping"
	"prometheus-manager/app/prom_server/internal/biz"
	"prometheus-manager/app/prom_server/internal/biz/dobo"
)

// PingService is a Ping service.
type PingService struct {
	ping.UnimplementedPingServer

	log *log.Helper

	uc *biz.PingBiz
}

// NewPingService new a Ping service.
func NewPingService(uc *biz.PingBiz, logger log.Logger) *PingService {
	return &PingService{
		uc:  uc,
		log: log.NewHelper(log.With(logger, "module", "service/ping")),
	}
}

// Check implements ping.Check
func (s *PingService) Check(ctx context.Context, in *ping.PingRequest) (*ping.PingReply, error) {
	g, err := s.uc.Ping(ctx, &dobo.Ping{Hello: in.Name})
	if err != nil {
		return nil, err
	}
	return &ping.PingReply{
		Name:      g.Hello,
		Version:   "",
		Namespace: "",
		Metadata:  nil,
	}, nil
}
