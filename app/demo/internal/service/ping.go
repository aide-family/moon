package service

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"prometheus-manager/api/ping"
	"prometheus-manager/app/demo/internal/biz"
)

// PingService is a Ping service.
type PingService struct {
	ping.UnimplementedPingServer

	log *log.Helper

	uc *biz.PingUseCase
}

// NewPingService new a Ping service.
func NewPingService(uc *biz.PingUseCase, logger log.Logger) *PingService {
	return &PingService{
		uc:  uc,
		log: log.NewHelper(log.With(logger, "module", "service/ping")),
	}
}

// Check implements ping.Check
func (s *PingService) Check(ctx context.Context, in *ping.PingRequest) (*ping.PingReply, error) {
	g, err := s.uc.Ping(ctx, &biz.Ping{Hello: in.Name})
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
